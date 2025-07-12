# Backtester コンポーネント設計書

## 概要

Backtesterは、FXバックテストシステムの統括コンポーネントとして、Market・Broker・Visualizerの連携を管理し、ユーザーに統一されたバックテスト実行環境を提供します。バックテストの制御、取引API、リアルタイム可視化、およびパフォーマンス統計の収集を担当します。

## アーキテクチャ

### コンポーネント構成

```go
type Backtester struct {
    config           Config
    market           market.Market
    broker           broker.Broker
    visualizer       visualizer.Visualizer
    initialized      bool
    statistics       *models.Statistics
    backtestController *BacktestController
    controlMutex     sync.RWMutex
    ctx              context.Context
    cancel           context.CancelFunc
}
```

Backtesterは以下のコンポーネントを統合管理します：
- **Market**: 市場データの管理と時系列進行
- **Broker**: 注文執行とポジション管理
- **Visualizer**: リアルタイム可視化とWebUI提供
- **BacktestController**: バックテスト実行制御（再生/一時停止/速度調整）

### 設定構造

#### Config構造
```go
type Config struct {
    Market     MarketConfig              `json:"market"`
    Broker     BrokerConfig              `json:"broker"`
    Backtest   BacktestConfig            `json:"backtest"`
    Visualizer models.VisualizerConfig   `json:"visualizer"`
}
```

#### MarketConfig
```go
type MarketConfig struct {
    DataProvider models.DataProviderConfig `json:"data_provider"`
}
```

#### BrokerConfig
```go
type BrokerConfig struct {
    InitialBalance float64 `json:"initial_balance"`
    Spread         float64 `json:"spread"`
}
```

#### BacktestConfig
```go
type BacktestConfig struct {
    StartTime *time.Time `json:"start_time,omitempty"`
    EndTime   *time.Time `json:"end_time,omitempty"`
    MaxSteps  *int       `json:"max_steps,omitempty"`
}
```

## 主要機能

### 1. バックテスター初期化

**NewBacktester()**: 設定に基づいてコンポーネントを作成・統合
```go
func NewBacktester(config Config) (*Backtester, error)
```

**処理フロー:**
1. 設定の妥当性検証（`validateConfig`）
2. DataProviderConfigに期間設定を適用
3. Market作成（`market.NewMarket`でdataProviderConfigから直接作成）
4. Broker作成（Market参照付き）
5. BacktestController作成（Visualizer有効時）
6. コンテキスト管理の設定

**Initialize()**: 各コンポーネントの初期化
```go
func (bt *Backtester) Initialize(ctx context.Context) error
```

### 2. バックテスト実行制御

**Forward()**: 時間進行とコンポーネント連携
```go
func (bt *Backtester) Forward() bool
```

**処理内容:**
- BacktestController制御（再生/一時停止状態チェック）
- 速度制御による待機処理
- Market時間進行（`market.Forward()`）
- Brokerポジション価格更新（`broker.UpdatePositions()`）
- Visualizerへのデータ通知（ローソク足・統計情報）

### 3. 取引API

#### 買い注文実行
```go
func (bt *Backtester) Buy(symbol string, size float64) error
```

#### 売り注文実行
```go
func (bt *Backtester) Sell(symbol string, size float64) error
```

**処理フロー:**
1. 入力値検証（サイズ・価格の妥当性）
2. 注文ID生成（"buy/sell-symbol-timestamp"形式）
3. MarketOrder作成
4. Broker経由での注文実行
5. Visualizerへのトレードイベント通知
6. 統計情報の更新

#### ポジション管理
```go
func (bt *Backtester) ClosePosition(positionID string) error
func (bt *Backtester) CloseAllPositions() error
```

### 4. データアクセスAPI

```go
func (bt *Backtester) GetCurrentTime() time.Time
func (bt *Backtester) GetCurrentPrice(symbol string) float64
func (bt *Backtester) GetPositions() []*models.Position
func (bt *Backtester) GetBalance() float64
func (bt *Backtester) GetTradeHistory() []*models.Trade
func (bt *Backtester) IsFinished() bool
```

### 5. バックテスト制御（BacktestController）

**BacktestController**: バックテストの実行制御を管理
```go
type BacktestController struct {
    bt              *Backtester
    speedCh         chan float64
    playCh          chan bool
    state           models.BacktestControlState
    mutex           sync.RWMutex
    ctx             context.Context
    cancel          context.CancelFunc
}
```

**制御機能:**
- `Play(speed)`: バックテスト開始/再開
- `Pause()`: バックテスト一時停止
- `SetSpeed(speed)`: 実行速度変更
- `GetState()`: 現在の制御状態取得
- `IsRunning()`: 実行状態確認

## データフロー

### 初期化フェーズ
```
Config → NewBacktester → Market/Broker/Visualizer作成 → Initialize → 実行準備完了
```

### 実行フェーズ
```
Forward() → Market時間進行 → Broker価格更新 → Visualizer通知 → 統計更新
     ↓
取引API(Buy/Sell) → Order作成 → Broker実行 → Position管理 → Visualizer通知
```

### 制御フェーズ
```
BacktestController → Play/Pause/SetSpeed → 実行制御 → Forward()での状態反映
```

## Visualizer統合

### リアルタイム可視化機能

**Visualizer連携:**
- ローソク足データのリアルタイム配信
- トレードイベントの可視化
- 統計情報のライブ更新
- バックテスト制御UI提供

**通知イベント:**
```go
bt.visualizer.OnCandleUpdate(candle)      // ローソク足更新
bt.visualizer.OnTradeEvent(trade)         // 取引イベント
bt.visualizer.OnStatisticsUpdate(stats)   // 統計情報更新
bt.visualizer.OnBacktestStateChange(state) // 状態変更
```

### Web UI制御
- WebSocketによるリアルタイム通信
- バックテスト制御（再生/一時停止/速度調整）
- チャート表示とトレード可視化
- パフォーマンス統計のダッシュボード

## エラーハンドリング

### 初期化エラー
```go
func NewBacktester(config Config) (*Backtester, error) {
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    // ...
}
```

**検証項目:**
- DataProvider設定の妥当性
- Broker設定（初期残高・スプレッド）
- Backtest設定（時間範囲・最大ステップ数）
- Visualizer設定

### 実行時エラー
```go
func (bt *Backtester) Buy(symbol string, size float64) error {
    if !bt.initialized {
        return errors.New("backtester not initialized")
    }
    if size <= 0 {
        return errors.New("order size must be positive")
    }
    // ...
}
```

**エラー種類:**
- 初期化前操作エラー
- 入力値検証エラー
- 市場データアクセスエラー
- 注文実行エラー
- ポジション管理エラー

## 統計情報管理

### Statistics構造
```go
type Statistics struct {
    InitialBalance float64
    CurrentBalance float64
    TotalTrades    int
    WinningTrades  int
    LosingTrades   int
    TotalPnL       float64
    // ...
}
```

**統計更新タイミング:**
- Forward()実行時: 残高更新
- 取引実行時: トレード統計更新
- ポジション決済時: PnL統計追加

## 並行処理とスレッドセーフ

### 同期制御
```go
controlMutex sync.RWMutex  // BacktestController状態同期
```

**制御方針:**
- BacktestController状態の読み書き保護
- Visualizer通知の非同期実行
- コンテキストによるキャンセル制御
- チャンネルによる制御信号伝達

## 設定例

### 基本設定
```go
config := Config{
    Market: MarketConfig{
        DataProvider: models.DataProviderConfig{
            FilePath: "data/EURUSD_M1.csv",
            Format:   "csv",
        },
    },
    Broker: BrokerConfig{
        InitialBalance: 10000.0,
        Spread:         0.0001, // 1 pip
    },
    Backtest: BacktestConfig{
        StartTime: &startTime,
        EndTime:   &endTime,
    },
    Visualizer: models.VisualizerConfig{
        Enabled: true,
        Port:    8080,
    },
}
```

### 時間範囲指定
```go
startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
endTime := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC)

config.Backtest.StartTime = &startTime
config.Backtest.EndTime = &endTime
```

## 使用例

### 基本的なバックテスト
```go
func main() {
    config := createConfig()
    
    bt, err := NewBacktester(config)
    if err != nil {
        log.Fatal(err)
    }
    
    err = bt.Initialize(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer bt.Stop()
    
    // シンプルな取引ロジック
    for !bt.IsFinished() {
        currentPrice := bt.GetCurrentPrice("EURUSD")
        
        if currentPrice > 1.1 && len(bt.GetPositions()) == 0 {
            bt.Buy("EURUSD", 10000.0)
        }
        
        if !bt.Forward() {
            break
        }
    }
    
    fmt.Printf("Final Balance: %.2f\n", bt.GetBalance())
}
```

### Visualizer付きバックテスト
```go
func runWithVisualization() {
    config := Config{
        Market: MarketConfig{...},
        Broker: BrokerConfig{...},
        Visualizer: models.VisualizerConfig{
            Enabled: true,
            Port:    8080,
        },
    }
    
    bt, err := NewBacktester(config)
    if err != nil {
        log.Fatal(err)
    }
    
    err = bt.Initialize(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // Webブラウザでlocalhost:8080にアクセス
    // UI経由でバックテスト制御が可能
    
    // プログラム側でも制御可能
    for !bt.IsFinished() {
        // 取引ロジック実行
        executeStrategy(bt)
        
        // Forward()はBacktestControllerの状態に応じて制御される
        if !bt.Forward() {
            break
        }
    }
}
```

## パフォーマンス考慮事項

### メモリ効率
- Statistics構造体による効率的な統計管理
- Visualizer通知の非同期処理
- 適切なキャッシュ戦略（Market側で実装）

### 処理速度
- Forward()での最小限の処理
- 速度制御による柔軟な実行速度調整
- コンテキストキャンセルによる即座の停止

## 拡張性

### 新機能対応
- 複数通貨ペア対応（マルチシンボル）
- 高度な注文タイプ（指値・逆指値）
- カスタム統計指標の追加
- プラグイン戦略システム

### 設定拡張
- 新しい設定項目の追加
- 環境変数による設定オーバーライド
- 設定ファイル読み込み機能

## 依存関係

### 直接依存
- `pkg/market`: Market インターフェース
- `pkg/broker`: Broker インターフェース  
- `pkg/visualizer`: Visualizer インターフェース
- `pkg/models`: データ構造定義

### 間接依存
- `pkg/data`: DataProvider（Market経由）
- WebSocket/HTTP: Visualizer通信
- Go標準ライブラリ: context, sync, time

## テスト戦略

### 単体テスト
- 各API機能の正常・異常系テスト
- 設定検証機能のテスト
- エラーハンドリングのテスト

### 統合テスト
- Market・Broker・Visualizer連携テスト
- BacktestController制御テスト
- 完全なバックテストフローテスト

### パフォーマンステスト
- 大容量データでの処理速度テスト
- メモリ使用量の監視
- 並行処理の安全性テスト

このアーキテクチャにより、Backtesterは統一されたインターフェースでバックテスト機能を提供し、リアルタイム可視化と高度な制御機能を統合した包括的なバックテスト環境を実現します。