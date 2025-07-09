package statistics

import (
	"time"
)

// MetricType はメトリクスの種類を表します。
type MetricType int

const (
	// 基本メトリクス
	MetricTotalPnL MetricType = iota
	MetricTotalReturn
	MetricWinRate
	MetricTotalTrades
	MetricWinningTrades
	MetricLosingTrades
	
	// 損益メトリクス
	MetricGrossProfit
	MetricGrossLoss
	MetricAverageWin
	MetricAverageLoss
	MetricLargestWin
	MetricLargestLoss
	
	// リスクメトリクス
	MetricMaxDrawdown
	MetricSharpeRatio
	MetricSortinoRatio
	MetricCalmarRatio
	MetricProfitFactor
	MetricStandardDeviation
	
	// 取引パフォーマンスメトリクス
	MetricMaxConsecutiveWins
	MetricMaxConsecutiveLosses
	MetricAverageHoldingPeriod
	MetricTradingFrequency
	MetricRiskRewardRatio
	MetricExpectedValue
)

// String はMetricTypeの文字列表現を返します。
func (mt MetricType) String() string {
	switch mt {
	case MetricTotalPnL:
		return "TotalPnL"
	case MetricTotalReturn:
		return "TotalReturn"
	case MetricWinRate:
		return "WinRate"
	case MetricTotalTrades:
		return "TotalTrades"
	case MetricWinningTrades:
		return "WinningTrades"
	case MetricLosingTrades:
		return "LosingTrades"
	case MetricGrossProfit:
		return "GrossProfit"
	case MetricGrossLoss:
		return "GrossLoss"
	case MetricAverageWin:
		return "AverageWin"
	case MetricAverageLoss:
		return "AverageLoss"
	case MetricLargestWin:
		return "LargestWin"
	case MetricLargestLoss:
		return "LargestLoss"
	case MetricMaxDrawdown:
		return "MaxDrawdown"
	case MetricSharpeRatio:
		return "SharpeRatio"
	case MetricSortinoRatio:
		return "SortinoRatio"
	case MetricCalmarRatio:
		return "CalmarRatio"
	case MetricProfitFactor:
		return "ProfitFactor"
	case MetricStandardDeviation:
		return "StandardDeviation"
	case MetricMaxConsecutiveWins:
		return "MaxConsecutiveWins"
	case MetricMaxConsecutiveLosses:
		return "MaxConsecutiveLosses"
	case MetricAverageHoldingPeriod:
		return "AverageHoldingPeriod"
	case MetricTradingFrequency:
		return "TradingFrequency"
	case MetricRiskRewardRatio:
		return "RiskRewardRatio"
	case MetricExpectedValue:
		return "ExpectedValue"
	default:
		return "Unknown"
	}
}

// Metric は単一のメトリクス値を表します。
type Metric struct {
	Type        MetricType  `json:"type"`
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
	Description string      `json:"description"`
}

// MetricsSet はメトリクスのセットを表します。
type MetricsSet struct {
	Timestamp time.Time           `json:"timestamp"`
	Metrics   map[string]*Metric  `json:"metrics"`
}

// NewMetricsSet は新しいMetricsSetを作成します。
func NewMetricsSet() *MetricsSet {
	return &MetricsSet{
		Timestamp: time.Now(),
		Metrics:   make(map[string]*Metric),
	}
}

// AddMetric はメトリクスを追加します。
func (ms *MetricsSet) AddMetric(metricType MetricType, value interface{}, unit, description string) {
	metric := &Metric{
		Type:        metricType,
		Name:        metricType.String(),
		Value:       value,
		Unit:        unit,
		Description: description,
	}
	ms.Metrics[metric.Name] = metric
}

// GetMetric は指定されたメトリクスを取得します。
func (ms *MetricsSet) GetMetric(metricType MetricType) *Metric {
	return ms.Metrics[metricType.String()]
}

// GenerateMetricsFromCalculator はCalculatorからメトリクスセットを生成します。
func GenerateMetricsFromCalculator(calculator *Calculator) *MetricsSet {
	metrics := NewMetricsSet()
	
	// 基本メトリクス
	metrics.AddMetric(MetricTotalPnL, calculator.CalculateTotalPnL(), "USD", "Total profit and loss")
	metrics.AddMetric(MetricWinRate, calculator.CalculateWinRate()*100, "%", "Percentage of winning trades")
	metrics.AddMetric(MetricTotalTrades, calculator.CalculateTotalTrades(), "count", "Total number of trades")
	
	// 損益メトリクス
	metrics.AddMetric(MetricAverageWin, calculator.CalculateAverageProfit(), "USD", "Average profit per winning trade")
	metrics.AddMetric(MetricAverageLoss, calculator.CalculateAverageLoss(), "USD", "Average loss per losing trade")
	metrics.AddMetric(MetricLargestWin, calculator.CalculateMaxProfit(), "USD", "Largest single profit")
	metrics.AddMetric(MetricLargestLoss, calculator.CalculateMaxLoss(), "USD", "Largest single loss")
	
	// リスクメトリクス
	metrics.AddMetric(MetricMaxDrawdown, calculator.CalculateMaxDrawdown(), "USD", "Maximum drawdown from peak")
	metrics.AddMetric(MetricSharpeRatio, calculator.CalculateSharpeRatio(), "ratio", "Risk-adjusted return measure")
	metrics.AddMetric(MetricSortinoRatio, calculator.CalculateSortinoRatio(), "ratio", "Downside risk-adjusted return")
	metrics.AddMetric(MetricCalmarRatio, calculator.CalculateCalmarRatio(), "ratio", "Return to max drawdown ratio")
	metrics.AddMetric(MetricProfitFactor, calculator.CalculateProfitFactor(), "ratio", "Gross profit to gross loss ratio")
	metrics.AddMetric(MetricStandardDeviation, calculator.CalculateStandardDeviation(), "USD", "Standard deviation of returns")
	
	// 取引パフォーマンスメトリクス
	metrics.AddMetric(MetricMaxConsecutiveWins, calculator.CalculateMaxConsecutiveWins(), "count", "Maximum consecutive winning trades")
	metrics.AddMetric(MetricMaxConsecutiveLosses, calculator.CalculateMaxConsecutiveLosses(), "count", "Maximum consecutive losing trades")
	metrics.AddMetric(MetricAverageHoldingPeriod, calculator.CalculateAverageHoldingPeriod().Hours(), "hours", "Average trade holding period")
	metrics.AddMetric(MetricTradingFrequency, calculator.CalculateTradingFrequency(), "trades/day", "Trading frequency per day")
	metrics.AddMetric(MetricRiskRewardRatio, calculator.CalculateRiskRewardRatio(), "ratio", "Average win to average loss ratio")
	metrics.AddMetric(MetricExpectedValue, calculator.CalculateExpectedValue(), "USD", "Expected value per trade")
	
	return metrics
}

// GetBasicMetrics は基本的なメトリクスのみを取得します。
func (ms *MetricsSet) GetBasicMetrics() map[string]*Metric {
	basic := make(map[string]*Metric)
	
	basicTypes := []MetricType{
		MetricTotalPnL,
		MetricTotalReturn,
		MetricWinRate,
		MetricTotalTrades,
		MetricProfitFactor,
		MetricMaxDrawdown,
		MetricSharpeRatio,
	}
	
	for _, metricType := range basicTypes {
		if metric := ms.GetMetric(metricType); metric != nil {
			basic[metricType.String()] = metric
		}
	}
	
	return basic
}

// GetRiskMetrics はリスク関連メトリクスのみを取得します。
func (ms *MetricsSet) GetRiskMetrics() map[string]*Metric {
	risk := make(map[string]*Metric)
	
	riskTypes := []MetricType{
		MetricMaxDrawdown,
		MetricSharpeRatio,
		MetricSortinoRatio,
		MetricCalmarRatio,
		MetricStandardDeviation,
	}
	
	for _, metricType := range riskTypes {
		if metric := ms.GetMetric(metricType); metric != nil {
			risk[metricType.String()] = metric
		}
	}
	
	return risk
}

// GetTradingMetrics は取引パフォーマンス関連メトリクスのみを取得します。
func (ms *MetricsSet) GetTradingMetrics() map[string]*Metric {
	trading := make(map[string]*Metric)
	
	tradingTypes := []MetricType{
		MetricMaxConsecutiveWins,
		MetricMaxConsecutiveLosses,
		MetricAverageHoldingPeriod,
		MetricTradingFrequency,
		MetricRiskRewardRatio,
		MetricExpectedValue,
	}
	
	for _, metricType := range tradingTypes {
		if metric := ms.GetMetric(metricType); metric != nil {
			trading[metricType.String()] = metric
		}
	}
	
	return trading
}