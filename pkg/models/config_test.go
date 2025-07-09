package models

import "testing"

// Config構造体のテスト
func TestConfig_NewDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()
	
	if config.Market.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", config.Market.Symbol)
	}
	
	if config.Market.DataProvider.Format != "csv" {
		t.Errorf("Expected format csv, got %s", config.Market.DataProvider.Format)
	}
	
	if config.Broker.InitialBalance != 10000.0 {
		t.Errorf("Expected initial balance 10000.0, got %f", config.Broker.InitialBalance)
	}
	
	if config.Broker.Spread != 0.0001 {
		t.Errorf("Expected spread 0.0001, got %f", config.Broker.Spread)
	}
}

func TestConfig_Validate(t *testing.T) {
	config := NewDefaultConfig()
	config.Market.DataProvider.FilePath = "./testdata/sample.csv"
	
	// 正常なケース
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}
	
	// 異常なケース - 空のファイルパス
	config.Market.DataProvider.FilePath = ""
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty file path")
	}
	
	// 異常なケース - 負の初期残高
	config = NewDefaultConfig()
	config.Market.DataProvider.FilePath = "./testdata/sample.csv"
	config.Broker.InitialBalance = -1000
	if err := config.Validate(); err == nil {
		t.Error("Expected error for negative initial balance")
	}
}

func TestMarketConfig_Validate(t *testing.T) {
	config := MarketConfig{
		DataProvider: DataProviderConfig{
			FilePath: "./testdata/sample.csv",
			Format:   "csv",
		},
		Symbol: "EURUSD",
	}
	
	// 正常なケース
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error for valid market config, got %v", err)
	}
	
	// 異常なケース - 空のシンボル
	config.Symbol = ""
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty symbol")
	}
}

func TestDataProviderConfig_Validate(t *testing.T) {
	config := DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	// 正常なケース
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error for valid data provider config, got %v", err)
	}
	
	// 異常なケース - 空のファイルパス
	config.FilePath = ""
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty file path")
	}
	
	// 異常なケース - 無効なフォーマット
	config.FilePath = "./testdata/sample.csv"
	config.Format = "xml"
	if err := config.Validate(); err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestBrokerConfig_Validate(t *testing.T) {
	config := BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	// 正常なケース
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error for valid broker config, got %v", err)
	}
	
	// 異常なケース - 負の初期残高
	config.InitialBalance = -1000
	if err := config.Validate(); err == nil {
		t.Error("Expected error for negative initial balance")
	}
	
	// 異常なケース - 負のスプレッド
	config.InitialBalance = 10000.0
	config.Spread = -0.0001
	if err := config.Validate(); err == nil {
		t.Error("Expected error for negative spread")
	}
}