// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

func main() {
	fmt.Println("FX Backtesting Library - Basic Example")
	fmt.Println("=====================================")

	// 1. Setup configuration
	config := models.NewConfig(
		10000.0, // Initial balance: $10,000
		0.0001,  // Spread: 1 pip
		0.0001,  // Commission: 0.01% per trade
		0.0,     // Slippage: 0 pips
		100.0,   // Leverage: 1:100
	)

	// 2. Create components
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	
	// Create a moving average crossover strategy
	// Fast MA: 3 periods, Slow MA: 5 periods, Position size: 1000 units
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 3, 5, 1000.0)

	// 3. Create backtester
	bt := backtester.NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// 4. Load historical data (using sample data)
	candles, err := dataProvider.LoadCSVData("../testdata/sample.csv")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	fmt.Printf("Loaded %d candles for backtesting\n", len(candles))

	// 5. Run backtest with progress callback
	fmt.Println("Running backtest...")
	start := time.Now()

	result, err := bt.RunWithCallback(candles, func(progress backtester.Progress) {
		if progress.ProcessedCandles%50 == 0 || progress.Percentage == 100.0 {
			fmt.Printf("Progress: %.1f%% (%d/%d candles processed)\n",
				progress.Percentage, progress.ProcessedCandles, progress.TotalCandles)
		}
	})

	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Backtest completed in %v\n\n", duration)

	// 6. Calculate statistics
	calc := statistics.NewCalculator()
	metrics := calc.CalculateMetrics(result)

	// 7. Generate and display results
	generator := statistics.NewReportGenerator()

	// Text report
	fmt.Println("BACKTEST RESULTS:")
	fmt.Println("=================")
	textReport := generator.GenerateTextReport(result, metrics)
	fmt.Print(textReport)

	// Display key metrics
	fmt.Println("\nKEY INSIGHTS:")
	fmt.Println("=============")
	if metrics.TotalReturn > 0 {
		fmt.Printf("✅ Strategy was PROFITABLE with %.2f%% return\n", metrics.TotalReturn)
	} else {
		fmt.Printf("❌ Strategy was UNPROFITABLE with %.2f%% loss\n", metrics.TotalReturn)
	}

	if metrics.SharpeRatio > 1.0 {
		fmt.Printf("✅ Excellent risk-adjusted returns (Sharpe: %.2f)\n", metrics.SharpeRatio)
	} else if metrics.SharpeRatio > 0 {
		fmt.Printf("⚠️  Moderate risk-adjusted returns (Sharpe: %.2f)\n", metrics.SharpeRatio)
	} else {
		fmt.Printf("❌ Poor risk-adjusted returns (Sharpe: %.2f)\n", metrics.SharpeRatio)
	}

	if result.WinRate > 50 {
		fmt.Printf("✅ Good win rate: %.1f%%\n", result.WinRate)
	} else {
		fmt.Printf("⚠️  Low win rate: %.1f%%\n", result.WinRate)
	}

	fmt.Printf("📊 Total trades executed: %d\n", result.TotalTrades)
	fmt.Printf("💰 Final balance: $%.2f (P&L: $%.2f)\n", result.FinalBalance, result.TotalPnL)

	fmt.Println("\n📄 For detailed analysis, you can also generate JSON or CSV reports.")
	fmt.Println("🚀 Try different strategy parameters to optimize performance!")
}