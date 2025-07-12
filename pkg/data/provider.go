package data

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// CandleIndex は軽量インデックスエントリです。
type CandleIndex struct {
	Timestamp  time.Time
	FileOffset int64
	LineNumber int
}

// DataProvider はデータ提供者のインターフェースです。
type DataProvider interface {
	// Time・Index変換
	TimeToIndex(time.Time) (int, error)
	IndexToTime(int) (time.Time, error)
	
	// 期間指定データ取得
	GetCandlesByTime(ctx context.Context, startTime, endTime time.Time) ([]models.Candle, error)
	GetCandlesByIndex(ctx context.Context, startIndex, endIndex int) ([]models.Candle, error)
	
	// 前データ取得
	GetPrevCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error)
	GetPrevCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error)
	
	// 後データ取得
	GetNextCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error)
	GetNextCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error)
}

// CSVProvider はCSVファイルからデータを提供します。
type CSVProvider struct {
	Config  models.DataProviderConfig
	index   []CandleIndex
	indexed bool
}

// NewCSVProvider は新しいCSVProviderを作成します。
func NewCSVProvider(config models.DataProviderConfig) *CSVProvider {
	return &CSVProvider{
		Config:  config,
		index:   make([]CandleIndex, 0),
		indexed: false,
	}
}

// buildIndex はファイルをスキャンして軽量インデックスを構築します。
func (p *CSVProvider) buildIndex() error {
	if p.indexed {
		return nil
	}

	// ファイルの存在確認
	if _, err := os.Stat(p.Config.FilePath); os.IsNotExist(err) {
		return errors.New("file not found: " + p.Config.FilePath)
	}

	file, err := os.Open(p.Config.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	parser := NewCSVParser(file)
	p.index = make([]CandleIndex, 0)
	lineNumber := 0

	for {
		candle, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				break
			}
			lineNumber++
			continue
		}

		// バリデーション
		if err := candle.Validate(); err != nil {
			lineNumber++
			continue
		}

		// バリデーション済みのデータを追加

		// インデックスに追加（ファイルオフセットは簡単化）
		p.index = append(p.index, CandleIndex{
			Timestamp:  candle.Timestamp,
			FileOffset: int64(lineNumber),
			LineNumber: lineNumber,
		})

		lineNumber++
	}

	// 時刻順でソート
	sort.Slice(p.index, func(i, j int) bool {
		return p.index[i].Timestamp.Before(p.index[j].Timestamp)
	})

	p.indexed = true
	return nil
}

// TimeToIndex は時刻をインデックスに変換します。
func (p *CSVProvider) TimeToIndex(t time.Time) (int, error) {
	if err := p.buildIndex(); err != nil {
		return -1, err
	}

	if len(p.index) == 0 {
		return -1, errors.New("no data available")
	}

	// 範囲外チェック
	if t.After(p.index[len(p.index)-1].Timestamp) {
		return -1, errors.New("time after last available data")
	}

	// 開始時刻より前の場合は0を返す
	if t.Before(p.index[0].Timestamp) {
		return 0, nil
	}

	// バイナリサーチで最も近い時刻を見つける
	left, right := 0, len(p.index)-1
	for left <= right {
		mid := (left + right) / 2
		if p.index[mid].Timestamp.Equal(t) {
			return mid, nil
		} else if p.index[mid].Timestamp.Before(t) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	// 最も近い時刻のインデックスを返す
	if right >= 0 {
		return right, nil
	}
	return 0, nil
}

// IndexToTime はインデックスを時刻に変換します。
func (p *CSVProvider) IndexToTime(index int) (time.Time, error) {
	if err := p.buildIndex(); err != nil {
		return time.Time{}, err
	}

	if index < 0 || index >= len(p.index) {
		return time.Time{}, errors.New("index out of range")
	}

	return p.index[index].Timestamp, nil
}

// GetCandlesByTime は指定された時間範囲のローソク足データを取得します。
func (p *CSVProvider) GetCandlesByTime(ctx context.Context, startTime, endTime time.Time) ([]models.Candle, error) {
	if startTime.After(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	startIndex, err := p.TimeToIndex(startTime)
	if err != nil {
		return nil, err
	}

	endIndex, err := p.TimeToIndex(endTime)
	if err != nil {
		return nil, err
	}

	return p.GetCandlesByIndex(ctx, startIndex, endIndex)
}

// GetCandlesByIndex は指定されたインデックス範囲のローソク足データを取得します。
func (p *CSVProvider) GetCandlesByIndex(ctx context.Context, startIndex, endIndex int) ([]models.Candle, error) {
	if err := p.buildIndex(); err != nil {
		return nil, err
	}

	if startIndex > endIndex {
		return nil, errors.New("start index must be less than or equal to end index")
	}

	if startIndex < 0 || endIndex >= len(p.index) {
		return nil, errors.New("index out of range")
	}

	candles := make([]models.Candle, 0, endIndex-startIndex+1)

	for i := startIndex; i <= endIndex; i++ {
		candle, err := p.getCandleAtIndex(i)
		if err != nil {
			continue
		}
		candles = append(candles, *candle)
	}

	return candles, nil
}

// GetPrevCandlesByTime は基準時刻より前のローソク足データを取得します。
func (p *CSVProvider) GetPrevCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error) {
	if count <= 0 {
		return []models.Candle{}, nil
	}

	baseIndex, err := p.TimeToIndex(baseTime)
	if err != nil {
		return nil, err
	}

	return p.GetPrevCandlesByIndex(ctx, baseIndex, count)
}

// GetPrevCandlesByIndex は基準インデックスより前のローソク足データを取得します。
func (p *CSVProvider) GetPrevCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error) {
	if err := p.buildIndex(); err != nil {
		return nil, err
	}

	if baseIndex < 0 || baseIndex >= len(p.index) {
		return nil, errors.New("base index out of range")
	}

	if count <= 0 {
		return []models.Candle{}, nil
	}

	startIndex := baseIndex - count
	if startIndex < 0 {
		startIndex = 0
	}

	endIndex := baseIndex - 1
	if endIndex < 0 {
		return []models.Candle{}, nil
	}

	return p.GetCandlesByIndex(ctx, startIndex, endIndex)
}

// GetNextCandlesByTime は基準時刻より後のローソク足データを取得します。
func (p *CSVProvider) GetNextCandlesByTime(ctx context.Context, baseTime time.Time, count int) ([]models.Candle, error) {
	if count <= 0 {
		return []models.Candle{}, nil
	}

	baseIndex, err := p.TimeToIndex(baseTime)
	if err != nil {
		return nil, err
	}

	return p.GetNextCandlesByIndex(ctx, baseIndex, count)
}

// GetNextCandlesByIndex は基準インデックスより後のローソク足データを取得します。
func (p *CSVProvider) GetNextCandlesByIndex(ctx context.Context, baseIndex int, count int) ([]models.Candle, error) {
	if err := p.buildIndex(); err != nil {
		return nil, err
	}

	if baseIndex < 0 || baseIndex >= len(p.index) {
		return nil, errors.New("base index out of range")
	}

	if count <= 0 {
		return []models.Candle{}, nil
	}

	startIndex := baseIndex + 1
	if startIndex >= len(p.index) {
		return []models.Candle{}, nil
	}

	endIndex := baseIndex + count
	if endIndex >= len(p.index) {
		endIndex = len(p.index) - 1
	}

	return p.GetCandlesByIndex(ctx, startIndex, endIndex)
}

// getCandleAtIndex は指定されたインデックスのローソク足データを取得します。
func (p *CSVProvider) getCandleAtIndex(index int) (*models.Candle, error) {
	if index < 0 || index >= len(p.index) {
		return nil, errors.New("index out of range")
	}

	file, err := os.Open(p.Config.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	parser := NewCSVParser(file)
	
	// 指定されたライン番号まで読み飛ばす
	targetLine := p.index[index].LineNumber
	for i := 0; i <= targetLine; i++ {
		candle, err := parser.Parse()
		if err != nil {
			return nil, err
		}
		if i == targetLine {
			return candle, nil
		}
	}

	return nil, errors.New("candle not found")
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