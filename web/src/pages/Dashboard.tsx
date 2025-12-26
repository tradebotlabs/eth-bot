import { useState, useMemo, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { TradingChart } from '../components/TradingChart';
import { useTradingStore } from '../stores/tradingStore';
import * as api from '../services/api';
import type { Candle } from '../types';

type BottomTab = 'activity' | 'positions' | 'trades' | 'signals';

export function Dashboard() {
  const [timeframe, setTimeframe] = useState('15m');
  const [bottomPanelOpen, setBottomPanelOpen] = useState(true);
  const [activeTab, setActiveTab] = useState<BottomTab>('activity');
  const [isMobile, setIsMobile] = useState(false);

  // Check for mobile viewport
  useEffect(() => {
    const checkMobile = () => setIsMobile(window.innerWidth <= 768);
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Get store state including live candle updates and current price
  const {
    positions, trades, signals, logs,
    candles: storeCandles, currentPrice, setCandles, accountStats, setAccountStats, setSignals
  } = useTradingStore();

  // Fetch initial dashboard data via REST API (fallback when WebSocket hasn't sent state yet)
  const { data: dashboardData } = useQuery({
    queryKey: ['dashboard'],
    queryFn: async () => {
      const res = await api.getDashboard();
      return res.data;
    },
    refetchInterval: 5000, // Refresh every 5s as backup
    staleTime: 2000,
  });

  // Set account stats from REST API if not already set by WebSocket
  useEffect(() => {
    if (dashboardData?.state && !accountStats) {
      const state = dashboardData.state;
      setAccountStats({
        equity: state.Equity || 10000,
        balance: state.AvailableBalance || 10000,
        unrealizedPnl: state.UnrealizedPnL || 0,
        realizedPnl: state.RealizedPnL || 0,
        totalTrades: state.TotalTrades || 0,
        winningTrades: Math.round((state.TotalTrades || 0) * (state.WinRate || 0)),
        losingTrades: Math.round((state.TotalTrades || 0) * (1 - (state.WinRate || 0))),
        winRate: state.WinRate || 0,
        profitFactor: 0,
        maxDrawdown: state.MaxDrawdown || 0,
        currentDrawdown: state.CurrentDrawdown || 0,
        sharpeRatio: 0,
        dailyPnl: state.DailyPnL || 0,
        weeklyPnl: 0,
      });
    }
  }, [dashboardData, accountStats, setAccountStats]);

  // Set signals from REST API if not already populated by WebSocket
  useEffect(() => {
    if (dashboardData?.signals && signals.length === 0) {
      // Convert backend SignalRecord format to frontend Signal format
      const formattedSignals = dashboardData.signals.map((record: { signal: unknown }) => record.signal);
      setSignals(formattedSignals);
    }
  }, [dashboardData, signals.length, setSignals]);

  // Fetch initial candles via REST API
  const { data: apiCandles = [], isLoading: isLoadingCandles } = useQuery<Candle[]>({
    queryKey: ['candles', 'ETHUSDT', timeframe],
    queryFn: async () => {
      const res = await api.getCandles('ETHUSDT', timeframe, 500);
      return res.data;
    },
    refetchInterval: 30000, // Refresh every 30s as backup
    staleTime: 10000,
  });

  // Store API candles in zustand store for WebSocket updates
  useEffect(() => {
    if (apiCandles.length > 0) {
      setCandles(timeframe, apiCandles);
    }
  }, [apiCandles, timeframe, setCandles]);

  // Get live candles from store (updated via WebSocket)
  const liveCandles = storeCandles.get(timeframe) || [];

  // Merge API candles with live updates - use live if available, otherwise API
  const candles = useMemo(() => {
    if (liveCandles.length > 0) {
      // Use store candles (includes WebSocket updates)
      return liveCandles;
    }
    return apiCandles;
  }, [liveCandles, apiCandles]);

  // Update last candle's close price with live WebSocket price for real-time effect
  const displayCandles = useMemo(() => {
    if (candles.length === 0 || currentPrice <= 0) return candles;

    // Clone candles and update the last one with live price
    const updated = [...candles];
    const lastIdx = updated.length - 1;
    if (lastIdx >= 0) {
      const lastCandle = { ...updated[lastIdx] };
      lastCandle.close = currentPrice;
      // Update high/low if current price exceeds them
      if (currentPrice > lastCandle.high) lastCandle.high = currentPrice;
      if (currentPrice < lastCandle.low) lastCandle.low = currentPrice;
      updated[lastIdx] = lastCandle;
    }
    return updated;
  }, [candles, currentPrice]);

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);

  const formatPnl = (pnl: number) => {
    const formatted = formatPrice(Math.abs(pnl));
    return pnl >= 0 ? `+${formatted}` : `-${formatted}`;
  };

  // Format duration from milliseconds
  const formatDuration = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) return `${days}d ${hours % 24}h`;
    if (hours > 0) return `${hours}h ${minutes % 60}m`;
    if (minutes > 0) return `${minutes}m`;
    return `${seconds}s`;
  };

  const tabs: { id: BottomTab; label: string; count: number }[] = [
    { id: 'activity', label: 'Bot Activity', count: logs.length },
    { id: 'positions', label: 'Positions', count: positions.length },
    { id: 'trades', label: 'Trades', count: trades.length },
    { id: 'signals', label: 'Signals', count: signals.length },
  ];

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      height: '100%',
      overflow: 'hidden',
      background: '#131722',
      position: 'relative',
    }}>
      {/* Main Chart Area */}
      <div style={{
        flex: 1,
        minHeight: 0,
        position: 'relative',
        display: 'flex',
      }}>
        {/* Chart Container */}
        <div style={{ flex: 1, minWidth: 0, position: 'relative' }}>
          {isLoadingCandles && displayCandles.length === 0 ? (
            <div style={{
              height: '100%',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              flexDirection: 'column',
              gap: '12px',
              color: '#787b86',
              background: '#131722',
            }}>
              <div style={{
                width: '60px',
                height: '60px',
                border: '3px solid #2a2e39',
                borderTopColor: '#2962ff',
                borderRadius: '50%',
                animation: 'spin 1s linear infinite',
              }} />
              <span style={{ fontSize: '13px' }}>Loading chart...</span>
            </div>
          ) : (
            <TradingChart
              candles={displayCandles}
              trades={trades}
              symbol="ETHUSDT"
              currentTimeframe={timeframe}
              onTimeframeChange={setTimeframe}
              height="100%"
            />
          )}
        </div>

      </div>

      {/* Bottom Panel - Tabbed Interface */}
      <div style={{
        height: bottomPanelOpen ? (isMobile ? '200px' : '180px') : (isMobile ? '40px' : '32px'),
        background: '#1e222d',
        borderTop: '1px solid #2a2e39',
        display: 'flex',
        flexDirection: 'column',
        flexShrink: 0,
        transition: 'height 0.2s ease',
      }}>
        {/* Tab Bar */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          height: isMobile ? '40px' : '32px',
          borderBottom: bottomPanelOpen ? '1px solid #2a2e39' : 'none',
          padding: isMobile ? '0 4px' : '0 8px',
          gap: isMobile ? '2px' : '4px',
          flexShrink: 0,
          overflowX: isMobile ? 'auto' : 'visible',
        }}>
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => {
                setActiveTab(tab.id);
                if (!bottomPanelOpen) setBottomPanelOpen(true);
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: isMobile ? '4px' : '6px',
                padding: isMobile ? '6px 8px' : '4px 10px',
                background: activeTab === tab.id && bottomPanelOpen ? '#2a2e39' : 'transparent',
                border: 'none',
                borderRadius: '4px',
                color: activeTab === tab.id && bottomPanelOpen ? '#d1d4dc' : '#787b86',
                fontSize: isMobile ? '10px' : '11px',
                cursor: 'pointer',
                transition: 'all 0.1s ease',
                flexShrink: 0,
                whiteSpace: 'nowrap',
              }}
            >
              {isMobile ? tab.label.split(' ')[0] : tab.label}
              {tab.count > 0 && (
                <span style={{
                  padding: '1px 5px',
                  background: activeTab === tab.id && bottomPanelOpen ? '#363a45' : '#2a2e39',
                  borderRadius: '3px',
                  fontSize: isMobile ? '9px' : '10px',
                }}>
                  {tab.count > 99 ? '99+' : tab.count}
                </span>
              )}
            </button>
          ))}

          <div style={{ flex: 1 }} />

          {/* Collapse Toggle */}
          <button
            onClick={() => setBottomPanelOpen(!bottomPanelOpen)}
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '24px',
              height: '24px',
              background: 'transparent',
              border: 'none',
              borderRadius: '4px',
              color: '#787b86',
              cursor: 'pointer',
            }}
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              {bottomPanelOpen ? (
                <polyline points="6 15 12 9 18 15" />
              ) : (
                <polyline points="6 9 12 15 18 9" />
              )}
            </svg>
          </button>
        </div>

        {/* Tab Content */}
        {bottomPanelOpen && (
          <div style={{ flex: 1, overflow: 'auto', padding: '8px' }}>
            {/* Bot Activity */}
            {activeTab === 'activity' && (
              <div style={{ fontFamily: 'monospace', fontSize: '11px' }}>
                {logs.length === 0 ? (
                  <div style={{ color: '#787b86', textAlign: 'center', padding: '20px' }}>
                    Waiting for bot activity...
                  </div>
                ) : (
                  logs.slice(0, 30).map((log) => (
                    <div
                      key={log.id}
                      style={{
                        display: 'flex',
                        alignItems: 'flex-start',
                        gap: '10px',
                        padding: '4px 8px',
                        borderRadius: '3px',
                        marginBottom: '2px',
                        background: log.type === 'error' ? 'rgba(239, 83, 80, 0.1)' :
                                    log.type === 'warning' ? 'rgba(255, 152, 0, 0.1)' :
                                    log.type === 'success' ? 'rgba(38, 166, 154, 0.1)' : 'transparent',
                      }}
                    >
                      <span style={{ color: '#4c525e', fontSize: '10px', flexShrink: 0 }}>
                        {log.timestamp.toLocaleTimeString('en-US', { hour12: false })}
                      </span>
                      <span style={{
                        color: log.type === 'error' ? '#ef5350' :
                               log.type === 'warning' ? '#ff9800' :
                               log.type === 'success' ? '#26a69a' :
                               log.type === 'signal' ? '#2962ff' :
                               log.type === 'trade' ? '#9c27b0' : '#787b86',
                      }}>
                        {log.message}
                        {log.details && <span style={{ color: '#4c525e', marginLeft: '6px' }}>({log.details})</span>}
                      </span>
                    </div>
                  ))
                )}
              </div>
            )}

            {/* Positions */}
            {activeTab === 'positions' && (
              <div>
                {positions.length === 0 ? (
                  <div style={{ color: '#787b86', textAlign: 'center', padding: '20px', fontSize: '12px' }}>
                    No open positions
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                    {positions.map((pos) => {
                      const duration = pos.entryTime ? Date.now() - new Date(pos.entryTime).getTime() : 0;
                      const pnlPercent = pos.entryPrice > 0 ? ((pos.currentPrice - pos.entryPrice) / pos.entryPrice) * 100 * (pos.side === 'LONG' ? 1 : -1) : 0;

                      return (
                        <div
                          key={pos.id}
                          style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '6px',
                            padding: '12px 16px',
                            borderRadius: '6px',
                            background: '#131722',
                            fontSize: '11px',
                            minWidth: '240px',
                          }}
                        >
                          {/* Header: Side + Quantity */}
                          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                              <span style={{
                                padding: '2px 8px',
                                borderRadius: '3px',
                                fontSize: '10px',
                                fontWeight: 600,
                                background: pos.side === 'LONG' ? 'rgba(38, 166, 154, 0.2)' : 'rgba(239, 83, 80, 0.2)',
                                color: pos.side === 'LONG' ? '#26a69a' : '#ef5350',
                              }}>
                                {pos.side}
                              </span>
                              <span style={{ color: '#d1d4dc', fontWeight: 500 }}>{pos.quantity.toFixed(4)} ETH</span>
                            </div>
                            <span style={{ color: '#787b86', fontSize: '10px' }}>{formatDuration(duration)}</span>
                          </div>

                          {/* Entry → Current Price */}
                          <div style={{ display: 'flex', alignItems: 'center', gap: '6px', color: '#787b86' }}>
                            <span>Entry: <span style={{ color: '#d1d4dc' }}>{formatPrice(pos.entryPrice)}</span></span>
                            <span style={{ color: '#4c525e' }}>→</span>
                            <span>Current: <span style={{ color: pos.currentPrice >= pos.entryPrice ? '#26a69a' : '#ef5350' }}>{formatPrice(pos.currentPrice || currentPrice)}</span></span>
                          </div>

                          {/* P&L */}
                          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <span style={{ color: '#787b86' }}>P&L:</span>
                            <span style={{
                              fontWeight: 600,
                              color: pos.unrealizedPnl >= 0 ? '#26a69a' : '#ef5350'
                            }}>
                              {formatPnl(pos.unrealizedPnl)}
                              <span style={{ fontSize: '10px', marginLeft: '4px' }}>
                                ({pnlPercent >= 0 ? '+' : ''}{pnlPercent.toFixed(2)}%)
                              </span>
                            </span>
                          </div>

                          {/* SL/TP */}
                          {(pos.stopLoss || pos.takeProfit) && (
                            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', color: '#787b86', fontSize: '10px' }}>
                              {pos.stopLoss && (
                                <span>SL: <span style={{ color: '#ef5350' }}>{formatPrice(pos.stopLoss)}</span></span>
                              )}
                              {pos.takeProfit && (
                                <span>TP: <span style={{ color: '#26a69a' }}>{formatPrice(pos.takeProfit)}</span></span>
                              )}
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}

            {/* Trades */}
            {activeTab === 'trades' && (
              <div>
                {trades.length === 0 ? (
                  <div style={{ color: '#787b86', textAlign: 'center', padding: '20px', fontSize: '12px' }}>
                    No trades yet
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                    {trades.slice(0, 10).map((trade) => (
                      <div
                        key={trade.id}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '12px',
                          padding: '10px 14px',
                          borderRadius: '6px',
                          background: '#131722',
                          fontSize: '12px',
                        }}
                      >
                        <span style={{
                          padding: '2px 8px',
                          borderRadius: '3px',
                          fontSize: '10px',
                          fontWeight: 600,
                          background: trade.side === 'LONG' ? 'rgba(38, 166, 154, 0.2)' : 'rgba(239, 83, 80, 0.2)',
                          color: trade.side === 'LONG' ? '#26a69a' : '#ef5350',
                        }}>
                          {trade.side}
                        </span>
                        <span style={{ color: '#787b86' }}>{trade.strategy}</span>
                        <span style={{
                          fontWeight: 600,
                          color: trade.pnl >= 0 ? '#26a69a' : '#ef5350'
                        }}>
                          {formatPnl(trade.pnl)}
                        </span>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Signals */}
            {activeTab === 'signals' && (
              <div>
                {signals.length === 0 ? (
                  <div style={{ color: '#787b86', textAlign: 'center', padding: '20px', fontSize: '12px' }}>
                    No signals yet
                  </div>
                ) : (
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px' }}>
                    {signals.slice(0, 10).map((signal, idx) => {
                      const timestamp = signal.timestamp ? new Date(signal.timestamp).toLocaleTimeString('en-US', { hour12: false }) : '';
                      const confidencePercent = Math.min(signal.confidence * 100, 100);

                      return (
                        <div
                          key={`signal-${idx}-${signal.strategy}-${signal.timestamp}`}
                          style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '6px',
                            padding: '12px 16px',
                            borderRadius: '6px',
                            background: '#131722',
                            fontSize: '11px',
                            minWidth: '220px',
                          }}
                        >
                          {/* Header: Direction + Strategy + Time */}
                          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                              <span style={{
                                padding: '2px 8px',
                                borderRadius: '3px',
                                fontSize: '10px',
                                fontWeight: 600,
                                background: signal.direction === 'LONG' ? 'rgba(38, 166, 154, 0.2)' :
                                           signal.direction === 'SHORT' ? 'rgba(239, 83, 80, 0.2)' : 'rgba(41, 98, 255, 0.2)',
                                color: signal.direction === 'LONG' ? '#26a69a' :
                                      signal.direction === 'SHORT' ? '#ef5350' : '#2962ff',
                              }}>
                                {signal.direction}
                              </span>
                              <span style={{ color: '#d1d4dc', fontWeight: 500 }}>{signal.strategy}</span>
                            </div>
                            <span style={{ color: '#787b86', fontSize: '10px' }}>{timestamp}</span>
                          </div>

                          {/* Entry, SL, TP prices */}
                          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', color: '#787b86', fontSize: '10px' }}>
                            {signal.price > 0 && (
                              <span>Entry: <span style={{ color: '#d1d4dc' }}>{formatPrice(signal.price)}</span></span>
                            )}
                            {signal.stopLoss > 0 && (
                              <span>SL: <span style={{ color: '#ef5350' }}>{formatPrice(signal.stopLoss)}</span></span>
                            )}
                            {signal.takeProfit > 0 && (
                              <span>TP: <span style={{ color: '#26a69a' }}>{formatPrice(signal.takeProfit)}</span></span>
                            )}
                          </div>

                          {/* Confidence bar */}
                          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                            <span style={{ color: '#787b86', fontSize: '10px' }}>Confidence:</span>
                            <div style={{
                              flex: 1,
                              height: '6px',
                              background: '#2a2e39',
                              borderRadius: '3px',
                              overflow: 'hidden',
                            }}>
                              <div style={{
                                width: `${confidencePercent}%`,
                                height: '100%',
                                background: confidencePercent >= 70 ? '#26a69a' :
                                           confidencePercent >= 50 ? '#ff9800' : '#ef5350',
                                borderRadius: '3px',
                                transition: 'width 0.3s ease',
                              }} />
                            </div>
                            <span style={{
                              color: confidencePercent >= 70 ? '#26a69a' :
                                    confidencePercent >= 50 ? '#ff9800' : '#ef5350',
                              fontWeight: 600,
                              fontSize: '11px',
                              minWidth: '32px',
                            }}>
                              {confidencePercent.toFixed(0)}%
                            </span>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* CSS Animation */}
      <style>{`
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
