import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import {
  AreaChart,
  Area,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ComposedChart,
  Bar,
  ReferenceLine,
} from 'recharts';
import * as api from '../services/api';
import type { RiskConfig } from '../types';

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


interface RiskGaugeProps {
  label: string;
  current: number;
  limit: number;
  unit?: string;
  inverse?: boolean;
}

function RiskGauge({ label, current, limit, unit = '%', inverse = false }: RiskGaugeProps) {
  const percentage = Math.min((Math.abs(current) / Math.abs(limit)) * 100, 100);
  const isWarning = percentage > 60 && percentage <= 80;
  const isDanger = percentage > 80;

  let barColor = 'var(--accent-green)';
  if (isWarning) barColor = '#ff9800';
  if (isDanger) barColor = 'var(--accent-red)';

  return (
    <div style={{
      padding: '16px',
      background: 'var(--bg-secondary)',
      borderRadius: '8px',
    }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginBottom: '8px',
        fontSize: '13px',
      }}>
        <span style={{ color: 'var(--text-secondary)' }}>{label}</span>
        <span style={{ fontWeight: 600 }}>
          {inverse ? '-' : ''}{Math.abs(current).toFixed(unit === '$' ? 0 : 1)}{unit} / {Math.abs(limit).toFixed(unit === '$' ? 0 : 1)}{unit}
        </span>
      </div>
      <div style={{
        height: '8px',
        background: 'var(--bg-tertiary)',
        borderRadius: '4px',
        overflow: 'hidden',
      }}>
        <div style={{
          width: `${percentage}%`,
          height: '100%',
          background: barColor,
          borderRadius: '4px',
          transition: 'width 0.3s ease',
        }} />
      </div>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        marginTop: '4px',
        fontSize: '11px',
        color: 'var(--text-tertiary)',
      }}>
        <span>{percentage.toFixed(0)}% used</span>
        <span style={{ color: isDanger ? 'var(--accent-red)' : isWarning ? '#ff9800' : 'inherit' }}>
          {isDanger ? 'Critical' : isWarning ? 'Warning' : 'Safe'}
        </span>
      </div>
    </div>
  );
}

export function Risk() {
  const queryClient = useQueryClient();
  const [editedConfig, setEditedConfig] = useState<Partial<RiskConfig> | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'settings'>('overview');
  const isMobile = useIsMobile();

  const { data: riskStatus } = useQuery({
    queryKey: ['riskStatus'],
    queryFn: async () => {
      const res = await api.getRiskStatus();
      return res.data;
    },
    refetchInterval: 5000,
  });

  const { data: riskConfig, isLoading } = useQuery<RiskConfig>({
    queryKey: ['riskConfig'],
    queryFn: async () => {
      const res = await api.getRiskConfig();
      return res.data;
    },
  });

  // Fetch chart data
  const { data: drawdownHistory = [] } = useQuery({
    queryKey: ['drawdownHistory'],
    queryFn: async () => {
      const res = await api.getDrawdown();
      return res.data;
    },
  });

  const { data: pnlHistory = [] } = useQuery({
    queryKey: ['pnlHistory'],
    queryFn: async () => {
      const res = await api.getRiskEvents(); // Using risk events or pnl history
      return res.data;
    },
  });

  const updateMutation = useMutation({
    mutationFn: (config: Partial<RiskConfig>) => api.updateRiskConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['riskConfig'] });
      setEditedConfig(null);
    },
  });

  const handleChange = (key: keyof RiskConfig, value: number | boolean) => {
    setEditedConfig((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleSave = () => {
    if (editedConfig) {
      updateMutation.mutate(editedConfig);
    }
  };

  const formatPercent = (value: number) => `${(value * 100).toFixed(2)}%`;
  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price);

  if (isLoading || !riskConfig) {
    return (
      <div className="animate-fade-in">
        <div className="page-header">
          <div className="skeleton" style={{ width: '200px', height: '32px', marginBottom: '8px' }} />
          <div className="skeleton" style={{ width: '300px', height: '20px' }} />
        </div>
        <div className="grid-stats section">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton" style={{ height: '100px', borderRadius: '8px' }} />
          ))}
        </div>
        <div className="skeleton" style={{ height: '300px', borderRadius: '8px' }} />
      </div>
    );
  }

  const currentConfig = { ...riskConfig, ...editedConfig };

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
            <h1 className="page-title">Risk Management</h1>
            {!isMobile && <p className="page-description">Monitor risk metrics and configure trading limits</p>}
          </div>
          <div style={{
            display: 'flex',
            gap: isMobile ? '8px' : '12px',
            alignItems: 'center',
            flexWrap: 'wrap',
          }}>
            {/* Tab Switcher */}
            <div style={{
              display: 'flex',
              background: 'var(--bg-secondary)',
              borderRadius: '8px',
              padding: '4px',
              flex: isMobile ? 1 : 'none',
            }}>
              <button
                onClick={() => setActiveTab('overview')}
                style={{
                  padding: isMobile ? '8px 12px' : '8px 16px',
                  fontSize: isMobile ? '12px' : '13px',
                  fontWeight: 500,
                  border: 'none',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  background: activeTab === 'overview' ? 'var(--bg-tertiary)' : 'transparent',
                  color: activeTab === 'overview' ? 'var(--text-primary)' : 'var(--text-secondary)',
                  flex: isMobile ? 1 : 'none',
                }}
              >
                Overview
              </button>
              <button
                onClick={() => setActiveTab('settings')}
                style={{
                  padding: isMobile ? '8px 12px' : '8px 16px',
                  fontSize: isMobile ? '12px' : '13px',
                  fontWeight: 500,
                  border: 'none',
                  borderRadius: '6px',
                  cursor: 'pointer',
                  background: activeTab === 'settings' ? 'var(--bg-tertiary)' : 'transparent',
                  color: activeTab === 'settings' ? 'var(--text-primary)' : 'var(--text-secondary)',
                  flex: isMobile ? 1 : 'none',
                }}
              >
                Settings
              </button>
            </div>

            {editedConfig && (
              <>
                <button
                  onClick={() => setEditedConfig(null)}
                  className="btn btn-secondary"
                  style={{ fontSize: isMobile ? '12px' : '13px' }}
                >
                  Cancel
                </button>
                <button
                  onClick={handleSave}
                  disabled={updateMutation.isPending}
                  className="btn btn-primary"
                  style={{ fontSize: isMobile ? '12px' : '13px' }}
                >
                  {updateMutation.isPending ? 'Saving...' : 'Save'}
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Risk Status Cards */}
      {riskStatus && (
        <div className="grid-stats section">
          <div className="stat-card">
            <div className="stat-label">Daily P&L</div>
            <div className={`stat-value ${riskStatus.dailyPnl >= 0 ? 'profit' : 'loss'}`}>
              {riskStatus.dailyPnl >= 0 ? '+' : ''}{formatPrice(riskStatus.dailyPnl)}
            </div>
            <div className="stat-change text-muted">
              Limit: {formatPrice(riskConfig.maxDailyLoss * -1)}
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-label">Weekly P&L</div>
            <div className={`stat-value ${riskStatus.weeklyPnl >= 0 ? 'profit' : 'loss'}`}>
              {riskStatus.weeklyPnl >= 0 ? '+' : ''}{formatPrice(riskStatus.weeklyPnl)}
            </div>
            <div className="stat-change text-muted">
              Limit: {formatPrice(riskConfig.maxWeeklyLoss * -1)}
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-label">Current Drawdown</div>
            <div className="stat-value loss">
              {formatPercent(riskStatus.currentDrawdown)}
            </div>
            <div className="stat-change text-muted">
              Max: {formatPercent(riskConfig.maxDrawdown)}
            </div>
          </div>

          <div className="stat-card">
            <div className="stat-label">Circuit Breaker</div>
            <div style={{
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
            }}>
              <div style={{
                width: '12px',
                height: '12px',
                borderRadius: '50%',
                background: riskStatus.circuitBreakerActive ? 'var(--accent-red)' : 'var(--accent-green)',
                boxShadow: `0 0 8px ${riskStatus.circuitBreakerActive ? 'var(--accent-red)' : 'var(--accent-green)'}`,
              }} />
              <span className={`stat-value ${riskStatus.circuitBreakerActive ? 'loss' : 'profit'}`} style={{ fontSize: '20px' }}>
                {riskStatus.circuitBreakerActive ? 'ACTIVE' : 'OK'}
              </span>
            </div>
            <div className="stat-change text-muted">
              Losses: {riskStatus.consecutiveLosses} / {riskConfig.consecutiveLossLimit}
            </div>
          </div>
        </div>
      )}

      {activeTab === 'overview' && (
        <>
          {/* Risk Utilization Gauges */}
          <div className="card section" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: '16px 20px',
              borderBottom: '1px solid var(--border-color)',
            }}>
              <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Risk Utilization</h3>
            </div>
            <div style={{ padding: isMobile ? '12px' : '20px' }}>
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(3, 1fr)', gap: isMobile ? '12px' : '16px' }}>
                <RiskGauge
                  label="Daily Loss Limit"
                  current={riskStatus?.dailyPnl || 0}
                  limit={riskConfig.maxDailyLoss}
                  unit="$"
                  inverse
                />
                <RiskGauge
                  label="Weekly Loss Limit"
                  current={riskStatus?.weeklyPnl || 0}
                  limit={riskConfig.maxWeeklyLoss}
                  unit="$"
                  inverse
                />
                <RiskGauge
                  label="Max Drawdown"
                  current={(riskStatus?.currentDrawdown || 0) * 100}
                  limit={riskConfig.maxDrawdown * 100}
                  unit="%"
                />
              </div>
            </div>
          </div>

          {/* Charts Row */}
          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : '1fr 1fr', gap: '16px' }} className="section">
            {/* Drawdown Chart */}
            <div className="card" style={{ overflow: 'hidden' }}>
              <div style={{
                padding: '16px 20px',
                borderBottom: '1px solid var(--border-color)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
                <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Drawdown History</h3>
                <span style={{ fontSize: '12px', color: 'var(--text-tertiary)' }}>Last 30 days</span>
              </div>
              <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '220px' : '280px' }}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={drawdownHistory}>
                    <defs>
                      <linearGradient id="drawdownGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#ef5350" stopOpacity={0.4} />
                        <stop offset="95%" stopColor="#ef5350" stopOpacity={0} />
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
                      tickFormatter={(value) => `${value}%`}
                      domain={[0, 'auto']}
                    />
                    <Tooltip
                      contentStyle={{
                        background: '#1e222d',
                        border: '1px solid #2a2e39',
                        borderRadius: '8px',
                        fontSize: '12px',
                      }}
                      formatter={(value: number | undefined) => [value !== undefined ? `${value.toFixed(2)}%` : '', 'Drawdown']}
                    />
                    <ReferenceLine
                      y={riskConfig.maxDrawdown * 100}
                      stroke="#ff9800"
                      strokeDasharray="5 5"
                      label={{
                        value: 'Max DD',
                        position: 'right',
                        fill: '#ff9800',
                        fontSize: 11,
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="drawdown"
                      stroke="#ef5350"
                      fill="url(#drawdownGradient)"
                      strokeWidth={2}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Daily P&L Chart */}
            <div className="card" style={{ overflow: 'hidden' }}>
              <div style={{
                padding: '16px 20px',
                borderBottom: '1px solid var(--border-color)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
              }}>
                <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Daily P&L</h3>
                <span style={{ fontSize: '12px', color: 'var(--text-tertiary)' }}>Last 14 days</span>
              </div>
              <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '220px' : '280px' }}>
                <ResponsiveContainer width="100%" height="100%">
                  <ComposedChart data={pnlHistory}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#2a2e39" vertical={false} />
                    <XAxis
                      dataKey="date"
                      axisLine={false}
                      tickLine={false}
                      tick={{ fontSize: 11, fill: '#787b86' }}
                      interval={0}
                      angle={-45}
                      textAnchor="end"
                      height={60}
                    />
                    <YAxis
                      axisLine={false}
                      tickLine={false}
                      tick={{ fontSize: 11, fill: '#787b86' }}
                      tickFormatter={(value) => `$${value}`}
                    />
                    <Tooltip
                      contentStyle={{
                        background: '#1e222d',
                        border: '1px solid #2a2e39',
                        borderRadius: '8px',
                        fontSize: '12px',
                      }}
                      formatter={(value: number | undefined, name: string | undefined) => [
                        value !== undefined ? ((name === 'pnl') ? `$${value.toFixed(0)}` : value) : '',
                        name === 'pnl' ? 'P&L' : 'Trades'
                      ]}
                    />
                    <ReferenceLine y={0} stroke="#787b86" />
                    <ReferenceLine
                      y={-riskConfig.maxDailyLoss}
                      stroke="#ef5350"
                      strokeDasharray="5 5"
                      label={{
                        value: 'Daily Limit',
                        position: 'right',
                        fill: '#ef5350',
                        fontSize: 11,
                      }}
                    />
                    <Bar
                      dataKey="pnl"
                      fill="#26a69a"
                      radius={[4, 4, 0, 0]}
                      barSize={20}
                      // Color based on positive/negative
                      shape={(props: any) => {
                        const { x, y, width, height, payload } = props;
                        const fill = payload.pnl >= 0 ? '#26a69a' : '#ef5350';
                        return <rect x={x} y={y} width={width} height={height} fill={fill} rx={4} />;
                      }}
                    />
                  </ComposedChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Equity Curve */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: '16px 20px',
              borderBottom: '1px solid var(--border-color)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <h3 style={{ fontSize: '15px', fontWeight: 600, margin: 0 }}>Equity vs Peak</h3>
              <div style={{ display: 'flex', gap: '16px', fontSize: '12px' }}>
                <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span style={{ width: '12px', height: '3px', background: '#2962ff', borderRadius: '2px' }} />
                  Equity
                </span>
                <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <span style={{ width: '12px', height: '3px', background: '#787b86', borderRadius: '2px' }} />
                  Peak
                </span>
              </div>
            </div>
            <div style={{ padding: isMobile ? '12px' : '16px', height: isMobile ? '200px' : '250px' }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={drawdownHistory}>
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
                    tickFormatter={(value) => `$${(value / 1000).toFixed(0)}k`}
                    domain={['dataMin - 5000', 'dataMax + 5000']}
                  />
                  <Tooltip
                    contentStyle={{
                      background: '#1e222d',
                      border: '1px solid #2a2e39',
                      borderRadius: '8px',
                      fontSize: '12px',
                    }}
                    formatter={(value: number | undefined) => [value !== undefined ? `$${value.toLocaleString()}` : '', '']}
                  />
                  <Line
                    type="monotone"
                    dataKey="peak"
                    stroke="#787b86"
                    strokeWidth={1}
                    strokeDasharray="5 5"
                    dot={false}
                  />
                  <Line
                    type="monotone"
                    dataKey="equity"
                    stroke="#2962ff"
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>
        </>
      )}

      {activeTab === 'settings' && (
        <>
          {/* Position Limits Card */}
          <div className="card section" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: isMobile ? '12px 16px' : '16px 20px',
              borderBottom: '1px solid var(--border-color)',
            }}>
              <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>Position Limits</h3>
            </div>
            <div style={{ padding: isMobile ? '12px' : '20px' }}>
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)', gap: isMobile ? '12px' : '20px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Position Size (%)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxPositionSize * 100}
                    onChange={(e) =>
                      handleChange('maxPositionSize', parseFloat(e.target.value) / 100)
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="1"
                    min="1"
                    max="100"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Risk Per Trade (%)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxRiskPerTrade * 100}
                    onChange={(e) =>
                      handleChange('maxRiskPerTrade', parseFloat(e.target.value) / 100)
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="0.1"
                    min="0.1"
                    max="10"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Open Positions
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxOpenPositions}
                    onChange={(e) =>
                      handleChange('maxOpenPositions', parseInt(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    min="1"
                    max="20"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Leverage
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxLeverage}
                    onChange={(e) =>
                      handleChange('maxLeverage', parseFloat(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="1"
                    min="1"
                    max="20"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Loss Limits Card */}
          <div className="card section" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: isMobile ? '12px 16px' : '16px 20px',
              borderBottom: '1px solid var(--border-color)',
            }}>
              <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>Loss Limits</h3>
            </div>
            <div style={{ padding: isMobile ? '12px' : '20px' }}>
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)', gap: isMobile ? '12px' : '20px' }}>
                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Daily Loss ($)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxDailyLoss}
                    onChange={(e) =>
                      handleChange('maxDailyLoss', parseFloat(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="100"
                    min="0"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Weekly Loss ($)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxWeeklyLoss}
                    onChange={(e) =>
                      handleChange('maxWeeklyLoss', parseFloat(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="100"
                    min="0"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Max Drawdown (%)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.maxDrawdown * 100}
                    onChange={(e) =>
                      handleChange('maxDrawdown', parseFloat(e.target.value) / 100)
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="1"
                    min="1"
                    max="50"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Min Risk/Reward Ratio
                  </label>
                  <input
                    type="number"
                    value={currentConfig.minRiskRewardRatio}
                    onChange={(e) =>
                      handleChange('minRiskRewardRatio', parseFloat(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    step="0.5"
                    min="1"
                    max="10"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Circuit Breaker Card */}
          <div className="card" style={{ overflow: 'hidden' }}>
            <div style={{
              padding: isMobile ? '12px 16px' : '16px 20px',
              borderBottom: '1px solid var(--border-color)',
            }}>
              <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>Circuit Breaker</h3>
            </div>
            <div style={{ padding: isMobile ? '12px' : '20px' }}>
              <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)', gap: isMobile ? '12px' : '20px' }}>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '16px',
                  background: 'var(--bg-secondary)',
                  borderRadius: '8px',
                }}>
                  <div>
                    <div style={{ fontWeight: 500, marginBottom: '4px' }}>Enable Circuit Breaker</div>
                    <div style={{ fontSize: '13px', color: 'var(--text-tertiary)' }}>
                      Automatically halt trading after consecutive losses
                    </div>
                  </div>
                  <button
                    onClick={() =>
                      handleChange('enableCircuitBreaker', !currentConfig.enableCircuitBreaker)
                    }
                    className={`toggle ${currentConfig.enableCircuitBreaker ? 'active' : ''}`}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Consecutive Loss Limit
                  </label>
                  <input
                    type="number"
                    value={currentConfig.consecutiveLossLimit}
                    onChange={(e) =>
                      handleChange('consecutiveLossLimit', parseInt(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    min="1"
                    max="20"
                  />
                </div>

                <div>
                  <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                    Halt Duration (hours)
                  </label>
                  <input
                    type="number"
                    value={currentConfig.haltDurationHours}
                    onChange={(e) =>
                      handleChange('haltDurationHours', parseInt(e.target.value))
                    }
                    className="input"
                    style={{ width: '100%' }}
                    min="1"
                    max="72"
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Warning Callout */}
          <div className="callout callout-warning" style={{ marginTop: '20px' }}>
            <div className="callout-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
            </div>
            <div className="callout-content">
              <strong>Risk Management Advisory</strong>
              <p style={{ margin: '4px 0 0', color: 'var(--text-secondary)' }}>
                These settings directly impact your trading safety. Ensure you understand the implications before making changes.
                Consider starting with conservative limits and adjusting based on your trading performance.
              </p>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
