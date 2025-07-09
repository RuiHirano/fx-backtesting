package models

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()
	
	if id1 == id2 {
		t.Error("Expected different IDs, got the same")
	}
	
	if len(id1) == 0 {
		t.Error("Expected non-empty ID")
	}
}

func TestParseOrderSide(t *testing.T) {
	tests := []struct {
		input    string
		expected OrderSide
		hasError bool
	}{
		{"buy", Buy, false},
		{"Buy", Buy, false},
		{"BUY", Buy, false},
		{"sell", Sell, false},
		{"Sell", Sell, false},
		{"SELL", Sell, false},
		{"invalid", Buy, true},
		{"", Buy, true},
	}
	
	for _, test := range tests {
		result, err := ParseOrderSide(test.input)
		
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s'", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v for input '%s', got %v", test.expected, test.input, result)
			}
		}
	}
}

func TestParseOrderType(t *testing.T) {
	tests := []struct {
		input    string
		expected OrderType
		hasError bool
	}{
		{"market", Market, false},
		{"Market", Market, false},
		{"MARKET", Market, false},
		{"limit", Limit, false},
		{"Limit", Limit, false},
		{"LIMIT", Limit, false},
		{"stop", Stop, false},
		{"Stop", Stop, false},
		{"STOP", Stop, false},
		{"invalid", Market, true},
		{"", Market, true},
	}
	
	for _, test := range tests {
		result, err := ParseOrderType(test.input)
		
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s'", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v for input '%s', got %v", test.expected, test.input, result)
			}
		}
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Value:   "test_value",
		Message: "test message",
	}
	
	errorMsg := err.Error()
	
	if !strings.Contains(errorMsg, "test_field") {
		t.Error("Expected error message to contain field name")
	}
	
	if !strings.Contains(errorMsg, "test message") {
		t.Error("Expected error message to contain message")
	}
}

func TestValidateStruct(t *testing.T) {
	// 正常なケース
	config := NewDefaultConfig()
	config.Market.DataProvider.FilePath = "./testdata/sample.csv"
	
	if err := ValidateStruct(&config); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}
	
	// 異常なケース
	config.Market.DataProvider.FilePath = ""
	if err := ValidateStruct(&config); err == nil {
		t.Error("Expected error for invalid config")
	}
}