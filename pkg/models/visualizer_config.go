package models

import "time"

// VisualizerConfig はVisualizerの設定を管理します
type VisualizerConfig struct {
	// 基本設定
	Enabled bool `json:"enabled"` // Visualizerを有効にするかどうか
	Port    int  `json:"port"`    // WebSocketサーバーのポート

	// タイムアウト設定
	ReadTimeout  time.Duration `json:"read_timeout"`  // 読み取りタイムアウト
	WriteTimeout time.Duration `json:"write_timeout"` // 書き込みタイムアウト

	// 接続管理設定
	MaxClients        int           `json:"max_clients"`        // 最大クライアント数
	HeartbeatInterval time.Duration `json:"heartbeat_interval"` // ハートビート間隔
	ClientTimeout     time.Duration `json:"client_timeout"`     // クライアントタイムアウト

	// データ処理設定
	BufferSize    int           `json:"buffer_size"`    // バッファサイズ
	BatchSize     int           `json:"batch_size"`     // バッチサイズ
	FlushInterval time.Duration `json:"flush_interval"` // フラッシュ間隔

	// ログ設定
	LogLevel      string `json:"log_level"`      // ログレベル
	LogFile       string `json:"log_file"`       // ログファイルパス
	EnableMetrics bool   `json:"enable_metrics"` // メトリクス有効化
}

// DefaultVisualizerConfig はデフォルトのVisualizer設定を返します
func DefaultVisualizerConfig() VisualizerConfig {
	return VisualizerConfig{
		Enabled:           true,
		Port:              8080,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxClients:        100,
		HeartbeatInterval: 30 * time.Second,
		ClientTimeout:     90 * time.Second,
		BufferSize:        1024,
		BatchSize:         100,
		FlushInterval:     1 * time.Second,
		LogLevel:          "info",
		LogFile:           "",
		EnableMetrics:     false,
	}
}

// DisabledVisualizerConfig はVisualizerを無効にした設定を返します
func DisabledVisualizerConfig() VisualizerConfig {
	config := DefaultVisualizerConfig()
	config.Enabled = false
	return config
}

// Validate はVisualizerConfigの妥当性を検証します
func (vc *VisualizerConfig) Validate() error {
	if !vc.Enabled {
		return nil // 無効時は検証不要
	}

	if vc.Port <= 0 || vc.Port > 65535 {
		return &ValidationError{
			Field:   "Port",
			Value:   vc.Port,
			Message: "port must be between 1 and 65535",
		}
	}

	if vc.MaxClients <= 0 {
		return &ValidationError{
			Field:   "MaxClients",
			Value:   vc.MaxClients,
			Message: "max clients must be positive",
		}
	}

	if vc.BufferSize <= 0 {
		return &ValidationError{
			Field:   "BufferSize",
			Value:   vc.BufferSize,
			Message: "buffer size must be positive",
		}
	}

	if vc.ReadTimeout <= 0 {
		return &ValidationError{
			Field:   "ReadTimeout",
			Value:   vc.ReadTimeout,
			Message: "read timeout must be positive",
		}
	}

	if vc.WriteTimeout <= 0 {
		return &ValidationError{
			Field:   "WriteTimeout",
			Value:   vc.WriteTimeout,
			Message: "write timeout must be positive",
		}
	}

	return nil
}