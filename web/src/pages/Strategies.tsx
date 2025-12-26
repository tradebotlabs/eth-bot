import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import type { DragEndEvent } from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { ResponsiveContainer, AreaChart, Area } from 'recharts';
import * as api from '../services/api';
import type { Strategy } from '../types';

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

// Generate mock performance data for sparklines
const generatePerformanceData = (days: number = 30, trend: number = 0) => {
  const data = [];
  let value = 100;
  for (let i = 0; i < days; i++) {
    value += (Math.random() - 0.45 + trend * 0.1) * 5;
    data.push({ day: i, value: Math.max(50, value) });
  }
  return data;
};

interface SortableStrategyCardProps {
  strategy: Strategy;
  onToggle: (strategy: Strategy) => void;
  isToggling: boolean;
  performanceData: { day: number; value: number }[];
  isMobile: boolean;
}

function SortableStrategyCard({ strategy, onToggle, isToggling, performanceData, isMobile }: SortableStrategyCardProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: strategy.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);

  const formatPriceCompact = (price: number) => {
    if (Math.abs(price) >= 1000) {
      return `${price >= 0 ? '+' : ''}$${(price / 1000).toFixed(1)}k`;
    }
    return `${price >= 0 ? '+' : ''}${formatPrice(price)}`;
  };

  const formatPercent = (value: number) => `${(value * 100).toFixed(1)}%`;

  const performance = strategy.performance || { totalPnl: 0, winRate: 0, profitFactor: 0, totalTrades: 0 };
  const isPositive = performance.totalPnl >= 0;
  const sparklineColor = isPositive ? '#26a69a' : '#ef5350';

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="card"
      {...attributes}
    >
      {/* Header with drag handle */}
      <div style={{
        padding: isMobile ? '12px' : '16px 20px',
        borderBottom: '1px solid var(--border-color)',
        display: 'flex',
        flexDirection: isMobile ? 'column' : 'row',
        alignItems: isMobile ? 'stretch' : 'center',
        gap: isMobile ? '10px' : '16px',
      }}>
        {/* Mobile: Top row with drag handle, name, toggle */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: isMobile ? '10px' : '16px',
          width: '100%',
        }}>
          {/* Drag Handle */}
          <div
            {...listeners}
            style={{
              cursor: 'grab',
              padding: '8px',
              margin: '-8px',
              color: 'var(--text-tertiary)',
              display: 'flex',
              alignItems: 'center',
              touchAction: 'none',
            }}
            title="Drag to reorder"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <circle cx="9" cy="6" r="1.5" />
              <circle cx="15" cy="6" r="1.5" />
              <circle cx="9" cy="12" r="1.5" />
              <circle cx="15" cy="12" r="1.5" />
              <circle cx="9" cy="18" r="1.5" />
              <circle cx="15" cy="18" r="1.5" />
            </svg>
          </div>

          {/* Strategy Info */}
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '2px', flexWrap: 'wrap' }}>
              <h3 style={{ fontSize: isMobile ? '14px' : '16px', fontWeight: 600, margin: 0 }}>{strategy.name}</h3>
              <span className={`badge ${strategy.enabled ? 'badge-success' : ''}`} style={{ fontSize: isMobile ? '9px' : '11px' }}>
                {strategy.enabled ? 'ACTIVE' : 'DISABLED'}
              </span>
            </div>
            {!isMobile && (
              <p style={{ fontSize: '13px', color: 'var(--text-secondary)', margin: 0 }}>
                {strategy.description}
              </p>
            )}
          </div>

          {/* Mini Performance Chart - Hidden on mobile */}
          {!isMobile && (
            <div style={{ width: '120px', height: '40px', flexShrink: 0 }}>
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={performanceData}>
                  <defs>
                    <linearGradient id={`gradient-${strategy.id}`} x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor={sparklineColor} stopOpacity={0.3}/>
                      <stop offset="95%" stopColor={sparklineColor} stopOpacity={0}/>
                    </linearGradient>
                  </defs>
                  <Area
                    type="monotone"
                    dataKey="value"
                    stroke={sparklineColor}
                    fill={`url(#gradient-${strategy.id})`}
                    strokeWidth={1.5}
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          )}

          {/* P&L Summary */}
          <div style={{ textAlign: 'right', minWidth: isMobile ? '70px' : '100px', flexShrink: 0 }}>
            <div style={{
              fontSize: isMobile ? '14px' : '18px',
              fontWeight: 600,
              color: isPositive ? 'var(--accent-green)' : 'var(--accent-red)',
            }}>
              {isMobile ? formatPriceCompact(performance.totalPnl) : (isPositive ? '+' : '') + formatPrice(performance.totalPnl)}
            </div>
            <div style={{ fontSize: isMobile ? '10px' : '12px', color: 'var(--text-tertiary)' }}>
              {performance.totalTrades} trades
            </div>
          </div>

          {/* Toggle */}
          <button
            onClick={() => onToggle(strategy)}
            className={`toggle ${strategy.enabled ? 'active' : ''}`}
            disabled={isToggling}
            style={{ flexShrink: 0 }}
          />
        </div>
      </div>

      {/* Metrics Grid */}
      <div style={{ padding: isMobile ? '12px' : '16px 20px' }}>
        <div style={{
          display: 'grid',
          gridTemplateColumns: isMobile ? 'repeat(3, 1fr)' : 'repeat(5, 1fr)',
          gap: isMobile ? '8px' : '12px'
        }}>
          <div style={{ padding: isMobile ? '8px' : '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
            <div style={{ fontSize: isMobile ? '9px' : '11px', color: 'var(--text-tertiary)', marginBottom: '4px', textTransform: 'uppercase' }}>Win Rate</div>
            <div style={{ fontSize: isMobile ? '14px' : '18px', fontWeight: 600 }}>{formatPercent(performance.winRate)}</div>
          </div>
          <div style={{ padding: isMobile ? '8px' : '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
            <div style={{ fontSize: isMobile ? '9px' : '11px', color: 'var(--text-tertiary)', marginBottom: '4px', textTransform: 'uppercase' }}>Profit Factor</div>
            <div style={{
              fontSize: isMobile ? '14px' : '18px',
              fontWeight: 600,
              color: performance.profitFactor >= 1.5 ? 'var(--accent-green)' : performance.profitFactor >= 1 ? 'var(--text-primary)' : 'var(--accent-red)'
            }}>
              {performance.profitFactor.toFixed(2)}
            </div>
          </div>
          <div style={{ padding: isMobile ? '8px' : '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
            <div style={{ fontSize: isMobile ? '9px' : '11px', color: 'var(--text-tertiary)', marginBottom: '4px', textTransform: 'uppercase' }}>Avg Win</div>
            <div style={{ fontSize: isMobile ? '14px' : '18px', fontWeight: 600, color: 'var(--accent-green)' }}>
              {isMobile ? formatPriceCompact((performance as any).avgWin || 0) : formatPrice((performance as any).avgWin || 0)}
            </div>
          </div>
          <div style={{ padding: isMobile ? '8px' : '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
            <div style={{ fontSize: isMobile ? '9px' : '11px', color: 'var(--text-tertiary)', marginBottom: '4px', textTransform: 'uppercase' }}>Avg Loss</div>
            <div style={{ fontSize: isMobile ? '14px' : '18px', fontWeight: 600, color: 'var(--accent-red)' }}>
              {isMobile ? formatPriceCompact((performance as any).avgLoss || 0) : formatPrice((performance as any).avgLoss || 0)}
            </div>
          </div>
          <div style={{ padding: isMobile ? '8px' : '12px', background: 'var(--bg-secondary)', borderRadius: '8px', textAlign: 'center' }}>
            <div style={{ fontSize: isMobile ? '9px' : '11px', color: 'var(--text-tertiary)', marginBottom: '4px', textTransform: 'uppercase' }}>Max DD</div>
            <div style={{ fontSize: isMobile ? '14px' : '18px', fontWeight: 600, color: 'var(--accent-red)' }}>
              {formatPercent((performance as any).maxDrawdown || 0)}
            </div>
          </div>
        </div>

        {/* Parameters Section */}
        {strategy.parameters && Object.keys(strategy.parameters).length > 0 && (
          <div style={{ marginTop: '16px', paddingTop: '16px', borderTop: '1px solid var(--border-color)' }}>
            <div style={{ fontSize: '11px', color: 'var(--text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '10px' }}>
              Parameters
            </div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px' }}>
              {Object.entries(strategy.parameters || {}).map(([key, value]) => (
                <div
                  key={key}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '6px',
                    padding: '4px 10px',
                    background: 'var(--bg-secondary)',
                    borderRadius: '4px',
                    fontSize: '12px',
                  }}
                >
                  <span style={{ color: 'var(--text-tertiary)' }}>
                    {key.replace(/([A-Z])/g, ' $1').trim()}:
                  </span>
                  <span style={{ fontWeight: 500 }}>{value}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export function Strategies() {
  const queryClient = useQueryClient();
  const [strategyOrder, setStrategyOrder] = useState<string[]>([]);
  const isMobile = useIsMobile();

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8, // Require 8px movement before drag starts
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 200, // 200ms delay for touch to differentiate from scroll
        tolerance: 5,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const { data: strategies = [], isLoading } = useQuery<Strategy[]>({
    queryKey: ['strategies'],
    queryFn: async () => {
      const res = await api.getStrategies();
      return res.data;
    },
  });

  // Initialize order when strategies load
  useMemo(() => {
    if (strategies.length > 0 && strategyOrder.length === 0) {
      setStrategyOrder(strategies.map(s => s.id));
    }
  }, [strategies, strategyOrder.length]);

  // Generate performance data for each strategy
  const performanceDataMap = useMemo(() => {
    const map: Record<string, { day: number; value: number }[]> = {};
    strategies.forEach(s => {
      const perf = s.performance || { totalPnl: 0 };
      const trend = perf.totalPnl >= 0 ? 0.3 : -0.3;
      map[s.id] = generatePerformanceData(30, trend);
    });
    return map;
  }, [strategies]);

  const sortedStrategies = useMemo(() => {
    if (strategyOrder.length === 0) return strategies;
    return [...strategies].sort((a, b) => {
      const aIndex = strategyOrder.indexOf(a.id);
      const bIndex = strategyOrder.indexOf(b.id);
      return aIndex - bIndex;
    });
  }, [strategies, strategyOrder]);

  const enableMutation = useMutation({
    mutationFn: (name: string) => api.enableStrategy(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['strategies'] }),
  });

  const disableMutation = useMutation({
    mutationFn: (name: string) => api.disableStrategy(name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['strategies'] }),
  });

  const toggleStrategy = (strategy: Strategy) => {
    if (strategy.enabled) {
      disableMutation.mutate(strategy.id);
    } else {
      enableMutation.mutate(strategy.id);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (over && active.id !== over.id) {
      setStrategyOrder((items) => {
        const oldIndex = items.indexOf(active.id as string);
        const newIndex = items.indexOf(over.id as string);
        return arrayMove(items, oldIndex, newIndex);
      });
    }
  };

  // Calculate totals
  const totals = useMemo(() => {
    const activeStrategies = strategies.filter(s => s.enabled);
    return {
      active: activeStrategies.length,
      total: strategies.length,
      totalPnl: strategies.reduce((sum, s) => sum + (s.performance?.totalPnl || 0), 0),
      totalTrades: strategies.reduce((sum, s) => sum + (s.performance?.totalTrades || 0), 0),
      avgWinRate: strategies.length > 0
        ? strategies.reduce((sum, s) => sum + (s.performance?.winRate || 0), 0) / strategies.length
        : 0,
    };
  }, [strategies]);

  const formatPrice = (price: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price);

  if (isLoading) {
    return (
      <div className="animate-fade-in">
        <div className="page-header">
          <div className="skeleton" style={{ width: '200px', height: '32px', marginBottom: '8px' }} />
          <div className="skeleton" style={{ width: '300px', height: '20px' }} />
        </div>
        <div className="grid-stats section">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton" style={{ height: '80px', borderRadius: '8px' }} />
          ))}
        </div>
        <div style={{ display: 'grid', gap: '12px' }}>
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton" style={{ height: '180px', borderRadius: '8px' }} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="animate-fade-in">
      <div className="page-header">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <h1 className="page-title">Trading Strategies</h1>
            <p className="page-description">Configure, monitor, and reorder your automated trading strategies</p>
          </div>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid-stats section">
        <div className="stat-card">
          <div className="stat-label">Active Strategies</div>
          <div className="stat-value">{totals.active} / {totals.total}</div>
          <div className="stat-change text-muted">
            {totals.active === totals.total ? 'All active' : `${totals.total - totals.active} disabled`}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total P&L</div>
          <div className={`stat-value ${totals.totalPnl >= 0 ? 'profit' : 'loss'}`}>
            {totals.totalPnl >= 0 ? '+' : ''}{formatPrice(totals.totalPnl)}
          </div>
          <div className="stat-change text-muted">Across all strategies</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total Trades</div>
          <div className="stat-value">{totals.totalTrades}</div>
          <div className="stat-change text-muted">Combined trades</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Avg Win Rate</div>
          <div className="stat-value">{(totals.avgWinRate * 100).toFixed(1)}%</div>
          <div className="stat-change text-muted">Average across strategies</div>
        </div>
      </div>

      {/* Drag hint */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: '8px',
        padding: isMobile ? '8px 12px' : '10px 16px',
        background: 'var(--bg-secondary)',
        borderRadius: '8px',
        marginBottom: isMobile ? '12px' : '16px',
        fontSize: isMobile ? '11px' : '13px',
        color: 'var(--text-secondary)',
      }}>
        <svg width={isMobile ? 14 : 16} height={isMobile ? 14 : 16} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M12 5v14M5 12h14" />
        </svg>
        {isMobile ? 'Hold & drag to reorder' : 'Drag strategies to reorder priority. Higher strategies are evaluated first.'}
      </div>

      {/* Strategy List with Drag and Drop */}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragEnd={handleDragEnd}
      >
        <SortableContext
          items={sortedStrategies.map(s => s.id)}
          strategy={verticalListSortingStrategy}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            {sortedStrategies.map((strategy) => (
              <SortableStrategyCard
                key={strategy.id}
                strategy={strategy}
                onToggle={toggleStrategy}
                isToggling={enableMutation.isPending || disableMutation.isPending}
                performanceData={performanceDataMap[strategy.id] || []}
                isMobile={isMobile}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>
    </div>
  );
}
