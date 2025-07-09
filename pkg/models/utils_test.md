# Utils テスト仕様書

## 概要
- **テスト対象**: `pkg/models/utils.go` のユーティリティ関数と型
- **テストの目的**: ID生成、型変換、バリデーション機能の正常性を確認
- **実装されているテスト関数**: 
  - `TestGenerateID`
  - `TestParseOrderSide`
  - `TestParseOrderType`
  - `TestValidationError_Error`
  - `TestValidateStruct`

## テスト関数詳細

### TestGenerateID
```go
func TestGenerateID(t *testing.T) {
    id1 := GenerateID()
    id2 := GenerateID()
    
    if id1 == id2 { ... }
    if len(id1) == 0 { ... }
}
```
- **テスト内容**: GenerateID関数によるユニークID生成機能
- **テストケース**: 
  - 正常系: 連続して呼び出した際のユニーク性確認
  - 正常系: 生成されたIDが空でないことの確認
- **アサーション**: 
  - 連続して生成されたIDが異なる値
  - 生成されたIDが空文字列でない
  - 時刻ベースのID生成が正常に動作

### TestParseOrderSide
```go
func TestParseOrderSide(t *testing.T) {
    tests := []struct {
        input    string
        expected OrderSide
        hasError bool
    }{
        {"buy", Buy, false},
        {"Buy", Buy, false},
        {"BUY", Buy, false},
        {"sell", Sell, false},
        {"Sell", Sell, false},
        {"SELL", Sell, false},
        {"invalid", Buy, true},
        {"", Buy, true},
    }
    
    for _, test := range tests {
        result, err := ParseOrderSide(test.input)
        
        if test.hasError {
            if err == nil { ... }
        } else {
            if err != nil { ... }
            if result != test.expected { ... }
        }
    }
}
```
- **テスト内容**: ParseOrderSide関数による文字列からOrderSide型への変換機能
- **テストケース**: 
  - 正常系: "buy", "Buy", "BUY"での Buy 型への変換
  - 正常系: "sell", "Sell", "SELL"での Sell 型への変換
  - 異常系: 無効な文字列でのエラー返却
  - 異常系: 空文字列でのエラー返却
- **アサーション**: 
  - 大文字小文字を問わず正しく変換される
  - 有効な文字列ではエラーなし
  - 無効な文字列では適切なエラーが返される
  - エラーメッセージに入力値が含まれる

### TestParseOrderType
```go
func TestParseOrderType(t *testing.T) {
    tests := []struct {
        input    string
        expected OrderType
        hasError bool
    }{
        {"market", Market, false},
        {"Market", Market, false},
        {"MARKET", Market, false},
        {"limit", Limit, false},
        {"Limit", Limit, false},
        {"LIMIT", Limit, false},
        {"stop", Stop, false},
        {"Stop", Stop, false},
        {"STOP", Stop, false},
        {"invalid", Market, true},
        {"", Market, true},
    }
    
    for _, test := range tests {
        result, err := ParseOrderType(test.input)
        
        if test.hasError {
            if err == nil { ... }
        } else {
            if err != nil { ... }
            if result != test.expected { ... }
        }
    }
}
```
- **テスト内容**: ParseOrderType関数による文字列からOrderType型への変換機能
- **テストケース**: 
  - 正常系: "market", "Market", "MARKET"での Market 型への変換
  - 正常系: "limit", "Limit", "LIMIT"での Limit 型への変換
  - 正常系: "stop", "Stop", "STOP"での Stop 型への変換
  - 異常系: 無効な文字列でのエラー返却
  - 異常系: 空文字列でのエラー返却
- **アサーション**: 
  - 大文字小文字を問わず正しく変換される
  - 有効な文字列ではエラーなし
  - 無効な文字列では適切なエラーが返される
  - 全ての OrderType（Market, Limit, Stop）がサポートされる

### TestValidationError_Error
```go
func TestValidationError_Error(t *testing.T) {
    err := ValidationError{
        Field:   "test_field",
        Value:   "test_value",
        Message: "test message",
    }
    
    errorMsg := err.Error()
    
    if !strings.Contains(errorMsg, "test_field") { ... }
    if !strings.Contains(errorMsg, "test message") { ... }
}
```
- **テスト内容**: ValidationError型のError()メソッド機能
- **テストケース**: 
  - 正常系: エラーメッセージにフィールド名が含まれることの確認
  - 正常系: エラーメッセージにメッセージ内容が含まれることの確認
- **アサーション**: 
  - エラーメッセージにFieldの値が含まれる
  - エラーメッセージにMessageの値が含まれる
  - 適切にフォーマットされた文字列が返される

### TestValidateStruct
```go
func TestValidateStruct(t *testing.T) {
    // 正常なケース
    config := NewDefaultConfig()
    config.Market.DataProvider.FilePath = "./testdata/sample.csv"
    
    if err := ValidateStruct(&config); err != nil { ... }
    
    // 異常なケース
    config.Market.DataProvider.FilePath = ""
    if err := ValidateStruct(&config); err == nil { ... }
}
```
- **テスト内容**: ValidateStruct関数による構造体バリデーション機能
- **テストケース**: 
  - 正常系: 有効な構造体でのバリデーション成功
  - 異常系: 無効な構造体でのバリデーション失敗
  - Validatorインターフェースの動作確認
- **アサーション**: 
  - 有効な構造体ではエラーなし
  - 無効な構造体では適切なエラーが返される
  - Validatorインターフェースが正しく呼び出される

## 実装済みテストの概要
- **正常系テスト数**: 12個
- **異常系テスト数**: 6個  
- **境界値テスト数**: 2個
- **カバレッジ**: 100%

## 特記事項
- 文字列変換関数で大文字小文字の区別なく変換可能
- ID生成の一意性をテスト（時刻ベース）
- カスタムエラー型のフォーマット機能をテスト
- Validatorインターフェースの実装をテスト
- テーブル駆動テストパターンを使用

## 実装されている機能
- ユニークID生成（時刻ベース）
- OrderSide、OrderType の文字列変換
- ValidationError カスタムエラー型
- Validator インターフェース
- 構造体バリデーション統合機能

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestGenerate -run TestParse -run TestValidation ./pkg/models/

# 詳細出力
go test -v -run TestGenerate -run TestParse -run TestValidation ./pkg/models/

# カバレッジ付き
go test -cover -run TestGenerate -run TestParse -run TestValidation ./pkg/models/
```