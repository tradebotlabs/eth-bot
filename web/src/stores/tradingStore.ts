import { create } from 'zustand';
import type {
  AccountStats,
  Position,
  Trade,
  Candle,
  Strategy,
  Signal,
  IndicatorValues,
  MarketRegime,
  Config,
  StatusUpdate,
} from '../types';

export interface LogEntry {
  id: string;
  timestamp: Date;
  type: 'info' | 'success' | 'warning' | 'error' | 'signal' | 'trade';
  message: string;
  details?: string;
}

interface TradingState {
  // Connection status
  wsConnected: boolean;
  setWsConnected: (connected: boolean) => void;

  // Logs
  logs: LogEntry[];
  addLog: (log: Omit<LogEntry, 'id' | 'timestamp'>) => void;
  clearLogs: () => void;

  // Trading status
  isRunning: boolean;
  mode: 'paper' | 'live';
  symbol: string;
  currentPrice: number;
  setStatus: (status: Partial<StatusUpdate>) => void;

  // Account
  accountStats: AccountStats | null;
  setAccountStats: (stats: AccountStats) => void;

  // Positions
  positions: Position[];
  setPositions: (positions: Position[]) => void;
  addPosition: (position: Position) => void;
  updatePosition: (id: string, update: Partial<Position>) => void;
  removePosition: (id: string) => void;

  // Trades
  trades: Trade[];
  setTrades: (trades: Trade[]) => void;
  addTrade: (trade: Trade) => void;

  // Market data
  candles: Map<string, Candle[]>;
  setCandles: (timeframe: string, candles: Candle[]) => void;
  addCandle: (timeframe: string, candle: Candle) => void;
  updateCandle: (timeframe: string, candle: Candle) => void;

  // Indicators
  indicators: IndicatorValues | null;
  setIndicators: (indicators: IndicatorValues) => void;

  // Market regime
  regime: MarketRegime | null;
  setRegime: (regime: MarketRegime) => void;

  // Strategies
  strategies: Strategy[];
  setStrategies: (strategies: Strategy[]) => void;
  updateStrategy: (name: string, update: Partial<Strategy>) => void;

  // Signals
  signals: Signal[];
  setSignals: (signals: Signal[]) => void;
  addSignal: (signal: Signal) => void;

  // Config
  config: Config | null;
  setConfig: (config: Config) => void;
}

export const useTradingStore = create<TradingState>((set) => ({
  // Connection status
  wsConnected: false,
  setWsConnected: (connected) => set({ wsConnected: connected }),

  // Logs
  logs: [],
  addLog: (log) =>
    set((state) => ({
      logs: [
        {
          ...log,
          id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
          timestamp: new Date(),
        },
        ...state.logs,
      ].slice(0, 100),
    })),
  clearLogs: () => set({ logs: [] }),

  // Trading status
  isRunning: false,
  mode: 'paper',
  symbol: 'ETHUSDT',
  currentPrice: 0,
  setStatus: (status) =>
    set((state) => ({
      isRunning: status.running ?? state.isRunning,
      mode: status.mode ?? state.mode,
      symbol: status.symbol ?? state.symbol,
      currentPrice: status.currentPrice ?? state.currentPrice,
    })),

  // Account
  accountStats: null,
  setAccountStats: (stats) => set({ accountStats: stats }),

  // Positions
  positions: [],
  setPositions: (positions) => set({ positions }),
  addPosition: (position) =>
    set((state) => ({ positions: [...state.positions, position] })),
  updatePosition: (id, update) =>
    set((state) => ({
      positions: state.positions.map((p) =>
        p.id === id ? { ...p, ...update } : p
      ),
    })),
  removePosition: (id) =>
    set((state) => ({
      positions: state.positions.filter((p) => p.id !== id),
    })),

  // Trades
  trades: [],
  setTrades: (trades) => set({ trades }),
  addTrade: (trade) =>
    set((state) => ({ trades: [trade, ...state.trades].slice(0, 100) })),

  // Market data
  candles: new Map(),
  setCandles: (timeframe, candles) =>
    set((state) => {
      const newCandles = new Map(state.candles);
      newCandles.set(timeframe, candles);
      return { candles: newCandles };
    }),
  addCandle: (timeframe, candle) =>
    set((state) => {
      const newCandles = new Map(state.candles);
      const existing = newCandles.get(timeframe) || [];
      newCandles.set(timeframe, [...existing, candle]);
      return { candles: newCandles };
    }),
  updateCandle: (timeframe, candle) =>
    set((state) => {
      const newCandles = new Map(state.candles);
      const existing = newCandles.get(timeframe) || [];
      if (existing.length > 0) {
        const lastIdx = existing.length - 1;
        if (existing[lastIdx].time === candle.time) {
          existing[lastIdx] = candle;
        } else {
          existing.push(candle);
        }
        newCandles.set(timeframe, [...existing]);
      }
      return { candles: newCandles };
    }),

  // Indicators
  indicators: null,
  setIndicators: (indicators) => set({ indicators }),

  // Market regime
  regime: null,
  setRegime: (regime) => set({ regime }),

  // Strategies
  strategies: [],
  setStrategies: (strategies) => set({ strategies }),
  updateStrategy: (name, update) =>
    set((state) => ({
      strategies: state.strategies.map((s) =>
        s.name === name ? { ...s, ...update } : s
      ),
    })),

  // Signals
  signals: [],
  setSignals: (signals) => set({ signals }),
  addSignal: (signal) =>
    set((state) => ({ signals: [signal, ...state.signals].slice(0, 50) })),

  // Config
  config: null,
  setConfig: (config) => set({ config }),
}));
