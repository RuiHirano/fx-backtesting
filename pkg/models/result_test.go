package models

import "testing"

// BacktestResult構造体のテスト
func TestBacktestResult_NewBacktestResult(t *testing.T) {
	initialBalance := 10000.0
	result := NewBacktestResult(initialBalance)
	
	if result.InitialBalance != initialBalance {
		t.Errorf("Expected initial balance %f, got %f", initialBalance, result.InitialBalance)
	}
	
	if result.FinalBalance != initialBalance {
		t.Errorf("Expected final balance %f, got %f", initialBalance, result.FinalBalance)
	}
	
	if len(result.TradeHistory) != 0 {
		t.Errorf("Expected empty trade history, got %d trades", len(result.TradeHistory))
	}
}

func TestBacktestResult_AddTrade(t *testing.T) {
	result := NewBacktestResult(10000.0)
	
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	trade := NewTradeFromPosition(position, 1.1010)
	
	result.AddTrade(*trade)
	
	if len(result.TradeHistory) != 1 {
		t.Errorf("Expected 1 trade in history, got %d", len(result.TradeHistory))
	}
	
	if result.TotalTrades != 1 {
		t.Errorf("Expected total trades 1, got %d", result.TotalTrades)
	}
	
	if result.WinningTrades != 1 {
		t.Errorf("Expected winning trades 1, got %d", result.WinningTrades)
	}
	
	if result.TotalPnL != trade.PnL {
		t.Errorf("Expected total PnL %f, got %f", trade.PnL, result.TotalPnL)
	}
}

func TestBacktestResult_UpdateStatistics(t *testing.T) {
	result := NewBacktestResult(10000.0)
	
	// 勝ち取引を追加
	position1 := NewPosition("pos-1", "EURUSD", Buy, 10000.0, 1.1000)
	trade1 := NewTradeFromPosition(position1, 1.1010)
	result.AddTrade(*trade1)
	
	// 負け取引を追加
	position2 := NewPosition("pos-2", "EURUSD", Buy, 10000.0, 1.1020)
	trade2 := NewTradeFromPosition(position2, 1.1000)
	result.AddTrade(*trade2)
	
	// 統計値の確認
	if result.TotalTrades != 2 {
		t.Errorf("Expected total trades 2, got %d", result.TotalTrades)
	}
	
	if result.WinningTrades != 1 {
		t.Errorf("Expected winning trades 1, got %d", result.WinningTrades)
	}
	
	if result.LosingTrades != 1 {
		t.Errorf("Expected losing trades 1, got %d", result.LosingTrades)
	}
	
	expectedWinRate := float64(1) / float64(2) * 100
	if result.WinRate != expectedWinRate {
		t.Errorf("Expected win rate %f, got %f", expectedWinRate, result.WinRate)
	}
	
	expectedTotalPnL := trade1.PnL + trade2.PnL
	if result.TotalPnL != expectedTotalPnL {
		t.Errorf("Expected total PnL %f, got %f", expectedTotalPnL, result.TotalPnL)
	}
	
	expectedFinalBalance := result.InitialBalance + expectedTotalPnL
	if result.FinalBalance != expectedFinalBalance {
		t.Errorf("Expected final balance %f, got %f", expectedFinalBalance, result.FinalBalance)
	}
}

func TestBacktestResult_Finalize(t *testing.T) {
	result := NewBacktestResult(10000.0)
	
	// Finalizeを呼び出し
	result.Finalize()
	
	// EndTimeが設定されていることを確認
	if result.EndTime.IsZero() {
		t.Error("Expected EndTime to be set after finalize")
	}
	
	// Durationが正の値であることを確認
	if result.Duration <= 0 {
		t.Error("Expected positive duration after finalize")
	}
}

func TestBacktestResult_GetSummary(t *testing.T) {
	result := NewBacktestResult(10000.0)
	
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	trade := NewTradeFromPosition(position, 1.1010)
	result.AddTrade(*trade)
	
	summary := result.GetSummary()
	
	// サマリーのキーが存在することを確認
	expectedKeys := []string{
		"total_return", "total_trades", "win_rate",
		"profit_factor", "max_drawdown", "sharpe_ratio",
	}
	
	for _, key := range expectedKeys {
		if _, exists := summary[key]; !exists {
			t.Errorf("Expected key %s in summary", key)
		}
	}
	
	// 値の確認
	if summary["total_trades"] != result.TotalTrades {
		t.Errorf("Expected total trades %d, got %v", result.TotalTrades, summary["total_trades"])
	}
	
	if summary["win_rate"] != result.WinRate {
		t.Errorf("Expected win rate %f, got %v", result.WinRate, summary["win_rate"])
	}
}