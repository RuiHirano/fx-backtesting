package models

import "testing"

// Order構造体のテスト
func TestOrder_NewMarketOrder(t *testing.T) {
	order := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
	
	if order.ID != "test-123" {
		t.Errorf("Expected ID test-123, got %s", order.ID)
	}
	
	if order.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", order.Symbol)
	}
	
	if order.Type != Market {
		t.Errorf("Expected type Market, got %v", order.Type)
	}
	
	if order.Side != Buy {
		t.Errorf("Expected side Buy, got %v", order.Side)
	}
	
	if order.Size != 10000.0 {
		t.Errorf("Expected size 10000.0, got %f", order.Size)
	}
}

func TestOrder_NewLimitOrder(t *testing.T) {
	order := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
	
	if order.Type != Limit {
		t.Errorf("Expected type Limit, got %v", order.Type)
	}
	
	if order.Price != 1.1000 {
		t.Errorf("Expected price 1.1000, got %f", order.Price)
	}
}

func TestOrder_Validate(t *testing.T) {
	// 正常なケース - Market注文
	order := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
	if err := order.Validate(); err != nil {
		t.Errorf("Expected no error for valid market order, got %v", err)
	}
	
	// 正常なケース - Limit注文
	order = NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
	if err := order.Validate(); err != nil {
		t.Errorf("Expected no error for valid limit order, got %v", err)
	}
	
	// 異常なケース - 負のサイズ
	order = NewMarketOrder("test-123", "EURUSD", Buy, -10000.0)
	if err := order.Validate(); err == nil {
		t.Error("Expected error for negative size")
	}
	
	// 異常なケース - 空のシンボル
	order = NewMarketOrder("test-123", "", Buy, 10000.0)
	if err := order.Validate(); err == nil {
		t.Error("Expected error for empty symbol")
	}
	
	// 異常なケース - Limit注文で価格が0
	order = NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 0)
	if err := order.Validate(); err == nil {
		t.Error("Expected error for limit order with zero price")
	}
}

func TestOrder_IsMarket(t *testing.T) {
	marketOrder := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
	if !marketOrder.IsMarket() {
		t.Error("Expected market order to be identified as market")
	}
	
	limitOrder := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
	if limitOrder.IsMarket() {
		t.Error("Expected limit order to not be identified as market")
	}
}

func TestOrder_IsLimit(t *testing.T) {
	limitOrder := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
	if !limitOrder.IsLimit() {
		t.Error("Expected limit order to be identified as limit")
	}
	
	marketOrder := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
	if marketOrder.IsLimit() {
		t.Error("Expected market order to not be identified as limit")
	}
}

func TestOrderType_String(t *testing.T) {
	tests := []struct {
		orderType OrderType
		expected  string
	}{
		{Market, "Market"},
		{Limit, "Limit"},
		{Stop, "Stop"},
		{OrderType(999), "Unknown"},
	}
	
	for _, test := range tests {
		if test.orderType.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.orderType.String())
		}
	}
}

func TestOrderSide_String(t *testing.T) {
	tests := []struct {
		orderSide OrderSide
		expected  string
	}{
		{Buy, "Buy"},
		{Sell, "Sell"},
		{OrderSide(999), "Unknown"},
	}
	
	for _, test := range tests {
		if test.orderSide.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.orderSide.String())
		}
	}
}