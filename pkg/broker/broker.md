# Broker コンポーネント設計書

## 概要

Brokerコンポーネントは、バックテストシステムにおいて取引の実行、ポジション管理、残高管理を担当するコンポーネントです。実際のFXブローカーの機能をシミュレートし、現実的な取引環境を提供することで、戦略の有効性を正確に評価できる環境を提供します。

## アーキテクチャ

### インターフェース設計

```go
type Broker interface {
    PlaceOrder(order *models.Order) error
    CancelOrder(orderID string) error
    GetPendingOrders() []*models.Order
    GetPositions() []*models.Position
    GetBalance() float64
    ClosePosition(positionID string) error
    UpdatePositions()
    ProcessPendingOrders()
    GetTradeHistory() []*models.Trade
}
```

Brokerインターフェースは、取引の実行から決済まで、トレーディングに必要な全ての機能を統一的に管理します。成行注文、指値注文、逆指値注文をサポートし、内部で保留中の注文を管理します。バックテストエンジンはこのインターフェースを通じて市場での取引をシミュレートします。

### 実装クラス

#### SimpleBroker

SimpleBrokerは、高度なFX取引機能を提供するBrokerインターフェースの実装です。成行注文、指値注文、逆指値注文をサポートし、レバレッジ取引、スプレッド、証拠金管理などの現実的な取引条件を考慮したシミュレーションを提供します。

**データ構造：**
```go
type SimpleBroker struct {
    config         models.BrokerConfig
    market         market.Market
    balance        float64
    positions      map[string]*models.Position
    pendingOrders  map[string]*models.Order
    tradeHistory   []*models.Trade
}
```

**主な機能：**
- 複数の注文種別のサポート（成行、指値、逆指値）
- 保留中注文の内部管理と自動約定処理
- レバレッジを考慮した証拠金計算（1:100レバレッジ）
- スプレッドを考慮した現実的な約定価格設定
- リアルタイムでのポジション価格更新
- 完了した取引の履歴管理
- 注文のキャンセル機能

## 機能詳細

### 1. 注文実行機能（PlaceOrder）

```go
func (b *SimpleBroker) PlaceOrder(order *models.Order) error
```

**目的**: 注文を受け付け、種別に応じて即座に実行または保留状態にする

**サポートする注文種別：**
- **成行注文（Market Order）**: 現在価格で即座に約定
- **指値注文（Limit Order）**: 指定価格に達した時に約定
- **逆指値注文（Stop Order）**: 指定価格を超えた時に約定

**処理フロー：**
1. 注文のバリデーションを実行（サイズ、価格、シンボルの妥当性）
2. 注文種別を判定する
3. **成行注文の場合:**
   - 市場から現在価格を取得する
   - スプレッドを適用した約定価格を計算する
     - 買い注文: `executionPrice = currentPrice + spread` (Ask価格)
     - 売り注文: `executionPrice = currentPrice - spread` (Bid価格)
   - 必要証拠金を計算し、残高チェックを実行する
   - 即座にポジションを作成し、残高を更新する
4. **指値・逆指値注文の場合:**
   - 注文を`pendingOrders`マップに保存する
   - 証拠金の事前確保は行わない（約定時に実行）
   - 注文IDを生成して管理する

**注文種別詳細：**

#### 成行注文（Market Order）
```go
type Order struct {
    ID     string
    Type   OrderType  // MarketOrder
    Symbol string
    Side   OrderSide  // Buy or Sell
    Size   float64
    // 価格指定は不要
}
```

#### 指値注文（Limit Order）
```go
type Order struct {
    ID         string
    Type       OrderType  // LimitOrder
    Symbol     string
    Side       OrderSide  // Buy or Sell
    Size       float64
    LimitPrice float64    // 約定希望価格
}
```
- **買い指値**: 現在価格が指値価格以下になった時に約定
- **売り指値**: 現在価格が指値価格以上になった時に約定

#### 逆指値注文（Stop Order）
```go
type Order struct {
    ID        string
    Type      OrderType  // StopOrder
    Symbol    string
    Side      OrderSide  // Buy or Sell
    Size      float64
    StopPrice float64    // トリガー価格
}
```
- **買い逆指値**: 現在価格が逆指値価格以上になった時に成行で約定
- **売り逆指値**: 現在価格が逆指値価格以下になった時に成行で約定

**エラーハンドリング：**
- 無効な注文サイズ（0以下）の場合はエラーを返す
- 無効なシンボルの場合はエラーを返す
- 成行注文で証拠金不足の場合は`insufficient balance`エラーを返す
- 指値・逆指値で無効な価格指定の場合はエラーを返す

**約定メカニズム：**
- 成行注文は即座に約定される
- 指値・逆指値注文は条件が満たされるまで保留される
- スプレッドによる実際のブローカー環境をシミュレート
- レバレッジにより少額の証拠金で大きなポジションを持てる

### 2. 注文キャンセル機能（CancelOrder）

```go
func (b *SimpleBroker) CancelOrder(orderID string) error
```

**目的**: 保留中の注文をキャンセルする

**処理フロー：**
1. 指定されたorderIDで保留注文を検索する
2. 注文が存在しない場合はエラーを返す
3. 注文が既に約定済みの場合はエラーを返す
4. `pendingOrders`マップから注文を削除する
5. 必要に応じて確保済み証拠金を解放する（将来実装）

**エラーハンドリング：**
- 存在しない注文IDの場合は`order not found`エラーを返す
- 既に約定済みの注文の場合は`order already executed`エラーを返す

### 3. 保留注文取得機能（GetPendingOrders）

```go
func (b *SimpleBroker) GetPendingOrders() []*models.Order
```

**目的**: 現在保留中の全注文を取得

**処理：**
- 内部マップ（`pendingOrders`）に保存されている全注文をスライスとして返す
- 注文には以下の情報が含まれる：
  - 注文ID、注文種別、シンボル、売買区分
  - 注文サイズ、指値価格、逆指値価格
  - 注文作成時刻、有効期限（将来実装）

**使用場面：**
- 戦略による保留注文の状況確認
- 注文管理とキャンセル判定
- 注文の重複チェック

### 4. 保留注文処理機能（ProcessPendingOrders）

```go
func (b *SimpleBroker) ProcessPendingOrders()
```

**目的**: 保留中の注文を現在の市場価格と照らし合わせて約定処理する

**処理フロー：**
1. 全ての保留注文を順次確認する
2. 各注文について現在の市場価格を取得する
3. 注文種別と価格条件を確認し、約定条件が満たされているかチェックする
4. 約定条件が満たされた場合：
   - 証拠金チェックを実行する
   - ポジションを作成する
   - 残高を更新する
   - 取引履歴を作成する
   - 保留注文から削除する
5. 約定条件が満たされない場合は次の注文へ進む

**約定条件：**
- **買い指値**: `currentPrice <= limitPrice`
- **売り指値**: `currentPrice >= limitPrice`
- **買い逆指値**: `currentPrice >= stopPrice`
- **売り逆指値**: `currentPrice <= stopPrice`

**使用タイミング：**
- 市場データ更新後（`market.Forward()`の後）
- `UpdatePositions()`と同時に実行
- バックテストのメインループ内

### 5. ポジション取得機能（GetPositions）

```go
func (b *SimpleBroker) GetPositions() []*models.Position
```

**目的**: 現在保有している全ポジションを取得

**処理：**
- 内部マップ（`positions`）に保存されている全ポジションをスライスとして返す
- ポジションは作成順ではなく、マップの順序で返される
- 各ポジションには以下の情報が含まれる：
  - ポジションID、シンボル、売買区分
  - ポジションサイズ、約定価格、現在価格
  - オープン時刻、損益情報

**使用場面：**
- 戦略による現在のポジション状況の確認
- リスク管理のためのエクスポージャー計算
- ポジション決済の対象選択

### 6. 残高取得機能（GetBalance）

```go
func (b *SimpleBroker) GetBalance() float64
```

**目的**: 現在の口座残高を取得

**処理：**
- 内部で管理している`balance`フィールドの値を返す
- この残高は証拠金として使用されているため、実質的な利用可能資金を表す
- ポジションの含み損益は残高に反映されない（決済時のみ反映）

**残高の変動要因：**
- 初期残高: 設定で指定された初期資金
- ポジション作成時: 必要証拠金の差し引き
- ポジション決済時: 証拠金の返却と損益の反映

### 7. ポジション決済機能（ClosePosition）

```go
func (b *SimpleBroker) ClosePosition(positionID string) error
```

**目的**: 指定されたポジションを決済する

**処理フロー：**
1. ポジションIDでポジションを検索し、存在しない場合はエラーを返す
2. 市場から現在価格を取得する
3. ポジション種別に応じてスプレッドを適用した決済価格を計算する
   - 買いポジション決済: `closePrice = currentPrice - spread` (Bid価格で売却)
   - 売りポジション決済: `closePrice = currentPrice + spread` (Ask価格で買戻し)
4. 損益を計算する
   - 買いポジション: `pnl = (closePrice - entryPrice) * size`
   - 売りポジション: `pnl = (entryPrice - closePrice) * size`
5. 口座残高を更新する
   - 証拠金を返却: `balance += requiredMargin`
   - 損益を反映: `balance += pnl`
6. 取引履歴を作成して保存する
7. ポジションを内部マップから削除する

**エラーハンドリング：**
- 存在しないポジションIDの場合はエラーを返す
- 無効な市場価格の場合はエラーを返す

### 8. ポジション更新機能（UpdatePositions）

```go
func (b *SimpleBroker) UpdatePositions()
```

**目的**: 全ポジションの現在価格を市場データで更新

**処理フロー：**
1. 保有している全ポジションを順次処理する
2. 各ポジションのシンボルについて市場から現在価格を取得する
3. 取得した価格が有効（0より大きい）な場合、ポジションの現在価格を更新する
4. ポジション内部で含み損益が自動的に再計算される
5. 保留注文の処理も同時に実行する（`ProcessPendingOrders()`を呼び出し）

**使用タイミング：**
- 市場データが更新された後（`market.Forward()`の後）
- リアルタイムでの損益計算が必要な場合
- Visualizerへのデータ送信前

### 9. 取引履歴取得機能（GetTradeHistory）

```go
func (b *SimpleBroker) GetTradeHistory() []*models.Trade
```

**目的**: 完了した取引の履歴を取得

**処理：**
- 内部で管理している`tradeHistory`スライスを返す
- 取引履歴には以下の情報が含まれる：
  - 取引ID、シンボル、売買区分、サイズ
  - 約定価格、決済価格、損益
  - オープン時刻、クローズ時刻、保有期間
  - 取引ステータス

**取引履歴の活用：**
- バックテスト結果の分析
- 戦略のパフォーマンス評価
- 統計情報の計算（勝率、平均損益など）

## データフロー

```
Strategy → Order → Broker → Position/PendingOrder → Market
    ↓         ↓        ↓              ↓               ↓
 判断ロジック → 注文配置 → 注文管理/約定処理 → 価格更新/約定チェック → 市場データ
```

### 処理フロー

1. **注文受付フェーズ**
   - 戦略からの注文を受け取る
   - 注文種別を判定（成行/指値/逆指値）
   - 成行注文は即座に約定、指値・逆指値は保留

2. **注文管理フェーズ**
   - 保留注文の内部管理
   - 市場価格更新時の約定条件チェック
   - 条件満足時の自動約定処理
   - 注文キャンセル処理

3. **ポジション管理フェーズ**
   - 市場データ更新に伴うポジション価格の更新
   - 含み損益の計算
   - 保留注文の約定チェックと処理
   - リスク管理（将来実装予定）

4. **決済フェーズ**
   - ポジション決済の実行
   - 損益確定と残高更新
   - 取引履歴の記録

### 注文種別による処理の違い

#### 成行注文の処理フロー
```
PlaceOrder → 価格取得 → 証拠金チェック → 即座に約定 → ポジション作成
```

#### 指値・逆指値注文の処理フロー
```
PlaceOrder → バリデーション → PendingOrdersに保存
                                    ↓
市場更新 → ProcessPendingOrders → 約定条件チェック → 約定/継続保留
```

## 設定とコンフィギュレーション

### BrokerConfig

```go
type BrokerConfig struct {
    InitialBalance float64 `json:"initial_balance"`
    Spread         float64 `json:"spread"`
}
```

**設定項目：**
- `InitialBalance`: 初期残高（デフォルト: 10,000.0）
- `Spread`: スプレッド（デフォルト: 0.0001 = 1 pip）

**設定例：**
```go
config := models.BrokerConfig{
    InitialBalance: 100000.0,  // 10万円
    Spread:         0.0002,    // 2 pips
}
```

### デフォルト設定

```go
func NewDefaultConfig() Config {
    return Config{
        Broker: BrokerConfig{
            InitialBalance: 10000.0,
            Spread:         0.0001, // 1 pip
        },
    }
}
```

## 使用例

### 基本的な使用方法

```go
// ブローカー設定
brokerConfig := models.BrokerConfig{
    InitialBalance: 50000.0,
    Spread:         0.0001,
}

// 市場の初期化（別途必要）
market := initializeMarket() // 実装は省略

// ブローカー作成
broker := NewSimpleBroker(brokerConfig, market)

// 1. 成行注文の実行
marketOrder := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
err := broker.PlaceOrder(marketOrder)
if err != nil {
    log.Printf("成行注文エラー: %v", err)
}

// 2. 指値注文の実行
currentPrice := market.GetCurrentPrice("EURUSD")
limitOrder := models.NewLimitOrder("order-2", "EURUSD", models.Buy, 10000.0, currentPrice-0.0010)
err = broker.PlaceOrder(limitOrder)
if err != nil {
    log.Printf("指値注文エラー: %v", err)
}

// 3. 逆指値注文の実行
stopOrder := models.NewStopOrder("order-3", "EURUSD", models.Sell, 10000.0, currentPrice-0.0020)
err = broker.PlaceOrder(stopOrder)
if err != nil {
    log.Printf("逆指値注文エラー: %v", err)
}

// ポジション確認
positions := broker.GetPositions()
fmt.Printf("現在のポジション数: %d\n", len(positions))

// 保留注文確認
pendingOrders := broker.GetPendingOrders()
fmt.Printf("保留中の注文数: %d\n", len(pendingOrders))

// 残高確認
balance := broker.GetBalance()
fmt.Printf("現在の残高: %.2f\n", balance)
```

### 高度な使用方法（バックテストループ）

```go
// バックテストシミュレーション
for !market.IsFinished() {
    // 市場データ更新
    market.Forward()
    
    // ポジション価格更新と保留注文処理
    broker.UpdatePositions()
    
    // 現在価格を取得
    currentPrice := market.GetCurrentPrice("EURUSD")
    
    // 戦略による判断
    if shouldBuy(market) {
        // 成行注文による即座の約定
        marketOrder := models.NewMarketOrder(
            generateOrderID(), 
            "EURUSD", 
            models.Buy, 
            10000.0,
        )
        broker.PlaceOrder(marketOrder)
        
        // 利食い用の指値注文を同時に設定
        takeProfitOrder := models.NewLimitOrder(
            generateOrderID(),
            "EURUSD",
            models.Sell,
            10000.0,
            currentPrice + 0.0050, // 50 pips利食い
        )
        broker.PlaceOrder(takeProfitOrder)
        
        // 損切り用の逆指値注文を設定
        stopLossOrder := models.NewStopOrder(
            generateOrderID(),
            "EURUSD", 
            models.Sell,
            10000.0,
            currentPrice - 0.0020, // 20 pips損切り
        )
        broker.PlaceOrder(stopLossOrder)
    }
    
    // 条件に基づく注文キャンセル
    pendingOrders := broker.GetPendingOrders()
    for _, order := range pendingOrders {
        if shouldCancelOrder(order, market) {
            broker.CancelOrder(order.ID)
        }
    }
    
    // 手動でのポジション決済判定
    positions := broker.GetPositions()
    for _, position := range positions {
        if shouldManualClose(position) {
            broker.ClosePosition(position.ID)
        }
    }
    
    // 統計情報の記録
    logAccountStatus(broker)
}

// 最終結果の表示
trades := broker.GetTradeHistory()
pendingOrders := broker.GetPendingOrders()
fmt.Printf("総取引数: %d\n", len(trades))
fmt.Printf("未約定注文数: %d\n", len(pendingOrders))
fmt.Printf("最終残高: %.2f\n", broker.GetBalance())
```

### 注文管理の例

```go
// 階段状の指値注文
basePrice := market.GetCurrentPrice("EURUSD")
for i := 0; i < 5; i++ {
    limitPrice := basePrice - float64(i)*0.0010 // 10 pipsずつ下の価格
    order := models.NewLimitOrder(
        fmt.Sprintf("limit-order-%d", i),
        "EURUSD",
        models.Buy,
        2000.0, // 小さなロットサイズ
        limitPrice,
    )
    broker.PlaceOrder(order)
}

// 一定時間後の注文キャンセル
time.Sleep(10 * time.Minute) // 実際のバックテストでは時間進行
pendingOrders := broker.GetPendingOrders()
for _, order := range pendingOrders {
    if time.Since(order.CreatedAt) > 1*time.Hour {
        broker.CancelOrder(order.ID)
    }
}
```

### エラーハンドリングの例

```go
// 残高不足の処理
order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 1000000.0)
if err := broker.PlaceOrder(order); err != nil {
    if strings.Contains(err.Error(), "insufficient balance") {
        log.Println("残高が不足しています")
        // ポジションサイズを調整するか、取引を見送る
    }
}

// 無効な指値価格の処理
limitOrder := models.NewLimitOrder("order-2", "EURUSD", models.Buy, 10000.0, -1.0)
if err := broker.PlaceOrder(limitOrder); err != nil {
    if strings.Contains(err.Error(), "invalid limit price") {
        log.Println("無効な指値価格が指定されました")
    }
}

// 存在しないポジションの処理
if err := broker.ClosePosition("invalid-id"); err != nil {
    if strings.Contains(err.Error(), "position not found") {
        log.Println("指定されたポジションが見つかりません")
    }
}

// 存在しない注文のキャンセル処理
if err := broker.CancelOrder("invalid-order-id"); err != nil {
    if strings.Contains(err.Error(), "order not found") {
        log.Println("指定された注文が見つかりません")
    }
}

// 約定済み注文のキャンセル処理
if err := broker.CancelOrder("executed-order-id"); err != nil {
    if strings.Contains(err.Error(), "order already executed") {
        log.Println("注文は既に約定済みです")
    }
}
```

## 設計上の考慮事項

### 1. レバレッジとリスク管理

**現在の実装：**
- 固定レバレッジ 1:100 を使用
- 証拠金は注文サイズの1%
- 含み損失による強制ロスカットは未実装

**将来の拡張：**
```go
type BrokerConfig struct {
    InitialBalance float64 `json:"initial_balance"`
    Spread         float64 `json:"spread"`
    Leverage       float64 `json:"leverage"`       // 追加
    MarginCall     float64 `json:"margin_call"`    // 追加
    StopOut        float64 `json:"stop_out"`       // 追加
}
```

### 2. スプレッドとスリッページ

**現在のアプローチ：**
- 固定スプレッドを使用
- 買い注文は Ask価格（現在価格 + スプレッド）で約定
- 売り注文は Bid価格（現在価格 - スプレッド）で約定
- スリッページは考慮していない

**現実的な考慮事項：**
- 時間帯や流動性によるスプレッドの変動
- 大口注文時のスリッページ
- ボラティリティによる価格ギャップ

### 3. 注文種別のサポート

**現在サポート：**
- 成行注文（Market Order）: 即座に約定
- 指値注文（Limit Order）: 指定価格での約定
- 逆指値注文（Stop Order）: 指定価格でのトリガー約定

**実装済みの注文構造：**
```go
type OrderType int

const (
    MarketOrder OrderType = iota
    LimitOrder           // 指値注文
    StopOrder            // 逆指値注文
)

type Order struct {
    ID          string    `json:"id"`
    Type        OrderType `json:"type"`
    Symbol      string    `json:"symbol"`
    Side        OrderSide `json:"side"`
    Size        float64   `json:"size"`
    LimitPrice  float64   `json:"limit_price,omitempty"`
    StopPrice   float64   `json:"stop_price,omitempty"`
    Status      OrderStatus `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
}
```

**将来の拡張候補：**
- ストップ指値注文（Stop Limit Order）
- OCO注文（One Cancels Other）
- 有効期限付き注文（Good Till Time）
- Fill or Kill注文
- Immediate or Cancel注文

```go
// 将来実装予定
const (
    StopLimitOrder OrderType = iota + 3
    OCOOrder
    GTTOrder
)

type AdvancedOrder struct {
    Order
    TimeInForce   string    `json:"time_in_force,omitempty"`
    ExpiryTime    time.Time `json:"expiry_time,omitempty"`
    LinkedOrderID string    `json:"linked_order_id,omitempty"` // OCO用
}
```

### 4. 複数通貨ペア対応

**現在の制限：**
- 単一通貨ペアのみサポート
- 証拠金計算も単一通貨ベース

**将来の拡張：**
```go
type MultiBroker struct {
    markets      map[string]market.Market
    positions    map[string]map[string]*models.Position  // symbol -> positions
    marginUsed   map[string]float64                      // symbol -> margin
    crossRates   map[string]float64                      // 通貨換算レート
}
```

## パフォーマンス考慮事項

### 1. ポジション・注文管理の効率化

**現在の実装：**
- ポジションをマップで管理（O(1)アクセス）
- 保留注文をマップで管理（O(1)アクセス）
- UpdatePositions()は全ポジションを順次更新（O(n)）
- ProcessPendingOrders()は全保留注文を順次チェック（O(m)）

**最適化案：**
- 大量ポジション時の差分更新
- 価格レベル別の注文インデックス構造
- 約定条件による注文の事前ソート
- バッチ処理による更新効率化

**注文処理の最適化：**
```go
// 価格レベル別の注文管理
type OrderBook struct {
    BuyLimits  map[float64][]*models.Order  // 価格レベル別の買い指値
    SellLimits map[float64][]*models.Order  // 価格レベル別の売り指値
    BuyStops   map[float64][]*models.Order  // 価格レベル別の買い逆指値
    SellStops  map[float64][]*models.Order  // 価格レベル別の売り逆指値
}
```

### 2. メモリ使用量

**現在の使用量：**
- ポジション: 1つあたり約200バイト
- 保留注文: 1つあたり約250バイト
- 取引履歴: 1つあたり約300バイト
- 10,000取引 + 1,000保留注文で約5.25MB

**最適化戦略：**
- 古い取引履歴の圧縮
- ポジション・注文構造体の最適化
- プール型オブジェクト管理
- 不要になった保留注文の自動削除

## エラーハンドリング

### 1. 注文実行エラー

**主なエラー種別：**
- `insufficient balance`: 証拠金不足（成行注文のみ）
- `invalid price`: 無効な市場価格
- `invalid symbol`: 無効なシンボル
- `invalid limit price`: 無効な指値価格
- `invalid stop price`: 無効な逆指値価格
- `invalid order type`: サポートされていない注文種別

**エラー対応策：**
```go
func (b *SimpleBroker) PlaceOrder(order *models.Order) error {
    // 基本バリデーション
    if order.Size <= 0 {
        return errors.New("order size must be positive")
    }
    
    if order.Symbol == "" {
        return errors.New("symbol is required")
    }
    
    // 注文種別別のバリデーション
    switch order.Type {
    case models.MarketOrder:
        // 成行注文の場合は価格と残高をチェック
        currentPrice := b.market.GetCurrentPrice(order.Symbol)
        if currentPrice <= 0.0 {
            return fmt.Errorf("invalid price for symbol %s", order.Symbol)
        }
        
        requiredMargin := b.calculateRequiredMargin(order, currentPrice)
        if b.balance < requiredMargin {
            return errors.New("insufficient balance")
        }
        
    case models.LimitOrder:
        // 指値注文の場合は指値価格をチェック
        if order.LimitPrice <= 0 {
            return errors.New("invalid limit price")
        }
        
    case models.StopOrder:
        // 逆指値注文の場合は逆指値価格をチェック
        if order.StopPrice <= 0 {
            return errors.New("invalid stop price")
        }
        
    default:
        return errors.New("invalid order type")
    }
    
    // 注文実行...
}
```

### 2. 注文管理エラー

**主なエラー種別：**
- `order not found`: 存在しない注文ID
- `order already executed`: 既に約定済み
- `order already cancelled`: 既にキャンセル済み
- `cannot cancel market order`: 成行注文はキャンセル不可

**エラー対応例：**
```go
func (b *SimpleBroker) CancelOrder(orderID string) error {
    order, exists := b.pendingOrders[orderID]
    if !exists {
        return fmt.Errorf("order not found: %s", orderID)
    }
    
    if order.Status == models.Executed {
        return fmt.Errorf("order already executed: %s", orderID)
    }
    
    if order.Status == models.Cancelled {
        return fmt.Errorf("order already cancelled: %s", orderID)
    }
    
    // キャンセル処理...
}
```

### 3. ポジション管理エラー

**主なエラー種別：**
- `position not found`: 存在しないポジション
- `position already closed`: 既に決済済み
- `invalid position state`: 無効な状態

### 4. リカバリー戦略

**エラー時の対応：**
- ログ出力による問題追跡
- 部分的な処理継続
- 一貫性チェックによる状態修復
- 注文とポジションの整合性チェック

## テスト戦略

### 1. ユニットテスト

**テストカバレッジ：**
- `broker_test.go`: 基本機能のテスト
- 正常系: 注文実行、ポジション管理、決済処理
- 異常系: エラーケース、境界値テスト

**主要テストケース：**
```go
// 注文実行テスト
func TestBroker_PlaceMarketOrder(t *testing.T)     // 成行注文テスト
func TestBroker_PlaceLimitOrder(t *testing.T)      // 指値注文テスト
func TestBroker_PlaceStopOrder(t *testing.T)       // 逆指値注文テスト

// 注文管理テスト
func TestBroker_CancelOrder(t *testing.T)          // 注文キャンセルテスト
func TestBroker_GetPendingOrders(t *testing.T)     // 保留注文取得テスト
func TestBroker_ProcessPendingOrders(t *testing.T) // 保留注文処理テスト

// ポジション管理テスト
func TestBroker_GetPositions(t *testing.T)         // ポジション取得テスト
func TestBroker_ClosePosition(t *testing.T)        // ポジション決済テスト
func TestBroker_UpdatePositions(t *testing.T)      // ポジション更新テスト

// その他
func TestBroker_GetBalance(t *testing.T)           // 残高管理テスト
func TestBroker_GetTradeHistory(t *testing.T)      // 取引履歴テスト
func TestBroker_ErrorHandling(t *testing.T)        // エラーハンドリングテスト
```

### 2. 統合テスト

**Market連携テスト：**
- 実際のCSVデータを使用した取引テスト
- 市場データ更新とポジション価格連動テスト
- 長時間実行での安定性テスト

### 3. パフォーマンステスト

**負荷テスト：**
- 大量ポジション（1000+）での性能測定
- メモリ使用量の監視
- 処理時間の測定

## 品質保証

### 1. コードカバレッジ

**目標カバレッジ：**
- 最低85%以上のカバレッジを維持
- 全エラーパスのテスト
- エッジケースの網羅

### 2. パフォーマンス目標

**処理速度：**
- 注文実行: 1秒間に10,000回以上
- 保留注文処理: 1秒間に5,000注文以上
- ポジション更新: 1秒間に5,000ポジション以上
- メモリ使用量: 1,000ポジション + 500保留注文で12MB以下

### 3. 精度保証

**計算精度：**
- 損益計算の小数点以下精度管理
- 累積誤差の防止
- 通貨換算精度の保証

## 拡張性

### 1. 新しいブローカータイプの追加

**ECNブローカー：**
```go
type ECNBroker struct {
    // 板情報ベースの約定
    orderBook map[string]*OrderBook
    // 可変スプレッド
    spreadProvider SpreadProvider
}
```

**デモブローカー：**
```go
type DemoBroker struct {
    // 実際のAPI接続
    apiClient ExternalBrokerAPI
    // リアルタイムデータ
    realTimeData DataFeed
}
```

### 2. リスク管理機能の追加

**マージンコール：**
```go
type RiskManager interface {
    CheckMarginCall(broker Broker) bool
    ExecuteStopOut(broker Broker) error
    CalculateMaxLotSize(broker Broker, symbol string) float64
}
```

**ポジションサイジング：**
```go
type PositionSizer interface {
    CalculatePositionSize(
        balance float64,
        riskPercentage float64,
        stopLoss float64,
    ) float64
}
```

### 3. 高度な注文機能

**一括決済：**
```go
func (b *SimpleBroker) CloseAllPositions() error
func (b *SimpleBroker) ClosePositionsBySymbol(symbol string) error
func (b *SimpleBroker) ClosePositionsBySide(side models.OrderSide) error
```

**条件付き注文：**
```go
type ConditionalOrder struct {
    Order     *models.Order
    Condition OrderCondition
    ExpiryTime time.Time
}
```

## 依存関係

### 直接依存

- `pkg/market`: Market インターフェース（価格データ取得）
- `pkg/models`: Order, Position, Trade データ構造
- Go標準ライブラリ: errors, fmt

### 間接依存

- 市場データ: CSV ファイルまたは外部データフィード
- 時間管理: バックテストエンジンによる時間進行

## 実装上の注意点

### 1. 並行処理

**現在の状況：**
- 単一ゴルーチンでの使用を想定
- 並行アクセスに対する保護は未実装

**将来の対応：**
```go
type SafeBroker struct {
    SimpleBroker
    mutex sync.RWMutex
}

func (b *SafeBroker) PlaceOrder(order *models.Order) error {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    return b.SimpleBroker.PlaceOrder(order)
}
```

### 2. 数値精度

**注意事項：**
- float64 の精度限界
- 累積誤差の可能性
- 通貨の最小単位（pip）の考慮

**対策：**
```go
import "github.com/shopspring/decimal"

type PreciseBroker struct {
    balance decimal.Decimal
    // decimal.Decimal を使用した高精度計算
}
```

### 3. 状態管理

**重要なポイント：**
- ポジションと残高の整合性維持
- 取引履歴の正確性
- エラー時の状態巻き戻し

## 設計哲学

### 1. シンプルさと拡張性のバランス

SimpleBrokerは基本的な機能に集中し、複雑な機能は将来の実装として明確に分離しています。これにより、理解しやすく保守しやすいコードベースを維持しつつ、必要に応じて機能を追加できる設計となっています。

### 2. 現実的なシミュレーション

スプレッド、レバレッジ、証拠金管理など、実際のFX取引の重要な要素を含めることで、バックテスト結果の信頼性を高めています。

### 3. テスタビリティ

各機能が独立してテスト可能な設計となっており、モックオブジェクトを使用した単体テストが容易に作成できます。