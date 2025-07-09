package models

import "testing"


// Trade構造体のテスト
func TestTrade_NewTradeFromPosition(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	exitPrice := 1.1010
	
	trade := NewTradeFromPosition(position, exitPrice)
	
	if trade.ID != position.ID {
		t.Errorf("Expected ID %s, got %s", position.ID, trade.ID)
	}
	
	if trade.Symbol != position.Symbol {
		t.Errorf("Expected symbol %s, got %s", position.Symbol, trade.Symbol)
	}
	
	if trade.Side != position.Side {
		t.Errorf("Expected side %v, got %v", position.Side, trade.Side)
	}
	
	if trade.Size != position.Size {
		t.Errorf("Expected size %f, got %f", position.Size, trade.Size)
	}
	
	if trade.EntryPrice != position.EntryPrice {
		t.Errorf("Expected entry price %f, got %f", position.EntryPrice, trade.EntryPrice)
	}
	
	if trade.ExitPrice != exitPrice {
		t.Errorf("Expected exit price %f, got %f", exitPrice, trade.ExitPrice)
	}
	
	if trade.Status != TradeClosed {
		t.Errorf("Expected status TradeClosed, got %v", trade.Status)
	}
	
	expectedPnL := (1.1010 - 1.1000) * 10000.0
	assertFloatEqual(t, expectedPnL, trade.PnL, "Trade PnL")
}

func TestTrade_IsWinning(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	
	// 勝ち取引
	trade := NewTradeFromPosition(position, 1.1010)
	if !trade.IsWinning() {
		t.Error("Expected winning trade")
	}
	
	// 負け取引
	trade = NewTradeFromPosition(position, 1.0990)
	if !trade.IsLosing() {
		t.Error("Expected losing trade")
	}
	
	// 引き分け
	trade = NewTradeFromPosition(position, 1.1000)
	if !trade.IsBreakeven() {
		t.Error("Expected breakeven trade")
	}
}

func TestTrade_IsLosing(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	
	// 負け取引
	trade := NewTradeFromPosition(position, 1.0990)
	if !trade.IsLosing() {
		t.Error("Expected losing trade")
	}
	
	// 勝ち取引
	trade = NewTradeFromPosition(position, 1.1010)
	if trade.IsLosing() {
		t.Error("Expected winning trade, not losing")
	}
}

func TestTrade_IsBreakeven(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	
	// 引き分け取引
	trade := NewTradeFromPosition(position, 1.1000)
	if !trade.IsBreakeven() {
		t.Error("Expected breakeven trade")
	}
	
	// 勝ち取引
	trade = NewTradeFromPosition(position, 1.1010)
	if trade.IsBreakeven() {
		t.Error("Expected winning trade, not breakeven")
	}
}

func TestTrade_GetPnLPercentage(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	trade := NewTradeFromPosition(position, 1.1010)
	
	expectedPercentage := ((1.1010 - 1.1000) / 1.1000) * 100
	actualPercentage := trade.GetPnLPercentage()
	
	assertFloatEqual(t, expectedPercentage, actualPercentage, "Trade PnL percentage")
}

func TestTrade_GetDurationHours(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	trade := NewTradeFromPosition(position, 1.1010)
	
	// 取引時間は非常に短いはずなので、0以上であることを確認
	if trade.GetDurationHours() < 0 {
		t.Error("Expected non-negative duration")
	}
}

func TestTrade_ToCSVRecord(t *testing.T) {
	position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
	trade := NewTradeFromPosition(position, 1.1010)
	
	record := trade.ToCSVRecord()
	
	expectedLength := 11
	if len(record) != expectedLength {
		t.Errorf("Expected CSV record length %d, got %d", expectedLength, len(record))
	}
	
	if record[0] != trade.ID {
		t.Errorf("Expected ID %s, got %s", trade.ID, record[0])
	}
	
	if record[1] != trade.Symbol {
		t.Errorf("Expected symbol %s, got %s", trade.Symbol, record[1])
	}
	
	if record[2] != trade.Side.String() {
		t.Errorf("Expected side %s, got %s", trade.Side.String(), record[2])
	}
}

func TestTradeStatus_String(t *testing.T) {
	tests := []struct {
		status   TradeStatus
		expected string
	}{
		{TradeOpen, "Open"},
		{TradeClosed, "Closed"},
		{TradeCanceled, "Canceled"},
		{TradeStatus(999), "Unknown"},
	}
	
	for _, test := range tests {
		if test.status.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.status.String())
		}
	}
}

func TestCalculateTradePnL(t *testing.T) {
	// 買い取引のテスト
	buyPnL := calculateTradePnL(Buy, 10000.0, 1.1000, 1.1010)
	expectedBuyPnL := (1.1010 - 1.1000) * 10000.0
	assertFloatEqual(t, expectedBuyPnL, buyPnL, "Buy trade PnL")
	
	// 売り取引のテスト
	sellPnL := calculateTradePnL(Sell, 10000.0, 1.1000, 1.0990)
	expectedSellPnL := (1.1000 - 1.0990) * 10000.0
	assertFloatEqual(t, expectedSellPnL, sellPnL, "Sell trade PnL")
}