package models

import (
	"testing"
	"time"
)

// Candle構造体のテスト
func TestCandle_NewCandle(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candle := NewCandle(timestamp, 1.1000, 1.1010, 1.0990, 1.1005, 1000.0)
	
	if candle.Timestamp != timestamp {
		t.Errorf("Expected timestamp %v, got %v", timestamp, candle.Timestamp)
	}
	
	if candle.Open != 1.1000 {
		t.Errorf("Expected open 1.1000, got %f", candle.Open)
	}
	
	if candle.High != 1.1010 {
		t.Errorf("Expected high 1.1010, got %f", candle.High)
	}
	
	if candle.Low != 1.0990 {
		t.Errorf("Expected low 1.0990, got %f", candle.Low)
	}
	
	if candle.Close != 1.1005 {
		t.Errorf("Expected close 1.1005, got %f", candle.Close)
	}
	
	if candle.Volume != 1000.0 {
		t.Errorf("Expected volume 1000.0, got %f", candle.Volume)
	}
}

func TestCandle_Validate(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 正常なケース
	candle := NewCandle(timestamp, 1.1000, 1.1010, 1.0990, 1.1005, 1000.0)
	if err := candle.Validate(); err != nil {
		t.Errorf("Expected no error for valid candle, got %v", err)
	}
	
	// 異常なケース - High < Low
	candle = NewCandle(timestamp, 1.1000, 1.0990, 1.1010, 1.1005, 1000.0)
	if err := candle.Validate(); err == nil {
		t.Error("Expected error when high < low")
	}
	
	// 異常なケース - 負の価格
	candle = NewCandle(timestamp, -1.1000, 1.1010, 1.0990, 1.1005, 1000.0)
	if err := candle.Validate(); err == nil {
		t.Error("Expected error for negative price")
	}
	
	// 異常なケース - 負のボリューム
	candle = NewCandle(timestamp, 1.1000, 1.1010, 1.0990, 1.1005, -1000.0)
	if err := candle.Validate(); err == nil {
		t.Error("Expected error for negative volume")
	}
}

func TestCandle_IsValidOHLC(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 正常なケース
	candle := NewCandle(timestamp, 1.1000, 1.1010, 1.0990, 1.1005, 1000.0)
	if !candle.IsValidOHLC() {
		t.Error("Expected valid OHLC")
	}
	
	// 異常なケース - High < Open
	candle = NewCandle(timestamp, 1.1010, 1.1000, 1.0990, 1.1005, 1000.0)
	if candle.IsValidOHLC() {
		t.Error("Expected invalid OHLC when high < open")
	}
	
	// 異常なケース - Low > Close
	candle = NewCandle(timestamp, 1.1000, 1.1010, 1.1010, 1.1005, 1000.0)
	if candle.IsValidOHLC() {
		t.Error("Expected invalid OHLC when low > close")
	}
}

func TestCandle_ToCSVRecord(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC)
	candle := NewCandle(timestamp, 1.10000, 1.10100, 1.09900, 1.10050, 1000.0)
	
	record := candle.ToCSVRecord()
	
	expectedLength := 6
	if len(record) != expectedLength {
		t.Errorf("Expected CSV record length %d, got %d", expectedLength, len(record))
	}
	
	expectedTimestamp := "2024-01-01 12:30:00"
	if record[0] != expectedTimestamp {
		t.Errorf("Expected timestamp %s, got %s", expectedTimestamp, record[0])
	}
	
	expectedOpen := "1.10000"
	if record[1] != expectedOpen {
		t.Errorf("Expected open %s, got %s", expectedOpen, record[1])
	}
	
	expectedVolume := "1000"
	if record[5] != expectedVolume {
		t.Errorf("Expected volume %s, got %s", expectedVolume, record[5])
	}
}