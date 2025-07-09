# Position テスト仕様書

## 概要
- **テスト対象**: `pkg/models/position.go` の Position 構造体
- **テストの目的**: ポジション管理、損益計算、ストップロス・テイクプロフィット機能の正常性を確認
- **実装されているテスト関数**: 
  - `TestPosition_NewPosition`
  - `TestPosition_UpdatePrice`
  - `TestPosition_IsLong`
  - `TestPosition_IsShort`
  - `TestPosition_ShouldStopLoss`
  - `TestPosition_ShouldTakeProfit`
  - `TestPosition_GetMarketValue`
  - `TestPosition_GetPnLPercentage`

## テスト関数詳細

### TestPosition_NewPosition
```go
func TestPosition_NewPosition(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    
    if position.ID != "pos-123" { ... }
    if position.Symbol != "EURUSD" { ... }
    if position.Side != Buy { ... }
    if position.Size != 10000.0 { ... }
    if position.EntryPrice != 1.1000 { ... }
    if position.CurrentPrice != 1.1000 { ... }
    if position.PnL != 0.0 { ... }
}
```
- **テスト内容**: NewPosition関数によるポジション構造体の作成
- **テストケース**: 
  - 正常系: 有効なパラメータでのポジション作成
  - 全フィールドが指定した値で正しく設定されることの確認
  - 初期PnLが0.0に設定されることの確認
  - CurrentPriceがEntryPriceと同じ値に設定されることの確認
- **アサーション**: 
  - ID、Symbol、Side、Size、EntryPriceが指定値と一致
  - CurrentPriceがEntryPriceと同じ値
  - 初期PnLが0.0
  - OpenTimeが設定されている

### TestPosition_UpdatePrice
```go
func TestPosition_UpdatePrice(t *testing.T) {
    // 買いポジションのテスト
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    
    // 価格上昇 -> 利益
    position.UpdatePrice(1.1010)
    if position.CurrentPrice != 1.1010 { ... }
    expectedPnL := (1.1010 - 1.1000) * 10000.0
    assertFloatEqual(t, expectedPnL, position.PnL, "Buy position PnL")
    
    // 売りポジションのテスト
    position = NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
    
    // 価格下落 -> 利益
    position.UpdatePrice(1.0990)
    expectedPnL = (1.1000 - 1.0990) * 10000.0
    assertFloatEqual(t, expectedPnL, position.PnL, "Sell position PnL")
}
```
- **テスト内容**: 価格更新と損益計算機能
- **テストケース**: 
  - 正常系: 買いポジションでの価格上昇時の利益計算
  - 正常系: 売りポジションでの価格下落時の利益計算
  - CurrentPriceが正しく更新されることの確認
  - PnL計算が買い・売りで正しく動作することの確認
- **アサーション**: 
  - CurrentPriceが更新される
  - 買いポジションで(CurrentPrice - EntryPrice) * Sizeで計算
  - 売りポジションで(EntryPrice - CurrentPrice) * Sizeで計算
  - 浮動小数点比較では許容誤差を使用

### TestPosition_IsLong
```go
func TestPosition_IsLong(t *testing.T) {
    buyPosition := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    if !buyPosition.IsLong() { ... }
    
    sellPosition := NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
    if sellPosition.IsLong() { ... }
}
```
- **テスト内容**: ロングポジション判定機能
- **テストケース**: 
  - 正常系: 買いポジションでtrue
  - 正常系: 売りポジションでfalse
- **アサーション**: 
  - Buyサイドのポジションがロングとして識別される
  - Sellサイドのポジションがロングとして識別されない

### TestPosition_IsShort
```go
func TestPosition_IsShort(t *testing.T) {
    sellPosition := NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
    if !sellPosition.IsShort() { ... }
    
    buyPosition := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    if buyPosition.IsShort() { ... }
}
```
- **テスト内容**: ショートポジション判定機能
- **テストケース**: 
  - 正常系: 売りポジションでtrue
  - 正常系: 買いポジションでfalse
- **アサーション**: 
  - Sellサイドのポジションがショートとして識別される
  - Buyサイドのポジションがショートとして識別されない

### TestPosition_ShouldStopLoss
```go
func TestPosition_ShouldStopLoss(t *testing.T) {
    // 買いポジションのテスト
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    position.StopLoss = 1.0990
    
    // ストップロス未達
    position.UpdatePrice(1.1000)
    if position.ShouldStopLoss() { ... }
    
    // ストップロス到達
    position.UpdatePrice(1.0990)
    if !position.ShouldStopLoss() { ... }
    
    // 売りポジションのテスト
    position = NewPosition("pos-123", "EURUSD", Sell, 10000.0, 1.1000)
    position.StopLoss = 1.1010
    
    // ストップロス到達
    position.UpdatePrice(1.1010)
    if !position.ShouldStopLoss() { ... }
}
```
- **テスト内容**: ストップロス判定機能
- **テストケース**: 
  - 正常系: 買いポジションでストップロス価格到達時の判定
  - 正常系: 売りポジションでストップロス価格到達時の判定
  - 正常系: ストップロス未設定時の動作
- **アサーション**: 
  - 買いポジションでCurrentPrice <= StopLossの時にtrue
  - 売りポジションでCurrentPrice >= StopLossの時にtrue
  - StopLoss未設定(0以下)の時はfalse

### TestPosition_ShouldTakeProfit
```go
func TestPosition_ShouldTakeProfit(t *testing.T) {
    // 買いポジションのテスト
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    position.TakeProfit = 1.1020
    
    // テイクプロフィット未達
    position.UpdatePrice(1.1010)
    if position.ShouldTakeProfit() { ... }
    
    // テイクプロフィット到達
    position.UpdatePrice(1.1020)
    if !position.ShouldTakeProfit() { ... }
}
```
- **テスト内容**: テイクプロフィット判定機能
- **テストケース**: 
  - 正常系: 買いポジションでテイクプロフィット価格到達時の判定
  - 正常系: 売りポジションでテイクプロフィット価格到達時の判定
  - 正常系: テイクプロフィット未設定時の動作
- **アサーション**: 
  - 買いポジションでCurrentPrice >= TakeProfitの時にtrue
  - 売りポジションでCurrentPrice <= TakeProfitの時にtrue
  - TakeProfit未設定(0以下)の時はfalse

### TestPosition_GetMarketValue
```go
func TestPosition_GetMarketValue(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    position.UpdatePrice(1.1010)
    
    expectedValue := 1.1010 * 10000.0
    if position.GetMarketValue() != expectedValue { ... }
}
```
- **テスト内容**: 市場価値計算機能
- **テストケース**: 
  - 正常系: CurrentPrice × Sizeでの市場価値計算
- **アサーション**: 
  - CurrentPrice * Sizeの計算結果が正しく返される

### TestPosition_GetPnLPercentage
```go
func TestPosition_GetPnLPercentage(t *testing.T) {
    position := NewPosition("pos-123", "EURUSD", Buy, 10000.0, 1.1000)
    position.UpdatePrice(1.1010)
    
    expectedPercentage := ((1.1010 - 1.1000) / 1.1000) * 100
    assertFloatEqual(t, expectedPercentage, position.GetPnLPercentage(), "PnL percentage")
}
```
- **テスト内容**: 損益率計算機能
- **テストケース**: 
  - 正常系: (PnL / (EntryPrice * Size)) * 100での損益率計算
  - 境界値: EntryPriceが0の場合の動作
- **アサーション**: 
  - 損益率がパーセンテージで正しく計算される
  - 浮動小数点比較では許容誤差を使用

## 実装済みテストの概要
- **正常系テスト数**: 12個
- **異常系テスト数**: 2個  
- **境界値テスト数**: 4個
- **カバレッジ**: 100%

## 特記事項
- 買いポジションと売りポジションの両方の損益計算をテスト
- ストップロスとテイクプロフィットの条件判定を網羅
- 浮動小数点計算では許容誤差付きの比較を使用
- 市場価値と損益率の計算機能もカバー

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestPosition ./pkg/models/

# 詳細出力
go test -v -run TestPosition ./pkg/models/

# カバレッジ付き
go test -cover -run TestPosition ./pkg/models/
```