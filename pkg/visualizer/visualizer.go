package visualizer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Visualizer は、バックテストエンジンとフロントエンド間の通信を管理するインターフェース
type Visualizer interface {
	// ライフサイクル管理
	Start(ctx context.Context, port int) error
	Stop() error
	IsRunning() bool

	// バックテストエンジンからのイベント受信
	OnCandleUpdate(candle *models.Candle) error
	OnTradeEvent(trade *models.Trade) error
	OnStatisticsUpdate(stats *models.Statistics) error
	OnBacktestStateChange(state BacktestState) error

	// フロントエンドからのコマンド処理
	OnControlCommand(cmd *ControlCommand) error

	// 接続管理
	GetConnectionCount() int
	BroadcastMessage(message interface{}) error

	// 設定
	SetConfig(config *Config) error
	GetConfig() *Config
}

// BacktestState はバックテストの状態を表す
type BacktestState int

const (
	BacktestStateIdle BacktestState = iota
	BacktestStateRunning
	BacktestStatePaused
	BacktestStateStopped
	BacktestStateCompleted
	BacktestStateError
)

// ControlCommand はフロントエンドからの制御コマンドを表す
type ControlCommand struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	ClientID  string                 `json:"client_id"`
	Timestamp time.Time              `json:"timestamp"`
}

// Message は WebSocket で送信されるメッセージの基本構造
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ClientID  string      `json:"client_id,omitempty"`
}

// Config は Visualizer の設定を管理
type Config struct {
	Port              int           `json:"port"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	MaxClients        int           `json:"max_clients"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	ClientTimeout     time.Duration `json:"client_timeout"`
	BufferSize        int           `json:"buffer_size"`
	LogLevel          string        `json:"log_level"`
}

// DefaultConfig はデフォルトの設定を返す
func DefaultConfig() *Config {
	return &Config{
		Port:              8080,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxClients:        100,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     90 * time.Second,
		BufferSize:        1024,
		LogLevel:          "info",
	}
}

// visualizerImpl は Visualizer インターフェースの実装
type visualizerImpl struct {
	config       *Config
	server       *http.Server
	upgrader     websocket.Upgrader
	clients      map[string]*Client
	clientsMutex sync.RWMutex
	isRunning    bool
	runningMutex sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	hub          *Hub
}

// Client は WebSocket クライアントを表す
type Client struct {
	id           string
	conn         *websocket.Conn
	send         chan []byte
	hub          *Hub
	lastActivity time.Time
	isActive     bool
	mutex        sync.RWMutex
}

// Hub は複数のクライアントを管理する
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewVisualizer は新しい Visualizer インスタンスを作成
func NewVisualizer(config *Config) Visualizer {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	hub := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	return &visualizerImpl{
		config:   config,
		clients:  make(map[string]*Client),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 開発環境用: 本番環境では適切な Origin チェックを実装
			},
		},
		ctx:    ctx,
		cancel: cancel,
		hub:    hub,
	}
}

// Start は Visualizer を開始
func (v *visualizerImpl) Start(ctx context.Context, port int) error {
	v.runningMutex.Lock()
	defer v.runningMutex.Unlock()

	if v.isRunning {
		return fmt.Errorf("visualizer is already running")
	}

	if port > 0 {
		v.config.Port = port
	}

	// Hub を開始
	go v.hub.run()

	// HTTP サーバーを設定
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", v.handleWebSocket)
	mux.HandleFunc("/health", v.handleHealth)

	v.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", v.config.Port),
		Handler:      mux,
		ReadTimeout:  v.config.ReadTimeout,
		WriteTimeout: v.config.WriteTimeout,
	}

	v.isRunning = true

	go func() {
		if err := v.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	fmt.Printf("Visualizer started on port %d\n", v.config.Port)
	return nil
}

// Stop は Visualizer を停止
func (v *visualizerImpl) Stop() error {
	v.runningMutex.Lock()
	defer v.runningMutex.Unlock()

	if !v.isRunning {
		return nil
	}

	v.cancel()

	// 全てのクライアント接続を閉じる
	v.clientsMutex.Lock()
	for _, client := range v.clients {
		client.conn.Close()
	}
	v.clientsMutex.Unlock()

	// サーバーを停止
	if v.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		v.server.Shutdown(ctx)
	}

	v.isRunning = false
	fmt.Println("Visualizer stopped")
	return nil
}

// IsRunning は Visualizer の実行状態を返す
func (v *visualizerImpl) IsRunning() bool {
	v.runningMutex.RLock()
	defer v.runningMutex.RUnlock()
	return v.isRunning
}

// OnCandleUpdate はローソク足データの更新を処理
func (v *visualizerImpl) OnCandleUpdate(candle *models.Candle) error {
	message := Message{
		Type:      "candle_update",
		Data:      candle,
		Timestamp: time.Now(),
	}

	return v.BroadcastMessage(message)
}

// OnTradeEvent は取引イベントを処理
func (v *visualizerImpl) OnTradeEvent(trade *models.Trade) error {
	message := Message{
		Type:      "trade_event",
		Data:      trade,
		Timestamp: time.Now(),
	}

	return v.BroadcastMessage(message)
}

// OnStatisticsUpdate は統計情報の更新を処理
func (v *visualizerImpl) OnStatisticsUpdate(stats *models.Statistics) error {
	message := Message{
		Type:      "statistics_update",
		Data:      stats,
		Timestamp: time.Now(),
	}

	return v.BroadcastMessage(message)
}

// OnBacktestStateChange はバックテストの状態変更を処理
func (v *visualizerImpl) OnBacktestStateChange(state BacktestState) error {
	message := Message{
		Type:      "backtest_state",
		Data:      state,
		Timestamp: time.Now(),
	}

	return v.BroadcastMessage(message)
}

// OnControlCommand はフロントエンドからの制御コマンドを処理
func (v *visualizerImpl) OnControlCommand(cmd *ControlCommand) error {
	// TODO: 実際のコマンド処理ロジックを実装
	fmt.Printf("Received control command: %s\n", cmd.Type)
	return nil
}

// GetConnectionCount は接続数を返す
func (v *visualizerImpl) GetConnectionCount() int {
	v.clientsMutex.RLock()
	defer v.clientsMutex.RUnlock()
	return len(v.clients)
}

// BroadcastMessage は全てのクライアントにメッセージを送信
func (v *visualizerImpl) BroadcastMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	v.hub.broadcast <- data
	return nil
}

// SetConfig は設定を更新
func (v *visualizerImpl) SetConfig(config *Config) error {
	if v.IsRunning() {
		return fmt.Errorf("cannot change config while running")
	}
	v.config = config
	return nil
}

// GetConfig は現在の設定を返す
func (v *visualizerImpl) GetConfig() *Config {
	return v.config
}

// handleWebSocket は WebSocket 接続を処理
func (v *visualizerImpl) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := v.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	client := &Client{
		id:           fmt.Sprintf("client_%d", time.Now().UnixNano()),
		conn:         conn,
		send:         make(chan []byte, v.config.BufferSize),
		hub:          v.hub,
		lastActivity: time.Now(),
		isActive:     true,
	}

	v.clientsMutex.Lock()
	v.clients[client.id] = client
	v.clientsMutex.Unlock()

	client.hub.register <- client

	// クライアントの読み書きを開始
	go client.writePump()
	go client.readPump()

	fmt.Printf("Client %s connected\n", client.id)
}

// handleHealth はヘルスチェックエンドポイント
func (v *visualizerImpl) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":      "healthy",
		"connections": v.GetConnectionCount(),
		"running":     v.IsRunning(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Hub の実行ループ
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// readPump はクライアントからのメッセージを処理
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		c.mutex.Lock()
		c.lastActivity = time.Now()
		c.mutex.Unlock()

		// メッセージを処理（今後実装）
		fmt.Printf("Received message from %s: %s\n", c.id, string(message))
	}
}

// writePump はクライアントへのメッセージを送信
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}