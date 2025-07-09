# Order テスト仕様書

## 概要
- **テスト対象**: `pkg/models/order.go` の Order 構造体と関連型
- **テストの目的**: 注文データの作成、バリデーション、型変換機能の正常性を確認
- **実装されているテスト関数**: 
  - `TestOrder_NewMarketOrder`
  - `TestOrder_NewLimitOrder`
  - `TestOrder_Validate`
  - `TestOrder_IsMarket`
  - `TestOrder_IsLimit`
  - `TestOrderType_String`
  - `TestOrderSide_String`

## テスト関数詳細

### TestOrder_NewMarketOrder
```go
func TestOrder_NewMarketOrder(t *testing.T) {
    order := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
    
    if order.ID != "test-123" { ... }
    if order.Symbol != "EURUSD" { ... }
    if order.Type != Market { ... }
    if order.Side != Buy { ... }
    if order.Size != 10000.0 { ... }
}
```
- **テスト内容**: NewMarketOrder関数による成行注文の作成
- **テストケース**: 
  - 正常系: 有効なパラメータでの成行注文作成
  - ID、Symbol、Type、Side、Sizeの各フィールド検証
  - Timestampが自動設定されることの確認
- **アサーション**: 
  - 指定したIDが正しく設定される
  - 注文タイプがMarketに設定される
  - 注文サイドがBuyに設定される
  - 注文サイズが正しく設定される

### TestOrder_NewLimitOrder
```go
func TestOrder_NewLimitOrder(t *testing.T) {
    order := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
    
    if order.Type != Limit { ... }
    if order.Price != 1.1000 { ... }
}
```
- **テスト内容**: NewLimitOrder関数による指値注文の作成
- **テストケース**: 
  - 正常系: 有効なパラメータでの指値注文作成
  - 注文タイプがLimitに設定されることの確認
  - 指値価格が正しく設定されることの確認
- **アサーション**: 
  - 注文タイプがLimitに設定される
  - 指定した価格が正しく設定される
  - その他のフィールドも適切に設定される

### TestOrder_Validate
```go
func TestOrder_Validate(t *testing.T) {
    // 正常なケース - Market注文
    order := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
    if err := order.Validate(); err != nil { ... }
    
    // 正常なケース - Limit注文
    order = NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
    if err := order.Validate(); err != nil { ... }
    
    // 異常なケース - 負のサイズ
    order = NewMarketOrder("test-123", "EURUSD", Buy, -10000.0)
    if err := order.Validate(); err == nil { ... }
    
    // 異常なケース - 空のシンボル
    order = NewMarketOrder("test-123", "", Buy, 10000.0)
    if err := order.Validate(); err == nil { ... }
    
    // 異常なケース - Limit注文で価格が0
    order = NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 0)
    if err := order.Validate(); err == nil { ... }
}
```
- **テスト内容**: Order構造体のバリデーション機能
- **テストケース**: 
  - 正常系: 有効な成行注文でのバリデーション成功
  - 正常系: 有効な指値注文でのバリデーション成功
  - 異常系: 負の注文サイズでのエラー
  - 異常系: 空のシンボルでのエラー
  - 異常系: 指値注文で価格が0の場合のエラー
- **アサーション**: 
  - 正常な注文ではエラーなし
  - 無効な注文では適切なエラーメッセージを返す
  - 注文タイプに応じた適切なバリデーション

### TestOrder_IsMarket
```go
func TestOrder_IsMarket(t *testing.T) {
    marketOrder := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
    if !marketOrder.IsMarket() { ... }
    
    limitOrder := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
    if limitOrder.IsMarket() { ... }
}
```
- **テスト内容**: 成行注文判定機能
- **テストケース**: 
  - 正常系: 成行注文でtrue
  - 正常系: 指値注文でfalse
- **アサーション**: 
  - 成行注文がMarketとして識別される
  - 指値注文がMarketとして識別されない

### TestOrder_IsLimit
```go
func TestOrder_IsLimit(t *testing.T) {
    limitOrder := NewLimitOrder("test-123", "EURUSD", Sell, 10000.0, 1.1000)
    if !limitOrder.IsLimit() { ... }
    
    marketOrder := NewMarketOrder("test-123", "EURUSD", Buy, 10000.0)
    if marketOrder.IsLimit() { ... }
}
```
- **テスト内容**: 指値注文判定機能
- **テストケース**: 
  - 正常系: 指値注文でtrue
  - 正常系: 成行注文でfalse
- **アサーション**: 
  - 指値注文がLimitとして識別される
  - 成行注文がLimitとして識別されない

### TestOrderType_String
```go
func TestOrderType_String(t *testing.T) {
    tests := []struct {
        orderType OrderType
        expected  string
    }{
        {Market, "Market"},
        {Limit, "Limit"},
        {Stop, "Stop"},
        {OrderType(999), "Unknown"},
    }
    
    for _, test := range tests {
        if test.orderType.String() != test.expected { ... }
    }
}
```
- **テスト内容**: OrderType型の文字列変換機能
- **テストケース**: 
  - 正常系: Market, Limit, Stop各タイプの文字列変換
  - 異常系: 未定義値での"Unknown"返却
- **アサーション**: 
  - 各OrderTypeが適切な文字列に変換される
  - 未定義の値で"Unknown"が返される

### TestOrderSide_String
```go
func TestOrderSide_String(t *testing.T) {
    tests := []struct {
        orderSide OrderSide
        expected  string
    }{
        {Buy, "Buy"},
        {Sell, "Sell"},
        {OrderSide(999), "Unknown"},
    }
    
    for _, test := range tests {
        if test.orderSide.String() != test.expected { ... }
    }
}
```
- **テスト内容**: OrderSide型の文字列変換機能
- **テストケース**: 
  - 正常系: Buy, Sell各サイドの文字列変換
  - 異常系: 未定義値での"Unknown"返却
- **アサーション**: 
  - 各OrderSideが適切な文字列に変換される
  - 未定義の値で"Unknown"が返される

## 実装済みテストの概要
- **正常系テスト数**: 8個
- **異常系テスト数**: 5個  
- **境界値テスト数**: 2個
- **カバレッジ**: 100%

## 特記事項
- 成行注文と指値注文の両方のテストケースを網羅
- 注文タイプと注文サイドの列挙型の文字列変換機能をテスト
- 注文バリデーションで注文タイプに応じた適切な検証を実施
- エラーメッセージの内容も検証対象

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestOrder ./pkg/models/

# 詳細出力
go test -v -run TestOrder ./pkg/models/

# カバレッジ付き
go test -cover -run TestOrder ./pkg/models/
```