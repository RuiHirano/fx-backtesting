package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
)

// CLIConfig はCLI設定を表します。
type CLIConfig struct {
	DataFile   string
	ConfigFile string
	OutputFile string
	Format     string
	ShowHelp   bool
}

// BacktestConfig はバックテスト設定を表します。
type BacktestConfig struct {
	Market models.MarketConfig `json:"market"`
	Broker models.BrokerConfig `json:"broker"`
}

// main はCLIアプリケーションのエントリーポイントです。
func main() {
	config, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	
	if config.ShowHelp {
		fmt.Print(getHelpMessage())
		os.Exit(0)
	}
	
	err = runBacktest(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs はコマンドライン引数を解析します。
func parseArgs(args []string) (*CLIConfig, error) {
	// フラグセット作成
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.Usage = func() {} // カスタムヘルプを使用
	
	config := &CLIConfig{}
	
	// フラグ定義
	fs.StringVar(&config.DataFile, "data", "", "Data file path (required)")
	fs.StringVar(&config.ConfigFile, "config", "", "Config file path (required)")
	fs.StringVar(&config.OutputFile, "output", "", "Output file path (optional, default: stdout)")
	fs.StringVar(&config.Format, "format", "text", "Output format (text/json/csv)")
	fs.BoolVar(&config.ShowHelp, "help", false, "Show help message")
	
	// 引数解析
	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	
	// ヘルプ表示の場合は早期リターン
	if config.ShowHelp {
		return config, nil
	}
	
	// 必須引数チェック
	if config.DataFile == "" {
		return nil, errors.New("data file is required (-data)")
	}
	
	if config.ConfigFile == "" {
		return nil, errors.New("config file is required (-config)")
	}
	
	// フォーマット検証
	if config.Format != "text" && config.Format != "json" && config.Format != "csv" {
		return nil, fmt.Errorf("invalid format: %s (must be text/json/csv)", config.Format)
	}
	
	return config, nil
}

// runBacktest はバックテストを実行します。
func runBacktest(config *CLIConfig) error {
	return runBacktestWithOutput(config, os.Stdout)
}

// runBacktestWithOutput は出力先を指定してバックテストを実行します。
func runBacktestWithOutput(config *CLIConfig, stdout io.Writer) error {
	// 設定ファイル読み込み
	backtestConfig, err := loadConfig(config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	// データファイルパス更新
	backtestConfig.Market.DataProvider.FilePath = config.DataFile
	
	// Backtester作成
	bt := backtester.NewBacktester(
		backtestConfig.Market.DataProvider,
		backtestConfig.Broker,
	)
	
	// 初期化
	ctx := context.Background()
	err = bt.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize backtester: %w", err)
	}
	
	// 簡単なバックテスト実行（例として）
	var trades []*models.Trade
	initialBalance := backtestConfig.Broker.InitialBalance
	
	// 基本的な取引ロジック（例）
	for !bt.IsFinished() {
		price := bt.GetCurrentPrice(backtestConfig.Market.Symbol)
		if price > 0 {
			// 簡単な買い戦略
			positions := bt.GetPositions()
			if len(positions) == 0 {
				err = bt.Buy(backtestConfig.Market.Symbol, 1000.0)
				if err == nil {
					// 取引記録作成（簡易版）
					trade := &models.Trade{
						ID:         fmt.Sprintf("trade-%d", len(trades)+1),
						Symbol:     backtestConfig.Market.Symbol,
						Side:       models.Buy,
						Size:       1000.0,
						EntryPrice: price,
						ExitPrice:  price,
						PnL:        0.0,
						Status:     models.TradeClosed,
						OpenTime:   bt.GetCurrentTime(),
						CloseTime:  bt.GetCurrentTime(),
						Duration:   0,
					}
					trades = append(trades, trade)
				}
			}
		}
		
		bt.Forward()
	}
	
	// 残りのポジションをクローズ
	err = bt.CloseAllPositions()
	if err != nil {
		return fmt.Errorf("failed to close positions: %w", err)
	}
	
	// 統計レポート生成
	report := statistics.NewReport(trades, initialBalance)
	
	// 出力生成
	var output string
	switch config.Format {
	case "json":
		output = report.GenerateJSONReport()
	case "csv":
		output = report.GenerateCSVReport()
	default:
		output = report.GenerateTextReport()
	}
	
	// 出力
	if config.OutputFile != "" {
		// ファイル出力
		err = writeToFile(config.OutputFile, output)
		if err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		// 標準出力
		_, err = fmt.Fprint(stdout, output)
		if err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}
	
	return nil
}

// loadConfig は設定ファイルを読み込みます。
func loadConfig(filename string) (*BacktestConfig, error) {
	// ファイル存在確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", filename)
	}
	
	// ファイル読み込み
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// JSON解析
	var config BacktestConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// 設定検証
	err = validateConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return &config, nil
}

// validateConfig は設定の妥当性を検証します。
func validateConfig(config *BacktestConfig) error {
	// Broker設定検証
	if config.Broker.InitialBalance <= 0 {
		return errors.New("initial balance must be positive")
	}
	
	if config.Broker.Spread < 0 {
		return errors.New("spread must be non-negative")
	}
	
	return nil
}

// writeToFile はファイルに出力します。
func writeToFile(filename, content string) error {
	// ディレクトリ作成
	dir := filepath.Dir(filename)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	
	// ファイル書き込み
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// getHelpMessage はヘルプメッセージを取得します。
func getHelpMessage() string {
	return `FX Backtesting Tool

Usage:
  backtester -data <data_file> -config <config_file> [options]

Required Arguments:
  -data     Data file path (CSV format)
  -config   Configuration file path (JSON format)

Options:
  -output   Output file path (default: stdout)
  -format   Output format: text, json, csv (default: text)
  -help     Show this help message

Examples:
  # Basic usage
  backtester -data ./data/EURUSD.csv -config ./config.json

  # With JSON output
  backtester -data ./data/EURUSD.csv -config ./config.json -format json

  # With output file
  backtester -data ./data/EURUSD.csv -config ./config.json -output ./results.txt

  # Show help
  backtester -help

Configuration File Format (JSON):
  {
    "market": {
      "data_provider": {
        "file_path": "./data/EURUSD.csv",
        "format": "csv"
      },
      "symbol": "EURUSD"
    },
    "broker": {
      "initial_balance": 10000.0,
      "spread": 0.0001
    }
  }
`
}