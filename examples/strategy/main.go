package main

import (
	"context"
	"fmt"
	"log"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
)

// SimpleMovingAverage は簡単な移動平均を計算します
type SimpleMovingAverage struct {
	period int
	prices []float64
}

func NewSMA(period int) *SimpleMovingAverage {
	return &SimpleMovingAverage{
		period: period,
		prices: make([]float64, 0, period),
	}
}

func (sma *SimpleMovingAverage) Add(price float64) {
	sma.prices = append(sma.prices, price)
	if len(sma.prices) > sma.period {
		sma.prices = sma.prices[1:]
	}
}

func (sma *SimpleMovingAverage) Value() float64 {
	if len(sma.prices) < sma.period {
		return 0
	}
	
	sum := 0.0
	for _, price := range sma.prices {
		sum += price
	}
	return sum / float64(len(sma.prices))
}

func main() {
	// 戦略的なバックテストの例（移動平均クロス戦略）
	fmt.Println("=== 移動平均クロス戦略バックテスト ===")

	// データプロバイダー設定
	dpConfig := models.DataProviderConfig{
		FilePath: "../../testdata/USDJPY_2024_01.csv", // 実際のデータファイルパス
		Format:   "csv",
	}

	// 市場に関する設定
	marketConfig := backtester.MarketConfig{
		DataProvider: dpConfig,
	}

	// ブローカーに関する設定
	brokerConfig := backtester.BrokerConfig{
		InitialBalance: 100000.0, // 初期残高: 10万円
		Spread:         0.01,     // スプレッド: 1銭
	}

	// バックテスト全体の設定
	config := backtester.Config{
		Market:     marketConfig,
		Broker:     brokerConfig,
		Backtest:   backtester.BacktestConfig{}, // 期間制限なし
		Visualizer: models.DisabledVisualizerConfig(),
	}

	// Backtester作成
	bt, err := backtester.NewBacktester(config)
	if err != nil {
		log.Fatalf("Failed to create backtester: %v", err)
	}

	// 初期化
	ctx := context.Background()
	err = bt.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize backtester: %v", err)
	}

	// 移動平均の準備
	shortMA := NewSMA(5)  // 5期間の短期移動平均
	longMA := NewSMA(20) // 20期間の長期移動平均

	fmt.Println("移動平均クロス戦略実行中...")

	for !bt.IsFinished() {
		// 現在価格取得
		currentPrice := bt.GetCurrentPrice()
		if currentPrice <= 0 {
			bt.Forward()
			continue
		}

		// 移動平均に価格を追加
		shortMA.Add(currentPrice)
		longMA.Add(currentPrice)

		// 移動平均値取得
		shortValue := shortMA.Value()
		longValue := longMA.Value()

		// 現在のポジション確認
		positions := bt.GetPositions()

		// 戦略実行（移動平均クロス）
		if shortValue > 0 && longValue > 0 {
			if len(positions) == 0 && shortValue > longValue {
				// ゴールデンクロス：買いエントリー
				err = bt.Buy("USDJPY", 1000.0)
				if err == nil {
					fmt.Printf("ゴールデンクロスで買いエントリー (価格: %.2f, 短期MA: %.2f, 長期MA: %.2f)\n",
						currentPrice, shortValue, longValue)
				}
			} else if len(positions) > 0 && shortValue < longValue {
				// デッドクロス：ポジションクローズ
				for _, pos := range positions {
					err = bt.ClosePosition(pos.ID)
					if err == nil {
						fmt.Printf("デッドクロスでクローズ (価格: %.2f, 短期MA: %.2f, 長期MA: %.2f)\n",
							currentPrice, shortValue, longValue)
					}
				}
			}
		}

		// 時間を進める
		bt.Forward()
	}

	// 残りのポジションをクローズ
	finalPositions := bt.GetPositions()
	for _, pos := range finalPositions {
		err = bt.ClosePosition(pos.ID)
		if err != nil {
			fmt.Printf("ポジション %s のクローズに失敗: %v\n", pos.ID, err)
		}
	}

	// 最終残高と取引履歴を取得
	finalBalance := bt.GetBalance()
	trades := bt.GetTradeHistory()

	// 結果表示
	fmt.Printf("\n=== バックテスト結果 ===\n")
	fmt.Printf("戦略: 移動平均クロス (短期: %d期間, 長期: %d期間)\n", 5, 20)
	fmt.Printf("初期残高: %.2f円\n", config.Broker.InitialBalance)
	fmt.Printf("最終残高: %.2f円\n", finalBalance)
	fmt.Printf("総損益: %.2f円\n", finalBalance-config.Broker.InitialBalance)
	fmt.Printf("リターン: %.2f%%\n", ((finalBalance-config.Broker.InitialBalance)/config.Broker.InitialBalance)*100)
	fmt.Printf("実行した取引数: %d\n", len(trades))

	// 詳細統計レポート生成
	if len(trades) > 0 {
		fmt.Printf("\n=== 詳細統計レポート ===\n")
		report := statistics.NewReport(trades, config.Broker.InitialBalance)

		// 基本統計
		calculator := statistics.NewCalculator(trades)
		fmt.Printf("勝率: %.2f%%\n", calculator.CalculateWinRate()*100)
		fmt.Printf("平均利益: %.2f円\n", calculator.CalculateAverageProfit())
		fmt.Printf("平均損失: %.2f円\n", calculator.CalculateAverageLoss())
		fmt.Printf("最大ドローダウン: %.2f円\n", calculator.CalculateMaxDrawdown())
		fmt.Printf("プロフィットファクター: %.2f\n", calculator.CalculateProfitFactor())
		fmt.Printf("シャープレシオ: %.4f\n", calculator.CalculateSharpeRatio())

		// 簡潔なサマリー
		fmt.Printf("\n簡潔サマリー: %s\n", report.GenerateCompactSummary())
	}

	fmt.Println("\n移動平均クロス戦略バックテスト完了!")
}