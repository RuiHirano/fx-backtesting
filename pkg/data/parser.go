package data

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// CSVParser はCSVファイルを解析します。
type CSVParser struct {
	reader *csv.Reader
}

// NewCSVParser は新しいCSVParserを作成します。
func NewCSVParser(reader io.Reader) *CSVParser {
	csvReader := csv.NewReader(reader)
	return &CSVParser{
		reader: csvReader,
	}
}

// Parse は次のローソク足データを解析します。
func (p *CSVParser) Parse() (*models.Candle, error) {
	record, err := p.reader.Read()
	if err != nil {
		return nil, err
	}
	
	// ヘッダー行をスキップ
	if strings.Contains(record[0], "timestamp") {
		return p.Parse() // 再帰的に次の行を読む
	}
	
	if len(record) < 6 {
		return nil, fmt.Errorf("invalid CSV record: expected 6 fields, got %d", len(record))
	}
	
	// タイムスタンプの解析
	timestamp, err := time.Parse("2006-01-02 15:04:05", record[0])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}
	
	// 価格データの解析
	open, err := strconv.ParseFloat(record[1], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid open price: %w", err)
	}
	
	high, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid high price: %w", err)
	}
	
	low, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid low price: %w", err)
	}
	
	close, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid close price: %w", err)
	}
	
	volume, err := strconv.ParseFloat(record[5], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid volume: %w", err)
	}
	
	return models.NewCandle(timestamp, open, high, low, close, volume), nil
}