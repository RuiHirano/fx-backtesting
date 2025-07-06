package broker

import (
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestMockBroker_GetBalance(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	balance := broker.GetBalance()
	if balance != config.InitialBalance {
		t.Errorf("Expected balance %v, got %v", config.InitialBalance, balance)
	}
}

func TestMockBroker_PlaceOrder_MarketBuy(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	// Set current market price
	broker.SetCurrentPrice("EURUSD", 1.0500)

	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())

	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// Check that position was created
	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(positions))
	}

	position := positions[0]
	if position.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", position.Symbol)
	}
	if position.Side != models.OrderSideBuy {
		t.Errorf("Expected buy side, got %v", position.Side)
	}
	if position.Size != 1000.0 {
		t.Errorf("Expected size 1000.0, got %v", position.Size)
	}

	// Check that balance was reduced by margin
	expectedMargin := config.CalculateMarginRequired(1000.0, 1.0500+config.Spread/2)
	expectedBalance := config.InitialBalance - expectedMargin
	currentBalance := broker.GetBalance()
	if currentBalance != expectedBalance {
		t.Errorf("Expected balance %v, got %v", expectedBalance, currentBalance)
	}
}

func TestMockBroker_PlaceOrder_MarketSell(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideSell, 1000.0, 0, 0, 0, time.Now())

	err := broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(positions))
	}

	position := positions[0]
	if position.Side != models.OrderSideSell {
		t.Errorf("Expected sell side, got %v", position.Side)
	}
}

func TestMockBroker_ClosePosition(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place an order to create a position
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())
	broker.PlaceOrder(order)

	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Fatalf("Expected 1 position after placing order")
	}

	positionID := positions[0].ID

	// Close the position
	err := broker.ClosePosition(positionID)
	if err != nil {
		t.Fatalf("ClosePosition failed: %v", err)
	}

	// Check that position was removed
	positions = broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions after closing, got %d", len(positions))
	}
}

func TestMockBroker_ClosePosition_InvalidID(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	err := broker.ClosePosition("invalid-id")
	if err == nil {
		t.Error("Expected error for invalid position ID")
	}
}

func TestMockBroker_InsufficientMargin(t *testing.T) {
	config := models.NewConfig(100.0, 0.0001, 0.0, 0.0, 1.0) // Low balance, no leverage
	broker := NewMockBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Try to place order that requires more margin than available balance
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())

	err := broker.PlaceOrder(order)
	if err == nil {
		t.Error("Expected error for insufficient margin")
	}
}

func TestMockBroker_UpdatePositions(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)

	// Place a buy order
	order := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())
	broker.PlaceOrder(order)

	// Change market price
	broker.SetCurrentPrice("EURUSD", 1.0520)

	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Fatalf("Expected 1 position")
	}

	position := positions[0]
	if position.CurrentPrice != 1.0520 {
		t.Errorf("Expected current price 1.0520, got %v", position.CurrentPrice)
	}

	// Check PnL calculation
	expectedPnL := (1.0520 - (1.0500 + config.Spread/2)) * 1000.0
	if position.PnL != expectedPnL {
		t.Errorf("Expected PnL %v, got %v", expectedPnL, position.PnL)
	}
}

func TestMockBroker_SetCurrentPrice(t *testing.T) {
	config := models.DefaultConfig()
	broker := NewMockBroker(config)

	broker.SetCurrentPrice("EURUSD", 1.0500)
	broker.SetCurrentPrice("GBPUSD", 1.2500)

	// This is testing internal state, we'll verify through position updates
	// Place orders and check execution prices
	order1 := models.NewOrder("order1", "EURUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())
	order2 := models.NewOrder("order2", "GBPUSD", models.OrderTypeMarket, models.OrderSideBuy, 1000.0, 0, 0, 0, time.Now())

	broker.PlaceOrder(order1)
	broker.PlaceOrder(order2)

	positions := broker.GetPositions()
	if len(positions) != 2 {
		t.Fatalf("Expected 2 positions, got %d", len(positions))
	}

	// Verify that different symbols have different entry prices
	eurPosition := positions[0]
	gbpPosition := positions[1]

	if eurPosition.EntryPrice == gbpPosition.EntryPrice {
		t.Error("Expected different entry prices for different symbols")
	}
}