package models

import (
	"errors"
	"strings"
	"time"
)

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
	Symbol     string    `json:"symbol"`
	Type       OrderType `json:"type"`
	Side       OrderSide `json:"side"`
	Size       float64   `json:"size"`
	Price      float64   `json:"price"`
	StopLoss   float64   `json:"stop_loss"`
	TakeProfit float64   `json:"take_profit"`
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
	
	if strings.TrimSpace(o.Symbol) == "" {
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