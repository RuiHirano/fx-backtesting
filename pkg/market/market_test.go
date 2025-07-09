package market

import (
	"context"
	"testing"

	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Market NewMarket テスト
func TestMarket_NewMarket(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	
	market := NewMarket(provider)
	
	if market == nil {
		t.Fatal("Expected market to be created")
	}
	
	// 初期化の確認
	if market.GetCurrentPrice("EURUSD") != 0.0 {
		t.Error("Expected initial price to be 0.0")
	}
	
	if market.IsFinished() {
		t.Error("Expected market to not be finished initially")
	}
}

// Market Forward テスト
func TestMarket_Forward(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化
	err := market.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 最初の価格取得
	initialPrice := market.GetCurrentPrice("EURUSD")
	if initialPrice == 0.0 {
		t.Error("Expected initial price to be set after Initialize")
	}
	
	// 時間進行
	hasNext := market.Forward()
	if !hasNext {
		t.Error("Expected to have next candle")
	}
	
	// 価格が更新されているか確認
	newPrice := market.GetCurrentPrice("EURUSD")
	if newPrice == 0.0 {
		t.Error("Expected price to be updated after Forward")
	}
}

// Market GetCurrentPrice テスト
func TestMarket_GetCurrentPrice(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化前
	price := market.GetCurrentPrice("EURUSD")
	if price != 0.0 {
		t.Errorf("Expected price 0.0 before initialization, got %f", price)
	}
	
	// 初期化
	err := market.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期化後
	price = market.GetCurrentPrice("EURUSD")
	if price <= 0.0 {
		t.Errorf("Expected positive price after initialization, got %f", price)
	}
	
	// 存在しないシンボル
	unknownPrice := market.GetCurrentPrice("UNKNOWN")
	if unknownPrice != 0.0 {
		t.Errorf("Expected price 0.0 for unknown symbol, got %f", unknownPrice)
	}
}

// Market IsFinished テスト
func TestMarket_IsFinished(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化前
	if market.IsFinished() {
		t.Error("Expected market to not be finished before initialization")
	}
	
	// 初期化
	err := market.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期化後
	if market.IsFinished() {
		t.Error("Expected market to not be finished after initialization")
	}
	
	// 全データを消費
	for market.Forward() {
		// 全データを読み切る
	}
	
	// 終了確認
	if !market.IsFinished() {
		t.Error("Expected market to be finished after consuming all data")
	}
}

// Market GetCurrentTime テスト
func TestMarket_GetCurrentTime(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化前
	currentTime := market.GetCurrentTime()
	if !currentTime.IsZero() {
		t.Error("Expected zero time before initialization")
	}
	
	// 初期化
	err := market.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期化後
	currentTime = market.GetCurrentTime()
	if currentTime.IsZero() {
		t.Error("Expected non-zero time after initialization")
	}
	
	// 時間進行
	initialTime := currentTime
	market.Forward()
	
	newTime := market.GetCurrentTime()
	if !newTime.After(initialTime) {
		t.Error("Expected time to advance after Forward")
	}
}

// Market GetCurrentCandle テスト
func TestMarket_GetCurrentCandle(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化前
	candle := market.GetCurrentCandle("EURUSD")
	if candle != nil {
		t.Error("Expected nil candle before initialization")
	}
	
	// 初期化
	err := market.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期化後
	candle = market.GetCurrentCandle("EURUSD")
	if candle == nil {
		t.Error("Expected non-nil candle after initialization")
	}
	
	// ローソク足データの検証
	if err := candle.Validate(); err != nil {
		t.Errorf("Expected valid candle, got validation error: %v", err)
	}
	
	// 存在しないシンボル
	unknownCandle := market.GetCurrentCandle("UNKNOWN")
	if unknownCandle != nil {
		t.Error("Expected nil candle for unknown symbol")
	}
}

// Market Initialize エラーハンドリングテスト
func TestMarket_InitializeError(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/nonexistent.csv",
		Format:   "csv",
	}
	
	provider := data.NewCSVProvider(config)
	market := NewMarket(provider)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初期化でエラーが発生することを確認
	err := market.Initialize(ctx)
	if err == nil {
		t.Error("Expected error from Initialize with nonexistent file")
	}
}