package models

import "time"

// OrderType represents the type of order
type OrderType int

const (
	OrderTypeMarket OrderType = iota
	OrderTypeLimit
	OrderTypeStop
)

// OrderSide represents the side of the order (buy or sell)
type OrderSide int

const (
	OrderSideBuy OrderSide = iota
	OrderSideSell
)

// Order represents a trading order
type Order struct {
	ID         string
	Symbol     string
	Type       OrderType
	Side       OrderSide
	Size       float64
	Price      float64
	StopLoss   float64
	TakeProfit float64
	Timestamp  time.Time
}

// NewOrder creates a new Order instance
func NewOrder(id, symbol string, orderType OrderType, side OrderSide, size, price, stopLoss, takeProfit float64, timestamp time.Time) Order {
	return Order{
		ID:         id,
		Symbol:     symbol,
		Type:       orderType,
		Side:       side,
		Size:       size,
		Price:      price,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
		Timestamp:  timestamp,
	}
}

// IsValid checks if the order is valid
func (o Order) IsValid() bool {
	if o.ID == "" || o.Symbol == "" {
		return false
	}
	
	if o.Size <= 0 {
		return false
	}
	
	// For limit orders, price must be positive
	if o.Type == OrderTypeLimit && o.Price <= 0 {
		return false
	}
	
	return true
}

// IsMarketOrder returns true if the order is a market order
func (o Order) IsMarketOrder() bool {
	return o.Type == OrderTypeMarket
}

// IsLimitOrder returns true if the order is a limit order
func (o Order) IsLimitOrder() bool {
	return o.Type == OrderTypeLimit
}

// IsBuyOrder returns true if the order is a buy order
func (o Order) IsBuyOrder() bool {
	return o.Side == OrderSideBuy
}

// IsSellOrder returns true if the order is a sell order
func (o Order) IsSellOrder() bool {
	return o.Side == OrderSideSell
}