package backtester

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/market"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/visualizer"
)

// VisualizerNotifier はVisualizerへの通知インターフェース
type VisualizerNotifier interface {
	OnCandleUpdate(candle *models.Candle) error
	OnTradeEvent(trade *models.Trade) error
	OnStatisticsUpdate(stats *models.Statistics) error
	OnBacktestStateChange(state BacktestState) error
}

// BacktestState はバックテストの状態を表す
type BacktestState int

const (
	BacktestStateIdle BacktestState = iota
	BacktestStateRunning
	BacktestStatePaused
	BacktestStateStopped
	BacktestStateCompleted
	BacktestStateError
)

// Backtester はバックテスト実行とユーザーAPIを提供する統括コンポーネントです。
type Backtester struct {
	market           market.Market
	broker           broker.Broker
	visualizer       VisualizerNotifier
	visualizerImpl   visualizer.Visualizer
	visualizerConfig models.VisualizerConfig
	initialized      bool
	statistics       *models.Statistics
}

// NewBacktester は新しいBacktesterを作成します。
func NewBacktester(dataConfig models.DataProviderConfig, brokerConfig models.BrokerConfig) *Backtester {
	return NewBacktesterWithVisualizer(dataConfig, brokerConfig, models.DisabledVisualizerConfig())
}

// NewBacktesterWithVisualizer はVisualizerConfigを含む新しいBacktesterを作成します。
func NewBacktesterWithVisualizer(dataConfig models.DataProviderConfig, brokerConfig models.BrokerConfig, visualizerConfig models.VisualizerConfig) *Backtester {
	// DataProvider作成
	provider := data.NewCSVProvider(dataConfig)
	
	// Market作成
	mkt := market.NewMarket(provider)
	
	// Broker作成  
	bkr := broker.NewSimpleBroker(brokerConfig, mkt)
	
	return &Backtester{
		market:           mkt,
		broker:           bkr,
		visualizer:       nil,
		visualizerImpl:   nil,
		visualizerConfig: visualizerConfig,
		initialized:      false,
		statistics:       models.NewStatistics(brokerConfig.InitialBalance),
	}
}

// VisualizerAdapter はVisualizerとBacktesterの型を適合させるアダプター
type VisualizerAdapter struct {
	visualizer visualizer.Visualizer
}

// NewVisualizerAdapter は新しいVisualizerAdapterを作成
func NewVisualizerAdapter(viz visualizer.Visualizer) *VisualizerAdapter {
	return &VisualizerAdapter{visualizer: viz}
}

// OnCandleUpdate はローソク足データ更新を転送
func (va *VisualizerAdapter) OnCandleUpdate(candle *models.Candle) error {
	return va.visualizer.OnCandleUpdate(candle)
}

// OnTradeEvent は取引イベントを転送
func (va *VisualizerAdapter) OnTradeEvent(trade *models.Trade) error {
	return va.visualizer.OnTradeEvent(trade)
}

// OnStatisticsUpdate は統計情報更新を転送
func (va *VisualizerAdapter) OnStatisticsUpdate(stats *models.Statistics) error {
	return va.visualizer.OnStatisticsUpdate(stats)
}

// OnBacktestStateChange はバックテスト状態変更を転送（型変換）
func (va *VisualizerAdapter) OnBacktestStateChange(state BacktestState) error {
	// backtester.BacktestState を visualizer.BacktestState に変換
	vizState := visualizer.BacktestState(state)
	return va.visualizer.OnBacktestStateChange(vizState)
}

// SetVisualizer はVisualizerを設定します（非推奨：NewBacktesterWithVisualizerを使用してください）。
func (bt *Backtester) SetVisualizer(visualizer VisualizerNotifier) error {
	bt.visualizer = visualizer
	return nil
}

// Initialize はBacktesterを初期化します。
func (bt *Backtester) Initialize(ctx context.Context) error {
	// VisualizerConfig検証
	if err := bt.visualizerConfig.Validate(); err != nil {
		return fmt.Errorf("invalid visualizer config: %w", err)
	}
	
	// Visualizer初期化（有効な場合のみ）
	if bt.visualizerConfig.Enabled {
		if err := bt.initializeVisualizer(ctx); err != nil {
			return fmt.Errorf("failed to initialize visualizer: %w", err)
		}
	}
	
	// Market初期化
	err := bt.market.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize market: %w", err)
	}
	
	bt.initialized = true
	
	// Visualizerに状態変更を通知
	if bt.visualizer != nil {
		bt.visualizer.OnBacktestStateChange(BacktestStateRunning)
	}
	
	return nil
}

// initializeVisualizer はVisualizerを初期化します（内部メソッド）
func (bt *Backtester) initializeVisualizer(ctx context.Context) error {
	// visualizer.Configに変換
	vizConfig := &visualizer.Config{
		Port:              bt.visualizerConfig.Port,
		ReadTimeout:       bt.visualizerConfig.ReadTimeout,
		WriteTimeout:      bt.visualizerConfig.WriteTimeout,
		MaxClients:        bt.visualizerConfig.MaxClients,
		HeartbeatInterval: bt.visualizerConfig.HeartbeatInterval,
		ClientTimeout:     bt.visualizerConfig.ClientTimeout,
		BufferSize:        bt.visualizerConfig.BufferSize,
		LogLevel:          bt.visualizerConfig.LogLevel,
	}
	
	// Visualizer作成
	bt.visualizerImpl = visualizer.NewVisualizer(vizConfig)
	
	// Visualizer開始
	if err := bt.visualizerImpl.Start(ctx, bt.visualizerConfig.Port); err != nil {
		return fmt.Errorf("failed to start visualizer: %w", err)
	}
	
	// アダプター作成と設定
	adapter := NewVisualizerAdapter(bt.visualizerImpl)
	bt.visualizer = adapter
	
	return nil
}

// Stop はBacktesterとVisualizerを停止します。
func (bt *Backtester) Stop() error {
	// Visualizerを停止
	if bt.visualizerImpl != nil {
		if err := bt.visualizerImpl.Stop(); err != nil {
			return fmt.Errorf("failed to stop visualizer: %w", err)
		}
	}
	
	// Visualizerに状態変更を通知
	if bt.visualizer != nil {
		bt.visualizer.OnBacktestStateChange(BacktestStateStopped)
	}
	
	bt.initialized = false
	return nil
}

// Forward は時間を次のステップに進めます。
func (bt *Backtester) Forward() bool {
	if !bt.initialized {
		return false
	}
	
	// Market時間進行
	hasNext := bt.market.Forward()
	
	// Broker側のポジション価格更新
	if hasNext {
		bt.broker.UpdatePositions()
		
		// Visualizerにローソク足データを通知
		if bt.visualizer != nil {
			// 現在のローソク足データを取得（とりあえずSAMPLEを固定）
			candle := bt.market.GetCurrentCandle("SAMPLE")
			if candle != nil {
				bt.visualizer.OnCandleUpdate(candle)
			}
			
			// 統計情報の更新と通知
			bt.statistics.UpdateBalance(bt.broker.GetBalance())
			bt.visualizer.OnStatisticsUpdate(bt.statistics)
		}
	}
	
	return hasNext
}

// IsFinished はバックテストが終了したかを確認します。
func (bt *Backtester) IsFinished() bool {
	if !bt.initialized {
		return false
	}
	return bt.market.IsFinished()
}

// GetCurrentTime は現在の時刻を取得します。
func (bt *Backtester) GetCurrentTime() time.Time {
	if !bt.initialized {
		return time.Time{}
	}
	return bt.market.GetCurrentTime()
}

// GetCurrentPrice は指定シンボルの現在価格を取得します。
func (bt *Backtester) GetCurrentPrice(symbol string) float64 {
	if !bt.initialized {
		return 0.0
	}
	return bt.market.GetCurrentPrice(symbol)
}

// Buy は買い注文を実行します。
func (bt *Backtester) Buy(symbol string, size float64) error {
	if !bt.initialized {
		return errors.New("backtester not initialized")
	}
	
	// 入力値検証
	if size <= 0 {
		return errors.New("order size must be positive")
	}
	
	// 現在価格確認（存在しないシンボルチェック）
	price := bt.market.GetCurrentPrice(symbol)
	if price <= 0 {
		return fmt.Errorf("invalid symbol or price: %s", symbol)
	}
	
	// 注文作成
	orderID := fmt.Sprintf("buy-%s-%d", symbol, time.Now().UnixNano())
	order := models.NewMarketOrder(orderID, symbol, models.Buy, size)
	
	// Broker経由で注文実行
	err := bt.broker.PlaceOrder(order)
	if err != nil {
		return err
	}
	
	// Visualizerにトレードイベントを通知（ポジションオープン）
	if bt.visualizer != nil {
		// ダミーのトレードイベントを作成
		trade := &models.Trade{
			ID:         orderID,
			Symbol:     symbol,
			Side:       models.Buy,
			Size:       size,
			EntryPrice: price,
			ExitPrice:  0, // まだクローズされていない
			PnL:        0,
			Status:     models.TradeOpen,
			OpenTime:   bt.market.GetCurrentTime(),
			CloseTime:  time.Time{},
		}
		bt.visualizer.OnTradeEvent(trade)
		
		// 統計情報を更新
		bt.statistics.UpdateBalance(bt.broker.GetBalance())
	}
	
	return nil
}

// Sell は売り注文を実行します。
func (bt *Backtester) Sell(symbol string, size float64) error {
	if !bt.initialized {
		return errors.New("backtester not initialized")
	}
	
	// 入力値検証
	if size <= 0 {
		return errors.New("order size must be positive")
	}
	
	// 現在価格確認（存在しないシンボルチェック）
	price := bt.market.GetCurrentPrice(symbol)
	if price <= 0 {
		return fmt.Errorf("invalid symbol or price: %s", symbol)
	}
	
	// 注文作成
	orderID := fmt.Sprintf("sell-%s-%d", symbol, time.Now().UnixNano())
	order := models.NewMarketOrder(orderID, symbol, models.Sell, size)
	
	// Broker経由で注文実行
	err := bt.broker.PlaceOrder(order)
	if err != nil {
		return err
	}
	
	// Visualizerにトレードイベントを通知（ポジションオープン）
	if bt.visualizer != nil {
		// ダミーのトレードイベントを作成
		trade := &models.Trade{
			ID:         orderID,
			Symbol:     symbol,
			Side:       models.Sell,
			Size:       size,
			EntryPrice: price,
			ExitPrice:  0, // まだクローズされていない
			PnL:        0,
			Status:     models.TradeOpen,
			OpenTime:   bt.market.GetCurrentTime(),
			CloseTime:  time.Time{},
		}
		bt.visualizer.OnTradeEvent(trade)
		
		// 統計情報を更新
		bt.statistics.UpdateBalance(bt.broker.GetBalance())
	}
	
	return nil
}

// GetPositions は現在の全ポジションを取得します。
func (bt *Backtester) GetPositions() []*models.Position {
	if !bt.initialized {
		return []*models.Position{}
	}
	return bt.broker.GetPositions()
}

// GetBalance は現在の残高を取得します。
func (bt *Backtester) GetBalance() float64 {
	if !bt.initialized {
		return 0.0
	}
	return bt.broker.GetBalance()
}

// ClosePosition は指定されたポジションを決済します。
func (bt *Backtester) ClosePosition(positionID string) error {
	if !bt.initialized {
		return errors.New("backtester not initialized")
	}
	
	// ポジション情報を取得（クローズ前）
	var position *models.Position
	positions := bt.broker.GetPositions()
	for _, p := range positions {
		if p.ID == positionID {
			position = p
			break
		}
	}
	
	// ポジションをクローズ
	err := bt.broker.ClosePosition(positionID)
	if err != nil {
		return err
	}
	
	// Visualizerにトレードイベントを通知（ポジションクローズ）
	if bt.visualizer != nil && position != nil {
		// クローズ後のトレード履歴から最新のトレードを取得
		tradeHistory := bt.broker.GetTradeHistory()
		if len(tradeHistory) > 0 {
			lastTrade := tradeHistory[len(tradeHistory)-1]
			bt.visualizer.OnTradeEvent(lastTrade)
			
			// 統計情報を更新
			bt.statistics.AddTrade(lastTrade.PnL)
			bt.statistics.UpdateBalance(bt.broker.GetBalance())
			bt.visualizer.OnStatisticsUpdate(bt.statistics)
		}
	}
	
	return nil
}

// CloseAllPositions は全ポジションを決済します。
func (bt *Backtester) CloseAllPositions() error {
	if !bt.initialized {
		return errors.New("backtester not initialized")
	}
	
	positions := bt.broker.GetPositions()
	for _, position := range positions {
		err := bt.broker.ClosePosition(position.ID)
		if err != nil {
			return fmt.Errorf("failed to close position %s: %w", position.ID, err)
		}
	}
	
	return nil
}

// GetTradeHistory は取引履歴を取得します。
func (bt *Backtester) GetTradeHistory() []*models.Trade {
	if !bt.initialized {
		return []*models.Trade{}
	}
	return bt.broker.GetTradeHistory()
}