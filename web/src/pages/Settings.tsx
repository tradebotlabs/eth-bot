import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as api from '../services/api';
import type { Config } from '../types';

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

export function Settings() {
  const queryClient = useQueryClient();
  const [editedConfig, setEditedConfig] = useState<Partial<Config> | null>(null);
  const isMobile = useIsMobile();

  const { data: config, isLoading } = useQuery<Config>({
    queryKey: ['config'],
    queryFn: async () => {
      const res = await api.getConfig();
      return res.data;
    },
  });

  const updateMutation = useMutation({
    mutationFn: (cfg: Partial<Config>) => api.updateConfig(cfg),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config'] });
      setEditedConfig(null);
    },
  });

  const handleTradingChange = (key: string, value: string | number | string[]) => {
    if (!config) return;
    setEditedConfig((prev) => ({
      ...prev,
      trading: {
        ...config.trading,
        ...(prev?.trading || {}),
        [key]: value,
      },
    }));
  };

  const handleIndicatorChange = (key: string, value: number) => {
    if (!config) return;
    setEditedConfig((prev) => ({
      ...prev,
      indicators: {
        ...config.indicators,
        ...(prev?.indicators || {}),
        [key]: value,
      },
    }));
  };

  const handleSave = () => {
    if (editedConfig) {
      updateMutation.mutate(editedConfig);
    }
  };

  if (isLoading || !config) {
    return (
      <div className="animate-fade-in">
        <div className="page-header">
          <div className="skeleton" style={{ width: '150px', height: '32px', marginBottom: '8px' }} />
          <div className="skeleton" style={{ width: '280px', height: '20px' }} />
        </div>
        <div className="skeleton section" style={{ height: '300px', borderRadius: '8px' }} />
        <div className="skeleton section" style={{ height: '200px', borderRadius: '8px' }} />
      </div>
    );
  }

  const currentConfig: Config = {
    trading: { ...config.trading, ...editedConfig?.trading },
    risk: { ...config.risk, ...editedConfig?.risk },
    indicators: { ...config.indicators, ...editedConfig?.indicators },
    strategies: { ...config.strategies, ...editedConfig?.strategies },
  };

  const timeframes = ['1m', '5m', '15m', '1h', '4h', '1d'];

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
            <h1 className="page-title">Settings</h1>
            {!isMobile && <p className="page-description">Configure trading parameters and indicator settings</p>}
          </div>
          {editedConfig && (
            <div style={{ display: 'flex', gap: isMobile ? '8px' : '12px' }}>
              <button
                onClick={() => setEditedConfig(null)}
                className="btn btn-secondary"
                style={{ fontSize: isMobile ? '12px' : '13px', flex: isMobile ? 1 : 'none' }}
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={updateMutation.isPending}
                className="btn btn-primary"
                style={{ fontSize: isMobile ? '12px' : '13px', flex: isMobile ? 1 : 'none' }}
              >
                {updateMutation.isPending ? 'Saving...' : 'Save'}
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Trading Settings Card */}
      <div className="card section" style={{ overflow: 'hidden' }}>
        <div style={{
          padding: isMobile ? '12px 16px' : '16px 20px',
          borderBottom: '1px solid var(--border-color)',
        }}>
          <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>Trading Settings</h3>
        </div>
        <div style={{ padding: isMobile ? '12px' : '20px' }}>
          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? '1fr' : 'repeat(2, 1fr)', gap: isMobile ? '12px' : '20px' }}>
            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                Symbol
              </label>
              <select
                value={currentConfig.trading.symbol}
                onChange={(e) => handleTradingChange('symbol', e.target.value)}
                className="input select"
                style={{ width: '100%' }}
              >
                <option value="ETHUSDT">ETHUSDT</option>
                <option value="BTCUSDT">BTCUSDT</option>
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                Trading Mode
              </label>
              <div style={{
                display: 'flex',
                gap: '8px',
                padding: '4px',
                background: 'var(--bg-secondary)',
                borderRadius: '8px',
              }}>
                <button
                  onClick={() => handleTradingChange('mode', 'paper')}
                  className="btn"
                  style={{
                    flex: 1,
                    background: currentConfig.trading.mode === 'paper' ? 'var(--accent-blue)' : 'transparent',
                    color: currentConfig.trading.mode === 'paper' ? 'white' : 'var(--text-secondary)',
                  }}
                >
                  Paper Trading
                </button>
                <button
                  onClick={() => handleTradingChange('mode', 'live')}
                  className="btn"
                  style={{
                    flex: 1,
                    background: currentConfig.trading.mode === 'live' ? 'var(--accent-yellow)' : 'transparent',
                    color: currentConfig.trading.mode === 'live' ? '#000' : 'var(--text-secondary)',
                  }}
                >
                  Live Trading
                </button>
              </div>
              {currentConfig.trading.mode === 'live' && (
                <div className="callout callout-warning" style={{ marginTop: '12px', padding: '12px' }}>
                  <span style={{ fontSize: '13px' }}>Warning: Live trading uses real funds</span>
                </div>
              )}
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                Primary Timeframe
              </label>
              <select
                value={currentConfig.trading.primaryTimeframe}
                onChange={(e) => handleTradingChange('primaryTimeframe', e.target.value)}
                className="input select"
                style={{ width: '100%' }}
              >
                {timeframes.map((tf) => (
                  <option key={tf} value={tf}>{tf}</option>
                ))}
              </select>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                Active Timeframes
              </label>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                {timeframes.map((tf) => (
                  <button
                    key={tf}
                    onClick={() => {
                      const current = currentConfig.trading.timeframes;
                      const updated = current.includes(tf)
                        ? current.filter((t) => t !== tf)
                        : [...current, tf];
                      handleTradingChange('timeframes', updated);
                    }}
                    className="btn btn-sm"
                    style={{
                      background: currentConfig.trading.timeframes.includes(tf)
                        ? 'var(--accent-blue)'
                        : 'var(--bg-secondary)',
                      color: currentConfig.trading.timeframes.includes(tf)
                        ? 'white'
                        : 'var(--text-secondary)',
                    }}
                  >
                    {tf}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                Initial Balance ($)
              </label>
              <input
                type="number"
                value={currentConfig.trading.initialBalance}
                onChange={(e) => handleTradingChange('initialBalance', parseFloat(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                step="1000"
                min="1000"
              />
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: '16px' }}>
              <div>
                <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                  Commission (%)
                </label>
                <input
                  type="number"
                  value={currentConfig.trading.commission * 100}
                  onChange={(e) => handleTradingChange('commission', parseFloat(e.target.value) / 100)}
                  className="input"
                  style={{ width: '100%' }}
                  step="0.01"
                  min="0"
                />
              </div>
              <div>
                <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                  Slippage (%)
                </label>
                <input
                  type="number"
                  value={currentConfig.trading.slippage * 100}
                  onChange={(e) => handleTradingChange('slippage', parseFloat(e.target.value) / 100)}
                  className="input"
                  style={{ width: '100%' }}
                  step="0.01"
                  min="0"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Indicator Settings Card */}
      <div className="card section" style={{ overflow: 'hidden' }}>
        <div style={{
          padding: isMobile ? '12px 16px' : '16px 20px',
          borderBottom: '1px solid var(--border-color)',
        }}>
          <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>Indicator Settings</h3>
        </div>
        <div style={{ padding: isMobile ? '12px' : '20px' }}>
          <div style={{ display: 'grid', gridTemplateColumns: isMobile ? 'repeat(2, 1fr)' : 'repeat(4, 1fr)', gap: isMobile ? '12px' : '20px' }}>
            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                RSI Period
              </label>
              <input
                type="number"
                value={currentConfig.indicators.rsiPeriod}
                onChange={(e) => handleIndicatorChange('rsiPeriod', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="50"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                MACD Fast
              </label>
              <input
                type="number"
                value={currentConfig.indicators.macdFast}
                onChange={(e) => handleIndicatorChange('macdFast', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="50"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                MACD Slow
              </label>
              <input
                type="number"
                value={currentConfig.indicators.macdSlow}
                onChange={(e) => handleIndicatorChange('macdSlow', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="100"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                MACD Signal
              </label>
              <input
                type="number"
                value={currentConfig.indicators.macdSignal}
                onChange={(e) => handleIndicatorChange('macdSignal', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="50"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                BB Period
              </label>
              <input
                type="number"
                value={currentConfig.indicators.bbPeriod}
                onChange={(e) => handleIndicatorChange('bbPeriod', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="100"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                BB Std Dev
              </label>
              <input
                type="number"
                value={currentConfig.indicators.bbStdDev}
                onChange={(e) => handleIndicatorChange('bbStdDev', parseFloat(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                step="0.5"
                min="0.5"
                max="5"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                ADX Period
              </label>
              <input
                type="number"
                value={currentConfig.indicators.adxPeriod}
                onChange={(e) => handleIndicatorChange('adxPeriod', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="50"
              />
            </div>

            <div>
              <label style={{ display: 'block', fontSize: '13px', color: 'var(--text-tertiary)', marginBottom: '8px' }}>
                ATR Period
              </label>
              <input
                type="number"
                value={currentConfig.indicators.atrPeriod}
                onChange={(e) => handleIndicatorChange('atrPeriod', parseInt(e.target.value))}
                className="input"
                style={{ width: '100%' }}
                min="2"
                max="50"
              />
            </div>
          </div>
        </div>
      </div>

      {/* API Configuration Card */}
      <div className="card" style={{ overflow: 'hidden' }}>
        <div style={{
          padding: isMobile ? '12px 16px' : '16px 20px',
          borderBottom: '1px solid var(--border-color)',
        }}>
          <h3 style={{ fontSize: isMobile ? '14px' : '15px', fontWeight: 600, margin: 0 }}>API Configuration</h3>
        </div>
        <div style={{ padding: isMobile ? '12px' : '20px' }}>
          <div className="callout callout-info">
            <div className="callout-icon">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="16" x2="12" y2="12" />
                <line x1="12" y1="8" x2="12.01" y2="8" />
              </svg>
            </div>
            <div className="callout-content">
              <p style={{ margin: 0, marginBottom: '12px' }}>
                API keys are configured via environment variables or config.yaml file for security.
                Please update the following environment variables:
              </p>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '8px 12px',
                  background: 'var(--bg-secondary)',
                  borderRadius: '6px',
                  fontSize: '13px',
                }}>
                  <code style={{ color: 'var(--accent-blue)', fontWeight: 500 }}>BINANCE_API_KEY</code>
                  <span style={{ color: 'var(--text-secondary)' }}>Your Binance API key</span>
                </div>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '8px 12px',
                  background: 'var(--bg-secondary)',
                  borderRadius: '6px',
                  fontSize: '13px',
                }}>
                  <code style={{ color: 'var(--accent-blue)', fontWeight: 500 }}>BINANCE_SECRET_KEY</code>
                  <span style={{ color: 'var(--text-secondary)' }}>Your Binance secret key</span>
                </div>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  padding: '8px 12px',
                  background: 'var(--bg-secondary)',
                  borderRadius: '6px',
                  fontSize: '13px',
                }}>
                  <code style={{ color: 'var(--accent-blue)', fontWeight: 500 }}>BINANCE_TESTNET</code>
                  <span style={{ color: 'var(--text-secondary)' }}>Set to "true" for testnet</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
