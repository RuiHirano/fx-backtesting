package models

import (
	"errors"
	"fmt"
	"time"
)

// Candle はローソク足データを表します。
type Candle struct {
	Timestamp time.Time `json:"timestamp" csv:"timestamp"`
	Open      float64   `json:"open" csv:"open"`
	High      float64   `json:"high" csv:"high"`
	Low       float64   `json:"low" csv:"low"`
	Close     float64   `json:"close" csv:"close"`
	Volume    float64   `json:"volume" csv:"volume"`
}

// NewCandle は新しいローソク足データを作成します。
func NewCandle(timestamp time.Time, open, high, low, close, volume float64) *Candle {
	return &Candle{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}
}

// Validate はローソク足データの妥当性を検証します。
func (c *Candle) Validate() error {
	if c.High < c.Low {
		return errors.New("high price must be greater than or equal to low price")
	}
	
	if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
		return errors.New("prices must be positive")
	}
	
	if c.Volume < 0 {
		return errors.New("volume must be non-negative")
	}
	
	return nil
}

// IsValidOHLC は四本値の妥当性をチェックします。
func (c *Candle) IsValidOHLC() bool {
	return c.High >= c.Open && c.High >= c.Close &&
		c.Low <= c.Open && c.Low <= c.Close &&
		c.High >= c.Low
}

// ToCSVRecord はCSV形式の文字列スライスに変換します。
func (c *Candle) ToCSVRecord() []string {
	return []string{
		c.Timestamp.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%.5f", c.Open),
		fmt.Sprintf("%.5f", c.High),
		fmt.Sprintf("%.5f", c.Low),
		fmt.Sprintf("%.5f", c.Close),
		fmt.Sprintf("%.0f", c.Volume),
	}
}