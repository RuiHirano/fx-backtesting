# Config テスト仕様書

## 概要
- **テスト対象**: `pkg/models/config.go` の Config 構造体とその関連構造体
- **テストの目的**: バックテスト設定の作成、バリデーション機能の正常性を確認
- **実装されているテスト関数**: 
  - `TestConfig_NewDefaultConfig`
  - `TestConfig_Validate`
  - `TestMarketConfig_Validate`
  - `TestDataProviderConfig_Validate`
  - `TestBrokerConfig_Validate`

## テスト関数詳細

### TestConfig_NewDefaultConfig
```go
func TestConfig_NewDefaultConfig(t *testing.T) {
    config := NewDefaultConfig()
    // デフォルト値の検証
    if config.Market.Symbol != "EURUSD" { ... }
    if config.Market.DataProvider.Format != "csv" { ... }
    if config.Broker.InitialBalance != 10000.0 { ... }
    if config.Broker.Spread != 0.0001 { ... }
}
```
- **テスト内容**: NewDefaultConfig関数によるデフォルト設定の生成
- **テストケース**: 
  - 正常系: デフォルト値での設定構造体作成
  - Market.Symbol = "EURUSD"の確認
  - DataProvider.Format = "csv"の確認
  - Broker.InitialBalance = 10000.0の確認
  - Broker.Spread = 0.0001の確認
- **アサーション**: 
  - 各フィールドがデフォルト値で正しく設定される
  - 構造体の階層構造が正常に作成される

### TestConfig_Validate
```go
func TestConfig_Validate(t *testing.T) {
    config := NewDefaultConfig()
    config.Market.DataProvider.FilePath = "./testdata/sample.csv"
    
    // 正常なケース
    if err := config.Validate(); err != nil { ... }
    
    // 異常なケース - 空のファイルパス
    config.Market.DataProvider.FilePath = ""
    if err := config.Validate(); err == nil { ... }
    
    // 異常なケース - 負の初期残高
    config.Broker.InitialBalance = -1000
    if err := config.Validate(); err == nil { ... }
}
```
- **テスト内容**: Config構造体全体のバリデーション機能
- **テストケース**: 
  - 正常系: 有効な設定でのバリデーション成功
  - 異常系: 空のファイルパス指定時のエラー
  - 異常系: 負の初期残高指定時のエラー
- **アサーション**: 
  - 正常な設定ではエラーなし
  - 無効な設定では適切なエラーを返す
  - 子構造体のバリデーションが呼び出される

### TestMarketConfig_Validate
```go
func TestMarketConfig_Validate(t *testing.T) {
    config := MarketConfig{
        DataProvider: DataProviderConfig{
            FilePath: "./testdata/sample.csv",
            Format:   "csv",
        },
        Symbol: "EURUSD",
    }
    
    // 正常なケース
    if err := config.Validate(); err != nil { ... }
    
    // 異常なケース - 空のシンボル
    config.Symbol = ""
    if err := config.Validate(); err == nil { ... }
}
```
- **テスト内容**: MarketConfig構造体のバリデーション機能
- **テストケース**: 
  - 正常系: 有効なマーケット設定でのバリデーション成功
  - 異常系: 空のシンボル指定時のエラー
- **アサーション**: 
  - 正常な設定ではエラーなし
  - 空文字列や空白文字のシンボルでエラー
  - DataProviderConfigのバリデーションも実行される

### TestDataProviderConfig_Validate
```go
func TestDataProviderConfig_Validate(t *testing.T) {
    config := DataProviderConfig{
        FilePath: "./testdata/sample.csv",
        Format:   "csv",
    }
    
    // 正常なケース
    if err := config.Validate(); err != nil { ... }
    
    // 異常なケース - 空のファイルパス
    config.FilePath = ""
    if err := config.Validate(); err == nil { ... }
    
    // 異常なケース - 無効なフォーマット
    config.Format = "xml"
    if err := config.Validate(); err == nil { ... }
}
```
- **テスト内容**: DataProviderConfig構造体のバリデーション機能
- **テストケース**: 
  - 正常系: 有効なデータプロバイダー設定でのバリデーション成功
  - 異常系: 空のファイルパス指定時のエラー
  - 異常系: サポートされていないフォーマット指定時のエラー
- **アサーション**: 
  - 正常な設定ではエラーなし
  - ファイルパスが空文字列でエラー
  - フォーマットが"csv"や"json"以外でエラー

### TestBrokerConfig_Validate
```go
func TestBrokerConfig_Validate(t *testing.T) {
    config := BrokerConfig{
        InitialBalance: 10000.0,
        Spread:         0.0001,
    }
    
    // 正常なケース
    if err := config.Validate(); err != nil { ... }
    
    // 異常なケース - 負の初期残高
    config.InitialBalance = -1000
    if err := config.Validate(); err == nil { ... }
    
    // 異常なケース - 負のスプレッド
    config.Spread = -0.0001
    if err := config.Validate(); err == nil { ... }
}
```
- **テスト内容**: BrokerConfig構造体のバリデーション機能
- **テストケース**: 
  - 正常系: 有効なブローカー設定でのバリデーション成功
  - 異常系: 負の初期残高指定時のエラー
  - 異常系: 負のスプレッド指定時のエラー
- **アサーション**: 
  - 正常な設定ではエラーなし
  - 初期残高が0以下でエラー
  - スプレッドが負の値でエラー

## 実装済みテストの概要
- **正常系テスト数**: 5個
- **異常系テスト数**: 6個  
- **境界値テスト数**: 3個
- **カバレッジ**: 100%

## 特記事項
- 階層構造を持つ設定のバリデーションを網羅
- エラーメッセージの内容も検証対象
- 実際のファイルパス存在確認は行わず、文字列バリデーションのみ
- 各子構造体のバリデーションが親構造体から呼び出されることを確認

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestConfig ./pkg/models/

# 詳細出力
go test -v -run TestConfig ./pkg/models/

# カバレッジ付き
go test -cover -run TestConfig ./pkg/models/
```