package statistics

import (
	"strings"
	"testing"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Report NewReport テスト
func TestReport_NewReport(t *testing.T) {
	trades := createTestTrades()
	initialBalance := 10000.0
	
	report := NewReport(trades, initialBalance)
	
	if report == nil {
		t.Fatal("Expected report to be created")
	}
	
	if report.calculator == nil {
		t.Fatal("Expected calculator to be initialized")
	}
	
	if report.result == nil {
		t.Fatal("Expected result to be initialized")
	}
	
	// 基本的な結果確認
	if report.result.InitialBalance != initialBalance {
		t.Errorf("Expected initial balance %f, got %f", initialBalance, report.result.InitialBalance)
	}
	
	if report.result.TotalTrades != len(trades) {
		t.Errorf("Expected %d trades, got %d", len(trades), report.result.TotalTrades)
	}
}

// Report GenerateTextReport テスト
func TestReport_GenerateTextReport(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	textReport := report.GenerateTextReport()
	
	if textReport == "" {
		t.Fatal("Expected text report to be generated")
	}
	
	// レポートに必要な要素が含まれているかチェック
	requiredElements := []string{
		"バックテスト結果レポート",
		"基本情報",
		"損益情報",
		"取引統計",
		"リスク指標",
		"取引パフォーマンス",
		"初期残高",
		"最終残高",
		"総損益",
		"勝率",
		"シャープレシオ",
		"最大ドローダウン",
	}
	
	for _, element := range requiredElements {
		if !strings.Contains(textReport, element) {
			t.Errorf("Text report missing required element: %s", element)
		}
	}
}

// Report GenerateJSONReport テスト
func TestReport_GenerateJSONReport(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	jsonReport := report.GenerateJSONReport()
	
	if jsonReport == "" {
		t.Fatal("Expected JSON report to be generated")
	}
	
	// JSON形式の基本的な構造確認
	requiredFields := []string{
		"summary",
		"detailed_metrics",
		"initial_balance",
		"final_balance",
		"total_pnl",
		"total_return",
		"win_rate",
		"profit_factor",
		"sharpe_ratio",
		"max_drawdown",
	}
	
	for _, field := range requiredFields {
		if !strings.Contains(jsonReport, field) {
			t.Errorf("JSON report missing required field: %s", field)
		}
	}
	
	// JSON形式かどうかの基本チェック
	if !strings.HasPrefix(jsonReport, "{") || !strings.HasSuffix(strings.TrimSpace(jsonReport), "}") {
		t.Error("Expected JSON report to be in valid JSON format")
	}
}

// Report GenerateCSVReport テスト
func TestReport_GenerateCSVReport(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	csvReport := report.GenerateCSVReport()
	
	if csvReport == "" {
		t.Fatal("Expected CSV report to be generated")
	}
	
	lines := strings.Split(strings.TrimSpace(csvReport), "\n")
	
	// ヘッダー行 + データ行の確認
	expectedLines := len(trades) + 1 // ヘッダー + データ
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines in CSV, got %d", expectedLines, len(lines))
	}
	
	// ヘッダー確認
	expectedHeader := "ID,Symbol,Side,Size,EntryPrice,ExitPrice,PnL,Status,OpenTime,CloseTime,DurationHours"
	if lines[0] != expectedHeader {
		t.Errorf("Expected CSV header: %s, got: %s", expectedHeader, lines[0])
	}
	
	// データ行の基本確認
	for i := 1; i < len(lines); i++ {
		fields := strings.Split(lines[i], ",")
		if len(fields) < 11 { // 最低限のフィールド数
			t.Errorf("Expected at least 11 fields in CSV line %d, got %d", i, len(fields))
		}
	}
}

// Report GenerateReport（フォーマット指定）テスト
func TestReport_GenerateReport(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	// テキスト形式
	textReport := report.GenerateReport(FormatText)
	if !strings.Contains(textReport, "バックテスト結果レポート") {
		t.Error("Expected text format report")
	}
	
	// JSON形式
	jsonReport := report.GenerateReport(FormatJSON)
	if !strings.Contains(jsonReport, "summary") {
		t.Error("Expected JSON format report")
	}
	
	// CSV形式
	csvReport := report.GenerateReport(FormatCSV)
	if !strings.Contains(csvReport, "ID,Symbol,Side") {
		t.Error("Expected CSV format report")
	}
}

// Report GetSummaryMetrics テスト
func TestReport_GetSummaryMetrics(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	metrics := report.GetSummaryMetrics()
	
	if metrics == nil {
		t.Fatal("Expected summary metrics to be returned")
	}
	
	// 必要なメトリクスが含まれているか確認
	requiredMetrics := []string{
		"total_return",
		"total_trades",
		"win_rate",
		"profit_factor",
		"max_drawdown",
		"sharpe_ratio",
		"sortino_ratio",
		"calmar_ratio",
		"risk_reward_ratio",
		"max_consecutive_wins",
		"max_consecutive_losses",
		"trading_frequency",
		"average_holding_period",
	}
	
	for _, metric := range requiredMetrics {
		if _, exists := metrics[metric]; !exists {
			t.Errorf("Summary metrics missing required metric: %s", metric)
		}
	}
	
	// データ型の基本確認
	if totalTrades, ok := metrics["total_trades"].(int); !ok || totalTrades != len(trades) {
		t.Errorf("Expected total_trades to be %d, got %v", len(trades), metrics["total_trades"])
	}
}

// Report GenerateCompactSummary テスト
func TestReport_GenerateCompactSummary(t *testing.T) {
	trades := createTestTrades()
	report := NewReport(trades, 10000.0)
	
	summary := report.GenerateCompactSummary()
	
	if summary == "" {
		t.Fatal("Expected compact summary to be generated")
	}
	
	// 簡潔なサマリーに含まれるべき要素
	requiredElements := []string{
		"リターン",
		"取引数",
		"勝率",
		"PF", // Profit Factor
		"DD", // Drawdown
		"SR", // Sharpe Ratio
	}
	
	for _, element := range requiredElements {
		if !strings.Contains(summary, element) {
			t.Errorf("Compact summary missing required element: %s", element)
		}
	}
	
	// パーセンテージ記号の確認
	if !strings.Contains(summary, "%") {
		t.Error("Expected compact summary to contain percentage symbols")
	}
}

// Report エラーハンドリングテスト
func TestReport_ErrorHandling(t *testing.T) {
	// 空の取引履歴でのレポート生成
	emptyTrades := []*models.Trade{}
	report := NewReport(emptyTrades, 10000.0)
	
	// テキストレポート
	textReport := report.GenerateTextReport()
	if textReport == "" {
		t.Error("Expected text report to be generated even with empty trades")
	}
	
	// JSONレポート
	jsonReport := report.GenerateJSONReport()
	if jsonReport == "" {
		t.Error("Expected JSON report to be generated even with empty trades")
	}
	
	// CSVレポート（ヘッダーのみになる）
	csvReport := report.GenerateCSVReport()
	lines := strings.Split(strings.TrimSpace(csvReport), "\n")
	if len(lines) != 1 { // ヘッダーのみ
		t.Errorf("Expected 1 line (header only) in empty CSV, got %d", len(lines))
	}
	
	// サマリーメトリクス
	metrics := report.GetSummaryMetrics()
	if metrics == nil {
		t.Error("Expected summary metrics to be returned even with empty trades")
	}
	
	// 簡潔サマリー
	summary := report.GenerateCompactSummary()
	if summary == "" {
		t.Error("Expected compact summary to be generated even with empty trades")
	}
}