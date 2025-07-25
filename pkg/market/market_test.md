# Market コンポーネント テスト計画書

## 概要

このドキュメントは、`Market`コンポーネントのテスト計画について記述します。テストは、設計書に記載された要件と機能を網羅し、コンポーネントの品質を保証することを目的とします。

## テストの種類

### 1. ユニットテスト (`market_test.go`)

ユニットテストは、`Market`コンポーネントの各機能が個別に正しく動作することを確認します。依存コンポーネントである`DataProvider`はモック化し、テストケースごとに特定のデータフローをシミュレートします。

### 2. 統合テスト

（将来のフェーズ）
DataProviderの実装（例：`CSVProvider`）と連携させ、実際のデータソースを用いたシナリオで`Market`コンポーネントが正しく動作することを確認します。

## ユニットテストのテストケース

### TestMarket_Initialize

| テストケースID | テスト内容 | 期待される結果 |
| :--- | :--- | :--- |
| INIT-001 | **正常系:** DataProviderから正常に初期キャッシュ分のデータが取得できる | - `initialized`フラグが`true`になる<br>- `candleCache`に`cacheSize`分のデータが格納される<br>- `currentIndex`が`0`になる<br>- `finished`フラグが`false`になる |
| INIT-002 | **準正常系:** DataProviderから取得できるデータが`cacheSize`より少ない | - `initialized`フラグが`true`になる<br>- `candleCache`に取得できた全データが格納される<br>- `currentIndex`が`0`になる<br>- `finished`フラグが`false`になる |
| INIT-003 | **準正常系:** DataProviderからデータが1件も取得できない | - `initialized`フラグが`true`になる<br>- `candleCache`が空になる<br>- `finished`フラグが`true`になる |
| INIT-004 | **異常系:** DataProviderの初期化でエラーが発生する | - `Initialize`がエラーを返す |

### TestMarket_Forward

| テストケースID | テスト内容 | 期待される結果 |
| :--- | :--- | :--- |
| FWD-001 | **正常系:** `Forward`を呼び出す | - `currentIndex`が1つインクリメントされる<br>- `GetCurrentTime`が次の足の時刻を返す<br>- `Forward`が`true`を返す |
| FWD-002 | **正常系:** キャッシュ補充の閾値に達した際に`Forward`を呼び出す | - `currentIndex`がインクリメントされる<br>- DataProviderから新しいデータが取得され、`candleCache`に追記される<br>- `Forward`が`true`を返す |
| FWD-003 | **準正常系:** 最後のデータで`Forward`を呼び出す | - `currentIndex`がインクリメントされる<br>- `finished`フラグが`true`になる<br>- `Forward`が`false`を返す |
| FWD-004 | **異常系:** `finished`状態で`Forward`を呼び出す | - `currentIndex`が変わらない<br>- `Forward`が`false`を返す |

### TestMarket_GetPrevCandles

| テストケースID | テスト内容 | 期待される結果 |
| :--- | :--- | :--- |
| GET-001 | **正常系:** `startTime`と`index`を指定し、過去のデータを取得する | - `startTime`から`index`の直前までの正しいローソク足のスライスが返される |
| GET-002 | **準正常系:** `startTime`が`index`の時刻より後の場合 | - 空のスライスが返される |
| GET-003 | **準正常系:** `index`が0の場合 | - 空のスライスが返される |
| GET-004 | **準正常系:** `startTime`に該当するデータがキャッシュにない場合 | - キャッシュの先頭から`index`の直前までのスライスが返される |
| GET-005 | **異常系:** `index`が範囲外（負数またはキャッシュサイズ以上）の場合 | - 空のスライスが返される |

### TestMarket_GetCurrentData

| テストケースID | テスト内容 | 期待される結果 |
| :--- | :--- | :--- |
| CUR-001 | **正常系:** `Forward`後に各種`GetCurrent`系メソッドを呼び出す | - `currentIndex`に対応する正しい`Candle`, `Price`, `Time`が返される |
| CUR-002 | **異常系:** 初期化前に`GetCurrent`系メソッドを呼び出す | - ゼロ値または`nil`が返される |

