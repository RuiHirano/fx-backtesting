package data

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// DataProvider defines the interface for data providers
type DataProvider interface {
	LoadCSVData(filePath string) ([]models.Candle, error)
	GetCandle(timestamp time.Time) (models.Candle, error)
}

// CSVDataProvider implements DataProvider for CSV files
type CSVDataProvider struct {
	candles map[time.Time]models.Candle
}

// NewCSVDataProvider creates a new CSV data provider
func NewCSVDataProvider() *CSVDataProvider {
	return &CSVDataProvider{
		candles: make(map[time.Time]models.Candle),
	}
}

// LoadCSVData loads historical data from a CSV file
func (p *CSVDataProvider) LoadCSVData(filePath string) ([]models.Candle, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var candles []models.Candle
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		candle, err := parseCSVLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line '%s': %w", line, err)
		}

		candles = append(candles, candle)
		p.candles[candle.Timestamp] = candle
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return candles, nil
}

// GetCandle retrieves a candle for a specific timestamp
func (p *CSVDataProvider) GetCandle(timestamp time.Time) (models.Candle, error) {
	candle, exists := p.candles[timestamp]
	if !exists {
		return models.Candle{}, fmt.Errorf("candle not found for timestamp: %v", timestamp)
	}
	return candle, nil
}

// parseCSVLine parses a single CSV line into a Candle
func parseCSVLine(line string) (models.Candle, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 6 {
		return models.Candle{}, fmt.Errorf("invalid CSV format: expected 6 fields, got %d", len(parts))
	}

	// Parse timestamp
	timestamp, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(parts[0]))
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Parse OHLC values
	open, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid open price: %w", err)
	}

	high, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid high price: %w", err)
	}

	low, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid low price: %w", err)
	}

	close, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid close price: %w", err)
	}

	// Parse volume
	volume, err := strconv.ParseInt(strings.TrimSpace(parts[5]), 10, 64)
	if err != nil {
		return models.Candle{}, fmt.Errorf("invalid volume: %w", err)
	}

	return models.NewCandle(timestamp, open, high, low, close, volume), nil
}