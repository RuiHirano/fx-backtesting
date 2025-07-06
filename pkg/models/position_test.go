package models

import (
	"math"
	"testing"
	"time"
)

func TestNewPosition(t *testing.T) {
	timestamp := time.Now()
	position := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0510, timestamp)

	if position.ID != "pos1" {
		t.Errorf("Expected ID 'pos1', got %s", position.ID)
	}
	if position.Symbol != "EURUSD" {
		t.Errorf("Expected symbol 'EURUSD', got %s", position.Symbol)
	}
	if position.Side != OrderSideBuy {
		t.Errorf("Expected side %v, got %v", OrderSideBuy, position.Side)
	}
	if position.Size != 1000.0 {
		t.Errorf("Expected size 1000.0, got %v", position.Size)
	}
	if position.EntryPrice != 1.0500 {
		t.Errorf("Expected entry price 1.0500, got %v", position.EntryPrice)
	}
	if position.CurrentPrice != 1.0510 {
		t.Errorf("Expected current price 1.0510, got %v", position.CurrentPrice)
	}
	if !position.OpenTime.Equal(timestamp) {
		t.Errorf("Expected open time %v, got %v", timestamp, position.OpenTime)
	}
}

func TestPosition_UpdatePrice(t *testing.T) {
	position := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0500, time.Now())

	position.UpdatePrice(1.0520)

	if position.CurrentPrice != 1.0520 {
		t.Errorf("Expected current price 1.0520, got %v", position.CurrentPrice)
	}
}

func TestPosition_CalculatePnL(t *testing.T) {
	tests := []struct {
		name          string
		side          OrderSide
		size          float64
		entryPrice    float64
		currentPrice  float64
		expectedPnL   float64
	}{
		{
			name:         "buy position profit",
			side:         OrderSideBuy,
			size:         1000.0,
			entryPrice:   1.0500,
			currentPrice: 1.0520,
			expectedPnL:  2.0,
		},
		{
			name:         "buy position loss",
			side:         OrderSideBuy,
			size:         1000.0,
			entryPrice:   1.0500,
			currentPrice: 1.0480,
			expectedPnL:  -2.0,
		},
		{
			name:         "sell position profit",
			side:         OrderSideSell,
			size:         1000.0,
			entryPrice:   1.0500,
			currentPrice: 1.0480,
			expectedPnL:  2.0,
		},
		{
			name:         "sell position loss",
			side:         OrderSideSell,
			size:         1000.0,
			entryPrice:   1.0500,
			currentPrice: 1.0520,
			expectedPnL:  -2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			position := NewPosition("test", "EURUSD", tt.side, tt.size, tt.entryPrice, tt.currentPrice, time.Now())
			pnl := position.CalculatePnL()

			if math.Abs(pnl-tt.expectedPnL) > 1e-10 {
				t.Errorf("Expected PnL %v, got %v", tt.expectedPnL, pnl)
			}
		})
	}
}

func TestPosition_IsProfitable(t *testing.T) {
	profitablePosition := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0520, time.Now())
	lossPosition := NewPosition("pos2", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0480, time.Now())

	if !profitablePosition.IsProfitable() {
		t.Error("Expected profitable position to return true")
	}
	if lossPosition.IsProfitable() {
		t.Error("Expected loss position to return false")
	}
}

func TestPosition_IsLoss(t *testing.T) {
	profitablePosition := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0520, time.Now())
	lossPosition := NewPosition("pos2", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0480, time.Now())

	if profitablePosition.IsLoss() {
		t.Error("Expected profitable position to return false")
	}
	if !lossPosition.IsLoss() {
		t.Error("Expected loss position to return true")
	}
}

func TestPosition_IsLong(t *testing.T) {
	longPosition := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0520, time.Now())
	shortPosition := NewPosition("pos2", "EURUSD", OrderSideSell, 1000.0, 1.0500, 1.0480, time.Now())

	if !longPosition.IsLong() {
		t.Error("Expected long position to return true")
	}
	if shortPosition.IsLong() {
		t.Error("Expected short position to return false")
	}
}

func TestPosition_IsShort(t *testing.T) {
	longPosition := NewPosition("pos1", "EURUSD", OrderSideBuy, 1000.0, 1.0500, 1.0520, time.Now())
	shortPosition := NewPosition("pos2", "EURUSD", OrderSideSell, 1000.0, 1.0500, 1.0480, time.Now())

	if longPosition.IsShort() {
		t.Error("Expected long position to return false")
	}
	if !shortPosition.IsShort() {
		t.Error("Expected short position to return true")
	}
}