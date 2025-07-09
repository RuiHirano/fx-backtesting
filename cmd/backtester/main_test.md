# CLI Backtester テスト仕様書

## 概要
- **テスト対象**: `cmd/backtester/main.go` の CLI アプリケーション
- **テスト目的**: コマンドライン引数解析、実行フロー、エラーハンドリングの検証
- **テスト対象メソッド**: 
  - `TestCLI_ArgumentParsing`
  - `TestCLI_ExecutionFlow`
  - `TestCLI_ErrorHandling`
  - `TestCLI_ConfigValidation`
  - `TestCLI_OutputGeneration`
  - `TestCLI_HelpMessages`

## テスト内容

### TestCLI_ArgumentParsing
```go
func TestCLI_ArgumentParsing(t *testing.T) {
    // 正常な引数解析テスト
    args := []string{
        "backtester",
        "-data", "./testdata/sample.csv",
        "-config", "./testdata/config.json",
        "-output", "./results.txt",
        "-format", "json",
    }
    
    config, err := parseArgs(args)
    
    // データファイル確認
    if config.DataFile != "./testdata/sample.csv" { ... }
    
    // 設定ファイル確認
    if config.ConfigFile != "./testdata/config.json" { ... }
    
    // 出力ファイル確認
    if config.OutputFile != "./results.txt" { ... }
    
    // フォーマット確認
    if config.Format != "json" { ... }
    
    // 必須引数なしテスト
    argsNoData := []string{"backtester", "-config", "./config.json"}
    _, err = parseArgs(argsNoData)
    if err == nil { t.Error("Expected error for missing data argument") }
    
    // デフォルト値テスト
    argsMinimal := []string{
        "backtester",
        "-data", "./testdata/sample.csv",
        "-config", "./testdata/config.json",
    }
    
    config, err = parseArgs(argsMinimal)
    if config.Format != "text" { ... }
    if config.OutputFile != "" { ... }
}
```
- **テスト目的**: CLI引数解析機能の検証
- **テスト条件**: 
  - 事前条件: 完全な引数セット、必須引数のみ、無効引数パターン
  - 入力値: 各種引数の組み合わせ（data, config, output, format, help）
  - 期待結果: 正確な引数解析、適切なエラー検出、デフォルト値設定
- **検証項目**: 
  - 正常な引数解析（全パラメータ）
  - 必須引数チェック（data, config）
  - 無効引数の検出（存在しないフラグ）
  - デフォルト値の設定（format=text, output=""）
  - ヘルプフラグの処理

### TestCLI_ExecutionFlow
```go
func TestCLI_ExecutionFlow(t *testing.T) {
    // テストデータの準備
    setupTestData(t)
    defer cleanupTestData(t)
    
    // 正常な実行フローテスト
    config := &CLIConfig{
        DataFile:   "./testdata/sample.csv",
        ConfigFile: "./testdata/config.json",
        Format:     "text",
        OutputFile: "./testdata/output.txt",
    }
    
    err := runBacktest(config)
    
    // 出力ファイルの存在確認
    if _, err := os.Stat(config.OutputFile); os.IsNotExist(err) { ... }
    
    // 出力ファイルの内容確認
    content, err := os.ReadFile(config.OutputFile)
    if len(content) == 0 { ... }
    
    // 出力内容の基本チェック
    if !strings.Contains(string(content), "バックテスト結果") { ... }
}
```
- **テスト目的**: バックテスト実行フロー全体の検証
- **テスト条件**: 
  - 事前条件: テストデータ（sample.csv, config.json）の準備
  - 入力値: 有効な設定でのバックテスト実行
  - 期待結果: 正常な実行完了、出力ファイル生成、適切な内容
- **検証項目**: 
  - 設定ファイル読み込み成功
  - Backtester初期化成功
  - バックテスト実行完了
  - 出力ファイル生成確認
  - 出力内容の妥当性（"バックテスト結果"文字列含有）

### TestCLI_ErrorHandling
```go
func TestCLI_ErrorHandling(t *testing.T) {
    // 存在しないデータファイル
    config := &CLIConfig{
        DataFile:   "./nonexistent.csv",
        ConfigFile: "./testdata/config.json",
        Format:     "text",
        OutputFile: "",
    }
    
    err := runBacktest(config)
    if err == nil { t.Error("Expected error for nonexistent data file") }
    
    // 存在しない設定ファイル
    config = &CLIConfig{
        DataFile:   "./testdata/sample.csv",
        ConfigFile: "./nonexistent.json",
        Format:     "text",
        OutputFile: "",
    }
    
    err = runBacktest(config)
    if err == nil { t.Error("Expected error for nonexistent config file") }
    
    // 無効なフォーマット
    config.Format = "invalid"
    err = runBacktest(config)
    if err == nil { t.Error("Expected error for invalid format") }
    
    // 無効な出力ディレクトリ
    config.OutputFile = "/invalid/path/output.txt"
    err = runBacktest(config)
    if err == nil { t.Error("Expected error for invalid output directory") }
}
```
- **テスト目的**: エラーハンドリングの検証
- **テスト条件**: 
  - 異常値: 存在しないファイル、無効な設定、不正なフォーマット、無効なパス
  - 期待結果: 適切なエラーメッセージ、処理の適切な停止
- **検証項目**: 
  - 存在しないデータファイルエラー
  - 存在しない設定ファイルエラー
  - 無効なフォーマットエラー
  - 無効な出力ディレクトリエラー
  - エラーメッセージの適切性

### TestCLI_ConfigValidation
```go
func TestCLI_ConfigValidation(t *testing.T) {
    setupTestData(t)
    defer cleanupTestData(t)
    
    // 正常な設定ファイル
    config, err := loadConfig("./testdata/config.json")
    if err != nil { ... }
    if config == nil { ... }
    
    // 設定内容の基本確認
    if config.Broker.InitialBalance <= 0 { ... }
    if config.Broker.Spread < 0 { ... }
    
    // 無効な設定ファイル
    invalidConfigPath := "./testdata/invalid_config.json"
    createInvalidConfig(t, invalidConfigPath)
    defer os.Remove(invalidConfigPath)
    
    _, err = loadConfig(invalidConfigPath)
    if err == nil { t.Error("Expected error for invalid config file") }
}
```
- **テスト目的**: 設定ファイル検証機能の検証
- **テスト条件**: 
  - 事前条件: 正常な設定ファイル、無効な設定ファイルを準備
  - 入力値: JSON形式の設定ファイル
  - 期待結果: 正常な設定の読み込み、無効な設定の拒否
- **検証項目**: 
  - JSON形式の設定ファイル読み込み成功
  - 設定項目の妥当性チェック（正の初期残高、非負のスプレッド）
  - 無効な設定の検出（負の初期残高、負のスプレッド）
  - 適切なエラーメッセージ

### TestCLI_OutputGeneration
```go
func TestCLI_OutputGeneration(t *testing.T) {
    setupTestData(t)
    defer cleanupTestData(t)
    
    // テキスト形式出力
    config := &CLIConfig{
        DataFile:   "./testdata/sample.csv",
        ConfigFile: "./testdata/config.json",
        Format:     "text",
        OutputFile: "./testdata/output_text.txt",
    }
    
    err := runBacktest(config)
    // JSON形式出力
    config.Format = "json"
    config.OutputFile = "./testdata/output_json.json"
    err = runBacktest(config)
    
    // CSV形式出力
    config.Format = "csv"
    config.OutputFile = "./testdata/output_csv.csv"
    err = runBacktest(config)
    
    // 出力ファイルの存在確認
    outputFiles := map[string]string{
        "text": "./testdata/output_text.txt",
        "json": "./testdata/output_json.json",
        "csv":  "./testdata/output_csv.csv",
    }
    for format, filename := range outputFiles {
        if _, err := os.Stat(filename); os.IsNotExist(err) { ... }
    }
    
    // 標準出力テスト
    config.OutputFile = ""
    config.Format = "text"
    
    var stdout bytes.Buffer
    err = runBacktestWithOutput(config, &stdout)
    if stdout.Len() == 0 { ... }
}
```
- **テスト目的**: 出力生成機能の検証
- **テスト条件**: 
  - 事前条件: 各種出力フォーマットの指定
  - 入力値: text, json, csv形式指定
  - 期待結果: 適切な形式での出力生成
- **検証項目**: 
  - 複数フォーマット対応（Text, JSON, CSV）
  - ファイル出力機能の確認
  - 標準出力への出力確認
  - 出力ファイルの存在確認

### TestCLI_HelpMessages
```go
func TestCLI_HelpMessages(t *testing.T) {
    // ヘルプメッセージの取得
    helpText := getHelpMessage()
    
    if helpText == "" { ... }
    
    // ヘルプメッセージの内容確認
    requiredElements := []string{
        "Usage:",
        "-data",
        "-config",
        "-output",
        "-format",
        "-help",
        "Examples:",
    }
    
    for _, element := range requiredElements {
        if !strings.Contains(helpText, element) { ... }
    }
    
    // ヘルプフラグでの終了テスト
    args := []string{"backtester", "-help"}
    config, err := parseArgs(args)
    if !config.ShowHelp { ... }
}
```
- **テスト目的**: ヘルプメッセージの検証
- **テスト条件**: 
  - 事前条件: ヘルプメッセージ生成機能
  - 入力値: -help フラグ
  - 期待結果: 適切なヘルプメッセージ表示
- **検証項目**: 
  - ヘルプメッセージの生成確認
  - 必要な要素の包含確認（Usage, 引数説明, Examples）
  - ヘルプフラグの正確な処理

## 実装された機能

### CLI引数解析
- **必須引数**: `-data` (データファイル), `-config` (設定ファイル)
- **オプション引数**: `-output` (出力ファイル), `-format` (出力形式), `-help` (ヘルプ)
- **デフォルト値**: format="text", output=""(標準出力)
- **バリデーション**: 必須引数チェック、フォーマット検証

### 実行フロー
1. **引数解析**: コマンドライン引数の処理
2. **設定読み込み**: JSON設定ファイルの読み込み・検証
3. **Backtester初期化**: Market・Broker統合初期化
4. **バックテスト実行**: 簡易取引戦略による実行
5. **レポート生成**: 統計レポートの生成
6. **出力**: 指定形式での結果出力

### エラーハンドリング
- **ファイルエラー**: 存在しないファイル、読み込みエラー
- **設定エラー**: 無効な設定値、JSON解析エラー
- **実行エラー**: Backtester初期化失敗、取引エラー
- **出力エラー**: ファイル書き込みエラー、ディレクトリ作成エラー

### 出力フォーマット
- **Text**: 日本語での詳細レポート
- **JSON**: 構造化データ形式
- **CSV**: 取引履歴データ

## 結果（テスト数と実績）
- **テスト関数数**: 6個
- **テストケース数**: 20個以上
- **カバレッジ**: 77.6%
- **正常系テスト**: 引数解析、実行フロー、出力生成
- **異常系テスト**: エラーハンドリング、設定検証

## CLI 仕様詳細

### 基本的な使用方法
```bash
# 基本実行
./backtester -data ./data/EURUSD.csv -config ./config.json

# 出力形式指定
./backtester -data ./data/EURUSD.csv -config ./config.json -format json

# 出力ファイル指定
./backtester -data ./data/EURUSD.csv -config ./config.json -output ./results.txt

# ヘルプ表示
./backtester -help
```

### 引数一覧
- **-data**: データファイルパス（必須）
- **-config**: 設定ファイルパス（必須）
- **-output**: 出力ファイルパス（オプション、デフォルト: 標準出力）
- **-format**: 出力フォーマット（text/json/csv、デフォルト: text）
- **-help**: ヘルプメッセージ表示

### 設定ファイル例
```json
{
  "market": {
    "data_provider": {
      "file_path": "./data/EURUSD.csv",
      "format": "csv"
    },
    "symbol": "EURUSD"
  },
  "broker": {
    "initial_balance": 10000.0,
    "spread": 0.0001
  }
}
```

## Examples

### 作成したサンプル
1. **basic_example.go**: 基本的なバックテスト実行例
2. **strategy_example.go**: 移動平均クロス戦略実装例
3. **config_example.go**: 各種設定パターン比較例
4. **README.md**: 使用方法と設定例の説明

### 実行方法
```bash
# Examples実行
cd examples
go run basic_example.go
go run strategy_example.go
go run config_example.go
```

## テスト実行
```bash
# 全テスト実行
go test ./cmd/backtester/... -v

# カバレッジ確認
go test ./cmd/backtester/... -cover

# 個別テスト実行
go test -run TestCLI_ArgumentParsing ./cmd/backtester/
go test -run TestCLI_ExecutionFlow ./cmd/backtester/
go test -run TestCLI_ErrorHandling ./cmd/backtester/
go test -run TestCLI_ConfigValidation ./cmd/backtester/
go test -run TestCLI_OutputGeneration ./cmd/backtester/
go test -run TestCLI_HelpMessages ./cmd/backtester/
```

## ヘルパー関数
- **setupTestData()**: テストデータ（CSV、JSON）の作成
- **cleanupTestData()**: テストデータの削除
- **createInvalidConfig()**: 無効な設定ファイルの作成
- **calculateTradeSize()**: 残高に応じた取引サイズ計算