# タイムフレーム選択機能 設計書

## 概要

Visual Modeのフロントエンドにタイムフレーム選択機能を追加し、ユーザーが複数の時間軸（1M、5M、15M、1H、4H、1D、1W）でチャートを表示できるようにする。

## 要件

### 機能要件
- 複数のタイムフレームから選択可能（1M、5M、15M、1H、4H、1D、1W）
- 現在選択中のタイムフレームを視覚的に識別可能
- タイムフレーム変更時にチャートが即座に更新
- 各タイムフレームのデータを個別に管理

### 非機能要件
- レスポンシブデザイン
- アクセシブルなUI
- 高速な切り替え（100ms以内）
- メモリ効率的なデータ管理

## UI設計

### タイムフレーム選択UI
- 位置: ControlPanelの右側に配置
- 表示: 横並びのボタン群
- スタイル: 既存のPlayPauseButtonと統一感のあるデザイン
- 状態: 選択中のボタンはハイライト表示

```
[Play/Pause] [Speed: ====|====] [1M] [5M] [15M] [1H] [4H] [1D] [1W]
```

### タイムフレーム定義
| 表示名 | 値 | 説明 |
|--------|-----|------|
| 1M | 1m | 1分足 |
| 5M | 5m | 5分足 |
| 15M | 15m | 15分足 |
| 1H | 1h | 1時間足 |
| 4H | 4h | 4時間足 |
| 1D | 1d | 日足 |
| 1W | 1w | 週足 |

## 技術設計

### 状態管理（Jotai）

```typescript
// タイムフレーム選択状態
const timeframeAtom = atom<string>('1m');

// タイムフレーム別ローソク足データ
const candleDataByTimeframeAtom = atom<Record<string, CandlestickData<Time>[]>>({
  '1m': [],
  '5m': [],
  '15m': [],
  '1h': [],
  '4h': [],
  '1d': [],
  '1w': []
});

// 現在表示中のローソク足データ（派生状態）
const currentCandleDataAtom = atom<CandlestickData<Time>[]>((get) => {
  const timeframe = get(timeframeAtom);
  const allData = get(candleDataByTimeframeAtom);
  return allData[timeframe] || [];
});
```

### コンポーネント構造

```typescript
// タイムフレーム選択コンポーネント
const TimeframeSelector = () => {
  const [selectedTimeframe, setSelectedTimeframe] = useAtom(timeframeAtom);
  
  const handleTimeframeChange = (timeframe: string) => {
    setSelectedTimeframe(timeframe);
    // WebSocketでバックエンドに通知（将来的に）
  };
  
  return (
    <TimeframeSelectorContainer>
      {TIMEFRAMES.map(timeframe => (
        <TimeframeButton
          key={timeframe.value}
          active={selectedTimeframe === timeframe.value}
          onClick={() => handleTimeframeChange(timeframe.value)}
        >
          {timeframe.label}
        </TimeframeButton>
      ))}
    </TimeframeSelectorContainer>
  );
};
```

### スタイリング（styled-components）

```typescript
const TimeframeSelectorContainer = styled.div`
  display: flex;
  gap: 4px;
  align-items: center;
`;

const TimeframeButton = styled.button<{ active: boolean }>`
  padding: 6px 12px;
  background: ${props => props.active ? '#4caf50' : '#666'};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  min-width: 36px;
  
  &:hover {
    background: ${props => props.active ? '#45a049' : '#777'};
  }
  
  &:disabled {
    background: #444;
    cursor: not-allowed;
  }
`;
```

## データフロー

### 1. タイムフレーム選択
1. ユーザーがタイムフレームボタンをクリック
2. `timeframeAtom`の状態が更新
3. `currentCandleDataAtom`が自動的に再計算
4. チャートが新しいデータで再描画

### 2. データ受信（将来的な拡張）
1. WebSocketでタイムフレーム変更をバックエンドに通知
2. バックエンドが対応するタイムフレームのデータを送信
3. フロントエンドが該当するタイムフレームのデータを更新

## 実装フェーズ

### フェーズ1: 基本UI実装
- [ ] タイムフレーム選択UIの追加
- [ ] 状態管理の実装
- [ ] チャート表示の切り替え
- [ ] スタイリングの完成

### フェーズ2: データ管理改善
- [ ] タイムフレーム別データ分離
- [ ] メモリ効率化
- [ ] データ永続化（localStorage）

### フェーズ3: バックエンド連携（将来的）
- [ ] WebSocketコマンド拡張
- [ ] バックエンドでのタイムフレーム処理
- [ ] リアルタイムデータ配信

## 制約事項

### 現在の制約
- バックエンドは1つのタイムフレームのデータのみ送信
- フロントエンドでのタイムフレーム変換は行わない
- 過去データの取得機能なし

### 将来的な改善点
- バックエンドでの複数タイムフレーム対応
- 過去データの動的取得
- タイムフレーム間のデータ変換アルゴリズム
- キャッシュ機能の実装

## テスト計画

### 単体テスト
- [ ] タイムフレーム選択の状態管理
- [ ] データ切り替えの動作確認
- [ ] UI操作のテスト

### 統合テスト
- [ ] WebSocket通信（将来的）
- [ ] チャート表示の正確性
- [ ] パフォーマンステスト

### ユーザビリティテスト
- [ ] タイムフレーム切り替えの直感性
- [ ] 視覚的フィードバックの適切性
- [ ] レスポンシブデザインの確認

## 参考資料

- [TradingView タイムフレーム選択UI](https://www.tradingview.com/)
- [Lightweight Charts ドキュメント](https://tradingview.github.io/lightweight-charts/)
- [Jotai 状態管理](https://jotai.org/)