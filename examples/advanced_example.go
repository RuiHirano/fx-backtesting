// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

func main() {
	fmt.Println("FX Backtesting Library - Advanced Example")
	fmt.Println("==========================================")
	fmt.Println("This example demonstrates strategy optimization across multiple parameter sets.")

	// Strategy parameter combinations to test
	strategies := []struct {
		name       string
		fastPeriod int
		slowPeriod int
		positionSize float64
	}{
		{"Conservative", 5, 20, 500.0},
		{"Moderate", 10, 30, 1000.0},
		{"Aggressive", 3, 10, 2000.0},
		{"Scalping", 2, 5, 500.0},
	}

	// Load data once
	dataProvider := data.NewCSVDataProvider()
	candles, err := dataProvider.LoadCSVData("../testdata/sample.csv")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	fmt.Printf("Testing %d strategy configurations on %d candles\n\n", len(strategies), len(candles))

	// Track results
	results := make([]*backtester.Result, len(strategies))
	metrics := make([]*statistics.Metrics, len(strategies))
	calc := statistics.NewCalculator()

	// Test each strategy
	for i, strat := range strategies {
		fmt.Printf("Testing %s strategy (Fast MA: %d, Slow MA: %d, Size: %.0f)\n",
			strat.name, strat.fastPeriod, strat.slowPeriod, strat.positionSize)

		// Create configuration for this test
		config := models.NewConfig(10000.0, 0.0001, 0.5, 0.0, 100.0)
		brokerInstance := broker.NewSimpleBroker(config)
		strategyInstance := strategy.NewMovingAverageStrategy(
			"EURUSD", strat.fastPeriod, strat.slowPeriod, strat.positionSize)

		bt := backtester.NewBacktester(dataProvider, brokerInstance, strategyInstance, config)

		// Run backtest
		start := time.Now()
		result, err := bt.Run(candles)
		if err != nil {
			log.Printf("Strategy %s failed: %v", strat.name, err)
			continue
		}
		duration := time.Since(start)

		// Calculate metrics
		strategyMetrics := calc.CalculateMetrics(result)

		// Store results
		results[i] = result
		metrics[i] = strategyMetrics

		// Display quick summary
		fmt.Printf("  ✓ Completed in %v | Return: %.2f%% | Sharpe: %.2f | Trades: %d\n",
			duration, strategyMetrics.TotalReturn, strategyMetrics.SharpeRatio, result.TotalTrades)
	}

	// Find best strategy
	fmt.Println("\nSTRATEGY COMPARISON:")
	fmt.Println("====================")
	
	bestStrategy := 0
	bestSharpe := -999.0
	
	for i, strat := range strategies {
		if metrics[i] == nil {
			continue
		}
		
		status := "📈"
		if metrics[i].TotalReturn < 0 {
			status = "📉"
		}
		
		fmt.Printf("%s %s: Return %.2f%% | Sharpe %.2f | Win Rate %.1f%% | Max DD %.2f%%\n",
			status, strat.name, metrics[i].TotalReturn, metrics[i].SharpeRatio,
			results[i].WinRate, metrics[i].MaxDrawdown)
		
		if metrics[i].SharpeRatio > bestSharpe {
			bestSharpe = metrics[i].SharpeRatio
			bestStrategy = i
		}
	}

	// Generate detailed report for best strategy
	if metrics[bestStrategy] != nil {
		fmt.Printf("\n🏆 BEST STRATEGY: %s\n", strategies[bestStrategy].name)
		fmt.Println("================================")

		generator := statistics.NewReportGenerator()
		detailedReport := generator.GenerateDetailedReport(results[bestStrategy], metrics[bestStrategy])
		
		// Save detailed report to file
		reportFile := "best_strategy_report.txt"
		if err := os.WriteFile(reportFile, []byte(detailedReport), 0644); err != nil {
			log.Printf("Failed to save report: %v", err)
		} else {
			fmt.Printf("📄 Detailed report saved to %s\n", reportFile)
		}

		// Generate JSON report for programmatic use
		jsonReport := generator.GenerateJSONReport(results[bestStrategy], metrics[bestStrategy])
		jsonFile := "best_strategy_data.json"
		if err := os.WriteFile(jsonFile, []byte(jsonReport), 0644); err != nil {
			log.Printf("Failed to save JSON data: %v", err)
		} else {
			fmt.Printf("📊 JSON data saved to %s\n", jsonFile)
		}

		// Generate CSV of all trades
		csvReport := generator.GenerateCSVReport(results[bestStrategy].Trades)
		csvFile := "best_strategy_trades.csv"
		if err := os.WriteFile(csvFile, []byte(csvReport), 0644); err != nil {
			log.Printf("Failed to save trade data: %v", err)
		} else {
			fmt.Printf("📈 Trade data saved to %s\n", csvFile)
		}

		// Risk assessment
		fmt.Println("\nRISK ASSESSMENT:")
		fmt.Println("================")
		if metrics[bestStrategy].MaxDrawdown > 20 {
			fmt.Println("⚠️  HIGH RISK: Maximum drawdown exceeds 20%")
		} else if metrics[bestStrategy].MaxDrawdown > 10 {
			fmt.Println("⚠️  MODERATE RISK: Maximum drawdown between 10-20%")
		} else {
			fmt.Println("✅ LOW RISK: Maximum drawdown under 10%")
		}

		if metrics[bestStrategy].SharpeRatio > 1.0 {
			fmt.Println("✅ EXCELLENT: Sharpe ratio indicates strong risk-adjusted returns")
		} else if metrics[bestStrategy].SharpeRatio > 0.5 {
			fmt.Println("✅ GOOD: Acceptable risk-adjusted returns")
		} else {
			fmt.Println("⚠️  POOR: Low risk-adjusted returns - consider strategy refinement")
		}
	}

	fmt.Println("\n🎯 RECOMMENDATIONS:")
	fmt.Println("===================")
	fmt.Println("1. Test with longer historical data periods for more robust results")
	fmt.Println("2. Consider implementing stop-loss and take-profit levels")
	fmt.Println("3. Analyze performance across different market conditions")
	fmt.Println("4. Implement portfolio diversification across multiple currency pairs")
	fmt.Println("5. Consider transaction costs and slippage in live trading")
}