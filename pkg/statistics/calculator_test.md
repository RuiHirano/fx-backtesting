# Statistics テスト仕様書

## 概要
- **テスト対象**: `pkg/statistics/` の統計計算・レポート生成・メトリクス管理機能
- **テスト目的**: バックテスト結果の統計分析、レポート生成、メトリクス管理の検証
- **テストコンポーネント**: 
  - **Calculator**: 統計計算エンジン
  - **Report**: レポート生成機能
  - **Metrics**: メトリクス定義・管理機能

## Calculator テスト内容

### TestCalculator_NewCalculator
```go
func TestCalculator_NewCalculator(t *testing.T) {
    trades := createTestTrades()
    calculator := NewCalculator(trades)
    
    // Calculator作成確認
    if calculator == nil { ... }
    
    // 初期状態確認
    if len(calculator.GetTrades()) != len(trades) { ... }
}
```
- **テスト目的**: Calculator コンストラクタの検証
- **テスト条件**: 取引履歴データを使用してCalculator作成
- **検証項目**: Calculator作成成功、初期状態確認、nilトレード除外処理

### TestCalculator_BasicMetrics
```go
func TestCalculator_BasicMetrics(t *testing.T) {
    trades := []*models.Trade{
        createTrade("trade-1", 100.0, time.Now()),   // 利益
        createTrade("trade-2", -50.0, time.Now()),   // 損失
        createTrade("trade-3", 200.0, time.Now()),   // 利益
        createTrade("trade-4", -30.0, time.Now()),   // 損失
        createTrade("trade-5", 80.0, time.Now()),    // 利益
    }
    calculator := NewCalculator(trades)
    
    // 総利益・損失テスト (300.0)
    totalPnL := calculator.CalculateTotalPnL()
    
    // 勝率テスト (60%)
    winRate := calculator.CalculateWinRate()
    
    // 平均利益・損失テスト
    avgProfit := calculator.CalculateAverageProfit()
    avgLoss := calculator.CalculateAverageLoss()
    
    // 最大利益・損失テスト
    maxProfit := calculator.CalculateMaxProfit()
    maxLoss := calculator.CalculateMaxLoss()
}
```
- **テスト目的**: 基本統計指標の計算検証
- **テスト条件**: 
  - 事前条件: 利益・損失が混在した5つの取引データ
  - 入力値: 利益3回（100, 200, 80）、損失2回（-50, -30）
  - 期待結果: 総PnL=300.0、勝率=60%、平均利益=126.67、平均損失=40.0
- **検証項目**: 
  - 総利益・損失の正確な計算
  - 勝率の正確な計算（3/5 = 60%）
  - 平均利益・損失の分離計算
  - 最大利益・損失の識別
  - 取引回数カウント

### TestCalculator_RiskMetrics
```go
func TestCalculator_RiskMetrics(t *testing.T) {
    // ドローダウンパターンのテストデータ
    trades := []*models.Trade{
        createTrade("trade-1", 100.0, time.Now()),   // 累積: 100
        createTrade("trade-2", -200.0, time.Now()),  // 累積: -100
        createTrade("trade-3", -100.0, time.Now()),  // 累積: -200
        createTrade("trade-4", 300.0, time.Now()),   // 累積: 100
        createTrade("trade-5", 50.0, time.Now()),    // 累積: 150
    }
    calculator := NewCalculator(trades)
    
    // 最大ドローダウンテスト (300.0)
    maxDrawdown := calculator.CalculateMaxDrawdown()
    
    // シャープレシオテスト
    sharpeRatio := calculator.CalculateSharpeRatio()
    
    // ソルティノレシオテスト
    sortinoRatio := calculator.CalculateSortinoRatio()
}
```
- **テスト目的**: リスク指標の計算検証
- **テスト条件**: 
  - 事前条件: ドローダウンが発生するパターンの取引データ
  - 入力値: 累積損益の変動（100→-100→-200→100→150）
  - 期待結果: 最大DD=300.0（100から-200への下落）
- **検証項目**: 
  - 最大ドローダウンの正確な計算
  - シャープレシオの正の値確認
  - ソルティノレシオの計算（下方偏差考慮）
  - リターン・リスク比の計算

### TestCalculator_AdvancedMetrics
```go
func TestCalculator_AdvancedMetrics(t *testing.T) {
    trades := createTestTrades()
    calculator := NewCalculator(trades)
    
    // カルマーレシオテスト
    calmarRatio := calculator.CalculateCalmarRatio()
    
    // プロフィットファクターテスト
    profitFactor := calculator.CalculateProfitFactor()
    
    // 期待値テスト
    expectedValue := calculator.CalculateExpectedValue()
    
    // 標準偏差テスト
    stdDev := calculator.CalculateStandardDeviation()
}
```
- **テスト目的**: 高度な統計指標の計算検証
- **テスト条件**: 複合的な取引データパターン
- **検証項目**: 
  - カルマーレシオ（リターン/最大DD）
  - プロフィットファクター（総利益/総損失）
  - 期待値（平均リターン）
  - 標準偏差（リターンのばらつき）

### TestCalculator_TradingMetrics
```go
func TestCalculator_TradingMetrics(t *testing.T) {
    // 時間間隔のある取引履歴作成
    baseTime := time.Now()
    trades := []*models.Trade{
        createTrade("trade-1", 100.0, baseTime),
        createTrade("trade-2", -50.0, baseTime.Add(24*time.Hour)),
        createTrade("trade-3", 75.0, baseTime.Add(48*time.Hour)),
    }
    calculator := NewCalculator(trades)
    
    // 取引パフォーマンス指標
    avgHoldingPeriod := calculator.CalculateAverageHoldingPeriod()
    maxConsecutiveWins := calculator.CalculateMaxConsecutiveWins()
    maxConsecutiveLosses := calculator.CalculateMaxConsecutiveLosses()
    tradingFrequency := calculator.CalculateTradingFrequency()
    riskRewardRatio := calculator.CalculateRiskRewardRatio()
}
```
- **テスト目的**: 取引関連指標の計算検証
- **テスト条件**: 時間間隔を持つ取引データ（24時間間隔）
- **検証項目**: 
  - 平均保有期間の計算
  - 最大連勝・連敗の追跡
  - 取引頻度（取引/日）の計算
  - リスクリワード比の計算

### TestCalculator_ErrorHandling
```go
func TestCalculator_ErrorHandling(t *testing.T) {
    // 空の取引履歴テスト
    emptyTrades := []*models.Trade{}
    calculator := NewCalculator(emptyTrades)
    
    // ゼロ除算回避確認
    totalPnL := calculator.CalculateTotalPnL()     // 0.0
    winRate := calculator.CalculateWinRate()       // 0.0
    avgProfit := calculator.CalculateAverageProfit() // 0.0
    
    // nilトレード処理テスト
    nilTrades := []*models.Trade{nil}
    calculatorWithNil := NewCalculator(nilTrades)
    totalTradesWithNil := calculatorWithNil.CalculateTotalTrades() // 0
}
```
- **テスト目的**: エラーハンドリングの検証
- **テスト条件**: 空データ、nilデータ、ゼロ除算ケース
- **検証項目**: 適切なデフォルト値返却、ゼロ除算回避、nil除外処理

## Report テスト内容

### TestReport_NewReport
- **テスト目的**: Report作成とBacktestResult統合の検証
- **検証項目**: Calculator・Result初期化、統計情報の自動計算

### TestReport_GenerateTextReport
- **テスト目的**: 日本語テキスト形式レポート生成の検証
- **検証項目**: 必要セクション（基本情報、損益情報、取引統計、リスク指標、取引パフォーマンス）の包含確認

### TestReport_GenerateJSONReport
- **テスト目的**: JSON形式レポート生成の検証
- **検証項目**: JSON構造の妥当性、必要フィールドの包含確認

### TestReport_GenerateCSVReport
- **テスト目的**: CSV形式取引履歴レポート生成の検証
- **検証項目**: ヘッダー行、データ行数、フィールド数の確認

### TestReport_GetSummaryMetrics
- **テスト目的**: 要約メトリクス取得機能の検証
- **検証項目**: 13種類の主要メトリクス包含確認、データ型の正確性

## Metrics テスト内容

### TestMetricsSet_NewMetricsSet
- **テスト目的**: MetricsSetコンストラクタの検証
- **検証項目**: 初期化状態、タイムスタンプ設定

### TestGenerateMetricsFromCalculator
- **テスト目的**: Calculator→MetricsSet変換機能の検証
- **検証項目**: 全メトリクス生成確認、値の妥当性検証

### TestMetricsSet_GetBasicMetrics / GetRiskMetrics / GetTradingMetrics
- **テスト目的**: メトリクス分類機能の検証
- **検証項目**: 適切なメトリクス分類、期待されるメトリクス数

## 結果（テスト数と実績）
- **Calculator テスト数**: 6個（全統計計算機能網羅）
- **Report テスト数**: 7個（全レポート形式対応）
- **Metrics テスト数**: 9個（メトリクス管理機能）
- **総テスト数**: 22個
- **カバレッジ**: 92.3%

## 実装された統計指標
### 基本統計
- Total PnL, Win Rate, Total Trades, Average Profit/Loss, Max Profit/Loss

### リスク指標  
- Max Drawdown, Sharpe Ratio, Sortino Ratio, Calmar Ratio, Standard Deviation

### 取引パフォーマンス
- Max Consecutive Wins/Losses, Average Holding Period, Trading Frequency, Risk Reward Ratio

### 高度指標
- Profit Factor, Expected Value, Return Risk Ratio

## レポート形式
1. **テキスト形式**: 日本語での詳細レポート（セクション分割）
2. **JSON形式**: 構造化データ（API連携対応）
3. **CSV形式**: 取引履歴詳細（スプレッドシート対応）

## メトリクス管理
- **22種類のメトリクス定義**: 基本・リスク・取引パフォーマンス分類
- **メトリクス分類機能**: 用途別メトリクス抽出
- **タイムスタンプ管理**: メトリクス生成時刻の記録

## 技術仕様詳細
1. **ゼロ除算対策**: 全計算関数でゼロ除算チェック実装
2. **nil データ処理**: 無効データの自動除外機能
3. **時間計算**: time.Duration による正確な期間計算
4. **浮動小数点精度**: 金融計算に適した精度管理
5. **メモリ効率**: 大量取引データ対応の効率的なデータ構造

## テスト実行
```bash
# 全テスト実行
go test ./pkg/statistics/... -v

# カバレッジ確認
go test ./pkg/statistics/... -cover

# 個別テスト実行
go test -run TestCalculator_BasicMetrics ./pkg/statistics/
go test -run TestReport_GenerateTextReport ./pkg/statistics/
go test -run TestGenerateMetricsFromCalculator ./pkg/statistics/
```