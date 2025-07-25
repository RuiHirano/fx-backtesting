import { atom, useAtom } from "jotai";
import {
  type CandlestickData,
  type ChartOptions,
  ColorType,
  type DeepPartial,
  type Time,
} from "lightweight-charts";
import { CandlestickSeries, Chart } from "lightweight-charts-react-components";
import { useEffect, useRef, useState } from "react";
import styled from "styled-components";

interface Trade {
  id: string;
  symbol: string;
  type: "buy" | "sell";
  amount: number;
  price: number;
  timestamp: string;
}

interface WebSocketMessage {
  type: "candle_update" | "trade_event" | "statistics_update" | "ping" | "pong";
  data?: any;
  timestamp?: string;
  message?: string;
}

interface ConnectionState {
  isConnected: boolean;
  error?: string;
}

// Timeframe definition
interface Timeframe {
  label: string;
  value: string;
}

const TIMEFRAMES: Timeframe[] = [
  { label: "1M", value: "1m" },
  { label: "5M", value: "5m" },
  { label: "15M", value: "15m" },
  { label: "1H", value: "1h" },
  { label: "4H", value: "4h" },
  { label: "1D", value: "1d" },
  { label: "1W", value: "1w" },
];

// Jotai atoms for state management
const connectionStateAtom = atom<ConnectionState>({ isConnected: false });
const timeframeAtom = atom<string>("1m");
const candleDataByTimeframeAtom = atom<Record<string, CandlestickData<Time>[]>>({
  "1m": [],
  "5m": [],
  "15m": [],
  "1h": [],
  "4h": [],
  "1d": [],
  "1w": [],
});
const tradesAtom = atom<Trade[]>([]);
const playbackStateAtom = atom<{ isPlaying: boolean; speed: number }>({
  isPlaying: false,
  speed: 1,
});

// Derived atom for current candle data
const currentCandleDataAtom = atom<CandlestickData<Time>[]>((get) => {
  const timeframe = get(timeframeAtom);
  const allData = get(candleDataByTimeframeAtom);
  return allData[timeframe] || [];
});

const ChartContainer = styled.div`
  width: 800px;
  height: 400px;
  display: flex;
  flex-direction: column;
`;

const ConnectionStatus = styled.div<{ isConnected: boolean }>`
  position: absolute;
  top: 10px;
  right: 10px;
  padding: 8px 12px;
  background: ${(props) => (props.isConnected ? "#4caf50" : "#f44336")};
  color: white;
  border-radius: 4px;
  font-size: 12px;
  z-index: 1000;
`;

const ErrorMessage = styled.div`
  position: absolute;
  top: 50px;
  right: 10px;
  padding: 8px 12px;
  background: #ff9800;
  color: white;
  border-radius: 4px;
  font-size: 12px;
  z-index: 1000;
  max-width: 300px;
`;

const StatusPanel = styled.div`
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  border-radius: 4px;
  font-size: 12px;
  z-index: 1000;
  min-width: 200px;
`;

const ControlPanel = styled.div`
  position: absolute;
  bottom: 10px;
  left: 10px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.8);
  color: white;
  border-radius: 4px;
  font-size: 12px;
  z-index: 1000;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 600px;
  flex-wrap: wrap;
`;

const PlayPauseButton = styled.button`
  padding: 8px 16px;
  background: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  min-width: 80px;

  &:hover {
    background: #45a049;
  }

  &:disabled {
    background: #666;
    cursor: not-allowed;
  }
`;

const SpeedSlider = styled.input`
  width: 100px;
  height: 4px;
  border-radius: 2px;
  background: #ddd;
  outline: none;

  &::-webkit-slider-thumb {
    appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #4caf50;
    cursor: pointer;
  }

  &::-moz-range-thumb {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #4caf50;
    cursor: pointer;
    border: none;
  }
`;

const SpeedLabel = styled.span`
  font-size: 12px;
  color: #ccc;
  min-width: 60px;
`;

const TimeframeSelectorContainer = styled.div`
  display: flex;
  gap: 4px;
  align-items: center;
  margin-left: 16px;
`;

const TimeframeButton = styled.button<{ active: boolean }>`
  padding: 6px 12px;
  background: ${(props) => (props.active ? "#4caf50" : "#666")};
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  min-width: 36px;

  &:hover {
    background: ${(props) => (props.active ? "#45a049" : "#777")};
  }

  &:disabled {
    background: #444;
    cursor: not-allowed;
  }
`;

const TimeframeLabel = styled.span`
  font-size: 12px;
  color: #ccc;
  margin-right: 8px;
`;

function App() {
  const ws = useRef<WebSocket | null>(null);

  const [connectionState, setConnectionState] = useAtom(connectionStateAtom);
  const [timeframe, setTimeframe] = useAtom(timeframeAtom);
  const [candleDataByTimeframe, setCandleDataByTimeframe] = useAtom(candleDataByTimeframeAtom);
  const [currentCandleData] = useAtom(currentCandleDataAtom);
  const [trades, setTrades] = useAtom(tradesAtom);
  const [playbackState, setPlaybackState] = useAtom(playbackStateAtom);
  const [statistics, setStatistics] = useState<any>(null);

  useEffect(() => {
    const connectWebSocket = () => {
      try {
        console.log("Attempting to connect to WebSocket...");
        ws.current = new WebSocket("ws://localhost:8080/ws");

        ws.current.onopen = () => {
          console.log("WebSocket connected successfully");
          setConnectionState({ isConnected: true });
          if (ws.current) {
            ws.current.send(
              JSON.stringify({ type: "ping", message: "Hello from React" })
            );
          }
        };

        ws.current.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            console.log("Received message:", message);

            switch (message.type) {
              case "candle_update":
                if (message.data) {
                  const candle = message.data;
                  const timeInSeconds = Math.floor(
                    new Date(candle.timestamp || candle.time).getTime() / 1000
                  );

                  const formattedCandle: CandlestickData<Time> = {
                    time: timeInSeconds as Time,
                    open: candle.open,
                    high: candle.high,
                    low: candle.low,
                    close: candle.close,
                  };

                  setCandleDataByTimeframe((prev) => {
                    const currentTimeframeData = prev[timeframe] || [];
                    const updatedData = [...currentTimeframeData, formattedCandle];
                    // Keep only the last 10000 candles
                    return {
                      ...prev,
                      [timeframe]: updatedData.slice(-10000),
                    };
                  });
                }
                break;

              case "trade_event":
                if (message.data) {
                  const trade: Trade = {
                    id: message.data.id,
                    symbol: message.data.symbol,
                    type: message.data.side === 0 ? "buy" : "sell",
                    amount: message.data.size,
                    price: message.data.entry_price || message.data.exit_price,
                    timestamp: message.timestamp || new Date().toISOString(),
                  };
                  setTrades((prev) => [...prev, trade]);
                }
                break;

              case "statistics_update":
                if (message.data) {
                  setStatistics(message.data);
                }
                break;

              case "pong":
                console.log("Received pong:", message.message);
                break;

              default:
                console.log("Unknown message type:", message.type);
            }
          } catch (error) {
            console.error("Error parsing WebSocket message:", error);
            setConnectionState((prev) => ({
              ...prev,
              error: "Message parsing error",
            }));
          }
        };

        ws.current.onclose = (event) => {
          console.log("WebSocket disconnected:", event);
          setConnectionState({ isConnected: false });
          setTimeout(() => {
            if (ws.current?.readyState === WebSocket.CLOSED) {
              console.log("Attempting to reconnect...");
              connectWebSocket();
            }
          }, 3000);
        };

        ws.current.onerror = (error) => {
          console.error("WebSocket error:", error);
          setConnectionState({ isConnected: false, error: "Connection error" });
        };
      } catch (error) {
        console.error("Failed to create WebSocket connection:", error);
        setConnectionState({ isConnected: false, error: "Failed to connect" });
      }
    };

    connectWebSocket();

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [setConnectionState, setTrades, setCandleDataByTimeframe, timeframe]);

  const handlePlayPause = () => {
    if (!connectionState.isConnected) return;

    const newIsPlaying = !playbackState.isPlaying;
    setPlaybackState((prev) => ({ ...prev, isPlaying: newIsPlaying }));

    if (ws.current) {
      const command = {
        type: newIsPlaying ? "play" : "pause",
        data: { speed: playbackState.speed },
        client_id: "react-client",
        timestamp: new Date().toISOString(),
      };
      ws.current.send(JSON.stringify(command));
    }
  };

  const handleSpeedChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const speed = parseFloat(event.target.value);
    setPlaybackState((prev) => ({ ...prev, speed }));

    if (ws.current && connectionState.isConnected) {
      const command = {
        type: "speed_change",
        data: { speed },
        client_id: "react-client",
        timestamp: new Date().toISOString(),
      };
      ws.current.send(JSON.stringify(command));
    }
  };

  const getSpeedLabel = (speed: number) => {
    if (speed < 1) return `${speed}x`;
    if (speed === 1) return "1x";
    return `${speed}x`;
  };

  const handleTimeframeChange = (newTimeframe: string) => {
    setTimeframe(newTimeframe);
    
    // Future: Send WebSocket command to backend
    // if (ws.current && connectionState.isConnected) {
    //   const command = {
    //     type: "timeframe_change",
    //     data: { timeframe: newTimeframe },
    //     client_id: "react-client",
    //     timestamp: new Date().toISOString(),
    //   };
    //   ws.current.send(JSON.stringify(command));
    // }
  };

  const options: DeepPartial<ChartOptions> = {
    layout: {
      background: {
        type: ColorType.Solid,
        color: "white",
      },
      textColor: "black",
      attributionLogo: false,
    },
    grid: {
      vertLines: {
        color: "rgba(197, 203, 206, 0.5)",
      },
      horzLines: {
        color: "rgba(197, 203, 206, 0.5)",
      },
    },
    rightPriceScale: {
      borderColor: "rgba(197, 203, 206, 0.8)",
      scaleMargins: {
        top: 0.1,
        bottom: 0.1,
      },
    },
    timeScale: {
      borderColor: "rgba(197, 203, 206, 0.8)",
      barSpacing: 4,
      visible: true,
      timeVisible: true,
    },
  };

  return (
    <ChartContainer>
      <ConnectionStatus isConnected={connectionState.isConnected}>
        {connectionState.isConnected ? "Connected" : "Disconnected"}
      </ConnectionStatus>

      {connectionState.error && (
        <ErrorMessage>Error: {connectionState.error}</ErrorMessage>
      )}

      <StatusPanel>
        <div>Timeframe: {timeframe.toUpperCase()}</div>
        <div>Candles: {currentCandleData.length}</div>
        <div>Trades: {trades.length}</div>
        {statistics && (
          <>
            <div>Balance: ${statistics.current_balance?.toFixed(2) || 0}</div>
            <div>Total Trades: {statistics.total_trades || 0}</div>
            <div>Win Rate: {statistics.win_rate?.toFixed(1) || 0}%</div>
          </>
        )}
      </StatusPanel>
      <Chart options={options} containerProps={{ style: { flexGrow: "1" } }}>
        <CandlestickSeries data={currentCandleData} />
      </Chart>

      <ControlPanel>
        <PlayPauseButton
          onClick={handlePlayPause}
          disabled={!connectionState.isConnected}
        >
          {playbackState.isPlaying ? "Pause" : "Play"}
        </PlayPauseButton>

        <SpeedLabel>Speed:</SpeedLabel>
        <SpeedSlider
          type="range"
          min="0.1"
          max="5"
          step="0.1"
          value={playbackState.speed}
          onChange={handleSpeedChange}
          disabled={!connectionState.isConnected}
        />
        <SpeedLabel>{getSpeedLabel(playbackState.speed)}</SpeedLabel>
        
        <TimeframeSelectorContainer>
          <TimeframeLabel>Timeframe:</TimeframeLabel>
          {TIMEFRAMES.map((tf) => (
            <TimeframeButton
              key={tf.value}
              active={timeframe === tf.value}
              onClick={() => handleTimeframeChange(tf.value)}
              disabled={!connectionState.isConnected}
            >
              {tf.label}
            </TimeframeButton>
          ))}
        </TimeframeSelectorContainer>
      </ControlPanel>
    </ChartContainer>
  );
}

export default App;
