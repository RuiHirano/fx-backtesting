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
	// 各種設定パターンの例
	fmt.Println("=== 各種設定パターンの例 ===")
	
	// パターン1: 高レバレッジ設定
	fmt.Println("\n--- パターン1: 高レバレッジ設定 ---")
	runBacktestWithConfig("高レバレッジ", backtester.BrokerConfig{
		InitialBalance: 50000.0,  // 初期残高: 5万円
		Spread:         0.005,    // スプレッド: 0.5銭（狭い）
	})
	
	// パターン2: 保守的設定
	fmt.Println("\n--- パターン2: 保守的設定 ---")
	runBacktestWithConfig("保守的", backtester.BrokerConfig{
		InitialBalance: 200000.0, // 初期残高: 20万円
		Spread:         0.02,     // スプレッド: 2銭（広い）
	})
	
	// パターン3: 標準設定
	fmt.Println("\n--- パターン3: 標準設定 ---")
	runBacktestWithConfig("標準", backtester.BrokerConfig{
		InitialBalance: 100000.0, // 初期残高: 10万円
		Spread:         0.01,     // スプレッド: 1銭
	})
	
	// パターン4: 小額設定
	fmt.Println("\n--- パターン4: 小額設定 ---")
	runBacktestWithConfig("小額", backtester.BrokerConfig{
		InitialBalance: 10000.0,  // 初期残高: 1万円
		Spread:         0.015,    // スプレッド: 1.5銭
	})
	
	fmt.Println("\n=== 設定比較完了 ===")
}

func runBacktestWithConfig(configName string, brokerConfig backtester.BrokerConfig) {
	// データプロバイダー設定
	dpConfig := models.DataProviderConfig{
		FilePath: "../../testdata/USDJPY_2024_01.csv",
		Format:   "csv",
	}

	// 市場に関する設定
	marketConfig := backtester.MarketConfig{
		DataProvider: dpConfig,
	}

	// バックテスト全体の設定
	config := backtester.Config{
		Market:     marketConfig,
		Broker:     brokerConfig,
		Visualizer: models.DisabledVisualizerConfig(),
	}

	// Backtester作成
	bt, err := backtester.NewBacktester(config)
	if err != nil {
		log.Printf("Failed to create backtester for %s: %v", configName, err)
		return
	}

	// 初期化
	ctx := context.Background()
	err = bt.Initialize(ctx)
	if err != nil {
		log.Printf("Failed to initialize backtester for %s: %v", configName, err)
		return
	}

	// 簡単な取引戦略
	maxTrades := 5 // 最大取引数を制限
	for i := 0; !bt.IsFinished() && len(bt.GetTradeHistory()) < maxTrades; i++ {
		currentPrice := bt.GetCurrentPrice("USDJPY")
		if currentPrice <= 0 {
			bt.Forward()
			continue
		}

		positions := bt.GetPositions()

		// 5回に1回取引
		if len(positions) == 0 {
			// 残高に応じて取引サイズを調整
			tradeSize := calculateTradeSize(bt.GetBalance(), config.Broker.InitialBalance)

			err = bt.Buy("USDJPY", tradeSize)
		}

		// 時間進行（間隔を空けて取引）
		for j := 0; j < 10 && !bt.IsFinished(); j++ {
			bt.Forward()
		}
		
		// ポジションがあればクローズ
		positionsToClose := bt.GetPositions()
		for _, pos := range positionsToClose {
			bt.ClosePosition(pos.ID)
		}
	}

	// 残りのポジションをクローズ
	bt.CloseAllPositions()

	// 結果表示
	finalBalance := bt.GetBalance()
	trades := bt.GetTradeHistory()
	pnl := finalBalance - config.Broker.InitialBalance
	returnPct := (pnl / config.Broker.InitialBalance) * 100

	fmt.Printf("設定: %s | 初期残高: %.0f円 | 最終残高: %.0f円 | 損益: %.0f円 | リターン: %.2f%% | 取引数: %d\n",
		configName, config.Broker.InitialBalance, finalBalance, pnl, returnPct, len(trades))

	// 統計サマリー
	if len(trades) > 0 {
		report := statistics.NewReport(trades, config.Broker.InitialBalance)
		fmt.Printf("統計: %s\n", report.GenerateCompactSummary())
	}
}

func calculateTradeSize(currentBalance, initialBalance float64) float64 {
	// 残高に応じて取引サイズを調整
	balanceRatio := currentBalance / initialBalance
	
	if balanceRatio > 1.2 {
		return 1500.0 // 残高が増えた場合は大きめ
	} else if balanceRatio < 0.8 {
		return 500.0  // 残高が減った場合は小さめ
	}
	
	return 1000.0 // 標準サイズ
}

// 設定例を生成する関数
func generateConfigExamples() {
	fmt.Println("\n=== 設定ファイル例 ===")
	
	// 基本設定例
	fmt.Println("基本設定 (config_basic.json):")
	fmt.Printf(`{
  "market": {
    "data_provider": {
      "file_path": "./data/USDJPY.csv",
      "format": "csv"
    },
    "symbol": "USDJPY"
  },
  "broker": {
    "initial_balance": 100000.0,
    "spread": 0.01
  }
}

`)
	
	// 高リスク設定例
	fmt.Println("高リスク設定 (config_high_risk.json):")
	fmt.Printf(`{
  "market": {
    "data_provider": {
      "file_path": "./data/USDJPY.csv",
      "format": "csv"
    },
    "symbol": "USDJPY"
  },
  "broker": {
    "initial_balance": 50000.0,
    "spread": 0.005
  }
}

`)
	
	// 保守的設定例
	fmt.Println("保守的設定 (config_conservative.json):")
	fmt.Printf(`{
  "market": {
    "data_provider": {
      "file_path": "./data/USDJPY.csv",
      "format": "csv"
    },
    "symbol": "USDJPY"
  },
  "broker": {
    "initial_balance": 200000.0,
    "spread": 0.02
  }
}

`)
}