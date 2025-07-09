package data

import (
	"context"
	"testing"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// DataProvider StreamData テスト
func TestDataProvider_StreamData(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	provider := NewCSVProvider(config)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	candleChan, err := provider.StreamData(ctx)
	if err != nil {
		t.Fatalf("Expected no error from StreamData, got %v", err)
	}
	
	// 最初のローソク足を取得
	candleData, ok := <-candleChan
	if !ok {
		t.Fatal("Expected to receive candle from channel")
	}
	
	if candleData.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", candleData.Symbol)
	}
	
	// チャネルクローズまで読み取り
	candleCount := 1
	for range candleChan {
		candleCount++
	}
	
	if candleCount == 0 {
		t.Error("Expected to receive at least one candle")
	}
}

// CSVProvider 初期化テスト
func TestCSVProvider_Initialize(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv", 
		Format:   "csv",
	}
	
	provider := NewCSVProvider(config)
	
	if provider == nil {
		t.Fatal("Expected provider to be created")
	}
	
	// 設定の確認
	if provider.Config.FilePath != config.FilePath {
		t.Errorf("Expected FilePath %s, got %s", config.FilePath, provider.Config.FilePath)
	}
	
	if provider.Config.Format != config.Format {
		t.Errorf("Expected Format %s, got %s", config.Format, provider.Config.Format)
	}
}

// CSVProvider エラーハンドリングテスト
func TestCSVProvider_ErrorHandling(t *testing.T) {
	// 存在しないファイル
	config := models.DataProviderConfig{
		FilePath: "./testdata/nonexistent.csv",
		Format:   "csv",
	}
	
	provider := NewCSVProvider(config)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	_, err := provider.StreamData(ctx)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// CSVParser パースエラーテスト
func TestCSVParser_ParseError(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/invalid.csv",
		Format:   "csv",
	}
	
	provider := NewCSVProvider(config)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	candleChan, err := provider.StreamData(ctx)
	if err != nil {
		// ファイルが存在しない場合はStreamDataでエラー
		return
	}
	
	// チャネルからの読み取りでエラーが発生することを確認
	for candleData := range candleChan {
		// 無効なデータでもCandleが作成される場合はバリデーションでチェック
		if err := candleData.Candle.Validate(); err == nil {
			t.Error("Expected validation error for invalid candle data")
		}
		break // 最初のエントリのみテスト
	}
}

// 大容量データテスト
func TestCSVProvider_LargeData(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "./testdata/large.csv",
		Format:   "csv",
	}
	
	provider := NewCSVProvider(config)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	candleChan, err := provider.StreamData(ctx)
	if err != nil {
		// ファイルが存在しない場合はスキップ
		t.Skip("Large test file not available")
	}
	
	candleCount := 0
	for range candleChan {
		candleCount++
		// メモリ使用量のテストのため、適当な数で止める
		if candleCount >= 1000 {
			cancel()
			break
		}
	}
	
	if candleCount == 0 {
		t.Error("Expected to process some candles")
	}
}