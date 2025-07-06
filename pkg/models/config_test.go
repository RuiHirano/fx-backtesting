package models

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig(10000.0, 0.0001, 0.0, 0.0, 100.0)

	if config.InitialBalance != 10000.0 {
		t.Errorf("Expected initial balance 10000.0, got %v", config.InitialBalance)
	}
	if config.Spread != 0.0001 {
		t.Errorf("Expected spread 0.0001, got %v", config.Spread)
	}
	if config.Commission != 0.0 {
		t.Errorf("Expected commission 0.0, got %v", config.Commission)
	}
	if config.Slippage != 0.0 {
		t.Errorf("Expected slippage 0.0, got %v", config.Slippage)
	}
	if config.Leverage != 100.0 {
		t.Errorf("Expected leverage 100.0, got %v", config.Leverage)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.InitialBalance <= 0 {
		t.Error("Default initial balance should be positive")
	}
	if config.Spread < 0 {
		t.Error("Default spread should be non-negative")
	}
	if config.Commission < 0 {
		t.Error("Default commission should be non-negative")
	}
	if config.Slippage < 0 {
		t.Error("Default slippage should be non-negative")
	}
	if config.Leverage <= 0 {
		t.Error("Default leverage should be positive")
	}
}

func TestConfig_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "valid config",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         0.0001,
				Commission:     0.0,
				Slippage:       0.0,
				Leverage:       100.0,
			},
			expected: true,
		},
		{
			name: "invalid - negative initial balance",
			config: Config{
				InitialBalance: -1000.0,
				Spread:         0.0001,
				Commission:     0.0,
				Slippage:       0.0,
				Leverage:       100.0,
			},
			expected: false,
		},
		{
			name: "invalid - negative spread",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         -0.0001,
				Commission:     0.0,
				Slippage:       0.0,
				Leverage:       100.0,
			},
			expected: false,
		},
		{
			name: "invalid - negative commission",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         0.0001,
				Commission:     -0.01,
				Slippage:       0.0,
				Leverage:       100.0,
			},
			expected: false,
		},
		{
			name: "invalid - negative slippage",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         0.0001,
				Commission:     0.0,
				Slippage:       -0.01,
				Leverage:       100.0,
			},
			expected: false,
		},
		{
			name: "invalid - zero leverage",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         0.0001,
				Commission:     0.0,
				Slippage:       0.0,
				Leverage:       0.0,
			},
			expected: false,
		},
		{
			name: "invalid - negative leverage",
			config: Config{
				InitialBalance: 10000.0,
				Spread:         0.0001,
				Commission:     0.0,
				Slippage:       0.0,
				Leverage:       -100.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsValid(); got != tt.expected {
				t.Errorf("Config.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConfig_CalculateMarginRequired(t *testing.T) {
	config := Config{
		Leverage: 100.0,
	}

	tests := []struct {
		name           string
		positionSize   float64
		entryPrice     float64
		expectedMargin float64
	}{
		{
			name:           "standard position",
			positionSize:   1000.0,
			entryPrice:     1.0500,
			expectedMargin: 10.5, // (1000 * 1.0500) / 100
		},
		{
			name:           "larger position",
			positionSize:   10000.0,
			entryPrice:     1.2000,
			expectedMargin: 120.0, // (10000 * 1.2000) / 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			margin := config.CalculateMarginRequired(tt.positionSize, tt.entryPrice)
			if margin != tt.expectedMargin {
				t.Errorf("Expected margin %v, got %v", tt.expectedMargin, margin)
			}
		})
	}
}

func TestConfig_CalculateCommission(t *testing.T) {
	config := Config{
		Commission: 0.01, // 1% commission
	}

	positionSize := 1000.0
	entryPrice := 1.0500
	expectedCommission := 10.5 // 1000 * 1.0500 * 0.01

	commission := config.CalculateCommission(positionSize, entryPrice)
	if commission != expectedCommission {
		t.Errorf("Expected commission %v, got %v", expectedCommission, commission)
	}
}

func TestConfig_ApplySpread(t *testing.T) {
	config := Config{
		Spread: 0.0002, // 2 pip spread
	}

	midPrice := 1.0500

	// For buy orders, price should be higher (ask price)
	buyPrice := config.ApplySpread(midPrice, OrderSideBuy)
	expectedBuyPrice := 1.0501 // 1.0500 + 0.0002/2
	if buyPrice != expectedBuyPrice {
		t.Errorf("Expected buy price %v, got %v", expectedBuyPrice, buyPrice)
	}

	// For sell orders, price should be lower (bid price)
	sellPrice := config.ApplySpread(midPrice, OrderSideSell)
	expectedSellPrice := 1.0499 // 1.0500 - 0.0002/2
	if sellPrice != expectedSellPrice {
		t.Errorf("Expected sell price %v, got %v", expectedSellPrice, sellPrice)
	}
}