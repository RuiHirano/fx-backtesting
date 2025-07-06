# FX バックテストライブラリ設計書

## 概要

Go言語で実装するFXバックテストライブラリ。.hstファイルのヒストリカルデータを使用して、取引戦略のバックテストを実行する。

## 要件

### 基本要件
- .hstファイルからヒストリカルデータを読み込み
- スプレッド設定の対応
- 初期投資額の設定
- 注文実行機能（成行、指値、逆指値）
- バックテスト結果の統計情報出力

### 対応データ形式
- MetaTrader 4/5の.hstファイル
- タイムフレーム：M1, M5, M15, M30, H1, H4, D1

## アーキテクチャ

### コアコンポーネント

#### 1. DataProvider
```go
type DataProvider interface {
    LoadHistoricalData(filePath string) ([]Candle, error)
    GetCandle(timestamp time.Time) (Candle, error)
}
```

#### 2. Backtester
```go
type Backtester struct {
    dataProvider DataProvider
    broker       Broker
    strategy     Strategy
    config       Config
}
```

#### 3. Broker
```go
type Broker interface {
    PlaceOrder(order Order) error
    ClosePosition(positionID string) error
    GetBalance() float64
    GetPositions() []Position
}
```

#### 4. Strategy
```go
type Strategy interface {
    OnTick(candle Candle, broker Broker) error
    OnOrderFill(order Order, broker Broker) error
}
```

### データ構造

#### Candle
```go
type Candle struct {
    Timestamp time.Time
    Open      float64
    High      float64
    Low       float64
    Close     float64
    Volume    int64
}
```

#### Order
```go
type Order struct {
    ID        string
    Symbol    string
    Type      OrderType
    Side      OrderSide
    Size      float64
    Price     float64
    StopLoss  float64
    TakeProfit float64
    Timestamp time.Time
}
```

#### Position
```go
type Position struct {
    ID          string
    Symbol      string
    Side        OrderSide
    Size        float64
    EntryPrice  float64
    CurrentPrice float64
    PnL         float64
    OpenTime    time.Time
}
```

#### Config
```go
type Config struct {
    InitialBalance float64
    Spread         float64
    Commission     float64
    Slippage       float64
    Leverage       float64
}
```

## 実装フロー

### 1. データ読み込み
1. .hstファイルを解析
2. Candleデータに変換
3. 時系列順にソート

### 2. バックテスト実行
1. 設定値の初期化
2. 各時点でのストラテジー実行
3. 注文処理とポジション管理
4. 損益計算

### 3. 結果出力
1. 統計情報の計算
2. パフォーマンス指標の算出
3. 取引履歴の出力

## 使用例

```go
// データプロバイダーの初期化
dataProvider := NewHstDataProvider()
data, err := dataProvider.LoadHistoricalData("EURUSD_M1.hst")

// 設定
config := Config{
    InitialBalance: 10000.0,
    Spread:         0.0001,
    Commission:     0.0,
    Slippage:       0.0,
    Leverage:       100.0,
}

// ストラテジーの実装
strategy := NewMovingAverageStrategy(20, 50)

// バックテスターの作成
backtester := NewBacktester(dataProvider, config)

// バックテスト実行
result, err := backtester.Run(strategy, data)

// 結果の表示
fmt.Printf("Total PnL: %.2f\n", result.TotalPnL)
fmt.Printf("Win Rate: %.2f%%\n", result.WinRate)
```

## パフォーマンス指標

### 基本指標
- 総損益（Total PnL）
- 勝率（Win Rate）
- 最大ドローダウン（Max Drawdown）
- シャープレシオ（Sharpe Ratio）
- プロフィットファクター（Profit Factor）

### 詳細指標
- 平均利益/損失
- 最大連勝/連敗
- 取引回数
- 保有期間統計

## プロジェクト構造

```
fx-backtesting/
├── cmd/
│   └── backtester/
│       └── main.go              # CLIエントリーポイント
├── pkg/
│   ├── backtester/
│   │   ├── backtester.go        # バックテスター本体
│   │   └── backtester_test.go   # バックテスターテスト
│   ├── broker/
│   │   ├── broker.go            # ブローカーインターフェース
│   │   ├── mock_broker.go       # モックブローカー
│   │   ├── simple_broker.go     # シンプルブローカー実装
│   │   └── broker_test.go       # ブローカーテスト
│   ├── data/
│   │   ├── provider.go          # データプロバイダーインターフェース
│   │   ├── hst_provider.go      # .hstファイルプロバイダー
│   │   ├── hst_parser.go        # .hstファイルパーサー
│   │   └── data_test.go         # データ関連テスト
│   ├── models/
│   │   ├── candle.go            # ローソク足データ
│   │   ├── order.go             # 注文データ
│   │   ├── position.go          # ポジションデータ
│   │   ├── config.go            # 設定データ
│   │   └── models_test.go       # モデルテスト
│   ├── strategy/
│   │   ├── strategy.go          # ストラテジーインターフェース
│   │   ├── ma_strategy.go       # 移動平均戦略
│   │   └── strategy_test.go     # ストラテジーテスト
│   ├── statistics/
│   │   ├── calculator.go        # 統計計算
│   │   ├── report.go            # レポート生成
│   │   └── statistics_test.go   # 統計テスト
│   └── utils/
│       ├── time.go              # 時間ユーティリティ
│       ├── math.go              # 数学ユーティリティ
│       └── utils_test.go        # ユーティリティテスト
├── testdata/
│   ├── sample.hst               # サンプルデータ
│   └── test_cases/              # テストケースデータ
├── docs/
│   ├── design.md                # 設計書
│   ├── api.md                   # API仕様書
│   └── examples/                # 使用例
├── examples/
│   ├── simple_backtest/
│   │   └── main.go              # シンプルなバックテスト例
│   └── advanced_strategy/
│       └── main.go              # 高度な戦略例
├── go.mod                       # Go modules
├── go.sum                       # Go modules checksum
├── Makefile                     # ビルドスクリプト
└── README.md                    # プロジェクト説明
```

### ディレクトリ説明
- **cmd/**: 実行可能なアプリケーションのエントリーポイント
- **pkg/**: 再利用可能なライブラリコード
- **testdata/**: テスト用のデータファイル
- **docs/**: ドキュメント類
- **examples/**: 使用例とサンプルコード

## 開発方針

### テスト駆動開発（TDD）
本プロジェクトでは、t-wadaのTDDアプローチに従って開発を行います。

#### TDDサイクル
1. **Red**: 失敗するテストを書く
2. **Green**: テストを通すための最小限のコードを書く
3. **Refactor**: テストを通したままコードを改善する

#### テスト戦略
- **単体テスト**: 各コンポーネントの動作を検証
- **統合テスト**: コンポーネント間の連携を検証
- **受け入れテスト**: 実際の使用例に基づいたエンドツーエンドテスト

#### テスト対象
- DataProviderの.hstファイル読み込み
- Brokerの注文処理とポジション管理
- Backtesterの実行ロジック
- 統計計算の正確性
- エラーハンドリング

#### テストデータ
- 実際の.hstファイルを使用したテストデータ
- エッジケース用の合成データ
- パフォーマンステスト用の大容量データ

## 拡張性

### 将来の機能拡張
- 複数通貨ペア対応
- リアルタイム取引対応
- 複数ストラテジー同時実行
- ポートフォリオ最適化
- 機械学習統合

### プラグインアーキテクチャ
- カスタムインジケーター
- カスタムオーダータイプ
- カスタムリスク管理
- カスタムレポート形式