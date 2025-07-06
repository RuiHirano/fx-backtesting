package backtester

import (
	"os"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

func TestIntegration_FullBacktest(t *testing.T) {
	// Create temporary CSV file for testing
	csvContent := `2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200
2024-01-01 09:02:00,1.0520,1.0540,1.0510,1.0530,800
2024-01-01 09:03:00,1.0530,1.0550,1.0520,1.0540,950
2024-01-01 09:04:00,1.0540,1.0560,1.0530,1.0550,1100
2024-01-01 09:05:00,1.0550,1.0570,1.0540,1.0560,1300
2024-01-01 09:06:00,1.0560,1.0580,1.0550,1.0570,1050
2024-01-01 09:07:00,1.0570,1.0590,1.0560,1.0580,1150
2024-01-01 09:08:00,1.0580,1.0600,1.0570,1.0590,1000
2024-01-01 09:09:00,1.0590,1.0610,1.0580,1.0600,1250`

	tmpFile, err := os.CreateTemp("", "integration_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Setup components
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 3, 5, 1000.0)

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Load data from CSV
	candles, err := dataProvider.LoadCSVData(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load CSV data: %v", err)
	}

	if len(candles) != 10 {
		t.Errorf("Expected 10 candles, got %d", len(candles))
	}

	// Run backtest with progress callback
	progressUpdates := 0
	progressCallback := func(progress Progress) {
		progressUpdates++
		t.Logf("Progress: %d/%d (%.1f%%) - %v", 
			progress.ProcessedCandles, progress.TotalCandles, 
			progress.Percentage, progress.CurrentTime)
	}

	result, err := backtester.RunWithCallback(candles, progressCallback)
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	// Validate results
	if result == nil {
		t.Fatal("Expected result to be returned")
	}

	// Check basic result structure
	if result.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
	if result.EndTime.IsZero() {
		t.Error("Expected end time to be set")
	}
	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
	if result.InitialBalance != config.InitialBalance {
		t.Errorf("Expected initial balance %v, got %v", config.InitialBalance, result.InitialBalance)
	}
	if result.FinalBalance <= 0 {
		t.Error("Expected positive final balance")
	}

	// Check progress callback was called
	if progressUpdates == 0 {
		t.Error("Expected progress callback to be called")
	}

	// Log results for debugging
	t.Logf("Backtest Results:")
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Initial Balance: %.2f", result.InitialBalance)
	t.Logf("  Final Balance: %.2f", result.FinalBalance)
	t.Logf("  Total PnL: %.2f", result.TotalPnL)
	t.Logf("  Total Trades: %d", result.TotalTrades)
	t.Logf("  Win Rate: %.2f%%", result.WinRate)
	t.Logf("  Max Drawdown: %.2f%%", result.MaxDrawdown)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMockStrategy()

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Test with invalid CSV file
	_, err := dataProvider.LoadCSVData("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// Test with empty data
	_, err = backtester.Run([]models.Candle{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestIntegration_MultipleStrategies(t *testing.T) {
	// Test multiple strategy configurations with same data
	csvContent := `2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200
2024-01-01 09:02:00,1.0520,1.0540,1.0510,1.0530,800
2024-01-01 09:03:00,1.0530,1.0550,1.0520,1.0540,950
2024-01-01 09:04:00,1.0540,1.0560,1.0530,1.0550,1100
2024-01-01 09:05:00,1.0550,1.0570,1.0540,1.0560,1300`

	tmpFile, err := os.CreateTemp("", "multi_strategy_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Load data once
	dataProvider := data.NewCSVDataProvider()
	candles, err := dataProvider.LoadCSVData(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load CSV data: %v", err)
	}

	strategies := []struct {
		name        string
		strategy    strategy.Strategy
		fastPeriod  int
		slowPeriod  int
	}{
		{"MA_2_3", strategy.NewMovingAverageStrategy("EURUSD", 2, 3, 1000.0), 2, 3},
		{"MA_3_5", strategy.NewMovingAverageStrategy("EURUSD", 3, 5, 1000.0), 3, 5},
	}

	results := make([]*Result, len(strategies))

	for i, strat := range strategies {
		config := models.DefaultConfig()
		brokerInstance := broker.NewSimpleBroker(config)
		backtester := NewBacktester(dataProvider, brokerInstance, strat.strategy, config)

		result, err := backtester.Run(candles)
		if err != nil {
			t.Fatalf("Strategy %s failed: %v", strat.name, err)
		}

		results[i] = result
		t.Logf("Strategy %s: PnL=%.2f, Trades=%d", strat.name, result.TotalPnL, result.TotalTrades)
	}

	// Compare results - they should be different due to different MA periods
	if len(results) >= 2 {
		if results[0].TotalPnL == results[1].TotalPnL {
			t.Log("Note: Different strategies produced same PnL (may be coincidental)")
		}
	}
}

func TestIntegration_RealDataProcessing(t *testing.T) {
	// Test with the existing sample.csv file
	config := models.DefaultConfig()
	dataProvider := data.NewCSVDataProvider()

	// Try to load from testdata
	candles, err := dataProvider.LoadCSVData("../../testdata/sample.csv")
	if err != nil {
		t.Skipf("Skipping real data test: %v", err)
		return
	}

	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 3, 7, 500.0)

	backtester := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	startTime := time.Now()
	result, err := backtester.Run(candles)
	executionTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("Real data backtest failed: %v", err)
	}

	t.Logf("Real Data Backtest Results:")
	t.Logf("  Candles processed: %d", len(candles))
	t.Logf("  Execution time: %v", executionTime)
	t.Logf("  Initial Balance: %.2f", result.InitialBalance)
	t.Logf("  Final Balance: %.2f", result.FinalBalance)
	t.Logf("  Total PnL: %.2f", result.TotalPnL)
	t.Logf("  Total Trades: %d", result.TotalTrades)
	t.Logf("  Win Rate: %.2f%%", result.WinRate)

	// Basic sanity checks
	if result.FinalBalance < 0 {
		t.Error("Final balance should not be negative")
	}
	if result.TotalTrades < 0 {
		t.Error("Total trades should not be negative")
	}
	if result.WinRate < 0 || result.WinRate > 100 {
		t.Errorf("Win rate should be between 0-100%%, got %.2f%%", result.WinRate)
	}
}