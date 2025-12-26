import axios from 'axios';
import type {
  Config,
  BacktestConfig,
  BacktestResult,
  Position,
  Candle,
  Strategy,
  RiskConfig,
} from '../types';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Dashboard
export const getDashboard = () => api.get('/dashboard');
export const getDashboardSummary = () => api.get('/dashboard/summary');
export const getEquityCurve = () => api.get('/dashboard/equity-curve');
export const getPerformance = () => api.get('/dashboard/performance');

// Trading State & Control
export const getTradingState = () => api.get('/trading/state');
export const startTrading = () => api.post('/trading/start');
export const stopTrading = () => api.post('/trading/stop');
export const pauseTrading = () => api.post('/trading/pause');
export const resumeTrading = () => api.post('/trading/resume');
export const getTradingMode = () => api.get('/trading/mode');
export const setTradingMode = (mode: 'paper' | 'live') =>
  api.post('/trading/mode', { mode });

// Positions
export const getPositions = () => api.get<Position[]>('/positions');
export const getPosition = (id: string) => api.get<Position>(`/positions/${id}`);
export const closePosition = (id: string) => api.post(`/positions/${id}/close`);
export const updateStopLoss = (id: string, stopLoss: number) =>
  api.put(`/positions/${id}/stop-loss`, { stopLoss });
export const updateTakeProfit = (id: string, takeProfit: number) =>
  api.put(`/positions/${id}/take-profit`, { takeProfit });

// Orders
export const getOrders = () => api.get('/orders');
export const getOpenOrders = () => api.get('/orders/open');
export const placeOrder = (order: { symbol: string; side: string; type: string; quantity: number; price?: number }) =>
  api.post('/orders', order);
export const cancelOrder = (id: string) => api.delete(`/orders/${id}`);

// Market Data (Candles, Ticker, Indicators)
export const getCandles = (symbol: string, timeframe: string, limit?: number) =>
  api.get<Candle[]>('/candles', {
    params: { symbol, timeframe, limit },
  });
export const getCandlesBySymbol = (symbol: string, timeframe: string) =>
  api.get<Candle[]>(`/candles/${symbol}/${timeframe}`);
export const getTicker = (symbol: string) =>
  api.get('/ticker', { params: { symbol } });
export const getIndicators = (symbol: string, timeframe: string) =>
  api.get('/indicators', { params: { symbol, timeframe } });
export const getRegime = () => api.get('/regime');

// Strategies
export const getStrategies = () => api.get<Strategy[]>('/strategies');
export const getStrategy = (name: string) => api.get<Strategy>(`/strategies/${name}`);
export const updateStrategy = (name: string, data: Partial<Strategy>) =>
  api.put(`/strategies/${name}`, data);
export const enableStrategy = (name: string) => api.post(`/strategies/${name}/enable`);
export const disableStrategy = (name: string) => api.post(`/strategies/${name}/disable`);
export const getSignals = (strategyName: string) => api.get(`/strategies/${strategyName}/signals`);

// Risk
export const getRiskStatus = () => api.get('/risk');
export const getRiskConfig = () => api.get<RiskConfig>('/risk/config');
export const updateRiskConfig = (config: Partial<RiskConfig>) =>
  api.put('/risk/config', config);
export const getRiskLimits = () => api.get('/risk/limits');
export const getDrawdown = () => api.get('/risk/drawdown');
export const getRiskEvents = () => api.get('/risk/events');
export const resetCircuitBreaker = () => api.post('/risk/circuit-breaker/reset');

// Settings
export const getSettings = () => api.get('/settings');
export const resetSettings = () => api.post('/settings/reset');
export const getTradingSettings = () => api.get('/settings/trading');
export const updateTradingSettings = (settings: Partial<Config['trading']>) =>
  api.put('/settings/trading', settings);
export const getIndicatorSettings = () => api.get('/settings/indicators');
export const updateIndicatorSettings = (settings: Partial<Config['indicators']>) =>
  api.put('/settings/indicators', settings);

// Config (combined settings)
export const getConfig = () => api.get<Config>('/settings');
export const updateConfig = (config: Partial<Config>) => api.put('/settings', config);

// Backtest
export const runBacktest = (config: BacktestConfig) =>
  api.post<{ id: string }>('/backtest', config);
export const getBacktestResults = () => api.get<BacktestResult[]>('/backtest/results');
export const getBacktestResult = (id: string) =>
  api.get<BacktestResult>(`/backtest/results/${id}`);

export default api;
