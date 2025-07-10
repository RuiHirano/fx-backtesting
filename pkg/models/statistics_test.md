# Statistics モジュール テストドキュメント

## 概要
このドキュメントは、`models.Statistics`構造体のテストケースとテスト戦略について説明します。

## テスト対象機能

### 1. 統計情報の作成 (`NewStatistics`)
- 初期残高を指定して統計情報を作成
- 初期値の正しい設定を確認

### 2. 残高更新 (`UpdateBalance`)
- 現在残高の更新
- 純利益の自動計算
- 最終更新時刻の更新

### 3. 取引追加 (`AddTrade`)
- 勝ち取引の追加
- 負け取引の追加
- 各種指標の自動計算

### 4. 指標計算 (`calculateMetrics`)
- 勝率計算
- プロフィットファクター計算
- 平均値計算（勝ち、負け、全体）

## テストケース

### TestNewStatistics
**目的**: 新しい統計情報の作成をテスト

**テストケース**:
1. `should create new statistics with initial balance`
   - 初期残高10,000で統計情報を作成
   - 初期値の正しい設定を確認
   - 取引回数が0であることを確認

**検証項目**:
- 統計情報オブジェクトが正しく作成される
- 初期残高と現在残高が同じ値に設定される
- 取引関連の初期値が0に設定される

### TestUpdateBalance
**目的**: 残高更新機能をテスト

**テストケース**:
1. `should update balance and net profit`
   - 初期残高10,000から11,000に更新
   - 純利益が1,000に計算される
   - 最終更新時刻が更新される

**検証項目**:
- 現在残高が正しく更新される
- 純利益が正しく計算される
- 最終更新時刻が更新される

### TestAddTrade
**目的**: 取引追加機能をテスト

**テストケース**:
1. `should add winning trade`
   - 利益100の勝ち取引を追加
   - 勝ち取引数が1に増加
   - 勝率が100%に計算される

2. `should add losing trade`
   - 損失-50の負け取引を追加
   - 負け取引数が1に増加
   - 勝率が0%に計算される

3. `should calculate metrics correctly`
   - 複数の取引を追加（勝ち: 100, 200、負け: -50）
   - 各種指標が正しく計算される

**検証項目**:
- 取引回数の正しい増加
- 勝ち/負け取引数の正しい分類
- 利益/損失の正しい累計
- 各種指標の正しい計算

### TestLastUpdated
**目的**: 最終更新時刻の管理をテスト

**テストケース**:
1. `should update last updated time`
   - 残高更新時に最終更新時刻が更新される
   - 更新前後の時刻が異なることを確認

**検証項目**:
- 最終更新時刻が正しく更新される
- 時刻の前後関係が正しい

## 指標計算の検証

### 勝率計算
```
勝率 = 勝ち取引数 / 総取引数 × 100
```

### プロフィットファクター
```
プロフィットファクター = 総利益 / 総損失の絶対値
```

### 平均値計算
```
平均勝ち = 総利益 / 勝ち取引数
平均負け = 総損失 / 負け取引数
平均利益 = 純利益 / 総取引数
```

## テスト実行

```bash
cd pkg/models
go test -v -run TestStatistics
```

## 期待される結果

全てのテストが成功し、以下のような出力が得られる：

```
=== RUN   TestNewStatistics
=== RUN   TestNewStatistics/should_create_new_statistics_with_initial_balance
--- PASS: TestNewStatistics (0.00s)
    --- PASS: TestNewStatistics/should_create_new_statistics_with_initial_balance (0.00s)
=== RUN   TestUpdateBalance
=== RUN   TestUpdateBalance/should_update_balance_and_net_profit
--- PASS: TestUpdateBalance (0.00s)
    --- PASS: TestUpdateBalance/should_update_balance_and_net_profit (0.00s)
=== RUN   TestAddTrade
=== RUN   TestAddTrade/should_add_winning_trade
=== RUN   TestAddTrade/should_add_losing_trade
=== RUN   TestAddTrade/should_calculate_metrics_correctly
--- PASS: TestAddTrade (0.00s)
    --- PASS: TestAddTrade/should_add_winning_trade (0.00s)
    --- PASS: TestAddTrade/should_add_losing_trade (0.00s)
    --- PASS: TestAddTrade/should_calculate_metrics_correctly (0.00s)
=== RUN   TestLastUpdated
=== RUN   TestLastUpdated/should_update_last_updated_time
--- PASS: TestLastUpdated (0.00s)
    --- PASS: TestLastUpdated/should_update_last_updated_time (0.00s)
PASS
```

## 今後のテスト拡張

1. **エッジケース**:
   - 0取引での指標計算
   - 非常に大きな値での計算
   - 負の初期残高での動作

2. **パフォーマンステスト**:
   - 大量の取引データでの処理速度
   - メモリ使用量の確認

3. **並行処理テスト**:
   - 複数ゴルーチンからの同時アクセス
   - データ競合の確認

## 依存関係

- `testing`パッケージ
- `time`パッケージ
- 外部依存関係なし（標準ライブラリのみ）