package strategy

import (
	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Strategy defines the interface for trading strategies
type Strategy interface {
	// OnTick is called for each new price data point
	OnTick(candle models.Candle, broker broker.Broker) error
	
	// OnOrderFill is called when an order is filled
	OnOrderFill(order models.Order, broker broker.Broker) error
	
	// GetName returns the strategy name
	GetName() string
	
	// SetParameters sets strategy parameters
	SetParameters(params map[string]interface{}) error
	
	// GetParameters returns current strategy parameters
	GetParameters() map[string]interface{}
	
	// Reset resets the strategy state
	Reset()
}

// StrategyBase provides common functionality for strategies
type StrategyBase struct {
	name       string
	parameters map[string]interface{}
}

// NewStrategyBase creates a new strategy base
func NewStrategyBase(name string) *StrategyBase {
	return &StrategyBase{
		name:       name,
		parameters: make(map[string]interface{}),
	}
}

// GetName returns the strategy name
func (s *StrategyBase) GetName() string {
	return s.name
}

// SetParameters sets strategy parameters
func (s *StrategyBase) SetParameters(params map[string]interface{}) error {
	s.parameters = make(map[string]interface{})
	for k, v := range params {
		s.parameters[k] = v
	}
	return nil
}

// GetParameters returns current strategy parameters
func (s *StrategyBase) GetParameters() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range s.parameters {
		result[k] = v
	}
	return result
}

// Reset resets the strategy state
func (s *StrategyBase) Reset() {
	s.parameters = make(map[string]interface{})
}

// MockStrategy implements Strategy for testing purposes
type MockStrategy struct {
	*StrategyBase
	tickCount      int
	orderFillCount int
}

// NewMockStrategy creates a new mock strategy
func NewMockStrategy() *MockStrategy {
	return &MockStrategy{
		StrategyBase:   NewStrategyBase("MockStrategy"),
		tickCount:      0,
		orderFillCount: 0,
	}
}

// OnTick handles new price data
func (s *MockStrategy) OnTick(candle models.Candle, broker broker.Broker) error {
	s.tickCount++
	return nil
}

// OnOrderFill handles order fills
func (s *MockStrategy) OnOrderFill(order models.Order, broker broker.Broker) error {
	s.orderFillCount++
	return nil
}

// GetTickCount returns the number of ticks processed
func (s *MockStrategy) GetTickCount() int {
	return s.tickCount
}

// GetOrderFillCount returns the number of order fills processed
func (s *MockStrategy) GetOrderFillCount() int {
	return s.orderFillCount
}

// Reset resets the mock strategy state
func (s *MockStrategy) Reset() {
	s.StrategyBase.Reset()
	s.tickCount = 0
	s.orderFillCount = 0
}