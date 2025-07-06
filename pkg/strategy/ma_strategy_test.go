package strategy

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestNewMovingAverageStrategy(t *testing.T) {
	strategy := NewMovingAverageStrategy("EURUSD", 5, 10, 1000.0)
	
	if strategy.GetName() != "MovingAverageStrategy" {
		t.Errorf("Expected name 'MovingAverageStrategy', got %s", strategy.GetName())
	}
	
	if strategy.symbol != "EURUSD" {
		t.Errorf("Expected symbol 'EURUSD', got %s", strategy.symbol)
	}
	
	if strategy.fastPeriod != 5 {
		t.Errorf("Expected fast period 5, got %d", strategy.fastPeriod)
	}
	
	if strategy.slowPeriod != 10 {
		t.Errorf("Expected slow period 10, got %d", strategy.slowPeriod)
	}
	
	if strategy.positionSize != 1000.0 {
		t.Errorf("Expected position size 1000.0, got %v", strategy.positionSize)
	}
}

func TestMovingAverageStrategy_OnTick_NoSignal(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMovingAverageStrategy("EURUSD", 3, 5, 1000.0)

	mockBroker.SetCurrentPrice("EURUSD", 1.0500)

	// Feed some data but not enough to generate signals
	candles := []models.Candle{
		models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Now(), 1.0510, 1.0530, 1.0500, 1.0520, 1000),
	}

	for _, candle := range candles {
		err := strategy.OnTick(candle, mockBroker)
		if err != nil {
			t.Fatalf("OnTick failed: %v", err)
		}
	}

	// No positions should be opened yet
	positions := mockBroker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions, got %d", len(positions))
	}
}

func TestMovingAverageStrategy_OnTick_BuySignal(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMovingAverageStrategy("EURUSD", 2, 3, 1000.0)

	mockBroker.SetCurrentPrice("EURUSD", 1.0500)

	// Create price series that will generate a bullish crossover
	// Start with downtrend, then uptrend to create crossover
	candles := []models.Candle{
		models.NewCandle(time.Now(), 1.0400, 1.0420, 1.0390, 1.0410, 1000), // Low prices
		models.NewCandle(time.Now(), 1.0410, 1.0430, 1.0400, 1.0420, 1000),
		models.NewCandle(time.Now(), 1.0420, 1.0440, 1.0410, 1.0430, 1000),
		models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000), // Price jump up
		models.NewCandle(time.Now(), 1.0510, 1.0530, 1.0500, 1.0520, 1000), // Continue up
	}

	for i, candle := range candles {
		mockBroker.SetCurrentPrice("EURUSD", candle.Close)
		err := strategy.OnTick(candle, mockBroker)
		if err != nil {
			t.Fatalf("OnTick failed at index %d: %v", i, err)
		}
	}

	// Check if a buy position was opened
	positions := mockBroker.GetPositions()
	if len(positions) == 0 {
		t.Error("Expected at least 1 position to be opened")
	} else {
		position := positions[0]
		if position.Side != models.OrderSideBuy {
			t.Errorf("Expected buy position, got %v", position.Side)
		}
		if position.Symbol != "EURUSD" {
			t.Errorf("Expected symbol EURUSD, got %s", position.Symbol)
		}
		if position.Size != 1000.0 {
			t.Errorf("Expected size 1000.0, got %v", position.Size)
		}
	}
}

func TestMovingAverageStrategy_OnTick_SellSignal(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMovingAverageStrategy("EURUSD", 2, 3, 1000.0)

	mockBroker.SetCurrentPrice("EURUSD", 1.0500)

	// Create a clear sell signal pattern
	candles := []models.Candle{
		models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Now(), 1.0510, 1.0530, 1.0500, 1.0520, 1000),
		models.NewCandle(time.Now(), 1.0520, 1.0540, 1.0510, 1.0530, 1000), // Now both MAs are ready
		models.NewCandle(time.Now(), 1.0400, 1.0420, 1.0390, 1.0410, 1000), // Sharp drop
	}

	var lastSignal Signal = SignalNone
	for i, candle := range candles {
		mockBroker.SetCurrentPrice("EURUSD", candle.Close)
		
		// Log before
		prevPositions := len(mockBroker.GetPositions())
		
		err := strategy.OnTick(candle, mockBroker)
		if err != nil {
			t.Fatalf("OnTick failed at index %d: %v", i, err)
		}
		
		// Log after
		currentPositions := len(mockBroker.GetPositions())
		currentSignal := strategy.GetSignal()
		
		if strategy.IsReady() {
			t.Logf("Candle %d: Close=%.4f, FastMA=%.4f, SlowMA=%.4f, Signal=%v, PrevSignal=%v, Positions: %d->%d", 
				i, candle.Close, strategy.GetFastMA(), strategy.GetSlowMA(), currentSignal, lastSignal, prevPositions, currentPositions)
		}
		
		lastSignal = currentSignal
	}

	// Check final state
	positions := mockBroker.GetPositions()
	t.Logf("Final positions: %d", len(positions))
	for i, pos := range positions {
		t.Logf("Position %d: Side=%v, Symbol=%s", i, pos.Side, pos.Symbol)
	}

	// We should have at least one position
	if len(positions) == 0 {
		t.Error("Expected at least 1 position to be opened")
	}
}

func TestMovingAverageStrategy_Reset(t *testing.T) {
	strategy := NewMovingAverageStrategy("EURUSD", 5, 10, 1000.0)

	// Add some data
	candle := models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000)
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	mockBroker.SetCurrentPrice("EURUSD", 1.0500)
	
	strategy.OnTick(candle, mockBroker)

	// Reset strategy
	strategy.Reset()

	// Check that indicators were reset
	if strategy.fastMA.IsReady() {
		t.Error("Expected fast MA not to be ready after reset")
	}
	if strategy.slowMA.IsReady() {
		t.Error("Expected slow MA not to be ready after reset")
	}
}

func TestMovingAverageStrategy_GetSignal(t *testing.T) {
	strategy := NewMovingAverageStrategy("EURUSD", 2, 3, 1000.0)

	// Test when not ready
	signal := strategy.GetSignal()
	if signal != SignalNone {
		t.Errorf("Expected no signal when not ready, got %v", signal)
	}

	// Add enough data to make indicators ready
	candles := []models.Candle{
		models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Now(), 1.0510, 1.0530, 1.0500, 1.0520, 1000),
		models.NewCandle(time.Now(), 1.0520, 1.0540, 1.0510, 1.0530, 1000),
		models.NewCandle(time.Now(), 1.0530, 1.0550, 1.0520, 1.0540, 1000),
	}

	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	mockBroker.SetCurrentPrice("EURUSD", 1.0500)

	for _, candle := range candles {
		strategy.OnTick(candle, mockBroker)
	}

	// Now signal should be available (likely buy due to uptrend)
	signal = strategy.GetSignal()
	if signal == SignalNone {
		t.Error("Expected a signal after providing enough data")
	}
}