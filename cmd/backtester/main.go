package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	
	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

const Version = "1.0.0"

// ConfigJSON represents the JSON configuration file structure
type ConfigJSON struct {
	InitialBalance float64 `json:"initial_balance"`
	Spread         float64 `json:"spread"`
	Commission     float64 `json:"commission"`
	Slippage       float64 `json:"slippage"`
	Leverage       float64 `json:"leverage"`
}

// ToConfig converts ConfigJSON to models.Config
func (c ConfigJSON) ToConfig() models.Config {
	return models.NewConfig(
		c.InitialBalance,
		c.Spread,
		c.Commission,
		c.Slippage,
		c.Leverage,
	)
}

// CLIConfig holds all command line configuration
type CLIConfig struct {
	DataFile     string
	ConfigFile   string
	OutputFile   string
	Strategy     string
	FastPeriod   int
	SlowPeriod   int
	PositionSize float64
	OutputFormat string
	ShowHelp     bool
	ShowVersion  bool
}

func main() {
	config := parseArguments(os.Args)
	
	if config.ShowHelp {
		showHelp()
		return
	}
	
	if config.ShowVersion {
		showVersion()
		return
	}
	
	if config.DataFile == "" {
		fmt.Fprintf(os.Stderr, "Error: Data file is required. Use --help for usage information.\n")
		os.Exit(1)
	}
	
	if err := runBacktest(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseArguments(args []string) *CLIConfig {
	config := &CLIConfig{
		Strategy:     "ma",
		FastPeriod:   3,
		SlowPeriod:   5,
		PositionSize: 1000.0,
		OutputFormat: "text",
	}
	
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	
	fs.StringVar(&config.DataFile, "data", "", "Path to CSV data file (required)")
	fs.StringVar(&config.ConfigFile, "config", "", "Path to configuration file")
	fs.StringVar(&config.OutputFile, "output", "", "Path to output file (default: stdout)")
	fs.StringVar(&config.Strategy, "strategy", config.Strategy, "Strategy to use (ma)")
	fs.IntVar(&config.FastPeriod, "fast-period", config.FastPeriod, "Fast moving average period")
	fs.IntVar(&config.SlowPeriod, "slow-period", config.SlowPeriod, "Slow moving average period")
	fs.Float64Var(&config.PositionSize, "position-size", config.PositionSize, "Position size for trades")
	fs.StringVar(&config.OutputFormat, "format", config.OutputFormat, "Output format (text, json, csv)")
	fs.BoolVar(&config.ShowHelp, "help", false, "Show help information")
	fs.BoolVar(&config.ShowVersion, "version", false, "Show version information")
	
	// Parse arguments, ignoring errors for now (we'll handle them in tests)
	fs.Parse(args[1:])
	
	return config
}

func showHelp() {
	fmt.Println("FX Backtesting Tool")
	fmt.Println("Usage: backtester [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --data FILE         Path to CSV data file (required)")
	fmt.Println("  --config FILE       Path to configuration file")
	fmt.Println("  --output FILE       Path to output file (default: stdout)")
	fmt.Println("  --strategy TYPE     Strategy to use (ma)")
	fmt.Println("  --fast-period N     Fast moving average period (default: 3)")
	fmt.Println("  --slow-period N     Slow moving average period (default: 5)")
	fmt.Println("  --position-size N   Position size for trades (default: 1000.0)")
	fmt.Println("  --format FORMAT     Output format: text, json, csv (default: text)")
	fmt.Println("  --help              Show this help message")
	fmt.Println("  --version           Show version information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  backtester --data data.csv")
	fmt.Println("  backtester --data data.csv --strategy ma --fast-period 5 --slow-period 10")
	fmt.Println("  backtester --data data.csv --output results.txt --format json")
}

func showVersion() {
	fmt.Printf("FX Backtesting Tool version %s\n", Version)
}

func runBacktest(config *CLIConfig) error {
	// Load configuration
	backtestConfig, err := loadConfig(config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Setup components
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(*backtestConfig)
	
	// Create strategy based on configuration
	var strategyInstance strategy.Strategy
	switch config.Strategy {
	case "ma":
		strategyInstance = strategy.NewMovingAverageStrategy(
			"EURUSD", // Default symbol for now
			config.FastPeriod,
			config.SlowPeriod,
			config.PositionSize,
		)
	default:
		return fmt.Errorf("unknown strategy: %s", config.Strategy)
	}
	
	// Create backtester
	bt := backtester.NewBacktester(dataProvider, brokerInstance, strategyInstance, *backtestConfig)
	
	// Load data
	candles, err := dataProvider.LoadCSVData(config.DataFile)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}
	
	// Run backtest with progress tracking
	fmt.Printf("Running backtest on %d candles...\n", len(candles))
	start := time.Now()
	
	result, err := bt.RunWithCallback(candles, func(progress backtester.Progress) {
		if progress.ProcessedCandles%100 == 0 || progress.Percentage == 100.0 {
			fmt.Printf("Progress: %.1f%% (%d/%d)\n", 
				progress.Percentage, progress.ProcessedCandles, progress.TotalCandles)
		}
	})
	
	if err != nil {
		return fmt.Errorf("backtest failed: %w", err)
	}
	
	duration := time.Since(start)
	fmt.Printf("Backtest completed in %v\n\n", duration)
	
	// Generate and output results
	return outputResults(result, config)
}

func loadConfig(configFile string) (*models.Config, error) {
	if configFile == "" {
		// Use default configuration
		config := models.DefaultConfig()
		return &config, nil
	}
	
	// Read JSON config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse JSON
	var configJSON ConfigJSON
	if err := json.Unmarshal(data, &configJSON); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}
	
	// Convert to models.Config
	config := configJSON.ToConfig()
	return &config, nil
}

func outputResults(result *backtester.Result, config *CLIConfig) error {
	// Calculate statistics
	calc := statistics.NewCalculator()
	metrics := calc.CalculateMetrics(result)
	
	// Generate report
	generator := statistics.NewReportGenerator()
	var report string
	
	switch config.OutputFormat {
	case "text":
		report = generator.GenerateTextReport(result, metrics)
	case "json":
		report = generator.GenerateJSONReport(result, metrics)
	case "csv":
		report = generator.GenerateCSVReport(result.Trades)
	default:
		return fmt.Errorf("unknown output format: %s", config.OutputFormat)
	}
	
	// Output results
	if config.OutputFile == "" {
		// Output to stdout
		fmt.Print(report)
	} else {
		// Output to file
		if err := os.WriteFile(config.OutputFile, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Results written to %s\n", config.OutputFile)
	}
	
	return nil
}