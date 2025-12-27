import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import * as api from '../services/api';
import {
  Line,
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
  ComposedChart,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ReferenceLine,
} from 'recharts';

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

export function Analytics() {
  const [timeRange, setTimeRange] = useState<'1W' | '1M' | '3M' | '6M' | 'YTD' | 'ALL'>('3M');
  const isMobile = useIsMobile();

  // Fetch real data
  const { data: performanceData } = useQuery({
    queryKey: ['performance'],
    queryFn: async () => {
      const res = await api.getPerformance();
      return res.data;
    },
  });

  const { data: equityCurveData = [] } = useQuery({
    queryKey: ['equityCurve', timeRange],
    queryFn: async () => {
      const res = await api.getEquityCurve();
      return res.data;
    },
  });

  const { data: strategiesData = [] } = useQuery({
    queryKey: ['strategies'],
    queryFn: async () => {
      const res = await api.getStrategies();
      return res.data;
    },
  });

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 }).format(price);

  // Use real data or empty fallbacks
  const equityCurve = equityCurveData;
  const strategies = strategiesData;

  const totalPnL = performanceData?.totalReturn || 0;
  const totalTrades = performanceData?.totalTrades || 0;
  const winRate = performanceData?.winRate || 0;
  const maxDrawdown = performanceData?.maxDrawdown || 0;

  // Fallback for empty data visuals
  const radarData = [
    { metric: 'Win Rate', value: winRate * 100, fullMark: 100 },
    { metric: 'Max Drawdown', value: (1 - maxDrawdown) * 100, fullMark: 100 },
    { metric: 'Profit Factor', value: (performanceData?.profitFactor || 0) * 20, fullMark: 100 },
    { metric: 'Total Trades', value: Math.min(totalTrades, 100), fullMark: 100 },
  ];

  const dailyReturns: { date: string; return: number; trades: number }[] = [];
  const hourlyDistribution: { hour: string; trades: number; pnl: number }[] = [];
  const winLossStreak: { trade: number; pnl: number }[] = [];
  const winningDays = 0;
  const losingDays = 0;
  const maxDailyGain = 0;
  const maxDailyLoss = 0;
  const avgSharpe = performanceData?.sharpeRatio || 0;

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div style={{
          display: 'flex',
          flexDirection: isMobile ? 'column' : 'row',
          alignItems: isMobile ? 'stretch' : 'center',
          justifyContent: 'space-between',
          gap: isMobile ? '12px' : '0',
        }}>
          <div>
            <h1 className="page-title">Analytics</h1>
            {!isMobile && <p className="page-description">Comprehensive performance metrics and trading analysis</p>}
          </div>
          <div style={{
            display: 'flex',
            background: 'var(--bg-secondary)',
            borderRadius: '8px',
            padding: '4px',
            overflowX: isMobile ? 'auto' : 'visible',
            WebkitOverflowScrolling: 'touch',
          }}>
            {(['1W', '1M', '3M', '6M', 'YTD', 'ALL'] as const).map(range => (
              <button
                key={range}
                onClick={() => setTimeRange(range)}
                style={{
                  padding: isMobile ? '6px 10px' : '6px 12px',
                  fontSize: isMobile ? '11px' : '12px',
                  fontWeight: 500,
                  border: 'none',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  background: timeRange === range ? 'var(--bg-tertiary)' : 'transparent',
                  color: timeRange === range ? 'var(--text-primary)' : 'var(--text-secondary)',
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                {range}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid-stats section">
        <div className="stat-card">
          <div className="stat-label">Total P&L</div>
          <div className={`stat-value ${totalPnL >= 0 ? 'profit' : 'loss'}`}>
            {totalPnL >= 0 ? '+' : ''}{formatPrice(totalPnL)}
          </div>
          <div className="stat-change text-muted">
            {strategies.filter(s => s.performance.totalPnl > 0).length}/{strategies.length} profitable strategies
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total Trades</div>
          <div className="stat-value">{totalTrades}</div>
          <div className="stat-change text-muted">
            {(totalTrades / 30).toFixed(1)} trades/day avg
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Avg Win Rate</div>
          <div className="stat-value">{(winRate * 100).toFixed(1)}%</div>
          <div className="stat-change text-muted">
            {winningDays}/{dailyReturns.length || performanceData?.totalTrades || 0} winning days
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Avg Sharpe</div>
          <div className={`stat-value ${avgSharpe >= 1 ? 'profit' : ''}`}>
            {avgSharpe.toFixed(2)}
          </div>
          <div className="stat-change text-muted">
            Risk-adjusted return
          </div>
        </div>
      </div>

      {/* Equity Curve - Full Width */}
      <div className="card section" style={{ overflow: 'hidden' }}>
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid var(--border-color)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Equity Curve</h3>
          <div style={{ display: 'flex', gap: '16px', fontSize: '12px' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ width: '12px', height: '3px', background: '#2962ff', borderRadius: '2px' }} />
              Portfolio
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <span style={{ width: '12px', height: '3px', background: '#787b86', borderRadius: '2px' }} />
              Benchmark
            </span>
          </div>
        </div>
        <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '220px' : '300px' }}>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={equityCurve}>
              <defs>
                <linearGradient id="portfolioGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#2962ff" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#2962ff" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
              <XAxis
                dataKey="date"
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 11, fill: '#787b86' }}
                interval="preserveStartEnd"
              />
              <YAxis
                axisLine={false}
                tickLine={false}
                tick={{ fontSize: 11, fill: '#787b86' }}
                tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`}
                domain={['dataMin - 5000', 'dataMax + 5000']}
              />
              <Tooltip
                contentStyle={{
                  background: '#1e222d',
                  border: '1px solid #2a2e39',
                  borderRadius: '8px',
                  fontSize: '12px',
                }}
                wrapperStyle={{ outline: 'none' }}
                allowEscapeViewBox={{ x: false, y: false }}
                formatter={(value: number | undefined) => [value !== undefined ? formatPrice(value) : '', '']}
              />
              <ReferenceLine y={100000} stroke="#787b86" strokeDasharray="5 5" />
              <Line
                type="monotone"
                dataKey="benchmark"
                stroke="#787b86"
                strokeWidth={1}
                strokeDasharray="5 5"
                dot={false}
              />
              <Area
                type="monotone"
                dataKey="equity"
                stroke="#2962ff"
                fill="url(#portfolioGradient)"
                strokeWidth={2}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Two Column Row */}
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: '16px' }} className="section">
        {/* Daily Returns */}
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{
            padding: '16px 20px',
            borderBottom: '1px solid var(--border-color)',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}>
            <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Daily Returns</h3>
            <div style={{ display: 'flex', gap: '12px', fontSize: '12px' }}>
              <span style={{ color: 'var(--accent-green)' }}>+{maxDailyGain.toFixed(2)}%</span>
              <span style={{ color: 'var(--accent-red)' }}>{maxDailyLoss.toFixed(2)}%</span>
            </div>
          </div>
          <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '200px' : '250px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <ComposedChart data={dailyReturns}>
                <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 10, fill: '#787b86' }}
                  interval={2}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 11, fill: '#787b86' }}
                  tickFormatter={(v) => `${v}%`}
                />
                <Tooltip
                  contentStyle={{
                    background: '#1e222d',
                    border: '1px solid #2a2e39',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  wrapperStyle={{ outline: 'none' }}
                  allowEscapeViewBox={{ x: false, y: false }}
                  formatter={(value: number | undefined, name: string | undefined) => [
                    value !== undefined ? ((name === 'return') ? `${value.toFixed(2)}%` : value) : '',
                    name === 'return' ? 'Return' : 'Trades'
                  ]}
                />
                <ReferenceLine y={0} stroke="#787b86" />
                <Bar dataKey="return" radius={[4, 4, 0, 0]} barSize={12}>
                  {dailyReturns.map((entry, index) => (
                    <Cell key={index} fill={entry.return >= 0 ? '#26a69a' : '#ef5350'} />
                  ))}
                </Bar>
                <Line
                  type="monotone"
                  dataKey="trades"
                  stroke="#ff9800"
                  strokeWidth={1}
                  dot={false}
                  yAxisId={1}
                />
                <YAxis
                  yAxisId={1}
                  orientation="right"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 11, fill: '#ff9800' }}
                />
              </ComposedChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Performance Radar */}
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{
            padding: '16px 20px',
            borderBottom: '1px solid var(--border-color)',
          }}>
            <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Performance Profile</h3>
          </div>
          <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '200px' : '250px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <RadarChart data={radarData}>
                <PolarGrid stroke="#2a2e39" />
                <PolarAngleAxis
                  dataKey="metric"
                  tick={{ fontSize: 11, fill: '#787b86' }}
                />
                <PolarRadiusAxis
                  angle={30}
                  domain={[0, 100]}
                  tick={{ fontSize: 10, fill: '#787b86' }}
                />
                <Radar
                  name="Performance"
                  dataKey="value"
                  stroke="#2962ff"
                  fill="#2962ff"
                  fillOpacity={0.3}
                />
              </RadarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Strategy Performance Table */}
      <div className="card section" style={{ overflow: 'hidden' }}>
        <div style={{
          padding: '16px 20px',
          borderBottom: '1px solid var(--border-color)',
        }}>
          <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Strategy Performance</h3>
        </div>
        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border-color)' }}>
                <th style={{ padding: '12px 20px', textAlign: 'left', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>Strategy</th>
                <th style={{ padding: '12px 20px', textAlign: 'right', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>Win Rate</th>
                <th style={{ padding: '12px 20px', textAlign: 'right', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>P&L</th>
                <th style={{ padding: '12px 20px', textAlign: 'right', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>Trades</th>
                <th style={{ padding: '12px 20px', textAlign: 'right', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>Sharpe</th>
                <th style={{ padding: '12px 20px', textAlign: 'right', fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', fontWeight: 600 }}>Performance</th>
              </tr>
            </thead>
            <tbody>
              {strategies.map((strategy, index) => (
                <tr key={index} style={{ borderBottom: '1px solid var(--border-color)' }}>
                  <td style={{ padding: '14px 20px', fontWeight: 500 }}>{strategy.name}</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>{(strategy.performance.winRate * 100).toFixed(1)}%</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>
                    <span className={strategy.performance.totalPnl >= 0 ? 'profit' : 'loss'} style={{ fontWeight: 600 }}>
                      {strategy.performance.totalPnl >= 0 ? '+' : ''}{formatPrice(strategy.performance.totalPnl)}
                    </span>
                  </td>
                  <td style={{ padding: '14px 20px', textAlign: 'right', color: 'var(--text-secondary)' }}>{strategy.performance.totalTrades}</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>
                    <span style={{ color: strategy.performance.profitFactor >= 2 ? 'var(--accent-green)' : strategy.performance.profitFactor >= 1 ? 'var(--text-primary)' : 'var(--accent-red)' }}>
                      {strategy.performance.profitFactor.toFixed(2)}
                    </span>
                  </td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>
                    <div style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '8px',
                      width: '100px',
                    }}>
                      <div style={{
                        flex: 1,
                        height: '6px',
                        background: 'var(--bg-tertiary)',
                        borderRadius: '3px',
                        overflow: 'hidden',
                      }}>
                        <div style={{
                          width: `${strategy.performance.winRate * 100}%`,
                          height: '100%',
                          background: strategy.performance.totalPnl >= 0 ? 'var(--accent-green)' : 'var(--accent-red)',
                          borderRadius: '3px',
                        }} />
                      </div>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Three Column Row */}
      <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr 1fr', gap: '16px' }}>
        {/* Hourly Distribution */}
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{
            padding: '16px 20px',
            borderBottom: '1px solid var(--border-color)',
          }}>
            <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Hourly Distribution</h3>
          </div>
          <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '180px' : '220px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={hourlyDistribution.filter((_, i) => i % 2 === 0)}>
                <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                <XAxis
                  dataKey="hour"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 9, fill: '#787b86' }}
                  interval={0}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 11, fill: '#787b86' }}
                />
                <Tooltip
                  contentStyle={{
                    background: '#1e222d',
                    border: '1px solid #2a2e39',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  wrapperStyle={{ outline: 'none' }}
                  allowEscapeViewBox={{ x: false, y: false }}
                />
                <Bar dataKey="trades" fill="#2962ff" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Win/Loss Streak */}
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{
            padding: '16px 20px',
            borderBottom: '1px solid var(--border-color)',
          }}>
            <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Trade Results</h3>
          </div>
          <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '180px' : '220px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={winLossStreak} barGap={0}>
                <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                <XAxis dataKey="trade" hide />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 11, fill: '#787b86' }}
                  tickFormatter={(v) => `$${v}`}
                />
                <Tooltip
                  contentStyle={{
                    background: '#1e222d',
                    border: '1px solid #2a2e39',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                  wrapperStyle={{ outline: 'none' }}
                  allowEscapeViewBox={{ x: false, y: false }}
                  formatter={(value: number | undefined) => [value !== undefined ? formatPrice(value) : '', 'P&L']}
                />
                <ReferenceLine y={0} stroke="#787b86" />
                <Bar dataKey="pnl" radius={[2, 2, 0, 0]}>
                  {winLossStreak.map((entry, index) => (
                    <Cell key={index} fill={entry.pnl >= 0 ? '#26a69a' : '#ef5350'} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Key Metrics */}
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{
            padding: '16px 20px',
            borderBottom: '1px solid var(--border-color)',
          }}>
            <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Key Metrics</h3>
          </div>
          <div style={{ padding: '16px' }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 12px', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Best Day</span>
                <span style={{ fontWeight: 600, color: 'var(--accent-green)' }}>+{maxDailyGain.toFixed(2)}%</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 12px', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Worst Day</span>
                <span style={{ fontWeight: 600, color: 'var(--accent-red)' }}>{maxDailyLoss.toFixed(2)}%</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 12px', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Win Days</span>
                <span style={{ fontWeight: 600 }}>{winningDays} ({((winningDays / dailyReturns.length) * 100).toFixed(0)}%)</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 12px', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Lose Days</span>
                <span style={{ fontWeight: 600 }}>{losingDays} ({((losingDays / dailyReturns.length) * 100).toFixed(0)}%)</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 12px', background: 'var(--bg-secondary)', borderRadius: '6px' }}>
                <span style={{ color: 'var(--text-secondary)', fontSize: '13px' }}>Avg Trades/Day</span>
                <span style={{ fontWeight: 600 }}>{(totalTrades / 30).toFixed(1)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
