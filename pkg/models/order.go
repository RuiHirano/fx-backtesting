package models

import (
	"errors"
	"strings"
	"time"
)

// OrderType は注文タイプを表します。
type OrderType int

const (
	MarketOrder OrderType = iota // 成行注文
	LimitOrder                   // 指値注文
	StopOrder                    // 逆指値注文
)

// String はOrderTypeの文字列表現を返します。
func (ot OrderType) String() string {
	switch ot {
	case MarketOrder:
		return "Market"
	case LimitOrder:
		return "Limit"
	case StopOrder:
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

// OrderStatus は注文状態を表します。
type OrderStatus int

const (
	Pending OrderStatus = iota // 保留中
	Executed                   // 約定済み
	Cancelled                  // キャンセル済み
	Rejected                   // 拒否済み
)

// String はOrderStatusの文字列表現を返します。
func (os OrderStatus) String() string {
	switch os {
	case Pending:
		return "Pending"
	case Executed:
		return "Executed"
	case Cancelled:
		return "Cancelled"
	case Rejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}

// Order は取引注文を表します。
type Order struct {
	ID          string      `json:"id"`
	Type        OrderType   `json:"type"`
	Symbol      string      `json:"symbol"`
	Side        OrderSide   `json:"side"`
	Size        float64     `json:"size"`
	LimitPrice  float64     `json:"limit_price,omitempty"`
	StopPrice   float64     `json:"stop_price,omitempty"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	ExecutedAt  time.Time   `json:"executed_at,omitempty"`
	ExecutedPrice float64   `json:"executed_price,omitempty"`
}

// NewMarketOrder は成行注文を作成します。
func NewMarketOrder(id, symbol string, side OrderSide, size float64) *Order {
	return &Order{
		ID:        id,
		Type:      MarketOrder,
		Symbol:    symbol,
		Side:      side,
		Size:      size,
		Status:    Pending,
		CreatedAt: time.Now(),
	}
}

// NewLimitOrder は指値注文を作成します。
func NewLimitOrder(id, symbol string, side OrderSide, size, limitPrice float64) *Order {
	return &Order{
		ID:         id,
		Type:       LimitOrder,
		Symbol:     symbol,
		Side:       side,
		Size:       size,
		LimitPrice: limitPrice,
		Status:     Pending,
		CreatedAt:  time.Now(),
	}
}

// NewStopOrder は逆指値注文を作成します。
func NewStopOrder(id, symbol string, side OrderSide, size, stopPrice float64) *Order {
	return &Order{
		ID:        id,
		Type:      StopOrder,
		Symbol:    symbol,
		Side:      side,
		Size:      size,
		StopPrice: stopPrice,
		Status:    Pending,
		CreatedAt: time.Now(),
	}
}

// Validate は注文データの妥当性を検証します。
func (o *Order) Validate() error {
	if o.Size <= 0 {
		return errors.New("order size must be positive")
	}
	
	if strings.TrimSpace(o.Symbol) == "" {
		return errors.New("symbol is required")
	}
	
	switch o.Type {
	case LimitOrder:
		if o.LimitPrice <= 0 {
			return errors.New("limit order must have positive limit price")
		}
	case StopOrder:
		if o.StopPrice <= 0 {
			return errors.New("stop order must have positive stop price")
		}
	}
	
	return nil
}

// IsMarket は成行注文かどうかを判定します。
func (o *Order) IsMarket() bool {
	return o.Type == MarketOrder
}

// IsLimit は指値注文かどうかを判定します。
func (o *Order) IsLimit() bool {
	return o.Type == LimitOrder
}

// IsStop は逆指値注文かどうかを判定します。
func (o *Order) IsStop() bool {
	return o.Type == StopOrder
}

// IsPending は保留中の注文かどうかを判定します。
func (o *Order) IsPending() bool {
	return o.Status == Pending
}

// IsExecuted は約定済みの注文かどうかを判定します。
func (o *Order) IsExecuted() bool {
	return o.Status == Executed
}

// IsCancelled はキャンセル済みの注文かどうかを判定します。
func (o *Order) IsCancelled() bool {
	return o.Status == Cancelled
}

// Execute は注文を約定状態にします。
func (o *Order) Execute(executedPrice float64) {
	o.Status = Executed
	o.ExecutedAt = time.Now()
	o.ExecutedPrice = executedPrice
}

// Cancel は注文をキャンセル状態にします。
func (o *Order) Cancel() {
	o.Status = Cancelled
}