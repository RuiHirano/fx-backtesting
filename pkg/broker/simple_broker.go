package broker

import (
	"fmt"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// SimpleBroker implements advanced broker features like stop loss, take profit, and risk management
type SimpleBroker struct {
	*MockBroker
	pendingOrders   map[string]models.Order
	positionOrders  map[string]models.Order // Maps position ID to original order with SL/TP
	nextOrderID     int
}

// NewSimpleBroker creates a new simple broker with advanced features
func NewSimpleBroker(config models.Config) *SimpleBroker {
	return &SimpleBroker{
		MockBroker:     NewMockBroker(config),
		pendingOrders:  make(map[string]models.Order),
		positionOrders: make(map[string]models.Order),
		nextOrderID:    1,
	}
}

// PlaceOrder places an order with support for limit orders and SL/TP
func (b *SimpleBroker) PlaceOrder(order models.Order) error {
	if !order.IsValid() {
		return fmt.Errorf("invalid order: %+v", order)
	}

	// For limit orders, add to pending orders
	if order.Type == models.OrderTypeLimit {
		orderID := fmt.Sprintf("order_%d", b.nextOrderID)
		b.nextOrderID++
		order.ID = orderID
		b.pendingOrders[orderID] = order
		return nil
	}

	// For market orders, execute immediately
	err := b.MockBroker.PlaceOrder(order)
	if err != nil {
		return err
	}

	// Store order for SL/TP tracking if position was created
	if order.StopLoss > 0 || order.TakeProfit > 0 {
		positions := b.GetPositions()
		if len(positions) > 0 {
			// Find the latest position (assumes it's the one just created)
			latestPosition := positions[len(positions)-1]
			b.positionOrders[latestPosition.ID] = order
		}
	}

	return nil
}

// GetPendingOrders returns all pending orders
func (b *SimpleBroker) GetPendingOrders() []models.Order {
	orders := make([]models.Order, 0, len(b.pendingOrders))
	for _, order := range b.pendingOrders {
		orders = append(orders, order)
	}
	return orders
}

// ProcessPendingOrders checks and executes pending limit orders
func (b *SimpleBroker) ProcessPendingOrders() {
	for orderID, order := range b.pendingOrders {
		marketPrice, exists := b.marketPrices[order.Symbol]
		if !exists {
			continue
		}

		shouldExecute := false

		if order.Side == models.OrderSideBuy && marketPrice <= order.Price {
			// Buy limit order triggered when price drops to or below limit price
			shouldExecute = true
		} else if order.Side == models.OrderSideSell && marketPrice >= order.Price {
			// Sell limit order triggered when price rises to or above limit price
			shouldExecute = true
		}

		if shouldExecute {
			// Convert to market order and execute
			marketOrder := order
			marketOrder.Type = models.OrderTypeMarket
			err := b.MockBroker.PlaceOrder(marketOrder)
			if err == nil {
				// Store for SL/TP tracking
				if order.StopLoss > 0 || order.TakeProfit > 0 {
					positions := b.GetPositions()
					if len(positions) > 0 {
						latestPosition := positions[len(positions)-1]
						b.positionOrders[latestPosition.ID] = order
					}
				}
				// Remove from pending orders
				delete(b.pendingOrders, orderID)
			}
		}
	}
}

// ProcessStopLosses checks and triggers stop loss orders
func (b *SimpleBroker) ProcessStopLosses() {
	for positionID, order := range b.positionOrders {
		if order.StopLoss <= 0 {
			continue
		}

		position, exists := b.positions[positionID]
		if !exists {
			delete(b.positionOrders, positionID)
			continue
		}

		marketPrice, exists := b.marketPrices[position.Symbol]
		if !exists {
			continue
		}

		shouldTrigger := false

		if position.Side == models.OrderSideBuy && marketPrice <= order.StopLoss {
			// Long position stop loss triggered when price drops to or below stop loss
			shouldTrigger = true
		} else if position.Side == models.OrderSideSell && marketPrice >= order.StopLoss {
			// Short position stop loss triggered when price rises to or above stop loss
			shouldTrigger = true
		}

		if shouldTrigger {
			b.ClosePosition(positionID)
			delete(b.positionOrders, positionID)
		}
	}
}

// ProcessTakeProfits checks and triggers take profit orders
func (b *SimpleBroker) ProcessTakeProfits() {
	for positionID, order := range b.positionOrders {
		if order.TakeProfit <= 0 {
			continue
		}

		position, exists := b.positions[positionID]
		if !exists {
			delete(b.positionOrders, positionID)
			continue
		}

		marketPrice, exists := b.marketPrices[position.Symbol]
		if !exists {
			continue
		}

		shouldTrigger := false

		if position.Side == models.OrderSideBuy && marketPrice >= order.TakeProfit {
			// Long position take profit triggered when price rises to or above take profit
			shouldTrigger = true
		} else if position.Side == models.OrderSideSell && marketPrice <= order.TakeProfit {
			// Short position take profit triggered when price drops to or below take profit
			shouldTrigger = true
		}

		if shouldTrigger {
			b.ClosePosition(positionID)
			delete(b.positionOrders, positionID)
		}
	}
}

// GetEquity returns the account equity (balance + unrealized PnL)
func (b *SimpleBroker) GetEquity() float64 {
	equity := b.balance
	
	for _, position := range b.positions {
		if currentPrice, exists := b.marketPrices[position.Symbol]; exists {
			// Calculate unrealized PnL
			var unrealizedPnL float64
			if position.Side == models.OrderSideBuy {
				unrealizedPnL = (currentPrice - position.EntryPrice) * position.Size
			} else {
				unrealizedPnL = (position.EntryPrice - currentPrice) * position.Size
			}
			equity += unrealizedPnL
		}
	}
	
	return equity
}

// GetUsedMargin returns the total margin used by open positions
func (b *SimpleBroker) GetUsedMargin() float64 {
	usedMargin := 0.0
	
	for _, position := range b.positions {
		margin := b.config.CalculateMarginRequired(position.Size, position.EntryPrice)
		usedMargin += margin
	}
	
	return usedMargin
}

// GetFreeMargin returns the available margin for new positions
func (b *SimpleBroker) GetFreeMargin() float64 {
	return b.GetEquity() - b.GetUsedMargin()
}

// GetMarginLevel returns the margin level as a percentage
func (b *SimpleBroker) GetMarginLevel() float64 {
	usedMargin := b.GetUsedMargin()
	if usedMargin == 0 {
		return 0 // No positions open
	}
	return (b.GetEquity() / usedMargin) * 100
}

// ProcessMarginCalls closes positions if margin level is too low
func (b *SimpleBroker) ProcessMarginCalls() {
	marginLevel := b.GetMarginLevel()
	
	// Stop out when margin level is below 50% or negative
	if marginLevel <= 50 {
		// Close all positions
		positions := b.GetPositions()
		
		for _, position := range positions {
			b.ClosePosition(position.ID)
			delete(b.positionOrders, position.ID)
		}
	}
}

// ClosePosition overrides to clean up position orders
func (b *SimpleBroker) ClosePosition(positionID string) error {
	err := b.MockBroker.ClosePosition(positionID)
	if err == nil {
		delete(b.positionOrders, positionID)
	}
	return err
}