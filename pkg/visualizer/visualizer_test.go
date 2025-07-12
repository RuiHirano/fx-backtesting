package visualizer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/gorilla/websocket"
)

// TestNewVisualizer は Visualizer インスタンスの作成をテスト
func TestNewVisualizer(t *testing.T) {
	// Red: テストを書く（失敗する状態）
	t.Run("should create visualizer with default config", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		if visualizer == nil {
			t.Error("Expected visualizer to be created, got nil")
		}
		
		config := visualizer.GetConfig()
		if config == nil {
			t.Error("Expected config to be set")
		}
		
		if config.Port != 8080 {
			t.Errorf("Expected default port to be 8080, got %d", config.Port)
		}
	})

	t.Run("should create visualizer with custom config", func(t *testing.T) {
		customConfig := &Config{
			Port:              9090,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      5 * time.Second,
			MaxClients:        50,
			HeartbeatInterval: 15 * time.Second,
			ClientTimeout:     45 * time.Second,
			BufferSize:        512,
			LogLevel:          "debug",
		}
		
		visualizer := NewVisualizer(customConfig)
		
		if visualizer == nil {
			t.Error("Expected visualizer to be created, got nil")
		}
		
		config := visualizer.GetConfig()
		if config.Port != 9090 {
			t.Errorf("Expected port to be 9090, got %d", config.Port)
		}
		
		if config.LogLevel != "debug" {
			t.Errorf("Expected log level to be 'debug', got '%s'", config.LogLevel)
		}
	})
}

// TestVisualizerLifecycle は Visualizer のライフサイクルをテスト
func TestVisualizerLifecycle(t *testing.T) {
	t.Run("should start and stop visualizer", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		// 初期状態は停止
		if visualizer.IsRunning() {
			t.Error("Expected visualizer to be stopped initially")
		}
		
		// 開始
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8081); err != nil {
			t.Errorf("Expected no error when starting, got %v", err)
		}
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// 実行中であることを確認
		if !visualizer.IsRunning() {
			t.Error("Expected visualizer to be running after start")
		}
		
		// 停止
		if err := visualizer.Stop(); err != nil {
			t.Errorf("Expected no error when stopping, got %v", err)
		}
		
		// 停止していることを確認
		if visualizer.IsRunning() {
			t.Error("Expected visualizer to be stopped after stop")
		}
	})

	t.Run("should not start twice", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8082); err != nil {
			t.Errorf("Expected no error on first start, got %v", err)
		}
		
		// 2回目の開始は失敗すべき
		if err := visualizer.Start(ctx, 8082); err == nil {
			t.Error("Expected error when starting twice")
		}
		
		visualizer.Stop()
	})
}

// TestWebSocketConnection は WebSocket 接続をテスト
func TestWebSocketConnection(t *testing.T) {
	t.Run("should handle websocket connection", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8083); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// WebSocket 接続を作成
		u := url.URL{Scheme: "ws", Host: "localhost:8083", Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("Failed to connect to websocket: %v", err)
		}
		defer conn.Close()
		
		// 接続数を確認
		if count := visualizer.GetConnectionCount(); count != 1 {
			t.Errorf("Expected connection count to be 1, got %d", count)
		}
	})
}

// TestMessageBroadcast はメッセージの配信をテスト
func TestMessageBroadcast(t *testing.T) {
	t.Run("should broadcast message to connected clients", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8084); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// WebSocket 接続を作成
		u := url.URL{Scheme: "ws", Host: "localhost:8084", Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("Failed to connect to websocket: %v", err)
		}
		defer conn.Close()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// メッセージを送信
		testMessage := Message{
			Type:      "test",
			Data:      "hello",
			Timestamp: time.Now(),
		}
		
		if err := visualizer.BroadcastMessage(testMessage); err != nil {
			t.Errorf("Failed to broadcast message: %v", err)
		}
		
		// メッセージを受信
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
		}
		
		var receivedMessage Message
		if err := json.Unmarshal(message, &receivedMessage); err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		
		if receivedMessage.Type != "test" {
			t.Errorf("Expected message type 'test', got '%s'", receivedMessage.Type)
		}
		
		if receivedMessage.Data != "hello" {
			t.Errorf("Expected message data 'hello', got '%v'", receivedMessage.Data)
		}
	})
}

// TestOnCandleUpdate はローソク足データの更新をテスト
func TestOnCandleUpdate(t *testing.T) {
	t.Run("should handle candle update", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8085); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// WebSocket 接続を作成
		u := url.URL{Scheme: "ws", Host: "localhost:8085", Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("Failed to connect to websocket: %v", err)
		}
		defer conn.Close()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// ローソク足データを作成
		candle := &models.Candle{
			Timestamp: time.Now(),
			Open:      150.0,
			High:      151.0,
			Low:       149.0,
			Close:     150.5,
			Volume:    1000,
		}
		
		// ローソク足データを送信
		if err := visualizer.OnCandleUpdate(candle); err != nil {
			t.Errorf("Failed to send candle update: %v", err)
		}
		
		// メッセージを受信
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
		}
		
		var receivedMessage Message
		if err := json.Unmarshal(message, &receivedMessage); err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		
		if receivedMessage.Type != "candle_update" {
			t.Errorf("Expected message type 'candle_update', got '%s'", receivedMessage.Type)
		}
		
		// データの検証
		candleData, ok := receivedMessage.Data.(map[string]interface{})
		if !ok {
			t.Error("Expected candle data to be a map")
		}
		
		if candleData["open"] != 150.0 {
			t.Errorf("Expected open price 150.0, got '%v'", candleData["open"])
		}
	})
}

// TestOnTradeEvent は取引イベントをテスト
func TestOnTradeEvent(t *testing.T) {
	t.Run("should handle trade event", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8086); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// WebSocket 接続を作成
		u := url.URL{Scheme: "ws", Host: "localhost:8086", Path: "/ws"}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			t.Errorf("Failed to connect to websocket: %v", err)
		}
		defer conn.Close()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// トレードデータを作成  
		trade := &models.Trade{
			ID:         "trade_1",
			Side:       models.Buy,
			Size:       1000,
			EntryPrice: 150.0,
			ExitPrice:  150.5,
			PnL:        500.0,
			Status:     models.TradeClosed,
			OpenTime:   time.Now().Add(-1 * time.Hour),
			CloseTime:  time.Now(),
			Duration:   1 * time.Hour,
		}
		
		// トレードイベントを送信
		if err := visualizer.OnTradeEvent(trade); err != nil {
			t.Errorf("Failed to send trade event: %v", err)
		}
		
		// メッセージを受信
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
		}
		
		var receivedMessage Message
		if err := json.Unmarshal(message, &receivedMessage); err != nil {
			t.Errorf("Failed to unmarshal message: %v", err)
		}
		
		if receivedMessage.Type != "trade_event" {
			t.Errorf("Expected message type 'trade_event', got '%s'", receivedMessage.Type)
		}
		
		// データの検証
		tradeData, ok := receivedMessage.Data.(map[string]interface{})
		if !ok {
			t.Error("Expected trade data to be a map")
		}
		
		if tradeData["symbol"] != "USDJPY" {
			t.Errorf("Expected symbol 'USDJPY', got '%v'", tradeData["symbol"])
		}
		
		if tradeData["side"] != float64(models.Buy) {
			t.Errorf("Expected side 'Buy', got '%v'", tradeData["side"])
		}
	})
}

// TestHealthEndpoint はヘルスチェックエンドポイントをテスト
func TestHealthEndpoint(t *testing.T) {
	t.Run("should return health status", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8087); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 少し待つ
		time.Sleep(100 * time.Millisecond)
		
		// HTTP リクエストを送信
		resp, err := http.Get("http://localhost:8087/health")
		if err != nil {
			t.Errorf("Failed to get health endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		// レスポンスを解析
		var health map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Errorf("Failed to decode health response: %v", err)
		}
		
		if health["status"] != "healthy" {
			t.Errorf("Expected status 'healthy', got '%v'", health["status"])
		}
		
		if health["running"] != true {
			t.Errorf("Expected running to be true, got '%v'", health["running"])
		}
	})
}

// TestConfigManagement は設定管理をテスト
func TestConfigManagement(t *testing.T) {
	t.Run("should not allow config change while running", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		ctx := context.Background()
		if err := visualizer.Start(ctx, 8088); err != nil {
			t.Errorf("Failed to start visualizer: %v", err)
		}
		defer visualizer.Stop()
		
		// 実行中に設定変更を試行
		newConfig := &Config{
			Port: 9999,
		}
		
		err := visualizer.SetConfig(newConfig)
		if err == nil {
			t.Error("Expected error when changing config while running")
		}
		
		if !strings.Contains(err.Error(), "cannot change config while running") {
			t.Errorf("Expected specific error message, got '%v'", err)
		}
	})
	
	t.Run("should allow config change when stopped", func(t *testing.T) {
		visualizer := NewVisualizer(nil)
		
		newConfig := &Config{
			Port:     9999,
			LogLevel: "debug",
		}
		
		if err := visualizer.SetConfig(newConfig); err != nil {
			t.Errorf("Expected no error when changing config while stopped, got %v", err)
		}
		
		config := visualizer.GetConfig()
		if config.Port != 9999 {
			t.Errorf("Expected port to be 9999, got %d", config.Port)
		}
		
		if config.LogLevel != "debug" {
			t.Errorf("Expected log level to be 'debug', got '%s'", config.LogLevel)
		}
	})
}