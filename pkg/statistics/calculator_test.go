package statistics

import (
	"math"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
)

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator()
	
	if calc == nil {
		t.Fatal("Expected calculator to be created")
	}
}

func TestCalculator_CalculateMetrics(t *testing.T) {
	calc := NewCalculator()
	
	// Create sample result with trades
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
	
	if metrics == nil {
		t.Fatal("Expected metrics to be calculated")
	}
	
	// Check basic metrics
	if metrics.TotalReturn != 10.0 {
		t.Errorf("Expected total return 10.0%%, got %.2f%%", metrics.TotalReturn)
	}
	
	if metrics.WinRate != 60.0 {
		t.Errorf("Expected win rate 60.0%%, got %.2f%%", metrics.WinRate)
	}
	
	if metrics.MaxDrawdown != 5.0 {
		t.Errorf("Expected max drawdown 5.0%%, got %.2f%%", metrics.MaxDrawdown)
	}
}

func TestCalculator_CalculateSharpeRatio(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
		{PnL: -50.0},
		{PnL: 200.0},
		{PnL: -30.0},
	}
	
	sharpe := calc.CalculateSharpeRatio(trades, 0.02) // 2% risk-free rate
	
	if sharpe == 0.0 {
		t.Error("Expected non-zero Sharpe ratio")
	}
	
	// Test with no trades
	sharpeEmpty := calc.CalculateSharpeRatio([]backtester.TradeResult{}, 0.02)
	if sharpeEmpty != 0.0 {
		t.Errorf("Expected zero Sharpe ratio for no trades, got %.4f", sharpeEmpty)
	}
}

func TestCalculator_CalculateProfitFactor(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
		{PnL: -50.0},
		{PnL: 200.0},
		{PnL: -30.0},
	}
	
	pf := calc.CalculateProfitFactor(trades)
	expectedGrossProfit := 450.0 // 100 + 150 + 200
	expectedGrossLoss := 80.0    // 50 + 30
	expectedPF := expectedGrossProfit / expectedGrossLoss
	
	if math.Abs(pf-expectedPF) > 0.001 {
		t.Errorf("Expected profit factor %.3f, got %.3f", expectedPF, pf)
	}
	
	// Test with no losing trades
	winningTrades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
	}
	pfWinning := calc.CalculateProfitFactor(winningTrades)
	if pfWinning <= 0 {
		t.Error("Expected positive profit factor for winning trades only")
	}
	
	// Test with no winning trades
	losingTrades := []backtester.TradeResult{
		{PnL: -100.0},
		{PnL: -150.0},
	}
	pfLosing := calc.CalculateProfitFactor(losingTrades)
	if pfLosing != 0.0 {
		t.Error("Expected zero profit factor for losing trades only")
	}
}

func TestCalculator_CalculateAverageWin(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
		{PnL: -50.0},
		{PnL: 200.0},
		{PnL: -30.0},
	}
	
	avgWin := calc.CalculateAverageWin(trades)
	expected := (100.0 + 150.0 + 200.0) / 3.0
	
	if math.Abs(avgWin-expected) > 0.001 {
		t.Errorf("Expected average win %.3f, got %.3f", expected, avgWin)
	}
	
	// Test with no winning trades
	losingTrades := []backtester.TradeResult{
		{PnL: -100.0},
		{PnL: -150.0},
	}
	avgWinLosing := calc.CalculateAverageWin(losingTrades)
	if avgWinLosing != 0.0 {
		t.Error("Expected zero average win for losing trades only")
	}
}

func TestCalculator_CalculateAverageLoss(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
		{PnL: -50.0},
		{PnL: 200.0},
		{PnL: -30.0},
	}
	
	avgLoss := calc.CalculateAverageLoss(trades)
	expected := (50.0 + 30.0) / 2.0
	
	if math.Abs(avgLoss-expected) > 0.001 {
		t.Errorf("Expected average loss %.3f, got %.3f", expected, avgLoss)
	}
	
	// Test with no losing trades
	winningTrades := []backtester.TradeResult{
		{PnL: 100.0},
		{PnL: 150.0},
	}
	avgLossWinning := calc.CalculateAverageLoss(winningTrades)
	if avgLossWinning != 0.0 {
		t.Error("Expected zero average loss for winning trades only")
	}
}

func TestCalculator_CalculateMaxConsecutiveWins(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},  // Win
		{PnL: 150.0},  // Win
		{PnL: 200.0},  // Win
		{PnL: -50.0},  // Loss
		{PnL: 80.0},   // Win
		{PnL: 90.0},   // Win
		{PnL: -30.0},  // Loss
	}
	
	maxWins := calc.CalculateMaxConsecutiveWins(trades)
	if maxWins != 3 {
		t.Errorf("Expected max consecutive wins 3, got %d", maxWins)
	}
	
	// Test with no trades
	maxWinsEmpty := calc.CalculateMaxConsecutiveWins([]backtester.TradeResult{})
	if maxWinsEmpty != 0 {
		t.Errorf("Expected max consecutive wins 0 for empty trades, got %d", maxWinsEmpty)
	}
}

func TestCalculator_CalculateMaxConsecutiveLosses(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{PnL: 100.0},  // Win
		{PnL: -50.0},  // Loss
		{PnL: -30.0},  // Loss
		{PnL: -20.0},  // Loss
		{PnL: 80.0},   // Win
		{PnL: -10.0},  // Loss
		{PnL: -15.0},  // Loss
	}
	
	maxLosses := calc.CalculateMaxConsecutiveLosses(trades)
	if maxLosses != 3 {
		t.Errorf("Expected max consecutive losses 3, got %d", maxLosses)
	}
}

func TestCalculator_CalculateAverageTradeDuration(t *testing.T) {
	calc := NewCalculator()
	
	trades := []backtester.TradeResult{
		{Duration: time.Hour},
		{Duration: 2 * time.Hour},
		{Duration: 30 * time.Minute},
	}
	
	avgDuration := calc.CalculateAverageTradeDuration(trades)
	expected := (time.Hour + 2*time.Hour + 30*time.Minute) / 3
	
	if avgDuration != expected {
		t.Errorf("Expected average duration %v, got %v", expected, avgDuration)
	}
	
	// Test with no trades
	avgDurationEmpty := calc.CalculateAverageTradeDuration([]backtester.TradeResult{})
	if avgDurationEmpty != 0 {
		t.Errorf("Expected zero duration for empty trades, got %v", avgDurationEmpty)
	}
}

func TestCalculator_CalculateMaxDrawdownPercent(t *testing.T) {
	calc := NewCalculator()
	
	// Test with equity curve that shows drawdown
	equityCurve := []float64{10000, 10500, 10200, 9800, 9500, 10000, 11000}
	
	maxDD := calc.CalculateMaxDrawdownPercent(equityCurve)
	
	// Peak was 10500, trough was 9500, so drawdown = (10500-9500)/10500 * 100 = 9.52%
	expected := (10500.0 - 9500.0) / 10500.0 * 100
	
	if math.Abs(maxDD-expected) > 0.01 {
		t.Errorf("Expected max drawdown %.2f%%, got %.2f%%", expected, maxDD)
	}
	
	// Test with always increasing equity
	increasingEquity := []float64{10000, 10500, 11000, 11500}
	maxDDIncreasing := calc.CalculateMaxDrawdownPercent(increasingEquity)
	if maxDDIncreasing != 0.0 {
		t.Errorf("Expected zero drawdown for increasing equity, got %.2f%%", maxDDIncreasing)
	}
}