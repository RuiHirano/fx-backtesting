package models

import "time"

// Position represents an open trading position
type Position struct {
	ID           string
	Symbol       string
	Side         OrderSide
	Size         float64
	EntryPrice   float64
	CurrentPrice float64
	PnL          float64
	OpenTime     time.Time
}

// NewPosition creates a new Position instance
func NewPosition(id, symbol string, side OrderSide, size, entryPrice, currentPrice float64, openTime time.Time) Position {
	position := Position{
		ID:           id,
		Symbol:       symbol,
		Side:         side,
		Size:         size,
		EntryPrice:   entryPrice,
		CurrentPrice: currentPrice,
		OpenTime:     openTime,
	}
	position.PnL = position.CalculatePnL()
	return position
}

// UpdatePrice updates the current price and recalculates PnL
func (p *Position) UpdatePrice(currentPrice float64) {
	p.CurrentPrice = currentPrice
	p.PnL = p.CalculatePnL()
}

// CalculatePnL calculates the profit/loss of the position
func (p Position) CalculatePnL() float64 {
	if p.Side == OrderSideBuy {
		// Long position: profit when price goes up
		return (p.CurrentPrice - p.EntryPrice) * p.Size
	} else {
		// Short position: profit when price goes down
		return (p.EntryPrice - p.CurrentPrice) * p.Size
	}
}

// IsProfitable returns true if the position is profitable
func (p Position) IsProfitable() bool {
	return p.CalculatePnL() > 0
}

// IsLoss returns true if the position is at a loss
func (p Position) IsLoss() bool {
	return p.CalculatePnL() < 0
}

// IsLong returns true if the position is long (buy)
func (p Position) IsLong() bool {
	return p.Side == OrderSideBuy
}

// IsShort returns true if the position is short (sell)
func (p Position) IsShort() bool {
	return p.Side == OrderSideSell
}