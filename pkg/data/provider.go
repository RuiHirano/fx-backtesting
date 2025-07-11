package data

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// DataProvider はデータ提供者のインターフェースです。
type DataProvider interface {
	StreamData(ctx context.Context) (<-chan models.Candle, error)
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
func (p *CSVProvider) StreamData(ctx context.Context) (<-chan models.Candle, error) {
	// ファイルの存在確認
	if _, err := os.Stat(p.Config.FilePath); os.IsNotExist(err) {
		return nil, errors.New("file not found: " + p.Config.FilePath)
	}
	
	// ファイルを開く
	file, err := os.Open(p.Config.FilePath)
	if err != nil {
		return nil, err
	}
	
	candleChan := make(chan models.Candle, 10)
	
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
			
			// 期間フィルタリング
			if !p.isWithinTimeRange(candle.Time) {
				continue
			}
			
			select {
			case candleChan <- *candle:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return candleChan, nil
}

// isWithinTimeRange は指定された時刻が期間内かどうかを判定します
func (p *CSVProvider) isWithinTimeRange(candleTime time.Time) bool {
	// 開始時刻のチェック
	if p.Config.StartTime != nil && candleTime.Before(*p.Config.StartTime) {
		return false
	}
	
	// 終了時刻のチェック
	if p.Config.EndTime != nil && candleTime.After(*p.Config.EndTime) {
		return false
	}
	
	return true
}

// extractSymbolFromFilename はファイル名からシンボルを推測します。
func (p *CSVProvider) extractSymbolFromFilename(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// "_"で分割して最初の部分をシンボルとする
	parts := strings.Split(name, "_")
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}

	// デフォルトはEURUSD
	return "EURUSD"
}