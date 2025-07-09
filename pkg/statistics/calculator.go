package statistics

import (
	"math"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Calculator は統計計算機能を提供します。
type Calculator struct {
	trades []*models.Trade
}

// NewCalculator は新しいCalculatorを作成します。
func NewCalculator(trades []*models.Trade) *Calculator {
	// nilトレードを除外
	validTrades := make([]*models.Trade, 0, len(trades))
	for _, trade := range trades {
		if trade != nil {
			validTrades = append(validTrades, trade)
		}
	}
	
	return &Calculator{
		trades: validTrades,
	}
}

// GetTrades は取引履歴を取得します。
func (c *Calculator) GetTrades() []*models.Trade {
	return c.trades
}

// CalculateTotalPnL は総利益・損失を計算します。
func (c *Calculator) CalculateTotalPnL() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	var total float64
	for _, trade := range c.trades {
		total += trade.PnL
	}
	return total
}

// CalculateWinRate は勝率を計算します。
func (c *Calculator) CalculateWinRate() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	winCount := 0
	for _, trade := range c.trades {
		if trade.IsWinning() {
			winCount++
		}
	}
	
	return float64(winCount) / float64(len(c.trades))
}

// CalculateAverageProfit は平均利益を計算します。
func (c *Calculator) CalculateAverageProfit() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	var totalProfit float64
	var profitCount int
	
	for _, trade := range c.trades {
		if trade.IsWinning() {
			totalProfit += trade.PnL
			profitCount++
		}
	}
	
	if profitCount == 0 {
		return 0.0
	}
	
	return totalProfit / float64(profitCount)
}

// CalculateAverageLoss は平均損失を計算します。
func (c *Calculator) CalculateAverageLoss() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	var totalLoss float64
	var lossCount int
	
	for _, trade := range c.trades {
		if trade.IsLosing() {
			totalLoss += -trade.PnL // 絶対値として計算
			lossCount++
		}
	}
	
	if lossCount == 0 {
		return 0.0
	}
	
	return totalLoss / float64(lossCount)
}

// CalculateMaxProfit は最大利益を計算します。
func (c *Calculator) CalculateMaxProfit() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	maxProfit := 0.0
	for _, trade := range c.trades {
		if trade.PnL > maxProfit {
			maxProfit = trade.PnL
		}
	}
	
	return maxProfit
}

// CalculateMaxLoss は最大損失を計算します。
func (c *Calculator) CalculateMaxLoss() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	maxLoss := 0.0
	for _, trade := range c.trades {
		if trade.PnL < 0 && -trade.PnL > maxLoss {
			maxLoss = -trade.PnL // 絶対値として返す
		}
	}
	
	return maxLoss
}

// CalculateTotalTrades は取引回数を計算します。
func (c *Calculator) CalculateTotalTrades() int {
	return len(c.trades)
}

// CalculateMaxDrawdown は最大ドローダウンを計算します。
func (c *Calculator) CalculateMaxDrawdown() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	var cumulativePnL float64
	var maxCumulative float64
	var maxDrawdown float64
	
	for _, trade := range c.trades {
		cumulativePnL += trade.PnL
		
		// 新しい高値更新
		if cumulativePnL > maxCumulative {
			maxCumulative = cumulativePnL
		}
		
		// ドローダウン計算
		drawdown := maxCumulative - cumulativePnL
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	
	return maxDrawdown
}

// CalculateSharpeRatio はシャープレシオを計算します。
func (c *Calculator) CalculateSharpeRatio() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	// リターンの計算
	returns := make([]float64, len(c.trades))
	for i, trade := range c.trades {
		returns[i] = trade.PnL
	}
	
	// 平均リターン
	meanReturn := c.CalculateTotalPnL() / float64(len(c.trades))
	
	// 標準偏差
	stdDev := c.calculateStandardDeviation(returns, meanReturn)
	
	if stdDev == 0 {
		return 0.0
	}
	
	// リスクフリーレートは0と仮定
	return meanReturn / stdDev
}

// CalculateSortinoRatio はソルティノレシオを計算します。
func (c *Calculator) CalculateSortinoRatio() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	// 負のリターンのみを使用して下方偏差を計算
	var negativeReturns []float64
	for _, trade := range c.trades {
		if trade.PnL < 0 {
			negativeReturns = append(negativeReturns, trade.PnL)
		}
	}
	
	if len(negativeReturns) == 0 {
		return math.Inf(1) // 負のリターンがない場合は無限大
	}
	
	// 平均リターン
	meanReturn := c.CalculateTotalPnL() / float64(len(c.trades))
	
	// 下方偏差
	downwardDev := c.calculateStandardDeviation(negativeReturns, 0)
	
	if downwardDev == 0 {
		return 0.0
	}
	
	return meanReturn / downwardDev
}

// CalculateReturnRiskRatio はリターン・リスク比を計算します。
func (c *Calculator) CalculateReturnRiskRatio() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	totalReturn := c.CalculateTotalPnL()
	maxDrawdown := c.CalculateMaxDrawdown()
	
	if maxDrawdown == 0 {
		if totalReturn > 0 {
			return math.Inf(1)
		}
		return 0.0
	}
	
	return totalReturn / maxDrawdown
}

// CalculateCalmarRatio はカルマーレシオを計算します。
func (c *Calculator) CalculateCalmarRatio() float64 {
	// Return/Risk Ratioと同じ計算
	return c.CalculateReturnRiskRatio()
}

// CalculateProfitFactor はプロフィットファクターを計算します。
func (c *Calculator) CalculateProfitFactor() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	var grossProfit, grossLoss float64
	
	for _, trade := range c.trades {
		if trade.IsWinning() {
			grossProfit += trade.PnL
		} else if trade.IsLosing() {
			grossLoss += -trade.PnL // 絶対値
		}
	}
	
	if grossLoss == 0 {
		if grossProfit > 0 {
			return math.Inf(1)
		}
		return 0.0
	}
	
	return grossProfit / grossLoss
}

// CalculateExpectedValue は期待値を計算します。
func (c *Calculator) CalculateExpectedValue() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	return c.CalculateTotalPnL() / float64(len(c.trades))
}

// CalculateStandardDeviation は標準偏差を計算します。
func (c *Calculator) CalculateStandardDeviation() float64 {
	if len(c.trades) == 0 {
		return 0.0
	}
	
	returns := make([]float64, len(c.trades))
	for i, trade := range c.trades {
		returns[i] = trade.PnL
	}
	
	mean := c.CalculateExpectedValue()
	return c.calculateStandardDeviation(returns, mean)
}

// CalculateAverageHoldingPeriod は平均保有期間を計算します。
func (c *Calculator) CalculateAverageHoldingPeriod() time.Duration {
	if len(c.trades) == 0 {
		return 0
	}
	
	var totalDuration time.Duration
	for _, trade := range c.trades {
		totalDuration += trade.Duration
	}
	
	return totalDuration / time.Duration(len(c.trades))
}

// CalculateMaxConsecutiveWins は最大連勝数を計算します。
func (c *Calculator) CalculateMaxConsecutiveWins() int {
	if len(c.trades) == 0 {
		return 0
	}
	
	maxWins := 0
	currentWins := 0
	
	for _, trade := range c.trades {
		if trade.IsWinning() {
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

// CalculateMaxConsecutiveLosses は最大連敗数を計算します。
func (c *Calculator) CalculateMaxConsecutiveLosses() int {
	if len(c.trades) == 0 {
		return 0
	}
	
	maxLosses := 0
	currentLosses := 0
	
	for _, trade := range c.trades {
		if trade.IsLosing() {
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

// CalculateTradingFrequency は取引頻度を計算します（1日あたりの取引数）。
func (c *Calculator) CalculateTradingFrequency() float64 {
	if len(c.trades) < 2 {
		return 0.0
	}
	
	// 最初と最後の取引時間から期間を計算
	firstTrade := c.trades[0]
	lastTrade := c.trades[len(c.trades)-1]
	
	duration := lastTrade.OpenTime.Sub(firstTrade.OpenTime)
	if duration <= 0 {
		return 0.0
	}
	
	days := duration.Hours() / 24.0
	if days == 0 {
		return float64(len(c.trades)) // 同日内の場合
	}
	
	return float64(len(c.trades)) / days
}

// CalculateRiskRewardRatio はリスクリワード比を計算します。
func (c *Calculator) CalculateRiskRewardRatio() float64 {
	avgProfit := c.CalculateAverageProfit()
	avgLoss := c.CalculateAverageLoss()
	
	if avgLoss == 0 {
		if avgProfit > 0 {
			return math.Inf(1)
		}
		return 0.0
	}
	
	return avgProfit / avgLoss
}

// calculateStandardDeviation は標準偏差を計算するヘルパー関数です。
func (c *Calculator) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	
	var sum float64
	for _, value := range values {
		diff := value - mean
		sum += diff * diff
	}
	
	variance := sum / float64(len(values)-1)
	return math.Sqrt(variance)
}