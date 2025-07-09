# FX バックテストライブラリ 開発計画書

## 1. 概要

この開発計画書は、FX バックテストライブラリの実装を TDD（テスト駆動開発）で進めるための詳細な計画です。設計書で定義されたアーキテクチャを実装し、高品質で保守性の高いコードベースを構築することを目指します。

## 2. プロジェクト構成

### 2.1 ディレクトリ構造

```
fx-backtesting/
├── cmd/                          # CLIアプリケーション
│   └── backtester/
│       └── main.go              # CLI エントリーポイント
├── pkg/                         # ライブラリコア
│   ├── models/                  # データモデル
│   │   ├── config.go           # 設定構造体
│   │   ├── candle.go           # ローソク足データ
│   │   ├── order.go            # 注文データ
│   │   ├── position.go         # ポジションデータ
│   │   ├── trade.go            # 取引履歴データ
│   │   ├── result.go           # バックテスト結果
│   │   └── models_test.go      # モデルテスト
│   ├── data/                   # データプロバイダー
│   │   ├── provider.go         # DataProviderインターフェース
│   │   ├── csv_provider.go     # CSV実装
│   │   ├── parser.go           # CSVパーサー
│   │   ├── provider_test.go    # プロバイダーテスト
│   │   └── parser_test.go      # パーサーテスト
│   ├── market/                 # 市場データ管理
│   │   ├── market.go           # Marketインターフェース
│   │   ├── market_impl.go      # Market実装
│   │   ├── market_test.go      # Marketテスト
│   │   └── mock_market.go      # テスト用モック
│   ├── broker/                 # 取引実行
│   │   ├── broker.go           # Brokerインターフェース
│   │   ├── simple_broker.go    # Broker実装
│   │   ├── broker_test.go      # Brokerテスト
│   │   └── mock_broker.go      # テスト用モック
│   ├── backtester/             # バックテスト統括
│   │   ├── backtester.go       # Backtester実装
│   │   ├── backtester_test.go  # Backtesterテスト
│   │   └── integration_test.go # 統合テスト
│   └── statistics/             # 統計計算・レポート
│       ├── calculator.go       # 統計計算エンジン
│       ├── report.go           # レポート生成
│       ├── metrics.go          # メトリクス定義
│       ├── formatter.go        # フォーマッター
│       ├── calculator_test.go  # 統計計算テスト
│       ├── report_test.go      # レポート生成テスト
│       ├── metrics_test.go     # メトリクステスト
│       └── formatter_test.go   # フォーマッターテスト
├── testdata/                   # テストデータ
│   ├── valid/
│   │   ├── sample.csv          # 正常データサンプル
│   │   └── large_dataset.csv   # 大容量データ
│   ├── invalid/
│   │   ├── malformed.csv       # 不正フォーマット
│   │   └── missing_columns.csv # カラム不足
│   └── edge_cases/
│       ├── empty.csv           # 空ファイル
│       └── single_row.csv      # 1行のみ
├── examples/                   # 使用例
│   ├── basic_example.go        # 基本的な使用方法
│   ├── advanced_example.go     # 高度な使用例
│   └── strategy_example.go     # 戦略実装例
├── docs/                       # 設計書・ドキュメント
│   ├── design.md               # 全体設計書
│   ├── dataProvider.md         # DataProvider設計書
│   ├── market.md               # Market設計書
│   ├── broker.md               # Broker設計書
│   ├── backtester.md           # Backtester設計書
│   ├── models.md               # Models設計書
│   ├── statistics.md           # Statistics設計書
│   └── development-plan.md     # この開発計画書
├── scripts/                   # 開発支援スクリプト
│   ├── test.sh                # テスト実行スクリプト
│   ├── coverage.sh            # カバレッジ測定
│   └── lint.sh                # 静的解析
├── go.mod                     # Go module定義
├── go.sum                     # 依存関係ロック
├── Makefile                   # ビルド・テスト自動化
└── README.md                  # プロジェクト概要
```

### 2.2 命名規則

- **パッケージ名**: 小文字、単語区切りなし（例: `backtester`, `dataProvider`）
- **ファイル名**: スネークケース（例: `csv_provider.go`, `market_test.go`）
- **構造体**: パスカルケース（例: `Backtester`, `MarketImpl`）
- **インターフェース**: パスカルケース（例: `Market`, `DataProvider`）
- **関数・メソッド**: キャメルケース（例: `Forward()`, `GetCurrentPrice()`）
- **定数**: 大文字スネークケース（例: `DEFAULT_SPREAD`）

## 3. TDD 開発フロー

### 3.1 TDD サイクル（t-wada式）

各機能の実装は以下の厳密な Red-Green-Refactor サイクルで進めます：

```
1. Red: 失敗するテストを書く
   - まず期待する仕様をテストコードで記述
   - テストを実行して「失敗」することを確認（Redの状態）
   - まだ実装が存在しないため、コンパイルエラーまたはテスト失敗となる

2. Green: テストを通すための最小限のコードを書く
   - テストが「成功」するための最小限の実装のみを行う
   - 綺麗なコードを書こうとせず、とにかくテストを通すことに集中
   - テストを実行して「成功」することを確認（Greenの状態）

3. Refactor: コードを改善・整理する
   - テストが通った状態でコードの品質を向上させる
   - 重複排除、命名改善、構造の最適化等を実施
   - 各改善後にテストを実行して「成功」を維持することを確認
```

### 3.2 TDD実行手順

#### 基本的な実行手順
```bash
# 1. Red: 失敗するテストを作成
# テストファイルを作成し、期待する仕様をテストコードで記述

# 2. テスト実行（失敗を確認）
go test ./pkg/models/candle_test.go -v
# → 出力例: FAIL (実装が存在しないためコンパイルエラーまたはテスト失敗)

# 3. Green: 最小限の実装
# テストが通るための最小限のコードを実装

# 4. テスト実行（成功を確認）
go test ./pkg/models/candle_test.go -v
# → 出力例: PASS

# 5. Refactor: コード改善
# テストが通った状態でコードの品質を向上

# 6. テスト実行（成功を維持）
go test ./pkg/models/candle_test.go -v
# → 出力例: PASS
```

#### テスト仕様書の作成
各テストファイルと同じディレクトリにマークダウン仕様書を作成します：

```
pkg/models/candle_test.go        → pkg/models/candle_test.md
pkg/models/config_test.go        → pkg/models/config_test.md
pkg/models/order_test.go         → pkg/models/order_test.md
pkg/models/position_test.go      → pkg/models/position_test.md
pkg/models/trade_test.go         → pkg/models/trade_test.md
pkg/models/result_test.go        → pkg/models/result_test.md
pkg/data/provider_test.go        → pkg/data/provider_test.md
```

#### テスト仕様書の構成
各xxx_test.mdファイルには、対応するxxx_test.goファイルの内容をわかりやすく記載します：

```markdown
# [Component] テスト仕様書

## 概要
- **テスト対象**: 対象ファイルとコンポーネント
- **テストの目的**: 何を検証するかの明確化
- **実装されているテスト関数**: xxx_test.goに含まれる全テスト関数のリスト

## テスト関数詳細

### TestXxx_Function1
```go
func TestXxx_Function1(t *testing.T) {
    // テストコードの概要をここに記載
}
```
- **テスト内容**: この関数が何をテストするか
- **テストケース**: 
  - 正常系: 期待される正常な動作
  - 異常系: エラーケースとその期待結果
  - 境界値: 境界値での動作確認
- **アサーション**: 何を検証しているか

### TestXxx_Function2
（同様の形式で各テスト関数を記載）

## 実装済みテストの概要
- **正常系テスト数**: X個
- **異常系テスト数**: Y個  
- **境界値テスト数**: Z個
- **カバレッジ**: XX%

## テスト実行方法
```bash
# 個別テスト実行
go test -run TestXxx ./pkg/path/

# 詳細出力
go test -v ./pkg/path/

# カバレッジ付き
go test -cover ./pkg/path/
```

### 3.3 開発順序

#### Phase 1: 基盤コンポーネント（1-2 週間）

**1. Models（データ構造）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
# 各コンポーネントのテスト仕様書をpkg内に作成
touch pkg/models/candle_test.md
touch pkg/models/config_test.md
touch pkg/models/order_test.md
touch pkg/models/position_test.md
touch pkg/models/trade_test.md
touch pkg/models/result_test.md

## ステップ2: Red - 失敗するテストを作成
# Candleから開始（最もシンプルなデータ構造）
1. pkg/models/candle_test.go を作成
   - TestCandle_NewCandle（コンストラクタテスト）
   - TestCandle_Validate（バリデーションテスト）
   - TestCandle_IsValidOHLC（OHLC検証テスト）
   - TestCandle_ToCSVRecord（シリアライゼーションテスト）

2. テスト実行（失敗確認）
   go test ./pkg/models/candle_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/models/candle.go を作成
   - Candle構造体の最小限定義
   - NewCandle関数の最小限実装
   - Validate関数の最小限実装
   - IsValidOHLC関数の最小限実装
   - ToCSVRecord関数の最小限実装

4. テスト実行（成功確認）
   go test ./pkg/models/candle_test.go -v
   # 出力: PASS

## ステップ4: Refactor - コード改善
5. コードの品質向上
   - エラーメッセージの改善
   - バリデーションロジックの最適化
   - CSV形式の精度向上
   - 構造体タグの追加

6. テスト実行（成功維持確認）
   go test ./pkg/models/candle_test.go -v
   # 出力: PASS

## 同様の手順でその他コンポーネントを実装
7. Config構造体 → Order構造体 → Position構造体 → Trade構造体 → Result構造体の順で実装
8. 各コンポーネントでRed-Green-Refactorサイクルを厳密に実行
9. 全コンポーネント完了後にutils.goとバリデーション機能を実装

## 最終確認
10. 全テスト実行とカバレッジ確認
    go test ./pkg/models/... -v -cover
    # 目標: カバレッジ85%以上

参考: models.md設計書を参照
```

**2. DataProvider（データ読み込み）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch pkg/data/provider_test.md
touch pkg/data/parser_test.md

## ステップ2: Red - 失敗するテストを作成
1. pkg/data/provider_test.go を作成
   - TestDataProvider_StreamData（データストリームテスト）
   - TestCSVProvider_Initialize（CSV初期化テスト）
   - TestCSVProvider_ErrorHandling（エラーハンドリングテスト）

2. テスト実行（失敗確認）
   go test ./pkg/data/provider_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/data/provider.go を作成
   - DataProviderインターフェース定義
   - CSVProvider構造体の最小限実装
   - StreamData関数の最小限実装

4. テスト実行（成功確認）
   go test ./pkg/data/provider_test.go -v
   # 出力: PASS

## ステップ4: Refactor - コード改善
5. コードの品質向上
   - エラーハンドリングの改善
   - メモリ効率の最適化
   - チャネル管理の最適化

6. テスト実行（成功維持確認）
   go test ./pkg/data/... -v
   # 出力: PASS

参考: dataProvider.md設計書を参照
```

#### Phase 2: 市場データ管理（1 週間）

**3. Market（市場データ管理）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch pkg/market/market_test.md

## ステップ2: Red - 失敗するテストを作成
1. pkg/market/market_test.go を作成
   - TestMarket_NewMarket（初期化テスト）
   - TestMarket_Forward（時間進行テスト）
   - TestMarket_GetCurrentPrice（価格取得テスト）
   - TestMarket_IsFinished（終了判定テスト）

2. テスト実行（失敗確認）
   go test ./pkg/market/market_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/market/market.go を作成
   - Marketインターフェース定義
   - MarketImpl構造体の最小限実装
   - 必要なメソッドの最小限実装

4. テスト実行（成功確認）
   go test ./pkg/market/market_test.go -v
   # 出力: PASS

## ステップ4: Refactor - コード改善
5. コードの品質向上
   - 状態管理の最適化
   - DataProvider統合の改善
   - エラーハンドリングの強化

6. テスト実行（成功維持確認）
   go test ./pkg/market/... -v
   # 出力: PASS

参考: market.md設計書を参照
```

#### Phase 3: 取引実行（1-2 週間）

**4. Broker（取引実行）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch pkg/broker/broker_test.md

## ステップ2: Red - 失敗するテストを作成
1. pkg/broker/broker_test.go を作成
   - TestBroker_PlaceOrder（注文実行テスト）
   - TestBroker_GetPositions（ポジション管理テスト）
   - TestBroker_GetBalance（残高管理テスト）
   - TestBroker_MarketIntegration（Market統合テスト）

2. テスト実行（失敗確認）
   go test ./pkg/broker/broker_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/broker/broker.go を作成
   - Brokerインターフェース定義
   - SimpleBroker構造体の最小限実装
   - 必要なメソッドの最小限実装

4. テスト実行（成功確認）
   go test ./pkg/broker/broker_test.go -v
   # 出力: PASS

## ステップ4: Refactor - コード改善
5. コードの品質向上
   - 取引ロジックの最適化
   - Market連携の改善
   - スプレッド計算の精緻化

6. テスト実行（成功維持確認）
   go test ./pkg/broker/... -v
   # 出力: PASS

参考: broker.md設計書を参照
```

#### Phase 4: バックテスト統括（1 週間）

**5. Backtester（統括コンポーネント）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch pkg/backtester/backtester_test.md

## ステップ2: Red - 失敗するテストを作成
1. pkg/backtester/backtester_test.go を作成
   - TestBacktester_NewBacktester（初期化テスト）
   - TestBacktester_Forward（バックテストループテスト）
   - TestBacktester_BuySell（ユーザーAPIテスト）
   - TestBacktester_Integration（統合テスト）

2. テスト実行（失敗確認）
   go test ./pkg/backtester/backtester_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/backtester/backtester.go を作成
   - Backtester構造体の最小限実装
   - ユーザーAPI（Buy, Sell, Forward等）の最小限実装
   - Market・Broker統合の最小限実装

4. テスト実行（成功確認）
   go test ./pkg/backtester/backtester_test.go -v
   # 出力: PASS

## ステップ4: Refactor - コード改善
5. コードの品質向上
   - APIの使いやすさ改善
   - エラーハンドリングの統一
   - パフォーマンス最適化

6. テスト実行（成功維持確認）
   go test ./pkg/backtester/... -v
   # 出力: PASS

参考: backtester.md設計書を参照
```

#### Phase 5: 統計・レポート（1 週間）

**6. Statistics（統計計算・レポート）**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch pkg/statistics/calculator_test.md
touch pkg/statistics/report_test.md
touch pkg/statistics/metrics_test.md
touch pkg/statistics/formatter_test.md

## ステップ2: Red - 失敗するテストを作成
1. pkg/statistics/calculator_test.go を作成
   - TestCalculator_CalculateMetrics（PnL、勝率、シャープレシオ等）
   - TestCalculator_RiskMetrics（最大ドローダウン、VaR）
   - TestCalculator_AdvancedMetrics（カルマーレシオ、ソルティノレシオ）

2. テスト実行（失敗確認）
   go test ./pkg/statistics/calculator_test.go -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. pkg/statistics/calculator.go を作成
   - Calculator構造体の最小限実装
   - CalculateMetrics関数の最小限実装
   - 各種統計計算の最小限実装

4. テスト実行（成功確認）
   go test ./pkg/statistics/calculator_test.go -v
   # 出力: PASS

## ステップ4: 他コンポーネント実装
5. 同様のRed-Green-Refactorサイクルでその他コンポーネント実装
   - Metrics（メトリクス定義）
   - Report（レポート生成）
   - Formatter（数値フォーマット）

## ステップ5: Refactor - コード改善
6. 全体の品質向上
   - 計算ロジックの最適化
   - レポート品質の向上
   - パフォーマンス改善

7. 最終テスト実行（成功維持確認）
   go test ./pkg/statistics/... -v -cover
   # 出力: PASS, カバレッジ85%以上

参考: statistics.md設計書を参照
```

#### Phase 6: CLI・使用例（1 週間）

**7. CLI・Examples**

```bash
# t-wada式TDDステップ

## ステップ1: テスト仕様書作成
touch cmd/backtester/main_test.md

## ステップ2: Red - 失敗するテストを作成
1. cmd/backtester/main_test.go を作成
   - TestCLI_ArgumentParsing（CLI引数解析テスト）
   - TestCLI_ExecutionFlow（実行フローテスト）
   - TestCLI_ErrorHandling（エラーハンドリングテスト）

2. テスト実行（失敗確認）
   go test ./cmd/backtester/... -v
   # 出力: FAIL（実装が存在しないため）

## ステップ3: Green - 最小限の実装
3. cmd/backtester/main.go を作成
   - CLI引数解析の最小限実装
   - 基本的な実行フローの最小限実装

4. テスト実行（成功確認）
   go test ./cmd/backtester/... -v
   # 出力: PASS

## ステップ4: Examples作成
5. examples/*.go を作成
   - 基本的な使用方法の例
   - 高度な戦略実装例
   - 各種設定パターンの例

## ステップ5: Refactor - コード改善
6. ユーザビリティの向上
   - CLIオプションの充実
   - ヘルプメッセージの改善
   - エラーメッセージの分かりやすさ向上

7. 最終テスト実行（成功維持確認）
   go test ./... -v
   # 出力: 全パッケージでPASS

8. テスト仕様書の更新
   # 実装完了後、各xxx_test.mdファイルを実際のテストコードに合わせて更新
   # 実装したテスト関数、テストケース、カバレッジ情報を記載
```

## 4. テスト仕様書作成ガイドライン

### 4.1 テスト仕様書の目的
- **実装されたテストコードの可読性向上**: xxx_test.goの内容をマークダウンで分かりやすく解説
- **テストケースの網羅性確認**: どのようなテストが実装されているかの可視化
- **新規開発者へのガイド**: テストコードの理解を助けるドキュメント
- **品質保証**: テストカバレッジと実装内容の記録

### 4.2 作成タイミング
```
1. Red段階: テストコード作成時にテスト仕様書も作成（テスト設計書として）
2. Green段階: 最小限実装後、テスト仕様書を実装に合わせて調整
3. Refactor段階: 最終的にテスト仕様書を実装されたコードに合わせて完成
```

### 4.3 必須記載項目
- **概要**: テスト対象とテストの目的
- **実装されているテスト関数**: xxx_test.goに含まれる全関数のリスト
- **テスト関数詳細**: 各関数の実装内容とテストケース
- **実装済みテストの概要**: 正常系・異常系・境界値テスト数とカバレッジ
- **テスト実行方法**: 具体的なコマンド例

### 4.4 品質基準
- **正確性**: 実装されたコードと仕様書の内容が一致している
- **完全性**: すべてのテスト関数が仕様書に記載されている
- **明確性**: 新規開発者がテストの目的と内容を理解できる
- **保守性**: コード変更時に仕様書も容易に更新できる構成

## 例: candle_test.md
Phase 1で実装されたpkg/models/candle_test.mdを参考として、同様の形式で各コンポーネントのテスト仕様書を作成してください。
