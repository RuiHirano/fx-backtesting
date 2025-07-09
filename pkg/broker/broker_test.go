package broker

import (
	"context"
	"testing"

	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/market"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Broker PlaceOrder テスト
func TestBroker_PlaceOrder(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// 成行買い注文
	order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
	
	err = broker.PlaceOrder(order)
	if err != nil {
		t.Fatalf("Expected no error from PlaceOrder, got %v", err)
	}
	
	// ポジション確認
	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 position, got %d", len(positions))
	}
	
	// 残高確認
	balance := broker.GetBalance()
	if balance >= brokerConfig.InitialBalance {
		t.Error("Expected balance to decrease after buying")
	}
}

// Broker GetPositions テスト
func TestBroker_GetPositions(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// 初期状態
	positions := broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions initially, got %d", len(positions))
	}
	
	// 注文実行
	order1 := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
	broker.PlaceOrder(order1)
	
	order2 := models.NewMarketOrder("order-2", "EURUSD", models.Sell, 5000.0)
	broker.PlaceOrder(order2)
	
	// ポジション確認
	positions = broker.GetPositions()
	if len(positions) != 2 {
		t.Errorf("Expected 2 positions, got %d", len(positions))
	}
	
	// ポジション内容確認
	buyPosition := positions[0]
	if buyPosition.Side != models.Buy {
		t.Errorf("Expected Buy position, got %v", buyPosition.Side)
	}
	if buyPosition.Size != 10000.0 {
		t.Errorf("Expected size 10000.0, got %f", buyPosition.Size)
	}
}

// Broker GetBalance テスト
func TestBroker_GetBalance(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// 初期残高確認
	balance := broker.GetBalance()
	if balance != brokerConfig.InitialBalance {
		t.Errorf("Expected initial balance %f, got %f", brokerConfig.InitialBalance, balance)
	}
	
	// 注文実行後の残高確認
	order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
	broker.PlaceOrder(order)
	
	newBalance := broker.GetBalance()
	if newBalance >= balance {
		t.Error("Expected balance to decrease after placing order")
	}
}

// Broker ClosePosition テスト
func TestBroker_ClosePosition(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// ポジション作成
	order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
	broker.PlaceOrder(order)
	
	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Fatalf("Expected 1 position, got %d", len(positions))
	}
	
	// ポジションクローズ
	positionID := positions[0].ID
	err = broker.ClosePosition(positionID)
	if err != nil {
		t.Fatalf("Expected no error from ClosePosition, got %v", err)
	}
	
	// ポジション確認
	positions = broker.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions after close, got %d", len(positions))
	}
}

// Broker Market統合テスト
func TestBroker_MarketIntegration(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// 現在価格取得
	currentPrice := mkt.GetCurrentPrice("EURUSD")
	if currentPrice <= 0.0 {
		t.Fatal("Expected positive current price")
	}
	
	// 注文実行
	order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
	broker.PlaceOrder(order)
	
	// ポジション価格確認
	positions := broker.GetPositions()
	if len(positions) != 1 {
		t.Fatalf("Expected 1 position, got %d", len(positions))
	}
	
	position := positions[0]
	
	// スプレッドを考慮した価格確認（買いはAsk価格）
	expectedEntryPrice := currentPrice + brokerConfig.Spread
	if position.EntryPrice != expectedEntryPrice {
		t.Errorf("Expected entry price %f, got %f", expectedEntryPrice, position.EntryPrice)
	}
	
	// 元の価格を保存
	originalPrice := position.CurrentPrice
	
	// 市場時間進行
	mkt.Forward()
	
	// ポジション更新
	broker.UpdatePositions()
	
	// 価格更新確認
	updatedPositions := broker.GetPositions()
	if len(updatedPositions) != 1 {
		t.Fatalf("Expected 1 position after update, got %d", len(updatedPositions))
	}
	
	updatedPosition := updatedPositions[0]
	if updatedPosition.CurrentPrice == originalPrice {
		t.Error("Expected position price to be updated")
	}
}

// Broker エラーハンドリングテスト
func TestBroker_ErrorHandling(t *testing.T) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	provider := data.NewCSVProvider(config)
	mkt := market.NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 1000.0, // 少額
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	
	// 残高不足での注文
	order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 100000.0) // 大きな注文
	err = broker.PlaceOrder(order)
	if err == nil {
		t.Error("Expected error for insufficient balance")
	}
	
	// 存在しないポジションのクローズ
	err = broker.ClosePosition("nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent position")
	}
}