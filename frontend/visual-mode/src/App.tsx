import { atom, useAtom } from "jotai";
import {
  type CandlestickData,
  type ChartOptions,
  ColorType,
  type DeepPartial,
  type Time,
} from "lightweight-charts";
import {
  CandlestickSeries,
  Chart,
  LineSeries,
  TimeScale,
  TimeScaleFitContentTrigger,
} from "lightweight-charts-react-components";
import { useEffect, useRef, useState } from "react";
import styled from "styled-components";

const data = [
  { time: "2023-01-01", value: 100 },
  { time: "2023-01-02", value: 101 },
  { time: "2023-01-03", value: 102 },
];

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

// Jotai atoms for state management
const connectionStateAtom = atom<ConnectionState>({ isConnected: false });
const candleDataAtom = atom<CandlestickData<Time>[]>([]);
const tradesAtom = atom<Trade[]>([]);

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

function App() {
  const ws = useRef<WebSocket | null>(null);

  const [connectionState, setConnectionState] = useAtom(connectionStateAtom);
  const [candleData, setCandleData] = useAtom(candleDataAtom);
  const [trades, setTrades] = useAtom(tradesAtom);
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

                  setCandleData((prev) => {
                    const updatedData = [...prev, formattedCandle];
                    // Keep only the last 10000 candles
                    return updatedData.slice(-10000);
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
  }, [setConnectionState, setTrades]);

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
      <Chart options={options} containerProps={{ style: { flexGrow: "1" } }}>
        <CandlestickSeries data={candleData} />
      </Chart>
    </ChartContainer>
  );
}

export default App;
