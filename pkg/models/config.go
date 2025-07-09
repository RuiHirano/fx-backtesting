package models

import (
	"errors"
	"strings"
)

// Config はバックテスト全体の設定を管理します。
type Config struct {
	Market MarketConfig `json:"market"`
	Broker BrokerConfig `json:"broker"`
}

// MarketConfig は市場データに関する設定です。
type MarketConfig struct {
	DataProvider DataProviderConfig `json:"data_provider"`
	Symbol       string             `json:"symbol"`
}

// DataProviderConfig はデータソースに関する設定です。
type DataProviderConfig struct {
	FilePath string `json:"file_path"`
	Format   string `json:"format"`
}

// BrokerConfig はブローカーに関する設定です。
type BrokerConfig struct {
	InitialBalance float64 `json:"initial_balance"`
	Spread         float64 `json:"spread"`
}

// NewDefaultConfig はデフォルト設定を生成します。
func NewDefaultConfig() Config {
	return Config{
		Market: MarketConfig{
			DataProvider: DataProviderConfig{
				Format: "csv",
			},
			Symbol: "EURUSD",
		},
		Broker: BrokerConfig{
			InitialBalance: 10000.0,
			Spread:         0.0001, // 1 pip
		},
	}
}

// Validate は設定の妥当性を検証します。
func (c *Config) Validate() error {
	if err := c.Market.Validate(); err != nil {
		return err
	}
	
	if err := c.Broker.Validate(); err != nil {
		return err
	}
	
	return nil
}

// Validate はMarketConfigの妥当性を検証します。
func (mc *MarketConfig) Validate() error {
	if err := mc.DataProvider.Validate(); err != nil {
		return err
	}
	
	if strings.TrimSpace(mc.Symbol) == "" {
		return errors.New("symbol is required")
	}
	
	return nil
}

// Validate はDataProviderConfigの妥当性を検証します。
func (dpc *DataProviderConfig) Validate() error {
	if strings.TrimSpace(dpc.FilePath) == "" {
		return errors.New("file path is required")
	}
	
	if dpc.Format != "csv" && dpc.Format != "json" {
		return errors.New("format must be 'csv' or 'json'")
	}
	
	return nil
}

// Validate はBrokerConfigの妥当性を検証します。
func (bc *BrokerConfig) Validate() error {
	if bc.InitialBalance <= 0 {
		return errors.New("initial balance must be positive")
	}
	
	if bc.Spread < 0 {
		return errors.New("spread must be non-negative")
	}
	
	return nil
}