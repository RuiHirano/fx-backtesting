package market

import (
	"context"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Market は市場データを管理するインターフェースです。
type Market interface {
	Initialize(ctx context.Context) error
	Forward() bool
	GetCurrentPrice(symbol string) float64
	GetCurrentTime() time.Time
	GetCurrentCandle(symbol string) *models.Candle
	IsFinished() bool
}

// MarketImpl はMarketインターフェースの実装です。
type MarketImpl struct {
	provider      data.DataProvider
	currentData   map[string]*models.Candle
	currentTime   time.Time
	candleChannel <-chan data.CandleData
	finished      bool
	initialized   bool
}

// NewMarket は新しいMarketを作成します。
func NewMarket(provider data.DataProvider) Market {
	return &MarketImpl{
		provider:    provider,
		currentData: make(map[string]*models.Candle),
		finished:    false,
		initialized: false,
	}
}

// Initialize は市場を初期化します。
func (m *MarketImpl) Initialize(ctx context.Context) error {
	// DataProviderからデータストリームを取得
	candleChannel, err := m.provider.StreamData(ctx)
	if err != nil {
		return err
	}
	
	m.candleChannel = candleChannel
	
	// 最初のデータを読み込み
	if candleData, ok := <-m.candleChannel; ok {
		m.currentData[candleData.Symbol] = candleData.Candle
		m.currentTime = candleData.Candle.Timestamp
	}
	
	m.initialized = true
	return nil
}

// Forward は次の時間に進みます。
func (m *MarketImpl) Forward() bool {
	if !m.initialized || m.finished {
		return false
	}
	
	// 次のデータを読み込み
	if candleData, ok := <-m.candleChannel; ok {
		m.currentData[candleData.Symbol] = candleData.Candle
		m.currentTime = candleData.Candle.Timestamp
		return true
	}
	
	// データがなくなった場合は終了
	m.finished = true
	return false
}

// GetCurrentPrice は現在の価格を取得します。
func (m *MarketImpl) GetCurrentPrice(symbol string) float64 {
	if candle, exists := m.currentData[symbol]; exists {
		return candle.Close
	}
	return 0.0
}

// GetCurrentTime は現在の時刻を取得します。
func (m *MarketImpl) GetCurrentTime() time.Time {
	return m.currentTime
}

// GetCurrentCandle は現在のローソク足を取得します。
func (m *MarketImpl) GetCurrentCandle(symbol string) *models.Candle {
	if candle, exists := m.currentData[symbol]; exists {
		return candle
	}
	return nil
}

// IsFinished は市場が終了したかを確認します。
func (m *MarketImpl) IsFinished() bool {
	return m.finished
}