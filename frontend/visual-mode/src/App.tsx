import React, { useEffect, useRef, useState } from "react";
import {
  createChart,
  type IChartApi,
  ColorType,
  type ISeriesApi,
  CandlestickSeries,
  type Time,
} from "lightweight-charts";
import { atom, useAtom } from "jotai";
import styled from "styled-components";

interface CandleData {
  time: Time;
  open: number;
  high: number;
  low: number;
  close: number;
}

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
const candleDataAtom = atom<CandleData[]>([]);
const tradesAtom = atom<Trade[]>([]);

const ChartContainer = styled.div`
  width: 100%;
  height: 100vh;
  position: relative;
  background: #1e1e1e;
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

// Chart component following TradingView tutorial pattern
interface ChartProps {
  data: CandleData[];
}

const Chart: React.FC<ChartProps> = ({ data }) => {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candlestickSeriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);
  const [isChartReady, setIsChartReady] = useState(false);

  // Initialize chart
  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: "#1e1e1e" },
        textColor: "#d1d4dc",
      },
      grid: {
        vertLines: {
          color: "#2a2a2a",
        },
        horzLines: {
          color: "#2a2a2a",
        },
      },
      timeScale: {
        borderColor: "#485c7b",
        timeVisible: true,
        secondsVisible: false,
      },
      rightPriceScale: {
        borderColor: "#485c7b",
      },
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
    });

    chartRef.current = chart;

    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#26a69a",
      downColor: "#ef5350",
      borderVisible: false,
      wickUpColor: "#26a69a",
      wickDownColor: "#ef5350",
    });

    candlestickSeriesRef.current = candlestickSeries;
    setIsChartReady(true);

    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        chartRef.current.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
      }
    };

    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      if (chartRef.current) {
        chartRef.current.remove();
        chartRef.current = null;
        candlestickSeriesRef.current = null;
        setIsChartReady(false);
      }
    };
  }, []);

  // Update chart data
  useEffect(() => {
    if (!isChartReady || !candlestickSeriesRef.current || data.length === 0)
      return;

    try {
      candlestickSeriesRef.current.setData(data);
    } catch (error) {
      console.error("Error updating chart data:", error);
    }
  }, [data, isChartReady]);

  return (
    <div ref={chartContainerRef} style={{ width: "100vw", height: "100vh" }} />
  );
};

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

          // Send ping to test connection
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
                  // Convert timestamp to seconds for lightweight-charts
                  const timeInSeconds = Math.floor(
                    new Date(candle.timestamp || candle.time).getTime() / 1000
                  );

                  const formattedCandle: CandleData = {
                    time: timeInSeconds as Time,
                    open: candle.open,
                    high: candle.high,
                    low: candle.low,
                    close: candle.close,
                  };

                  setCandleData((prev) => {
                    const newData = [...prev, formattedCandle];
                    return newData.slice(-1000); // Keep last 1000 candles
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

          // Auto-reconnect after 3 seconds
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
  }, [setConnectionState, setCandleData, setTrades]);

  return (
    <ChartContainer>
      <ConnectionStatus isConnected={connectionState.isConnected}>
        {connectionState.isConnected ? "Connected" : "Disconnected"}
      </ConnectionStatus>

      {connectionState.error && (
        <ErrorMessage>Error: {connectionState.error}</ErrorMessage>
      )}

      <StatusPanel>
        <div>Candles: {candleData.length}</div>
        <div>Trades: {trades.length}</div>
        {statistics && (
          <>
            <div>Balance: ${statistics.current_balance?.toFixed(2) || 0}</div>
            <div>Total Trades: {statistics.total_trades || 0}</div>
            <div>Win Rate: {statistics.win_rate?.toFixed(1) || 0}%</div>
          </>
        )}
      </StatusPanel>

      <Chart data={candleData} />
    </ChartContainer>
  );
}

export default App;
