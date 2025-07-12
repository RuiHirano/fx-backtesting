package market

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDataProvider is a mock for the DataProvider interface
type MockDataProvider struct {
	mock.Mock
}

func (m *MockDataProvider) TimeToIndex(t time.Time) (int, error) {
	args := m.Called(t)
	return args.Int(0), args.Error(1)
}

func (m *MockDataProvider) IndexToTime(i int) (time.Time, error) {
	args := m.Called(i)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockDataProvider) GetCandlesByTime(ctx context.Context, startTime, endTime time.Time) ([]models.Candle, error) {
	args := m.Called(ctx, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func (m *MockDataProvider) GetCandlesByIndex(ctx context.Context, startIndex, endIndex int) ([]models.Candle, error) {
	args := m.Called(ctx, startIndex, endIndex)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func (m *MockDataProvider) GetPrevCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error) {
	args := m.Called(ctx, baseTime, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func (m *MockDataProvider) GetPrevCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error) {
	args := m.Called(ctx, baseIndex, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func (m *MockDataProvider) GetNextCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error) {
	args := m.Called(ctx, baseTime, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func (m *MockDataProvider) GetNextCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error) {
	args := m.Called(ctx, baseIndex, count)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Candle), args.Error(1)
}

func TestMarket_Initialize(t *testing.T) {
	t.Run("INIT-001: Normal case", func(t *testing.T) {
		mockProvider := new(MockDataProvider)
		cacheSize := 500
		candles := make([]models.Candle, cacheSize)
		for i := 0; i < cacheSize; i++ {
			candles[i] = models.Candle{Timestamp: time.Now().Add(time.Duration(i) * time.Minute)}
		}

		mockProvider.On("GetCandlesByIndex", mock.Anything, 0, cacheSize-1).Return(candles, nil)

		market := NewMarket(models.MarketConfig{})
		market.provider = mockProvider
		err := market.Initialize(context.Background())

		assert.NoError(t, err)
		assert.True(t, market.initialized)
		assert.False(t, market.IsFinished())
		assert.Len(t, market.candleCache, cacheSize)
		assert.Equal(t, 0, market.currentIndex)
		mockProvider.AssertExpectations(t)
	})

	t.Run("INIT-003: No data available", func(t *testing.T) {
		mockProvider := new(MockDataProvider)
		cacheSize := 500
		mockProvider.On("GetCandlesByIndex", mock.Anything, 0, cacheSize-1).Return([]models.Candle{}, nil)

		market := NewMarket(models.MarketConfig{})
		market.provider = mockProvider
		err := market.Initialize(context.Background())

		assert.NoError(t, err)
		assert.True(t, market.initialized)
		assert.True(t, market.IsFinished())
		assert.Empty(t, market.candleCache)
		mockProvider.AssertExpectations(t)
	})

	t.Run("INIT-004: Error from provider", func(t *testing.T) {
		mockProvider := new(MockDataProvider)
		cacheSize := 500
		mockProvider.On("GetCandlesByIndex", mock.Anything, 0, cacheSize-1).Return(nil, errors.New("provider error"))

		market := NewMarket(models.MarketConfig{})
		market.provider = mockProvider
		err := market.Initialize(context.Background())

		assert.Error(t, err)
		assert.Equal(t, "provider error", err.Error())
		assert.False(t, market.initialized)
		mockProvider.AssertExpectations(t)
	})
}

func TestMarket_Forward(t *testing.T) {
	initialTime := time.Now()
	setup := func(t *testing.T, cacheSize, initialDataSize, refillDataSize, refillThreshold int) (*MarketImpl, *MockDataProvider) {
		mockProvider := new(MockDataProvider)

		initialCandles := make([]models.Candle, initialDataSize)
		for i := 0; i < initialDataSize; i++ {
			initialCandles[i] = models.Candle{Timestamp: initialTime.Add(time.Duration(i) * time.Minute), Close: float64(100 + i)}
		}
		mockProvider.On("GetCandlesByIndex", mock.Anything, 0, cacheSize-1).Return(initialCandles, nil).Once()

		if refillDataSize > 0 {
			refillCandles := make([]models.Candle, refillDataSize)
			for i := 0; i < refillDataSize; i++ {
				refillCandles[i] = models.Candle{Timestamp: initialTime.Add(time.Duration(initialDataSize+i) * time.Minute), Close: float64(100 + initialDataSize + i)}
			}
			mockProvider.On("GetCandlesByIndex", mock.Anything, initialDataSize, initialDataSize+cacheSize-1).Return(refillCandles, nil).Once()
		} else {
			// When refillDataSize is 0, expect a call for more data and return an empty slice
			mockProvider.On("GetCandlesByIndex", mock.Anything, mock.Anything, mock.Anything).Return([]models.Candle{}, nil).Maybe()
		}

		market := NewMarket(models.MarketConfig{})
		market.provider = mockProvider
		market.refillThreshold = refillThreshold
		market.cacheSize = cacheSize
		err := market.Initialize(context.Background())
		assert.NoError(t, err)
		return market, mockProvider
	}

	t.Run("FWD-001: Forward once", func(t *testing.T) {
		market, _ := setup(t, 20, 20, 0, 10) // Keep refillThreshold=10 for this test
		success := market.Forward()
		assert.True(t, success)
		assert.Equal(t, 1, market.currentIndex)
		assert.Equal(t, initialTime.Add(1*time.Minute), market.GetCurrentTime())
		assert.Equal(t, 101.0, market.GetCurrentPrice())
	})

	t.Run("FWD-002: Forward to trigger refill", func(t *testing.T) {
		market, mockProvider := setup(t, 20, 20, 10, 10) // cacheSize=20, initialData=20, refillData=10, refillThreshold=10

		// Forward until refill threshold is met (currentIndex will be 10)
		for i := 0; i < 10; i++ {
			market.Forward()
		}
		assert.Equal(t, 10, market.currentIndex)

		// This forward should use the refilled cache
		success := market.Forward()
		assert.True(t, success)
		assert.Equal(t, 11, market.currentIndex)
		assert.Len(t, market.candleCache, 30) // 20 initial + 10 refill
		mockProvider.AssertExpectations(t)
	})

	t.Run("FWD-003: Forward to the end", func(t *testing.T) {
		market, _ := setup(t, 10, 10, 0, 1) // cacheSize=10, initialData=10, no refill, refillThreshold=1

		// Forward 9 times
		for i := 0; i < 9; i++ {
			assert.True(t, market.Forward(), "Forward should succeed at step %d", i)
		}
		assert.Equal(t, 9, market.currentIndex)
		assert.False(t, market.IsFinished())

		// The 10th forward fails as we are at the end
		success := market.Forward()
		assert.False(t, success)
		assert.True(t, market.IsFinished())
	})
}

func TestMarket_GetPrevCandles(t *testing.T) {
	mockProvider := new(MockDataProvider)
	baseTime := time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC)
	candles := make([]models.Candle, 20)
	for i := 0; i < 20; i++ {
		candles[i] = models.Candle{Timestamp: baseTime.Add(time.Duration(i) * time.Minute)}
	}
	mockProvider.On("GetCandlesByIndex", mock.Anything, 0, 499).Return(candles, nil)
	// Mock additional calls for Forward operations that might trigger refill
	mockProvider.On("GetCandlesByIndex", mock.Anything, mock.Anything, mock.Anything).Return([]models.Candle{}, nil).Maybe()

	market := NewMarket(models.MarketConfig{})
	market.provider = mockProvider
	market.Initialize(context.Background())

	// Move forward 10 steps, so currentIndex is 10
	for i := 0; i < 10; i++ {
		market.Forward()
	}
	assert.Equal(t, 10, market.currentIndex)

	t.Run("GET-001: Get previous 5 candles from index 10", func(t *testing.T) {
		startTime := baseTime.Add(5 * time.Minute) // from 5th minute
		prevCandles := market.GetPrevCandles(startTime, 10)
		assert.Len(t, prevCandles, 5)
		assert.Equal(t, baseTime.Add(5*time.Minute), prevCandles[0].Timestamp)
		assert.Equal(t, baseTime.Add(9*time.Minute), prevCandles[4].Timestamp)
	})

	t.Run("GET-002: Start time is after index time", func(t *testing.T) {
		startTime := baseTime.Add(15 * time.Minute)
		prevCandles := market.GetPrevCandles(startTime, 10)
		assert.Empty(t, prevCandles)
	})

	t.Run("GET-005: Index out of range", func(t *testing.T) {
		startTime := baseTime
		prevCandles := market.GetPrevCandles(startTime, -1)
		assert.Empty(t, prevCandles)

		prevCandles = market.GetPrevCandles(startTime, 100)
		assert.Empty(t, prevCandles)
	})
}