package models

import (
	"math"
	"testing"
)

const floatTolerance = 1e-9

func assertFloatEqual(t *testing.T, expected, actual float64, message string) {
	if math.Abs(expected-actual) > floatTolerance {
		t.Errorf("%s: expected %.9f, got %.9f", message, expected, actual)
	}
}

// Position構造体のテスト
func TestPosition_NewPosition(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	
	if position.ID != "pos-123" {
		t.Errorf("Expected ID pos-123, got %s", position.ID)
	}
	
	if position.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", position.Symbol)
	}
	
	if position.Side != Buy {
		t.Errorf("Expected side Buy, got %v", position.Side)
	}
	
	if position.Size != 10000.0 {
		t.Errorf("Expected size 10000.0, got %f", position.Size)
	}
	
	if position.EntryPrice != 1.1000 {
		t.Errorf("Expected entry price 1.1000, got %f", position.EntryPrice)
	}
	
	if position.CurrentPrice != 1.1000 {
		t.Errorf("Expected current price 1.1000, got %f", position.CurrentPrice)
	}
	
	if position.PnL != 0.0 {
		t.Errorf("Expected PnL 0.0, got %f", position.PnL)
	}
}

func TestPosition_UpdatePrice(t *testing.T) {
	// 買いポジションのテスト
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	
	// 価格上昇 -> 利益
	position.UpdatePrice(1.1010)
	if position.CurrentPrice != 1.1010 {
		t.Errorf("Expected current price 1.1010, got %f", position.CurrentPrice)
	}
	expectedPnL := (1.1010 - 1.1000) * 10000.0
	assertFloatEqual(t, expectedPnL, position.PnL, "Buy position PnL")
	
	// 売りポジションのテスト
	position = NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
	
	// 価格下落 -> 利益
	position.UpdatePrice(1.0990)
	expectedPnL = (1.1000 - 1.0990) * 10000.0
	assertFloatEqual(t, expectedPnL, position.PnL, "Sell position PnL")
}

func TestPosition_IsLong(t *testing.T) {
	buyPosition := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	if !buyPosition.IsLong() {
		t.Error("Expected buy position to be long")
	}
	
	sellPosition := NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
	if sellPosition.IsLong() {
		t.Error("Expected sell position to not be long")
	}
}

func TestPosition_IsShort(t *testing.T) {
	sellPosition := NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
	if !sellPosition.IsShort() {
		t.Error("Expected sell position to be short")
	}
	
	buyPosition := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	if buyPosition.IsShort() {
		t.Error("Expected buy position to not be short")
	}
}

func TestPosition_ShouldStopLoss(t *testing.T) {
	// 買いポジションのテスト
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	position.StopLoss = 1.0990
	
	// ストップロス未達
	position.UpdatePrice(1.1000)
	if position.ShouldStopLoss() {
		t.Error("Expected no stop loss trigger at entry price")
	}
	
	// ストップロス到達
	position.UpdatePrice(1.0990)
	if !position.ShouldStopLoss() {
		t.Error("Expected stop loss trigger at stop loss price")
	}
	
	// 売りポジションのテスト
	position = NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
	position.StopLoss = 1.1010
	
	// ストップロス到達
	position.UpdatePrice(1.1010)
	if !position.ShouldStopLoss() {
		t.Error("Expected stop loss trigger for sell position")
	}
}

func TestPosition_ShouldTakeProfit(t *testing.T) {
	// 買いポジションのテスト
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	position.TakeProfit = 1.1020
	
	// テイクプロフィット未達
	position.UpdatePrice(1.1010)
	if position.ShouldTakeProfit() {
		t.Error("Expected no take profit trigger below target")
	}
	
	// テイクプロフィット到達
	position.UpdatePrice(1.1020)
	if !position.ShouldTakeProfit() {
		t.Error("Expected take profit trigger at target price")
	}
}

func TestPosition_GetMarketValue(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	position.UpdatePrice(1.1010)
	
	expectedValue := 1.1010 * 10000.0
	if position.GetMarketValue() != expectedValue {
		t.Errorf("Expected market value %f, got %f", expectedValue, position.GetMarketValue())
	}
}

func TestPosition_GetPnLPercentage(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	position.UpdatePrice(1.1010)
	
	expectedPercentage := ((1.1010 - 1.1000) / 1.1000) * 100
	assertFloatEqual(t, expectedPercentage, position.GetPnLPercentage(), "PnL percentage")
}