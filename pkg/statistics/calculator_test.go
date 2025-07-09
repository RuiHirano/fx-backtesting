package statistics

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Calculator NewCalculator テスト
func TestCalculator_NewCalculator(t *testing.T) {
	// テスト用取引履歴作成
	trades := createTestTrades()
	
	calculator := NewCalculator(trades)
	
	if calculator == nil {
		t.Fatal("Expected calculator to be created")
	}
	
	// 初期状態確認
	if len(calculator.GetTrades()) != len(trades) {
		t.Errorf("Expected %d trades, got %d", len(trades), len(calculator.GetTrades()))
	}
}

// Calculator 基本統計指標テスト
func TestCalculator_BasicMetrics(t *testing.T) {
	// テスト用取引履歴作成（利益・損失の混在）
	trades := []*models.Trade{
		createTrade("trade-1", 100.0, time.Now()),   // 利益
		createTrade("trade-2", -50.0, time.Now()),   // 損失
		createTrade("trade-3", 200.0, time.Now()),   // 利益
		createTrade("trade-4", -30.0, time.Now()),   // 損失
		createTrade("trade-5", 80.0, time.Now()),    // 利益
	}
	
	calculator := NewCalculator(trades)
	
	// 総利益・損失テスト
	totalPnL := calculator.CalculateTotalPnL()
	expectedTotal := 100.0 - 50.0 + 200.0 - 30.0 + 80.0 // 300.0
	if totalPnL != expectedTotal {
		t.Errorf("Expected total PnL %f, got %f", expectedTotal, totalPnL)
	}
	
	// 勝率テスト
	winRate := calculator.CalculateWinRate()
	expectedWinRate := 3.0 / 5.0 // 60%
	if winRate != expectedWinRate {
		t.Errorf("Expected win rate %f, got %f", expectedWinRate, winRate)
	}
	
	// 平均利益テスト
	avgProfit := calculator.CalculateAverageProfit()
	expectedAvgProfit := (100.0 + 200.0 + 80.0) / 3.0 // 126.67
	if avgProfit != expectedAvgProfit {
		t.Errorf("Expected average profit %f, got %f", expectedAvgProfit, avgProfit)
	}
	
	// 平均損失テスト
	avgLoss := calculator.CalculateAverageLoss()
	expectedAvgLoss := (50.0 + 30.0) / 2.0 // 40.0
	if avgLoss != expectedAvgLoss {
		t.Errorf("Expected average loss %f, got %f", expectedAvgLoss, avgLoss)
	}
	
	// 最大利益テスト
	maxProfit := calculator.CalculateMaxProfit()
	expectedMaxProfit := 200.0
	if maxProfit != expectedMaxProfit {
		t.Errorf("Expected max profit %f, got %f", expectedMaxProfit, maxProfit)
	}
	
	// 最大損失テスト
	maxLoss := calculator.CalculateMaxLoss()
	expectedMaxLoss := 50.0
	if maxLoss != expectedMaxLoss {
		t.Errorf("Expected max loss %f, got %f", expectedMaxLoss, maxLoss)
	}
	
	// 取引回数テスト
	totalTrades := calculator.CalculateTotalTrades()
	if totalTrades != 5 {
		t.Errorf("Expected 5 trades, got %d", totalTrades)
	}
}

// Calculator リスク指標テスト
func TestCalculator_RiskMetrics(t *testing.T) {
	// テスト用取引履歴作成（ドローダウンパターン）
	trades := []*models.Trade{
		createTrade("trade-1", 100.0, time.Now()),
		createTrade("trade-2", -200.0, time.Now()),
		createTrade("trade-3", -100.0, time.Now()),
		createTrade("trade-4", 300.0, time.Now()),
		createTrade("trade-5", 50.0, time.Now()),
	}
	
	calculator := NewCalculator(trades)
	
	// 最大ドローダウンテスト
	maxDrawdown := calculator.CalculateMaxDrawdown()
	// 累積: 100, -100, -200, 100, 150
	// ドローダウン: 0, 200, 300, 0, 0
	expectedMaxDrawdown := 300.0
	if maxDrawdown != expectedMaxDrawdown {
		t.Errorf("Expected max drawdown %f, got %f", expectedMaxDrawdown, maxDrawdown)
	}
	
	// シャープレシオテスト
	sharpeRatio := calculator.CalculateSharpeRatio()
	if sharpeRatio <= 0 {
		t.Error("Expected positive Sharpe ratio")
	}
	
	// ソルティノレシオテスト
	sortinoRatio := calculator.CalculateSortinoRatio()
	if sortinoRatio <= 0 {
		t.Error("Expected positive Sortino ratio")
	}
	
	// リターン・リスク比テスト
	returnRiskRatio := calculator.CalculateReturnRiskRatio()
	if returnRiskRatio <= 0 {
		t.Error("Expected positive return/risk ratio")
	}
}

// Calculator 高度統計指標テスト
func TestCalculator_AdvancedMetrics(t *testing.T) {
	trades := createTestTrades()
	calculator := NewCalculator(trades)
	
	// カルマーレシオテスト
	calmarRatio := calculator.CalculateCalmarRatio()
	if calmarRatio <= 0 {
		t.Error("Expected positive Calmar ratio")
	}
	
	// プロフィットファクターテスト
	profitFactor := calculator.CalculateProfitFactor()
	if profitFactor <= 0 {
		t.Error("Expected positive profit factor")
	}
	
	// 期待値テスト
	expectedValue := calculator.CalculateExpectedValue()
	if expectedValue == 0 {
		t.Error("Expected non-zero expected value")
	}
	
	// 標準偏差テスト
	stdDev := calculator.CalculateStandardDeviation()
	if stdDev <= 0 {
		t.Error("Expected positive standard deviation")
	}
}

// Calculator 取引関連指標テスト
func TestCalculator_TradingMetrics(t *testing.T) {
	// 時間間隔のある取引履歴作成
	baseTime := time.Now()
	trades := []*models.Trade{
		createTrade("trade-1", 100.0, baseTime),
		createTrade("trade-2", -50.0, baseTime.Add(24*time.Hour)),
		createTrade("trade-3", 75.0, baseTime.Add(48*time.Hour)),
	}
	
	calculator := NewCalculator(trades)
	
	// 平均保有期間テスト
	avgHoldingPeriod := calculator.CalculateAverageHoldingPeriod()
	if avgHoldingPeriod <= 0 {
		t.Error("Expected positive average holding period")
	}
	
	// 最大連勝テスト
	maxConsecutiveWins := calculator.CalculateMaxConsecutiveWins()
	if maxConsecutiveWins < 0 {
		t.Error("Expected non-negative max consecutive wins")
	}
	
	// 最大連敗テスト
	maxConsecutiveLosses := calculator.CalculateMaxConsecutiveLosses()
	if maxConsecutiveLosses < 0 {
		t.Error("Expected non-negative max consecutive losses")
	}
	
	// 取引頻度テスト（1日あたりの取引数）
	tradingFrequency := calculator.CalculateTradingFrequency()
	if tradingFrequency <= 0 {
		t.Error("Expected positive trading frequency")
	}
	
	// リスクリワード比テスト
	riskRewardRatio := calculator.CalculateRiskRewardRatio()
	if riskRewardRatio <= 0 {
		t.Error("Expected positive risk reward ratio")
	}
}

// Calculator エラーハンドリングテスト
func TestCalculator_ErrorHandling(t *testing.T) {
	// 空の取引履歴テスト
	emptyTrades := []*models.Trade{}
	calculator := NewCalculator(emptyTrades)
	
	// 統計計算時のエラーハンドリング確認
	totalPnL := calculator.CalculateTotalPnL()
	if totalPnL != 0.0 {
		t.Errorf("Expected 0 PnL for empty trades, got %f", totalPnL)
	}
	
	winRate := calculator.CalculateWinRate()
	if winRate != 0.0 {
		t.Errorf("Expected 0 win rate for empty trades, got %f", winRate)
	}
	
	// ゼロ除算回避確認
	avgProfit := calculator.CalculateAverageProfit()
	if avgProfit != 0.0 {
		t.Errorf("Expected 0 average profit for empty trades, got %f", avgProfit)
	}
	
	// nilトレード処理テスト
	nilTrades := []*models.Trade{nil}
	calculatorWithNil := NewCalculator(nilTrades)
	
	totalTradesWithNil := calculatorWithNil.CalculateTotalTrades()
	if totalTradesWithNil != 0 {
		t.Errorf("Expected 0 trades when including nil, got %d", totalTradesWithNil)
	}
}

// ヘルパー関数: テスト用取引作成
func createTrade(id string, pnl float64, timestamp time.Time) *models.Trade {
	return &models.Trade{
		ID:        id,
		Symbol:    "EURUSD",
		Side:      models.Buy,
		Size:      10000.0,
		EntryPrice: 1.1000,
		ExitPrice:  1.1000 + (pnl / 10000.0), // PnLから逆算
		PnL:       pnl,
		Status:    models.TradeClosed,
		OpenTime:  timestamp,
		CloseTime: timestamp.Add(time.Hour),
		Duration:  time.Hour,
	}
}

// ヘルパー関数: テスト用取引履歴作成
func createTestTrades() []*models.Trade {
	baseTime := time.Now()
	return []*models.Trade{
		createTrade("trade-1", 150.0, baseTime),
		createTrade("trade-2", -100.0, baseTime.Add(time.Hour)),
		createTrade("trade-3", 200.0, baseTime.Add(2*time.Hour)),
		createTrade("trade-4", -75.0, baseTime.Add(3*time.Hour)),
		createTrade("trade-5", 125.0, baseTime.Add(4*time.Hour)),
		createTrade("trade-6", 80.0, baseTime.Add(5*time.Hour)),
		createTrade("trade-7", -60.0, baseTime.Add(6*time.Hour)),
	}
}