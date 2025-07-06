package broker

import (
	"fmt"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Broker defines the interface for trade execution
type Broker interface {
	PlaceOrder(order models.Order) error
	ClosePosition(positionID string) error
	GetBalance() float64
	GetPositions() []models.Position
}

// MockBroker implements Broker for backtesting
type MockBroker struct {
	config      models.Config
	balance     float64
	positions   map[string]models.Position
	marketPrices map[string]float64
	nextPositionID int
}

// NewMockBroker creates a new mock broker
func NewMockBroker(config models.Config) *MockBroker {
	return &MockBroker{
		config:      config,
		balance:     config.InitialBalance,
		positions:   make(map[string]models.Position),
		marketPrices: make(map[string]float64),
		nextPositionID: 1,
	}
}

// GetBalance returns the current account balance
func (b *MockBroker) GetBalance() float64 {
	return b.balance
}

// GetPositions returns all open positions
func (b *MockBroker) GetPositions() []models.Position {
	positions := make([]models.Position, 0, len(b.positions))
	for _, position := range b.positions {
		// Update position with current market price
		if currentPrice, exists := b.marketPrices[position.Symbol]; exists {
			position.UpdatePrice(currentPrice)
		}
		positions = append(positions, position)
	}
	return positions
}

// PlaceOrder executes a trading order
func (b *MockBroker) PlaceOrder(order models.Order) error {
	if !order.IsValid() {
		return fmt.Errorf("invalid order: %+v", order)
	}

	// Get current market price
	marketPrice, exists := b.marketPrices[order.Symbol]
	if !exists {
		return fmt.Errorf("no market price available for symbol: %s", order.Symbol)
	}

	// Apply spread to get execution price
	executionPrice := b.config.ApplySpread(marketPrice, order.Side)

	// Calculate required margin
	requiredMargin := b.config.CalculateMarginRequired(order.Size, executionPrice)

	// Check if sufficient balance for margin
	if b.balance < requiredMargin {
		return fmt.Errorf("insufficient margin: required %v, available %v", requiredMargin, b.balance)
	}

	// Calculate commission
	commission := b.config.CalculateCommission(order.Size, executionPrice)

	// Check if sufficient balance for commission
	if b.balance < requiredMargin + commission {
		return fmt.Errorf("insufficient balance for margin and commission: required %v, available %v", requiredMargin + commission, b.balance)
	}

	// Create position
	positionID := fmt.Sprintf("pos_%d", b.nextPositionID)
	b.nextPositionID++

	position := models.NewPosition(
		positionID,
		order.Symbol,
		order.Side,
		order.Size,
		executionPrice,
		marketPrice,
		order.Timestamp,
	)

	// Deduct margin and commission from balance
	b.balance -= requiredMargin + commission

	// Store position
	b.positions[positionID] = position

	return nil
}

// ClosePosition closes an open position
func (b *MockBroker) ClosePosition(positionID string) error {
	position, exists := b.positions[positionID]
	if !exists {
		return fmt.Errorf("position not found: %s", positionID)
	}

	// Get current market price
	marketPrice, exists := b.marketPrices[position.Symbol]
	if !exists {
		return fmt.Errorf("no market price available for symbol: %s", position.Symbol)
	}

	// Apply spread for closing (opposite side)
	var closingSide models.OrderSide
	if position.Side == models.OrderSideBuy {
		closingSide = models.OrderSideSell
	} else {
		closingSide = models.OrderSideBuy
	}
	
	closingPrice := b.config.ApplySpread(marketPrice, closingSide)

	// Calculate PnL with closing price
	var pnl float64
	if position.Side == models.OrderSideBuy {
		pnl = (closingPrice - position.EntryPrice) * position.Size
	} else {
		pnl = (position.EntryPrice - closingPrice) * position.Size
	}

	// Calculate commission for closing
	commission := b.config.CalculateCommission(position.Size, closingPrice)

	// Return margin to balance and apply PnL and commission
	returnedMargin := b.config.CalculateMarginRequired(position.Size, position.EntryPrice)
	b.balance += returnedMargin + pnl - commission

	// Remove position
	delete(b.positions, positionID)

	return nil
}

// SetCurrentPrice sets the current market price for a symbol (for testing/simulation)
func (b *MockBroker) SetCurrentPrice(symbol string, price float64) {
	b.marketPrices[symbol] = price
}