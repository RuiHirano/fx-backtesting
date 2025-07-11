// Package visualmode demonstrates the visual mode functionality
// To run this example independently:
//
//	go run visualize_example.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RuiHirano/fx-backtesting/pkg/backtester"
	"github.com/RuiHirano/fx-backtesting/pkg/models"
)

// SimpleMovingAverageStrategy はシンプルな移動平均クロス戦略を実装します
type SimpleMovingAverageStrategy struct {
	backtester *backtester.Backtester
	prices     []float64
	windowSize int
}

// NewSimpleMovingAverageStrategy は新しいSimpleMovingAverageStrategyを作成します
func NewSimpleMovingAverageStrategy(bt *backtester.Backtester, windowSize int) *SimpleMovingAverageStrategy {
	return &SimpleMovingAverageStrategy{
		backtester: bt,
		prices:     make([]float64, 0),
		windowSize: windowSize,
	}
}

// calculateMovingAverage は移動平均を計算します
func (s *SimpleMovingAverageStrategy) calculateMovingAverage() float64 {
	if len(s.prices) < s.windowSize {
		return 0
	}
	
	sum := 0.0
	for i := len(s.prices) - s.windowSize; i < len(s.prices); i++ {
		sum += s.prices[i]
	}
	
	return sum / float64(s.windowSize)
}

// onTick は新しい価格データを処理します
func (s *SimpleMovingAverageStrategy) onTick(price float64, symbol string) error {
	s.prices = append(s.prices, price)
	
	// 過去のデータが多すぎる場合は削除
	if len(s.prices) > s.windowSize*2 {
		s.prices = s.prices[1:]
	}
	
	// 移動平均の計算
	ma := s.calculateMovingAverage()
	if ma == 0 {
		return nil // まだ計算できない
	}
	
	currentPrice := price
	
	// 現在のポジション状況を確認
	positions := s.backtester.GetPositions()
	hasPosition := len(positions) > 0
	
	// シンプルな戦略: 現在価格が移動平均より上なら買い、下なら売り
	if currentPrice > ma && !hasPosition {
		// 買いシグナル
		fmt.Printf("📈 買いシグナル: 現在価格=%.5f, MA=%.5f\n", currentPrice, ma)
		return s.backtester.Buy(symbol, 1000)
	} else if currentPrice < ma && hasPosition {
		// 売りシグナル（ポジション決済）
		fmt.Printf("📉 売りシグナル: 現在価格=%.5f, MA=%.5f\n", currentPrice, ma)
		for _, pos := range positions {
			if err := s.backtester.ClosePosition(pos.ID); err != nil {
				return fmt.Errorf("ポジション決済エラー: %w", err)
			}
		}
	}
	
	return nil
}

// printBacktestStatistics は統計情報を表示します
func printBacktestStatistics(bt *backtester.Backtester) {
	balance := bt.GetBalance()
	positions := bt.GetPositions()
	trades := bt.GetTradeHistory()
	
	fmt.Printf("\n📊 === 統計情報 ===\n")
	fmt.Printf("現在残高: %.2f\n", balance)
	fmt.Printf("オープンポジション数: %d\n", len(positions))
	fmt.Printf("総取引数: %d\n", len(trades))
	
	if len(trades) > 0 {
		totalPnL := 0.0
		wins := 0
		
		for _, trade := range trades {
			totalPnL += trade.PnL
			if trade.PnL > 0 {
				wins++
			}
		}
		
		winRate := float64(wins) / float64(len(trades)) * 100
		fmt.Printf("総損益: %.2f\n", totalPnL)
		fmt.Printf("勝率: %.1f%%\n", winRate)
	}
	fmt.Printf("==================\n\n")
}

func main() {
	fmt.Println("🚀 FX Backtesting Visual Mode Example")
	fmt.Println("======================================")
	
	// 設定
	dataConfig := models.DataProviderConfig{
		FilePath: "../../testdata/USDJPY_2024_01.csv", // 実際のデータファイルパスに変更してください
		Format:   "csv",
	}
	
	brokerConfig := models.BrokerConfig{
		InitialBalance: 100000.0, // 初期残高 10万円
		Spread:         0.0001,   // 0.1 pips
	}
	
	// Visualizer設定
	visualizerConfig := models.DefaultVisualizerConfig()
	visualizerConfig.Port = 8080
	
	// Backtester作成（Visualizer統合）
	fmt.Println("🤖 Backtester を初期化中...")
	bt := backtester.NewBacktesterWithVisualizer(dataConfig, brokerConfig, visualizerConfig)
	
	// Graceful shutdown用のコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Backtester初期化（Visualizerも自動的に開始される）
	if err := bt.Initialize(ctx); err != nil {
		log.Fatalf("Backtester初期化エラー: %v", err)
	}
	defer bt.Stop()
	
	fmt.Printf("✅ BacktesterとVisualizer（ポート %d）が初期化されました\n", 8080)
	fmt.Println("🌐 フロントエンドを開始するには:")
	fmt.Println("   1. cd ../../frontend/visual-mode")
	fmt.Println("   2. npm run dev")
	fmt.Println("   3. ブラウザで表示されるURL（通常 http://localhost:5173 または http://localhost:5174）を開く")
	
	// 戦略作成
	strategy := NewSimpleMovingAverageStrategy(bt, 10) // 10期移動平均
	
	fmt.Println("📈 シンプル移動平均戦略を開始します")
	fmt.Println("戦略: 現在価格が10期移動平均より上で買い、下で売り")
	fmt.Println()
	
	// シグナル処理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigCh
		fmt.Println("\n🛑 終了シグナルを受信しました...")
		cancel()
	}()
	
	// バックテスト実行
	fmt.Println("🔄 バックテスト開始...")
	
	stepCount := 0
	lastStatsTime := time.Now()
	
	// データに応じたシンボルを取得（USDJPYまたはSAMPLE）
	symbol := "USDJPY" // 実際のデータファイルに合わせて変更
	
	for !bt.IsFinished() {
		select {
		case <-ctx.Done():
			fmt.Println("バックテストを中断します...")
			goto cleanup
		default:
		}
		
		// 時間を進める
		if !bt.Forward() {
			break
		}
		
		stepCount++
		
		// 現在価格を取得
		currentPrice := bt.GetCurrentPrice(symbol)
		if currentPrice > 0 {
			// 戦略実行
			if err := strategy.onTick(currentPrice, symbol); err != nil {
				fmt.Printf("⚠️  戦略実行エラー: %v\n", err)
			}
		}
		
		// 進捗表示（100ステップごと）
		if stepCount%100 == 0 {
			fmt.Printf("⏰ ステップ %d 処理完了 (現在時刻: %s, 価格: %.5f)\n", 
				stepCount, bt.GetCurrentTime().Format("2006-01-02 15:04:05"), currentPrice)
		}
		
		// 統計情報表示（30秒ごと）
		if time.Since(lastStatsTime) >= 30*time.Second {
			printBacktestStatistics(bt)
			lastStatsTime = time.Now()
		}
		
		// 可視化のため少し待機（実際の運用では不要）
		time.Sleep(50 * time.Millisecond)
	}
	
cleanup:
	// 残りのポジションを決済
	fmt.Println("\n🔄 残ポジションを決済中...")
	if err := bt.CloseAllPositions(); err != nil {
		fmt.Printf("⚠️  決済エラー: %v\n", err)
	}
	
	// 最終統計情報
	fmt.Println("\n🏁 バックテスト完了!")
	printBacktestStatistics(bt)
	
	// 詳細なトレード履歴
	trades := bt.GetTradeHistory()
	if len(trades) > 0 {
		fmt.Printf("📋 === 取引履歴 (最新5件) ===\n")
		start := len(trades) - 5
		if start < 0 {
			start = 0
		}
		
		for i := start; i < len(trades); i++ {
			trade := trades[i]
			status := "❌"
			if trade.PnL > 0 {
				status = "✅"
			}
			
			fmt.Printf("%s [%s] %s %.0f@%.5f → %.5f (損益: %.2f)\n",
				status,
				trade.OpenTime.Format("15:04:05"),
				trade.Side.String(),
				trade.Size,
				trade.EntryPrice,
				trade.ExitPrice,
				trade.PnL,
			)
		}
		fmt.Printf("========================\n")
	}
	
	fmt.Println("\n💡 Visualizer サーバーは引き続き実行中です")
	fmt.Println("   フロントエンドでリアルタイムデータを確認できます")
	fmt.Println("   終了するには Ctrl+C を押してください")
	
	// サーバーを実行し続ける
	<-ctx.Done()
	fmt.Println("\n👋 プログラムを終了します...")
}