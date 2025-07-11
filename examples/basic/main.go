package main

import (
	"context"
	"fmt"
	"log"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
	"github.com/RuiHirano/fx-backtesting/pkg/statistics"
)

func main() {
	// 基本的なバックテストの例
	fmt.Println("=== 基本的なバックテスト例 ===")

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

	fmt.Println("バックテスト実行中...")

	// 簡単な取引戦略の実行
	for i := 0; !bt.IsFinished() && i < 2000; i++ { // 最初の2000ステップのみ実行
		// 100ステップごとに買い注文
		if i%100 == 0 {
			positions := bt.GetPositions()
			if len(positions) == 0 {
				currentPrice := bt.GetCurrentPrice("USDJPY")
				err := bt.Buy("USDJPY", 1000.0) // 1000通貨単位で買い
				if err == nil {
					fmt.Printf("ステップ %d: 価格 %.2f で買い注文実行\n", i, currentPrice)
				}
			}
		}

		// 50ステップごとにポジションをクローズ
		if i%50 == 0 {
			positions := bt.GetPositions()
			if len(positions) > 0 {
				for _, pos := range positions {
					bt.ClosePosition(pos.ID)
				}
			}
		}

		// 時間を進める
		bt.Forward()
	}

	// 残りのポジションをクローズ
	bt.CloseAllPositions()

	// 最終残高と取引履歴を取得
	finalBalance := bt.GetBalance()
	trades := bt.GetTradeHistory()

	// 結果表示
	fmt.Printf("\n=== バックテスト結果 ===\n")
	fmt.Printf("初期残高: %.2f円\n", config.Broker.InitialBalance)
	fmt.Printf("最終残高: %.2f円\n", finalBalance)
	fmt.Printf("総損益: %.2f円\n", finalBalance-config.Broker.InitialBalance)
	fmt.Printf("リターン: %.2f%%\n", ((finalBalance-config.Broker.InitialBalance)/config.Broker.InitialBalance)*100)
	fmt.Printf("実行した取引数: %d\n", len(trades))

	// 統計レポート生成（取引があった場合）
	if len(trades) > 0 {
		fmt.Printf("\n=== 統計レポート ===\n")
		report := statistics.NewReport(trades, config.Broker.InitialBalance)
		fmt.Print(report.GenerateCompactSummary())
	}

	fmt.Println("\n\nバックテスト完了!")
}