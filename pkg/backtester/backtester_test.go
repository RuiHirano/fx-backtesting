package backtester

import (
	"context"
	"testing"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Backtester NewBacktester テスト
func TestBacktester_NewBacktester(t *testing.T) {
	// Market・Broker設定
	dataConfig := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	// Backtester作成
	backtester := NewBacktester(dataConfig, brokerConfig)
	
	if backtester == nil {
		t.Fatal("Expected backtester to be created")
	}
	
	// 初期化
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期状態確認
	if backtester.IsFinished() {
		t.Error("Expected backtester to not be finished initially")
	}
	
	// 初期残高確認
	balance := backtester.GetBalance()
	if balance != brokerConfig.InitialBalance {
		t.Errorf("Expected initial balance %f, got %f", brokerConfig.InitialBalance, balance)
	}
	
	// 初期ポジション確認
	positions := backtester.GetPositions()
	if len(positions) != 0 {
		t.Errorf("Expected 0 initial positions, got %d", len(positions))
	}
}

// Backtester Forward（時間進行）テスト
func TestBacktester_Forward(t *testing.T) {
	backtester := createTestBacktester(t)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期時刻取得
	initialTime := backtester.GetCurrentTime()
	if initialTime.IsZero() {
		t.Error("Expected non-zero initial time")
	}
	
	// 初期価格取得
	initialPrice := backtester.GetCurrentPrice("EURUSD")
	if initialPrice <= 0.0 {
		t.Error("Expected positive initial price")
	}
	
	// 時間進行
	hasNext := backtester.Forward()
	if !hasNext {
		t.Error("Expected to have next candle")
	}
	
	// 時間更新確認
	newTime := backtester.GetCurrentTime()
	if !newTime.After(initialTime) {
		t.Error("Expected time to advance after Forward")
	}
	
	// 価格更新確認
	newPrice := backtester.GetCurrentPrice("EURUSD")
	if newPrice <= 0.0 {
		t.Error("Expected positive price after Forward")
	}
	
	// 全データ消費まで進行
	stepCount := 1
	for backtester.Forward() {
		stepCount++
		if stepCount > 100 { // 無限ループ防止
			break
		}
	}
	
	// 終了状態確認
	if !backtester.IsFinished() {
		t.Error("Expected backtester to be finished after consuming all data")
	}
}

// Backtester Buy・Sell API テスト
func TestBacktester_BuySell(t *testing.T) {
	backtester := createTestBacktester(t)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 初期残高確認
	initialBalance := backtester.GetBalance()
	
	// 買い注文実行
	err = backtester.Buy("EURUSD", 10000.0)
	if err != nil {
		t.Fatalf("Expected no error from Buy, got %v", err)
	}
	
	// ポジション確認
	positions := backtester.GetPositions()
	if len(positions) != 1 {
		t.Errorf("Expected 1 position after buy, got %d", len(positions))
	}
	
	if len(positions) > 0 {
		position := positions[0]
		if position.Side != models.Buy {
			t.Errorf("Expected Buy position, got %v", position.Side)
		}
		if position.Size != 10000.0 {
			t.Errorf("Expected position size 10000.0, got %f", position.Size)
		}
	}
	
	// 残高変動確認
	newBalance := backtester.GetBalance()
	if newBalance >= initialBalance {
		t.Error("Expected balance to decrease after buy order")
	}
	
	// 売り注文実行
	err = backtester.Sell("EURUSD", 5000.0)
	if err != nil {
		t.Fatalf("Expected no error from Sell, got %v", err)
	}
	
	// ポジション確認（2つになる）
	positions = backtester.GetPositions()
	if len(positions) != 2 {
		t.Errorf("Expected 2 positions after sell, got %d", len(positions))
	}
}

// Backtester ポジション管理テスト
func TestBacktester_PositionManagement(t *testing.T) {
	backtester := createTestBacktester(t)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 複数ポジション作成
	err = backtester.Buy("EURUSD", 10000.0)
	if err != nil {
		t.Fatalf("Expected no error from first Buy, got %v", err)
	}
	
	err = backtester.Sell("EURUSD", 8000.0)
	if err != nil {
		t.Fatalf("Expected no error from Sell, got %v", err)
	}
	
	err = backtester.Buy("EURUSD", 12000.0)
	if err != nil {
		t.Fatalf("Expected no error from second Buy, got %v", err)
	}
	
	// ポジション一覧確認
	positions := backtester.GetPositions()
	if len(positions) != 3 {
		t.Errorf("Expected 3 positions, got %d", len(positions))
	}
	
	// 特定ポジションクローズ
	if len(positions) > 0 {
		positionID := positions[0].ID
		err = backtester.ClosePosition(positionID)
		if err != nil {
			t.Fatalf("Expected no error from ClosePosition, got %v", err)
		}
		
		// ポジション数確認
		updatedPositions := backtester.GetPositions()
		if len(updatedPositions) != 2 {
			t.Errorf("Expected 2 positions after close, got %d", len(updatedPositions))
		}
	}
	
	// 全ポジションクローズ
	err = backtester.CloseAllPositions()
	if err != nil {
		t.Fatalf("Expected no error from CloseAllPositions, got %v", err)
	}
	
	// ポジション数確認
	finalPositions := backtester.GetPositions()
	if len(finalPositions) != 0 {
		t.Errorf("Expected 0 positions after close all, got %d", len(finalPositions))
	}
}

// Backtester 統合テスト
func TestBacktester_Integration(t *testing.T) {
	backtester := createTestBacktester(t)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 複数の時間ステップで取引実行
	tradeCount := 0
	maxSteps := 3
	
	for step := 0; step < maxSteps && !backtester.IsFinished(); step++ {
		// 現在価格取得
		price := backtester.GetCurrentPrice("EURUSD")
		if price <= 0.0 {
			t.Errorf("Expected positive price at step %d", step)
		}
		
		// ステップごとに異なる取引パターン
		switch step {
		case 0:
			// 買い注文
			err = backtester.Buy("EURUSD", 10000.0)
			if err != nil {
				t.Fatalf("Expected no error from Buy at step %d, got %v", step, err)
			}
			tradeCount++
			
		case 1:
			// 売り注文
			err = backtester.Sell("EURUSD", 5000.0)
			if err != nil {
				t.Fatalf("Expected no error from Sell at step %d, got %v", step, err)
			}
			tradeCount++
			
		case 2:
			// ポジション一部決済
			positions := backtester.GetPositions()
			if len(positions) > 0 {
				err = backtester.ClosePosition(positions[0].ID)
				if err != nil {
					t.Fatalf("Expected no error from ClosePosition at step %d, got %v", step, err)
				}
			}
		}
		
		// 時間進行
		if step < maxSteps-1 {
			hasNext := backtester.Forward()
			if !hasNext {
				break // データ終了
			}
		}
	}
	
	// 最終状態確認
	finalPositions := backtester.GetPositions()
	finalBalance := backtester.GetBalance()
	
	// ログ出力（デバッグ用）
	t.Logf("Final positions: %d", len(finalPositions))
	t.Logf("Final balance: %f", finalBalance)
	t.Logf("Trade count: %d", tradeCount)
	
	// 基本的な整合性確認
	if finalBalance < 0 {
		t.Error("Expected non-negative final balance")
	}
}

// Backtester エラーハンドリングテスト
func TestBacktester_ErrorHandling(t *testing.T) {
	backtester := createTestBacktester(t)
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	err := backtester.Initialize(ctx)
	if err != nil {
		t.Fatalf("Expected no error from Initialize, got %v", err)
	}
	
	// 無効なサイズでの注文
	err = backtester.Buy("EURUSD", 0.0)
	if err == nil {
		t.Error("Expected error for zero size order")
	}
	
	err = backtester.Buy("EURUSD", -1000.0)
	if err == nil {
		t.Error("Expected error for negative size order")
	}
	
	// 存在しないシンボル
	err = backtester.Buy("INVALID", 1000.0)
	if err == nil {
		t.Error("Expected error for invalid symbol")
	}
	
	// 残高不足での大きな注文
	err = backtester.Buy("EURUSD", 10000000.0) // 非常に大きな注文
	if err == nil {
		t.Error("Expected error for insufficient balance")
	}
	
	// 存在しないポジションのクローズ
	err = backtester.ClosePosition("nonexistent-id")
	if err == nil {
		t.Error("Expected error for nonexistent position")
	}
	
	// 初期化前の操作エラー（新しいbacktesterで）
	uninitializedBacktester := createTestBacktester(t)
	
	err = uninitializedBacktester.Buy("EURUSD", 1000.0)
	if err == nil {
		t.Error("Expected error for uninitialized backtester")
	}
}

// ヘルパー関数: テスト用Backtester作成
func createTestBacktester(_ *testing.T) *Backtester {
	dataConfig := models.DataProviderConfig{
		FilePath: "./testdata/sample.csv",
		Format:   "csv",
	}
	
	brokerConfig := models.BrokerConfig{
		InitialBalance: 10000.0,
		Spread:         0.0001,
	}
	
	return NewBacktester(dataConfig, brokerConfig)
}