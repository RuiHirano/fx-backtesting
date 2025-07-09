# Statistics 設計書

## 1. 概要

Statistics パッケージは、FXバックテストライブラリにおいて統計計算とレポート生成を担当するコンポーネントです。取引履歴から各種パフォーマンス指標を計算し、テキスト、JSON、CSVなど複数のフォーマットでレポートを生成します。

## 2. 責務

- バックテスト結果の統計計算
- パフォーマンス指標の算出（PnL、勝率、シャープレシオ等）
- リスク指標の計算（最大ドローダウン、ボラティリティ等）
- 複数フォーマットでのレポート生成
- 時系列分析とトレンド分析
- 比較レポートの生成

## 3. ファイル構成

```
pkg/statistics/
├── calculator.go       # 統計計算エンジン
├── report.go           # レポート生成
├── metrics.go          # メトリクス定義
├── formatter.go        # フォーマッター
├── calculator_test.go  # 統計計算テスト
├── report_test.go      # レポート生成テスト
├── metrics_test.go     # メトリクステスト
└── formatter_test.go   # フォーマッターテスト
```

## 4. アーキテクチャ

### 4.1 Calculator（統計計算エンジン）

```go
package statistics

import (
    "fx-backtesting/pkg/models"
    "math"
    "time"
)

// Calculator は統計計算を行うメインエンジンです。
type Calculator struct {
    riskFreeRate float64 // リスクフリーレート（年率）
}

// NewCalculator は新しい統計計算エンジンを作成します。
func NewCalculator() *Calculator {
    return &Calculator{
        riskFreeRate: 0.02, // デフォルト2%
    }
}

// SetRiskFreeRate はリスクフリーレートを設定します。
func (c *Calculator) SetRiskFreeRate(rate float64) {
    c.riskFreeRate = rate
}

// CalculateMetrics は取引履歴から各種メトリクスを計算します。
func (c *Calculator) CalculateMetrics(trades []models.Trade, initialBalance float64) *Metrics {
    if len(trades) == 0 {
        return &Metrics{InitialBalance: initialBalance}
    }

    metrics := &Metrics{
        InitialBalance: initialBalance,
        TotalTrades:    len(trades),
    }

    c.calculateBasicMetrics(metrics, trades)
    c.calculateRiskMetrics(metrics, trades, initialBalance)
    c.calculateAdvancedMetrics(metrics, trades, initialBalance)

    return metrics
}

// calculateBasicMetrics は基本的なメトリクスを計算します。
func (c *Calculator) calculateBasicMetrics(metrics *Metrics, trades []models.Trade) {
    var totalPnL, grossProfit, grossLoss float64
    var winCount, lossCount int
    var largestWin, largestLoss float64

    for _, trade := range trades {
        totalPnL += trade.PnL

        if trade.IsWinning() {
            winCount++
            grossProfit += trade.PnL
            if trade.PnL > largestWin {
                largestWin = trade.PnL
            }
        } else if trade.IsLosing() {
            lossCount++
            grossLoss += math.Abs(trade.PnL)
            if trade.PnL < largestLoss {
                largestLoss = trade.PnL
            }
        }
    }

    metrics.TotalPnL = totalPnL
    metrics.FinalBalance = metrics.InitialBalance + totalPnL
    metrics.TotalReturn = (totalPnL / metrics.InitialBalance) * 100

    metrics.WinningTrades = winCount
    metrics.LosingTrades = lossCount
    metrics.WinRate = float64(winCount) / float64(len(trades)) * 100

    metrics.GrossProfit = grossProfit
    metrics.GrossLoss = grossLoss
    metrics.LargestWin = largestWin
    metrics.LargestLoss = largestLoss

    if winCount > 0 {
        metrics.AverageWin = grossProfit / float64(winCount)
    }
    if lossCount > 0 {
        metrics.AverageLoss = grossLoss / float64(lossCount)
    }
    if grossLoss > 0 {
        metrics.ProfitFactor = grossProfit / grossLoss
    }
}

// calculateRiskMetrics はリスク関連メトリクスを計算します。
func (c *Calculator) calculateRiskMetrics(metrics *Metrics, trades []models.Trade, initialBalance float64) {
    // 日次リターンの計算
    dailyReturns := c.calculateDailyReturns(trades, initialBalance)
    
    // ボラティリティの計算
    metrics.Volatility = c.calculateVolatility(dailyReturns)
    
    // シャープレシオの計算
    metrics.SharpeRatio = c.calculateSharpeRatio(dailyReturns, c.riskFreeRate)
    
    // 最大ドローダウンの計算
    maxDD, maxDDDate := c.calculateMaxDrawdown(trades, initialBalance)
    metrics.MaxDrawdown = maxDD
    metrics.MaxDrawdownDate = maxDDDate
    
    // VaR（Value at Risk）の計算
    metrics.VaR95 = c.calculateVaR(dailyReturns, 0.95)
    metrics.VaR99 = c.calculateVaR(dailyReturns, 0.99)
}

// calculateAdvancedMetrics は高度なメトリクスを計算します。
func (c *Calculator) calculateAdvancedMetrics(metrics *Metrics, trades []models.Trade, initialBalance float64) {
    // カルマーレシオの計算
    if metrics.MaxDrawdown != 0 {
        annualizedReturn := metrics.TotalReturn * (365.0 / c.calculateTradingDays(trades))
        metrics.CalmarRatio = annualizedReturn / math.Abs(metrics.MaxDrawdown)
    }
    
    // ソルティノレシオの計算
    metrics.SortinoRatio = c.calculateSortinoRatio(trades, c.riskFreeRate)
    
    // 平均取引時間の計算
    metrics.AverageTradeDuration = c.calculateAverageTradeDuration(trades)
    
    // 最大連勝・連敗の計算
    metrics.MaxConsecutiveWins, metrics.MaxConsecutiveLosses = c.calculateConsecutiveWinsLosses(trades)
}
```

### 4.2 Metrics（メトリクス定義）

```go
package statistics

import "time"

// Metrics はバックテスト結果の統計情報を保持します。
type Metrics struct {
    // 基本情報
    InitialBalance float64 `json:"initial_balance"`
    FinalBalance   float64 `json:"final_balance"`
    TotalPnL       float64 `json:"total_pnl"`
    TotalReturn    float64 `json:"total_return_percent"`
    
    // 取引統計
    TotalTrades    int     `json:"total_trades"`
    WinningTrades  int     `json:"winning_trades"`
    LosingTrades   int     `json:"losing_trades"`
    WinRate        float64 `json:"win_rate_percent"`
    
    // 損益統計
    GrossProfit  float64 `json:"gross_profit"`
    GrossLoss    float64 `json:"gross_loss"`
    AverageWin   float64 `json:"average_win"`
    AverageLoss  float64 `json:"average_loss"`
    LargestWin   float64 `json:"largest_win"`
    LargestLoss  float64 `json:"largest_loss"`
    ProfitFactor float64 `json:"profit_factor"`
    
    // リスク指標
    MaxDrawdown      float64   `json:"max_drawdown_percent"`
    MaxDrawdownDate  time.Time `json:"max_drawdown_date"`
    Volatility       float64   `json:"volatility_percent"`
    SharpeRatio      float64   `json:"sharpe_ratio"`
    SortinoRatio     float64   `json:"sortino_ratio"`
    CalmarRatio      float64   `json:"calmar_ratio"`
    
    // VaR（Value at Risk）
    VaR95 float64 `json:"var_95_percent"`
    VaR99 float64 `json:"var_99_percent"`
    
    // 取引パターン
    AverageTradeDuration  time.Duration `json:"average_trade_duration"`
    MaxConsecutiveWins    int           `json:"max_consecutive_wins"`
    MaxConsecutiveLosses  int           `json:"max_consecutive_losses"`
    
    // 詳細データ
    DailyReturns    []DailyReturn    `json:"daily_returns,omitempty"`
    MonthlyReturns  []MonthlyReturn  `json:"monthly_returns,omitempty"`
    DrawdownPeriods []DrawdownPeriod `json:"drawdown_periods,omitempty"`
}

// DailyReturn は日次リターンを表します。
type DailyReturn struct {
    Date            time.Time `json:"date"`
    Return          float64   `json:"return_percent"`
    CumulativeReturn float64   `json:"cumulative_return_percent"`
    Balance         float64   `json:"balance"`
}

// MonthlyReturn は月次リターンを表します。
type MonthlyReturn struct {
    Year   int     `json:"year"`
    Month  int     `json:"month"`
    Return float64 `json:"return_percent"`
    Trades int     `json:"trades"`
}

// DrawdownPeriod はドローダウン期間を表します。
type DrawdownPeriod struct {
    StartDate    time.Time `json:"start_date"`
    EndDate      time.Time `json:"end_date"`
    Duration     time.Duration `json:"duration"`
    MaxDrawdown  float64   `json:"max_drawdown_percent"`
    Recovery     bool      `json:"recovered"`
}

// GetRiskAdjustedReturn はリスク調整後リターンを返します。
func (m *Metrics) GetRiskAdjustedReturn() float64 {
    if m.Volatility == 0 {
        return 0
    }
    return m.TotalReturn / m.Volatility
}

// GetExpectancy は期待値を返します。
func (m *Metrics) GetExpectancy() float64 {
    if m.TotalTrades == 0 {
        return 0
    }
    return m.TotalPnL / float64(m.TotalTrades)
}

// IsOverallProfitable は全体的に利益が出ているかを判定します。
func (m *Metrics) IsOverallProfitable() bool {
    return m.TotalPnL > 0
}
```

### 4.3 Report（レポート生成）

```go
package statistics

import (
    "fmt"
    "strings"
    "time"
)

// ReportGenerator はレポート生成を担当します。
type ReportGenerator struct {
    formatter Formatter
}

// NewReportGenerator は新しいレポート生成器を作成します。
func NewReportGenerator() *ReportGenerator {
    return &ReportGenerator{
        formatter: NewFormatter(),
    }
}

// GenerateTextReport はテキスト形式のレポートを生成します。
func (rg *ReportGenerator) GenerateTextReport(metrics *Metrics) string {
    var report strings.Builder
    
    report.WriteString("📊 BACKTEST PERFORMANCE REPORT\n")
    report.WriteString("═══════════════════════════════════════════════════════════════\n\n")
    
    // 基本統計
    report.WriteString("📈 BASIC STATISTICS\n")
    report.WriteString("───────────────────────────────────────────────────────────────\n")
    report.WriteString(fmt.Sprintf("Initial Balance:      %s\n", rg.formatter.FormatCurrency(metrics.InitialBalance)))
    report.WriteString(fmt.Sprintf("Final Balance:        %s\n", rg.formatter.FormatCurrency(metrics.FinalBalance)))
    report.WriteString(fmt.Sprintf("Total P&L:            %s\n", rg.formatter.FormatPnL(metrics.TotalPnL)))
    report.WriteString(fmt.Sprintf("Total Return:         %s\n", rg.formatter.FormatPercentage(metrics.TotalReturn)))
    report.WriteString("\n")
    
    // 取引統計
    report.WriteString("🎯 TRADE STATISTICS\n")
    report.WriteString("───────────────────────────────────────────────────────────────\n")
    report.WriteString(fmt.Sprintf("Total Trades:         %d\n", metrics.TotalTrades))
    report.WriteString(fmt.Sprintf("Winning Trades:       %d\n", metrics.WinningTrades))
    report.WriteString(fmt.Sprintf("Losing Trades:        %d\n", metrics.LosingTrades))
    report.WriteString(fmt.Sprintf("Win Rate:             %s\n", rg.formatter.FormatPercentage(metrics.WinRate)))
    report.WriteString(fmt.Sprintf("Profit Factor:        %.2f\n", metrics.ProfitFactor))
    report.WriteString("\n")
    
    // 損益分析
    report.WriteString("💰 PROFIT & LOSS ANALYSIS\n")
    report.WriteString("───────────────────────────────────────────────────────────────\n")
    report.WriteString(fmt.Sprintf("Gross Profit:         %s\n", rg.formatter.FormatCurrency(metrics.GrossProfit)))
    report.WriteString(fmt.Sprintf("Gross Loss:           %s\n", rg.formatter.FormatCurrency(-metrics.GrossLoss)))
    report.WriteString(fmt.Sprintf("Average Win:          %s\n", rg.formatter.FormatCurrency(metrics.AverageWin)))
    report.WriteString(fmt.Sprintf("Average Loss:         %s\n", rg.formatter.FormatCurrency(-metrics.AverageLoss)))
    report.WriteString(fmt.Sprintf("Largest Win:          %s\n", rg.formatter.FormatCurrency(metrics.LargestWin)))
    report.WriteString(fmt.Sprintf("Largest Loss:         %s\n", rg.formatter.FormatCurrency(metrics.LargestLoss)))
    report.WriteString("\n")
    
    // リスク指標
    report.WriteString("⚠️  RISK METRICS\n")
    report.WriteString("───────────────────────────────────────────────────────────────\n")
    report.WriteString(fmt.Sprintf("Maximum Drawdown:     %s\n", rg.formatter.FormatPercentage(metrics.MaxDrawdown)))
    if !metrics.MaxDrawdownDate.IsZero() {
        report.WriteString(fmt.Sprintf("Max DD Date:          %s\n", metrics.MaxDrawdownDate.Format("2006-01-02")))
    }
    report.WriteString(fmt.Sprintf("Volatility:           %s\n", rg.formatter.FormatPercentage(metrics.Volatility)))
    report.WriteString(fmt.Sprintf("Sharpe Ratio:         %.3f\n", metrics.SharpeRatio))
    report.WriteString(fmt.Sprintf("Sortino Ratio:        %.3f\n", metrics.SortinoRatio))
    report.WriteString(fmt.Sprintf("Calmar Ratio:         %.3f\n", metrics.CalmarRatio))
    report.WriteString("\n")
    
    // パフォーマンス評価
    report.WriteString(rg.generatePerformanceAssessment(metrics))
    
    return report.String()
}

// GenerateJSONReport はJSON形式のレポートを生成します。
func (rg *ReportGenerator) GenerateJSONReport(metrics *Metrics) (string, error) {
    jsonBytes, err := json.MarshalIndent(metrics, "", "  ")
    if err != nil {
        return "", fmt.Errorf("failed to generate JSON report: %w", err)
    }
    return string(jsonBytes), nil
}

// GenerateCSVReport はCSV形式のレポートを生成します。
func (rg *ReportGenerator) GenerateCSVReport(metrics *Metrics) string {
    var csv strings.Builder
    
    // ヘッダー
    csv.WriteString("Metric,Value\n")
    
    // データ行
    csv.WriteString(fmt.Sprintf("Initial Balance,%.2f\n", metrics.InitialBalance))
    csv.WriteString(fmt.Sprintf("Final Balance,%.2f\n", metrics.FinalBalance))
    csv.WriteString(fmt.Sprintf("Total P&L,%.2f\n", metrics.TotalPnL))
    csv.WriteString(fmt.Sprintf("Total Return %%,%.2f\n", metrics.TotalReturn))
    csv.WriteString(fmt.Sprintf("Total Trades,%d\n", metrics.TotalTrades))
    csv.WriteString(fmt.Sprintf("Win Rate %%,%.2f\n", metrics.WinRate))
    csv.WriteString(fmt.Sprintf("Profit Factor,%.3f\n", metrics.ProfitFactor))
    csv.WriteString(fmt.Sprintf("Max Drawdown %%,%.2f\n", metrics.MaxDrawdown))
    csv.WriteString(fmt.Sprintf("Sharpe Ratio,%.3f\n", metrics.SharpeRatio))
    csv.WriteString(fmt.Sprintf("Volatility %%,%.2f\n", metrics.Volatility))
    
    return csv.String()
}

// generatePerformanceAssessment はパフォーマンス評価を生成します。
func (rg *ReportGenerator) generatePerformanceAssessment(metrics *Metrics) string {
    var assessment strings.Builder
    
    assessment.WriteString("🎯 PERFORMANCE ASSESSMENT\n")
    assessment.WriteString("───────────────────────────────────────────────────────────────\n")
    
    // 収益性評価
    if metrics.TotalReturn > 20 {
        assessment.WriteString("✅ Excellent profitability\n")
    } else if metrics.TotalReturn > 10 {
        assessment.WriteString("✅ Good profitability\n")
    } else if metrics.TotalReturn > 0 {
        assessment.WriteString("⚠️  Modest profitability\n")
    } else {
        assessment.WriteString("❌ Strategy is unprofitable\n")
    }
    
    // 勝率評価
    if metrics.WinRate > 60 {
        assessment.WriteString("✅ High win rate\n")
    } else if metrics.WinRate > 50 {
        assessment.WriteString("✅ Good win rate\n")
    } else if metrics.WinRate > 40 {
        assessment.WriteString("⚠️  Moderate win rate\n")
    } else {
        assessment.WriteString("❌ Low win rate\n")
    }
    
    // リスク評価
    if metrics.MaxDrawdown < 5 {
        assessment.WriteString("✅ Low risk (max drawdown < 5%)\n")
    } else if metrics.MaxDrawdown < 10 {
        assessment.WriteString("⚠️  Moderate risk (max drawdown < 10%)\n")
    } else if metrics.MaxDrawdown < 20 {
        assessment.WriteString("⚠️  High risk (max drawdown < 20%)\n")
    } else {
        assessment.WriteString("❌ Very high risk (max drawdown > 20%)\n")
    }
    
    // シャープレシオ評価
    if metrics.SharpeRatio > 2.0 {
        assessment.WriteString("✅ Excellent risk-adjusted returns\n")
    } else if metrics.SharpeRatio > 1.0 {
        assessment.WriteString("✅ Good risk-adjusted returns\n")
    } else if metrics.SharpeRatio > 0.5 {
        assessment.WriteString("⚠️  Moderate risk-adjusted returns\n")
    } else {
        assessment.WriteString("❌ Poor risk-adjusted returns\n")
    }
    
    assessment.WriteString("\n")
    return assessment.String()
}
```

### 4.4 Formatter（フォーマッター）

```go
package statistics

import (
    "fmt"
    "strings"
)

// Formatter は数値フォーマットを担当します。
type Formatter struct {
    currencySymbol string
    decimalPlaces  int
}

// NewFormatter は新しいフォーマッターを作成します。
func NewFormatter() *Formatter {
    return &Formatter{
        currencySymbol: "$",
        decimalPlaces:  2,
    }
}

// FormatCurrency は通貨形式でフォーマットします。
func (f *Formatter) FormatCurrency(value float64) string {
    return fmt.Sprintf("%s%,.2f", f.currencySymbol, value)
}

// FormatPercentage はパーセント形式でフォーマットします。
func (f *Formatter) FormatPercentage(value float64) string {
    return fmt.Sprintf("%.2f%%", value)
}

// FormatPnL は損益をカラー付きでフォーマットします。
func (f *Formatter) FormatPnL(value float64) string {
    if value > 0 {
        return fmt.Sprintf("✅ +%s", f.FormatCurrency(value))
    } else if value < 0 {
        return fmt.Sprintf("❌ %s", f.FormatCurrency(value))
    }
    return fmt.Sprintf("➖ %s", f.FormatCurrency(value))
}

// FormatRatio は比率をフォーマットします。
func (f *Formatter) FormatRatio(value float64, precision int) string {
    return fmt.Sprintf("%.*f", precision, value)
}

// SetCurrencySymbol は通貨記号を設定します。
func (f *Formatter) SetCurrencySymbol(symbol string) {
    f.currencySymbol = symbol
}
```

## 5. 高度な統計計算

### 5.1 リスク指標計算

```go
// calculateSharpeRatio はシャープレシオを計算します。
func (c *Calculator) calculateSharpeRatio(dailyReturns []float64, riskFreeRate float64) float64 {
    if len(dailyReturns) < 2 {
        return 0
    }
    
    // 平均リターン計算
    var sum float64
    for _, ret := range dailyReturns {
        sum += ret
    }
    avgReturn := sum / float64(len(dailyReturns))
    
    // 年率換算
    annualizedReturn := avgReturn * 252 // 営業日ベース
    dailyRiskFree := riskFreeRate / 252
    
    // 標準偏差計算
    var variance float64
    for _, ret := range dailyReturns {
        variance += math.Pow(ret-avgReturn, 2)
    }
    variance = variance / float64(len(dailyReturns)-1)
    volatility := math.Sqrt(variance) * math.Sqrt(252) // 年率換算
    
    if volatility == 0 {
        return 0
    }
    
    return (annualizedReturn - riskFreeRate) / volatility
}

// calculateMaxDrawdown は最大ドローダウンを計算します。
func (c *Calculator) calculateMaxDrawdown(trades []models.Trade, initialBalance float64) (float64, time.Time) {
    if len(trades) == 0 {
        return 0, time.Time{}
    }
    
    var maxDrawdown float64
    var maxDrawdownDate time.Time
    var peak float64 = initialBalance
    balance := initialBalance
    
    for _, trade := range trades {
        balance += trade.PnL
        
        if balance > peak {
            peak = balance
        }
        
        drawdown := (peak - balance) / peak * 100
        if drawdown > maxDrawdown {
            maxDrawdown = drawdown
            maxDrawdownDate = trade.CloseTime
        }
    }
    
    return maxDrawdown, maxDrawdownDate
}

// calculateVaR はValue at Riskを計算します。
func (c *Calculator) calculateVaR(dailyReturns []float64, confidence float64) float64 {
    if len(dailyReturns) == 0 {
        return 0
    }
    
    // リターンをソート
    sorted := make([]float64, len(dailyReturns))
    copy(sorted, dailyReturns)
    sort.Float64s(sorted)
    
    // 信頼水準に対応するパーセンタイルを取得
    index := int((1.0 - confidence) * float64(len(sorted)))
    if index >= len(sorted) {
        index = len(sorted) - 1
    }
    
    return math.Abs(sorted[index]) * 100 // パーセント表示
}
```

### 5.2 時系列分析

```go
// calculateDailyReturns は日次リターンを計算します。
func (c *Calculator) calculateDailyReturns(trades []models.Trade, initialBalance float64) []float64 {
    if len(trades) == 0 {
        return []float64{}
    }
    
    // 日付別に取引をグループ化
    dailyPnL := make(map[string]float64)
    
    for _, trade := range trades {
        date := trade.CloseTime.Format("2006-01-02")
        dailyPnL[date] += trade.PnL
    }
    
    // 日次リターンを計算
    var returns []float64
    balance := initialBalance
    
    for _, pnl := range dailyPnL {
        if balance > 0 {
            returns = append(returns, pnl/balance)
        }
        balance += pnl
    }
    
    return returns
}

// calculateMonthlyReturns は月次リターンを計算します。
func (c *Calculator) calculateMonthlyReturns(trades []models.Trade, initialBalance float64) []MonthlyReturn {
    if len(trades) == 0 {
        return []MonthlyReturn{}
    }
    
    // 月別に取引をグループ化
    monthlyData := make(map[string]*MonthlyReturn)
    
    for _, trade := range trades {
        key := fmt.Sprintf("%d-%02d", trade.CloseTime.Year(), trade.CloseTime.Month())
        
        if monthlyData[key] == nil {
            monthlyData[key] = &MonthlyReturn{
                Year:  trade.CloseTime.Year(),
                Month: int(trade.CloseTime.Month()),
            }
        }
        
        monthlyData[key].Return += trade.PnL
        monthlyData[key].Trades++
    }
    
    // パーセント換算
    balance := initialBalance
    var results []MonthlyReturn
    
    for _, data := range monthlyData {
        if balance > 0 {
            data.Return = (data.Return / balance) * 100
        }
        balance += data.Return
        results = append(results, *data)
    }
    
    return results
}
```

## 6. レポート比較機能

### 6.1 複数戦略比較

```go
// ComparisonReport は複数の戦略を比較するレポートです。
type ComparisonReport struct {
    Strategies []StrategyResult `json:"strategies"`
    Summary    ComparisonSummary `json:"summary"`
}

// StrategyResult は戦略の結果を表します。
type StrategyResult struct {
    Name    string   `json:"name"`
    Metrics *Metrics `json:"metrics"`
}

// ComparisonSummary は比較サマリーを表します。
type ComparisonSummary struct {
    BestReturn      string  `json:"best_return_strategy"`
    BestSharpe      string  `json:"best_sharpe_strategy"`
    LowestDrawdown  string  `json:"lowest_drawdown_strategy"`
    BestWinRate     string  `json:"best_win_rate_strategy"`
}

// GenerateComparisonReport は比較レポートを生成します。
func (rg *ReportGenerator) GenerateComparisonReport(strategies []StrategyResult) *ComparisonReport {
    report := &ComparisonReport{
        Strategies: strategies,
        Summary:    ComparisonSummary{},
    }
    
    // 最高パフォーマンスの戦略を特定
    var bestReturn, bestSharpe, lowestDD, bestWinRate float64
    
    for _, strategy := range strategies {
        metrics := strategy.Metrics
        
        if metrics.TotalReturn > bestReturn {
            bestReturn = metrics.TotalReturn
            report.Summary.BestReturn = strategy.Name
        }
        
        if metrics.SharpeRatio > bestSharpe {
            bestSharpe = metrics.SharpeRatio
            report.Summary.BestSharpe = strategy.Name
        }
        
        if lowestDD == 0 || metrics.MaxDrawdown < lowestDD {
            lowestDD = metrics.MaxDrawdown
            report.Summary.LowestDrawdown = strategy.Name
        }
        
        if metrics.WinRate > bestWinRate {
            bestWinRate = metrics.WinRate
            report.Summary.BestWinRate = strategy.Name
        }
    }
    
    return report
}
```

## 7. テスト項目

### 7.1 統計計算テスト

#### 正常系テスト
- **基本統計**
  - PnL計算の正確性
  - 勝率計算の正確性
  - プロフィットファクター計算

- **リスク指標**
  - シャープレシオ計算
  - 最大ドローダウン計算
  - VaR計算の正確性

- **時系列分析**
  - 日次リターン計算
  - 月次リターン計算
  - ボラティリティ計算

#### 異常系テスト
- **エッジケース**
  - 空の取引履歴
  - 単一取引
  - 全勝・全敗ケース

#### 境界値テスト
- **極端な値**
  - 非常に大きな損益
  - ゼロリターン
  - 負のリターン

### 7.2 レポート生成テスト

#### フォーマットテスト
- **テキストレポート**
  - フォーマットの正確性
  - 特殊文字の処理
  - レイアウトの確認

- **JSONレポート**
  - JSON構造の妥当性
  - エスケープ処理
  - データ型の正確性

- **CSVレポート**
  - CSV形式の正確性
  - 区切り文字の処理
  - ヘッダー行の確認

### 7.3 テスト実行方法

```bash
# 全テスト実行
go test ./pkg/statistics/... -v

# カバレッジ確認
go test -cover ./pkg/statistics/...

# ベンチマークテスト
go test -bench . ./pkg/statistics/...

# 統計計算の精度テスト
go test -run TestCalculator_Precision ./pkg/statistics/
```

## 8. パフォーマンス考慮事項

### 8.1 大量データ処理

```go
// 大量取引履歴の効率的処理
func (c *Calculator) ProcessLargeDataset(trades []models.Trade, batchSize int) *Metrics {
    totalBatches := (len(trades) + batchSize - 1) / batchSize
    
    // バッチ処理で統計を計算
    var aggregatedMetrics *Metrics
    
    for i := 0; i < totalBatches; i++ {
        start := i * batchSize
        end := start + batchSize
        if end > len(trades) {
            end = len(trades)
        }
        
        batchMetrics := c.CalculateMetrics(trades[start:end], 0)
        aggregatedMetrics = c.aggregateMetrics(aggregatedMetrics, batchMetrics)
    }
    
    return aggregatedMetrics
}
```

### 8.2 メモリ最適化

```go
// ストリーミング統計計算
type StreamingCalculator struct {
    runningSum    float64
    runningSquareSum float64
    count         int
    min           float64
    max           float64
}

func (sc *StreamingCalculator) AddValue(value float64) {
    sc.count++
    sc.runningSum += value
    sc.runningSquareSum += value * value
    
    if sc.count == 1 || value < sc.min {
        sc.min = value
    }
    if sc.count == 1 || value > sc.max {
        sc.max = value
    }
}

func (sc *StreamingCalculator) GetStatistics() (mean, variance float64) {
    if sc.count == 0 {
        return 0, 0
    }
    
    mean = sc.runningSum / float64(sc.count)
    if sc.count > 1 {
        variance = (sc.runningSquareSum - sc.runningSum*mean) / float64(sc.count-1)
    }
    return mean, variance
}
```

この設計により、包括的で高精度な統計計算とレポート生成機能を提供し、バックテスト結果の詳細な分析を可能にします。