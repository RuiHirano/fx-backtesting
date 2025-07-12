package market

import (
	"context"
	"sync"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/data"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// Market defines the interface for the market simulation.
type Market interface {
	Initialize(ctx context.Context) error
	Forward() bool
	GetCurrentPrice(symbol string) float64
	GetCurrentTime() time.Time
	GetCurrentCandle(symbol string) *models.Candle
	GetPrevCandles(startTime time.Time, index int) []*models.Candle
	IsFinished() bool
}

// MarketImpl implements the Market interface.
type MarketImpl struct {
	provider        data.DataProvider
	candleCache     []*models.Candle
	currentIndex    int
	cacheSize       int
	refillThreshold int
	finished        bool
	initialized     bool
	mu              sync.Mutex
	lastIndexFetched int
}

// NewMarket creates a new MarketImpl with default cache settings.
func NewMarket(marketConfig models.MarketConfig) *MarketImpl {
	provider := data.NewCSVProvider(marketConfig.DataProvider)
	cacheSize := 500
	refillThreshold := 100
	
	return &MarketImpl{
		provider:        provider,
		cacheSize:       cacheSize,
		refillThreshold: refillThreshold,
		currentIndex:    -1, // Start before the first element
		candleCache:     make([]*models.Candle, 0, cacheSize),
	}
}

// Initialize fetches the initial set of candles into the cache.
func (m *MarketImpl) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	candles, err := m.provider.GetCandlesByIndex(ctx, 0, m.cacheSize-1)
	if err != nil {
		return err
	}

	for i := range candles {
		m.candleCache = append(m.candleCache, &candles[i])
	}

	m.lastIndexFetched = len(m.candleCache) - 1

	if len(m.candleCache) == 0 {
		m.finished = true
	} else {
		m.currentIndex = 0
	}

	m.initialized = true
	return nil
}

// Forward moves the market to the next time step.
func (m *MarketImpl) Forward() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.finished {
		return false
	}

	// Check if we need to refill the cache
	if len(m.candleCache)-m.currentIndex <= m.refillThreshold {
		startIndex := m.lastIndexFetched + 1
		endIndex := startIndex + m.cacheSize - 1
		newCandles, err := m.provider.GetCandlesByIndex(context.Background(), startIndex, endIndex)
		if err == nil && len(newCandles) > 0 {
			for i := range newCandles {
				m.candleCache = append(m.candleCache, &newCandles[i])
			}
			m.lastIndexFetched = endIndex
		}
	}

	if m.currentIndex+1 >= len(m.candleCache) {
		m.finished = true
		return false
	}

	m.currentIndex++
	return true
}

// GetCurrentPrice returns the closing price of the current candle.
func (m *MarketImpl) GetCurrentPrice(symbol string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.currentIndex < 0 || m.currentIndex >= len(m.candleCache) {
		return 0.0
	}
	return m.candleCache[m.currentIndex].Close
}

// GetCurrentTime returns the timestamp of the current candle.
func (m *MarketImpl) GetCurrentTime() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.currentIndex < 0 || m.currentIndex >= len(m.candleCache) {
		return time.Time{}
	}
	return m.candleCache[m.currentIndex].Timestamp
}

// GetCurrentCandle returns the current candle.
func (m *MarketImpl) GetCurrentCandle(symbol string) *models.Candle {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.currentIndex < 0 || m.currentIndex >= len(m.candleCache) {
		return nil
	}
	return m.candleCache[m.currentIndex]
}

// GetPrevCandles returns a slice of candles from startTime up to (but not including) the given index.
func (m *MarketImpl) GetPrevCandles(startTime time.Time, index int) []*models.Candle {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || index <= 0 || index > len(m.candleCache) {
		return []*models.Candle{}
	}

	if startTime.After(m.candleCache[index-1].Timestamp) {
		return []*models.Candle{}
	}

	var startIndex int
	// Find the starting index by searching backwards from the given index
	for i := index - 1; i >= 0; i-- {
		if m.candleCache[i].Timestamp.Before(startTime) {
			startIndex = i + 1
			break
		}
	}

	// Ensure slice is not out of bounds
	if startIndex >= index {
		return []*models.Candle{}
	}

	return m.candleCache[startIndex:index]
}

// IsFinished returns true if the market simulation has ended.
func (m *MarketImpl) IsFinished() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.finished
}