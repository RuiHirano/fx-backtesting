package data

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// CandleData はシンボル付きローソク足データを表します。
type CandleData struct {
	Symbol string
	Candle *models.Candle
}

// DataProvider はデータ提供者のインターフェースです。
type DataProvider interface {
	StreamData(ctx context.Context) (<-chan CandleData, error)
}

// CSVProvider はCSVファイルからデータを提供します。
type CSVProvider struct {
	Config models.DataProviderConfig
}

// NewCSVProvider は新しいCSVProviderを作成します。
func NewCSVProvider(config models.DataProviderConfig) *CSVProvider {
	return &CSVProvider{
		Config: config,
	}
}

// StreamData はローソク足データをストリーミングします。
func (p *CSVProvider) StreamData(ctx context.Context) (<-chan CandleData, error) {
	// ファイルの存在確認
	if _, err := os.Stat(p.Config.FilePath); os.IsNotExist(err) {
		return nil, errors.New("file not found: " + p.Config.FilePath)
	}
	
	// ファイルを開く
	file, err := os.Open(p.Config.FilePath)
	if err != nil {
		return nil, err
	}
	
	// ファイル名からシンボルを推測
	symbol := p.extractSymbolFromFilename(p.Config.FilePath)
	
	candleChan := make(chan CandleData, 10)
	
	go func() {
		defer close(candleChan)
		defer file.Close()
		
		parser := NewCSVParser(file)
		
		for {
			candle, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					// ファイル終端に達した
					break
				}
				// パースエラーはログに記録してスキップ
				continue
			}
			
			// バリデーション
			if err := candle.Validate(); err != nil {
				// 無効なデータはスキップ
				continue
			}
			
			candleData := CandleData{
				Symbol: symbol,
				Candle: candle,
			}
			
			select {
			case candleChan <- candleData:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return candleChan, nil
}

// extractSymbolFromFilename はファイル名からシンボルを推測します。
func (p *CSVProvider) extractSymbolFromFilename(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	
	// 一般的なシンボル形式を推測
	upper := strings.ToUpper(name)
	if len(upper) >= 6 && (strings.Contains(upper, "USD") || strings.Contains(upper, "EUR") || strings.Contains(upper, "JPY")) {
		return upper
	}
	
	// デフォルトはEURUSD
	return "EURUSD"
}