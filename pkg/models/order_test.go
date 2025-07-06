package models

import (
	"testing"
	"time"
)

func TestNewOrder(t *testing.T) {
	timestamp := time.Now()
	order := NewOrder("order1", "EURUSD", OrderTypeMarket, OrderSideBuy, 1000.0, 1.0500, 1.0450, 1.0550, timestamp)

	if order.ID != "order1" {
		t.Errorf("Expected ID 'order1', got %s", order.ID)
	}
	if order.Symbol != "EURUSD" {
		t.Errorf("Expected symbol 'EURUSD', got %s", order.Symbol)
	}
	if order.Type != OrderTypeMarket {
		t.Errorf("Expected type %v, got %v", OrderTypeMarket, order.Type)
	}
	if order.Side != OrderSideBuy {
		t.Errorf("Expected side %v, got %v", OrderSideBuy, order.Side)
	}
	if order.Size != 1000.0 {
		t.Errorf("Expected size 1000.0, got %v", order.Size)
	}
	if order.Price != 1.0500 {
		t.Errorf("Expected price 1.0500, got %v", order.Price)
	}
	if order.StopLoss != 1.0450 {
		t.Errorf("Expected stop loss 1.0450, got %v", order.StopLoss)
	}
	if order.TakeProfit != 1.0550 {
		t.Errorf("Expected take profit 1.0550, got %v", order.TakeProfit)
	}
	if !order.Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, order.Timestamp)
	}
}

func TestOrder_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		order    Order
		expected bool
	}{
		{
			name: "valid market order",
			order: Order{
				ID:     "order1",
				Symbol: "EURUSD",
				Type:   OrderTypeMarket,
				Side:   OrderSideBuy,
				Size:   1000.0,
				Price:  1.0500,
			},
			expected: true,
		},
		{
			name: "invalid - empty ID",
			order: Order{
				ID:     "",
				Symbol: "EURUSD",
				Type:   OrderTypeMarket,
				Side:   OrderSideBuy,
				Size:   1000.0,
				Price:  1.0500,
			},
			expected: false,
		},
		{
			name: "invalid - empty symbol",
			order: Order{
				ID:     "order1",
				Symbol: "",
				Type:   OrderTypeMarket,
				Side:   OrderSideBuy,
				Size:   1000.0,
				Price:  1.0500,
			},
			expected: false,
		},
		{
			name: "invalid - zero size",
			order: Order{
				ID:     "order1",
				Symbol: "EURUSD",
				Type:   OrderTypeMarket,
				Side:   OrderSideBuy,
				Size:   0,
				Price:  1.0500,
			},
			expected: false,
		},
		{
			name: "invalid - negative price",
			order: Order{
				ID:     "order1",
				Symbol: "EURUSD",
				Type:   OrderTypeLimit,
				Side:   OrderSideBuy,
				Size:   1000.0,
				Price:  -1.0500,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.order.IsValid(); got != tt.expected {
				t.Errorf("Order.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestOrder_IsMarketOrder(t *testing.T) {
	marketOrder := Order{Type: OrderTypeMarket}
	limitOrder := Order{Type: OrderTypeLimit}

	if !marketOrder.IsMarketOrder() {
		t.Error("Expected market order to return true")
	}
	if limitOrder.IsMarketOrder() {
		t.Error("Expected limit order to return false")
	}
}

func TestOrder_IsLimitOrder(t *testing.T) {
	marketOrder := Order{Type: OrderTypeMarket}
	limitOrder := Order{Type: OrderTypeLimit}

	if marketOrder.IsLimitOrder() {
		t.Error("Expected market order to return false")
	}
	if !limitOrder.IsLimitOrder() {
		t.Error("Expected limit order to return true")
	}
}

func TestOrder_IsBuyOrder(t *testing.T) {
	buyOrder := Order{Side: OrderSideBuy}
	sellOrder := Order{Side: OrderSideSell}

	if !buyOrder.IsBuyOrder() {
		t.Error("Expected buy order to return true")
	}
	if sellOrder.IsBuyOrder() {
		t.Error("Expected sell order to return false")
	}
}

func TestOrder_IsSellOrder(t *testing.T) {
	buyOrder := Order{Side: OrderSideBuy}
	sellOrder := Order{Side: OrderSideSell}

	if buyOrder.IsSellOrder() {
		t.Error("Expected buy order to return false")
	}
	if !sellOrder.IsSellOrder() {
		t.Error("Expected sell order to return true")
	}
}