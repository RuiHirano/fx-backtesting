package models

import "time"

// Position は保有ポジションを表します。
type Position struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Side         OrderSide `json:"side"`
	Size         float64   `json:"size"`
	EntryPrice   float64   `json:"entry_price"`
	CurrentPrice float64   `json:"current_price"`
	PnL          float64   `json:"pnl"`
	OpenTime     time.Time `json:"open_time"`
	StopLoss     float64   `json:"stop_loss,omitempty"`
	TakeProfit   float64   `json:"take_profit,omitempty"`
}

// NewPosition は新しいポジションを作成します。
func NewPosition(id, symbol string, side OrderSide, size, entryPrice float64) *Position {
	return &Position{
		ID:           id,
		Symbol:       symbol,
		Side:         side,
		Size:         size,
		EntryPrice:   entryPrice,
		CurrentPrice: entryPrice,
		PnL:          0.0,
		OpenTime:     time.Now(),
	}
}

// UpdatePrice は現在価格を更新し、PnLを再計算します。
func (p *Position) UpdatePrice(currentPrice float64) {
	p.CurrentPrice = currentPrice
	p.calculatePnL()
}

// calculatePnL は損益を計算します。
func (p *Position) calculatePnL() {
	if p.Side == Buy {
		p.PnL = (p.CurrentPrice - p.EntryPrice) * p.Size
	} else {
		p.PnL = (p.EntryPrice - p.CurrentPrice) * p.Size
	}
}

// IsLong は買いポジションかどうかを判定します。
func (p *Position) IsLong() bool {
	return p.Side == Buy
}

// IsShort は売りポジションかどうかを判定します。
func (p *Position) IsShort() bool {
	return p.Side == Sell
}

// ShouldStopLoss はストップロス条件に達しているかを判定します。
func (p *Position) ShouldStopLoss() bool {
	if p.StopLoss <= 0 {
		return false
	}
	
	if p.IsLong() {
		return p.CurrentPrice <= p.StopLoss
	}
	return p.CurrentPrice >= p.StopLoss
}

// ShouldTakeProfit はテイクプロフィット条件に達しているかを判定します。
func (p *Position) ShouldTakeProfit() bool {
	if p.TakeProfit <= 0 {
		return false
	}
	
	if p.IsLong() {
		return p.CurrentPrice >= p.TakeProfit
	}
	return p.CurrentPrice <= p.TakeProfit
}

// GetMarketValue は現在の市場価値を返します。
func (p *Position) GetMarketValue() float64 {
	return p.CurrentPrice * p.Size
}

// GetPnLPercentage は損益率を返します。
func (p *Position) GetPnLPercentage() float64 {
	if p.EntryPrice == 0 {
		return 0
	}
	return (p.PnL / (p.EntryPrice * p.Size)) * 100
}