package models

import (
	"fmt"
	"time"
)

// TradeStatus は取引ステータスを表します。
type TradeStatus int

const (
	TradeOpen TradeStatus = iota // オープン中
	TradeClosed                  // クローズ済み
	TradeCanceled                // キャンセル済み
)

// String はTradeStatusの文字列表現を返します。
func (ts TradeStatus) String() string {
	switch ts {
	case TradeOpen:
		return "Open"
	case TradeClosed:
		return "Closed"
	case TradeCanceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}

// Trade は完了した取引を表します。
type Trade struct {
	ID         string        `json:"id"`
	Side       OrderSide     `json:"side"`
	Size       float64       `json:"size"`
	EntryPrice float64       `json:"entry_price"`
	ExitPrice  float64       `json:"exit_price"`
	PnL        float64       `json:"pnl"`
	Status     TradeStatus   `json:"status"`
	OpenTime   time.Time     `json:"open_time"`
	CloseTime  time.Time     `json:"close_time"`
	Duration   time.Duration `json:"duration"`
}

// NewTradeFromPosition はポジションから取引履歴を作成します。
func NewTradeFromPosition(position *Position, exitPrice float64, pnl float64, closeTime time.Time) *Trade {
	return &Trade{
		ID:         position.ID,
		Side:       position.Side,
		Size:       position.Size,
		EntryPrice: position.EntryPrice,
		ExitPrice:  exitPrice,
		PnL:        pnl,
		Status:     TradeClosed,
		OpenTime:   position.OpenTime,
		CloseTime:  closeTime,
		Duration:   closeTime.Sub(position.OpenTime),
	}
}

// calculateTradePnL は取引の損益を計算します。
func calculateTradePnL(side OrderSide, size, entryPrice, exitPrice float64) float64 {
	if side == Buy {
		return (exitPrice - entryPrice) * size
	}
	return (entryPrice - exitPrice) * size
}

// IsWinning は勝ち取引かどうかを判定します。
func (t *Trade) IsWinning() bool {
	return t.PnL > 0
}

// IsLosing は負け取引かどうかを判定します。
func (t *Trade) IsLosing() bool {
	return t.PnL < 0
}

// IsBreakeven は損益なしかどうかを判定します。
func (t *Trade) IsBreakeven() bool {
	return t.PnL == 0
}

// GetPnLPercentage は損益率を返します。
func (t *Trade) GetPnLPercentage() float64 {
	if t.EntryPrice == 0 {
		return 0
	}
	return (t.PnL / (t.EntryPrice * t.Size)) * 100
}

// GetDurationHours は取引時間を時間単位で返します。
func (t *Trade) GetDurationHours() float64 {
	return t.Duration.Hours()
}

// ToCSVRecord はCSV形式の文字列スライスに変換します。
func (t *Trade) ToCSVRecord() []string {
	return []string{
		t.ID,
		t.Side.String(),
		fmt.Sprintf("%.2f", t.Size),
		fmt.Sprintf("%.5f", t.EntryPrice),
		fmt.Sprintf("%.5f", t.ExitPrice),
		fmt.Sprintf("%.2f", t.PnL),
		t.Status.String(),
		t.OpenTime.Format("2006-01-02 15:04:05"),
		t.CloseTime.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%.2f", t.GetDurationHours()),
	}
}