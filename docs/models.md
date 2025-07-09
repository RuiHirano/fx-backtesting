# Models 設計書

## 1. 概要

Models パッケージは、FXバックテストライブラリで使用される全てのデータ構造を定義します。設定情報、市場データ、取引情報、結果データなど、システム全体で共有される基本的なデータモデルを提供します。

## 2. 責務

- システム全体で使用されるデータ構造の定義
- データバリデーション機能の提供
- JSON/CSVシリアライゼーション対応
- 型安全性とデータ整合性の保証
- 構造体間の変換・ヘルパー機能

## 3. ファイル構成

```
pkg/models/
├── config.go           # 設定構造体
├── candle.go           # ローソク足データ
├── order.go            # 注文データ
├── position.go         # ポジションデータ
├── trade.go            # 取引履歴データ
├── result.go           # バックテスト結果
└── models_test.go      # 全モデルのテスト
```

## 4. データ構造定義

### 4.1 Config（設定）

システム全体の設定を管理する構造体群です。

```go
package models

// Config はバックテスト全体の設定を管理します。
type Config struct {
    Market MarketConfig `json:"market" validate:"required"`
    Broker BrokerConfig `json:"broker" validate:"required"`
}

// MarketConfig は市場データに関する設定です。
type MarketConfig struct {
    DataProvider DataProviderConfig `json:"data_provider" validate:"required"`
    Symbol       string             `json:"symbol" validate:"required"`
}

// DataProviderConfig はデータソースに関する設定です。
type DataProviderConfig struct {
    FilePath string `json:"file_path" validate:"required,file"`
    Format   string `json:"format" validate:"required,oneof=csv json"`
}

// BrokerConfig はブローカーに関する設定です。
type BrokerConfig struct {
    InitialBalance float64 `json:"initial_balance" validate:"required,gt=0"`
    Spread         float64 `json:"spread" validate:"gte=0"`
}

// NewDefaultConfig はデフォルト設定を生成します。
func NewDefaultConfig() Config {
    return Config{
        Market: MarketConfig{
            DataProvider: DataProviderConfig{
                Format: "csv",
            },
            Symbol: "EURUSD",
        },
        Broker: BrokerConfig{
            InitialBalance: 10000.0,
            Spread:         0.0001, // 1 pip
        },
    }
}

// Validate は設定の妥当性を検証します。
func (c *Config) Validate() error {
    // バリデーションロジック
    return nil
}
```

### 4.2 Candle（ローソク足データ）

市場データの基本単位となるローソク足データです。

```go
package models

import "time"

// Candle はローソク足データを表します。
type Candle struct {
    Timestamp time.Time `json:"timestamp" csv:"timestamp"`
    Open      float64   `json:"open" csv:"open" validate:"gt=0"`
    High      float64   `json:"high" csv:"high" validate:"gt=0"`
    Low       float64   `json:"low" csv:"low" validate:"gt=0"`
    Close     float64   `json:"close" csv:"close" validate:"gt=0"`
    Volume    float64   `json:"volume" csv:"volume" validate:"gte=0"`
}

// NewCandle は新しいローソク足データを作成します。
func NewCandle(timestamp time.Time, open, high, low, close, volume float64) *Candle {
    return &Candle{
        Timestamp: timestamp,
        Open:      open,
        High:      high,
        Low:       low,
        Close:     close,
        Volume:    volume,
    }
}

// Validate はローソク足データの妥当性を検証します。
func (c *Candle) Validate() error {
    if c.High < c.Low {
        return errors.New("high price must be greater than or equal to low price")
    }
    if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
        return errors.New("prices must be positive")
    }
    if c.Volume < 0 {
        return errors.New("volume must be non-negative")
    }
    return nil
}

// IsValidOHLC は四本値の妥当性をチェックします。
func (c *Candle) IsValidOHLC() bool {
    return c.High >= c.Open && c.High >= c.Close &&
           c.Low <= c.Open && c.Low <= c.Close &&
           c.High >= c.Low
}

// ToCSVRecord はCSV形式の文字列スライスに変換します。
func (c *Candle) ToCSVRecord() []string {
    return []string{
        c.Timestamp.Format("2006-01-02 15:04:05"),
        fmt.Sprintf("%.5f", c.Open),
        fmt.Sprintf("%.5f", c.High),
        fmt.Sprintf("%.5f", c.Low),
        fmt.Sprintf("%.5f", c.Close),
        fmt.Sprintf("%.0f", c.Volume),
    }
}
```

### 4.3 Order（注文データ）

取引注文を表現するデータ構造です。

```go
package models

import "time"

// OrderType は注文タイプを表します。
type OrderType int

const (
    Market OrderType = iota // 成行注文
    Limit                   // 指値注文
    Stop                    // 逆指値注文
)

// String はOrderTypeの文字列表現を返します。
func (ot OrderType) String() string {
    switch ot {
    case Market:
        return "Market"
    case Limit:
        return "Limit"
    case Stop:
        return "Stop"
    default:
        return "Unknown"
    }
}

// OrderSide は注文方向を表します。
type OrderSide int

const (
    Buy OrderSide = iota // 買い注文
    Sell                 // 売り注文
)

// String はOrderSideの文字列表現を返します。
func (os OrderSide) String() string {
    switch os {
    case Buy:
        return "Buy"
    case Sell:
        return "Sell"
    default:
        return "Unknown"
    }
}

// Order は取引注文を表します。
type Order struct {
    ID         string    `json:"id"`
    Symbol     string    `json:"symbol" validate:"required"`
    Type       OrderType `json:"type"`
    Side       OrderSide `json:"side"`
    Size       float64   `json:"size" validate:"gt=0"`
    Price      float64   `json:"price" validate:"gte=0"`
    StopLoss   float64   `json:"stop_loss" validate:"gte=0"`
    TakeProfit float64   `json:"take_profit" validate:"gte=0"`
    Timestamp  time.Time `json:"timestamp"`
}

// NewMarketOrder は成行注文を作成します。
func NewMarketOrder(id, symbol string, side OrderSide, size float64) *Order {
    return &Order{
        ID:        id,
        Symbol:    symbol,
        Type:      Market,
        Side:      side,
        Size:      size,
        Timestamp: time.Now(),
    }
}

// NewLimitOrder は指値注文を作成します。
func NewLimitOrder(id, symbol string, side OrderSide, size, price float64) *Order {
    return &Order{
        ID:        id,
        Symbol:    symbol,
        Type:      Limit,
        Side:      side,
        Size:      size,
        Price:     price,
        Timestamp: time.Now(),
    }
}

// Validate は注文データの妥当性を検証します。
func (o *Order) Validate() error {
    if o.Size <= 0 {
        return errors.New("order size must be positive")
    }
    if o.Type == Limit && o.Price <= 0 {
        return errors.New("limit order must have positive price")
    }
    if o.Symbol == "" {
        return errors.New("symbol is required")
    }
    return nil
}

// IsMarket は成行注文かどうかを判定します。
func (o *Order) IsMarket() bool {
    return o.Type == Market
}

// IsLimit は指値注文かどうかを判定します。
func (o *Order) IsLimit() bool {
    return o.Type == Limit
}
```

### 4.4 Position（ポジションデータ）

保有ポジションを表現するデータ構造です。

```go
package models

import "time"

// Position は保有ポジションを表します。
type Position struct {
    ID           string    `json:"id"`
    Symbol       string    `json:"symbol"`
    Side         OrderSide `json:"side"`
    Size         float64   `json:"size"`
    EntryPrice   float64   `json:"entry_price"`
    CurrentPrice float64   `json:"current_price"`
    PnL          float64   `json:"pnl"`
    OpenTime     time.Time `json:"open_time"`
    StopLoss     float64   `json:"stop_loss,omitempty"`
    TakeProfit   float64   `json:"take_profit,omitempty"`
}

// NewPosition は新しいポジションを作成します。
func NewPosition(id, symbol string, side OrderSide, size, entryPrice float64) *Position {
    return &Position{
        ID:           id,
        Symbol:       symbol,
        Side:         side,
        Size:         size,
        EntryPrice:   entryPrice,
        CurrentPrice: entryPrice,
        PnL:          0.0,
        OpenTime:     time.Now(),
    }
}

// UpdatePrice は現在価格を更新し、PnLを再計算します。
func (p *Position) UpdatePrice(currentPrice float64) {
    p.CurrentPrice = currentPrice
    p.calculatePnL()
}

// calculatePnL は損益を計算します。
func (p *Position) calculatePnL() {
    if p.Side == Buy {
        p.PnL = (p.CurrentPrice - p.EntryPrice) * p.Size
    } else {
        p.PnL = (p.EntryPrice - p.CurrentPrice) * p.Size
    }
}

// IsLong は買いポジションかどうかを判定します。
func (p *Position) IsLong() bool {
    return p.Side == Buy
}

// IsShort は売りポジションかどうかを判定します。
func (p *Position) IsShort() bool {
    return p.Side == Sell
}

// ShouldStopLoss はストップロス条件に達しているかを判定します。
func (p *Position) ShouldStopLoss() bool {
    if p.StopLoss <= 0 {
        return false
    }
    
    if p.IsLong() {
        return p.CurrentPrice <= p.StopLoss
    }
    return p.CurrentPrice >= p.StopLoss
}

// ShouldTakeProfit はテイクプロフィット条件に達しているかを判定します。
func (p *Position) ShouldTakeProfit() bool {
    if p.TakeProfit <= 0 {
        return false
    }
    
    if p.IsLong() {
        return p.CurrentPrice >= p.TakeProfit
    }
    return p.CurrentPrice <= p.TakeProfit
}

// GetMarketValue は現在の市場価値を返します。
func (p *Position) GetMarketValue() float64 {
    return p.CurrentPrice * p.Size
}

// GetPnLPercentage は損益率を返します。
func (p *Position) GetPnLPercentage() float64 {
    if p.EntryPrice == 0 {
        return 0
    }
    return (p.PnL / (p.EntryPrice * p.Size)) * 100
}
```

### 4.5 Trade（取引履歴データ）

完了した取引の履歴を表現するデータ構造です。

```go
package models

import "time"

// TradeStatus は取引ステータスを表します。
type TradeStatus int

const (
    TradeOpen TradeStatus = iota // オープン中
    TradeClosed                  // クローズ済み
    TradeCanceled                // キャンセル済み
)

// String はTradeStatusの文字列表現を返します。
func (ts TradeStatus) String() string {
    switch ts {
    case TradeOpen:
        return "Open"
    case TradeClosed:
        return "Closed"
    case TradeCanceled:
        return "Canceled"
    default:
        return "Unknown"
    }
}

// Trade は完了した取引を表します。
type Trade struct {
    ID         string      `json:"id"`
    Symbol     string      `json:"symbol"`
    Side       OrderSide   `json:"side"`
    Size       float64     `json:"size"`
    EntryPrice float64     `json:"entry_price"`
    ExitPrice  float64     `json:"exit_price"`
    PnL        float64     `json:"pnl"`
    Status     TradeStatus `json:"status"`
    OpenTime   time.Time   `json:"open_time"`
    CloseTime  time.Time   `json:"close_time"`
    Duration   time.Duration `json:"duration"`
}

// NewTradeFromPosition はポジションから取引履歴を作成します。
func NewTradeFromPosition(position *Position, exitPrice float64) *Trade {
    closeTime := time.Now()
    pnl := calculateTradePnL(position.Side, position.Size, position.EntryPrice, exitPrice)
    
    return &Trade{
        ID:         position.ID,
        Symbol:     position.Symbol,
        Side:       position.Side,
        Size:       position.Size,
        EntryPrice: position.EntryPrice,
        ExitPrice:  exitPrice,
        PnL:        pnl,
        Status:     TradeClosed,
        OpenTime:   position.OpenTime,
        CloseTime:  closeTime,
        Duration:   closeTime.Sub(position.OpenTime),
    }
}

// calculateTradePnL は取引の損益を計算します。
func calculateTradePnL(side OrderSide, size, entryPrice, exitPrice float64) float64 {
    if side == Buy {
        return (exitPrice - entryPrice) * size
    }
    return (entryPrice - exitPrice) * size
}

// IsWinning は勝ち取引かどうかを判定します。
func (t *Trade) IsWinning() bool {
    return t.PnL > 0
}

// IsLosing は負け取引かどうかを判定します。
func (t *Trade) IsLosing() bool {
    return t.PnL < 0
}

// IsBreakeven は損益なしかどうかを判定します。
func (t *Trade) IsBreakeven() bool {
    return t.PnL == 0
}

// GetPnLPercentage は損益率を返します。
func (t *Trade) GetPnLPercentage() float64 {
    if t.EntryPrice == 0 {
        return 0
    }
    return (t.PnL / (t.EntryPrice * t.Size)) * 100
}

// GetDurationHours は取引時間を時間単位で返します。
func (t *Trade) GetDurationHours() float64 {
    return t.Duration.Hours()
}

// ToCSVRecord はCSV形式の文字列スライスに変換します。
func (t *Trade) ToCSVRecord() []string {
    return []string{
        t.ID,
        t.Symbol,
        t.Side.String(),
        fmt.Sprintf("%.2f", t.Size),
        fmt.Sprintf("%.5f", t.EntryPrice),
        fmt.Sprintf("%.5f", t.ExitPrice),
        fmt.Sprintf("%.2f", t.PnL),
        t.Status.String(),
        t.OpenTime.Format("2006-01-02 15:04:05"),
        t.CloseTime.Format("2006-01-02 15:04:05"),
        fmt.Sprintf("%.2f", t.GetDurationHours()),
    }
}
```

### 4.6 Result（バックテスト結果）

バックテストの最終結果を表現するデータ構造です。

```go
package models

import "time"

// BacktestResult はバックテストの結果を表します。
type BacktestResult struct {
    // 基本情報
    StartTime time.Time `json:"start_time"`
    EndTime   time.Time `json:"end_time"`
    Duration  time.Duration `json:"duration"`
    
    // 残高情報
    InitialBalance float64 `json:"initial_balance"`
    FinalBalance   float64 `json:"final_balance"`
    TotalPnL       float64 `json:"total_pnl"`
    TotalReturn    float64 `json:"total_return_percent"`
    
    // 取引統計
    TotalTrades    int     `json:"total_trades"`
    WinningTrades  int     `json:"winning_trades"`
    LosingTrades   int     `json:"losing_trades"`
    WinRate        float64 `json:"win_rate_percent"`
    
    // 損益統計
    GrossProfit    float64 `json:"gross_profit"`
    GrossLoss      float64 `json:"gross_loss"`
    AverageWin     float64 `json:"average_win"`
    AverageLoss    float64 `json:"average_loss"`
    LargestWin     float64 `json:"largest_win"`
    LargestLoss    float64 `json:"largest_loss"`
    
    // リスク指標
    MaxDrawdown      float64 `json:"max_drawdown"`
    MaxDrawdownDate  time.Time `json:"max_drawdown_date"`
    SharpeRatio      float64 `json:"sharpe_ratio"`
    ProfitFactor     float64 `json:"profit_factor"`
    
    // 取引履歴
    TradeHistory []Trade `json:"trade_history"`
    
    // 日次統計
    DailyReturns []DailyReturn `json:"daily_returns,omitempty"`
}

// DailyReturn は日次リターンを表します。
type DailyReturn struct {
    Date   time.Time `json:"date"`
    Return float64   `json:"return"`
    PnL    float64   `json:"pnl"`
    Balance float64  `json:"balance"`
}

// NewBacktestResult は新しいバックテスト結果を作成します。
func NewBacktestResult(initialBalance float64) *BacktestResult {
    return &BacktestResult{
        StartTime:      time.Now(),
        InitialBalance: initialBalance,
        FinalBalance:   initialBalance,
        TradeHistory:   make([]Trade, 0),
        DailyReturns:   make([]DailyReturn, 0),
    }
}

// AddTrade は取引を結果に追加します。
func (br *BacktestResult) AddTrade(trade Trade) {
    br.TradeHistory = append(br.TradeHistory, trade)
    br.updateStatistics()
}

// updateStatistics は統計情報を更新します。
func (br *BacktestResult) updateStatistics() {
    br.TotalTrades = len(br.TradeHistory)
    
    if br.TotalTrades == 0 {
        return
    }
    
    var totalPnL, grossProfit, grossLoss float64
    var winningTrades, losingTrades int
    var largestWin, largestLoss float64
    
    for _, trade := range br.TradeHistory {
        totalPnL += trade.PnL
        
        if trade.IsWinning() {
            winningTrades++
            grossProfit += trade.PnL
            if trade.PnL > largestWin {
                largestWin = trade.PnL
            }
        } else if trade.IsLosing() {
            losingTrades++
            grossLoss += -trade.PnL
            if trade.PnL < largestLoss {
                largestLoss = trade.PnL
            }
        }
    }
    
    br.TotalPnL = totalPnL
    br.FinalBalance = br.InitialBalance + totalPnL
    br.TotalReturn = (totalPnL / br.InitialBalance) * 100
    
    br.WinningTrades = winningTrades
    br.LosingTrades = losingTrades
    br.WinRate = float64(winningTrades) / float64(br.TotalTrades) * 100
    
    br.GrossProfit = grossProfit
    br.GrossLoss = grossLoss
    br.LargestWin = largestWin
    br.LargestLoss = largestLoss
    
    if winningTrades > 0 {
        br.AverageWin = grossProfit / float64(winningTrades)
    }
    if losingTrades > 0 {
        br.AverageLoss = grossLoss / float64(losingTrades)
    }
    if grossLoss > 0 {
        br.ProfitFactor = grossProfit / grossLoss
    }
}

// Finalize はバックテスト結果を確定します。
func (br *BacktestResult) Finalize() {
    br.EndTime = time.Now()
    br.Duration = br.EndTime.Sub(br.StartTime)
    br.updateStatistics()
}

// GetSummary は結果の要約を返します。
func (br *BacktestResult) GetSummary() map[string]interface{} {
    return map[string]interface{}{
        "total_return":    br.TotalReturn,
        "total_trades":    br.TotalTrades,
        "win_rate":        br.WinRate,
        "profit_factor":   br.ProfitFactor,
        "max_drawdown":    br.MaxDrawdown,
        "sharpe_ratio":    br.SharpeRatio,
    }
}
```

## 5. バリデーション

### 5.1 バリデーション戦略

```go
// Validator はデータバリデーションのインターフェースです。
type Validator interface {
    Validate() error
}

// ValidationError はバリデーションエラーを表します。
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (ve ValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// ValidateStruct は構造体全体のバリデーションを行います。
func ValidateStruct(v Validator) error {
    return v.Validate()
}
```

## 6. ユーティリティ関数

### 6.1 変換・ヘルパー関数

```go
// GenerateID はユニークなIDを生成します。
func GenerateID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ParseOrderSide は文字列からOrderSideに変換します。
func ParseOrderSide(s string) (OrderSide, error) {
    switch strings.ToLower(s) {
    case "buy":
        return Buy, nil
    case "sell":
        return Sell, nil
    default:
        return Buy, fmt.Errorf("invalid order side: %s", s)
    }
}

// ParseOrderType は文字列からOrderTypeに変換します。
func ParseOrderType(s string) (OrderType, error) {
    switch strings.ToLower(s) {
    case "market":
        return Market, nil
    case "limit":
        return Limit, nil
    case "stop":
        return Stop, nil
    default:
        return Market, fmt.Errorf("invalid order type: %s", s)
    }
}
```

## 7. テスト項目

### 7.1 単体テスト

#### 正常系テスト
- **構造体作成・初期化**
  - 各構造体の正常な作成
  - デフォルト値の確認
  - コンストラクタ関数の動作

- **バリデーション**
  - 正常データでのバリデーション成功
  - 境界値での動作確認

- **計算機能**
  - PnL計算の正確性
  - 損益率計算の正確性
  - 統計値計算の正確性

#### 異常系テスト
- **バリデーション失敗**
  - 不正な値でのバリデーションエラー
  - 必須フィールド未設定エラー
  - 範囲外値での拒否

- **変換エラー**
  - 不正な文字列からの変換失敗
  - 型変換エラーの処理

### 7.2 統合テスト

- **他コンポーネントとの連携**
  - MarketからのCandle受信
  - BrokerでのOrder処理
  - StatisticsでのResult計算

### 7.3 テスト実行方法

```bash
# 全テスト実行
go test ./pkg/models/... -v

# カバレッジ確認
go test -cover ./pkg/models/...

# ベンチマークテスト
go test -bench . ./pkg/models/...
```

## 8. 拡張性

### 8.1 新しいデータ型の追加

```go
// 新しい注文タイプの追加例
const (
    Market OrderType = iota
    Limit
    Stop
    TrailingStop // 新しい注文タイプ
)
```

### 8.2 カスタムフィールドの追加

```go
// カスタムメタデータの追加
type Order struct {
    // 既存フィールド...
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

この設計により、型安全で拡張性の高いデータモデルを提供し、システム全体の基盤となる堅牢なデータ構造を実現します。