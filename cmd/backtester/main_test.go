package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// CLI ArgumentParsing テスト
func TestCLI_ArgumentParsing(t *testing.T) {
	// 正常な引数解析テスト
	args := []string{
		"backtester",
		"-data", "./testdata/sample.csv",
		"-config", "./testdata/config.json",
		"-output", "./results.txt",
		"-format", "json",
	}
	
	config, err := parseArgs(args)
	if err != nil {
		t.Fatalf("Expected no error from parseArgs, got %v", err)
	}
	
	if config.DataFile != "./testdata/sample.csv" {
		t.Errorf("Expected data file './testdata/sample.csv', got '%s'", config.DataFile)
	}
	
	if config.ConfigFile != "./testdata/config.json" {
		t.Errorf("Expected config file './testdata/config.json', got '%s'", config.ConfigFile)
	}
	
	if config.OutputFile != "./results.txt" {
		t.Errorf("Expected output file './results.txt', got '%s'", config.OutputFile)
	}
	
	if config.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", config.Format)
	}
	
	// 必須引数なしのテスト
	argsNoData := []string{"backtester", "-config", "./config.json"}
	_, err = parseArgs(argsNoData)
	if err == nil {
		t.Error("Expected error for missing data argument")
	}
	
	argsNoConfig := []string{"backtester", "-data", "./data.csv"}
	_, err = parseArgs(argsNoConfig)
	if err == nil {
		t.Error("Expected error for missing config argument")
	}
	
	// デフォルト値テスト
	argsMinimal := []string{
		"backtester",
		"-data", "./testdata/sample.csv",
		"-config", "./testdata/config.json",
	}
	
	config, err = parseArgs(argsMinimal)
	if err != nil {
		t.Fatalf("Expected no error from minimal args, got %v", err)
	}
	
	if config.Format != "text" {
		t.Errorf("Expected default format 'text', got '%s'", config.Format)
	}
	
	if config.OutputFile != "" {
		t.Errorf("Expected default output file '', got '%s'", config.OutputFile)
	}
}

// CLI ExecutionFlow テスト
func TestCLI_ExecutionFlow(t *testing.T) {
	// テストデータの準備
	setupTestData(t)
	defer cleanupTestData(t)
	
	// 正常な実行フローテスト
	config := &CLIConfig{
		DataFile:   "./testdata/sample.csv",
		ConfigFile: "./testdata/config.json",
		Format:     "text",
		OutputFile: "./testdata/output.txt",
	}
	
	err := runBacktest(config)
	if err != nil {
		t.Fatalf("Expected no error from runBacktest, got %v", err)
	}
	
	// 出力ファイルの存在確認
	if _, err := os.Stat(config.OutputFile); os.IsNotExist(err) {
		t.Errorf("Expected output file to be created at %s", config.OutputFile)
	}
	
	// 出力ファイルの内容確認
	content, err := os.ReadFile(config.OutputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Expected output file to contain data")
	}
	
	// 出力内容の基本チェック
	contentStr := string(content)
	if !strings.Contains(contentStr, "バックテスト結果") {
		t.Error("Expected output to contain backtest results")
	}
}

// CLI ErrorHandling テスト
func TestCLI_ErrorHandling(t *testing.T) {
	// 存在しないデータファイル
	config := &CLIConfig{
		DataFile:   "./nonexistent.csv",
		ConfigFile: "./testdata/config.json",
		Format:     "text",
		OutputFile: "",
	}
	
	err := runBacktest(config)
	if err == nil {
		t.Error("Expected error for nonexistent data file")
	}
	
	// 存在しない設定ファイル
	config = &CLIConfig{
		DataFile:   "./testdata/sample.csv",
		ConfigFile: "./nonexistent.json",
		Format:     "text",
		OutputFile: "",
	}
	
	err = runBacktest(config)
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
	
	// 無効なフォーマット
	config = &CLIConfig{
		DataFile:   "./testdata/sample.csv",
		ConfigFile: "./testdata/config.json",
		Format:     "invalid",
		OutputFile: "",
	}
	
	err = runBacktest(config)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
	
	// 無効な出力ディレクトリ
	config = &CLIConfig{
		DataFile:   "./testdata/sample.csv",
		ConfigFile: "./testdata/config.json",
		Format:     "text",
		OutputFile: "/invalid/path/output.txt",
	}
	
	err = runBacktest(config)
	if err == nil {
		t.Error("Expected error for invalid output directory")
	}
}

// CLI ConfigValidation テスト
func TestCLI_ConfigValidation(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)
	
	// 正常な設定ファイル
	config, err := loadConfig("./testdata/config.json")
	if err != nil {
		t.Fatalf("Expected no error from loadConfig, got %v", err)
	}
	
	if config == nil {
		t.Fatal("Expected config to be loaded")
	}
	
	// 設定内容の基本確認
	if config.Broker.InitialBalance <= 0 {
		t.Error("Expected positive initial balance")
	}
	
	if config.Broker.Spread < 0 {
		t.Error("Expected non-negative spread")
	}
	
	// 無効な設定ファイル
	invalidConfigPath := "./testdata/invalid_config.json"
	createInvalidConfig(t, invalidConfigPath)
	defer os.Remove(invalidConfigPath)
	
	_, err = loadConfig(invalidConfigPath)
	if err == nil {
		t.Error("Expected error for invalid config file")
	}
}

// CLI OutputGeneration テスト
func TestCLI_OutputGeneration(t *testing.T) {
	setupTestData(t)
	defer cleanupTestData(t)
	
	// テキスト形式出力
	config := &CLIConfig{
		DataFile:   "./testdata/sample.csv",
		ConfigFile: "./testdata/config.json",
		Format:     "text",
		OutputFile: "./testdata/output_text.txt",
	}
	
	err := runBacktest(config)
	if err != nil {
		t.Fatalf("Expected no error from text output, got %v", err)
	}
	
	// JSON形式出力
	config.Format = "json"
	config.OutputFile = "./testdata/output_json.json"
	
	err = runBacktest(config)
	if err != nil {
		t.Fatalf("Expected no error from json output, got %v", err)
	}
	
	// CSV形式出力
	config.Format = "csv"
	config.OutputFile = "./testdata/output_csv.csv"
	
	err = runBacktest(config)
	if err != nil {
		t.Fatalf("Expected no error from csv output, got %v", err)
	}
	
	// 出力ファイルの存在確認
	outputFiles := map[string]string{
		"text": "./testdata/output_text.txt",
		"json": "./testdata/output_json.json",
		"csv":  "./testdata/output_csv.csv",
	}
	for format, filename := range outputFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Expected %s output file to be created at %s", format, filename)
		}
	}
	
	// 標準出力テスト
	config.OutputFile = ""
	config.Format = "text"
	
	var stdout bytes.Buffer
	err = runBacktestWithOutput(config, &stdout)
	if err != nil {
		t.Fatalf("Expected no error from stdout output, got %v", err)
	}
	
	if stdout.Len() == 0 {
		t.Error("Expected output to stdout")
	}
}

// CLI HelpMessages テスト
func TestCLI_HelpMessages(t *testing.T) {
	// ヘルプメッセージの取得
	helpText := getHelpMessage()
	
	if helpText == "" {
		t.Fatal("Expected help message to be generated")
	}
	
	// ヘルプメッセージの内容確認
	requiredElements := []string{
		"Usage:",
		"-data",
		"-config",
		"-output",
		"-format",
		"-help",
		"Examples:",
	}
	
	for _, element := range requiredElements {
		if !strings.Contains(helpText, element) {
			t.Errorf("Help message missing required element: %s", element)
		}
	}
	
	// ヘルプフラグでの終了テスト
	args := []string{"backtester", "-help"}
	config, err := parseArgs(args)
	if err != nil {
		t.Fatalf("Expected no error from help args, got %v", err)
	}
	
	if !config.ShowHelp {
		t.Error("Expected ShowHelp to be true")
	}
}

// ヘルパー関数: テストデータセットアップ
func setupTestData(t *testing.T) {
	// テストディレクトリ作成
	err := os.MkdirAll("./testdata", 0755)
	if err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}
	
	// サンプルCSVファイル作成
	csvContent := `timestamp,open,high,low,close,volume
2025-01-01 00:00:00,1.1000,1.1010,1.0990,1.1005,1000
2025-01-01 00:01:00,1.1005,1.1015,1.0995,1.1010,1200
2025-01-01 00:02:00,1.1010,1.1020,1.1000,1.1015,1100
2025-01-01 00:03:00,1.1015,1.1025,1.1005,1.1020,1300
2025-01-01 00:04:00,1.1020,1.1030,1.1010,1.1025,1400`
	
	err = os.WriteFile("./testdata/sample.csv", []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create sample CSV: %v", err)
	}
	
	// 設定ファイル作成
	configContent := `{
  "market": {
    "data_provider": {
      "file_path": "./testdata/sample.csv",
      "format": "csv"
    },
    "symbol": "EURUSD"
  },
  "broker": {
    "initial_balance": 10000.0,
    "spread": 0.0001
  }
}`
	
	err = os.WriteFile("./testdata/config.json", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}

// ヘルパー関数: テストデータクリーンアップ
func cleanupTestData(t *testing.T) {
	os.RemoveAll("./testdata")
}

// ヘルパー関数: 無効な設定ファイル作成
func createInvalidConfig(t *testing.T, filename string) {
	invalidContent := `{
  "broker": {
    "initial_balance": -1000.0,
    "spread": -0.1
  }
}`
	err := os.WriteFile(filename, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}
}