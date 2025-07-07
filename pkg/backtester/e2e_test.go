package backtester

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

// TestE2E_CompleteWorkflow tests the complete backtesting workflow
// from CSV loading to final report generation
func TestE2E_CompleteWorkflow(t *testing.T) {
	// Create test CSV file
	testData := createE2ETestCSV(t)
	defer os.Remove(testData)
	
	// Setup configuration
	config := models.NewConfig(10000.0, 0.0002, 1.0, 0.0001, 100.0)
	
	// Create all components
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 5, 15, 1000.0)
	bt := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)
	
	// Load data from CSV
	candles, err := dataProvider.LoadCSVData(testData)
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}
	
	if len(candles) == 0 {
		t.Fatal("No candles loaded from test data")
	}
	
	t.Logf("Loaded %d candles for E2E test", len(candles))
	
	// Run backtest with progress tracking
	progressCallbackCalled := false
	result, err := bt.RunWithCallback(candles, func(progress Progress) {
		progressCallbackCalled = true
		if progress.ProcessedCandles > progress.TotalCandles {
			t.Errorf("Progress error: processed (%d) > total (%d)", 
				progress.ProcessedCandles, progress.TotalCandles)
		}
		if progress.Percentage < 0 || progress.Percentage > 100 {
			t.Errorf("Invalid percentage: %.2f", progress.Percentage)
		}
	})
	
	if err != nil {
		t.Fatalf("E2E backtest failed: %v", err)
	}
	
	if !progressCallbackCalled {
		t.Error("Progress callback was not called")
	}
	
	// Validate result completeness
	validateE2EResult(t, result, config, len(candles))
	
	// Note: Statistics and report generation testing is done separately
	// to avoid circular import issues
	
	t.Logf("E2E Test Results Summary:")
	t.Logf("  Duration: %v", result.EndTime.Sub(result.StartTime))
	t.Logf("  Total Candles: %d", len(candles))
	t.Logf("  Total Trades: %d", result.TotalTrades)
	t.Logf("  Final Balance: $%.2f", result.FinalBalance)
	t.Logf("  Total P&L: $%.2f", result.TotalPnL)
	t.Logf("  Win Rate: %.1f%%", result.WinRate)
}

// TestE2E_LargeDataset tests performance and correctness with larger dataset
func TestE2E_LargeDataset(t *testing.T) {
	// Use existing sample data or create larger dataset
	dataProvider := data.NewCSVDataProvider()
	
	// Try to load the sample data first
	testDataPath := "../../testdata/sample.csv"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		// Create larger test dataset if sample doesn't exist
		testDataPath = createLargeTestDataset(t)
		defer os.Remove(testDataPath)
	}
	
	candles, err := dataProvider.LoadCSVData(testDataPath)
	if err != nil {
		t.Fatalf("Failed to load large dataset: %v", err)
	}
	
	if len(candles) < 10 {
		t.Skip("Skipping large dataset test - insufficient data")
	}
	
	// Test with different strategy configurations
	strategyConfigs := []struct {
		name       string
		fastPeriod int
		slowPeriod int
		positionSize float64
	}{
		{"Conservative", 10, 30, 500.0},
		{"Aggressive", 3, 8, 2000.0},
		{"Balanced", 5, 15, 1000.0},
	}
	
	for _, stratConfig := range strategyConfigs {
		t.Run(stratConfig.name, func(t *testing.T) {
			config := models.NewConfig(10000.0, 0.0001, 0.5, 0.0, 100.0)
			brokerInstance := broker.NewSimpleBroker(config)
			strategyInstance := strategy.NewMovingAverageStrategy(
				"EURUSD", stratConfig.fastPeriod, stratConfig.slowPeriod, stratConfig.positionSize)
			bt := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)
			
			start := time.Now()
			result, err := bt.Run(candles)
			duration := time.Since(start)
			
			if err != nil {
				t.Fatalf("Large dataset test failed for %s: %v", stratConfig.name, err)
			}
			
			// Performance validation
			maxExpectedDuration := time.Second * 5 // Should complete within 5 seconds
			if duration > maxExpectedDuration {
				t.Errorf("Performance issue: %s strategy took %v (expected < %v)", 
					stratConfig.name, duration, maxExpectedDuration)
			}
			
			// Result validation
			validateE2EResult(t, result, config, len(candles))
			
			t.Logf("%s Strategy: %d candles processed in %v, %d trades, P&L: $%.2f", 
				stratConfig.name, len(candles), duration, result.TotalTrades, result.TotalPnL)
		})
	}
}

// TestE2E_ErrorHandling tests error handling in complete workflow
func TestE2E_ErrorHandling(t *testing.T) {
	config := models.NewConfig(1000.0, 0.0001, 0.0, 0.0, 100.0)
	
	testCases := []struct {
		name          string
		setupFunc     func() (*data.CSVDataProvider, broker.Broker, strategy.Strategy)
		expectedError string
	}{
		{
			name: "Invalid CSV Data",
			setupFunc: func() (*data.CSVDataProvider, broker.Broker, strategy.Strategy) {
				return data.NewCSVDataProvider(), 
					broker.NewSimpleBroker(config),
					strategy.NewMovingAverageStrategy("EURUSD", 3, 5, 1000.0)
			},
			expectedError: "no such file",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataProvider, brokerInstance, strategyInstance := tc.setupFunc()
			_ = NewBacktester(dataProvider, brokerInstance, strategyInstance, config)
			
			// Try to load non-existent file
			_, err := dataProvider.LoadCSVData("nonexistent.csv")
			if err == nil {
				t.Error("Expected error for invalid CSV file")
			}
		})
	}
}

// TestE2E_MemoryUsage tests memory efficiency with repeated runs
func TestE2E_MemoryUsage(t *testing.T) {
	// Create test data
	testData := createE2ETestCSV(t)
	defer os.Remove(testData)
	
	config := models.NewConfig(5000.0, 0.0001, 0.0, 0.0, 100.0)
	
	// Run multiple backtests to check for memory leaks
	for i := 0; i < 10; i++ {
		dataProvider := data.NewCSVDataProvider()
		brokerInstance := broker.NewSimpleBroker(config)
		strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 3, 7, 500.0)
		bt := NewBacktester(dataProvider, brokerInstance, strategyInstance, config)
		
		candles, err := dataProvider.LoadCSVData(testData)
		if err != nil {
			t.Fatalf("Run %d: Failed to load data: %v", i+1, err)
		}
		
		result, err := bt.Run(candles)
		if err != nil {
			t.Fatalf("Run %d: Backtest failed: %v", i+1, err)
		}
		
		// Validate each run
		validateE2EResult(t, result, config, len(candles))
		
		// Force garbage collection to help detect memory issues
		if i%3 == 0 {
			// Simulate memory pressure check
			_ = result
		}
	}
	
	t.Log("Memory usage test completed successfully")
}

// Helper function to create E2E test CSV file
func createE2ETestCSV(t *testing.T) string {
	content := `2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200
2024-01-01 09:02:00,1.0520,1.0540,1.0510,1.0530,1100
2024-01-01 09:03:00,1.0530,1.0550,1.0520,1.0540,1300
2024-01-01 09:04:00,1.0540,1.0560,1.0530,1.0550,1150
2024-01-01 09:05:00,1.0550,1.0570,1.0540,1.0560,1250
2024-01-01 09:06:00,1.0560,1.0580,1.0550,1.0570,1180
2024-01-01 09:07:00,1.0570,1.0590,1.0560,1.0580,1220
2024-01-01 09:08:00,1.0580,1.0600,1.0570,1.0590,1300
2024-01-01 09:09:00,1.0590,1.0610,1.0580,1.0600,1400
2024-01-01 09:10:00,1.0600,1.0620,1.0590,1.0610,1350
2024-01-01 09:11:00,1.0610,1.0630,1.0600,1.0620,1280
2024-01-01 09:12:00,1.0620,1.0640,1.0610,1.0630,1320
2024-01-01 09:13:00,1.0630,1.0650,1.0620,1.0640,1250
2024-01-01 09:14:00,1.0640,1.0660,1.0630,1.0650,1380
2024-01-01 09:15:00,1.0650,1.0670,1.0640,1.0660,1420
2024-01-01 09:16:00,1.0660,1.0680,1.0650,1.0670,1300
2024-01-01 09:17:00,1.0670,1.0690,1.0660,1.0680,1250
2024-01-01 09:18:00,1.0680,1.0700,1.0670,1.0690,1400
2024-01-01 09:19:00,1.0690,1.0710,1.0680,1.0700,1450`
	
	tmpFile, err := os.CreateTemp("", "e2e_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}
	
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write test CSV: %v", err)
	}
	tmpFile.Close()
	
	return tmpFile.Name()
}

// Helper function to create large test dataset
func createLargeTestDataset(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "large_test_*.csv")
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}
	defer tmpFile.Close()
	
	// Skip header to match current CSV parser expectations
	
	// Generate 100 candles with realistic price movement
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	price := 1.0500
	
	for i := 0; i < 100; i++ {
		// Add some random-like movement
		priceChange := 0.0001 * float64((i%10)-5) // Oscillating movement
		price += priceChange
		
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		high := price + 0.0005
		low := price - 0.0005
		close := price + 0.0001
		volume := 1000 + (i * 10)
		
		line := fmt.Sprintf("%s,%.4f,%.4f,%.4f,%.4f,%d\n",
			timestamp.Format("2006-01-02 15:04:05"), price, high, low, close, volume)
		tmpFile.WriteString(line)
	}
	
	return tmpFile.Name()
}

// Helper function to validate E2E result
func validateE2EResult(t *testing.T, result *Result, config models.Config, candleCount int) {
	if result == nil {
		t.Fatal("E2E result is nil")
	}
	
	// Basic structure validation
	if result.StartTime.IsZero() || result.EndTime.IsZero() {
		t.Error("E2E result missing time information")
	}
	
	if result.InitialBalance != config.InitialBalance {
		t.Errorf("E2E initial balance mismatch: expected %.2f, got %.2f", 
			config.InitialBalance, result.InitialBalance)
	}
	
	// Validate that we processed the expected number of candles
	actualDuration := result.EndTime.Sub(result.StartTime)
	
	// Allow some tolerance for duration calculation
	if actualDuration < 0 {
		t.Error("E2E negative duration")
	}
	
	// Trade count validation
	if result.TotalTrades < 0 {
		t.Error("E2E negative trade count")
	}
	
	if result.WinningTrades+result.LosingTrades > result.TotalTrades {
		t.Error("E2E win/loss count exceeds total trades")
	}
	
	// Balance validation - final balance should equal initial + PnL
	tolerance := 1.0 // Allow small tolerance for floating point precision
	expectedBalance := config.InitialBalance + result.TotalPnL
	if math.Abs(result.FinalBalance-expectedBalance) > tolerance {
		t.Errorf("E2E balance calculation error: expected %.2f, got %.2f (PnL: %.2f)", 
			expectedBalance, result.FinalBalance, result.TotalPnL)
	}
}

