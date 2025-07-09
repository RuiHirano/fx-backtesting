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
	GetPositions() []*models.Position
	GetBalance() float64
	ClosePosition(positionID string) error
	UpdatePositions()
}

// SimpleBroker はBrokerインターフェースのシンプルな実装です。
type SimpleBroker struct {
	config    models.BrokerConfig
	market    market.Market
	balance   float64
	positions map[string]*models.Position
}

// NewSimpleBroker は新しいSimpleBrokerを作成します。
func NewSimpleBroker(config models.BrokerConfig, market market.Market) Broker {
	return &SimpleBroker{
		config:    config,
		market:    market,
		balance:   config.InitialBalance,
		positions: make(map[string]*models.Position),
	}
}

// PlaceOrder は注文を実行します。
func (b *SimpleBroker) PlaceOrder(order *models.Order) error {
	// 現在価格を取得
	currentPrice := b.market.GetCurrentPrice(order.Symbol)
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
	if order.Side == models.Buy && b.balance < requiredMargin {
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

	return nil
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
	currentPrice := b.market.GetCurrentPrice(position.Symbol)
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

	// ポジション削除
	delete(b.positions, positionID)

	return nil
}

// UpdatePositions は全ポジションの現在価格を更新します。
func (b *SimpleBroker) UpdatePositions() {
	for _, position := range b.positions {
		currentPrice := b.market.GetCurrentPrice(position.Symbol)
		if currentPrice > 0.0 {
			position.CurrentPrice = currentPrice
		}
	}
}