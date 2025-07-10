# Visual Mode Example

このサンプルは、FXバックテストライブラリのビジュアルモード機能を実演します。

## 概要

`visualize_example.go`は以下の機能を実演します：

- ✅ **Visualizerサーバー**: WebSocketベースのリアルタイムデータ配信
- ✅ **Backtester連携**: バックテストエンジンとVisualizerの統合
- ✅ **サンプル戦略**: 10期移動平均クロス戦略
- ✅ **リアルタイム可視化**: チャート、取引、統計情報の表示

## 必要条件

1. **Goのインストール** (1.22+)
2. **Node.jsのインストール** (18+)
3. **FXヒストリカルデータ**: CSV形式のローソク足データ

## セットアップ手順

### 1. データファイルの準備

```bash
# testdataディレクトリにUSDJPYデータを配置
cp your_data.csv testdata/USDJPY_2024_01.csv
```

データファイルの形式：
```csv
timestamp,open,high,low,close,volume
2024-01-01 00:00:00,150.123,150.145,150.098,150.134,1000
2024-01-01 00:01:00,150.134,150.156,150.112,150.143,1200
...
```

### 2. フロントエンドの起動

```bash
cd ../../frontend/visual-mode
npm install
npm run dev
```

フロントエンドが `http://localhost:5173` （またはポートが使用中の場合は別のポート）で起動します。

### 3. バックエンドの起動

```bash
cd examples/visualmode
go run main.go
```

## 実行結果

```
🚀 FX Backtesting Visual Mode Example
======================================
📡 Visualizer サーバーを開始中...
✅ Visualizer サーバーがポート 8080 で開始されました
🌐 ブラウザで http://localhost:3000 を開いてフロントエンドを確認してください
🤖 Backtester を初期化中...
✅ Backtester が初期化されました
📈 シンプル移動平均戦略を開始します
戦略: 現在価格が10期移動平均より上で買い、下で売り

🔄 バックテスト開始...
⏰ ステップ 100 処理完了 (現在時刻: 2024-01-01 01:40:00, 価格: 150.145)
📈 買いシグナル: 現在価格=150.156, MA=150.143
⏰ ステップ 200 処理完了 (現在時刻: 2024-01-01 03:20:00, 価格: 150.178)
📉 売りシグナル: 現在価格=150.132, MA=150.155

📊 === 統計情報 ===
現在残高: 100250.00
オープンポジション数: 0
総取引数: 1
総損益: 250.00
勝率: 100.0%
==================
```

## 機能詳細

### VisualizerAdapter

型の不一致を解決するアダプターパターンを実装：

```go
type VisualizerAdapter struct {
    visualizer visualizer.Visualizer
}

func (va *VisualizerAdapter) OnBacktestStateChange(state backtester.BacktestState) error {
    vizState := visualizer.BacktestState(state)
    return va.visualizer.OnBacktestStateChange(vizState)
}
```

### SimpleMovingAverageStrategy

シンプルな移動平均クロス戦略：

```go
type SimpleMovingAverageStrategy struct {
    backtester *backtester.Backtester
    prices     []float64
    windowSize int
}
```

**戦略ロジック:**
- 現在価格 > 移動平均: 買いシグナル
- 現在価格 < 移動平均: 売りシグナル（決済）

### リアルタイム通知

バックテスト実行中に以下のイベントが自動的にフロントエンドに送信されます：

1. **ローソク足データ**: 各Forward()実行時
2. **取引イベント**: Buy/Sell/ClosePosition実行時
3. **統計情報**: 残高、損益、勝率等
4. **状態変更**: バックテスト開始/終了時

## WebSocket通信

### エンドポイント
- **WebSocket**: `ws://localhost:8080/ws`
- **ヘルスチェック**: `http://localhost:8080/health`

### メッセージ形式

```json
{
  "type": "candle_update",
  "data": {
    "timestamp": "2024-01-01T00:00:00Z",
    "open": 150.123,
    "high": 150.145,
    "low": 150.098,
    "close": 150.134,
    "volume": 1000
  },
  "timestamp": "2024-01-01T00:00:01Z"
}
```

```json
{
  "type": "trade_event",
  "data": {
    "id": "buy-USDJPY-1234567890",
    "symbol": "USDJPY",
    "side": 0,
    "size": 1000,
    "entry_price": 150.123,
    "exit_price": 150.156,
    "pnl": 330.0,
    "status": 1
  },
  "timestamp": "2024-01-01T00:00:01Z"
}
```

## カスタマイズ

### データファイルの変更

```go
dataConfig := models.DataProviderConfig{
    FilePath: "./testdata/YOUR_DATA.csv", // ← ここを変更
    Format:   "csv",
}
```

### ブローカー設定の調整

```go
brokerConfig := models.BrokerConfig{
    InitialBalance: 100000.0, // 初期残高
    Spread:         0.0001,   // スプレッド (0.1 pips)
}
```

### 戦略パラメータの調整

```go
strategy := NewSimpleMovingAverageStrategy(bt, 20) // ← 移動平均期間を変更
```

### 更新頻度の調整

```go
// 可視化のための待機時間
time.Sleep(50 * time.Millisecond) // ← ここを変更
```

## トラブルシューティング

### 1. 404 Page Not Found エラー

```
404 Page Not Found - http://localhost:8080 を開こうとしている
```

**解決方法**: 
- フロントエンドを先に起動してください: `cd ../../frontend/visual-mode && npm run dev`
- ブラウザで表示されるURL（通常 `http://localhost:5173` や `http://localhost:5174`）を開いてください（8080ではありません）
- ポート8080はWebSocketサーバー用で、ウェブページは提供していません

### 2. データファイルが見つからない

```
Backtester初期化エラー: failed to initialize market: file not found: ./testdata/USDJPY_2024_01.csv
```

**解決方法**: データファイルパスを確認し、正しいファイルを配置してください。

### 3. WebSocket接続エラー

```
WebSocket error: dial tcp 127.0.0.1:8080: connect: connection refused
```

**解決方法**: Visualizerサーバーが正常に起動していることを確認してください。

### 4. 取引エラー

```
⚠️ 戦略実行エラー: invalid symbol or price: USDJPY
```

**解決方法**: データファイルのシンボル名とコード内のシンボル名が一致していることを確認してください。

## 次のステップ

1. **カスタム戦略の実装**: より複雑な取引ロジックの追加
2. **パフォーマンス最適化**: 大量データでの処理速度向上
3. **UI機能拡張**: フロントエンドでの追加機能実装
4. **リスク管理**: ストップロス、テイクプロフィットの実装

## 参考資料

- [Visual Mode Design Document](../docs/visual_mode_design.md)
- [Visualizer Component Design](../docs/visualizer.md)
- [Backtester Documentation](../docs/backtester.md)