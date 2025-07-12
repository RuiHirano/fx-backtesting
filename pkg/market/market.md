# Market コンポーネント設計書

## 概要

Marketコンポーネントは、バックテストシステムにおいて市場データの管理と時系列シミュレーションを担当するコンポーネントです。DataProviderから取得したローソク足データを基に、時系列順にデータを提供し、バックテストの実行エンジンとして機能します。

## アーキテクチャ

### インターフェース設計

```go
type Market interface {
    Initialize(ctx context.Context) error
    Forward() bool
    GetCurrentPrice(symbol string) float64
    GetCurrentTime() time.Time
    GetCurrentCandle(symbol string) *models.Candle
    GetPrevCandles(startTime time.Time, index int) []*models.Candle
    IsFinished() bool
}
```

Marketインターフェースは、時系列データの管理と時間の進行を統一的に管理します。バックテストエンジンはこのインターフェースを通じて市場データにアクセスし、時間を進めながら戦略を実行します。

### 実装クラス

#### MarketImpl

MarketImplは、DataProviderから取得したデータを内部キャッシュに保持し、時系列順に効率的にデータを提供することで、バックテストの実行環境を提供します。

**データ構造：**
```go
type MarketImpl struct {
    provider        data.DataProvider
    candleCache     []*models.Candle
    currentIndex    int
    cacheSize       int // 例: 500
    refillThreshold int // 例: 100
    finished        bool
    initialized     bool
}
```

**主な機能：**
- DataProviderとの連携による効率的なデータ取得とキャッシング
- キャッシュを利用した高速な時系列データの順次アクセス
- バックテスト実行のためのデータアクセス機能
- 市場終了状態の管理

## 機能詳細

### 1. 初期化機能（Initialize）

```go
func (m *MarketImpl) Initialize(ctx context.Context) error
```

**目的**: 市場データの初期化と初期キャッシュの構築

**処理フロー：**
1. DataProviderから初期データ（`cacheSize`分）をまとめて取得し、`candleCache`に格納する。
2. 最初のローソク足データが存在する場合、`currentIndex`を`0`に設定する。
3. `currentTime`を最初のローソク足の時刻に設定する。
4. 初期化フラグを設定する。

**エラーハンドリング：**
- DataProviderからのデータ取得でエラーが発生した場合はエラーを返す。
- 初期データが1件も取得できない場合は、`finished`フラグを`true`に設定し、正常に初期化を完了する。

### 2. 時間進行機能（Forward）

```go
func (m *MarketImpl) Forward() bool
```

**目的**: 次の時間ステップに進む

**処理フロー：**
1. `currentIndex`をインクリメントする。
2. `currentIndex`がキャッシュの終わりに近づいた場合（例: `cacheSize - currentIndex < refillThreshold`）、DataProviderから不足分のデータを非同期で取得し、`candleCache`の後方に追記する。
3. キャッシュを補充しても新しいデータが取得できず、`currentIndex`がキャッシュの末尾に達した場合、`finished`フラグを`true`に設定する。
4. `finished`フラグが`true`でなければ、`currentTime`を更新する。

**戻り値：**
- `true`: 正常に次の時間に進んだ場合。
- `false`: 進むべき次のデータが存在しない場合（既に終了している）。

### 3. 現在価格取得機能（GetCurrentPrice）

```go
func (m *MarketImpl) GetCurrentPrice(symbol string) float64
```

**目的**: 指定されたシンボルの現在価格を取得

**処理：**
- `candleCache[currentIndex]`の終値を返す。
- データが存在しない場合は0.0を返す。

**注意点：**
- 現在の実装では単一シンボルのみをサポート
- 将来的にはマルチシンボル対応を検討

### 4. 現在時刻取得機能（GetCurrentTime）

```go
func (m *MarketImpl) GetCurrentTime() time.Time
```

**目的**: 現在のシミュレーション時刻を取得

**処理：**
- `candleCache[currentIndex]`のタイムスタンプを返す。
- バックテストの時間軸を提供

### 5. 現在ローソク足取得機能（GetCurrentCandle）

```go
func (m *MarketImpl) GetCurrentCandle(symbol string) *models.Candle
```

**目的**: 指定されたシンボルの現在のローソク足データを取得

**処理：**
- `candleCache[currentIndex]`のローソク足（OHLCV）を返す。
- データが存在しない場合はnilを返す。

### 6. 過去ローソク足取得機能（GetPrevCandles）

```go
func (m *MarketImpl) GetPrevCandles(startTime time.Time, index int) []*models.Candle
```

**目的**: 指定された`startTime`から、指定された`index`の直前までのローソク足データを取得する。

**処理フロー：**
1. `index`が0未満、またはキャッシュサイズ以上の場合、あるいは`startTime`が`candleCache[index].Time`より後である場合は、空のスライスを返す。
2. `candleCache`内を`index`から逆方向に探索し、時刻が`startTime`以降である最初のインデックス（`startIndex`）を見つける。
3. `candleCache`から `[startIndex : index]` の範囲のデータをスライスとして取得して返す。
4. `startTime`に該当するデータがキャッシュ内に見つからない場合、キャッシュの先頭から`index`までのデータを返す。

**戻り値：**
- 指定された時間範囲とインデックスに基づいた過去のローソク足データを含むスライス。
- 条件に合うデータがない場合は、空のスライス。

### 7. 終了状態確認機能（IsFinished）

```go
func (m *MarketImpl) IsFinished() bool
```

**目的**: 市場データの終了状態を確認

**処理：**
- `finished`フラグの値を返す。
- バックテストの完了判定に使用

## データフロー

```
DataProvider → StreamData → Market → Forward → Strategy/Broker
    ↓             ↓          ↓         ↓
 CSVファイル → チャンネル → 時系列管理 → バックテスト実行
```

### 処理フロー

1. **初期化フェーズ**
   - DataProviderからデータストリームを取得
   - 最初のローソク足データを読み込み
   - 初期状態を設定

2. **実行フェーズ**
   - Forward()によって時間を進める
   - 各時間ステップで現在データを更新
   - 戦略やブローカーが市場データにアクセス

3. **終了フェーズ**
   - データストリームの終了を検出
   - 終了フラグを設定
   - バックテストの完了を通知

## 使用例

### 基本的な使用方法

```go
// DataProviderの設定
config := models.DataProviderConfig{
    FilePath: "data/EURUSD_M1.csv",
    Format:   "csv",
}

// MarketImplの作成
marketConfig := models.MarketConfig{
    DataProvider: config,
    Symbol:       "EURUSD",
}
market := NewMarket(marketConfig)

// 初期化
ctx := context.Background()
err := market.Initialize(ctx)
if err != nil {
    log.Fatal(err)
}

// バックテストループ
for !market.IsFinished() {
    // 現在の市場データを取得
    currentTime := market.GetCurrentTime()
    currentPrice := market.GetCurrentPrice("EURUSD")
    currentCandle := market.GetCurrentCandle("EURUSD")
    
    // 戦略の実行
    // strategy.Execute(currentTime, currentPrice, currentCandle)
    
    // 次の時間に進む
    if !market.Forward() {
        break
    }
}
```

### 高度な使用方法

```go
// 特定の時間範囲でのバックテスト
startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
endTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)

config := models.DataProviderConfig{
    FilePath: "data/EURUSD_M1.csv",
    Format:   "csv",
}

marketConfig := models.MarketConfig{
    DataProvider: config,
    Symbol:       "EURUSD",
}
market := NewMarket(marketConfig)

// 初期化
ctx := context.Background()
err := market.Initialize(ctx)
if err != nil {
    log.Fatal(err)
}

// 時間制限付きバックテスト
for !market.IsFinished() {
    currentTime := market.GetCurrentTime()
    
    // 時間制限のチェック
    if currentTime.After(endTime) {
        break
    }
    
    // 戦略の実行
    executeStrategy(market)
    
    // 次の時間に進む
    if !market.Forward() {
        break
    }
}
```

## 設計上の考慮事項

### 1. 時系列データの管理

**現在のアプローチ：**
- DataProviderからデータをチャンクで取得し、内部キャッシュ(`candleCache`)に保持する。
- `currentIndex`を移動させることで、時間を線形的に進める。
- キャッシュが一定量（`refillThreshold`）を下回ると、次のデータチャンクを非同期で補充する。

**利点：**
- `Forward`操作がメモリ内で完結するため、非常に高速に実行できる。
- データアクセスの度にDataProviderに問い合わせる必要がなく、I/Oのオーバーヘッドを削減できる。
- 将来的に、キャッシュ内のデータを複数インデックスで参照することで、時間を行き来するような高度な機能（例：`GetPrevCandle`）の実装が容易になる。

**制限：**
- 最初に全データを読み込むわけではないため、全体のデータサイズを事前に把握することは難しい。
- キャッシュサイズと補充の閾値のバランスが、パフォーマンスとメモリ使用量に影響を与える。

### 2. マルチシンボル対応

**現在の状況：**
- 単一シンボルのみサポート
- symbol引数があるが実際は使用されていない

**将来の拡張：**
```go
type MarketImpl struct {
    providers     map[string]data.DataProvider
    currentData   map[string]*models.Candle
    currentTime   time.Time
    synchronized  bool
}
```

### 3. パフォーマンス最適化

**現在のボトルネック：**
- チャンネル操作による若干のオーバーヘッド
- 全データの履歴を保持している

**最適化案：**
- バッファリングによる効率化
- 不要な履歴データの削除
- バッチ処理による高速化

## エラーハンドリング

### 1. 初期化エラー

- DataProviderの初期化エラー
- データファイルの読み込みエラー
- 不正なデータフォーマット

### 2. 実行時エラー

- データストリームの中断
- 不正なシンボル指定
- 初期化前のアクセス

### 3. リカバリー戦略

- エラー時の適切な状態設定
- ログ出力による問題の追跡
- 部分的な復旧機能

## テスト戦略

### 1. ユニットテスト

- `market_test.go`: 基本機能のテスト
- 正常系: 正常なデータフローのテスト
- 異常系: エラーケースのテスト

### 2. 統合テスト

- DataProviderとの連携テスト
- 実際のCSVファイルを使用したテスト
- 長時間実行のテスト

### 3. パフォーマンステスト

- 大量データでの処理速度テスト
- メモリ使用量の測定
- 並行処理のテスト

## 拡張性

### 1. 新しい機能の追加

**時間操作機能：**
```go
type Market interface {
    // 既存メソッド...
    
    // 時間操作機能
    SetTime(time.Time) error
    GetNextCandle(symbol string, offset int) *models.Candle
}
```

**履歴データアクセス：**
```go
type Market interface {
    // 既存メソッド...
    
    // 履歴データアクセス
    GetCandleAtTime(symbol string, time time.Time) *models.Candle
}
```

### 2. 新しい市場タイプの追加

```go
type RealtimeMarket struct {
    // リアルタイムデータ対応
}

type ReplayMarket struct {
    // 高速リプレイ対応
}
```

## 依存関係

### 直接依存

- `pkg/data`: DataProviderインターフェース
- `pkg/models`: Candleデータ構造
- Go標準ライブラリ: context, time

### 間接依存

- CSVファイル: 市場データソース
- ファイルシステム: データファイルへのアクセス

## 実装上の注意点

### 1. メモリ管理

- currentDataの無制限な蓄積を避ける
- 必要に応じて古いデータを削除
- 大量データ処理時のメモリ使用量を監視

### 2. 並行処理

- 現在の実装は並行処理に対応していない
- 将来的にはgoroutineセーフな実装を検討
- チャンネル操作での競合状態に注意

### 3. エラー処理

- 初期化エラーの適切な処理
- データ終了時の状態管理
- リソースの適切な解放

## 品質保証

### 1. コードカバレッジ

- 最低80%以上のカバレッジを維持
- 全エラーパスのテスト
- エッジケースの網羅

### 2. パフォーマンス目標

- 1秒間に1000回以上のForward()実行
- 1GBのCSVファイルを1分以内で処理
- メモリ使用量を100MB以下に抑制

### 3. 可読性とメンテナンス性

- 明確な責任分離
- 適切なコメントとドキュメント
- 一貫したコーディングスタイル