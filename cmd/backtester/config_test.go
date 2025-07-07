package main

import (
	"os"
	"testing"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestLoadConfig_DefaultConfig(t *testing.T) {
	config, err := loadConfig("")
	if err != nil {
		t.Fatalf("Expected no error loading default config, got: %v", err)
	}
	
	if config == nil {
		t.Fatal("Expected config to be returned")
	}
	
	// Check default values
	if config.InitialBalance != 10000.0 {
		t.Errorf("Expected initial balance 10000.0, got %.2f", config.InitialBalance)
	}
	if config.Spread != 0.0001 {
		t.Errorf("Expected spread 0.0001, got %.6f", config.Spread)
	}
}

func TestLoadConfig_JSONFile(t *testing.T) {
	// Create temporary JSON config file
	jsonContent := `{
		"initial_balance": 5000.0,
		"spread": 0.0002,
		"commission": 0.5,
		"slippage": 0.0001,
		"leverage": 50
	}`
	
	tmpFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(jsonContent); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	tmpFile.Close()
	
	config, err := loadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error loading JSON config, got: %v", err)
	}
	
	// Check loaded values
	if config.InitialBalance != 5000.0 {
		t.Errorf("Expected initial balance 5000.0, got %.2f", config.InitialBalance)
	}
	if config.Spread != 0.0002 {
		t.Errorf("Expected spread 0.0002, got %.6f", config.Spread)
	}
	if config.Commission != 0.5 {
		t.Errorf("Expected commission 0.5, got %.2f", config.Commission)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	_, err := loadConfig("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create temporary invalid JSON file
	invalidJSON := `{
		"initial_balance": 5000.0,
		"spread": invalid_value
	}`
	
	tmpFile, err := os.CreateTemp("", "invalid_config_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(invalidJSON); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	tmpFile.Close()
	
	_, err = loadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON config")
	}
}

func TestConfigJSON_FromConfig(t *testing.T) {
	config := models.DefaultConfig()
	
	configJSON := ConfigJSON{
		InitialBalance: config.InitialBalance,
		Spread:         config.Spread,
		Commission:     config.Commission,
		Slippage:       config.Slippage,
		Leverage:       config.Leverage,
	}
	
	resultConfig := configJSON.ToConfig()
	
	if resultConfig.InitialBalance != config.InitialBalance {
		t.Errorf("Expected initial balance %.2f, got %.2f", 
			config.InitialBalance, resultConfig.InitialBalance)
	}
	if resultConfig.Spread != config.Spread {
		t.Errorf("Expected spread %.6f, got %.6f", 
			config.Spread, resultConfig.Spread)
	}
}