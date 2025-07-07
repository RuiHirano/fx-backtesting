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
	fmt.Println("FX Backtesting Debug - Strategy Analysis")
	fmt.Println("========================================")

	// Setup configuration with minimal costs
	config := models.NewConfig(
		10000.0, // Initial balance: $10,000
		0.0,     // No spread for debugging
		0.0,     // No commission for debugging
		0.0,     // No slippage
		100.0,   // Leverage: 1:100
	)

	// Create components
	dataProvider := data.NewCSVDataProvider()
	brokerInstance := broker.NewSimpleBroker(config)
	
	// Create strategy with shorter periods for more signals
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 2, 4, 1000.0)

	// Create backtester
	bt := backtester.NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

	// Load data
	candles, err := dataProvider.LoadCSVData("../testdata/sample.csv")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	fmt.Printf("Loaded %d candles for analysis\n", len(candles))

	// Run backtest with detailed progress
	fmt.Println("Running detailed backtest...")
	start := time.Now()

	result, err := bt.RunWithCallback(candles, func(progress backtester.Progress) {
		fmt.Printf("Candle %d: %.1f%% complete\n", 
			progress.ProcessedCandles, progress.Percentage)
	})

	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	duration := time.Since(start)
	fmt.Printf("Backtest completed in %v\n\n", duration)

	// Detailed analysis
	fmt.Println("DETAILED ANALYSIS:")
	fmt.Println("==================")
	fmt.Printf("🕐 Period: %v to %v\n", result.StartTime, result.EndTime)
	fmt.Printf("💰 Initial Balance: $%.2f\n", result.InitialBalance)
	fmt.Printf("💰 Final Balance: $%.2f\n", result.FinalBalance)
	fmt.Printf("📊 Balance Change: $%.2f\n", result.FinalBalance - result.InitialBalance)
	fmt.Printf("📈 Total P&L: $%.2f\n", result.TotalPnL)
	fmt.Printf("🔢 Total Trades: %d\n", result.TotalTrades)
	fmt.Printf("✅ Winning Trades: %d\n", result.WinningTrades)
	fmt.Printf("❌ Losing Trades: %d\n", result.LosingTrades)
	fmt.Printf("📊 Win Rate: %.1f%%\n", result.WinRate)

	// Check current broker state
	currentBalance := brokerInstance.GetBalance()
	positions := brokerInstance.GetPositions()
	
	fmt.Printf("\nBROKER STATE:\n")
	fmt.Printf("=============\n")
	fmt.Printf("💰 Current Balance: $%.2f\n", currentBalance)
	fmt.Printf("📈 Open Positions: %d\n", len(positions))
	
	for i, pos := range positions {
		pnl := pos.CalculatePnL()
		fmt.Printf("  Position %d: %s %.0f units @ %.5f, PnL: $%.2f\n", 
			i+1, pos.Symbol, pos.Size, pos.EntryPrice, pnl)
	}

	// Generate statistics
	calc := statistics.NewCalculator()
	metrics := calc.CalculateMetrics(result)

	fmt.Printf("\nSTATISTICS:\n")
	fmt.Printf("===========\n")
	fmt.Printf("📊 Total Return: %.2f%%\n", metrics.TotalReturn)
	fmt.Printf("📈 Sharpe Ratio: %.2f\n", metrics.SharpeRatio)
	fmt.Printf("📉 Max Drawdown: %.2f%%\n", metrics.MaxDrawdown)
	fmt.Printf("🎯 Profit Factor: %.2f\n", metrics.ProfitFactor)

	fmt.Printf("\nCONCLUSION:\n")
	fmt.Printf("===========\n")
	if len(positions) > 0 {
		fmt.Printf("✅ Strategy is working - %d positions were opened\n", len(positions))
		fmt.Printf("💡 Consider adding position closing logic for completed trades\n")
	} else {
		fmt.Printf("⚠️  No positions opened - check strategy parameters or data\n")
	}

	if result.TotalTrades == 0 && len(positions) > 0 {
		fmt.Printf("ℹ️  Positions are open but not closed (no completed trades)\n")
		fmt.Printf("💰 Balance reduction is from margin requirements, not losses\n")
	}
}