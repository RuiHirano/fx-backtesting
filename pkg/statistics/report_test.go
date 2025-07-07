package statistics

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
)

func TestNewReportGenerator(t *testing.T) {
	generator := NewReportGenerator()
	
	if generator == nil {
		t.Fatal("Expected report generator to be created")
	}
}

func TestReportGenerator_GenerateTextReport(t *testing.T) {
	generator := NewReportGenerator()
	calc := NewCalculator()
	
	// Create sample result
	result := &backtester.Result{
		StartTime:      time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
		Duration:       8 * time.Hour,
		InitialBalance: 10000.0,
		FinalBalance:   11000.0,
		TotalPnL:       1000.0,
		TotalTrades:    5,
		WinningTrades:  3,
		LosingTrades:   2,
		WinRate:        60.0,
		MaxDrawdown:    5.0,
		Trades: []backtester.TradeResult{
			{PnL: 200.0, Duration: time.Hour},
			{PnL: -100.0, Duration: 30 * time.Minute},
			{PnL: 300.0, Duration: 2 * time.Hour},
			{PnL: -50.0, Duration: 45 * time.Minute},
			{PnL: 650.0, Duration: 3 * time.Hour},
		},
	}
	
	metrics := calc.CalculateMetrics(result)
	report := generator.GenerateTextReport(result, metrics)
	
	if report == "" {
		t.Error("Expected non-empty text report")
	}
	
	// Check that report contains key information
	expectedContents := []string{
		"BACKTEST RESULTS",
		"PERFORMANCE SUMMARY",
		"TRADE STATISTICS",
		"RISK METRICS",
		"Total Return",
		"Win Rate",
		"Sharpe Ratio",
		"Max Drawdown",
	}
	
	for _, content := range expectedContents {
		if !strings.Contains(report, content) {
			t.Errorf("Report missing expected content: %s", content)
		}
	}
}

func TestReportGenerator_GenerateJSONReport(t *testing.T) {
	generator := NewReportGenerator()
	calc := NewCalculator()
	
	// Create sample result
	result := &backtester.Result{
		StartTime:      time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC),
		Duration:       8 * time.Hour,
		InitialBalance: 10000.0,
		FinalBalance:   11000.0,
		TotalPnL:       1000.0,
		TotalTrades:    5,
		WinningTrades:  3,
		LosingTrades:   2,
		WinRate:        60.0,
		MaxDrawdown:    5.0,
		Trades: []backtester.TradeResult{
			{PnL: 200.0},
			{PnL: -100.0},
		},
	}
	
	metrics := calc.CalculateMetrics(result)
	jsonReport := generator.GenerateJSONReport(result, metrics)
	
	if jsonReport == "" {
		t.Error("Expected non-empty JSON report")
	}
	
	// Verify it's valid JSON
	var reportData map[string]interface{}
	err := json.Unmarshal([]byte(jsonReport), &reportData)
	if err != nil {
		t.Fatalf("Invalid JSON report: %v", err)
	}
	
	// Check that key sections exist
	expectedSections := []string{"backtest_summary", "performance_metrics", "trade_statistics", "risk_metrics"}
	for _, section := range expectedSections {
		if _, exists := reportData[section]; !exists {
			t.Errorf("JSON report missing section: %s", section)
		}
	}
}

func TestReportGenerator_GenerateCSVReport(t *testing.T) {
	generator := NewReportGenerator()
	
	// Create sample trades
	trades := []backtester.TradeResult{
		{
			EntryTime:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			Symbol:     "EURUSD",
			Side:       0, // Buy
			Size:       1000.0,
			EntryPrice: 1.0500,
			ExitPrice:  1.0520,
			PnL:        200.0,
			Duration:   time.Hour,
		},
		{
			EntryTime:  time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			ExitTime:   time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
			Symbol:     "EURUSD",
			Side:       1, // Sell
			Size:       1000.0,
			EntryPrice: 1.0520,
			ExitPrice:  1.0510,
			PnL:        100.0,
			Duration:   30 * time.Minute,
		},
	}
	
	csvReport := generator.GenerateCSVReport(trades)
	
	if csvReport == "" {
		t.Error("Expected non-empty CSV report")
	}
	
	// Check CSV headers
	lines := strings.Split(csvReport, "\n")
	if len(lines) < 2 {
		t.Error("Expected at least header and one data row")
	}
	
	header := lines[0]
	expectedHeaders := []string{"Entry Time", "Exit Time", "Symbol", "Side", "Size", "Entry Price", "Exit Price", "PnL", "Duration"}
	for _, expectedHeader := range expectedHeaders {
		if !strings.Contains(header, expectedHeader) {
			t.Errorf("CSV header missing: %s", expectedHeader)
		}
	}
	
	// Check data rows
	if len(lines) < 3 {
		t.Error("Expected at least 2 data rows")
	}
}

func TestReportGenerator_GenerateDetailedReport(t *testing.T) {
	generator := NewReportGenerator()
	calc := NewCalculator()
	
	// Create comprehensive test data
	result := &backtester.Result{
		StartTime:      time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2024, 12, 31, 17, 0, 0, 0, time.UTC),
		Duration:       365 * 24 * time.Hour,
		InitialBalance: 10000.0,
		FinalBalance:   12000.0,
		TotalPnL:       2000.0,
		TotalTrades:    100,
		WinningTrades:  60,
		LosingTrades:   40,
		WinRate:        60.0,
		MaxDrawdown:    8.5,
		Trades:         generateTestTrades(100),
	}
	
	metrics := calc.CalculateMetrics(result)
	report := generator.GenerateDetailedReport(result, metrics)
	
	if report == "" {
		t.Error("Expected non-empty detailed report")
	}
	
	// Check for detailed sections
	expectedSections := []string{
		"DETAILED BACKTEST ANALYSIS",
		"EXECUTIVE SUMMARY",
		"PERFORMANCE ANALYSIS",
		"RISK ANALYSIS",
		"TRADE DISTRIBUTION",
		"MONTHLY PERFORMANCE",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(report, section) {
			t.Errorf("Detailed report missing section: %s", section)
		}
	}
}

func TestReportGenerator_FormatDuration(t *testing.T) {
	generator := NewReportGenerator()
	
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{time.Hour, "1h 0m"},
		{90 * time.Minute, "1h 30m"},
		{30 * time.Minute, "0h 30m"},
		{25 * time.Hour, "1d 1h"},
		{0, "0h 0m"},
	}
	
	for _, tc := range testCases {
		formatted := generator.formatDuration(tc.duration)
		if formatted != tc.expected {
			t.Errorf("Expected duration format %s, got %s", tc.expected, formatted)
		}
	}
}

func TestReportGenerator_FormatPercentage(t *testing.T) {
	generator := NewReportGenerator()
	
	testCases := []struct {
		value    float64
		expected string
	}{
		{15.5678, "15.57%"},
		{-5.1234, "-5.12%"},
		{0.0, "0.00%"},
		{100.0, "100.00%"},
	}
	
	for _, tc := range testCases {
		formatted := generator.formatPercentage(tc.value)
		if formatted != tc.expected {
			t.Errorf("Expected percentage format %s, got %s", tc.expected, formatted)
		}
	}
}

// Helper function to generate test trades
func generateTestTrades(count int) []backtester.TradeResult {
	trades := make([]backtester.TradeResult, count)
	baseTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	
	for i := 0; i < count; i++ {
		// Generate alternating wins and losses with some randomness
		pnl := float64(100 + i*5)
		if i%3 == 0 {
			pnl = -pnl / 2 // Some losses
		}
		
		trades[i] = backtester.TradeResult{
			EntryTime:  baseTime.Add(time.Duration(i) * time.Hour),
			ExitTime:   baseTime.Add(time.Duration(i+1) * time.Hour),
			Symbol:     "EURUSD",
			Side:       0,
			Size:       1000.0,
			EntryPrice: 1.0500 + float64(i)*0.0001,
			ExitPrice:  1.0500 + float64(i)*0.0001 + pnl/10000,
			PnL:        pnl,
			Duration:   time.Hour,
		}
	}
	
	return trades
}