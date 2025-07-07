package statistics

import (
	"math"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
)

// Metrics represents comprehensive trading statistics
type Metrics struct {
	// Basic Performance
	TotalReturn    float64 // Total return percentage
	AnnualizedReturn float64 // Annualized return percentage
	TotalPnL       float64 // Total profit/loss
	
	// Risk Metrics
	MaxDrawdown    float64 // Maximum drawdown percentage
	SharpeRatio    float64 // Sharpe ratio
	SortinoRatio   float64 // Sortino ratio
	CalmarRatio    float64 // Calmar ratio (Annual return / Max drawdown)
	
	// Trade Statistics
	TotalTrades    int     // Total number of trades
	WinningTrades  int     // Number of winning trades
	LosingTrades   int     // Number of losing trades
	WinRate        float64 // Win rate percentage
	ProfitFactor   float64 // Gross profit / Gross loss
	
	// Trade Analysis
	AverageWin     float64 // Average winning trade
	AverageLoss    float64 // Average losing trade
	LargestWin     float64 // Largest winning trade
	LargestLoss    float64 // Largest losing trade
	
	// Consistency Metrics
	MaxConsecutiveWins   int           // Maximum consecutive wins
	MaxConsecutiveLosses int           // Maximum consecutive losses
	AverageTradeDuration time.Duration // Average trade duration
	
	// Additional Statistics
	StandardDeviation float64 // Standard deviation of returns
	VaR95            float64 // 95% Value at Risk
	VaR99            float64 // 99% Value at Risk
}

// Calculator provides statistical calculation functionality
type Calculator struct{}

// NewCalculator creates a new statistics calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateMetrics calculates comprehensive trading metrics from backtest results
func (c *Calculator) CalculateMetrics(result *backtester.Result) *Metrics {
	if result == nil {
		return &Metrics{}
	}
	
	// Initialize basic metrics even if no trades
	metrics := &Metrics{
		TotalPnL:      result.TotalPnL,
		TotalTrades:   result.TotalTrades,
		WinningTrades: result.WinningTrades,
		LosingTrades:  result.LosingTrades,
		WinRate:       result.WinRate,
		MaxDrawdown:   result.MaxDrawdown,
	}
	
	// Calculate total return
	if result.InitialBalance > 0 {
		metrics.TotalReturn = ((result.FinalBalance - result.InitialBalance) / result.InitialBalance) * 100
	}
	
	// If no trades, return early with basic metrics
	if result.TotalTrades == 0 {
		return metrics
	}

	// Additional calculations for when we have trades
	// Note: TotalReturn already calculated above

	// Calculate annualized return
	if result.Duration > 0 {
		yearsElapsed := result.Duration.Hours() / (24 * 365)
		if yearsElapsed > 0 {
			metrics.AnnualizedReturn = (math.Pow(result.FinalBalance/result.InitialBalance, 1/yearsElapsed) - 1) * 100
		}
	}

	// Calculate trade-based metrics
	metrics.ProfitFactor = c.CalculateProfitFactor(result.Trades)
	metrics.SharpeRatio = c.CalculateSharpeRatio(result.Trades, 0.02) // Assume 2% risk-free rate
	metrics.AverageWin = c.CalculateAverageWin(result.Trades)
	metrics.AverageLoss = c.CalculateAverageLoss(result.Trades)
	metrics.LargestWin = c.CalculateLargestWin(result.Trades)
	metrics.LargestLoss = c.CalculateLargestLoss(result.Trades)
	metrics.MaxConsecutiveWins = c.CalculateMaxConsecutiveWins(result.Trades)
	metrics.MaxConsecutiveLosses = c.CalculateMaxConsecutiveLosses(result.Trades)
	metrics.AverageTradeDuration = c.CalculateAverageTradeDuration(result.Trades)

	// Calculate risk metrics
	metrics.StandardDeviation = c.CalculateStandardDeviation(result.Trades)
	metrics.SortinoRatio = c.CalculateSortinoRatio(result.Trades, 0.02)
	
	if metrics.MaxDrawdown > 0 {
		metrics.CalmarRatio = metrics.AnnualizedReturn / metrics.MaxDrawdown
	}

	// Calculate VaR
	metrics.VaR95 = c.CalculateVaR(result.Trades, 0.05)
	metrics.VaR99 = c.CalculateVaR(result.Trades, 0.01)

	return metrics
}

// CalculateSharpeRatio calculates the Sharpe ratio
func (c *Calculator) CalculateSharpeRatio(trades []backtester.TradeResult, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	// Calculate average return
	totalReturn := 0.0
	for _, trade := range trades {
		totalReturn += trade.PnL
	}
	avgReturn := totalReturn / float64(len(trades))

	// Calculate standard deviation
	stdDev := c.CalculateStandardDeviation(trades)
	
	if stdDev == 0 {
		return 0.0
	}

	// Convert risk-free rate to per-trade basis (assuming daily trades)
	dailyRiskFreeRate := riskFreeRate / 365
	
	return (avgReturn - dailyRiskFreeRate) / stdDev
}

// CalculateSortinoRatio calculates the Sortino ratio (focuses on downside deviation)
func (c *Calculator) CalculateSortinoRatio(trades []backtester.TradeResult, riskFreeRate float64) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	// Calculate average return
	totalReturn := 0.0
	for _, trade := range trades {
		totalReturn += trade.PnL
	}
	avgReturn := totalReturn / float64(len(trades))

	// Calculate downside deviation (only negative returns)
	sumSquaredDownside := 0.0
	downsideCount := 0
	for _, trade := range trades {
		if trade.PnL < 0 {
			sumSquaredDownside += math.Pow(trade.PnL, 2)
			downsideCount++
		}
	}

	if downsideCount == 0 {
		return math.Inf(1) // Infinite Sortino ratio if no downside
	}

	downsideDeviation := math.Sqrt(sumSquaredDownside / float64(downsideCount))
	
	if downsideDeviation == 0 {
		return 0.0
	}

	dailyRiskFreeRate := riskFreeRate / 365
	return (avgReturn - dailyRiskFreeRate) / downsideDeviation
}

// CalculateProfitFactor calculates the profit factor (gross profit / gross loss)
func (c *Calculator) CalculateProfitFactor(trades []backtester.TradeResult) float64 {
	grossProfit := 0.0
	grossLoss := 0.0

	for _, trade := range trades {
		if trade.PnL > 0 {
			grossProfit += trade.PnL
		} else if trade.PnL < 0 {
			grossLoss += math.Abs(trade.PnL)
		}
	}

	if grossLoss == 0 {
		if grossProfit > 0 {
			return math.Inf(1) // Infinite profit factor
		}
		return 0.0
	}

	return grossProfit / grossLoss
}

// CalculateAverageWin calculates the average winning trade
func (c *Calculator) CalculateAverageWin(trades []backtester.TradeResult) float64 {
	totalWin := 0.0
	winCount := 0

	for _, trade := range trades {
		if trade.PnL > 0 {
			totalWin += trade.PnL
			winCount++
		}
	}

	if winCount == 0 {
		return 0.0
	}

	return totalWin / float64(winCount)
}

// CalculateAverageLoss calculates the average losing trade
func (c *Calculator) CalculateAverageLoss(trades []backtester.TradeResult) float64 {
	totalLoss := 0.0
	lossCount := 0

	for _, trade := range trades {
		if trade.PnL < 0 {
			totalLoss += math.Abs(trade.PnL)
			lossCount++
		}
	}

	if lossCount == 0 {
		return 0.0
	}

	return totalLoss / float64(lossCount)
}

// CalculateLargestWin finds the largest winning trade
func (c *Calculator) CalculateLargestWin(trades []backtester.TradeResult) float64 {
	largestWin := 0.0

	for _, trade := range trades {
		if trade.PnL > largestWin {
			largestWin = trade.PnL
		}
	}

	return largestWin
}

// CalculateLargestLoss finds the largest losing trade
func (c *Calculator) CalculateLargestLoss(trades []backtester.TradeResult) float64 {
	largestLoss := 0.0

	for _, trade := range trades {
		if trade.PnL < 0 && math.Abs(trade.PnL) > largestLoss {
			largestLoss = math.Abs(trade.PnL)
		}
	}

	return largestLoss
}

// CalculateMaxConsecutiveWins calculates the maximum consecutive winning trades
func (c *Calculator) CalculateMaxConsecutiveWins(trades []backtester.TradeResult) int {
	maxWins := 0
	currentWins := 0

	for _, trade := range trades {
		if trade.PnL > 0 {
			currentWins++
			if currentWins > maxWins {
				maxWins = currentWins
			}
		} else {
			currentWins = 0
		}
	}

	return maxWins
}

// CalculateMaxConsecutiveLosses calculates the maximum consecutive losing trades
func (c *Calculator) CalculateMaxConsecutiveLosses(trades []backtester.TradeResult) int {
	maxLosses := 0
	currentLosses := 0

	for _, trade := range trades {
		if trade.PnL < 0 {
			currentLosses++
			if currentLosses > maxLosses {
				maxLosses = currentLosses
			}
		} else {
			currentLosses = 0
		}
	}

	return maxLosses
}

// CalculateAverageTradeDuration calculates the average trade duration
func (c *Calculator) CalculateAverageTradeDuration(trades []backtester.TradeResult) time.Duration {
	if len(trades) == 0 {
		return 0
	}

	totalDuration := time.Duration(0)
	for _, trade := range trades {
		totalDuration += trade.Duration
	}

	return totalDuration / time.Duration(len(trades))
}

// CalculateStandardDeviation calculates the standard deviation of trade returns
func (c *Calculator) CalculateStandardDeviation(trades []backtester.TradeResult) float64 {
	if len(trades) <= 1 {
		return 0.0
	}

	// Calculate mean
	total := 0.0
	for _, trade := range trades {
		total += trade.PnL
	}
	mean := total / float64(len(trades))

	// Calculate variance
	sumSquaredDiff := 0.0
	for _, trade := range trades {
		diff := trade.PnL - mean
		sumSquaredDiff += diff * diff
	}
	variance := sumSquaredDiff / float64(len(trades)-1)

	return math.Sqrt(variance)
}

// CalculateVaR calculates Value at Risk at the given confidence level
func (c *Calculator) CalculateVaR(trades []backtester.TradeResult, alpha float64) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	// Sort trades by PnL (ascending)
	sortedPnL := make([]float64, len(trades))
	for i, trade := range trades {
		sortedPnL[i] = trade.PnL
	}

	// Simple bubble sort for small datasets
	for i := 0; i < len(sortedPnL); i++ {
		for j := 0; j < len(sortedPnL)-1-i; j++ {
			if sortedPnL[j] > sortedPnL[j+1] {
				sortedPnL[j], sortedPnL[j+1] = sortedPnL[j+1], sortedPnL[j]
			}
		}
	}

	// Find the alpha percentile
	index := int(float64(len(sortedPnL)) * alpha)
	if index >= len(sortedPnL) {
		index = len(sortedPnL) - 1
	}

	return math.Abs(sortedPnL[index])
}

// CalculateMaxDrawdownPercent calculates maximum drawdown from equity curve
func (c *Calculator) CalculateMaxDrawdownPercent(equityCurve []float64) float64 {
	if len(equityCurve) == 0 {
		return 0.0
	}

	maxDrawdown := 0.0
	peak := equityCurve[0]

	for _, equity := range equityCurve {
		if equity > peak {
			peak = equity
		}

		drawdown := (peak - equity) / peak * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}