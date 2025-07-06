package backtester

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

func TestNewBacktester(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMockStrategy()

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	if backtester == nil {
		t.Fatal("Expected backtester to be created")
	}

	if backtester.GetConfig().InitialBalance != config.InitialBalance {
		t.Errorf("Expected initial balance %v, got %v", config.InitialBalance, backtester.GetConfig().InitialBalance)
	}
}

func TestBacktester_Run_EmptyData(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMockStrategy()

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Run with empty data
	result, err := backtester.Run([]models.Candle{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
	if result != nil {
		t.Error("Expected nil result for empty data")
	}
}

func TestBacktester_Run_SimpleStrategy(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	mockStrategy := strategy.NewMockStrategy()
	var strategyInstance strategy.Strategy = mockStrategy

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Create test data
	candles := []models.Candle{
		models.NewCandle(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Date(2024, 1, 1, 9, 1, 0, 0, time.UTC), 1.0510, 1.0530, 1.0500, 1.0520, 1200),
		models.NewCandle(time.Date(2024, 1, 1, 9, 2, 0, 0, time.UTC), 1.0520, 1.0540, 1.0510, 1.0530, 800),
	}

	result, err := backtester.Run(candles)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	// Check that strategy received all ticks
	if mockStrategy.GetTickCount() != len(candles) {
		t.Errorf("Expected %d ticks, got %d", len(candles), mockStrategy.GetTickCount())
	}

	// Check basic result fields
	if result.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
	if result.EndTime.IsZero() {
		t.Error("Expected end time to be set")
	}
	if result.InitialBalance != config.InitialBalance {
		t.Errorf("Expected initial balance %v, got %v", config.InitialBalance, result.InitialBalance)
	}
}

func TestBacktester_Run_WithTradingStrategy(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	
	// Use a moving average strategy that will generate trades
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 2, 3, 1000.0)

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Create test data with trend that will trigger trades
	candles := []models.Candle{
		models.NewCandle(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Date(2024, 1, 1, 9, 1, 0, 0, time.UTC), 1.0510, 1.0530, 1.0500, 1.0520, 1200),
		models.NewCandle(time.Date(2024, 1, 1, 9, 2, 0, 0, time.UTC), 1.0520, 1.0540, 1.0510, 1.0530, 800),
		models.NewCandle(time.Date(2024, 1, 1, 9, 3, 0, 0, time.UTC), 1.0530, 1.0550, 1.0520, 1.0540, 900),
		models.NewCandle(time.Date(2024, 1, 1, 9, 4, 0, 0, time.UTC), 1.0540, 1.0560, 1.0530, 1.0550, 1100),
	}

	result, err := backtester.Run(candles)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Debug: check final positions and balance
	positions := brokerInstance.GetPositions()
	t.Logf("Final positions: %d", len(positions))
	t.Logf("Final balance: %v", result.FinalBalance)
	t.Logf("Total trades: %d", result.TotalTrades)

	// Check that some trading activity occurred
	// For MA strategy, we expect positions to be opened even if not all are closed
	if len(positions) == 0 && result.TotalTrades == 0 {
		t.Error("Expected at least some trading activity (positions or completed trades)")
	}

	// Check final balance
	if result.FinalBalance <= 0 {
		t.Error("Expected positive final balance")
	}

	// Check duration
	expectedDuration := candles[len(candles)-1].Timestamp.Sub(candles[0].Timestamp)
	if result.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, result.Duration)
	}
}

func TestBacktester_GetProgress(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMockStrategy()

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Initially no progress
	progress := backtester.GetProgress()
	if progress.ProcessedCandles != 0 {
		t.Errorf("Expected 0 processed candles, got %d", progress.ProcessedCandles)
	}
	if progress.TotalCandles != 0 {
		t.Errorf("Expected 0 total candles, got %d", progress.TotalCandles)
	}
	if progress.Percentage != 0.0 {
		t.Errorf("Expected 0%% progress, got %v", progress.Percentage)
	}
}

func TestBacktester_Reset(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	mockStrategy := strategy.NewMockStrategy()
	var strategyInstance strategy.Strategy = mockStrategy

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Run a simple backtest
	candles := []models.Candle{
		models.NewCandle(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
	}
	backtester.Run(candles)

	// Reset
	backtester.Reset()

	// Check that state was reset
	progress := backtester.GetProgress()
	if progress.ProcessedCandles != 0 {
		t.Errorf("Expected 0 processed candles after reset, got %d", progress.ProcessedCandles)
	}

	// Check that strategy was reset
	if mockStrategy.GetTickCount() != 0 {
		t.Errorf("Expected 0 tick count after reset, got %d", mockStrategy.GetTickCount())
	}
}

func TestBacktester_RunWithCallback(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMockStrategy()

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Create test data
	candles := []models.Candle{
		models.NewCandle(time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC), 1.0500, 1.0520, 1.0490, 1.0510, 1000),
		models.NewCandle(time.Date(2024, 1, 1, 9, 1, 0, 0, time.UTC), 1.0510, 1.0530, 1.0500, 1.0520, 1200),
	}

	callbackCalls := 0
	progressCallback := func(progress Progress) {
		callbackCalls++
		if progress.Percentage < 0 || progress.Percentage > 100 {
			t.Errorf("Invalid progress percentage: %v", progress.Percentage)
		}
	}

	result, err := backtester.RunWithCallback(candles, progressCallback)
	if err != nil {
		t.Fatalf("RunWithCallback failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	if callbackCalls == 0 {
		t.Error("Expected progress callback to be called")
	}
}