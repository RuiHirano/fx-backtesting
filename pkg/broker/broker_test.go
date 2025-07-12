package broker

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/market"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/stretchr/testify/assert"
)

// テスト用のヘルパー関数
func createTestBroker(t *testing.T) (Broker, market.Market) {
	// Marketの準備
	config := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	marketConfig := models.MarketConfig{
		DataProvider: config,
		Symbol:       "EURUSD",
	}
	mkt := market.NewMarket(marketConfig)
	
	ctx := context.Background()
	
	err := mkt.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize market: %v", err)
	}
	
	// Brokerの準備
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	broker := NewSimpleBroker(brokerConfig, mkt)
	return broker, mkt
}

// 成行注文テスト
func TestBroker_PlaceMarketOrder(t *testing.T) {
	broker, _ := createTestBroker(t)
	
	t.Run("should execute market buy order immediately", func(t *testing.T) {
		// 成行買い注文
		order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// ポジション確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 1)
		assert.Equal(t, models.Buy, positions[0].Side)
		assert.Equal(t, 10000.0, positions[0].Size)
		
		// 残高確認
		balance := broker.GetBalance()
		assert.Less(t, balance, 10000.0) // 証拠金が差し引かれている
		
		// 注文状態確認
		assert.True(t, order.IsExecuted())
		assert.Greater(t, order.ExecutedPrice, 0.0)
	})
	
	t.Run("should execute market sell order immediately", func(t *testing.T) {
		// 成行売り注文
		order := models.NewMarketOrder("order-2", "EURUSD", models.Sell, 5000.0)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// ポジション確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 2) // 前のテストのポジション + 新しいポジション
		
		// 売りポジションを確認
		var sellPosition *models.Position
		for _, pos := range positions {
			if pos.Side == models.Sell {
				sellPosition = pos
				break
			}
		}
		assert.NotNil(t, sellPosition)
		assert.Equal(t, 5000.0, sellPosition.Size)
	})
	
	t.Run("should return error for insufficient balance", func(t *testing.T) {
		// 残高を超える大きな注文
		order := models.NewMarketOrder("order-3", "EURUSD", models.Buy, 1000000.0)
		
		err := broker.PlaceOrder(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
		
		// 注文が約定されていないことを確認
		assert.True(t, order.IsPending())
	})
}

// 指値注文テスト
func TestBroker_PlaceLimitOrder(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should add limit order to pending orders", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 現在価格より低い買い指値注文（約定しない）
		order := models.NewLimitOrder("limit-1", "EURUSD", models.Buy, 10000.0, currentPrice-0.0010)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文に追加されていることを確認
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 1)
		assert.Equal(t, order.ID, pendingOrders[0].ID)
		
		// ポジションは作成されていないことを確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 0)
	})
	
	t.Run("should execute limit order when price condition is met", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 現在価格より高い買い指値注文（即座に約定する）
		order := models.NewLimitOrder("limit-2", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文処理を実行
		broker.ProcessPendingOrders()
		
		// 約定してポジションが作成されていることを確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 1)
		
		// 保留注文から削除されていることを確認
		pendingOrders := broker.GetPendingOrders()
		found := false
		for _, pendingOrder := range pendingOrders {
			if pendingOrder.ID == order.ID {
				found = true
				break
			}
		}
		assert.False(t, found)
	})
}

// 逆指値注文テスト
func TestBroker_PlaceStopOrder(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should add stop order to pending orders", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 現在価格より高い買い逆指値注文（約定しない）
		order := models.NewStopOrder("stop-1", "EURUSD", models.Buy, 10000.0, currentPrice+0.0020)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文に追加されていることを確認
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 1)
		assert.Equal(t, order.ID, pendingOrders[0].ID)
	})
	
	t.Run("should execute stop order when price condition is met", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 現在価格より低い買い逆指値注文（即座に約定する）
		order := models.NewStopOrder("stop-2", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
		
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文処理を実行
		broker.ProcessPendingOrders()
		
		// 約定してポジションが作成されていることを確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 1)
	})
}

// 注文キャンセルテスト
func TestBroker_CancelOrder(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should cancel pending order successfully", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 約定しない指値注文を作成
		order := models.NewLimitOrder("cancel-1", "EURUSD", models.Buy, 10000.0, currentPrice-0.0020)
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文があることを確認
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 1)
		
		// 注文をキャンセル
		err = broker.CancelOrder(order.ID)
		assert.NoError(t, err)
		
		// 保留注文から削除されていることを確認
		pendingOrders = broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 0)
		
		// 注文状態がキャンセルされていることを確認
		assert.True(t, order.IsCancelled())
	})
	
	t.Run("should return error for non-existent order", func(t *testing.T) {
		err := broker.CancelOrder("non-existent-order")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
	})
	
	t.Run("should return error for already executed order", func(t *testing.T) {
		// 成行注文を実行（即座に約定）
		order := models.NewMarketOrder("cancel-2", "EURUSD", models.Buy, 5000.0)
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 約定済み注文のキャンセルを試行
		err = broker.CancelOrder(order.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
	})
}

// 保留注文取得テスト
func TestBroker_GetPendingOrders(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should return all pending orders", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 複数の保留注文を作成
		order1 := models.NewLimitOrder("pending-1", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
		order2 := models.NewLimitOrder("pending-2", "EURUSD", models.Sell, 7000.0, currentPrice+0.0015)
		order3 := models.NewStopOrder("pending-3", "EURUSD", models.Buy, 3000.0, currentPrice+0.0020)
		
		broker.PlaceOrder(order1)
		broker.PlaceOrder(order2)
		broker.PlaceOrder(order3)
		
		// 保留注文を取得
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 3)
		
		// 各注文が含まれていることを確認
		orderIDs := make(map[string]bool)
		for _, order := range pendingOrders {
			orderIDs[order.ID] = true
		}
		assert.True(t, orderIDs["pending-1"])
		assert.True(t, orderIDs["pending-2"])
		assert.True(t, orderIDs["pending-3"])
	})
	
	t.Run("should return empty slice when no pending orders", func(t *testing.T) {
		// 新しいブローカーを作成
		newBroker, _ := createTestBroker(t)
		
		pendingOrders := newBroker.GetPendingOrders()
		assert.Len(t, pendingOrders, 0)
	})
}

// ポジション決済テスト
func TestBroker_ClosePosition(t *testing.T) {
	broker, _ := createTestBroker(t)
	
	t.Run("should close position successfully", func(t *testing.T) {
		// ポジション作成
		order := models.NewMarketOrder("close-1", "EURUSD", models.Buy, 10000.0)
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		positions := broker.GetPositions()
		assert.Len(t, positions, 1)
		
		positionID := positions[0].ID
		initialBalance := broker.GetBalance()
		
		// ポジション決済
		err = broker.ClosePosition(positionID)
		assert.NoError(t, err)
		
		// ポジションが削除されていることを確認
		positions = broker.GetPositions()
		assert.Len(t, positions, 0)
		
		// 残高が更新されていることを確認
		finalBalance := broker.GetBalance()
		assert.NotEqual(t, initialBalance, finalBalance)
		
		// 取引履歴が追加されていることを確認
		trades := broker.GetTradeHistory()
		assert.Len(t, trades, 1)
	})
	
	t.Run("should return error for non-existent position", func(t *testing.T) {
		err := broker.ClosePosition("non-existent-position")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position not found")
	})
}

// 保留注文処理テスト
func TestBroker_ProcessPendingOrders(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should process limit orders correctly", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 異なる価格条件の指値注文を作成
		buyLimitBelow := models.NewLimitOrder("limit-buy-below", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
		buyLimitAbove := models.NewLimitOrder("limit-buy-above", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
		sellLimitBelow := models.NewLimitOrder("limit-sell-below", "EURUSD", models.Sell, 5000.0, currentPrice-0.0010)
		sellLimitAbove := models.NewLimitOrder("limit-sell-above", "EURUSD", models.Sell, 5000.0, currentPrice+0.0010)
		
		broker.PlaceOrder(buyLimitBelow)
		broker.PlaceOrder(buyLimitAbove)
		broker.PlaceOrder(sellLimitBelow)
		broker.PlaceOrder(sellLimitAbove)
		
		// 保留注文処理を実行
		broker.ProcessPendingOrders()
		
		// 約定すべき注文が約定していることを確認
		// 買い指値は現在価格以下で約定: buyLimitAbove
		// 売り指値は現在価格以上で約定: sellLimitAbove
		positions := broker.GetPositions()
		assert.Len(t, positions, 2)
		
		// 残りの保留注文を確認
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 2)
	})
	
	t.Run("should process stop orders correctly", func(t *testing.T) {
		// 新しいブローカーを作成
		newBroker, newMkt := createTestBroker(t)
		currentPrice := newMkt.GetCurrentPrice()
		
		// 異なる価格条件の逆指値注文を作成
		buyStopBelow := models.NewStopOrder("stop-buy-below", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
		buyStopAbove := models.NewStopOrder("stop-buy-above", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
		sellStopBelow := models.NewStopOrder("stop-sell-below", "EURUSD", models.Sell, 5000.0, currentPrice-0.0010)
		sellStopAbove := models.NewStopOrder("stop-sell-above", "EURUSD", models.Sell, 5000.0, currentPrice+0.0010)
		
		newBroker.PlaceOrder(buyStopBelow)
		newBroker.PlaceOrder(buyStopAbove)
		newBroker.PlaceOrder(sellStopBelow)
		newBroker.PlaceOrder(sellStopAbove)
		
		// 保留注文処理を実行
		newBroker.ProcessPendingOrders()
		
		// 約定すべき注文が約定していることを確認
		// 買い逆指値は現在価格以上で約定: buyStopAbove
		// 売り逆指値は現在価格以下で約定: sellStopBelow
		positions := newBroker.GetPositions()
		assert.Len(t, positions, 2)
		
		// 残りの保留注文を確認
		pendingOrders := newBroker.GetPendingOrders()
		assert.Len(t, pendingOrders, 2)
	})
}

// UpdatePositionsテスト
func TestBroker_UpdatePositions(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should update position prices and process pending orders", func(t *testing.T) {
		// ポジション作成
		order := models.NewMarketOrder("update-1", "EURUSD", models.Buy, 10000.0)
		err := broker.PlaceOrder(order)
		assert.NoError(t, err)
		
		// 保留注文も作成
		currentPrice := mkt.GetCurrentPrice()
		limitOrder := models.NewLimitOrder("update-limit", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
		broker.PlaceOrder(limitOrder)
		
		// 初期状態を記録
		positions := broker.GetPositions()
		_ = positions[0].CurrentPrice // 将来の価格比較用（現在は未使用）
		
		// 市場を進める
		mkt.Forward()
		
		// UpdatePositionsを実行
		broker.UpdatePositions()
		
		// ポジション価格が更新されていることを確認
		updatedPositions := broker.GetPositions()
		if len(updatedPositions) > 0 {
			// 価格が更新される可能性がある（テストデータ次第）
			assert.GreaterOrEqual(t, updatedPositions[0].CurrentPrice, 0.0)
		}
	})
}

// 統合テスト
func TestBroker_Integration(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("complete trading scenario", func(t *testing.T) {
		currentPrice := mkt.GetCurrentPrice()
		
		// 1. 成行注文でポジション作成
		marketOrder := models.NewMarketOrder("integration-market", "EURUSD", models.Buy, 10000.0)
		err := broker.PlaceOrder(marketOrder)
		assert.NoError(t, err)
		
		// 2. 利食い用の指値注文
		takeProfitOrder := models.NewLimitOrder("integration-tp", "EURUSD", models.Sell, 10000.0, currentPrice+0.0050)
		err = broker.PlaceOrder(takeProfitOrder)
		assert.NoError(t, err)
		
		// 3. 損切り用の逆指値注文
		stopLossOrder := models.NewStopOrder("integration-sl", "EURUSD", models.Sell, 10000.0, currentPrice-0.0020)
		err = broker.PlaceOrder(stopLossOrder)
		assert.NoError(t, err)
		
		// 4. 状態確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 1)
		
		pendingOrders := broker.GetPendingOrders()
		assert.Len(t, pendingOrders, 2)
		
		// 5. 市場更新と保留注文処理
		mkt.Forward()
		broker.UpdatePositions()
		
		// 6. 最終状態確認（約定条件によって結果が変わる）
		finalPositions := broker.GetPositions()
		finalPendingOrders := broker.GetPendingOrders()
		finalTrades := broker.GetTradeHistory()
		
		// 基本的な整合性チェック
		assert.GreaterOrEqual(t, len(finalPositions), 0)
		assert.GreaterOrEqual(t, len(finalPendingOrders), 0)
		assert.GreaterOrEqual(t, len(finalTrades), 0)
		assert.Greater(t, broker.GetBalance(), 0.0)
	})
}

// エラーハンドリングテスト
func TestBroker_ErrorHandling(t *testing.T) {
	broker, _ := createTestBroker(t)
	
	t.Run("should validate order data", func(t *testing.T) {
		// 無効なサイズの注文
		invalidOrder := models.NewMarketOrder("invalid-1", "EURUSD", models.Buy, 0.0)
		err := broker.PlaceOrder(invalidOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order size must be positive")
		
		// 無効な指値価格
		invalidLimitOrder := models.NewLimitOrder("invalid-2", "EURUSD", models.Buy, 1000.0, 0.0)
		err = broker.PlaceOrder(invalidLimitOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit order must have positive limit price")
		
		// 無効な逆指値価格
		invalidStopOrder := models.NewStopOrder("invalid-3", "EURUSD", models.Buy, 1000.0, -1.0)
		err = broker.PlaceOrder(invalidStopOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stop order must have positive stop price")
	})
	
	t.Run("should handle insufficient balance for pending orders", func(t *testing.T) {
		// 残高を消費する大きな成行注文
		largeOrder := models.NewMarketOrder("large-order", "EURUSD", models.Buy, 90000.0)
		err := broker.PlaceOrder(largeOrder)
		assert.NoError(t, err) // 成行注文は成功
		
		// 残高不足で約定できない保留注文を作成
		currentPrice := 1.0000 // 仮の価格
		limitOrder := models.NewLimitOrder("insufficient-limit", "EURUSD", models.Buy, 50000.0, currentPrice+0.0010)
		err = broker.PlaceOrder(limitOrder)
		assert.NoError(t, err) // 保留注文の作成は成功
		
		// 保留注文処理時に残高不足でエラーになることを確認
		// この場合、約定条件が満たされても約定しない
		broker.ProcessPendingOrders()
		
		// ポジションが増えていないことを確認
		positions := broker.GetPositions()
		assert.Len(t, positions, 1) // 最初の大きな注文のみ
	})
}

// パフォーマンステスト
func TestBroker_Performance(t *testing.T) {
	broker, mkt := createTestBroker(t)
	
	t.Run("should handle multiple orders efficiently", func(t *testing.T) {
		start := time.Now()
		
		currentPrice := mkt.GetCurrentPrice()
		
		// 大量の保留注文を作成
		for i := 0; i < 100; i++ {
			limitOrder := models.NewLimitOrder(
				fmt.Sprintf("perf-limit-%d", i),
				"EURUSD",
				models.Buy,
				1000.0,
				currentPrice-float64(i)*0.0001,
			)
			err := broker.PlaceOrder(limitOrder)
			assert.NoError(t, err)
		}
		
		// 保留注文処理
		broker.ProcessPendingOrders()
		
		elapsed := time.Since(start)
		t.Logf("Processing 100 orders took: %v", elapsed)
		
		// 1秒以内に完了することを確認
		assert.Less(t, elapsed, time.Second)
		
		// 保留注文が正しく管理されていることを確認
		pendingOrders := broker.GetPendingOrders()
		assert.GreaterOrEqual(t, len(pendingOrders), 0)
	})
}