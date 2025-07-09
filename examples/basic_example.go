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
	dataConfig := models.DataProviderConfig{
		FilePath: "../testdata/USDJPY_2024_01.csv", // 実際のデータファイルパス
		Format:   "csv",
	}
	
	// ブローカー設定
	brokerConfig := models.BrokerConfig{
		InitialBalance: 100000.0, // 初期残高: 10万円
		Spread:         0.01,     // スプレッド: 1銭
	}
	
	// Backtester作成
	bt := backtester.NewBacktester(dataConfig, brokerConfig)
	
	// 初期化
	ctx := context.Background()
	err := bt.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize backtester: %v", err)
	}
	
	// 簡単な取引戦略の実行
	var trades []*models.Trade
	tradeCount := 0
	
	fmt.Println("バックテスト実行中...")
	
	for !bt.IsFinished() {
		fmt.Printf("isFinished: %v, 現在の時間: %s\n", bt.IsFinished(), bt.GetCurrentTime().Format("2006-01-02 15:04:05"))
		// 現在価格取得
		currentPrice := bt.GetCurrentPrice("USDJPY")
		if currentPrice <= 0 {
			bt.Forward()
			continue
		}
		fmt.Printf("現在価格: %.2f\n", currentPrice)
		
		// 現在のポジション確認
		positions := bt.GetPositions()
		
		// 簡単な戦略: 10回に1回買い注文
		if len(positions) == 0 && tradeCount < 10 {
			err = bt.Buy("USDJPY", 1000.0) // 1000通貨単位で買い
			if err == nil {
				tradeCount++
				fmt.Printf("取引 %d: 価格 %.2f で買い注文実行\n", tradeCount, currentPrice)
				
				// 取引記録（簡易版）
				trade := &models.Trade{
					ID:         fmt.Sprintf("trade-%d", tradeCount),
					Symbol:     "USDJPY",
					Side:       models.Buy,
					Size:       1000.0,
					EntryPrice: currentPrice,
					ExitPrice:  currentPrice, // 仮の値
					PnL:        0.0,          // 後で計算
					Status:     models.TradeOpen,
					OpenTime:   bt.GetCurrentTime(),
				}
				trades = append(trades, trade)
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
	
	// 最終残高
	finalBalance := bt.GetBalance()
	
	// 結果表示
	fmt.Printf("\n=== バックテスト結果 ===\n")
	fmt.Printf("初期残高: %.2f円\n", brokerConfig.InitialBalance)
	fmt.Printf("最終残高: %.2f円\n", finalBalance)
	fmt.Printf("総損益: %.2f円\n", finalBalance-brokerConfig.InitialBalance)
	fmt.Printf("リターン: %.2f%%\n", ((finalBalance-brokerConfig.InitialBalance)/brokerConfig.InitialBalance)*100)
	fmt.Printf("実行した取引数: %d\n", tradeCount)
	
	// 統計レポート生成（取引があった場合）
	if len(trades) > 0 {
		fmt.Printf("\n=== 統計レポート ===\n")
		report := statistics.NewReport(trades, brokerConfig.InitialBalance)
		fmt.Print(report.GenerateCompactSummary())
	}
	
	fmt.Println("\n\nバックテスト完了!")
}