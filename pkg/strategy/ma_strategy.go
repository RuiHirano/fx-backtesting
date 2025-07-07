package strategy

import (
	"fmt"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/indicators"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Signal represents trading signals
type Signal int

const (
	SignalNone Signal = iota
	SignalBuy
	SignalSell
)

// MovingAverageStrategy implements a moving average crossover strategy
type MovingAverageStrategy struct {
	*StrategyBase
	symbol          string
	fastPeriod      int
	slowPeriod      int
	positionSize    float64
	fastMA          *indicators.SMA
	slowMA          *indicators.SMA
	lastFastValue   float64
	lastSlowValue   float64
	previousSignal  Signal
	hasPosition     bool
}

// NewMovingAverageStrategy creates a new moving average crossover strategy
func NewMovingAverageStrategy(symbol string, fastPeriod, slowPeriod int, positionSize float64) *MovingAverageStrategy {
	if fastPeriod >= slowPeriod {
		panic("Fast period must be less than slow period")
	}

	strategy := &MovingAverageStrategy{
		StrategyBase:   NewStrategyBase("MovingAverageStrategy"),
		symbol:         symbol,
		fastPeriod:     fastPeriod,
		slowPeriod:     slowPeriod,
		positionSize:   positionSize,
		fastMA:         indicators.NewSMA(fastPeriod),
		slowMA:         indicators.NewSMA(slowPeriod),
		previousSignal: SignalNone,
		hasPosition:    false,
	}

	// Set default parameters
	params := map[string]interface{}{
		"symbol":       symbol,
		"fastPeriod":   fastPeriod,
		"slowPeriod":   slowPeriod,
		"positionSize": positionSize,
	}
	strategy.SetParameters(params)

	return strategy
}

// OnTick processes new price data and generates trading signals
func (s *MovingAverageStrategy) OnTick(candle models.Candle, broker broker.Broker) error {
	// Update moving averages with closing price
	s.lastFastValue = s.fastMA.Update(candle.Close)
	s.lastSlowValue = s.slowMA.Update(candle.Close)

	// Only trade if both indicators are ready
	if !s.fastMA.IsReady() || !s.slowMA.IsReady() {
		return nil
	}

	// Get current signal
	signal := s.GetSignal()

	// Check for crossover signals
	var err error
	if signal == SignalBuy && s.previousSignal != SignalBuy {
		// Fast MA crossed above slow MA - Buy signal
		if s.hasPosition {
			// Close any existing positions first
			s.closeAllPositions(broker)
		}
		err = s.openPosition(models.OrderSideBuy, broker)
	} else if signal == SignalSell && s.previousSignal != SignalSell {
		// Fast MA crossed below slow MA - Sell signal
		if s.hasPosition {
			// Close any existing positions first
			s.closeAllPositions(broker)
		}
		err = s.openPosition(models.OrderSideSell, broker)
	}

	// Update previous signal for next comparison
	s.previousSignal = signal

	return err
}

// OnOrderFill handles order fill events
func (s *MovingAverageStrategy) OnOrderFill(order models.Order, broker broker.Broker) error {
	// Mark that we have a position
	s.hasPosition = true
	return nil
}

// GetSignal returns the current trading signal based on MA relationship
func (s *MovingAverageStrategy) GetSignal() Signal {
	if !s.fastMA.IsReady() || !s.slowMA.IsReady() {
		return SignalNone
	}

	if s.lastFastValue > s.lastSlowValue {
		return SignalBuy
	} else if s.lastFastValue < s.lastSlowValue {
		return SignalSell
	}

	return SignalNone
}

// openPosition opens a new position
func (s *MovingAverageStrategy) openPosition(side models.OrderSide, broker broker.Broker) error {
	// Generate unique order ID
	orderID := fmt.Sprintf("ma_strategy_%s_%d", s.symbol, time.Now().UnixNano())

	// Create market order
	order := models.NewOrder(
		orderID,
		s.symbol,
		models.OrderTypeMarket,
		side,
		s.positionSize,
		0, // Market price
		0, // No stop loss for now
		0, // No take profit for now
		time.Now(),
	)

	// Place order
	err := broker.PlaceOrder(order)
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}

	return nil
}

// closeAllPositions closes all open positions for this symbol
func (s *MovingAverageStrategy) closeAllPositions(broker broker.Broker) error {
	positions := broker.GetPositions()
	for _, position := range positions {
		if position.Symbol == s.symbol {
			err := broker.ClosePosition(position.ID)
			if err != nil {
				return fmt.Errorf("failed to close position %s: %w", position.ID, err)
			}
		}
	}
	s.hasPosition = false
	return nil
}

// Reset resets the strategy state
func (s *MovingAverageStrategy) Reset() {
	s.StrategyBase.Reset()
	s.fastMA.Reset()
	s.slowMA.Reset()
	s.lastFastValue = 0
	s.lastSlowValue = 0
	s.previousSignal = SignalNone
	s.hasPosition = false

	// Restore parameters
	params := map[string]interface{}{
		"symbol":       s.symbol,
		"fastPeriod":   s.fastPeriod,
		"slowPeriod":   s.slowPeriod,
		"positionSize": s.positionSize,
	}
	s.SetParameters(params)
}

// GetFastMA returns the fast moving average value
func (s *MovingAverageStrategy) GetFastMA() float64 {
	return s.lastFastValue
}

// GetSlowMA returns the slow moving average value
func (s *MovingAverageStrategy) GetSlowMA() float64 {
	return s.lastSlowValue
}

// IsReady returns true if the strategy has enough data to generate signals
func (s *MovingAverageStrategy) IsReady() bool {
	return s.fastMA.IsReady() && s.slowMA.IsReady()
}