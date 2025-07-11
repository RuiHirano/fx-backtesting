package backtester

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/broker"
	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/market"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/visualizer"
)

// BacktestState はバックテストの状態を表す
type BacktestState = models.BacktestState

const (
	BacktestStateIdle = models.BacktestStateIdle
	BacktestStateRunning = models.BacktestStateRunning
	BacktestStatePaused = models.BacktestStatePaused
	BacktestStateStopped = models.BacktestStateStopped
	BacktestStateCompleted = models.BacktestStateCompleted
	BacktestStateError = models.BacktestStateError
)

// Backtester はバックテスト実行とユーザーAPIを提供する統括コンポーネントです。
type Backtester struct {
	market           market.Market
	broker           broker.Broker
	visualizer   visualizer.Visualizer
	visualizerConfig models.VisualizerConfig
	initialized      bool
	statistics       *models.Statistics
	// バックテスト制御関連
	backtestController *BacktestController
	controlMutex     sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// BacktestController はバックテストのコントロールを管理
type BacktestController struct {
	bt              *Backtester
	speedCh         chan float64
	playCh          chan bool
	state           models.BacktestControlState
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
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
	
	// コンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	
	bt := &Backtester{
		market:           mkt,
		broker:           bkr,
		visualizer:   nil,
		visualizerConfig: visualizerConfig,
		initialized:      false,
		statistics:       models.NewStatistics(brokerConfig.InitialBalance),
		ctx:              ctx,
		cancel:           cancel,
	}
	
	// BacktestControllerを作成
	if visualizerConfig.Enabled {
		bt.backtestController = NewBacktestController(bt)
	}
	
	return bt
}

// NewBacktestController は新しいBacktestControllerを作成
func NewBacktestController(bt *Backtester) *BacktestController {
	ctx, cancel := context.WithCancel(context.Background())
	
	controller := &BacktestController{
		bt:      bt,
		speedCh: make(chan float64, 1),
		playCh:  make(chan bool, 1),
		state:   models.BacktestControlState{IsPlaying: false, Speed: 1.0, State: models.BacktestStateIdle},
		ctx:     ctx,
		cancel:  cancel,
	}
	
	// コントロールループを開始
	go controller.controlLoop()
	
	return controller
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
		bt.visualizer.OnBacktestStateChange(models.BacktestStateRunning)
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
	bt.visualizer = visualizer.NewVisualizer(vizConfig)
	
	// BacktestControllerをVisualizerに設定
	if bt.backtestController != nil {
		bt.visualizer.SetBacktestController(bt.backtestController)
	}
	
	// Visualizer開始
	if err := bt.visualizer.Start(ctx, bt.visualizerConfig.Port); err != nil {
		return fmt.Errorf("failed to start visualizer: %w", err)
	}
	
	return nil
}

// Stop はBacktesterとVisualizerを停止します。
func (bt *Backtester) Stop() error {
	// BacktestControllerを停止
	if bt.backtestController != nil {
		bt.backtestController.Stop()
	}
	
	// Visualizerを停止
	if bt.visualizer != nil {
		if err := bt.visualizer.Stop(); err != nil {
			return fmt.Errorf("failed to stop visualizer: %w", err)
		}
	}
	
	// Visualizerに状態変更を通知
	if bt.visualizer != nil {
		bt.visualizer.OnBacktestStateChange(models.BacktestStateStopped)
	}
	
	bt.initialized = false
	return nil
}

// SetContext はバックテストのコンテキストを設定します
func (bt *Backtester) SetContext(ctx context.Context) {
	bt.ctx = ctx
}

// Forward は時間を次のステップに進めます。
func (bt *Backtester) Forward() bool {
	if !bt.initialized {
		return false
	}
	
	// コントロールモードが有効な場合のチェック
	if bt.backtestController != nil {
		// コントロールモードではコントローラーが再生状態の時のみ進む
		for !bt.backtestController.IsRunning() {
			// コンテキストのキャンセルをチェック
			select {
			case <-bt.ctx.Done():
				fmt.Println("Backtest interrupted by context cancellation")
				return false
			default:
				// 一時停止中は実際に待機
				time.Sleep(100 * time.Millisecond)
				
				// バックテストが完全に終了した場合のチェック
				if bt.market.IsFinished() {
					return false
				}
			}
		}
		
		// 速度制御のための待機
		bt.controlMutex.RLock()
		speed := bt.backtestController.GetState().Speed
		bt.controlMutex.RUnlock()
		
		if speed > 0 {
			waitTime := time.Duration(float64(time.Millisecond*50) / speed)
			// 速度制御の待機中もコンテキストをチェック
			select {
			case <-bt.ctx.Done():
				fmt.Println("Backtest interrupted during speed control")
				return false
			case <-time.After(waitTime):
				// 速度制御の待機終了
			}
		}
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
			fmt.Printf("Current Candle: %v\n", candle)
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

// BacktestController のメソッド群

// Play はバックテストを開始/再開
func (bc *BacktestController) Play(speed float64) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.state.IsPlaying = true
	bc.state.Speed = speed
	bc.state.State = models.BacktestStateRunning
	
	// 非ブロッキングで状態を送信
	select {
	case bc.playCh <- true:
	default:
	}
	
	select {
	case bc.speedCh <- speed:
	default:
	}
	
	fmt.Printf("Backtest started with speed: %f\n", speed)
	return nil
}

// Pause はバックテストを一時停止
func (bc *BacktestController) Pause() error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.state.IsPlaying = false
	bc.state.State = models.BacktestStatePaused
	
	// 非ブロッキングで状態を送信
	select {
	case bc.playCh <- false:
	default:
	}
	
	fmt.Printf("Backtest paused\n")
	return nil
}

// SetSpeed はバックテストの速度を設定
func (bc *BacktestController) SetSpeed(speed float64) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	bc.state.Speed = speed
	
	// 非ブロッキングで速度を送信
	select {
	case bc.speedCh <- speed:
	default:
	}
	
	fmt.Printf("Backtest speed changed to: %f\n", speed)
	return nil
}

// GetState は現在の状態を取得
func (bc *BacktestController) GetState() models.BacktestControlState {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.state
}

// IsRunning はバックテストが実行中かを確認
func (bc *BacktestController) IsRunning() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	return bc.state.IsPlaying
}

// controlLoop はコントロールループを実行
func (bc *BacktestController) controlLoop() {
	for {
		select {
		case <-bc.ctx.Done():
			return
		case isPlaying := <-bc.playCh:
			bc.bt.controlMutex.Lock()
			if isPlaying {
				fmt.Printf("Backtest control: Play\n")
			} else {
				fmt.Printf("Backtest control: Pause\n")
			}
			bc.bt.controlMutex.Unlock()
		case speed := <-bc.speedCh:
			bc.bt.controlMutex.Lock()
			bc.state.Speed = speed
			bc.bt.controlMutex.Unlock()
			fmt.Printf("Backtest control: Speed changed to %f\n", speed)
		}
	}
}

// Stop はBacktestControllerを停止
func (bc *BacktestController) Stop() {
	bc.cancel()
}

// Cancel はバックテストをキャンセルします
func (bt *Backtester) Cancel() {
	if bt.cancel != nil {
		bt.cancel()
	}
}