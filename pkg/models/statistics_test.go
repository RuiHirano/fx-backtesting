package models

import (
	"math"
	"testing"
	"time"
)

// abs は絶対値を返すヘルパー関数
func abs(x float64) float64 {
	return math.Abs(x)
}

// TestNewStatistics は新しい統計情報の作成をテスト
func TestNewStatistics(t *testing.T) {
	t.Run("should create new statistics with initial balance", func(t *testing.T) {
		initialBalance := 10000.0
		stats := NewStatistics(initialBalance)
		
		if stats == nil {
			t.Error("Expected statistics to be created, got nil")
		}
		
		if stats.InitialBalance != initialBalance {
			t.Errorf("Expected initial balance to be %f, got %f", initialBalance, stats.InitialBalance)
		}
		
		if stats.CurrentBalance != initialBalance {
			t.Errorf("Expected current balance to be %f, got %f", initialBalance, stats.CurrentBalance)
		}
		
		if stats.TotalTrades != 0 {
			t.Errorf("Expected total trades to be 0, got %d", stats.TotalTrades)
		}
		
		if stats.WinningTrades != 0 {
			t.Errorf("Expected winning trades to be 0, got %d", stats.WinningTrades)
		}
		
		if stats.LosingTrades != 0 {
			t.Errorf("Expected losing trades to be 0, got %d", stats.LosingTrades)
		}
	})
}

// TestUpdateBalance は残高更新をテスト
func TestUpdateBalance(t *testing.T) {
	t.Run("should update balance and net profit", func(t *testing.T) {
		stats := NewStatistics(10000.0)
		newBalance := 11000.0
		
		stats.UpdateBalance(newBalance)
		
		if stats.CurrentBalance != newBalance {
			t.Errorf("Expected current balance to be %f, got %f", newBalance, stats.CurrentBalance)
		}
		
		expectedNetProfit := newBalance - 10000.0
		if stats.NetProfit != expectedNetProfit {
			t.Errorf("Expected net profit to be %f, got %f", expectedNetProfit, stats.NetProfit)
		}
	})
}

// TestAddTrade は取引追加をテスト
func TestAddTrade(t *testing.T) {
	t.Run("should add winning trade", func(t *testing.T) {
		stats := NewStatistics(10000.0)
		profit := 100.0
		
		stats.AddTrade(profit)
		
		if stats.TotalTrades != 1 {
			t.Errorf("Expected total trades to be 1, got %d", stats.TotalTrades)
		}
		
		if stats.WinningTrades != 1 {
			t.Errorf("Expected winning trades to be 1, got %d", stats.WinningTrades)
		}
		
		if stats.LosingTrades != 0 {
			t.Errorf("Expected losing trades to be 0, got %d", stats.LosingTrades)
		}
		
		if stats.TotalProfit != profit {
			t.Errorf("Expected total profit to be %f, got %f", profit, stats.TotalProfit)
		}
		
		if stats.WinRate != 100.0 {
			t.Errorf("Expected win rate to be 100.0, got %f", stats.WinRate)
		}
	})
	
	t.Run("should add losing trade", func(t *testing.T) {
		stats := NewStatistics(10000.0)
		loss := -50.0
		
		stats.AddTrade(loss)
		
		if stats.TotalTrades != 1 {
			t.Errorf("Expected total trades to be 1, got %d", stats.TotalTrades)
		}
		
		if stats.WinningTrades != 0 {
			t.Errorf("Expected winning trades to be 0, got %d", stats.WinningTrades)
		}
		
		if stats.LosingTrades != 1 {
			t.Errorf("Expected losing trades to be 1, got %d", stats.LosingTrades)
		}
		
		if stats.TotalLoss != loss {
			t.Errorf("Expected total loss to be %f, got %f", loss, stats.TotalLoss)
		}
		
		if stats.WinRate != 0.0 {
			t.Errorf("Expected win rate to be 0.0, got %f", stats.WinRate)
		}
	})
	
	t.Run("should calculate metrics correctly", func(t *testing.T) {
		stats := NewStatistics(10000.0)
		
		// 勝ち取引を追加
		stats.AddTrade(100.0)
		stats.AddTrade(200.0)
		
		// 負け取引を追加
		stats.AddTrade(-50.0)
		
		if stats.TotalTrades != 3 {
			t.Errorf("Expected total trades to be 3, got %d", stats.TotalTrades)
		}
		
		if stats.WinningTrades != 2 {
			t.Errorf("Expected winning trades to be 2, got %d", stats.WinningTrades)
		}
		
		if stats.LosingTrades != 1 {
			t.Errorf("Expected losing trades to be 1, got %d", stats.LosingTrades)
		}
		
		expectedWinRate := 2.0 / 3.0 * 100.0
		if abs(stats.WinRate - expectedWinRate) > 0.0001 {
			t.Errorf("Expected win rate to be %f, got %f", expectedWinRate, stats.WinRate)
		}
		
		expectedAverageWin := 300.0 / 2.0
		if stats.AverageWin != expectedAverageWin {
			t.Errorf("Expected average win to be %f, got %f", expectedAverageWin, stats.AverageWin)
		}
		
		expectedAverageLoss := -50.0 / 1.0
		if stats.AverageLoss != expectedAverageLoss {
			t.Errorf("Expected average loss to be %f, got %f", expectedAverageLoss, stats.AverageLoss)
		}
		
		expectedProfitFactor := 300.0 / 50.0
		if stats.ProfitFactor != expectedProfitFactor {
			t.Errorf("Expected profit factor to be %f, got %f", expectedProfitFactor, stats.ProfitFactor)
		}
	})
}

// TestLastUpdated は最終更新時刻をテスト
func TestLastUpdated(t *testing.T) {
	t.Run("should update last updated time", func(t *testing.T) {
		stats := NewStatistics(10000.0)
		initialTime := stats.LastUpdated
		
		// 少し待つ
		time.Sleep(10 * time.Millisecond)
		
		stats.UpdateBalance(11000.0)
		
		if !stats.LastUpdated.After(initialTime) {
			t.Error("Expected last updated time to be updated")
		}
	})
}