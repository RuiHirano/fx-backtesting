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
	fmt.Println("FX Backtesting Library - Strategy Debug")
	fmt.Println("======================================")

	// Create sample data - rising prices
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

	// Create configuration
	config := models.NewConfig(10000.0, 0.0001, 0.0, 0.0, 100.0)

	// Create components
	brokerInstance := broker.NewSimpleBroker(config)
	strategyInstance := strategy.NewMovingAverageStrategy("EURUSD", 2, 5, 10000.0)

	fmt.Println("Tracing strategy decision making...")

	// Manually process candles and trace strategy decisions
	for i, candle := range candles {
		fmt.Printf("\n=== Candle %d: Price %.4f ===\n", i+1, candle.Close)
		
		// Set price
		brokerInstance.SetCurrentPrice("EURUSD", candle.Close)
		
		// Process the candle to update moving averages
		err := strategyInstance.OnTick(candle, brokerInstance)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		
		// Check if strategy is ready
		if strategyInstance.IsReady() {
			signal := strategyInstance.GetSignal()
			fastMA := strategyInstance.GetFastMA()
			slowMA := strategyInstance.GetSlowMA()
			
			fmt.Printf("  FastMA: %.4f, SlowMA: %.4f\n", fastMA, slowMA)
			fmt.Printf("  Signal: %d (1=Buy, 2=Sell, 0=None)\n", signal)
			
			// Check positions
			positions := brokerInstance.GetPositions()
			fmt.Printf("  Positions: %d\n", len(positions))
			for _, pos := range positions {
				fmt.Printf("    - %s: Size %.0f, Entry %.4f, Current %.4f\n", 
					pos.ID, pos.Size, pos.EntryPrice, pos.CurrentPrice)
			}
			
			fmt.Printf("  Balance: $%.2f\n", brokerInstance.GetBalance())
		}
	}

	// Final state
	fmt.Printf("\n=== FINAL STATE ===\n")
	fmt.Printf("Final balance: $%.2f\n", brokerInstance.GetBalance())
	fmt.Printf("Total PnL: $%.2f\n", brokerInstance.GetBalance()-10000.0)
	fmt.Printf("Final positions: %d\n", len(brokerInstance.GetPositions()))
}