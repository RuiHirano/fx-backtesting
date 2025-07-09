package statistics

import (
	"fmt"
	"strings"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// ReportFormat はレポートフォーマットを表します。
type ReportFormat int

const (
	FormatText ReportFormat = iota
	FormatJSON
	FormatCSV
)

// Report はバックテスト結果のレポート生成機能を提供します。
type Report struct {
	calculator *Calculator
	result     *models.BacktestResult
}

// NewReport は新しいReportを作成します。
func NewReport(trades []*models.Trade, initialBalance float64) *Report {
	calculator := NewCalculator(trades)
	
	// BacktestResultを作成して統計情報を設定
	result := models.NewBacktestResult(initialBalance)
	for _, trade := range trades {
		result.AddTrade(*trade)
	}
	result.Finalize()
	
	// 高度な統計指標を設定
	result.MaxDrawdown = calculator.CalculateMaxDrawdown()
	result.SharpeRatio = calculator.CalculateSharpeRatio()
	
	return &Report{
		calculator: calculator,
		result:     result,
	}
}

// GenerateTextReport はテキスト形式のレポートを生成します。
func (r *Report) GenerateTextReport() string {
	var sb strings.Builder
	
	sb.WriteString("=== バックテスト結果レポート ===\n\n")
	
	// 基本情報
	sb.WriteString("【基本情報】\n")
	sb.WriteString(fmt.Sprintf("期間: %s ～ %s\n", 
		r.result.StartTime.Format("2006-01-02 15:04:05"),
		r.result.EndTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("実行時間: %v\n", r.result.Duration))
	sb.WriteString(fmt.Sprintf("初期残高: %.2f\n", r.result.InitialBalance))
	sb.WriteString(fmt.Sprintf("最終残高: %.2f\n", r.result.FinalBalance))
	sb.WriteString("\n")
	
	// 損益情報
	sb.WriteString("【損益情報】\n")
	sb.WriteString(fmt.Sprintf("総損益: %.2f\n", r.result.TotalPnL))
	sb.WriteString(fmt.Sprintf("総リターン: %.2f%%\n", r.result.TotalReturn))
	sb.WriteString(fmt.Sprintf("総利益: %.2f\n", r.result.GrossProfit))
	sb.WriteString(fmt.Sprintf("総損失: %.2f\n", r.result.GrossLoss))
	sb.WriteString(fmt.Sprintf("最大利益: %.2f\n", r.result.LargestWin))
	sb.WriteString(fmt.Sprintf("最大損失: %.2f\n", r.result.LargestLoss))
	sb.WriteString("\n")
	
	// 取引統計
	sb.WriteString("【取引統計】\n")
	sb.WriteString(fmt.Sprintf("総取引数: %d\n", r.result.TotalTrades))
	sb.WriteString(fmt.Sprintf("勝ち取引: %d\n", r.result.WinningTrades))
	sb.WriteString(fmt.Sprintf("負け取引: %d\n", r.result.LosingTrades))
	sb.WriteString(fmt.Sprintf("勝率: %.2f%%\n", r.result.WinRate))
	sb.WriteString(fmt.Sprintf("平均利益: %.2f\n", r.result.AverageWin))
	sb.WriteString(fmt.Sprintf("平均損失: %.2f\n", r.result.AverageLoss))
	sb.WriteString("\n")
	
	// リスク指標
	sb.WriteString("【リスク指標】\n")
	sb.WriteString(fmt.Sprintf("最大ドローダウン: %.2f\n", r.result.MaxDrawdown))
	sb.WriteString(fmt.Sprintf("シャープレシオ: %.4f\n", r.result.SharpeRatio))
	sb.WriteString(fmt.Sprintf("プロフィットファクター: %.4f\n", r.result.ProfitFactor))
	sb.WriteString(fmt.Sprintf("ソルティノレシオ: %.4f\n", r.calculator.CalculateSortinoRatio()))
	sb.WriteString(fmt.Sprintf("カルマーレシオ: %.4f\n", r.calculator.CalculateCalmarRatio()))
	sb.WriteString("\n")
	
	// 取引パフォーマンス
	sb.WriteString("【取引パフォーマンス】\n")
	avgHolding := r.calculator.CalculateAverageHoldingPeriod()
	sb.WriteString(fmt.Sprintf("平均保有期間: %.2f時間\n", avgHolding.Hours()))
	sb.WriteString(fmt.Sprintf("最大連勝: %d\n", r.calculator.CalculateMaxConsecutiveWins()))
	sb.WriteString(fmt.Sprintf("最大連敗: %d\n", r.calculator.CalculateMaxConsecutiveLosses()))
	sb.WriteString(fmt.Sprintf("取引頻度: %.2f取引/日\n", r.calculator.CalculateTradingFrequency()))
	sb.WriteString(fmt.Sprintf("リスクリワード比: %.4f\n", r.calculator.CalculateRiskRewardRatio()))
	sb.WriteString("\n")
	
	return sb.String()
}

// GenerateJSONReport はJSON形式のレポートを生成します。
func (r *Report) GenerateJSONReport() string {
	// 基本的なJSON形式のレポート
	return fmt.Sprintf(`{
  "summary": {
    "initial_balance": %.2f,
    "final_balance": %.2f,
    "total_pnl": %.2f,
    "total_return": %.2f,
    "total_trades": %d,
    "win_rate": %.2f,
    "profit_factor": %.4f,
    "max_drawdown": %.2f,
    "sharpe_ratio": %.4f
  },
  "detailed_metrics": {
    "gross_profit": %.2f,
    "gross_loss": %.2f,
    "largest_win": %.2f,
    "largest_loss": %.2f,
    "average_win": %.2f,
    "average_loss": %.2f,
    "max_consecutive_wins": %d,
    "max_consecutive_losses": %d,
    "sortino_ratio": %.4f,
    "calmar_ratio": %.4f,
    "risk_reward_ratio": %.4f,
    "trading_frequency": %.2f,
    "average_holding_hours": %.2f
  }
}`,
		r.result.InitialBalance,
		r.result.FinalBalance,
		r.result.TotalPnL,
		r.result.TotalReturn,
		r.result.TotalTrades,
		r.result.WinRate,
		r.result.ProfitFactor,
		r.result.MaxDrawdown,
		r.result.SharpeRatio,
		r.result.GrossProfit,
		r.result.GrossLoss,
		r.result.LargestWin,
		r.result.LargestLoss,
		r.result.AverageWin,
		r.result.AverageLoss,
		r.calculator.CalculateMaxConsecutiveWins(),
		r.calculator.CalculateMaxConsecutiveLosses(),
		r.calculator.CalculateSortinoRatio(),
		r.calculator.CalculateCalmarRatio(),
		r.calculator.CalculateRiskRewardRatio(),
		r.calculator.CalculateTradingFrequency(),
		r.calculator.CalculateAverageHoldingPeriod().Hours(),
	)
}

// GenerateCSVReport はCSV形式の取引履歴レポートを生成します。
func (r *Report) GenerateCSVReport() string {
	var sb strings.Builder
	
	// ヘッダー
	sb.WriteString("ID,Symbol,Side,Size,EntryPrice,ExitPrice,PnL,Status,OpenTime,CloseTime,DurationHours\n")
	
	// 取引履歴
	for _, trade := range r.calculator.GetTrades() {
		record := trade.ToCSVRecord()
		sb.WriteString(strings.Join(record, ","))
		sb.WriteString("\n")
	}
	
	return sb.String()
}

// GenerateReport は指定されたフォーマットでレポートを生成します。
func (r *Report) GenerateReport(format ReportFormat) string {
	switch format {
	case FormatJSON:
		return r.GenerateJSONReport()
	case FormatCSV:
		return r.GenerateCSVReport()
	default:
		return r.GenerateTextReport()
	}
}

// GetSummaryMetrics は要約メトリクスを取得します。
func (r *Report) GetSummaryMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_return":         r.result.TotalReturn,
		"total_trades":         r.result.TotalTrades,
		"win_rate":             r.result.WinRate,
		"profit_factor":        r.result.ProfitFactor,
		"max_drawdown":         r.result.MaxDrawdown,
		"sharpe_ratio":         r.result.SharpeRatio,
		"sortino_ratio":        r.calculator.CalculateSortinoRatio(),
		"calmar_ratio":         r.calculator.CalculateCalmarRatio(),
		"risk_reward_ratio":    r.calculator.CalculateRiskRewardRatio(),
		"max_consecutive_wins": r.calculator.CalculateMaxConsecutiveWins(),
		"max_consecutive_losses": r.calculator.CalculateMaxConsecutiveLosses(),
		"trading_frequency":    r.calculator.CalculateTradingFrequency(),
		"average_holding_period": r.calculator.CalculateAverageHoldingPeriod().Hours(),
	}
}

// GenerateCompactSummary は簡潔な要約を生成します。
func (r *Report) GenerateCompactSummary() string {
	return fmt.Sprintf(
		"リターン: %.2f%% | 取引数: %d | 勝率: %.1f%% | PF: %.2f | DD: %.2f | SR: %.2f",
		r.result.TotalReturn,
		r.result.TotalTrades,
		r.result.WinRate,
		r.result.ProfitFactor,
		r.result.MaxDrawdown,
		r.result.SharpeRatio,
	)
}