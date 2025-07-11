# Data Provider 設計書

## 概要

Data Providerは、バックテストシステムにおいて外部データソースから市場データを読み込み、ローソク足データとして提供する責任を持つコンポーネントです。期間指定データ取得、前後データ取得、TimeとIndexの相互変換機能を提供します。

## アーキテクチャ

### インターフェース設計

```go
type DataProvider interface {
    // Time・Index変換
    TimeToIndex(time.Time) (int, error)
    IndexToTime(int) (time.Time, error)
    
    // 期間指定データ取得
    GetCandlesByTime(ctx context.Context, startTime, endTime time.Time) ([]models.Candle, error)
    GetCandlesByIndex(ctx context.Context, startIndex, endIndex int) ([]models.Candle, error)
    
    // 前データ取得
    GetPrevCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error)
    GetPrevCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error)
    
    // 後データ取得
    GetNextCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error)
    GetNextCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error)
}
```

DataProviderインターフェースは、データソースの種類に関わらず統一的なAPIを提供します。現在はCSVファイルからのデータ読み込みのみをサポートしていますが、将来的にはAPI連携、データベース接続等の拡張が可能な設計となっています。

### 機能概要

1. **期間指定取得**: 開始・終了時刻（またはインデックス）を指定してデータを一括取得
2. **前後データ取得**: 基準時刻（またはインデックス）から前後のデータを取得
3. **変換機能**: 時刻とインデックスの相互変換

### 実装クラス

#### CSVProvider

CSVファイルからローソク足データを読み込むプロバイダーです。

**データ構造：**
```go
type CandleIndex struct {
    Timestamp  time.Time
    FileOffset int64    // ファイル内の位置
    LineNumber int      // 行番号（0ベース）
}

type CSVProvider struct {
    Config  models.DataProviderConfig
    index   []CandleIndex  // 軽量インデックス
    file    *os.File       // ファイルハンドル
    indexed bool          // インデックス構築済みフラグ
}
```

**主な機能：**
- CSVファイルの存在確認とオープン
- 軽量インデックスの構築（時刻とファイル位置のみ）
- 高速な時刻検索（バイナリサーチ）
- 必要な分のみのデータ読み込み（メモリ効率）
- データのバリデーション
- エラーハンドリング

**設定項目：**
- `FilePath`: CSVファイルのパス
- `Format`: データフォーマット（現在は"csv"のみサポート）
- `StartTime`: データの開始時刻（オプション）
- `EndTime`: データの終了時刻（オプション）

## データフロー

### 期間指定・前後データ取得
```
CSVファイル → インデックス構築 → 軽量インデックス → 位置特定 → 必要部分のみ読み込み → データ抽出
```

### 処理フロー

1. **初期化**: 初回アクセス時にファイルをスキャンして軽量インデックスを構築
2. **インデックス構築**: 時刻とファイル位置のマッピングを作成
3. **データ検索**: バイナリサーチによる高速位置特定
4. **データ読み込み**: ファイルシークで必要部分のみを読み込み
5. **データ抽出**: オンデマンドでパースして結果を返却

## CSVデータフォーマット

期待されるCSVフォーマット：
```
Date,Time,Open,High,Low,Close,Volume
2024.01.01,00:00,1.1000,1.1050,1.0950,1.1025,1000.0
```

**フィールド説明：**
- Date: 日付 (YYYY.MM.DD形式)
- Time: 時刻 (HH:MM形式)
- Open: 始値
- High: 高値
- Low: 安値
- Close: 終値
- Volume: 出来高

## 新機能の詳細

### 1. 変換機能

#### TimeToIndex
指定された時刻に最も近いローソク足のインデックスを返します。
```go
index, err := provider.TimeToIndex(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC))
```

#### IndexToTime
指定されたインデックスのローソク足の時刻を返します。
```go
timestamp, err := provider.IndexToTime(100)
```

### 2. 期間指定データ取得

#### GetCandlesByTime
開始・終了時刻を指定してローソク足データを取得します。
```go
candles, err := provider.GetCandlesByTime(ctx, startTime, endTime)
```

#### GetCandlesByIndex
開始・終了インデックスを指定してローソク足データを取得します。
```go
candles, err := provider.GetCandlesByIndex(ctx, 100, 200)
```

### 3. 前後データ取得

#### GetPrevCandlesByTime / GetPrevCandlesByIndex
基準時刻（またはインデックス）より前のN個のローソク足を取得します。
```go
prevCandles, err := provider.GetPrevCandlesByTime(ctx, baseTime, 10)
prevCandles, err := provider.GetPrevCandlesByIndex(ctx, 150, 10)
```

#### GetNextCandlesByTime / GetNextCandlesByIndex
基準時刻（またはインデックス）より後のN個のローソク足を取得します。
```go
nextCandles, err := provider.GetNextCandlesByTime(ctx, baseTime, 5)
nextCandles, err := provider.GetNextCandlesByIndex(ctx, 150, 5)
```

## エラーハンドリング

### ファイル関連エラー
- ファイルが存在しない場合は即座にエラーを返す
- ファイルオープンエラーは呼び出し元に伝播

### パーシングエラー
- 無効なCSVレコードは警告ログを出力してスキップ
- EOFに達した場合は正常終了

### データバリデーションエラー
- 無効なローソク足データはスキップして処理続行
- 価格データの整合性チェック（High >= Low等）

### 新機能のエラー
- 範囲外インデックスアクセス時はエラーを返す
- 存在しない時刻指定時は最も近い時刻のデータを返す
- データ未ロード時は自動的にロードを実行

## パフォーマンス特性

### 期間指定・前後データ取得
- **軽量インデックス**: 時刻とファイル位置のみをメモリに保持
- **バイナリサーチ**: O(log n)の高速時刻検索
- **ファイルシーク**: 必要な部分のみを読み込み
- **オンデマンドパース**: 要求されたデータのみをパース

### 使い分け指針
- **期間指定**: 特定期間の分析、バックテストに適している
- **前後データ取得**: 移動平均、テクニカル分析に適している

### メモリ効率性
- **インデックスサイズ**: 1M行のCSVファイルでも約24MB程度（1行あたり24バイト）
- **データ読み込み**: 必要な分のみファイルから読み込み、メモリ使用量を大幅削減
- **スケーラビリティ**: 数GBのCSVファイルでも効率的に処理可能

## 拡張性

### 新しいデータプロバイダーの追加

DataProviderインターフェースを実装することで、新しいデータソースを追加できます：

```go
type APIProvider struct {
    Config models.DataProviderConfig
}

func (p *APIProvider) StreamData(ctx context.Context) (<-chan models.Candle, error) {
    // API連携の実装
}
```

### 設定の拡張

models.DataProviderConfigに新しいフィールドを追加することで、プロバイダー固有の設定を拡張できます。

## 使用例

### 基本的な使用方法

```go
config := models.DataProviderConfig{
    FilePath: "data/EURUSD_M1.csv",
    Format: "csv",
    StartTime: &startTime,
    EndTime: &endTime,
}

provider := NewCSVProvider(config)
```

### 期間指定データ取得
```go
// 時刻指定
candles, err := provider.GetCandlesByTime(ctx, startTime, endTime)
if err != nil {
    log.Fatal(err)
}

// インデックス指定
candles, err := provider.GetCandlesByIndex(ctx, 100, 200)
if err != nil {
    log.Fatal(err)
}
```

### 前後データ取得
```go
// 時刻ベース
prevCandles, err := provider.GetPrevCandlesByTime(ctx, baseTime, 10)
nextCandles, err := provider.GetNextCandlesByTime(ctx, baseTime, 5)

// インデックスベース
prevCandles, err := provider.GetPrevCandlesByIndex(ctx, 150, 10)
nextCandles, err := provider.GetNextCandlesByIndex(ctx, 150, 5)
```

### 変換機能
```go
// 時刻からインデックスへ
index, err := provider.TimeToIndex(someTime)
if err != nil {
    log.Fatal(err)
}

// インデックスから時刻へ
timestamp, err := provider.IndexToTime(50)
if err != nil {
    log.Fatal(err)
}
```

## 依存関係

- `pkg/models`: データ構造とバリデーション
- `pkg/data/parser.go`: CSVパーシング機能
- Go標準ライブラリ: context, os, io, time等

## テスト

- ユニットテスト: `provider_test.go`
- 正常系: 有効なCSVファイルからのデータ読み込み
- 異常系: ファイル不存在、無効なフォーマット、パーシングエラー
- 機能テスト:
  - 期間指定データ取得（Time版・Index版）
  - 前後データ取得（Time版・Index版）
  - 変換機能（TimeToIndex・IndexToTime）
  - 境界値テスト（範囲外アクセス、存在しない時刻等）

## 実装上の注意点

### メモリ使用量
- 軽量インデックスのみをメモリに保持（大幅な使用量削減）
- 1M行のCSVファイルでも約24MB程度のメモリ使用量
- 実際のデータは必要時のみファイルから読み込み

### 初期化コスト
- 初回アクセス時にファイル全体をスキャンしてインデックスを構築
- データ自体は読み込まず、位置情報のみを記録するため高速
- アプリケーション起動時の事前インデックス構築を推奨

### ファイルアクセス
- 複数のデータ取得操作で同じファイルハンドルを使用
- ファイルシークによる効率的な位置移動
- 必要に応じてファイルクローズ・リオープンを実装

### データ更新
- 現在の実装では静的データのみをサポート
- 動的データ更新が必要な場合は再設計が必要
- インデックス再構築のコストを考慮