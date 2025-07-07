package test

import (
	"math"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

// TestStatisticsIntegration_BacktesterResults validates that statistics
// calculations are consistent with backtest results
func TestStatisticsIntegration_BacktesterResults(t *testing.T) {
	// Create test scenario
	candles := createTestCandles()
	config := models.NewConfig(10000.0, 0.0001, 0.5, 0.0, 100.0)
	
	// Run backtest
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 3, 7, 1000.0)
	bt := backtester.NewBacktester(dataProvider, brokerInstance, strategyInstance, config)
	
	result, err := bt.Run(candles)
	if err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}
	
	// Calculate statistics
	calc := statistics.NewCalculator()
	metrics := calc.CalculateMetrics(result)
	
	// Validate consistency between backtest results and statistics
	validateResultStatisticsConsistency(t, result, metrics)
	
	// Generate reports and validate they contain expected data
	generator := statistics.NewReportGenerator()
	
	textReport := generator.GenerateTextReport(result, metrics)
	if len(textReport) == 0 {
		t.Error("Text report is empty")
	}
	
	jsonReport := generator.GenerateJSONReport(result, metrics)
	if len(jsonReport) == 0 {
		t.Error("JSON report is empty")
	}
	
	if result.TotalTrades > 0 {
		csvReport := generator.GenerateCSVReport(result.Trades)
		if len(csvReport) == 0 {
			t.Error("CSV report is empty despite having trades")
		}
	}
	
	detailedReport := generator.GenerateDetailedReport(result, metrics)
	if len(detailedReport) == 0 {
		t.Error("Detailed report is empty")
	}
}

// TestStatisticsIntegration_ReportAccuracy validates report content accuracy
func TestStatisticsIntegration_ReportAccuracy(t *testing.T) {
	// Create a scenario with known expected results
	result := createKnownResult()
	
	calc := statistics.NewCalculator()
	metrics := calc.CalculateMetrics(result)
	
	// Validate specific calculations
	validateSpecificMetrics(t, result, metrics)
	
	// Test report generation
	generator := statistics.NewReportGenerator()
	
	// Test text report contains key metrics
	textReport := generator.GenerateTextReport(result, metrics)
	validateReportContent(t, textReport, result, metrics)
}

// TestStatisticsIntegration_EdgeCases tests statistics with edge case scenarios
func TestStatisticsIntegration_EdgeCases(t *testing.T) {
	testCases := []struct {
		name   string
		result *backtester.Result
	}{
		{
			name:   "No Trades",
			result: createNoTradesResult(),
		},
		{
			name:   "All Winning Trades",
			result: createAllWinningTradesResult(),
		},
		{
			name:   "All Losing Trades",
			result: createAllLosingTradesResult(),
		},
	}
	
	calc := statistics.NewCalculator()
	generator := statistics.NewReportGenerator()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := calc.CalculateMetrics(tc.result)
			
			// Validate edge case handling
			validateEdgeCaseMetrics(t, tc.result, metrics, tc.name)
			
			// Ensure reports can be generated without errors
			textReport := generator.GenerateTextReport(tc.result, metrics)
			if len(textReport) == 0 {
				t.Errorf("Empty text report for %s", tc.name)
			}
			
			jsonReport := generator.GenerateJSONReport(tc.result, metrics)
			if len(jsonReport) == 0 {
				t.Errorf("Empty JSON report for %s", tc.name)
			}
		})
	}
}

// Helper functions
func createTestCandles() []models.Candle {
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	candles := make([]models.Candle, 15)
	
	for i := 0; i < 15; i++ {
		price := 1.0500 + float64(i)*0.0005
		candles[i] = models.Candle{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      price,
			High:      price + 0.0003,
			Low:       price - 0.0002,
			Close:     price + 0.0001,
			Volume:    1000 + int64(i*50),
		}
	}
	return candles
}

func createKnownResult() *backtester.Result {
	return &backtester.Result{
		StartTime:       time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 1, 9, 10, 0, 0, time.UTC),
		InitialBalance:  10000.0,
		FinalBalance:    10250.0,
		TotalPnL:        250.0,
		TotalTrades:     10,
		WinningTrades:   6,
		LosingTrades:    4,
		WinRate:         60.0,
		MaxDrawdown:     5.2,
		Trades: []backtester.TradeResult{
			{
				Symbol:     "EURUSD",
				Side:       models.OrderSideBuy,
				Size:       1000.0,
				EntryPrice: 1.0500,
				ExitPrice:  1.0520,
				PnL:        20.0,
				EntryTime:  time.Date(2024, 1, 1, 9, 1, 0, 0, time.UTC),
				ExitTime:   time.Date(2024, 1, 1, 9, 2, 0, 0, time.UTC),
				Duration:   time.Minute,
			},
			// Add more test trades as needed
		},
	}
}

func createNoTradesResult() *backtester.Result {
	return &backtester.Result{
		StartTime:       time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 1, 9, 10, 0, 0, time.UTC),
		InitialBalance:  10000.0,
		FinalBalance:    10000.0,
		TotalPnL:        0.0,
		TotalTrades:     0,
		WinningTrades:   0,
		LosingTrades:    0,
		WinRate:         0.0,
		MaxDrawdown:     0.0,
		Trades:          []backtester.TradeResult{},
	}
}

func createAllWinningTradesResult() *backtester.Result {
	return &backtester.Result{
		StartTime:       time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 1, 9, 5, 0, 0, time.UTC),
		InitialBalance:  10000.0,
		FinalBalance:    10150.0,
		TotalPnL:        150.0,
		TotalTrades:     5,
		WinningTrades:   5,
		LosingTrades:    0,
		WinRate:         100.0,
		MaxDrawdown:     0.0,
		Trades:          []backtester.TradeResult{}, // Simplified for test
	}
}

func createAllLosingTradesResult() *backtester.Result {
	return &backtester.Result{
		StartTime:       time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2024, 1, 1, 9, 3, 0, 0, time.UTC),
		InitialBalance:  10000.0,
		FinalBalance:    9850.0,
		TotalPnL:        -150.0,
		TotalTrades:     3,
		WinningTrades:   0,
		LosingTrades:    3,
		WinRate:         0.0,
		MaxDrawdown:     1.5,
		Trades:          []backtester.TradeResult{}, // Simplified for test
	}
}

func validateResultStatisticsConsistency(t *testing.T, result *backtester.Result, metrics *statistics.Metrics) {
	tolerance := 0.01
	
	// Basic count validation
	if metrics.TotalTrades != result.TotalTrades {
		t.Errorf("TotalTrades mismatch: result=%d, metrics=%d", 
			result.TotalTrades, metrics.TotalTrades)
	}
	
	if metrics.WinningTrades != result.WinningTrades {
		t.Errorf("WinningTrades mismatch: result=%d, metrics=%d", 
			result.WinningTrades, metrics.WinningTrades)
	}
	
	if metrics.LosingTrades != result.LosingTrades {
		t.Errorf("LosingTrades mismatch: result=%d, metrics=%d", 
			result.LosingTrades, metrics.LosingTrades)
	}
	
	// Win rate validation
	if math.Abs(metrics.WinRate-result.WinRate) > tolerance {
		t.Errorf("WinRate mismatch: result=%.2f, metrics=%.2f", 
			result.WinRate, metrics.WinRate)
	}
	
	// P&L validation
	if math.Abs(metrics.TotalPnL-result.TotalPnL) > tolerance {
		t.Errorf("TotalPnL mismatch: result=%.2f, metrics=%.2f", 
			result.TotalPnL, metrics.TotalPnL)
	}
	
	// Total return calculation
	expectedTotalReturn := ((result.FinalBalance - result.InitialBalance) / result.InitialBalance) * 100
	if math.Abs(metrics.TotalReturn-expectedTotalReturn) > tolerance {
		t.Errorf("TotalReturn calculation error: expected %.2f, got %.2f", 
			expectedTotalReturn, metrics.TotalReturn)
	}
}

func validateSpecificMetrics(t *testing.T, result *backtester.Result, metrics *statistics.Metrics) {
	// Test specific metric calculations with known inputs
	
	if result.TotalTrades > 0 {
		// Win rate should match calculation
		expectedWinRate := (float64(result.WinningTrades) / float64(result.TotalTrades)) * 100
		if math.Abs(metrics.WinRate-expectedWinRate) > 0.01 {
			t.Errorf("Win rate calculation error: expected %.2f, got %.2f", 
				expectedWinRate, metrics.WinRate)
		}
	}
	
	// Validate range bounds
	if metrics.WinRate < 0 || metrics.WinRate > 100 {
		t.Errorf("Win rate out of bounds: %.2f", metrics.WinRate)
	}
	
	if metrics.MaxDrawdown < 0 {
		t.Errorf("Negative max drawdown: %.2f", metrics.MaxDrawdown)
	}
}

func validateReportContent(t *testing.T, report string, result *backtester.Result, metrics *statistics.Metrics) {
	// Check that key information is present in the report
	requiredElements := []string{
		"BACKTEST RESULTS",
		"PERFORMANCE SUMMARY",
		"TRADE STATISTICS",
		"RISK METRICS",
	}
	
	for _, element := range requiredElements {
		if !contains(report, element) {
			t.Errorf("Report missing required element: %s", element)
		}
	}
	
	// Check that numerical values are reasonable
	if result.TotalTrades > 0 && !contains(report, "Total Trades") {
		t.Error("Report missing trade count information")
	}
}

func validateEdgeCaseMetrics(t *testing.T, result *backtester.Result, metrics *statistics.Metrics, caseName string) {
	switch caseName {
	case "No Trades":
		if metrics.TotalTrades != 0 {
			t.Errorf("No trades case should have 0 trades, got %d", metrics.TotalTrades)
		}
		if metrics.WinRate != 0 {
			t.Errorf("No trades case should have 0%% win rate, got %.2f", metrics.WinRate)
		}
		
	case "All Winning Trades":
		if metrics.WinRate != 100.0 {
			t.Errorf("All winning trades should have 100%% win rate, got %.2f", metrics.WinRate)
		}
		if metrics.LosingTrades != 0 {
			t.Errorf("All winning trades should have 0 losing trades, got %d", metrics.LosingTrades)
		}
		
	case "All Losing Trades":
		if metrics.WinRate != 0.0 {
			t.Errorf("All losing trades should have 0%% win rate, got %.2f", metrics.WinRate)
		}
		if metrics.WinningTrades != 0 {
			t.Errorf("All losing trades should have 0 winning trades, got %d", metrics.WinningTrades)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsInternal(s, substr))))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}