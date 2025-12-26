import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
  BarChart,
  Bar,
  Cell,
  ReferenceLine,
} from 'recharts';
import * as api from '../services/api';
import type { BacktestConfig, BacktestResult, Strategy } from '../types';

// Custom hook for mobile detection
function useIsMobile() {
  const [isMobile, setIsMobile] = useState(false);
  useEffect(() => {
    const checkMobile = () => setIsMobile(window.innerWidth <= 768);
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);
  return isMobile;
}

// Generate mock monthly returns data
const generateMonthlyReturns = () => {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  return months.map(month => ({
    month,
    return: (Math.random() - 0.4) * 20,
  }));
};

// Generate mock trade distribution
const generateTradeDistribution = () => {
  const ranges = [
    { range: '-5%+', count: Math.floor(Math.random() * 5) },
    { range: '-3 to -5%', count: Math.floor(Math.random() * 10) },
    { range: '-1 to -3%', count: Math.floor(Math.random() * 20) },
    { range: '0 to -1%', count: Math.floor(Math.random() * 25) },
    { range: '0 to 1%', count: Math.floor(Math.random() * 30) },
    { range: '1 to 3%', count: Math.floor(Math.random() * 25) },
    { range: '3 to 5%', count: Math.floor(Math.random() * 15) },
    { range: '5%+', count: Math.floor(Math.random() * 8) },
  ];
  return ranges;
};

export function Backtest() {
  const queryClient = useQueryClient();
  const [selectedResult, setSelectedResult] = useState<BacktestResult | null>(null);
  const [showConfig, setShowConfig] = useState(true);
  const isMobile = useIsMobile();

  // Auto-hide config panel on mobile
  useEffect(() => {
    if (isMobile) {
      setShowConfig(false);
    }
  }, [isMobile]);

  const [config, setConfig] = useState<BacktestConfig>({
    symbol: 'ETHUSDT',
    interval: '15m',
    startDate: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    endDate: new Date().toISOString().split('T')[0],
    initialCapital: 100000,
    commission: 0.001,
    slippage: 0.0005,
    riskPerTrade: 0.02,
    strategies: [],
  });

  const { data: strategies = [] } = useQuery<Strategy[]>({
    queryKey: ['strategies'],
    queryFn: async () => {
      const res = await api.getStrategies();
      return res.data;
    },
  });

  const { data: results = [] } = useQuery<BacktestResult[]>({
    queryKey: ['backtestResults'],
    queryFn: async () => {
      const res = await api.getBacktestResults();
      return res.data;
    },
  });

  // Generate chart data for selected result
  const monthlyReturns = useMemo(() => generateMonthlyReturns(), [selectedResult]);
  const tradeDistribution = useMemo(() => generateTradeDistribution(), [selectedResult]);

  const runMutation = useMutation({
    mutationFn: (cfg: BacktestConfig) => api.runBacktest(cfg),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backtestResults'] });
    },
  });

  const handleStrategyToggle = (strategyId: string) => {
    setConfig((prev) => ({
      ...prev,
      strategies: prev.strategies.includes(strategyId)
        ? prev.strategies.filter((s) => s !== strategyId)
        : [...prev.strategies, strategyId],
    }));
  };

  const handleRunBacktest = () => {
    if (config.strategies.length === 0) {
      alert('Please select at least one strategy');
      return;
    }
    runMutation.mutate(config);
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
    }).format(price);

  const formatPercent = (value: number) => `${(value * 100).toFixed(2)}%`;

  // Calculate additional metrics
  const getAdditionalMetrics = (result: BacktestResult) => {
    const avgWin = result.metrics.avgWin || (result.metrics.totalReturn > 0 ? result.metrics.totalReturn / result.metrics.winningTrades : 0);
    const avgLoss = result.metrics.avgLoss || (result.metrics.totalReturn < 0 ? Math.abs(result.metrics.totalReturn) / result.metrics.losingTrades : 0);
    const expectancy = (result.metrics.winRate * avgWin) - ((1 - result.metrics.winRate) * avgLoss);
    const recoveryFactor = result.metrics.maxDrawdown > 0 ? result.metrics.totalReturn / result.metrics.maxDrawdown : 0;

    return {
      avgWin,
      avgLoss,
      expectancy,
      recoveryFactor,
      riskRewardRatio: avgLoss > 0 ? avgWin / avgLoss : 0,
    };
  };

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h1 className="page-title">Backtesting</h1>
            <p className="page-description">Test your strategies against historical data</p>
          </div>
          <div style={{ display: 'flex', gap: '12px' }}>
            <button
              onClick={() => setShowConfig(!showConfig)}
              className="btn btn-secondary"
              style={{ display: 'flex', alignItems: 'center', gap: '8px' }}
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707" />
                <circle cx="12" cy="12" r="4" />
              </svg>
              {showConfig ? 'Hide Config' : 'Show Config'}
            </button>
          </div>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid-stats section">
        <div className="stat-card">
          <div className="stat-label">Total Backtests</div>
          <div className="stat-value">{results.length}</div>
          <div className="stat-change text-muted">
            {results.filter(r => r.status === 'completed').length} completed
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Best Return</div>
          <div className="stat-value profit">
            {results.length > 0 && results.some(r => r.status === 'completed')
              ? formatPercent(Math.max(...results.filter(r => r.status === 'completed').map(r => r.metrics.totalReturn)))
              : '--'}
          </div>
          <div className="stat-change text-muted">Best performing</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Avg Win Rate</div>
          <div className="stat-value">
            {results.length > 0 && results.some(r => r.status === 'completed')
              ? formatPercent(results.filter(r => r.status === 'completed').reduce((sum, r) => sum + r.metrics.winRate, 0) / results.filter(r => r.status === 'completed').length)
              : '--'}
          </div>
          <div className="stat-change text-muted">Average across all</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Best Sharpe</div>
          <div className="stat-value">
            {results.length > 0 && results.some(r => r.status === 'completed')
              ? Math.max(...results.filter(r => r.status === 'completed').map(r => r.metrics.sharpeRatio)).toFixed(2)
              : '--'}
          </div>
          <div className="stat-change text-muted">Highest ratio</div>
        </div>
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: showConfig ? (isMobile ? '1fr' : '350px 1fr') : '1fr',
        gap: '16px',
      }}>
        {/* Configuration Panel */}
        {showConfig && (
          <div className="card" style={{ overflow: 'hidden', height: 'fit-content' }}>
            <div style={{
              padding: '16px 20px',
              borderBottom: '1px solid var(--border-color)',
            }}>
              <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Configuration</h3>
            </div>
            <div style={{ padding: '20px' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '12px' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                      Symbol
                    </label>
                    <select
                      value={config.symbol}
                      onChange={(e) => setConfig((prev) => ({ ...prev, symbol: e.target.value }))}
                      className="input select"
                      style={{ width: '100%' }}
                    >
                      <option value="ETHUSDT">ETHUSDT</option>
                      <option value="BTCUSDT">BTCUSDT</option>
                    </select>
                  </div>

                  <div>
                    <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                      Timeframe
                    </label>
                    <select
                      value={config.interval}
                      onChange={(e) => setConfig((prev) => ({ ...prev, interval: e.target.value }))}
                      className="input select"
                      style={{ width: '100%' }}
                    >
                      <option value="1m">1m</option>
                      <option value="5m">5m</option>
                      <option value="15m">15m</option>
                      <option value="1h">1h</option>
                      <option value="4h">4h</option>
                      <option value="1d">1d</option>
                    </select>
                  </div>
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '12px' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                      Start
                    </label>
                    <input
                      type="date"
                      value={config.startDate}
                      onChange={(e) => setConfig((prev) => ({ ...prev, startDate: e.target.value }))}
                      className="input"
                      style={{ width: '100%' }}
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                      End
                    </label>
                    <input
                      type="date"
                      value={config.endDate}
                      onChange={(e) => setConfig((prev) => ({ ...prev, endDate: e.target.value }))}
                      className="input"
                      style={{ width: '100%' }}
                    />
                  </div>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Initial Capital ($)
                  </label>
                  <input
                    type="number"
                    value={config.initialCapital}
                    onChange={(e) => setConfig((prev) => ({ ...prev, initialCapital: parseFloat(e.target.value) }))}
                    className="input"
                    style={{ width: '100%' }}
                    step="1000"
                    min="1000"
                  />
                </div>

                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
                  <div>
                    <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-tertiary)', marginBottom: '6px' }}>
                      Comm (%)
                    </label>
                    <input
                      type="number"
                      value={config.commission * 100}
                      onChange={(e) => setConfig((prev) => ({ ...prev, commission: parseFloat(e.target.value) / 100 }))}
                      className="input"
                      style={{ width: '100%' }}
                      step="0.01"
                      min="0"
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-tertiary)', marginBottom: '6px' }}>
                      Slip (%)
                    </label>
                    <input
                      type="number"
                      value={config.slippage * 100}
                      onChange={(e) => setConfig((prev) => ({ ...prev, slippage: parseFloat(e.target.value) / 100 }))}
                      className="input"
                      style={{ width: '100%' }}
                      step="0.01"
                      min="0"
                    />
                  </div>
                  <div>
                    <label style={{ display: 'block', fontSize: '12px', color: 'var(--text-tertiary)', marginBottom: '6px' }}>
                      Risk (%)
                    </label>
                    <input
                      type="number"
                      value={config.riskPerTrade * 100}
                      onChange={(e) => setConfig((prev) => ({ ...prev, riskPerTrade: parseFloat(e.target.value) / 100 }))}
                      className="input"
                      style={{ width: '100%' }}
                      step="0.5"
                      min="0.5"
                      max="10"
                    />
                  </div>
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Strategies ({config.strategies.length} selected)
                  </label>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    {strategies.map((strategy) => (
                      <label
                        key={strategy.id}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '10px',
                          padding: '8px 12px',
                          background: config.strategies.includes(strategy.id)
                            ? 'rgba(35, 131, 226, 0.1)'
                            : 'var(--bg-secondary)',
                          borderRadius: '6px',
                          cursor: 'pointer',
                          border: config.strategies.includes(strategy.id)
                            ? '1px solid var(--accent-blue)'
                            : '1px solid transparent',
                          transition: 'all 0.15s ease',
                          fontSize: '13px',
                        }}
                      >
                        <input
                          type="checkbox"
                          checked={config.strategies.includes(strategy.id)}
                          onChange={() => handleStrategyToggle(strategy.id)}
                          style={{ width: '14px', height: '14px', accentColor: 'var(--accent-blue)' }}
                        />
                        <span>{strategy.name}</span>
                      </label>
                    ))}
                  </div>
                </div>

                <button
                  onClick={handleRunBacktest}
                  disabled={runMutation.isPending || config.strategies.length === 0}
                  className="btn btn-primary"
                  style={{ width: '100%', padding: '12px' }}
                >
                  {runMutation.isPending ? (
                    <span style={{ display: 'flex', alignItems: 'center', gap: '8px', justifyContent: 'center' }}>
                      <svg className="animate-pulse" width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="12" r="10" opacity="0.3" />
                        <path d="M12 2a10 10 0 0 1 10 10h-2a8 8 0 0 0-8-8V2z" />
                      </svg>
                      Running...
                    </span>
                  ) : 'Run Backtest'}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Main Content Area */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {/* Results List */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: '16px 20px',
              borderBottom: '1px solid var(--border-color)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}>
              <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Backtest Results</h3>
              <span className="badge">{results.length}</span>
            </div>

            {results.length === 0 ? (
              <div className="empty-state" style={{ padding: '40px 20px' }}>
                <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <div className="empty-state-title">No backtest results yet</div>
                <div className="empty-state-text">Configure and run a backtest to see results</div>
              </div>
            ) : (
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '12px', padding: '16px' }}>
                {results.map((result) => (
                  <button
                    key={result.id}
                    onClick={() => setSelectedResult(selectedResult?.id === result.id ? null : result)}
                    style={{
                      padding: '16px',
                      textAlign: 'left',
                      background: selectedResult?.id === result.id
                        ? 'rgba(35, 131, 226, 0.1)'
                        : 'var(--bg-secondary)',
                      border: selectedResult?.id === result.id
                        ? '1px solid var(--accent-blue)'
                        : '1px solid var(--border-color)',
                      borderRadius: '8px',
                      cursor: 'pointer',
                      transition: 'all 0.15s ease',
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                      <span style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '14px' }}>
                        {result.config.strategies.join(', ').substring(0, 20)}{result.config.strategies.join(', ').length > 20 ? '...' : ''}
                      </span>
                      <span className={`badge ${
                        result.status === 'completed' ? 'badge-success' :
                        result.status === 'failed' ? 'badge-danger' : 'badge-warning'
                      }`}>
                        {result.status}
                      </span>
                    </div>
                    <div style={{ fontSize: '12px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                      {result.config.symbol} | {result.config.interval} | {result.config.startDate}
                    </div>
                    {result.status === 'completed' && (
                      <div style={{ display: 'flex', gap: '16px', fontSize: '13px' }}>
                        <div>
                          <span style={{ color: 'var(--text-tertiary)' }}>Return: </span>
                          <span className={result.metrics.totalReturn >= 0 ? 'profit' : 'loss'} style={{ fontWeight: 600 }}>
                            {formatPercent(result.metrics.totalReturn)}
                          </span>
                        </div>
                        <div>
                          <span style={{ color: 'var(--text-tertiary)' }}>Win: </span>
                          <span style={{ fontWeight: 500 }}>{formatPercent(result.metrics.winRate)}</span>
                        </div>
                        <div>
                          <span style={{ color: 'var(--text-tertiary)' }}>Trades: </span>
                          <span style={{ fontWeight: 500 }}>{result.metrics.totalTrades}</span>
                        </div>
                      </div>
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Selected Result Details */}
          {selectedResult && selectedResult.status === 'completed' && (
            <>
              {/* Metrics Grid */}
              <div className="card" style={{ overflow: 'hidden' }}>
                <div style={{
                  padding: '16px 20px',
                  borderBottom: '1px solid var(--border-color)',
                }}>
                  <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Performance Metrics</h3>
                </div>
                <div style={{ padding: isMobile ? '12px' : '20px' }}>
                  <div style={{ display: 'grid', gridTemplateColumns: isMobile ? 'repeat(2, 1fr)' : 'repeat(6, 1fr)', gap: isMobile ? '8px' : '12px' }}>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Total Return</div>
                      <div style={{ fontSize: '22px', fontWeight: 700 }} className={selectedResult.metrics.totalReturn >= 0 ? 'profit' : 'loss'}>
                        {selectedResult.metrics.totalReturn >= 0 ? '+' : ''}{formatPercent(selectedResult.metrics.totalReturn)}
                      </div>
                    </div>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Sharpe Ratio</div>
                      <div style={{ fontSize: '22px', fontWeight: 700, color: selectedResult.metrics.sharpeRatio >= 1 ? 'var(--accent-green)' : 'var(--text-primary)' }}>
                        {selectedResult.metrics.sharpeRatio.toFixed(2)}
                      </div>
                    </div>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Max Drawdown</div>
                      <div style={{ fontSize: '22px', fontWeight: 700, color: 'var(--accent-red)' }}>
                        {formatPercent(selectedResult.metrics.maxDrawdown)}
                      </div>
                    </div>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Win Rate</div>
                      <div style={{ fontSize: '22px', fontWeight: 700 }}>
                        {formatPercent(selectedResult.metrics.winRate)}
                      </div>
                    </div>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Profit Factor</div>
                      <div style={{ fontSize: '22px', fontWeight: 700, color: selectedResult.metrics.profitFactor >= 1.5 ? 'var(--accent-green)' : selectedResult.metrics.profitFactor >= 1 ? 'var(--text-primary)' : 'var(--accent-red)' }}>
                        {selectedResult.metrics.profitFactor.toFixed(2)}
                      </div>
                    </div>
                    <div style={{ padding: '16px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                      <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', marginBottom: '6px' }}>Total Trades</div>
                      <div style={{ fontSize: '22px', fontWeight: 700 }}>
                        {selectedResult.metrics.totalTrades}
                      </div>
                    </div>
                  </div>

                  {/* Additional Metrics Row */}
                  <div style={{ display: 'grid', gridTemplateColumns: isMobile ? 'repeat(2, 1fr)' : 'repeat(5, 1fr)', gap: isMobile ? '8px' : '12px', marginTop: '12px' }}>
                    {(() => {
                      const extra = getAdditionalMetrics(selectedResult);
                      return (
                        <>
                          <div style={{ padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '4px' }}>Winning</div>
                            <div style={{ fontSize: '16px', fontWeight: 600, color: 'var(--accent-green)' }}>
                              {selectedResult.metrics.winningTrades}
                            </div>
                          </div>
                          <div style={{ padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '4px' }}>Losing</div>
                            <div style={{ fontSize: '16px', fontWeight: 600, color: 'var(--accent-red)' }}>
                              {selectedResult.metrics.losingTrades}
                            </div>
                          </div>
                          <div style={{ padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '4px' }}>Avg Win</div>
                            <div style={{ fontSize: '16px', fontWeight: 600, color: 'var(--accent-green)' }}>
                              {formatPercent(extra.avgWin)}
                            </div>
                          </div>
                          <div style={{ padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '4px' }}>Avg Loss</div>
                            <div style={{ fontSize: '16px', fontWeight: 600, color: 'var(--accent-red)' }}>
                              {formatPercent(extra.avgLoss)}
                            </div>
                          </div>
                          <div style={{ padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
                            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', marginBottom: '4px' }}>R:R Ratio</div>
                            <div style={{ fontSize: '16px', fontWeight: 600 }}>
                              {extra.riskRewardRatio.toFixed(2)}
                            </div>
                          </div>
                        </>
                      );
                    })()}
                  </div>
                </div>
              </div>

              {/* Charts Row */}
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '2fr 1fr', gap: '16px' }}>
                {/* Equity Curve */}
                <div className="card" style={{ overflow: 'hidden' }}>
                  <div style={{
                    padding: '16px 20px',
                    borderBottom: '1px solid var(--border-color)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}>
                    <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Equity Curve</h3>
                    <span style={{ fontSize: '12px', color: 'var(--text-tertiary)' }}>
                      Final: {formatPrice(selectedResult.equityCurve[selectedResult.equityCurve.length - 1]?.equity || config.initialCapital)}
                    </span>
                  </div>
                  <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '200px' : '250px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={selectedResult.equityCurve}>
                        <defs>
                          <linearGradient id="equityGradient" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#26a69a" stopOpacity={0.3}/>
                            <stop offset="95%" stopColor="#26a69a" stopOpacity={0}/>
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                        <XAxis dataKey="time" hide />
                        <YAxis
                          domain={['auto', 'auto']}
                          tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`}
                          axisLine={false}
                          tickLine={false}
                          tick={{ fontSize: 11, fill: '#787b86' }}
                        />
                        <Tooltip
                          formatter={(value) => [formatPrice(value as number), 'Equity']}
                          contentStyle={{
                            background: '#1e222d',
                            border: '1px solid #2a2e39',
                            borderRadius: '8px',
                            fontSize: '12px',
                          }}
                        />
                        <ReferenceLine
                          y={config.initialCapital}
                          stroke="#787b86"
                          strokeDasharray="5 5"
                        />
                        <Area
                          type="monotone"
                          dataKey="equity"
                          stroke="#26a69a"
                          fill="url(#equityGradient)"
                          strokeWidth={2}
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </div>

                {/* Monthly Returns */}
                <div className="card" style={{ overflow: 'hidden' }}>
                  <div style={{
                    padding: '16px 20px',
                    borderBottom: '1px solid var(--border-color)',
                  }}>
                    <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Monthly Returns</h3>
                  </div>
                  <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '200px' : '250px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={monthlyReturns}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                        <XAxis
                          dataKey="month"
                          axisLine={false}
                          tickLine={false}
                          tick={{ fontSize: 10, fill: '#787b86' }}
                        />
                        <YAxis
                          axisLine={false}
                          tickLine={false}
                          tick={{ fontSize: 11, fill: '#787b86' }}
                          tickFormatter={(v) => `${v}%`}
                        />
                        <Tooltip
                          formatter={(value: number | undefined) => [value !== undefined ? `${value.toFixed(2)}%` : '', 'Return']}
                          contentStyle={{
                            background: '#1e222d',
                            border: '1px solid #2a2e39',
                            borderRadius: '8px',
                            fontSize: '12px',
                          }}
                        />
                        <ReferenceLine y={0} stroke="#787b86" />
                        <Bar dataKey="return" radius={[4, 4, 0, 0]}>
                          {monthlyReturns.map((entry, index) => (
                            <Cell key={index} fill={entry.return >= 0 ? '#26a69a' : '#ef5350'} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              </div>

              {/* Second Charts Row */}
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: '16px' }}>
                {/* Drawdown Chart */}
                <div className="card" style={{ overflow: 'hidden' }}>
                  <div style={{
                    padding: '16px 20px',
                    borderBottom: '1px solid var(--border-color)',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}>
                    <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Drawdown</h3>
                    <span style={{ fontSize: '12px', color: 'var(--accent-red)' }}>
                      Max: {formatPercent(selectedResult.metrics.maxDrawdown)}
                    </span>
                  </div>
                  <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '180px' : '200px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={selectedResult.drawdownCurve}>
                        <defs>
                          <linearGradient id="ddGradient" x1="0" y1="0" x2="0" y2="1">
                            <stop offset="5%" stopColor="#ef5350" stopOpacity={0.4}/>
                            <stop offset="95%" stopColor="#ef5350" stopOpacity={0}/>
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                        <XAxis dataKey="time" hide />
                        <YAxis
                          domain={['auto', 0]}
                          tickFormatter={(v) => `${(v * 100).toFixed(0)}%`}
                          axisLine={false}
                          tickLine={false}
                          tick={{ fontSize: 11, fill: '#787b86' }}
                        />
                        <Tooltip
                          formatter={(value) => [formatPercent(value as number), 'Drawdown']}
                          contentStyle={{
                            background: '#1e222d',
                            border: '1px solid #2a2e39',
                            borderRadius: '8px',
                            fontSize: '12px',
                          }}
                        />
                        <Area
                          type="monotone"
                          dataKey="drawdown"
                          stroke="#ef5350"
                          fill="url(#ddGradient)"
                          strokeWidth={2}
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                </div>

                {/* Trade Distribution */}
                <div className="card" style={{ overflow: 'hidden' }}>
                  <div style={{
                    padding: '16px 20px',
                    borderBottom: '1px solid var(--border-color)',
                  }}>
                    <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Trade Distribution</h3>
                  </div>
                  <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '180px' : '200px' }}>
                    <ResponsiveContainer width="100%" height="100%">
                      <BarChart data={tradeDistribution} layout="vertical">
                        <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" horizontal={false} />
                        <XAxis type="number" axisLine={false} tickLine={false} tick={{ fontSize: 11, fill: '#787b86' }} />
                        <YAxis
                          dataKey="range"
                          type="category"
                          axisLine={false}
                          tickLine={false}
                          tick={{ fontSize: 10, fill: '#787b86' }}
                          width={70}
                        />
                        <Tooltip
                          formatter={(value: number | undefined) => [value !== undefined ? value : '', 'Trades']}
                          contentStyle={{
                            background: '#1e222d',
                            border: '1px solid #2a2e39',
                            borderRadius: '8px',
                            fontSize: '12px',
                          }}
                        />
                        <Bar dataKey="count" radius={[0, 4, 4, 0]}>
                          {tradeDistribution.map((_, index) => (
                            <Cell key={index} fill={index < 4 ? '#ef5350' : '#26a69a'} />
                          ))}
                        </Bar>
                      </BarChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* Running/Failed State */}
          {selectedResult && selectedResult.status !== 'completed' && (
            <div className="card" style={{ overflow: 'hidden' }}>
              <div className="empty-state" style={{ padding: '60px 20px' }}>
                {selectedResult.status === 'running' ? (
                  <>
                    <div className="animate-pulse" style={{ marginBottom: '16px' }}>
                      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="var(--accent-yellow)" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 6v6l4 2" />
                      </svg>
                    </div>
                    <div className="empty-state-title">Backtest in progress</div>
                    <div className="empty-state-text">Progress: {(selectedResult.progress * 100).toFixed(0)}%</div>
                    <div style={{
                      marginTop: '16px',
                      width: '200px',
                      height: '6px',
                      background: 'var(--bg-active)',
                      borderRadius: '3px',
                      overflow: 'hidden',
                    }}>
                      <div style={{
                        width: `${selectedResult.progress * 100}%`,
                        height: '100%',
                        background: 'var(--accent-yellow)',
                        transition: 'width 0.3s ease',
                      }} />
                    </div>
                  </>
                ) : (
                  <>
                    <svg className="empty-state-icon" viewBox="0 0 24 24" fill="none" stroke="var(--accent-red)" strokeWidth="2">
                      <circle cx="12" cy="12" r="10" />
                      <line x1="15" y1="9" x2="9" y2="15" />
                      <line x1="9" y1="9" x2="15" y2="15" />
                    </svg>
                    <div className="empty-state-title" style={{ color: 'var(--accent-red)' }}>Backtest failed</div>
                    {selectedResult.error && (
                      <div className="empty-state-text">{selectedResult.error}</div>
                    )}
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
