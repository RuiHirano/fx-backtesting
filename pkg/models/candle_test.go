package models

import (
	"math"
	"testing"
	"time"
)

func TestNewCandle(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	
	candle := NewCandle(timestamp, 1.0500, 1.0520, 1.0490, 1.0510, 1000)
	
	if candle.Timestamp != timestamp {
		t.Errorf("Expected timestamp %v, got %v", timestamp, candle.Timestamp)
	}
	if candle.Open != 1.0500 {
		t.Errorf("Expected open 1.0500, got %v", candle.Open)
	}
	if candle.High != 1.0520 {
		t.Errorf("Expected high 1.0520, got %v", candle.High)
	}
	if candle.Low != 1.0490 {
		t.Errorf("Expected low 1.0490, got %v", candle.Low)
	}
	if candle.Close != 1.0510 {
		t.Errorf("Expected close 1.0510, got %v", candle.Close)
	}
	if candle.Volume != 1000 {
		t.Errorf("Expected volume 1000, got %v", candle.Volume)
	}
}

func TestCandle_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		candle   Candle
		expected bool
	}{
		{
			name: "valid candle",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      1.0500,
				High:      1.0520,
				Low:       1.0490,
				Close:     1.0510,
				Volume:    1000,
			},
			expected: true,
		},
		{
			name: "invalid - high less than open",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      1.0500,
				High:      1.0480,
				Low:       1.0490,
				Close:     1.0510,
				Volume:    1000,
			},
			expected: false,
		},
		{
			name: "invalid - low greater than close",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      1.0500,
				High:      1.0520,
				Low:       1.0515,
				Close:     1.0510,
				Volume:    1000,
			},
			expected: false,
		},
		{
			name: "invalid - negative volume",
			candle: Candle{
				Timestamp: time.Now(),
				Open:      1.0500,
				High:      1.0520,
				Low:       1.0490,
				Close:     1.0510,
				Volume:    -100,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.IsValid(); got != tt.expected {
				t.Errorf("Candle.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCandle_IsBullish(t *testing.T) {
	bullishCandle := Candle{
		Open:  1.0500,
		Close: 1.0510,
	}
	
	bearishCandle := Candle{
		Open:  1.0510,
		Close: 1.0500,
	}
	
	if !bullishCandle.IsBullish() {
		t.Error("Expected bullish candle to return true")
	}
	
	if bearishCandle.IsBullish() {
		t.Error("Expected bearish candle to return false")
	}
}

func TestCandle_IsBearish(t *testing.T) {
	bullishCandle := Candle{
		Open:  1.0500,
		Close: 1.0510,
	}
	
	bearishCandle := Candle{
		Open:  1.0510,
		Close: 1.0500,
	}
	
	if bullishCandle.IsBearish() {
		t.Error("Expected bullish candle to return false")
	}
	
	if !bearishCandle.IsBearish() {
		t.Error("Expected bearish candle to return true")
	}
}

func TestCandle_BodySize(t *testing.T) {
	candle := Candle{
		Open:  1.0500,
		Close: 1.0510,
	}
	
	expected := 0.0010
	got := candle.BodySize()
	if math.Abs(got-expected) > 1e-10 {
		t.Errorf("Expected body size %v, got %v", expected, got)
	}
}