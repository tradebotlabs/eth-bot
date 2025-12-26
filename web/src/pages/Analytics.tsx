import { useState, useMemo, useEffect } from 'react';
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

// Generate mock performance data
const generateEquityCurve = (days: number = 90) => {
  const data = [];
  let equity = 100000;
  let benchmark = 100000;

  for (let i = 0; i < days; i++) {
    const date = new Date();
    date.setDate(date.getDate() - (days - i));

    equity += (Math.random() - 0.45) * 1500;
    benchmark += (Math.random() - 0.48) * 1000;

    data.push({
      date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      equity: Math.round(equity),
      benchmark: Math.round(benchmark),
    });
  }
  return data;
};

const generateDailyReturns = (days: number = 30) => {
  const data = [];
  for (let i = 0; i < days; i++) {
    const date = new Date();
    date.setDate(date.getDate() - (days - i));
    data.push({
      date: date.toLocaleDateString('en-US', { weekday: 'short', day: 'numeric' }),
      return: (Math.random() - 0.45) * 4,
      trades: Math.floor(Math.random() * 12) + 1,
    });
  }
  return data;
};

const generateHourlyDistribution = () => {
  const hours = [];
  for (let i = 0; i < 24; i++) {
    hours.push({
      hour: `${i.toString().padStart(2, '0')}:00`,
      trades: Math.floor(Math.random() * 20),
      pnl: (Math.random() - 0.4) * 500,
    });
  }
  return hours;
};

const generateStrategyPerformance = () => [
  { strategy: 'Trend Following', winRate: 0.62, pnl: 4500, trades: 45, sharpe: 1.8 },
  { strategy: 'Mean Reversion', winRate: 0.55, pnl: 2200, trades: 68, sharpe: 1.2 },
  { strategy: 'Breakout', winRate: 0.48, pnl: -800, trades: 32, sharpe: 0.6 },
  { strategy: 'Scalping', winRate: 0.71, pnl: 1800, trades: 156, sharpe: 2.1 },
  { strategy: 'Momentum', winRate: 0.58, pnl: 3200, trades: 28, sharpe: 1.5 },
];

const generateWinLossStreak = () => {
  const data = [];
  for (let i = 0; i < 50; i++) {
    data.push({
      trade: i + 1,
      result: Math.random() > 0.4 ? 1 : -1,
      pnl: (Math.random() - 0.4) * 200,
    });
  }
  return data;
};

const generateRadarData = () => [
  { metric: 'Win Rate', value: 62, fullMark: 100 },
  { metric: 'Profit Factor', value: 75, fullMark: 100 },
  { metric: 'Sharpe Ratio', value: 68, fullMark: 100 },
  { metric: 'Risk/Reward', value: 82, fullMark: 100 },
  { metric: 'Consistency', value: 58, fullMark: 100 },
  { metric: 'Recovery', value: 71, fullMark: 100 },
];

export function Analytics() {
  const [timeRange, setTimeRange] = useState<'1W' | '1M' | '3M' | '6M' | 'YTD' | 'ALL'>('3M');
  const isMobile = useIsMobile();

  // Generate chart data
  const equityCurve = useMemo(() => generateEquityCurve(90), []);
  const dailyReturns = useMemo(() => generateDailyReturns(30), []);
  const hourlyDistribution = useMemo(() => generateHourlyDistribution(), []);
  const strategyPerformance = useMemo(() => generateStrategyPerformance(), []);
  const winLossStreak = useMemo(() => generateWinLossStreak(), []);
  const radarData = useMemo(() => generateRadarData(), []);

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 }).format(price);

  // Calculate summary metrics
  const totalPnL = strategyPerformance.reduce((sum, s) => sum + s.pnl, 0);
  const totalTrades = strategyPerformance.reduce((sum, s) => sum + s.trades, 0);
  const avgWinRate = strategyPerformance.reduce((sum, s) => sum + s.winRate, 0) / strategyPerformance.length;
  const avgSharpe = strategyPerformance.reduce((sum, s) => sum + s.sharpe, 0) / strategyPerformance.length;

  const winningDays = dailyReturns.filter(d => d.return > 0).length;
  const losingDays = dailyReturns.filter(d => d.return < 0).length;
  const maxDailyGain = Math.max(...dailyReturns.map(d => d.return));
  const maxDailyLoss = Math.min(...dailyReturns.map(d => d.return));

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
            {strategyPerformance.filter(s => s.pnl > 0).length}/{strategyPerformance.length} profitable strategies
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
          <div className="stat-value">{(avgWinRate * 100).toFixed(1)}%</div>
          <div className="stat-change text-muted">
            {winningDays}/{dailyReturns.length} winning days
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
                  <stop offset="5%" stopColor="#2962ff" stopOpacity={0.3}/>
                  <stop offset="95%" stopColor="#2962ff" stopOpacity={0}/>
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
              {strategyPerformance.map((strategy, index) => (
                <tr key={index} style={{ borderBottom: '1px solid var(--border-color)' }}>
                  <td style={{ padding: '14px 20px', fontWeight: 500 }}>{strategy.strategy}</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>{(strategy.winRate * 100).toFixed(1)}%</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>
                    <span className={strategy.pnl >= 0 ? 'profit' : 'loss'} style={{ fontWeight: 600 }}>
                      {strategy.pnl >= 0 ? '+' : ''}{formatPrice(strategy.pnl)}
                    </span>
                  </td>
                  <td style={{ padding: '14px 20px', textAlign: 'right', color: 'var(--text-secondary)' }}>{strategy.trades}</td>
                  <td style={{ padding: '14px 20px', textAlign: 'right' }}>
                    <span style={{ color: strategy.sharpe >= 1.5 ? 'var(--accent-green)' : strategy.sharpe >= 1 ? 'var(--text-primary)' : 'var(--accent-red)' }}>
                      {strategy.sharpe.toFixed(2)}
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
                          width: `${strategy.winRate * 100}%`,
                          height: '100%',
                          background: strategy.pnl >= 0 ? 'var(--accent-green)' : 'var(--accent-red)',
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
