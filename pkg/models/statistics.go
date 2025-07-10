package models

import "time"

// Statistics はバックテストの統計情報を表す構造体
type Statistics struct {
	// 基本統計
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	TotalTrades      int       `json:"total_trades"`
	WinningTrades    int       `json:"winning_trades"`
	LosingTrades     int       `json:"losing_trades"`
	
	// 収益統計
	InitialBalance   float64   `json:"initial_balance"`
	CurrentBalance   float64   `json:"current_balance"`
	TotalProfit      float64   `json:"total_profit"`
	TotalLoss        float64   `json:"total_loss"`
	NetProfit        float64   `json:"net_profit"`
	
	// パフォーマンス指標
	WinRate          float64   `json:"win_rate"`
	ProfitFactor     float64   `json:"profit_factor"`
	MaxDrawdown      float64   `json:"max_drawdown"`
	MaxDrawdownPct   float64   `json:"max_drawdown_pct"`
	
	// 平均値
	AverageWin       float64   `json:"average_win"`
	AverageLoss      float64   `json:"average_loss"`
	AverageProfit    float64   `json:"average_profit"`
	
	// 連続記録
	MaxConsecutiveWins   int   `json:"max_consecutive_wins"`
	MaxConsecutiveLosses int   `json:"max_consecutive_losses"`
	
	// 更新時刻
	LastUpdated      time.Time `json:"last_updated"`
}

// NewStatistics は新しい統計情報を作成
func NewStatistics(initialBalance float64) *Statistics {
	return &Statistics{
		StartTime:        time.Now(),
		InitialBalance:   initialBalance,
		CurrentBalance:   initialBalance,
		LastUpdated:      time.Now(),
	}
}

// UpdateBalance は残高を更新
func (s *Statistics) UpdateBalance(newBalance float64) {
	s.CurrentBalance = newBalance
	s.NetProfit = s.CurrentBalance - s.InitialBalance
	s.LastUpdated = time.Now()
}

// AddTrade は取引を統計に追加
func (s *Statistics) AddTrade(profit float64) {
	s.TotalTrades++
	
	if profit > 0 {
		s.WinningTrades++
		s.TotalProfit += profit
	} else {
		s.LosingTrades++
		s.TotalLoss += profit
	}
	
	s.calculateMetrics()
	s.LastUpdated = time.Now()
}

// calculateMetrics は各種指標を計算
func (s *Statistics) calculateMetrics() {
	// 勝率計算
	if s.TotalTrades > 0 {
		s.WinRate = float64(s.WinningTrades) / float64(s.TotalTrades) * 100
	}
	
	// プロフィットファクター計算
	if s.TotalLoss != 0 {
		s.ProfitFactor = s.TotalProfit / (-s.TotalLoss)
	}
	
	// 平均値計算
	if s.WinningTrades > 0 {
		s.AverageWin = s.TotalProfit / float64(s.WinningTrades)
	}
	
	if s.LosingTrades > 0 {
		s.AverageLoss = s.TotalLoss / float64(s.LosingTrades)
	}
	
	if s.TotalTrades > 0 {
		s.AverageProfit = s.NetProfit / float64(s.TotalTrades)
	}
}