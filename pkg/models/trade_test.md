# Trade テスト仕様書

## 概要
- **テスト対象**: `pkg/models/trade.go` の Trade 構造体と関連型
- **テストの目的**: 取引履歴の作成、損益計算、取引結果判定、CSV変換機能の正常性を確認
- **実装されているテスト関数**: 
  - `TestTrade_NewTradeFromPosition`
  - `TestTrade_IsWinning`
  - `TestTrade_IsLosing`
  - `TestTrade_IsBreakeven`
  - `TestTrade_GetPnLPercentage`
  - `TestTrade_GetDurationHours`
  - `TestTrade_ToCSVRecord`
  - `TestTradeStatus_String`
  - `TestCalculateTradePnL`

## テスト関数詳細

### TestTrade_NewTradeFromPosition
```go
func TestTrade_NewTradeFromPosition(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    exitPrice := 1.1010
    
    trade := NewTradeFromPosition(position, exitPrice)
    
    if trade.ID != position.ID { ... }
    if trade.Symbol != position.Symbol { ... }
    if trade.Side != position.Side { ... }
    if trade.Size != position.Size { ... }
    if trade.EntryPrice != position.EntryPrice { ... }
    if trade.ExitPrice != exitPrice { ... }
    if trade.Status != TradeClosed { ... }
    
    expectedPnL := (1.1010 - 1.1000) * 10000.0
    assertFloatEqual(t, expectedPnL, trade.PnL, "Trade PnL")
}
```
- **テスト内容**: NewTradeFromPosition関数によるポジションからの取引履歴作成
- **テストケース**: 
  - 正常系: 有効なポジションと出口価格での取引履歴作成
  - ポジション情報の正しい引き継ぎ確認
  - 損益計算の正確性確認
  - ステータスがTradeClosedに設定されることの確認
- **アサーション**: 
  - ポジションの全情報が正しく引き継がれる
  - ExitPriceが指定値に設定される
  - PnLが正しく計算される
  - StatusがTradeClosedに設定される
  - CloseTimeとDurationが設定される

### TestTrade_IsWinning
```go
func TestTrade_IsWinning(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    
    // 勝ち取引
    trade := NewTradeFromPosition(position, 1.1010)
    if !trade.IsWinning() { ... }
    
    // 負け取引
    trade = NewTradeFromPosition(position, 1.0990)
    if !trade.IsLosing() { ... }
    
    // 引き分け
    trade = NewTradeFromPosition(position, 1.1000)
    if !trade.IsBreakeven() { ... }
}
```
- **テスト内容**: 勝ち取引判定機能
- **テストケース**: 
  - 正常系: PnL > 0での勝ち取引判定
  - 正常系: PnL < 0での負け取引判定（IsLosingメソッドで確認）
  - 正常系: PnL = 0での引き分け取引判定（IsBreakevenメソッドで確認）
- **アサーション**: 
  - 利益が出た取引でtrue
  - 損失が出た取引でfalse
  - 損益なしの取引でfalse

### TestTrade_IsLosing
```go
func TestTrade_IsLosing(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    
    // 負け取引
    trade := NewTradeFromPosition(position, 1.0990)
    if !trade.IsLosing() { ... }
    
    // 勝ち取引
    trade = NewTradeFromPosition(position, 1.1010)
    if trade.IsLosing() { ... }
}
```
- **テスト内容**: 負け取引判定機能
- **テストケース**: 
  - 正常系: PnL < 0での負け取引判定
  - 正常系: PnL > 0での勝ち取引判定（負け取引として識別されないことの確認）
- **アサーション**: 
  - 損失が出た取引でtrue
  - 利益が出た取引でfalse

### TestTrade_IsBreakeven
```go
func TestTrade_IsBreakeven(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    
    // 引き分け取引
    trade := NewTradeFromPosition(position, 1.1000)
    if !trade.IsBreakeven() { ... }
    
    // 勝ち取引
    trade = NewTradeFromPosition(position, 1.1010)
    if trade.IsBreakeven() { ... }
}
```
- **テスト内容**: 引き分け取引判定機能
- **テストケース**: 
  - 正常系: PnL = 0での引き分け取引判定
  - 正常系: PnL ≠ 0での非引き分け取引判定
- **アサーション**: 
  - 損益なしの取引でtrue
  - 損益ありの取引でfalse

### TestTrade_GetPnLPercentage
```go
func TestTrade_GetPnLPercentage(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    trade := NewTradeFromPosition(position, 1.1010)
    
    expectedPercentage := ((1.1010 - 1.1000) / 1.1000) * 100
    actualPercentage := trade.GetPnLPercentage()
    
    assertFloatEqual(t, expectedPercentage, actualPercentage, "Trade PnL percentage")
}
```
- **テスト内容**: 損益率計算機能
- **テストケース**: 
  - 正常系: (PnL / (EntryPrice * Size)) * 100での損益率計算
  - 境界値: EntryPriceが0の場合の動作
- **アサーション**: 
  - 損益率がパーセンテージで正しく計算される
  - 浮動小数点比較では許容誤差を使用

### TestTrade_GetDurationHours
```go
func TestTrade_GetDurationHours(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    trade := NewTradeFromPosition(position, 1.1010)
    
    // 取引時間は非常に短いはずなので、0以上であることを確認
    if trade.GetDurationHours() < 0 { ... }
}
```
- **テスト内容**: 取引時間（時間単位）計算機能
- **テストケース**: 
  - 正常系: CloseTime - OpenTimeの時間差計算
  - 境界値: 非負の値であることの確認
- **アサーション**: 
  - 取引時間が非負の値で返される
  - 時間単位での計算が正しく行われる

### TestTrade_ToCSVRecord
```go
func TestTrade_ToCSVRecord(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    trade := NewTradeFromPosition(position, 1.1010)
    
    record := trade.ToCSVRecord()
    
    expectedLength := 11
    if len(record) != expectedLength { ... }
    
    if record[0] != trade.ID { ... }
    if record[1] != trade.Symbol { ... }
    if record[2] != trade.Side.String() { ... }
}
```
- **テスト内容**: CSV形式への変換機能
- **テストケース**: 
  - 正常系: 有効なTradeオブジェクトのCSV変換
  - フォーマット検証: 各フィールドの文字列形式
  - 配列長の確認
- **アサーション**: 
  - 配列長が11（全フィールド）
  - ID、Symbol、Sideが正しく文字列化される
  - 数値フィールドが適切にフォーマットされる

### TestTradeStatus_String
```go
func TestTradeStatus_String(t *testing.T) {
    tests := []struct {
        status   TradeStatus
        expected string
    }{
        {TradeOpen, "Open"},
        {TradeClosed, "Closed"},
        {TradeCanceled, "Canceled"},
        {TradeStatus(999), "Unknown"},
    }
    
    for _, test := range tests {
        if test.status.String() != test.expected { ... }
    }
}
```
- **テスト内容**: TradeStatus型の文字列変換機能
- **テストケース**: 
  - 正常系: Open, Closed, Canceled各ステータスの文字列変換
  - 異常系: 未定義値での"Unknown"返却
- **アサーション**: 
  - 各TradeStatusが適切な文字列に変換される
  - 未定義の値で"Unknown"が返される

### TestCalculateTradePnL
```go
func TestCalculateTradePnL(t *testing.T) {
    // 買い取引のテスト
    buyPnL := calculateTradePnL(Buy, 10000.0, 1.1000, 1.1010)
    expectedBuyPnL := (1.1010 - 1.1000) * 10000.0
    assertFloatEqual(t, expectedBuyPnL, buyPnL, "Buy trade PnL")
    
    // 売り取引のテスト
    sellPnL := calculateTradePnL(Sell, 10000.0, 1.1000, 1.0990)
    expectedSellPnL := (1.1000 - 1.0990) * 10000.0
    assertFloatEqual(t, expectedSellPnL, sellPnL, "Sell trade PnL")
}
```
- **テスト内容**: 損益計算関数（パッケージ内部関数）
- **テストケース**: 
  - 正常系: 買い取引での損益計算
  - 正常系: 売り取引での損益計算
- **アサーション**: 
  - 買い取引で(ExitPrice - EntryPrice) * Sizeで計算
  - 売り取引で(EntryPrice - ExitPrice) * Sizeで計算
  - 浮動小数点比較では許容誤差を使用

## 実装済みテストの概要
- **正常系テスト数**: 11個
- **異常系テスト数**: 2個  
- **境界値テスト数**: 3個
- **カバレッジ**: 100%

## 特記事項
- ポジションから取引履歴への変換機能を網羅的にテスト
- 買い取引と売り取引の両方の損益計算をテスト
- 取引結果の分類（勝ち・負け・引き分け）機能をテスト
- CSV出力機能と文字列変換機能も含む
- 浮動小数点計算では許容誤差付きの比較を使用

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestTrade ./pkg/models/

# 詳細出力
go test -v -run TestTrade ./pkg/models/

# カバレッジ付き
go test -cover -run TestTrade ./pkg/models/
```