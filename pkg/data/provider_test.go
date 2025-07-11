package data

import (
	"context"
	"testing"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

func TestCSVProvider_TimeToIndex(t *testing.T) {
	tests := []struct {
		name    string
		config  models.DataProviderConfig
		time    time.Time
		want    int
		wantErr bool
	}{
		{
			name: "valid time conversion",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			time:    time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			want:    0,
			wantErr: false,
		},
		{
			name: "time not found - should return nearest",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			time:    time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC),
			want:    30,
			wantErr: false,
		},
		{
			name: "time before first entry",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			time:    time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC),
			want:    0,
			wantErr: false,
		},
		{
			name: "time after last entry",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			time:    time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
			want:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			got, err := provider.TimeToIndex(tt.time)
			if (err != nil) != tt.wantErr {
				t.Errorf("TimeToIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TimeToIndex() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCSVProvider_IndexToTime(t *testing.T) {
	tests := []struct {
		name    string
		config  models.DataProviderConfig
		index   int
		want    time.Time
		wantErr bool
	}{
		{
			name: "valid index conversion",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			index:   0,
			want:    time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name: "index out of range - negative",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			index:   -1,
			want:    time.Time{},
			wantErr: true,
		},
		{
			name: "index out of range - too large",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			index:   10000,
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			got, err := provider.IndexToTime(tt.index)
			if (err != nil) != tt.wantErr {
				t.Errorf("IndexToTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("IndexToTime() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCSVProvider_GetCandlesByTime(t *testing.T) {
	tests := []struct {
		name      string
		config    models.DataProviderConfig
		startTime time.Time
		endTime   time.Time
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid time range",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			wantCount: 61,
			wantErr:   false,
		},
		{
			name: "invalid time range - start after end",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startTime: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "empty range",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startTime: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			endTime:   time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetCandlesByTime(ctx, tt.startTime, tt.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCandlesByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetCandlesByTime() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_GetCandlesByIndex(t *testing.T) {
	tests := []struct {
		name       string
		config     models.DataProviderConfig
		startIndex int
		endIndex   int
		wantCount  int
		wantErr    bool
	}{
		{
			name: "valid index range",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startIndex: 0,
			endIndex:   59,
			wantCount:  60,
			wantErr:    false,
		},
		{
			name: "invalid index range - start after end",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startIndex: 59,
			endIndex:   0,
			wantCount:  0,
			wantErr:    true,
		},
		{
			name: "index out of range",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startIndex: 0,
			endIndex:   10000,
			wantCount:  0,
			wantErr:    true,
		},
		{
			name: "single index",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			startIndex: 0,
			endIndex:   0,
			wantCount:  1,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetCandlesByIndex(ctx, tt.startIndex, tt.endIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCandlesByIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetCandlesByIndex() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_GetPrevCandlesByTime(t *testing.T) {
	tests := []struct {
		name      string
		config    models.DataProviderConfig
		baseTime  time.Time
		count     int
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid prev candles",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			count:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "request more than available",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 9, 5, 0, 0, time.UTC),
			count:     10,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "base time at start",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			count:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "zero count",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			count:     0,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetPrevCandlesByTime(ctx, tt.baseTime, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrevCandlesByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetPrevCandlesByTime() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_GetPrevCandlesByIndex(t *testing.T) {
	tests := []struct {
		name      string
		config    models.DataProviderConfig
		baseIndex int
		count     int
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid prev candles",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 60,
			count:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "request more than available",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 5,
			count:     10,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "base index at start",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 0,
			count:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "negative base index",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: -1,
			count:     10,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetPrevCandlesByIndex(ctx, tt.baseIndex, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrevCandlesByIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetPrevCandlesByIndex() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_GetNextCandlesByTime(t *testing.T) {
	tests := []struct {
		name      string
		config    models.DataProviderConfig
		baseTime  time.Time
		count     int
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid next candles",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
			count:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "request more than available",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 16, 55, 0, 0, time.UTC),
			count:     10,
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "base time at end",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseTime:  time.Date(2024, 1, 1, 16, 59, 0, 0, time.UTC),
			count:     10,
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetNextCandlesByTime(ctx, tt.baseTime, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextCandlesByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetNextCandlesByTime() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_GetNextCandlesByIndex(t *testing.T) {
	tests := []struct {
		name      string
		config    models.DataProviderConfig
		baseIndex int
		count     int
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid next candles",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 0,
			count:     10,
			wantCount: 10,
			wantErr:   false,
		},
		{
			name: "request more than available",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 475,
			count:     10,
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "base index at end",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 479,
			count:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "base index out of range",
			config: models.DataProviderConfig{
				FilePath: "testdata/sample.csv",
				Format:   "csv",
			},
			baseIndex: 10000,
			count:     10,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewCSVProvider(tt.config)
			ctx := context.Background()
			got, err := provider.GetNextCandlesByIndex(ctx, tt.baseIndex, tt.count)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextCandlesByIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("GetNextCandlesByIndex() got %d candles, want %d", len(got), tt.wantCount)
			}
		})
	}
}

func TestCSVProvider_FileNotExists(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "testdata/nonexistent.csv",
		Format:   "csv",
	}

	provider := NewCSVProvider(config)
	ctx := context.Background()

	// All methods should return an error when file doesn't exist
	_, err := provider.TimeToIndex(time.Now())
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.IndexToTime(0)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetCandlesByTime(ctx, time.Now(), time.Now())
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetCandlesByIndex(ctx, 0, 1)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetPrevCandlesByTime(ctx, time.Now(), 1)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetPrevCandlesByIndex(ctx, 1, 1)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetNextCandlesByTime(ctx, time.Now(), 1)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	_, err = provider.GetNextCandlesByIndex(ctx, 1, 1)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestCSVProvider_DataConsistency(t *testing.T) {
	config := models.DataProviderConfig{
		FilePath: "testdata/sample.csv",
		Format:   "csv",
	}

	provider := NewCSVProvider(config)
	ctx := context.Background()

	// Test data consistency between different methods
	startTime := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	// Get candles by time
	candlesByTime, err := provider.GetCandlesByTime(ctx, startTime, endTime)
	if err != nil {
		t.Fatalf("GetCandlesByTime failed: %v", err)
	}

	// Get equivalent candles by index
	startIndex, err := provider.TimeToIndex(startTime)
	if err != nil {
		t.Fatalf("TimeToIndex failed: %v", err)
	}

	endIndex, err := provider.TimeToIndex(endTime)
	if err != nil {
		t.Fatalf("TimeToIndex failed: %v", err)
	}

	candlesByIndex, err := provider.GetCandlesByIndex(ctx, startIndex, endIndex)
	if err != nil {
		t.Fatalf("GetCandlesByIndex failed: %v", err)
	}

	// Compare results
	if len(candlesByTime) != len(candlesByIndex) {
		t.Errorf("Data inconsistency: GetCandlesByTime returned %d candles, GetCandlesByIndex returned %d",
			len(candlesByTime), len(candlesByIndex))
	}

	// Compare first and last candles
	if len(candlesByTime) > 0 && len(candlesByIndex) > 0 {
		if !candlesByTime[0].Timestamp.Equal(candlesByIndex[0].Timestamp) {
			t.Errorf("First candle timestamp mismatch: %v vs %v",
				candlesByTime[0].Timestamp, candlesByIndex[0].Timestamp)
		}

		lastIdx := len(candlesByTime) - 1
		if !candlesByTime[lastIdx].Timestamp.Equal(candlesByIndex[lastIdx].Timestamp) {
			t.Errorf("Last candle timestamp mismatch: %v vs %v",
				candlesByTime[lastIdx].Timestamp, candlesByIndex[lastIdx].Timestamp)
		}
	}
}