package broker

import (
	"errors"
	"fmt"

	"github.com/RuiHirano/fx-backtesting/pkg/market"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Broker はブローカー機能を提供するインターフェースです。
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

// SimpleBroker はBrokerインターフェースの高度な実装です。
type SimpleBroker struct {
	config        models.BrokerConfig
	market        market.Market
	balance       float64
	positions     map[string]*models.Position
	pendingOrders map[string]*models.Order
	tradeHistory  []*models.Trade
}

// NewSimpleBroker は新しいSimpleBrokerを作成します。
func NewSimpleBroker(config models.BrokerConfig, market market.Market) Broker {
	return &SimpleBroker{
		config:        config,
		market:        market,
		balance:       config.InitialBalance,
		positions:     make(map[string]*models.Position),
		pendingOrders: make(map[string]*models.Order),
		tradeHistory:  make([]*models.Trade, 0),
	}
}

// PlaceOrder は注文を受け付け、種別に応じて即座に実行または保留状態にします。
func (b *SimpleBroker) PlaceOrder(order *models.Order) error {
	// 注文のバリデーション
	if err := order.Validate(); err != nil {
		return err
	}

	// 注文種別に応じた処理
	switch order.Type {
	case models.MarketOrder:
		return b.executeMarketOrder(order)
	case models.LimitOrder, models.StopOrder:
		return b.addPendingOrder(order)
	default:
		return fmt.Errorf("unsupported order type: %v", order.Type)
	}
}

// executeMarketOrder は成行注文を即座に実行します。
func (b *SimpleBroker) executeMarketOrder(order *models.Order) error {
	// 現在価格を取得
	currentPrice := b.market.GetCurrentPrice()
	if currentPrice <= 0.0 {
		return fmt.Errorf("invalid price for symbol %s", order.Symbol)
	}

	// スプレッドを適用した実行価格を計算
	var executionPrice float64
	if order.Side == models.Buy {
		executionPrice = currentPrice + b.config.Spread // Ask価格
	} else {
		executionPrice = currentPrice - b.config.Spread // Bid価格
	}

	// 必要証拠金を計算（1:100レバレッジを想定）
	requiredMargin := (order.Size * executionPrice) / 100.0

	// 残高チェック
	if b.balance < requiredMargin {
		return errors.New("insufficient balance")
	}

	// ポジション作成
	position := &models.Position{
		ID:           fmt.Sprintf("pos-%s", order.ID),
		Symbol:       order.Symbol,
		Side:         order.Side,
		Size:         order.Size,
		EntryPrice:   executionPrice,
		CurrentPrice: currentPrice,
		OpenTime:     b.market.GetCurrentTime(),
	}

	// ポジション保存
	b.positions[position.ID] = position

	// 残高更新（証拠金を差し引く）
	b.balance -= requiredMargin

	// 注文を約定状態に更新
	order.Execute(executionPrice)

	return nil
}

// addPendingOrder は指値・逆指値注文を保留リストに追加します。
func (b *SimpleBroker) addPendingOrder(order *models.Order) error {
	// 保留注文として保存
	b.pendingOrders[order.ID] = order
	return nil
}

// CancelOrder は保留中の注文をキャンセルします。
func (b *SimpleBroker) CancelOrder(orderID string) error {
	order, exists := b.pendingOrders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.IsExecuted() {
		return fmt.Errorf("order already executed: %s", orderID)
	}

	if order.IsCancelled() {
		return fmt.Errorf("order already cancelled: %s", orderID)
	}

	// 注文をキャンセル状態に更新
	order.Cancel()
	
	// 保留注文リストから削除
	delete(b.pendingOrders, orderID)

	return nil
}

// GetPendingOrders は現在保留中の全注文を取得します。
func (b *SimpleBroker) GetPendingOrders() []*models.Order {
	orders := make([]*models.Order, 0, len(b.pendingOrders))
	for _, order := range b.pendingOrders {
		orders = append(orders, order)
	}
	return orders
}

// GetPositions は全ポジションを取得します。
func (b *SimpleBroker) GetPositions() []*models.Position {
	positions := make([]*models.Position, 0, len(b.positions))
	for _, position := range b.positions {
		positions = append(positions, position)
	}
	return positions
}

// GetBalance は現在の残高を取得します。
func (b *SimpleBroker) GetBalance() float64 {
	return b.balance
}

// ClosePosition はポジションをクローズします。
func (b *SimpleBroker) ClosePosition(positionID string) error {
	position, exists := b.positions[positionID]
	if !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	// 現在価格を取得
	currentPrice := b.market.GetCurrentPrice()
	if currentPrice <= 0.0 {
		return fmt.Errorf("invalid price for symbol %s", position.Symbol)
	}

	// スプレッドを適用したクローズ価格を計算
	var closePrice float64
	if position.Side == models.Buy {
		closePrice = currentPrice - b.config.Spread // Bid価格で売却
	} else {
		closePrice = currentPrice + b.config.Spread // Ask価格で買戻し
	}

	// 損益計算
	var pnl float64
	if position.Side == models.Buy {
		pnl = (closePrice - position.EntryPrice) * position.Size
	} else {
		pnl = (position.EntryPrice - closePrice) * position.Size
	}

	// 残高更新（証拠金を返却し、損益を反映）
	requiredMargin := (position.EntryPrice * position.Size) / 100.0
	b.balance += requiredMargin // 証拠金返却
	b.balance += pnl            // 損益反映

	// 取引履歴を作成して保存
	trade := models.NewTradeFromPosition(position, closePrice, pnl, b.market.GetCurrentTime())
	b.tradeHistory = append(b.tradeHistory, trade)

	// ポジション削除
	delete(b.positions, positionID)

	return nil
}

// GetTradeHistory は取引履歴を取得します。
func (b *SimpleBroker) GetTradeHistory() []*models.Trade {
	return b.tradeHistory
}

// UpdatePositions は全ポジションの現在価格を更新し、保留注文も処理します。
func (b *SimpleBroker) UpdatePositions() {
	// ポジション価格更新
	for _, position := range b.positions {
		currentPrice := b.market.GetCurrentPrice()
		if currentPrice > 0.0 {
			position.CurrentPrice = currentPrice
		}
	}
	
	// 保留注文の処理
	b.ProcessPendingOrders()
}

// ProcessPendingOrders は保留中の注文を現在の市場価格と照らし合わせて約定処理します。
func (b *SimpleBroker) ProcessPendingOrders() {
	executedOrders := make([]string, 0)
	
	for orderID, order := range b.pendingOrders {
		if !order.IsPending() {
			continue
		}
		
		currentPrice := b.market.GetCurrentPrice()
		if currentPrice <= 0.0 {
			continue
		}
		
		// 約定条件をチェック
		shouldExecute := false
		
		switch order.Type {
		case models.LimitOrder:
			if order.Side == models.Buy {
				// 買い指値: 現在価格が指値価格以下になった時に約定
				shouldExecute = currentPrice <= order.LimitPrice
			} else {
				// 売り指値: 現在価格が指値価格以上になった時に約定
				shouldExecute = currentPrice >= order.LimitPrice
			}
		case models.StopOrder:
			if order.Side == models.Buy {
				// 買い逆指値: 現在価格が逆指値価格以上になった時に約定
				shouldExecute = currentPrice >= order.StopPrice
			} else {
				// 売り逆指値: 現在価格が逆指値価格以下になった時に約定
				shouldExecute = currentPrice <= order.StopPrice
			}
		}
		
		if shouldExecute {
			if err := b.executePendingOrder(order, currentPrice); err == nil {
				executedOrders = append(executedOrders, orderID)
			}
		}
	}
	
	// 約定した注文を保留リストから削除
	for _, orderID := range executedOrders {
		delete(b.pendingOrders, orderID)
	}
}

// executePendingOrder は保留注文を約定させます。
func (b *SimpleBroker) executePendingOrder(order *models.Order, currentPrice float64) error {
	// スプレッドを適用した実行価格を計算
	var executionPrice float64
	if order.Side == models.Buy {
		executionPrice = currentPrice + b.config.Spread // Ask価格
	} else {
		executionPrice = currentPrice - b.config.Spread // Bid価格
	}
	
	// 必要証拠金を計算
	requiredMargin := (order.Size * executionPrice) / 100.0
	
	// 残高チェック
	if b.balance < requiredMargin {
		// 証拠金不足の場合は約定させない
		return errors.New("insufficient balance for pending order execution")
	}
	
	// ポジション作成
	position := &models.Position{
		ID:           fmt.Sprintf("pos-%s", order.ID),
		Symbol:       order.Symbol,
		Side:         order.Side,
		Size:         order.Size,
		EntryPrice:   executionPrice,
		CurrentPrice: currentPrice,
		OpenTime:     b.market.GetCurrentTime(),
	}
	
	// ポジション保存
	b.positions[position.ID] = position
	
	// 残高更新
	b.balance -= requiredMargin
	
	// 注文を約定状態に更新
	order.Execute(executionPrice)
	
	return nil
}