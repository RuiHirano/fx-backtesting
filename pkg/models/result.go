package models

import (
	"time"
)

// BacktestResult はバックテストの結果を表します。
type BacktestResult struct {
	// 基本情報
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	
	// 残高情報
	InitialBalance float64 `json:"initial_balance"`
	FinalBalance   float64 `json:"final_balance"`
	TotalPnL       float64 `json:"total_pnl"`
	TotalReturn    float64 `json:"total_return_percent"`
	
	// 取引統計
	TotalTrades   int     `json:"total_trades"`
	WinningTrades int     `json:"winning_trades"`
	LosingTrades  int     `json:"losing_trades"`
	WinRate       float64 `json:"win_rate_percent"`
	
	// 損益統計
	GrossProfit float64 `json:"gross_profit"`
	GrossLoss   float64 `json:"gross_loss"`
	AverageWin  float64 `json:"average_win"`
	AverageLoss float64 `json:"average_loss"`
	LargestWin  float64 `json:"largest_win"`
	LargestLoss float64 `json:"largest_loss"`
	
	// リスク指標
	MaxDrawdown     float64   `json:"max_drawdown"`
	MaxDrawdownDate time.Time `json:"max_drawdown_date"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	ProfitFactor    float64   `json:"profit_factor"`
	
	// 取引履歴
	TradeHistory []Trade `json:"trade_history"`
	
	// 日次統計
	DailyReturns []DailyReturn `json:"daily_returns,omitempty"`
}

// DailyReturn は日次リターンを表します。
type DailyReturn struct {
	Date    time.Time `json:"date"`
	Return  float64   `json:"return"`
	PnL     float64   `json:"pnl"`
	Balance float64   `json:"balance"`
}

// NewBacktestResult は新しいバックテスト結果を作成します。
func NewBacktestResult(initialBalance float64) *BacktestResult {
	return &BacktestResult{
		StartTime:      time.Now(),
		InitialBalance: initialBalance,
		FinalBalance:   initialBalance,
		TradeHistory:   make([]Trade, 0),
		DailyReturns:   make([]DailyReturn, 0),
	}
}

// AddTrade は取引を結果に追加します。
func (br *BacktestResult) AddTrade(trade Trade) {
	br.TradeHistory = append(br.TradeHistory, trade)
	br.updateStatistics()
}

// updateStatistics は統計情報を更新します。
func (br *BacktestResult) updateStatistics() {
	br.TotalTrades = len(br.TradeHistory)
	
	if br.TotalTrades == 0 {
		return
	}
	
	var totalPnL, grossProfit, grossLoss float64
	var winningTrades, losingTrades int
	var largestWin, largestLoss float64

	for _, trade := range br.TradeHistory {
		totalPnL += trade.PnL
		
		if trade.IsWinning() {
			winningTrades++
			grossProfit += trade.PnL
			if trade.PnL > largestWin {
				largestWin = trade.PnL
			}
		} else if trade.IsLosing() {
			losingTrades++
			grossLoss += -trade.PnL
			if trade.PnL < largestLoss {
				largestLoss = trade.PnL
			}
		}
	}
	
	br.TotalPnL = totalPnL
	br.FinalBalance = br.InitialBalance + totalPnL
	br.TotalReturn = (totalPnL / br.InitialBalance) * 100
	
	br.WinningTrades = winningTrades
	br.LosingTrades = losingTrades
	br.WinRate = float64(winningTrades) / float64(br.TotalTrades) * 100
	
	br.GrossProfit = grossProfit
	br.GrossLoss = grossLoss
	br.LargestWin = largestWin
	br.LargestLoss = largestLoss
	
	if winningTrades > 0 {
		br.AverageWin = grossProfit / float64(winningTrades)
	}
	if losingTrades > 0 {
		br.AverageLoss = grossLoss / float64(losingTrades)
	}
	if grossLoss > 0 {
		br.ProfitFactor = grossProfit / grossLoss
	}
}

// Finalize はバックテスト結果を確定します。
func (br *BacktestResult) Finalize() {
	br.EndTime = time.Now()
	br.Duration = br.EndTime.Sub(br.StartTime)
	br.updateStatistics()
}

// GetSummary は結果の要約を返します。
func (br *BacktestResult) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"total_return":   br.TotalReturn,
		"total_trades":   br.TotalTrades,
		"win_rate":       br.WinRate,
		"profit_factor":  br.ProfitFactor,
		"max_drawdown":   br.MaxDrawdown,
		"sharpe_ratio":   br.SharpeRatio,
	}
}