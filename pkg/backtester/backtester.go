package backtester

import (
	"fmt"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/strategy"
)

// Progress represents backtest progress information
type Progress struct {
	ProcessedCandles int
	TotalCandles     int
	Percentage       float64
	CurrentTime      time.Time
}

// Result represents backtest results
type Result struct {
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	InitialBalance float64
	FinalBalance   float64
	TotalPnL       float64
	TotalTrades    int
	WinningTrades  int
	LosingTrades   int
	WinRate        float64
	MaxDrawdown    float64
	Trades         []TradeResult
}

// TradeResult represents a single trade result
type TradeResult struct {
	EntryTime  time.Time
	ExitTime   time.Time
	Symbol     string
	Side       models.OrderSide
	Size       float64
	EntryPrice float64
	ExitPrice  float64
	PnL        float64
	Duration   time.Duration
}

// ProgressCallback is called during backtest execution to report progress
type ProgressCallback func(progress Progress)

// Backtester executes backtests
type Backtester struct {
	dataProvider data.DataProvider
	broker       broker.Broker
	strategy     strategy.Strategy
	config       models.Config
	progress     Progress
}

// NewBacktester creates a new backtester
func NewBacktester(dataProvider data.DataProvider, brokerInstance broker.Broker, strategyInstance strategy.Strategy, config models.Config) *Backtester {
	return &Backtester{
		dataProvider: dataProvider,
		broker:       brokerInstance,
		strategy:     strategyInstance,
		config:       config,
		progress:     Progress{},
	}
}

// Run executes a backtest with the given candle data
func (b *Backtester) Run(candles []models.Candle) (*Result, error) {
	return b.RunWithCallback(candles, nil)
}

// RunWithCallback executes a backtest with progress callbacks
func (b *Backtester) RunWithCallback(candles []models.Candle, callback ProgressCallback) (*Result, error) {
	if len(candles) == 0 {
		return nil, fmt.Errorf("no data provided for backtesting")
	}

	// Reset state
	b.Reset()

	// Initialize progress
	b.progress.TotalCandles = len(candles)
	b.progress.ProcessedCandles = 0

	startTime := time.Now()
	result := &Result{
		StartTime:      candles[0].Timestamp,
		EndTime:        candles[len(candles)-1].Timestamp,
		InitialBalance: b.config.InitialBalance,
		Trades:         make([]TradeResult, 0),
	}
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Track positions for trade calculation
	var openPositions []models.Position
	var lastBalance = b.broker.GetBalance()

	// Process each candle
	for i, candle := range candles {
		// Update broker with current price
		symbol := "EURUSD" // Default symbol - in a real implementation this would come from the data
		if simpleBroker, ok := b.broker.(*broker.SimpleBroker); ok {
			simpleBroker.SetCurrentPrice(symbol, candle.Close)
			
			// Process pending orders and risk management
			simpleBroker.ProcessPendingOrders()
			simpleBroker.ProcessStopLosses()
			simpleBroker.ProcessTakeProfits()
			simpleBroker.ProcessMarginCalls()
		} else if mockBroker, ok := b.broker.(*broker.MockBroker); ok {
			mockBroker.SetCurrentPrice(symbol, candle.Close)
		}

		// Get positions before strategy execution
		positionsBefore := b.broker.GetPositions()

		// Execute strategy
		err := b.strategy.OnTick(candle, b.broker)
		if err != nil {
			return nil, fmt.Errorf("strategy error at candle %d: %w", i, err)
		}

		// Get positions after strategy execution
		positionsAfter := b.broker.GetPositions()

		// Check for new positions (trades)
		if len(positionsAfter) > len(positionsBefore) {
			// New position opened
			for _, pos := range positionsAfter {
				found := false
				for _, oldPos := range positionsBefore {
					if pos.ID == oldPos.ID {
						found = true
						break
					}
				}
				if !found {
					openPositions = append(openPositions, pos)
				}
			}
		}

		// Check for closed positions
		for _, oldPos := range positionsBefore {
			found := false
			for _, pos := range positionsAfter {
				if pos.ID == oldPos.ID {
					found = true
					break
				}
			}
			if !found {
				// Position was closed, calculate trade result
				currentBalance := b.broker.GetBalance()
				pnl := currentBalance - lastBalance
				lastBalance = currentBalance

				trade := TradeResult{
					EntryTime:  oldPos.OpenTime,
					ExitTime:   candle.Timestamp,
					Symbol:     oldPos.Symbol,
					Side:       oldPos.Side,
					Size:       oldPos.Size,
					EntryPrice: oldPos.EntryPrice,
					ExitPrice:  candle.Close,
					PnL:        pnl,
					Duration:   candle.Timestamp.Sub(oldPos.OpenTime),
				}
				result.Trades = append(result.Trades, trade)

				// Remove from open positions
				for j, openPos := range openPositions {
					if openPos.ID == oldPos.ID {
						openPositions = append(openPositions[:j], openPositions[j+1:]...)
						break
					}
				}
			}
		}

		// Update progress
		b.progress.ProcessedCandles = i + 1
		b.progress.Percentage = float64(i+1) / float64(len(candles)) * 100
		b.progress.CurrentTime = candle.Timestamp

		// Call progress callback if provided
		if callback != nil {
			callback(b.progress)
		}
	}

	// Calculate final results
	result.FinalBalance = b.broker.GetBalance()
	result.TotalPnL = result.FinalBalance - result.InitialBalance
	result.TotalTrades = len(result.Trades)

	// Calculate trade statistics
	winningTrades := 0
	losingTrades := 0
	maxDrawdown := 0.0
	runningBalance := result.InitialBalance

	for _, trade := range result.Trades {
		if trade.PnL > 0 {
			winningTrades++
		} else if trade.PnL < 0 {
			losingTrades++
		}

		runningBalance += trade.PnL
		drawdown := (result.InitialBalance - runningBalance) / result.InitialBalance * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	result.WinningTrades = winningTrades
	result.LosingTrades = losingTrades
	if result.TotalTrades > 0 {
		result.WinRate = float64(winningTrades) / float64(result.TotalTrades) * 100
	}
	result.MaxDrawdown = maxDrawdown

	executionTime := time.Since(startTime)
	fmt.Printf("Backtest completed in %v\n", executionTime)

	return result, nil
}

// GetProgress returns the current backtest progress
func (b *Backtester) GetProgress() Progress {
	return b.progress
}

// GetConfig returns the backtester configuration
func (b *Backtester) GetConfig() models.Config {
	return b.config
}

// Reset resets the backtester state
func (b *Backtester) Reset() {
	b.progress = Progress{}
	b.strategy.Reset()
	
	// Reset broker if it has a reset method
	if resettable, ok := b.broker.(interface{ Reset() }); ok {
		resettable.Reset()
	}
}