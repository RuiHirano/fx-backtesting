# Backtester テスト仕様書

## 概要
- **テスト対象**: `pkg/backtester/backtester.go` の Backtester 統括コンポーネント
- **テスト目的**: Market・Broker統合によるバックテスト実行とユーザーAPI提供の検証
- **テスト対象メソッド**: 
  - `TestBacktester_NewBacktester`
  - `TestBacktester_Forward`
  - `TestBacktester_BuySell`
  - `TestBacktester_PositionManagement`
  - `TestBacktester_Integration`
  - `TestBacktester_ErrorHandling`

## テスト内容

### TestBacktester_NewBacktester
```go
func TestBacktester_NewBacktester(t *testing.T) {
    // Market・Broker設定
    dataConfig := models.DataProviderConfig{
        FilePath: "./testdata/sample.csv",
        Format:   "csv",
    }
    
    brokerConfig := models.BrokerConfig{
        InitialBalance: 10000.0,
        Spread:         0.0001,
    }
    
    // Backtester作成
    backtester := NewBacktester(dataConfig, brokerConfig)
    
    // 初期化
    err := backtester.Initialize(ctx)
    
    // 初期状態確認
    balance := backtester.GetBalance()
    positions := backtester.GetPositions()
}
```
- **テスト目的**: Backtester コンストラクタとInitializeの検証
- **テスト条件**: 
  - 事前条件: DataProviderConfig・BrokerConfig設定
  - 入力値: CSV ファイルパス、初期残高10000.0、スプレッド0.0001
  - 期待結果: Backtester作成成功、初期化成功
- **検証項目**: 
  - NewBacktester()成功
  - Initialize()がエラーなく完了
  - 初期状態でIsFinished()=false
  - 初期残高が設定値と一致
  - 初期ポジション数=0
- **重要ポイント**: 
  - Market・Broker統合の確認
  - 初期状態の整合性確認
  - コンポーネント連携テスト

### TestBacktester_Forward
```go
func TestBacktester_Forward(t *testing.T) {
    backtester := createTestBacktester(t)
    err := backtester.Initialize(ctx)
    
    // 初期時刻・価格取得
    initialTime := backtester.GetCurrentTime()
    initialPrice := backtester.GetCurrentPrice("EURUSD")
    
    // 時間進行
    hasNext := backtester.Forward()
    
    // 時間・価格更新確認
    newTime := backtester.GetCurrentTime()
    newPrice := backtester.GetCurrentPrice("EURUSD")
    
    // 全データ消費まで進行
    for backtester.Forward() { }
    
    // 終了状態確認
    if !backtester.IsFinished() { ... }
}
```
- **テスト目的**: バックテストループ（時間進行）機能の検証
- **テスト条件**: 
  - 事前条件: Backtester初期化済み
  - 入力値: Forward()による時間進行
  - 期待結果: 時間・価格の順次更新、データ終了時の適切な終了
- **検証項目**: 
  - 初期時刻が非ゼロ値
  - 初期価格が正の値
  - Forward()後の時間進行確認
  - 価格データ更新確認
  - 全データ消費後のIsFinished()=true
- **重要ポイント**: 
  - Market時間進行との連携
  - ポジション価格自動更新
  - データ終了時の適切な終了判定
  - 時系列データ処理の正確性

### TestBacktester_BuySell
```go
func TestBacktester_BuySell(t *testing.T) {
    backtester := createTestBacktester(t)
    err := backtester.Initialize(ctx)
    
    initialBalance := backtester.GetBalance()
    
    // 買い注文実行
    err = backtester.Buy("EURUSD", 10000.0)
    
    // ポジション確認
    positions := backtester.GetPositions()
    if len(positions) != 1 { ... }
    
    position := positions[0]
    if position.Side != models.Buy { ... }
    if position.Size != 10000.0 { ... }
    
    // 残高変動確認
    newBalance := backtester.GetBalance()
    if newBalance >= initialBalance { ... }
    
    // 売り注文実行
    err = backtester.Sell("EURUSD", 5000.0)
    
    positions = backtester.GetPositions()
    if len(positions) != 2 { ... }
}
```
- **テスト目的**: ユーザーAPI（買い・売り注文）の検証
- **テスト条件**: 
  - 事前条件: Backtester初期化済み、市場価格取得可能
  - 入力値: Buy("EURUSD", 10000.0)、Sell("EURUSD", 5000.0)
  - 期待結果: ポジション作成、残高変動、複数ポジション管理
- **検証項目**: 
  - Buy()がエラーなく完了
  - 買いポジション1つ作成
  - ポジション詳細（Side=Buy、Size=10000.0）
  - 残高減少（証拠金差し引き）
  - Sell()実行後、ポジション2つ
- **重要ポイント**: 
  - ユーザーAPI簡単性
  - Broker連携による注文実行
  - 複数ポジション同時保持
  - 証拠金計算・残高管理

### TestBacktester_PositionManagement
```go
func TestBacktester_PositionManagement(t *testing.T) {
    backtester := createTestBacktester(t)
    err := backtester.Initialize(ctx)
    
    // 複数ポジション作成
    err = backtester.Buy("EURUSD", 10000.0)
    err = backtester.Sell("EURUSD", 8000.0)
    err = backtester.Buy("EURUSD", 12000.0)
    
    // ポジション一覧確認
    positions := backtester.GetPositions()
    if len(positions) != 3 { ... }
    
    // 特定ポジションクローズ
    positionID := positions[0].ID
    err = backtester.ClosePosition(positionID)
    
    updatedPositions := backtester.GetPositions()
    if len(updatedPositions) != 2 { ... }
    
    // 全ポジションクローズ
    err = backtester.CloseAllPositions()
    
    finalPositions := backtester.GetPositions()
    if len(finalPositions) != 0 { ... }
}
```
- **テスト目的**: ポジション管理機能の検証
- **テスト条件**: 
  - 事前条件: 複数ポジション作成（買い2、売り1）
  - 入力値: ClosePosition(ID)、CloseAllPositions()
  - 期待結果: ポジション個別・全体決済機能
- **検証項目**: 
  - 複数ポジション作成（3つ）
  - 個別ポジションクローズ成功
  - ポジション数減少確認（3→2）
  - 全ポジションクローズ成功
  - 最終ポジション数=0
- **重要ポイント**: 
  - ポジション管理の柔軟性
  - 個別・一括決済機能
  - Broker連携による損益計算
  - ポジション状態追跡

### TestBacktester_Integration
```go
func TestBacktester_Integration(t *testing.T) {
    backtester := createTestBacktester(t)
    err := backtester.Initialize(ctx)
    
    tradeCount := 0
    maxSteps := 3
    
    for step := 0; step < maxSteps && !backtester.IsFinished(); step++ {
        // 現在価格取得
        price := backtester.GetCurrentPrice("EURUSD")
        
        // ステップごとに異なる取引パターン
        switch step {
        case 0:
            err = backtester.Buy("EURUSD", 10000.0)
            tradeCount++
        case 1:
            err = backtester.Sell("EURUSD", 5000.0)
            tradeCount++
        case 2:
            positions := backtester.GetPositions()
            if len(positions) > 0 {
                err = backtester.ClosePosition(positions[0].ID)
            }
        }
        
        // 時間進行
        if step < maxSteps-1 {
            hasNext := backtester.Forward()
        }
    }
    
    // 最終状態確認
    finalPositions := backtester.GetPositions()
    finalBalance := backtester.GetBalance()
}
```
- **テスト目的**: Market・Broker統合機能の検証
- **テスト条件**: 
  - 事前条件: 時系列での複数取引実行
  - 入力値: 時間進行と取引の組み合わせ
  - 期待結果: 統合的なバックテスト実行
- **検証項目**: 
  - 各ステップで価格取得成功
  - 時間進行に合わせた取引実行
  - 複数種類の操作組み合わせ
  - 最終状態の整合性確認
  - 残高非負値確認
- **重要ポイント**: 
  - エンドツーエンド動作確認
  - Market・Broker連携の安定性
  - 時系列取引処理
  - リアルタイム価格反映

### TestBacktester_ErrorHandling
```go
func TestBacktester_ErrorHandling(t *testing.T) {
    backtester := createTestBacktester(t)
    err := backtester.Initialize(ctx)
    
    // 無効なサイズでの注文
    err = backtester.Buy("EURUSD", 0.0)
    if err == nil { t.Error("Expected error for zero size order") }
    
    err = backtester.Buy("EURUSD", -1000.0)
    if err == nil { t.Error("Expected error for negative size order") }
    
    // 存在しないシンボル
    err = backtester.Buy("INVALID", 1000.0)
    if err == nil { t.Error("Expected error for invalid symbol") }
    
    // 残高不足での大きな注文
    err = backtester.Buy("EURUSD", 10000000.0)
    if err == nil { t.Error("Expected error for insufficient balance") }
    
    // 存在しないポジションのクローズ
    err = backtester.ClosePosition("nonexistent-id")
    if err == nil { t.Error("Expected error for nonexistent position") }
    
    // 初期化前の操作エラー
    uninitializedBacktester := createTestBacktester(t)
    err = uninitializedBacktester.Buy("EURUSD", 1000.0)
    if err == nil { t.Error("Expected error for uninitialized backtester") }
}
```
- **テスト目的**: エラーハンドリングの検証
- **テスト条件**: 
  - 異常値: ゼロ・負のサイズ、無効シンボル、大きな注文、存在しないID
  - 異常状態: 初期化前操作
  - 期待結果: 適切なエラー発生と処理
- **検証項目**: 
  - ゼロサイズ注文でエラー
  - 負サイズ注文でエラー
  - 無効シンボルでエラー
  - 残高不足でエラー
  - 存在しないポジションクローズでエラー
  - 初期化前操作でエラー
- **重要ポイント**: 
  - 入力値検証の徹底
  - 状態チェックの適切性
  - エラーメッセージの明確性
  - システム安定性確保

## 結果（テスト数と実績）
- **正常系テスト数**: 5個
- **異常系テスト数**: 1個  
- **エラーハンドリング**: 6パターン
- **カバレッジ**: 78.7%

## 依存関係
- **Market連携**: pkg/market のMarket実装に依存
- **Broker連携**: pkg/broker のBroker実装に依存
- **データ提供**: pkg/data のCSVProvider機能に依存
- **モデル**: pkg/models のOrder、Position、BrokerConfig、DataProviderConfig
- **統合機能**: Market時間進行とBroker価格更新の連携

## ユーザーAPI仕様詳細
1. **NewBacktester(dataConfig, brokerConfig)**: Backtester作成
2. **Initialize(ctx)**: Market・Broker初期化
3. **Buy(symbol, size)**: 買い注文実行（MarketOrder作成・実行）
4. **Sell(symbol, size)**: 売り注文実行（MarketOrder作成・実行）
5. **Forward()**: 時間進行（Market.Forward()・Broker.UpdatePositions()）
6. **GetPositions()**: 全ポジション取得
7. **GetBalance()**: 現在残高取得
8. **GetCurrentPrice(symbol)**: 現在価格取得
9. **GetCurrentTime()**: 現在時刻取得
10. **IsFinished()**: バックテスト終了判定
11. **ClosePosition(id)**: 個別ポジション決済
12. **CloseAllPositions()**: 全ポジション決済

## 技術仕様詳細
1. **統合アーキテクチャ**: Market(価格・時間) + Broker(注文・ポジション)
2. **自動注文ID生成**: "buy-SYMBOL-timestamp"、"sell-SYMBOL-timestamp"形式
3. **価格連携**: Market価格をBrokerスプレッド適用で実行
4. **時間進行**: Market.Forward()によるデータストリーム進行
5. **ポジション更新**: Forward()時の自動価格更新
6. **エラー処理**: 初期化チェック、入力値検証、Broker連携エラー伝播
7. **状態管理**: initialized フラグによる操作制御

## テストデータ
- **sample.csv**: テスト用ローソク足データ（6行のEURUSDデータ）
  - 価格範囲: 1.1005 ～ 1.1030
  - 時間範囲: 2025-01-01 00:00:00 ～ 00:05:00
  - 価格トレンド: 上昇傾向（+25pips）

## テスト実行
```bash
# 全テスト実行
go test ./pkg/backtester/... -v

# カバレッジ確認
go test ./pkg/backtester/... -cover

# 個別テスト実行
go test -run TestBacktester_NewBacktester ./pkg/backtester/
go test -run TestBacktester_Forward ./pkg/backtester/
go test -run TestBacktester_BuySell ./pkg/backtester/
go test -run TestBacktester_PositionManagement ./pkg/backtester/
go test -run TestBacktester_Integration ./pkg/backtester/
go test -run TestBacktester_ErrorHandling ./pkg/backtester/
```