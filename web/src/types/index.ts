export interface Candle {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface Position {
  id: string;
  symbol: string;
  side: 'LONG' | 'SHORT';
  entryPrice: number;
  currentPrice: number;
  quantity: number;
  unrealizedPnl: number;
  unrealizedPnlPercent: number;
  stopLoss: number;
  takeProfit: number;
  entryTime: string;
  strategy: string;
}

export interface Trade {
  id: string;
  symbol: string;
  side: 'LONG' | 'SHORT';
  entryPrice: number;
  exitPrice: number;
  quantity: number;
  pnl: number;
  pnlPercent: number;
  commission: number;
  entryTime: string;
  exitTime: string;
  strategy: string;
  exitReason: string;
}

export interface AccountStats {
  equity: number;
  balance: number;
  unrealizedPnl: number;
  realizedPnl: number;
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRate: number;
  profitFactor: number;
  maxDrawdown: number;
  currentDrawdown: number;
  sharpeRatio: number;
  dailyPnl: number;
  weeklyPnl: number;
}

export interface IndicatorValues {
  rsi: number;
  macd: {
    macd: number;
    signal: number;
    histogram: number;
  };
  bollinger: {
    upper: number;
    middle: number;
    lower: number;
    width: number;
  };
  adx: {
    adx: number;
    plusDI: number;
    minusDI: number;
  };
  atr: number;
  ema20: number;
  ema50: number;
  sma200: number;
}

export interface MarketRegime {
  regime: 'trending_up' | 'trending_down' | 'ranging' | 'volatile' | 'unknown';
  strength: number;
  volatility: number;
  trend: number;
}

export interface Strategy {
  id: string;
  name: string;
  enabled: boolean;
  description: string;
  parameters: Record<string, number>;
  performance: {
    totalTrades: number;
    winRate: number;
    profitFactor: number;
    totalPnl: number;
  };
}

export interface Signal {
  strategy: string;
  direction: 'LONG' | 'SHORT' | 'NONE';
  price: number;
  stopLoss: number;
  takeProfit: number;
  confidence: number;
  timestamp: string;
}

export interface RiskConfig {
  maxPositionSize: number;
  maxRiskPerTrade: number;
  maxDailyLoss: number;
  maxWeeklyLoss: number;
  maxDrawdown: number;
  maxOpenPositions: number;
  maxLeverage: number;
  enableCircuitBreaker: boolean;
  consecutiveLossLimit: number;
  haltDurationHours: number;
  minRiskRewardRatio: number;
}

export interface TradingConfig {
  symbol: string;
  mode: 'paper' | 'live';
  timeframes: string[];
  primaryTimeframe: string;
  initialBalance: number;
  commission: number;
  slippage: number;
}

export interface IndicatorConfig {
  rsiPeriod: number;
  macdFast: number;
  macdSlow: number;
  macdSignal: number;
  bbPeriod: number;
  bbStdDev: number;
  adxPeriod: number;
  atrPeriod: number;
}

export interface StrategiesConfig {
  enabled: string[];
}

export interface Config {
  trading: TradingConfig;
  risk: RiskConfig;
  indicators: IndicatorConfig;
  strategies: StrategiesConfig;
}

export interface BacktestConfig {
  symbol: string;
  interval: string;
  startDate: string;
  endDate: string;
  initialCapital: number;
  commission: number;
  slippage: number;
  riskPerTrade: number;
  strategies: string[];
}

export interface BacktestResult {
  id: string;
  config: BacktestConfig;
  metrics: {
    totalReturn: number;
    annualizedReturn: number;
    sharpeRatio: number;
    sortinoRatio: number;
    maxDrawdown: number;
    winRate: number;
    profitFactor: number;
    totalTrades: number;
    winningTrades: number;
    losingTrades: number;
    avgWin: number;
    avgLoss: number;
    expectancy: number;
    calmarRatio: number;
  };
  trades: Trade[];
  equityCurve: { time: string; equity: number; drawdown: number; return: number }[];
  monthlyReturns?: Record<string, number>;
  strategyStats?: Record<string, {
    name: string;
    totalTrades: number;
    winRate: number;
    profitFactor: number;
    netProfit: number;
    contribution: number;
  }>;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  error?: string;
  startTime: string;
  endTime?: string;
  executionTime?: string;
}

export interface WSMessage {
  type: string;
  data: unknown;
  timestamp: string;
}

export interface StatusUpdate {
  running: boolean;
  mode: 'paper' | 'live';
  symbol: string;
  currentPrice: number;
  equity: number;
  openPositions: number;
  todayPnl: number;
  regime: MarketRegime;
}

export interface TickerUpdate {
  symbol: string;
  price: number;
  change24h: number;
  volume24h: number;
  high24h: number;
  low24h: number;
}

export interface CandleUpdate {
  symbol: string;
  timeframe: string;
  candle: Candle;
}
