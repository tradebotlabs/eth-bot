import { useEffect, useCallback } from 'react';
import { wsService } from '../services/websocket';
import { useTradingStore } from '../stores/tradingStore';
import type { WSMessage, CandleUpdate, Position, Trade, Signal, IndicatorValues, MarketRegime, Candle } from '../types';

// Backend state format
interface BackendState {
  Mode: number;
  IsRunning: boolean;
  IsPaused: boolean;
  CurrentPrice: number;
  Equity: number;
  AvailableBalance: number;
  UnrealizedPnL: number;
  RealizedPnL: number;
  DailyPnL: number;
  OpenPositions: number;
  TotalTrades: number;
  WinRate: number;
  CurrentDrawdown: number;
  MaxDrawdown: number;
  CurrentRegime: string;
}

// Backend StateUpdate format (what actually comes from WebSocket)
interface BackendStateUpdate {
  State?: BackendState;
  Summary?: {
    equity: number;
    availableBalance: number;
    unrealizedPnL: number;
    realizedPnL: number;
    dailyPnL: number;
    weeklyPnL: number;
    openPositions: number;
    totalTrades: number;
    winRate: number;
  };
}

// Backend indicators format
interface BackendIndicators {
  rsi: number;
  macd: { macd: number; signal: number; histogram: number };
  bb: { upper: number; middle: number; lower: number; width: number };
  adx: { adx: number; plusDI: number; minusDI: number };
  atr: number;
  regime: string;
}

// Backend candle format
interface BackendCandle {
  symbol: string;
  timeframe: string;
  timestamp: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  isClosed: boolean;
}

export function useWebSocket() {
  const {
    setWsConnected,
    setStatus,
    setAccountStats,
    setPositions,
    addPosition,
    updatePosition,
    removePosition,
    addTrade,
    updateCandle,
    setIndicators,
    setRegime,
    addSignal,
    addLog,
  } = useTradingStore();

  const handleMessage = useCallback(
    (message: WSMessage) => {
      switch (message.type) {
        case 'state': {
          // Handle backend StateUpdate format (nested State and Summary)
          const stateUpdate = message.data as BackendStateUpdate;
          const state = stateUpdate.State;
          const summary = stateUpdate.Summary;

          if (state) {
            setStatus({
              running: state.IsRunning,
              mode: state.Mode === 0 ? 'paper' : 'live',
              currentPrice: state.CurrentPrice,
            });

            setAccountStats({
              equity: summary?.equity ?? state.Equity,
              balance: summary?.availableBalance ?? state.AvailableBalance,
              unrealizedPnl: summary?.unrealizedPnL ?? state.UnrealizedPnL,
              realizedPnl: summary?.realizedPnL ?? state.RealizedPnL,
              totalTrades: summary?.totalTrades ?? state.TotalTrades,
              winningTrades: Math.round((summary?.totalTrades ?? state.TotalTrades) * (summary?.winRate ?? state.WinRate)),
              losingTrades: Math.round((summary?.totalTrades ?? state.TotalTrades) * (1 - (summary?.winRate ?? state.WinRate))),
              winRate: summary?.winRate ?? state.WinRate,
              profitFactor: 0,
              maxDrawdown: state.MaxDrawdown,
              currentDrawdown: state.CurrentDrawdown,
              sharpeRatio: 0,
              dailyPnl: summary?.dailyPnL ?? state.DailyPnL,
              weeklyPnl: summary?.weeklyPnL ?? 0,
            });

            if (state.CurrentRegime) {
              const regimeMap: Record<string, MarketRegime['regime']> = {
                'trending_up': 'trending_up',
                'trending_down': 'trending_down',
                'ranging': 'ranging',
                'volatile': 'volatile',
              };
              setRegime({
                regime: regimeMap[state.CurrentRegime] || 'unknown',
                strength: 0.5,
                volatility: 0,
                trend: 0,
              });
            }
          }
          break;
        }

        case 'status':
          // Legacy format support
          setStatus(message.data as Parameters<typeof setStatus>[0]);
          break;

        case 'account':
          setAccountStats(message.data as Parameters<typeof setAccountStats>[0]);
          break;

        case 'positions':
          setPositions(message.data as Position[]);
          break;

        case 'position': {
          // Handle backend position format with eventType
          const posData = message.data as Position & { eventType?: string };
          if (posData.eventType === 'opened') {
            addPosition(posData);
          } else if (posData.eventType === 'closed') {
            removePosition(posData.id);
          } else {
            updatePosition(posData.id, posData);
          }
          break;
        }

        case 'position_opened':
          addPosition(message.data as Position);
          break;

        case 'position_updated': {
          const posUpdate = message.data as Position;
          updatePosition(posUpdate.id, posUpdate);
          break;
        }

        case 'position_closed': {
          const closedPos = message.data as { id: string };
          removePosition(closedPos.id);
          break;
        }

        case 'trade':
          addTrade(message.data as Trade);
          break;

        case 'candle': {
          // Handle backend candle format
          const candleData = message.data as BackendCandle | CandleUpdate;
          if ('timestamp' in candleData && typeof candleData.timestamp === 'string') {
            // Backend format - convert to frontend format
            const candle: Candle = {
              time: new Date(candleData.timestamp).getTime(),
              open: candleData.open,
              high: candleData.high,
              low: candleData.low,
              close: candleData.close,
              volume: candleData.volume,
            };
            updateCandle(candleData.timeframe, candle);
            // Log candle updates for visibility
            if (candleData.isClosed) {
              addLog({ type: 'info', message: `Candle closed: ${candleData.timeframe}`, details: `$${candleData.close.toFixed(2)}` });
            }
          } else if ('candle' in candleData) {
            // Already in frontend format
            updateCandle(candleData.timeframe, candleData.candle);
          }
          break;
        }

        case 'indicators': {
          // Handle backend indicators format
          const ind = message.data as BackendIndicators | IndicatorValues;
          if ('bb' in ind && ind.bb) {
            // Backend format - convert to frontend format
            setIndicators({
              rsi: ind.rsi,
              macd: ind.macd,
              bollinger: {
                upper: ind.bb.upper,
                middle: ind.bb.middle,
                lower: ind.bb.lower,
                width: ind.bb.width,
              },
              adx: {
                adx: ind.adx.adx,
                plusDI: ind.adx.plusDI,
                minusDI: ind.adx.minusDI,
              },
              atr: ind.atr,
              ema20: 0,
              ema50: 0,
              sma200: 0,
            });
            if (ind.regime) {
              const regimeMap: Record<string, MarketRegime['regime']> = {
                'trending_up': 'trending_up',
                'trending_down': 'trending_down',
                'ranging': 'ranging',
                'volatile': 'volatile',
              };
              setRegime({
                regime: regimeMap[ind.regime] || 'unknown',
                strength: 0.5,
                volatility: 0,
                trend: 0,
              });
            }
          } else if ('bollinger' in ind) {
            // Already in frontend format
            setIndicators(ind as IndicatorValues);
          }
          break;
        }

        case 'regime':
          setRegime(message.data as MarketRegime);
          break;

        case 'signal': {
          // Handle backend signal format
          const signalData = message.data as { signal?: Signal; approved?: boolean; reason?: string } | Signal;
          if ('signal' in signalData && signalData.signal) {
            addSignal(signalData.signal);
            const sig = signalData.signal;
            const status = signalData.approved ? 'approved' : 'rejected';
            addLog({
              type: 'signal',
              message: `Signal: ${sig.direction} ${sig.strategy}`,
              details: `${status} - ${signalData.reason || 'confidence: ' + Math.min(sig.confidence * 100, 100).toFixed(0) + '%'}`
            });
          } else {
            const sig = signalData as Signal;
            addSignal(sig);
            addLog({ type: 'signal', message: `Signal: ${sig.direction} ${sig.strategy}` });
          }
          break;
        }

        case 'risk': {
          const riskData = message.data as { isHalted?: boolean; haltReason?: string; level?: number };
          if (riskData.isHalted) {
            addLog({ type: 'warning', message: 'Trading halted', details: riskData.haltReason });
          }
          break;
        }

        case 'error': {
          const errorData = message.data as { message?: string; code?: string };
          addLog({ type: 'error', message: errorData.message || 'Unknown error', details: errorData.code });
          break;
        }

        case 'price': {
          // Real-time price update (high frequency)
          const priceData = message.data as { symbol: string; price: number; timestamp: string };
          setStatus({ currentPrice: priceData.price });
          break;
        }

        case 'pong':
          // Server response to ping - no action needed
          break;

        default:
          console.log('Unknown message type:', message.type);
      }
    },
    [
      setStatus,
      setAccountStats,
      setPositions,
      addPosition,
      updatePosition,
      removePosition,
      addTrade,
      updateCandle,
      setIndicators,
      setRegime,
      addSignal,
      addLog,
    ]
  );

  useEffect(() => {
    // Connect WebSocket (singleton - will reuse existing connection)
    wsService.connect();

    // Subscribe to all messages
    const unsubscribe = wsService.subscribe('*', handleMessage);

    // Track if we've logged connection status to avoid duplicates
    let hasLoggedConnect = false;
    let hasLoggedDisconnect = false;

    // Handle connection status
    const unsubOpen = wsService.onOpen(() => {
      setWsConnected(true);
      if (!hasLoggedConnect) {
        hasLoggedConnect = true;
        hasLoggedDisconnect = false;
        addLog({ type: 'success', message: 'Connected to trading server' });
      }
    });
    const unsubClose = wsService.onClose(() => {
      setWsConnected(false);
      if (!hasLoggedDisconnect) {
        hasLoggedDisconnect = true;
        hasLoggedConnect = false;
        addLog({ type: 'warning', message: 'Disconnected from trading server' });
      }
    });

    // If already connected, update state
    if (wsService.isConnected()) {
      setWsConnected(true);
    }

    return () => {
      unsubscribe();
      unsubOpen();
      unsubClose();
      // Don't disconnect on cleanup - keep connection alive across HMR
    };
  }, [handleMessage, setWsConnected, addLog]);

  return {
    isConnected: wsService.isConnected(),
    send: wsService.send.bind(wsService),
  };
}
