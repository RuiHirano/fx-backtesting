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
)

// Backtester はバックテスト実行とユーザーAPIを提供する統括コンポーネントです。
type Backtester struct {
	market      market.Market
	broker      broker.Broker
	initialized bool
}

// NewBacktester は新しいBacktesterを作成します。
func NewBacktester(dataConfig models.DataProviderConfig, brokerConfig models.BrokerConfig) *Backtester {
	// DataProvider作成
	provider := data.NewCSVProvider(dataConfig)
	
	// Market作成
	mkt := market.NewMarket(provider)
	
	// Broker作成  
	bkr := broker.NewSimpleBroker(brokerConfig, mkt)
	
	return &Backtester{
		market:      mkt,
		broker:      bkr,
		initialized: false,
	}
}

// Initialize はBacktesterを初期化します。
func (bt *Backtester) Initialize(ctx context.Context) error {
	// Market初期化
	err := bt.market.Initialize(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize market: %w", err)
	}
	
	bt.initialized = true
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
	return bt.broker.PlaceOrder(order)
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
	return bt.broker.PlaceOrder(order)
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
	return bt.broker.ClosePosition(positionID)
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