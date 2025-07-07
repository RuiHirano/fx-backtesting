package main

import (
	"fmt"
	"log"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

func main() {
	fmt.Println("FX Backtesting Library - Debug Example")
	fmt.Println("=====================================")

	// Create sample data
	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	candles := []models.Candle{
		{Open: 1.1000, High: 1.1010, Low: 1.0990, Close: 1.1005, Volume: 1000, Timestamp: startTime},
		{Open: 1.1005, High: 1.1015, Low: 1.0995, Close: 1.1010, Volume: 1000, Timestamp: startTime.Add(time.Minute)},
		{Open: 1.1010, High: 1.1020, Low: 1.1000, Close: 1.1015, Volume: 1000, Timestamp: startTime.Add(2 * time.Minute)},
		{Open: 1.1015, High: 1.1025, Low: 1.1005, Close: 1.1020, Volume: 1000, Timestamp: startTime.Add(3 * time.Minute)},
		{Open: 1.1020, High: 1.1030, Low: 1.1010, Close: 1.1025, Volume: 1000, Timestamp: startTime.Add(4 * time.Minute)},
		{Open: 1.1025, High: 1.1035, Low: 1.1015, Close: 1.1030, Volume: 1000, Timestamp: startTime.Add(5 * time.Minute)},
		{Open: 1.1030, High: 1.1040, Low: 1.1020, Close: 1.1035, Volume: 1000, Timestamp: startTime.Add(6 * time.Minute)},
		{Open: 1.1035, High: 1.1045, Low: 1.1025, Close: 1.1040, Volume: 1000, Timestamp: startTime.Add(7 * time.Minute)},
		{Open: 1.1040, High: 1.1050, Low: 1.1030, Close: 1.1045, Volume: 1000, Timestamp: startTime.Add(8 * time.Minute)},
		{Open: 1.1045, High: 1.1055, Low: 1.1035, Close: 1.1050, Volume: 1000, Timestamp: startTime.Add(9 * time.Minute)},
	}

	fmt.Printf("Loaded %d candles for backtesting\n", len(candles))

	// Create configuration
	config := models.NewConfig(
		10000.0, // Initial balance
		0.0001,  // Spread (1 pip)
		0.0,     // Commission (0%)
		0.0,     // Slippage
		100.0,   // Leverage
	)

	// Create components
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 2, 5, 10000.0)

	fmt.Println("Running backtest with debug information...")
	fmt.Printf("Initial balance: $%.2f\n", brokerInstance.GetBalance())

	// Run backtest with detailed tracking
	for i, candle := range candles {
		fmt.Printf("\nCandle %d: Price %.4f, Time: %s\n", i+1, candle.Close, candle.Timestamp.Format("15:04:05"))
		
		// Track balance before
		balanceBefore := brokerInstance.GetBalance()
		positionsBefore := brokerInstance.GetPositions()
		fmt.Printf("  Before: Balance $%.2f, Positions: %d\n", balanceBefore, len(positionsBefore))

		// Set current price
		brokerInstance.SetCurrentPrice("EURUSD", candle.Close)

		// Process broker operations
		brokerInstance.ProcessPendingOrders()
		brokerInstance.ProcessStopLosses()
		brokerInstance.ProcessTakeProfits()
		brokerInstance.ProcessMarginCalls()

		// Track balance after broker processing
		balanceAfterBroker := brokerInstance.GetBalance()
		positionsAfterBroker := brokerInstance.GetPositions()
		if balanceAfterBroker != balanceBefore {
			fmt.Printf("  After Broker Processing: Balance $%.2f (Change: $%.2f), Positions: %d\n", 
				balanceAfterBroker, balanceAfterBroker-balanceBefore, len(positionsAfterBroker))
		}

		// Check strategy state before execution
		if strategyInstance.IsReady() {
			signal := strategyInstance.GetSignal()
			fastMA := strategyInstance.GetFastMA()
			slowMA := strategyInstance.GetSlowMA()
			fmt.Printf("  Strategy: FastMA=%.4f, SlowMA=%.4f, Signal=%d\n", fastMA, slowMA, signal)
		}

		// Execute strategy
		err := strategyInstance.OnTick(candle, brokerInstance)
		if err != nil {
			log.Fatalf("Strategy error: %v", err)
		}

		// Track balance after strategy
		balanceAfter := brokerInstance.GetBalance()
		positionsAfter := brokerInstance.GetPositions()
		
		if balanceAfter != balanceAfterBroker {
			fmt.Printf("  After Strategy: Balance $%.2f (Change: $%.2f), Positions: %d\n", 
				balanceAfter, balanceAfter-balanceAfterBroker, len(positionsAfter))
		}

		// Show position details
		if len(positionsAfter) > 0 {
			fmt.Printf("  Current Positions:\n")
			for _, pos := range positionsAfter {
				fmt.Printf("    - ID: %s, Side: %s, Size: %.0f, Entry: %.4f, Current: %.4f\n", 
					pos.ID, pos.Side, pos.Size, pos.EntryPrice, pos.CurrentPrice)
			}
		}

		// Check for new positions
		if len(positionsAfter) > len(positionsAfterBroker) {
			fmt.Printf("  *** NEW POSITION OPENED! ***\n")
		}

		// Check for closed positions
		if len(positionsAfter) < len(positionsAfterBroker) {
			fmt.Printf("  *** POSITION CLOSED! ***\n")
		}

		// Calculate margin information
		usedMargin := brokerInstance.GetUsedMargin()
		equity := brokerInstance.GetEquity()
		freeMargin := brokerInstance.GetFreeMargin()
		
		if usedMargin > 0 {
			fmt.Printf("  Margin: Used $%.2f, Equity $%.2f, Free $%.2f\n", 
				usedMargin, equity, freeMargin)
		}
	}

	fmt.Printf("\nFinal balance: $%.2f\n", brokerInstance.GetBalance())
	fmt.Printf("Total PnL: $%.2f\n", brokerInstance.GetBalance()-10000.0)
	fmt.Printf("Final positions: %d\n", len(brokerInstance.GetPositions()))
}