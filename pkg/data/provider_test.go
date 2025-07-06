package data

import (
	"os"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestCSVDataProvider_LoadCSVData(t *testing.T) {
	// Create a temporary CSV file for testing
	csvContent := `2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200
2024-01-01 09:02:00,1.0520,1.0540,1.0510,1.0530,800`

	tmpFile, err := os.CreateTemp("", "test_data_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	provider := NewCSVDataProvider()
	candles, err := provider.LoadCSVData(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadCSVData failed: %v", err)
	}

	if len(candles) != 3 {
		t.Errorf("Expected 3 candles, got %d", len(candles))
	}

	// Check first candle
	expected := models.Candle{
		Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		Open:      1.0500,
		High:      1.0520,
		Low:       1.0490,
		Close:     1.0510,
		Volume:    1000,
	}

	if !candlesEqual(candles[0], expected) {
		t.Errorf("First candle mismatch: got %+v, want %+v", candles[0], expected)
	}
}

func TestCSVDataProvider_LoadCSVData_InvalidFile(t *testing.T) {
	provider := NewCSVDataProvider()
	_, err := provider.LoadCSVData("nonexistent.csv")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCSVDataProvider_LoadCSVData_InvalidFormat(t *testing.T) {
	// Create a temporary CSV file with invalid format
	csvContent := `invalid,data,format`

	tmpFile, err := os.CreateTemp("", "test_invalid_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	provider := NewCSVDataProvider()
	_, err = provider.LoadCSVData(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid CSV format")
	}
}

func TestCSVDataProvider_GetCandle(t *testing.T) {
	// Create test data
	csvContent := `2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000
2024-01-01 09:01:00,1.0510,1.0530,1.0500,1.0520,1200`

	tmpFile, err := os.CreateTemp("", "test_get_candle_*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(csvContent); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	provider := NewCSVDataProvider()
	_, err = provider.LoadCSVData(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadCSVData failed: %v", err)
	}

	// Test existing timestamp
	timestamp := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	candle, err := provider.GetCandle(timestamp)
	if err != nil {
		t.Fatalf("GetCandle failed: %v", err)
	}

	if candle.Timestamp != timestamp {
		t.Errorf("Expected timestamp %v, got %v", timestamp, candle.Timestamp)
	}

	// Test non-existent timestamp
	nonExistentTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	_, err = provider.GetCandle(nonExistentTime)
	if err == nil {
		t.Error("Expected error for non-existent timestamp")
	}
}

func TestParseCSVLine(t *testing.T) {
	line := "2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,1000"
	
	candle, err := parseCSVLine(line)
	if err != nil {
		t.Fatalf("parseCSVLine failed: %v", err)
	}

	expected := models.Candle{
		Timestamp: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
		Open:      1.0500,
		High:      1.0520,
		Low:       1.0490,
		Close:     1.0510,
		Volume:    1000,
	}

	if !candlesEqual(candle, expected) {
		t.Errorf("Parsed candle mismatch: got %+v, want %+v", candle, expected)
	}
}

func TestParseCSVLine_InvalidFormat(t *testing.T) {
	tests := []string{
		"invalid",
		"2024-01-01 09:00:00,invalid,1.0520,1.0490,1.0510,1000",
		"2024-01-01 09:00:00,1.0500,1.0520,1.0490,1.0510,invalid",
		"2024-01-01 09:00:00,1.0500,1.0520,1.0490", // missing fields
	}

	for _, test := range tests {
		_, err := parseCSVLine(test)
		if err == nil {
			t.Errorf("Expected error for invalid line: %s", test)
		}
	}
}

// Helper function to compare candles
func candlesEqual(a, b models.Candle) bool {
	return a.Timestamp.Equal(b.Timestamp) &&
		a.Open == b.Open &&
		a.High == b.High &&
		a.Low == b.Low &&
		a.Close == b.Close &&
		a.Volume == b.Volume
}