package broker

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestSimpleBroker_StopLoss(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewSimpleBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place buy order with stop loss
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 1.0450, 0, time.Now())
	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// Move price down to trigger stop loss
	broker.SetCurrentPrice("EURUSD", 1.0440)
	
	// Process stop loss triggers
	broker.ProcessStopLosses()

	// Position should be closed
	positions := broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions after stop loss trigger, got %d", len(positions))
	}
}

func TestSimpleBroker_TakeProfit(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewSimpleBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place buy order with take profit
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 1.0550, time.Now())
	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// Move price up to trigger take profit
	broker.SetCurrentPrice("EURUSD", 1.0560)
	
	// Process take profit triggers
	broker.ProcessTakeProfits()

	// Position should be closed
	positions := broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions after take profit trigger, got %d", len(positions))
	}
}

func TestSimpleBroker_LimitOrder(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewSimpleBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place limit buy order below current price
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeLimit, models.OrderSideBuy, 1000.0, 1.0480, 0, 0, time.Now())
	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// Order should be pending
	pendingOrders := broker.GetPendingOrders()
	if len(pendingOrders) != 1 {
		t.Errorf("Expected 1 pending order, got %d", len(pendingOrders))
	}

	// No position should exist yet
	positions := broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions before limit order execution, got %d", len(positions))
	}

	// Move price down to trigger limit order
	broker.SetCurrentPrice("EURUSD", 1.0470)
	broker.ProcessPendingOrders()

	// Order should be executed and position created
	positions = broker.GetPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 position after limit order execution, got %d", len(positions))
	}

	// Pending order should be removed
	pendingOrders = broker.GetPendingOrders()
	if len(pendingOrders) != 0 {
		t.Errorf("Expected 0 pending orders after execution, got %d", len(pendingOrders))
	}
}

func TestSimpleBroker_MarginCall(t *testing.T) {
	config := models.NewConfig(1000.0, 0.0001, 0.0, 0.0, 10.0) // Low balance, low leverage
	broker := NewSimpleBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place large position that uses most of the balance
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 9000.0, 0, 0, 0, time.Now())
	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// Move price down significantly to create large loss
	broker.SetCurrentPrice("EURUSD", 1.0200) // 300 pips down

	// Check margin level
	marginLevel := broker.GetMarginLevel()
	t.Logf("Margin level: %v", marginLevel)
	
	// Process margin calls
	broker.ProcessMarginCalls()

	// Check if position was closed due to margin call
	positions := broker.GetPositions()
	marginLevelAfter := broker.GetMarginLevel()
	t.Logf("Positions after margin call: %d, Margin level: %v", len(positions), marginLevelAfter)
	
	if marginLevel <= 50.0 && len(positions) != 0 {
		t.Errorf("Expected 0 positions after margin call at %v%%, got %d", marginLevel, len(positions))
	}
}

func TestSimpleBroker_GetEquity(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewSimpleBroker(config)

	// Initial equity should equal balance
	initialEquity := broker.GetEquity()
	if initialEquity != config.InitialBalance {
		t.Errorf("Expected initial equity %v, got %v", config.InitialBalance, initialEquity)
	}

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place order
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())
	broker.PlaceOrder(order)

	// Move price up (profitable)
	broker.SetCurrentPrice("EURUSD", 1.0520)

	equity := broker.GetEquity()
	balance := broker.GetBalance()

	// Equity should be higher than balance due to unrealized profit
	if equity <= balance {
		t.Errorf("Expected equity %v to be higher than balance %v", equity, balance)
	}
}

func TestSimpleBroker_GetUsedMargin(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewSimpleBroker(config)

	// Initially no margin used
	usedMargin := broker.GetUsedMargin()
	if usedMargin != 0 {
		t.Errorf("Expected initial used margin 0, got %v", usedMargin)
	}

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place order
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())
	broker.PlaceOrder(order)

	// Check used margin
	usedMargin = broker.GetUsedMargin()
	expectedMargin := config.CalculateMarginRequired(1000.0, 1.0500+config.Spread/2)
	if usedMargin != expectedMargin {
		t.Errorf("Expected used margin %v, got %v", expectedMargin, usedMargin)
	}
}