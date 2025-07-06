package strategy

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestMockStrategy_OnTick(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMockStrategy()

	candle := models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000)

	err := strategy.OnTick(candle, mockBroker)
	if err != nil {
		t.Fatalf("OnTick failed: %v", err)
	}

	// Check that OnTick was called
	if strategy.GetTickCount() != 1 {
		t.Errorf("Expected tick count 1, got %d", strategy.GetTickCount())
	}
}

func TestMockStrategy_OnOrderFill(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMockStrategy()

	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 1.0500, 0, 0, time.Now())

	err := strategy.OnOrderFill(order, mockBroker)
	if err != nil {
		t.Fatalf("OnOrderFill failed: %v", err)
	}

	// Check that OnOrderFill was called
	if strategy.GetOrderFillCount() != 1 {
		t.Errorf("Expected order fill count 1, got %d", strategy.GetOrderFillCount())
	}
}

func TestMockStrategy_GetName(t *testing.T) {
	strategy := NewMockStrategy()
	name := strategy.GetName()
	
	if name == "" {
		t.Error("Expected non-empty strategy name")
	}
}

func TestMockStrategy_Reset(t *testing.T) {
	config := models.DefaultConfig()
	mockBroker := broker.NewMockBroker(config)
	strategy := NewMockStrategy()

	// Generate some activity
	candle := models.NewCandle(time.Now(), 1.0500, 1.0520, 1.0490, 1.0510, 1000)
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 1.0500, 0, 0, time.Now())

	strategy.OnTick(candle, mockBroker)
	strategy.OnOrderFill(order, mockBroker)

	// Reset strategy
	strategy.Reset()

	// Check that counters were reset
	if strategy.GetTickCount() != 0 {
		t.Errorf("Expected tick count 0 after reset, got %d", strategy.GetTickCount())
	}
	if strategy.GetOrderFillCount() != 0 {
		t.Errorf("Expected order fill count 0 after reset, got %d", strategy.GetOrderFillCount())
	}
}

func TestMockStrategy_SetParameters(t *testing.T) {
	strategy := NewMockStrategy()

	params := map[string]interface{}{
		"period": 20,
		"threshold": 0.01,
	}

	err := strategy.SetParameters(params)
	if err != nil {
		t.Fatalf("SetParameters failed: %v", err)
	}

	// Verify parameters were set
	retrievedParams := strategy.GetParameters()
	if retrievedParams["period"] != 20 {
		t.Errorf("Expected period 20, got %v", retrievedParams["period"])
	}
	if retrievedParams["threshold"] != 0.01 {
		t.Errorf("Expected threshold 0.01, got %v", retrievedParams["threshold"])
	}
}

func TestStrategyBase_GetName(t *testing.T) {
	base := NewStrategyBase("TestStrategy")
	
	if base.GetName() != "TestStrategy" {
		t.Errorf("Expected name 'TestStrategy', got %s", base.GetName())
	}
}

func TestStrategyBase_SetParameters(t *testing.T) {
	base := NewStrategyBase("TestStrategy")

	params := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}

	err := base.SetParameters(params)
	if err != nil {
		t.Fatalf("SetParameters failed: %v", err)
	}

	retrievedParams := base.GetParameters()
	if retrievedParams["param1"] != "value1" {
		t.Errorf("Expected param1 'value1', got %v", retrievedParams["param1"])
	}
	if retrievedParams["param2"] != 42 {
		t.Errorf("Expected param2 42, got %v", retrievedParams["param2"])
	}
}

func TestStrategyBase_Reset(t *testing.T) {
	base := NewStrategyBase("TestStrategy")

	params := map[string]interface{}{
		"param1": "value1",
	}
	base.SetParameters(params)

	base.Reset()

	// Parameters should be cleared after reset
	retrievedParams := base.GetParameters()
	if len(retrievedParams) != 0 {
		t.Errorf("Expected empty parameters after reset, got %v", retrievedParams)
	}
}