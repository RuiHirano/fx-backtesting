package statistics

import (
	"testing"
	"time"
)

// MetricsSet NewMetricsSet テスト
func TestMetricsSet_NewMetricsSet(t *testing.T) {
	metrics := NewMetricsSet()
	
	if metrics == nil {
		t.Fatal("Expected metrics set to be created")
	}
	
	if metrics.Metrics == nil {
		t.Fatal("Expected metrics map to be initialized")
	}
	
	if metrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	if len(metrics.Metrics) != 0 {
		t.Errorf("Expected empty metrics map, got %d items", len(metrics.Metrics))
	}
}

// MetricsSet AddMetric テスト
func TestMetricsSet_AddMetric(t *testing.T) {
	metrics := NewMetricsSet()
	
	// メトリクス追加
	metrics.AddMetric(MetricTotalPnL, 1000.0, "USD", "Total profit and loss")
	
	if len(metrics.Metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics.Metrics))
	}
	
	// 追加されたメトリクスの確認
	metric := metrics.GetMetric(MetricTotalPnL)
	if metric == nil {
		t.Fatal("Expected metric to be found")
	}
	
	if metric.Type != MetricTotalPnL {
		t.Errorf("Expected metric type %v, got %v", MetricTotalPnL, metric.Type)
	}
	
	if metric.Name != "TotalPnL" {
		t.Errorf("Expected metric name 'TotalPnL', got '%s'", metric.Name)
	}
	
	if metric.Value != 1000.0 {
		t.Errorf("Expected metric value 1000.0, got %v", metric.Value)
	}
	
	if metric.Unit != "USD" {
		t.Errorf("Expected metric unit 'USD', got '%s'", metric.Unit)
	}
	
	if metric.Description != "Total profit and loss" {
		t.Errorf("Expected metric description 'Total profit and loss', got '%s'", metric.Description)
	}
	
	// 複数メトリクス追加
	metrics.AddMetric(MetricWinRate, 0.65, "%", "Win rate percentage")
	
	if len(metrics.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics.Metrics))
	}
}

// MetricsSet GetMetric テスト
func TestMetricsSet_GetMetric(t *testing.T) {
	metrics := NewMetricsSet()
	
	// 存在しないメトリクス
	metric := metrics.GetMetric(MetricTotalPnL)
	if metric != nil {
		t.Error("Expected nil for non-existent metric")
	}
	
	// メトリクス追加後
	metrics.AddMetric(MetricTotalPnL, 1500.0, "USD", "Test metric")
	
	metric = metrics.GetMetric(MetricTotalPnL)
	if metric == nil {
		t.Fatal("Expected metric to be found")
	}
	
	if metric.Value != 1500.0 {
		t.Errorf("Expected metric value 1500.0, got %v", metric.Value)
	}
}

// GenerateMetricsFromCalculator テスト
func TestGenerateMetricsFromCalculator(t *testing.T) {
	trades := createTestTrades()
	calculator := NewCalculator(trades)
	
	metricsSet := GenerateMetricsFromCalculator(calculator)
	
	if metricsSet == nil {
		t.Fatal("Expected metrics set to be generated")
	}
	
	// 基本メトリクスの確認
	basicMetrics := []MetricType{
		MetricTotalPnL,
		MetricWinRate,
		MetricTotalTrades,
		MetricAverageWin,
		MetricAverageLoss,
		MetricMaxDrawdown,
		MetricSharpeRatio,
		MetricProfitFactor,
	}
	
	for _, metricType := range basicMetrics {
		metric := metricsSet.GetMetric(metricType)
		if metric == nil {
			t.Errorf("Expected metric %s to be generated", metricType.String())
		} else {
			// 値の型チェック
			if metric.Value == nil {
				t.Errorf("Expected metric %s to have a value", metricType.String())
			}
		}
	}
	
	// 取引数の具体的確認
	totalTradesMetric := metricsSet.GetMetric(MetricTotalTrades)
	if totalTradesMetric != nil {
		if totalTrades, ok := totalTradesMetric.Value.(int); !ok || totalTrades != len(trades) {
			t.Errorf("Expected total trades %d, got %v", len(trades), totalTradesMetric.Value)
		}
	}
	
	// 勝率の範囲確認
	winRateMetric := metricsSet.GetMetric(MetricWinRate)
	if winRateMetric != nil {
		if winRate, ok := winRateMetric.Value.(float64); ok {
			if winRate < 0 || winRate > 100 {
				t.Errorf("Expected win rate between 0-100, got %f", winRate)
			}
		}
	}
}

// MetricsSet GetBasicMetrics テスト
func TestMetricsSet_GetBasicMetrics(t *testing.T) {
	trades := createTestTrades()
	calculator := NewCalculator(trades)
	metricsSet := GenerateMetricsFromCalculator(calculator)
	
	basicMetrics := metricsSet.GetBasicMetrics()
	
	if basicMetrics == nil {
		t.Fatal("Expected basic metrics to be returned")
	}
	
	// 期待される基本メトリクス
	expectedBasicMetrics := []string{
		"TotalPnL",
		"WinRate",
		"TotalTrades",
		"ProfitFactor",
		"MaxDrawdown",
		"SharpeRatio",
	}
	
	for _, metricName := range expectedBasicMetrics {
		if _, exists := basicMetrics[metricName]; !exists {
			t.Errorf("Expected basic metric %s to be included", metricName)
		}
	}
	
	// 基本メトリクス以外が含まれていないことを確認
	expectedBasicCount := len(expectedBasicMetrics)
	// TotalReturnは生成されていない可能性があるので、実際の数をチェック
	if len(basicMetrics) > expectedBasicCount+1 { // +1 for potential TotalReturn
		t.Errorf("Expected at most %d basic metrics, got %d", expectedBasicCount+1, len(basicMetrics))
	}
}

// MetricsSet GetRiskMetrics テスト
func TestMetricsSet_GetRiskMetrics(t *testing.T) {
	trades := createTestTrades()
	calculator := NewCalculator(trades)
	metricsSet := GenerateMetricsFromCalculator(calculator)
	
	riskMetrics := metricsSet.GetRiskMetrics()
	
	if riskMetrics == nil {
		t.Fatal("Expected risk metrics to be returned")
	}
	
	// 期待されるリスクメトリクス
	expectedRiskMetrics := []string{
		"MaxDrawdown",
		"SharpeRatio",
		"SortinoRatio",
		"CalmarRatio",
		"StandardDeviation",
	}
	
	for _, metricName := range expectedRiskMetrics {
		if _, exists := riskMetrics[metricName]; !exists {
			t.Errorf("Expected risk metric %s to be included", metricName)
		}
	}
	
	if len(riskMetrics) != len(expectedRiskMetrics) {
		t.Errorf("Expected %d risk metrics, got %d", len(expectedRiskMetrics), len(riskMetrics))
	}
}

// MetricsSet GetTradingMetrics テスト
func TestMetricsSet_GetTradingMetrics(t *testing.T) {
	trades := createTestTrades()
	calculator := NewCalculator(trades)
	metricsSet := GenerateMetricsFromCalculator(calculator)
	
	tradingMetrics := metricsSet.GetTradingMetrics()
	
	if tradingMetrics == nil {
		t.Fatal("Expected trading metrics to be returned")
	}
	
	// 期待される取引メトリクス
	expectedTradingMetrics := []string{
		"MaxConsecutiveWins",
		"MaxConsecutiveLosses",
		"AverageHoldingPeriod",
		"TradingFrequency",
		"RiskRewardRatio",
		"ExpectedValue",
	}
	
	for _, metricName := range expectedTradingMetrics {
		if _, exists := tradingMetrics[metricName]; !exists {
			t.Errorf("Expected trading metric %s to be included", metricName)
		}
	}
	
	if len(tradingMetrics) != len(expectedTradingMetrics) {
		t.Errorf("Expected %d trading metrics, got %d", len(expectedTradingMetrics), len(tradingMetrics))
	}
}

// MetricType String テスト
func TestMetricType_String(t *testing.T) {
	testCases := []struct {
		metricType MetricType
		expected   string
	}{
		{MetricTotalPnL, "TotalPnL"},
		{MetricWinRate, "WinRate"},
		{MetricSharpeRatio, "SharpeRatio"},
		{MetricMaxDrawdown, "MaxDrawdown"},
		{MetricProfitFactor, "ProfitFactor"},
		{MetricTotalTrades, "TotalTrades"},
		{MetricAverageHoldingPeriod, "AverageHoldingPeriod"},
	}
	
	for _, tc := range testCases {
		result := tc.metricType.String()
		if result != tc.expected {
			t.Errorf("Expected %s for metric type %v, got %s", tc.expected, tc.metricType, result)
		}
	}
	
	// 不明なメトリクスタイプ
	unknownMetric := MetricType(999)
	if unknownMetric.String() != "Unknown" {
		t.Errorf("Expected 'Unknown' for unknown metric type, got %s", unknownMetric.String())
	}
}

// Metric 構造体テスト
func TestMetric_Structure(t *testing.T) {
	metric := &Metric{
		Type:        MetricTotalPnL,
		Name:        "TotalPnL",
		Value:       1500.75,
		Unit:        "USD",
		Description: "Total profit and loss from all trades",
	}
	
	if metric.Type != MetricTotalPnL {
		t.Errorf("Expected metric type %v, got %v", MetricTotalPnL, metric.Type)
	}
	
	if metric.Name != "TotalPnL" {
		t.Errorf("Expected metric name 'TotalPnL', got '%s'", metric.Name)
	}
	
	if metric.Value != 1500.75 {
		t.Errorf("Expected metric value 1500.75, got %v", metric.Value)
	}
	
	if metric.Unit != "USD" {
		t.Errorf("Expected metric unit 'USD', got '%s'", metric.Unit)
	}
	
	if metric.Description == "" {
		t.Error("Expected metric description to be set")
	}
}

// MetricsSet タイムスタンプテスト
func TestMetricsSet_Timestamp(t *testing.T) {
	before := time.Now()
	metrics := NewMetricsSet()
	after := time.Now()
	
	if metrics.Timestamp.Before(before) || metrics.Timestamp.After(after) {
		t.Error("Expected timestamp to be set to current time")
	}
}