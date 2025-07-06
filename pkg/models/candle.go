package models

import (
	"math"
	"time"
)

// Candle represents a single candlestick data point
type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}

// NewCandle creates a new Candle instance
func NewCandle(timestamp time.Time, open, high, low, close float64, volume int64) Candle {
	return Candle{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}
}

// IsValid checks if the candle data is valid
func (c Candle) IsValid() bool {
	// Check if high is the highest price
	if c.High < c.Open || c.High < c.Close || c.High < c.Low {
		return false
	}
	
	// Check if low is the lowest price
	if c.Low > c.Open || c.Low > c.Close || c.Low > c.High {
		return false
	}
	
	// Check if volume is non-negative
	if c.Volume < 0 {
		return false
	}
	
	return true
}

// IsBullish returns true if the candle is bullish (close > open)
func (c Candle) IsBullish() bool {
	return c.Close > c.Open
}

// IsBearish returns true if the candle is bearish (close < open)
func (c Candle) IsBearish() bool {
	return c.Close < c.Open
}

// BodySize returns the absolute size of the candle body
func (c Candle) BodySize() float64 {
	return math.Abs(c.Close - c.Open)
}