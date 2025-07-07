package statistics

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
)

// ReportGenerator provides report generation functionality
type ReportGenerator struct{}

// NewReportGenerator creates a new report generator
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// GenerateTextReport generates a human-readable text report
func (r *ReportGenerator) GenerateTextReport(result *backtester.Result, metrics *Metrics) string {
	var report strings.Builder
	
	report.WriteString("================================================================================\n")
	report.WriteString("                           BACKTEST RESULTS                                    \n")
	report.WriteString("================================================================================\n\n")
	
	// Performance Summary
	report.WriteString("PERFORMANCE SUMMARY\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(fmt.Sprintf("Strategy Period:     %s to %s\n", 
		result.StartTime.Format("2006-01-02 15:04"), 
		result.EndTime.Format("2006-01-02 15:04")))
	report.WriteString(fmt.Sprintf("Duration:            %s\n", r.formatDuration(result.Duration)))
	report.WriteString(fmt.Sprintf("Initial Balance:     $%.2f\n", result.InitialBalance))
	report.WriteString(fmt.Sprintf("Final Balance:       $%.2f\n", result.FinalBalance))
	report.WriteString(fmt.Sprintf("Total Return:        %s\n", r.formatPercentage(metrics.TotalReturn)))
	report.WriteString(fmt.Sprintf("Annualized Return:   %s\n", r.formatPercentage(metrics.AnnualizedReturn)))
	report.WriteString(fmt.Sprintf("Total PnL:           $%.2f\n", result.TotalPnL))
	report.WriteString("\n")
	
	// Trade Statistics
	report.WriteString("TRADE STATISTICS\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(fmt.Sprintf("Total Trades:        %d\n", result.TotalTrades))
	report.WriteString(fmt.Sprintf("Winning Trades:      %d\n", result.WinningTrades))
	report.WriteString(fmt.Sprintf("Losing Trades:       %d\n", result.LosingTrades))
	report.WriteString(fmt.Sprintf("Win Rate:            %s\n", r.formatPercentage(result.WinRate)))
	report.WriteString(fmt.Sprintf("Profit Factor:       %.2f\n", metrics.ProfitFactor))
	report.WriteString(fmt.Sprintf("Average Win:         $%.2f\n", metrics.AverageWin))
	report.WriteString(fmt.Sprintf("Average Loss:        $%.2f\n", metrics.AverageLoss))
	report.WriteString(fmt.Sprintf("Largest Win:         $%.2f\n", metrics.LargestWin))
	report.WriteString(fmt.Sprintf("Largest Loss:        $%.2f\n", metrics.LargestLoss))
	report.WriteString(fmt.Sprintf("Avg Trade Duration:  %s\n", r.formatDuration(metrics.AverageTradeDuration)))
	report.WriteString("\n")
	
	// Risk Metrics
	report.WriteString("RISK METRICS\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(fmt.Sprintf("Max Drawdown:        %s\n", r.formatPercentage(result.MaxDrawdown)))
	report.WriteString(fmt.Sprintf("Sharpe Ratio:        %.2f\n", metrics.SharpeRatio))
	report.WriteString(fmt.Sprintf("Sortino Ratio:       %.2f\n", metrics.SortinoRatio))
	report.WriteString(fmt.Sprintf("Calmar Ratio:        %.2f\n", metrics.CalmarRatio))
	report.WriteString(fmt.Sprintf("VaR (95%%):           $%.2f\n", metrics.VaR95))
	report.WriteString(fmt.Sprintf("VaR (99%%):           $%.2f\n", metrics.VaR99))
	report.WriteString("\n")
	
	// Consistency Metrics
	report.WriteString("CONSISTENCY METRICS\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(fmt.Sprintf("Max Consecutive Wins:   %d\n", metrics.MaxConsecutiveWins))
	report.WriteString(fmt.Sprintf("Max Consecutive Losses: %d\n", metrics.MaxConsecutiveLosses))
	report.WriteString(fmt.Sprintf("Standard Deviation:     %.2f\n", metrics.StandardDeviation))
	report.WriteString("\n")
	
	report.WriteString("================================================================================\n")
	
	return report.String()
}

// GenerateJSONReport generates a JSON-formatted report
func (r *ReportGenerator) GenerateJSONReport(result *backtester.Result, metrics *Metrics) string {
	report := map[string]interface{}{
		"backtest_summary": map[string]interface{}{
			"start_time":      result.StartTime,
			"end_time":        result.EndTime,
			"duration_hours":  result.Duration.Hours(),
			"initial_balance": result.InitialBalance,
			"final_balance":   result.FinalBalance,
		},
		"performance_metrics": map[string]interface{}{
			"total_return_percent":     metrics.TotalReturn,
			"annualized_return_percent": metrics.AnnualizedReturn,
			"total_pnl":               result.TotalPnL,
		},
		"trade_statistics": map[string]interface{}{
			"total_trades":         result.TotalTrades,
			"winning_trades":       result.WinningTrades,
			"losing_trades":        result.LosingTrades,
			"win_rate_percent":     result.WinRate,
			"profit_factor":        metrics.ProfitFactor,
			"average_win":          metrics.AverageWin,
			"average_loss":         metrics.AverageLoss,
			"largest_win":          metrics.LargestWin,
			"largest_loss":         metrics.LargestLoss,
			"avg_trade_duration_minutes": metrics.AverageTradeDuration.Minutes(),
		},
		"risk_metrics": map[string]interface{}{
			"max_drawdown_percent": result.MaxDrawdown,
			"sharpe_ratio":         metrics.SharpeRatio,
			"sortino_ratio":        metrics.SortinoRatio,
			"calmar_ratio":         metrics.CalmarRatio,
			"var_95":               metrics.VaR95,
			"var_99":               metrics.VaR99,
			"standard_deviation":   metrics.StandardDeviation,
		},
		"consistency_metrics": map[string]interface{}{
			"max_consecutive_wins":   metrics.MaxConsecutiveWins,
			"max_consecutive_losses": metrics.MaxConsecutiveLosses,
		},
	}
	
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "Failed to generate JSON report: %s"}`, err.Error())
	}
	
	return string(jsonData)
}

// GenerateCSVReport generates a CSV report of all trades
func (r *ReportGenerator) GenerateCSVReport(trades []backtester.TradeResult) string {
	var csv strings.Builder
	
	// CSV Header
	csv.WriteString("Entry Time,Exit Time,Symbol,Side,Size,Entry Price,Exit Price,PnL,Duration (minutes)\n")
	
	// CSV Data
	for _, trade := range trades {
		side := "Buy"
		if trade.Side == 1 {
			side = "Sell"
		}
		
		csv.WriteString(fmt.Sprintf("%s,%s,%s,%s,%.2f,%.5f,%.5f,%.2f,%.1f\n",
			trade.EntryTime.Format("2006-01-02 15:04:05"),
			trade.ExitTime.Format("2006-01-02 15:04:05"),
			trade.Symbol,
			side,
			trade.Size,
			trade.EntryPrice,
			trade.ExitPrice,
			trade.PnL,
			trade.Duration.Minutes(),
		))
	}
	
	return csv.String()
}

// GenerateDetailedReport generates a comprehensive detailed report
func (r *ReportGenerator) GenerateDetailedReport(result *backtester.Result, metrics *Metrics) string {
	var report strings.Builder
	
	report.WriteString("================================================================================\n")
	report.WriteString("                        DETAILED BACKTEST ANALYSIS                            \n")
	report.WriteString("================================================================================\n\n")
	
	// Executive Summary
	report.WriteString("EXECUTIVE SUMMARY\n")
	report.WriteString("------------------------------------------\n")
	summary := r.generateExecutiveSummary(result, metrics)
	report.WriteString(summary)
	report.WriteString("\n\n")
	
	// Performance Analysis
	report.WriteString("PERFORMANCE ANALYSIS\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(r.generatePerformanceAnalysis(result, metrics))
	report.WriteString("\n\n")
	
	// Risk Analysis
	report.WriteString("RISK ANALYSIS\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(r.generateRiskAnalysis(metrics))
	report.WriteString("\n\n")
	
	// Trade Distribution
	report.WriteString("TRADE DISTRIBUTION\n")
	report.WriteString("------------------------------------------\n")
	report.WriteString(r.generateTradeDistribution(result.Trades))
	report.WriteString("\n\n")
	
	// Monthly Performance (if applicable)
	if result.Duration >= 30*24*time.Hour {
		report.WriteString("MONTHLY PERFORMANCE\n")
		report.WriteString("------------------------------------------\n")
		report.WriteString(r.generateMonthlyPerformance(result.Trades))
		report.WriteString("\n\n")
	}
	
	report.WriteString("================================================================================\n")
	
	return report.String()
}

// Helper methods for detailed report generation

func (r *ReportGenerator) generateExecutiveSummary(result *backtester.Result, metrics *Metrics) string {
	var summary strings.Builder
	
	// Performance assessment
	if metrics.TotalReturn > 0 {
		summary.WriteString("Strategy Performance: POSITIVE\n")
	} else {
		summary.WriteString("Strategy Performance: NEGATIVE\n")
	}
	
	// Risk assessment
	if metrics.SharpeRatio > 1.0 {
		summary.WriteString("Risk-Adjusted Returns: EXCELLENT (Sharpe > 1.0)\n")
	} else if metrics.SharpeRatio > 0.5 {
		summary.WriteString("Risk-Adjusted Returns: GOOD (Sharpe > 0.5)\n")
	} else if metrics.SharpeRatio > 0 {
		summary.WriteString("Risk-Adjusted Returns: FAIR (Sharpe > 0)\n")
	} else {
		summary.WriteString("Risk-Adjusted Returns: POOR (Sharpe <= 0)\n")
	}
	
	// Win rate assessment
	if result.WinRate > 60 {
		summary.WriteString("Trade Accuracy: HIGH (Win Rate > 60%)\n")
	} else if result.WinRate > 40 {
		summary.WriteString("Trade Accuracy: MODERATE (Win Rate 40-60%)\n")
	} else {
		summary.WriteString("Trade Accuracy: LOW (Win Rate < 40%)\n")
	}
	
	return summary.String()
}

func (r *ReportGenerator) generatePerformanceAnalysis(result *backtester.Result, metrics *Metrics) string {
	var analysis strings.Builder
	
	analysis.WriteString(fmt.Sprintf("Total Return: %s over %s\n", 
		r.formatPercentage(metrics.TotalReturn), r.formatDuration(result.Duration)))
	analysis.WriteString(fmt.Sprintf("Compound Annual Growth Rate: %s\n", 
		r.formatPercentage(metrics.AnnualizedReturn)))
	analysis.WriteString(fmt.Sprintf("Profit Factor: %.2f (Gross Profit / Gross Loss)\n", metrics.ProfitFactor))
	
	if metrics.ProfitFactor > 1.5 {
		analysis.WriteString("  → Excellent profit factor (> 1.5)\n")
	} else if metrics.ProfitFactor > 1.0 {
		analysis.WriteString("  → Acceptable profit factor (> 1.0)\n")
	} else {
		analysis.WriteString("  → Poor profit factor (< 1.0) - Strategy loses money\n")
	}
	
	return analysis.String()
}

func (r *ReportGenerator) generateRiskAnalysis(metrics *Metrics) string {
	var analysis strings.Builder
	
	analysis.WriteString(fmt.Sprintf("Maximum Drawdown: %s\n", r.formatPercentage(metrics.MaxDrawdown)))
	if metrics.MaxDrawdown > 20 {
		analysis.WriteString("  → HIGH RISK: Drawdown > 20%\n")
	} else if metrics.MaxDrawdown > 10 {
		analysis.WriteString("  → MODERATE RISK: Drawdown 10-20%\n")
	} else {
		analysis.WriteString("  → LOW RISK: Drawdown < 10%\n")
	}
	
	analysis.WriteString(fmt.Sprintf("Value at Risk (95%%): $%.2f\n", metrics.VaR95))
	analysis.WriteString(fmt.Sprintf("Sortino Ratio: %.2f (Focuses on downside risk)\n", metrics.SortinoRatio))
	
	return analysis.String()
}

func (r *ReportGenerator) generateTradeDistribution(trades []backtester.TradeResult) string {
	var distribution strings.Builder
	
	if len(trades) == 0 {
		return "No trades to analyze.\n"
	}
	
	// Calculate PnL distribution
	winCount := 0
	lossCount := 0
	breakEvenCount := 0
	
	for _, trade := range trades {
		if trade.PnL > 0 {
			winCount++
		} else if trade.PnL < 0 {
			lossCount++
		} else {
			breakEvenCount++
		}
	}
	
	distribution.WriteString(fmt.Sprintf("Winning Trades: %d (%.1f%%)\n", 
		winCount, float64(winCount)/float64(len(trades))*100))
	distribution.WriteString(fmt.Sprintf("Losing Trades: %d (%.1f%%)\n", 
		lossCount, float64(lossCount)/float64(len(trades))*100))
	distribution.WriteString(fmt.Sprintf("Break-even Trades: %d (%.1f%%)\n", 
		breakEvenCount, float64(breakEvenCount)/float64(len(trades))*100))
	
	return distribution.String()
}

func (r *ReportGenerator) generateMonthlyPerformance(trades []backtester.TradeResult) string {
	var performance strings.Builder
	
	// Group trades by month
	monthlyPnL := make(map[string]float64)
	for _, trade := range trades {
		month := trade.EntryTime.Format("2006-01")
		monthlyPnL[month] += trade.PnL
	}
	
	performance.WriteString("Month-by-Month Performance:\n")
	for month, pnl := range monthlyPnL {
		performance.WriteString(fmt.Sprintf("  %s: $%.2f\n", month, pnl))
	}
	
	return performance.String()
}

// Utility formatting methods

func (r *ReportGenerator) formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func (r *ReportGenerator) formatPercentage(value float64) string {
	return fmt.Sprintf("%.2f%%", value)
}