# FX Backtesting Examples

このディレクトリには、FX バックテストライブラリの使用例が含まれています。

## 含まれる例

### 1. basic_example.go
基本的なバックテストの実行例です。

**機能:**
- 簡単な設定でのバックテスト実行
- 基本的な買い戦略の実装
- 結果の表示

**実行方法:**
```bash
go run basic_example.go
```

### 2. strategy_example.go
移動平均クロス戦略を使用した高度なバックテスト例です。

**機能:**
- 短期・長期移動平均の計算
- ゴールデンクロス/デッドクロス戦略
- 詳細な統計レポート生成

**実行方法:**
```bash
go run strategy_example.go
```

### 3. config_example.go
各種設定パターンの比較例です。

**機能:**
- 複数の設定での同時バックテスト
- 設定による結果の違いの比較
- 残高に応じた取引サイズの調整

**実行方法:**
```bash
go run config_example.go
```

## 前提条件

### データファイル
これらの例を実行するには、以下の場所にデータファイルが必要です：
- `../testdata/USDJPY_2024_01.csv`

### データファイル形式
CSV ファイルは以下の形式である必要があります：
```csv
timestamp,open,high,low,close,volume
2024-01-01 00:00:00,150.00,150.10,149.90,150.05,1000
2024-01-01 00:01:00,150.05,150.15,149.95,150.10,1200
...
```

## CLI アプリケーション

実際のデータでバックテストを実行するには、CLI アプリケーションを使用できます：

```bash
# CLI アプリケーションをビルド
go build -o backtester ../cmd/backtester/main.go

# 基本的な実行
./backtester -data ../testdata/USDJPY_2024_01.csv -config config.json

# JSON形式で結果を出力
./backtester -data ../testdata/USDJPY_2024_01.csv -config config.json -format json

# ファイルに結果を保存
./backtester -data ../testdata/USDJPY_2024_01.csv -config config.json -output results.txt
```

## 設定ファイル例

### config.json
```json
{
  "market": {
    "data_provider": {
      "file_path": "../testdata/USDJPY_2024_01.csv",
      "format": "csv"
    },
    "symbol": "USDJPY"
  },
  "broker": {
    "initial_balance": 100000.0,
    "spread": 0.01
  }
}
```

## カスタム戦略の実装

独自の戦略を実装するには、以下のパターンを参考にしてください：

```go
package main

import (
    "context"
    "github.com/RuiHirano/fx-backtesting/pkg/backtester"
    "github.com/RuiHirano/fx-backtesting/pkg/models"
)

func main() {
    // 1. 設定
    dataConfig := models.DataProviderConfig{
        FilePath: "your_data.csv",
        Format:   "csv",
    }
    brokerConfig := models.BrokerConfig{
        InitialBalance: 100000.0,
        Spread:         0.01,
    }
    
    // 2. Backtester作成・初期化
    bt := backtester.NewBacktester(dataConfig, brokerConfig)
    ctx := context.Background()
    bt.Initialize(ctx)
    
    // 3. 戦略ループ
    for !bt.IsFinished() {
        // 現在価格取得
        price := bt.GetCurrentPrice("USDJPY")
        
        // あなたの戦略ロジック
        if shouldBuy(price) {
            bt.Buy("USDJPY", 1000.0)
        }
        
        if shouldSell(price) {
            bt.Sell("USDJPY", 1000.0)
        }
        
        // 時間進行
        bt.Forward()
    }
    
    // 4. 結果確認
    finalBalance := bt.GetBalance()
    // ...
}

func shouldBuy(price float64) bool {
    // あなたの買い条件
    return false
}

func shouldSell(price float64) bool {
    // あなたの売り条件
    return false
}
```

## 注意事項

1. **データの品質**: 正確なバックテストには高品質なデータが必要です
2. **スプレッド**: 実際の取引コストを反映するようにスプレッドを設定してください
3. **オーバーフィッティング**: 過去のデータに最適化しすぎないよう注意してください
4. **リスク管理**: 実際の取引では適切なリスク管理が重要です

## サポート

質問や問題がある場合は、プロジェクトのメインドキュメントを参照してください。