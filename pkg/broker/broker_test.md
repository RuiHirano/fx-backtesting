# Broker テストドキュメント

## 概要
- **対象コンポーネント**: `pkg/broker/broker.go` の SimpleBroker 実装
- **テスト目的**: 高度な注文管理機能を持つFXブローカーシミュレーションの検証
- **テスト対象機能**: 
  - 成行・指値・逆指値注文の実行
  - 保留注文の管理と自動約定
  - ポジション管理と決済
  - 注文キャンセル機能
  - エラーハンドリング
  - 統合シナリオ

## テスト構成

### 主要テストケース
1. **TestBroker_PlaceMarketOrder** - 成行注文テスト
2. **TestBroker_PlaceLimitOrder** - 指値注文テスト
3. **TestBroker_PlaceStopOrder** - 逆指値注文テスト
4. **TestBroker_CancelOrder** - 注文キャンセルテスト
5. **TestBroker_GetPendingOrders** - 保留注文取得テスト
6. **TestBroker_ProcessPendingOrders** - 保留注文処理テスト
7. **TestBroker_ClosePosition** - ポジション決済テスト
8. **TestBroker_UpdatePositions** - ポジション更新テスト
9. **TestBroker_Integration** - 統合テスト
10. **TestBroker_ErrorHandling** - エラーハンドリングテスト
11. **TestBroker_Performance** - パフォーマンステスト

## 詳細テスト仕様

### TestBroker_PlaceMarketOrder
```go
func TestBroker_PlaceMarketOrder(t *testing.T) {
    t.Run("should execute market buy order immediately", func(t *testing.T) {
        order := models.NewMarketOrder("order-1", "EURUSD", models.Buy, 10000.0)
        err := broker.PlaceOrder(order)
        // 検証: 即座に約定、ポジション作成、残高更新
    })
    
    t.Run("should execute market sell order immediately", func(t *testing.T) {
        order := models.NewMarketOrder("order-2", "EURUSD", models.Sell, 5000.0)
        err := broker.PlaceOrder(order)
        // 検証: 売りポジション作成
    })
    
    t.Run("should return error for insufficient balance", func(t *testing.T) {
        order := models.NewMarketOrder("order-3", "EURUSD", models.Buy, 1000000.0)
        err := broker.PlaceOrder(order)
        // 検証: 証拠金不足エラー
    })
}
```

**テスト目的**: 成行注文の即座約定機能を検証
**検証項目**:
- 買い・売り成行注文の即座実行
- スプレッド適用による適正な約定価格（Ask/Bid）
- レバレッジ1:100による証拠金計算
- 残高不足時の適切なエラーハンドリング
- 注文状態の正しい更新（Pending → Executed）

### TestBroker_PlaceLimitOrder
```go
func TestBroker_PlaceLimitOrder(t *testing.T) {
    t.Run("should add limit order to pending orders", func(t *testing.T) {
        currentPrice := mkt.GetCurrentPrice("EURUSD")
        order := models.NewLimitOrder("limit-1", "EURUSD", models.Buy, 10000.0, currentPrice-0.0010)
        err := broker.PlaceOrder(order)
        // 検証: 保留注文リストに追加、ポジション未作成
    })
    
    t.Run("should execute limit order when price condition is met", func(t *testing.T) {
        order := models.NewLimitOrder("limit-2", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
        err := broker.PlaceOrder(order)
        broker.ProcessPendingOrders()
        // 検証: 価格条件満足時の自動約定
    })
}
```

**テスト目的**: 指値注文の条件付き約定機能を検証
**検証項目**:
- 約定条件を満たさない場合の保留注文追加
- 約定条件（買い指値: 現在価格 ≤ 指値価格、売り指値: 現在価格 ≥ 指値価格）
- 条件満足時の自動約定とポジション作成
- 保留注文リストからの自動削除

### TestBroker_PlaceStopOrder
```go
func TestBroker_PlaceStopOrder(t *testing.T) {
    t.Run("should add stop order to pending orders", func(t *testing.T) {
        order := models.NewStopOrder("stop-1", "EURUSD", models.Buy, 10000.0, currentPrice+0.0020)
        err := broker.PlaceOrder(order)
        // 検証: 保留注文への追加
    })
    
    t.Run("should execute stop order when price condition is met", func(t *testing.T) {
        order := models.NewStopOrder("stop-2", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
        err := broker.PlaceOrder(order)
        broker.ProcessPendingOrders()
        // 検証: トリガー条件満足時の約定
    })
}
```

**テスト目的**: 逆指値注文のトリガー機能を検証
**検証項目**:
- トリガー条件を満たさない場合の保留注文追加
- トリガー条件（買い逆指値: 現在価格 ≥ 逆指値価格、売り逆指値: 現在価格 ≤ 逆指値価格）
- 条件満足時の成行約定実行
- 損切り・ブレイクアウト戦略への対応

### TestBroker_CancelOrder
```go
func TestBroker_CancelOrder(t *testing.T) {
    t.Run("should cancel pending order successfully", func(t *testing.T) {
        order := models.NewLimitOrder("cancel-1", "EURUSD", models.Buy, 10000.0, currentPrice-0.0020)
        broker.PlaceOrder(order)
        err := broker.CancelOrder(order.ID)
        // 検証: 注文キャンセルと保留リストからの削除
    })
    
    t.Run("should return error for non-existent order", func(t *testing.T) {
        err := broker.CancelOrder("non-existent-order")
        // 検証: 存在しない注文IDのエラーハンドリング
    })
    
    t.Run("should return error for already executed order", func(t *testing.T) {
        order := models.NewMarketOrder("cancel-2", "EURUSD", models.Buy, 5000.0)
        broker.PlaceOrder(order)
        err := broker.CancelOrder(order.ID)
        // 検証: 約定済み注文キャンセル時のエラー
    })
}
```

**テスト目的**: 注文キャンセル機能の検証
**検証項目**:
- 保留注文の正常なキャンセル
- 注文状態の適切な更新（Pending → Cancelled）
- 存在しない注文IDに対するエラーハンドリング
- 約定済み注文キャンセル試行時のエラー

### TestBroker_GetPendingOrders
```go
func TestBroker_GetPendingOrders(t *testing.T) {
    t.Run("should return all pending orders", func(t *testing.T) {
        // 複数の保留注文を作成
        order1 := models.NewLimitOrder("pending-1", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
        order2 := models.NewLimitOrder("pending-2", "EURUSD", models.Sell, 7000.0, currentPrice+0.0015)
        order3 := models.NewStopOrder("pending-3", "EURUSD", models.Buy, 3000.0, currentPrice+0.0020)
        // 検証: 全保留注文の正確な取得
    })
    
    t.Run("should return empty slice when no pending orders", func(t *testing.T) {
        pendingOrders := newBroker.GetPendingOrders()
        // 検証: 保留注文がない場合の空スライス返却
    })
}
```

**テスト目的**: 保留注文取得機能の検証
**検証項目**:
- 全保留注文の正確な取得
- 注文の種類（指値・逆指値）に関わらない一括取得
- 保留注文がない場合の適切な空スライス返却

### TestBroker_ProcessPendingOrders
```go
func TestBroker_ProcessPendingOrders(t *testing.T) {
    t.Run("should process limit orders correctly", func(t *testing.T) {
        // 異なる価格条件の指値注文を作成
        buyLimitBelow := models.NewLimitOrder("limit-buy-below", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
        buyLimitAbove := models.NewLimitOrder("limit-buy-above", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
        // 検証: 約定条件に基づく選択的実行
    })
    
    t.Run("should process stop orders correctly", func(t *testing.T) {
        // 異なるトリガー条件の逆指値注文を作成
        buyStopBelow := models.NewStopOrder("stop-buy-below", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
        buyStopAbove := models.NewStopOrder("stop-buy-above", "EURUSD", models.Buy, 5000.0, currentPrice+0.0010)
        // 検証: トリガー条件に基づく選択的実行
    })
}
```

**テスト目的**: 保留注文の自動処理ロジックを検証
**検証項目**:
- 指値注文の約定条件判定ロジック
- 逆指値注文のトリガー条件判定ロジック
- 条件を満たした注文のみの選択的約定
- 約定済み注文の保留リストからの自動削除

### TestBroker_ClosePosition
```go
func TestBroker_ClosePosition(t *testing.T) {
    t.Run("should close position successfully", func(t *testing.T) {
        order := models.NewMarketOrder("close-1", "EURUSD", models.Buy, 10000.0)
        broker.PlaceOrder(order)
        positions := broker.GetPositions()
        positionID := positions[0].ID
        err := broker.ClosePosition(positionID)
        // 検証: ポジション決済、残高更新、取引履歴追加
    })
    
    t.Run("should return error for non-existent position", func(t *testing.T) {
        err := broker.ClosePosition("non-existent-position")
        // 検証: 存在しないポジションIDのエラーハンドリング
    })
}
```

**テスト目的**: ポジション決済機能の検証
**検証項目**:
- 正常なポジション決済処理
- スプレッドを考慮した決済価格計算
- 損益計算と残高への反映
- 取引履歴への記録
- 存在しないポジションIDのエラーハンドリング

### TestBroker_UpdatePositions
```go
func TestBroker_UpdatePositions(t *testing.T) {
    t.Run("should update position prices and process pending orders", func(t *testing.T) {
        order := models.NewMarketOrder("update-1", "EURUSD", models.Buy, 10000.0)
        broker.PlaceOrder(order)
        limitOrder := models.NewLimitOrder("update-limit", "EURUSD", models.Buy, 5000.0, currentPrice-0.0010)
        broker.PlaceOrder(limitOrder)
        mkt.Forward()
        broker.UpdatePositions()
        // 検証: ポジション価格更新と保留注文の同時処理
    })
}
```

**テスト目的**: ポジション更新と保留注文処理の統合機能を検証
**検証項目**:
- 市場データ更新に伴うポジション価格の自動更新
- 保留注文の同時処理（ProcessPendingOrders）
- 含み損益の自動再計算

### TestBroker_Integration
```go
func TestBroker_Integration(t *testing.T) {
    t.Run("complete trading scenario", func(t *testing.T) {
        // 1. 成行注文でポジション作成
        marketOrder := models.NewMarketOrder("integration-market", "EURUSD", models.Buy, 10000.0)
        broker.PlaceOrder(marketOrder)
        
        // 2. 利食い用の指値注文
        takeProfitOrder := models.NewLimitOrder("integration-tp", "EURUSD", models.Sell, 10000.0, currentPrice+0.0050)
        broker.PlaceOrder(takeProfitOrder)
        
        // 3. 損切り用の逆指値注文
        stopLossOrder := models.NewStopOrder("integration-sl", "EURUSD", models.Sell, 10000.0, currentPrice-0.0020)
        broker.PlaceOrder(stopLossOrder)
        
        // 4. 市場更新と保留注文処理
        mkt.Forward()
        broker.UpdatePositions()
        // 検証: 完全な取引シナリオの動作
    })
}
```

**テスト目的**: 実際の取引シナリオでの統合動作を検証
**検証項目**:
- 成行・指値・逆指値注文の組み合わせ使用
- 利食い・損切り戦略の実装
- 市場データ更新との連携
- 複数注文の同時管理

### TestBroker_ErrorHandling
```go
func TestBroker_ErrorHandling(t *testing.T) {
    t.Run("should validate order data", func(t *testing.T) {
        invalidOrder := models.NewMarketOrder("invalid-1", "EURUSD", models.Buy, 0.0)
        err := broker.PlaceOrder(invalidOrder)
        // 検証: 注文データバリデーション
    })
    
    t.Run("should handle insufficient balance for pending orders", func(t *testing.T) {
        largeOrder := models.NewMarketOrder("large-order", "EURUSD", models.Buy, 90000.0)
        broker.PlaceOrder(largeOrder)
        limitOrder := models.NewLimitOrder("insufficient-limit", "EURUSD", models.Buy, 50000.0, currentPrice+0.0010)
        broker.PlaceOrder(limitOrder)
        broker.ProcessPendingOrders()
        // 検証: 保留注文約定時の証拠金不足処理
    })
}
```

**テスト目的**: エラーハンドリングの堅牢性を検証
**検証項目**:
- 注文データの包括的バリデーション
- 証拠金不足の適切な検出と処理
- 存在しない注文/ポジションIDの処理
- システムの安定性と整合性維持

### TestBroker_Performance
```go
func TestBroker_Performance(t *testing.T) {
    t.Run("should handle multiple orders efficiently", func(t *testing.T) {
        start := time.Now()
        // 100個の保留注文を作成
        for i := 0; i < 100; i++ {
            limitOrder := models.NewLimitOrder(...)
            broker.PlaceOrder(limitOrder)
        }
        broker.ProcessPendingOrders()
        elapsed := time.Since(start)
        // 検証: 処理時間が1秒以内
    })
}
```

**テスト目的**: 大量注文処理時のパフォーマンスを検証
**検証項目**:
- 100注文の処理時間（目標: 1秒以内）
- メモリ使用量の効率性
- スケーラビリティの確認

## テスト環境とデータ

### テストヘルパー関数
```go
func createTestBroker(t *testing.T) (Broker, market.Market) {
    // MarketとBrokerの標準的なテスト環境を構築
    config := models.DataProviderConfig{
        FilePath: "./testdata/sample.csv",
        Format:   "csv",
    }
    
    brokerConfig := models.BrokerConfig{
        InitialBalance: 10000.0,
        Spread:         0.0001, // 1 pip
    }
    
    return broker, market
}
```

### テストデータ（sample.csv）
```csv
timestamp,open,high,low,close,volume
2025-01-01 00:00:00,1.1000,1.1010,1.0990,1.1005,1000
2025-01-01 00:01:00,1.1005,1.1015,1.0995,1.1010,1200
2025-01-01 00:02:00,1.1010,1.1020,1.1000,1.1015,1100
2025-01-01 00:03:00,1.1015,1.1025,1.1005,1.1020,1300
2025-01-01 00:04:00,1.1020,1.1030,1.1010,1.1025,1400
2025-01-01 00:05:00,1.1025,1.1035,1.1015,1.1030,1500
```

**データ特性**:
- 通貨ペア: EURUSD
- 価格範囲: 1.1005 - 1.1030
- 時間範囲: 6分間のデータ
- 上昇トレンド（テスト用）

## 重要な検証ポイント

### 1. 注文種別による処理の違い
- **成行注文**: 即座約定、スプレッド適用、残高チェック
- **指値注文**: 保留登録、条件チェック、自動約定
- **逆指値注文**: 保留登録、トリガーチェック、成行約定

### 2. 価格計算の正確性
- **買い成行**: Ask価格（現在価格 + スプレッド）
- **売り成行**: Bid価格（現在価格 - スプレッド）
- **買い指値**: 現在価格 ≤ 指値価格で約定
- **売り指値**: 現在価格 ≥ 指値価格で約定
- **買い逆指値**: 現在価格 ≥ 逆指値価格で約定
- **売り逆指値**: 現在価格 ≤ 逆指値価格で約定

### 3. 証拠金管理
- **レバレッジ**: 1:100（必要証拠金 = ポジションサイズ × 約定価格 / 100）
- **残高チェック**: 成行注文時と保留注文約定時
- **証拠金返却**: ポジション決済時

### 4. 状態管理の整合性
- **注文状態**: Pending → Executed/Cancelled
- **ポジション管理**: 作成 → 更新 → 決済
- **保留注文**: 追加 → 処理 → 削除/約定
- **取引履歴**: 決済時の自動記録

## テスト実行方法

### 基本実行
```bash
# 全テスト実行
go test ./pkg/broker -v

# カバレッジ測定
go test ./pkg/broker -cover

# ベンチマーク実行
go test ./pkg/broker -bench=.
```

### 個別テスト実行
```bash
# 成行注文テスト
go test -run TestBroker_PlaceMarketOrder ./pkg/broker/

# 指値注文テスト
go test -run TestBroker_PlaceLimitOrder ./pkg/broker/

# 逆指値注文テスト
go test -run TestBroker_PlaceStopOrder ./pkg/broker/

# 注文キャンセルテスト
go test -run TestBroker_CancelOrder ./pkg/broker/

# 統合テスト
go test -run TestBroker_Integration ./pkg/broker/

# エラーハンドリングテスト
go test -run TestBroker_ErrorHandling ./pkg/broker/

# パフォーマンステスト
go test -run TestBroker_Performance ./pkg/broker/
```

### デバッグ実行
```bash
# 詳細ログ付き実行
go test ./pkg/broker -v -args -test.v

# 特定のサブテスト実行
go test -run "TestBroker_PlaceMarketOrder/should_execute_market_buy_order_immediately" ./pkg/broker/
```

## 期待される結果

### 成功条件
- 全テストケースがパス
- カバレッジ85%以上
- パフォーマンステストが1秒以内で完了
- メモリリークなし

### パフォーマンス目標
- **注文実行**: 1秒間に10,000回以上
- **保留注文処理**: 1秒間に5,000注文以上
- **メモリ使用量**: 1,000ポジション + 500保留注文で12MB以下

## 依存関係
- **Market**: pkg/market のMarket実装（価格取得、時間進行）
- **Models**: pkg/models の Order、Position、Trade、BrokerConfig
- **Data Provider**: CSV形式のテストデータ
- **Testing Framework**: testify/assert

## 注意事項
1. **市場データ依存**: テスト結果は sample.csv の価格データに依存
2. **時間進行**: Market.Forward() による時間進行がテストに影響
3. **浮動小数点**: 価格計算での浮動小数点精度に注意
4. **並行処理**: 現在の実装は単一ゴルーチン前提
5. **リソース管理**: テスト後のリソース解放は不要（GCが処理）

## トラブルシューティング

### よくある問題
1. **"index out of range"**: Market初期化エラー、sample.csvの確認
2. **"insufficient balance"**: 初期残高設定の確認
3. **約定しない保留注文**: 価格条件の設定ミス
4. **テストタイムアウト**: パフォーマンス問題、ループの確認

### デバッグ手順
1. テストデータ（sample.csv）の存在確認
2. Market初期化の成功確認
3. 注文データのバリデーション確認
4. 価格条件の論理確認
5. メモリ使用量の監視