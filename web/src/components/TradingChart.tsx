import { useEffect, useRef, useCallback, useState, useMemo } from 'react';
import {
  createChart,
  ColorType,
  CrosshairMode,
  CandlestickSeries,
  LineSeries,
  HistogramSeries,
  createSeriesMarkers,
} from 'lightweight-charts';
import type { IChartApi, ISeriesApi, CandlestickData, Time, LineData, SeriesMarker, ISeriesMarkersPluginApi } from 'lightweight-charts';
import type { Candle, IndicatorValues, Trade } from '../types';

interface TradingChartProps {
  candles: Candle[];
  indicators?: IndicatorValues;
  trades?: Trade[];
  symbol?: string;
  onTimeframeChange?: (timeframe: string) => void;
  currentTimeframe?: string;
  height?: number | string;
}

const TIMEFRAMES = ['1m', '5m', '15m', '1h', '4h', '1d'];

export function TradingChart({
  candles,
  indicators: _indicators, // Reserved for future backend indicator overlay
  trades = [],
  symbol = 'ETHUSDT',
  onTimeframeChange,
  currentTimeframe = '15m',
  height = 500,
}: TradingChartProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const chartRef = useRef<IChartApi | null>(null);
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null);
  const markersPluginRef = useRef<ISeriesMarkersPluginApi<Time> | null>(null);
  const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
  const ema20Ref = useRef<ISeriesApi<'Line'> | null>(null);
  const ema50Ref = useRef<ISeriesApi<'Line'> | null>(null);
  const bbUpperRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbLowerRef = useRef<ISeriesApi<'Line'> | null>(null);
  const bbMiddleRef = useRef<ISeriesApi<'Line'> | null>(null);
  const dataLoadedRef = useRef(false);

  const [showVolume, setShowVolume] = useState(true);
  const [showEMA, setShowEMA] = useState(true);
  const [showBB, setShowBB] = useState(false);
  // OHLC data - always show last candle, update on hover
  const [hoveredOHLC, setHoveredOHLC] = useState<{
    time: string;
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
  } | null>(null);

  // Memoize candle data transformation - deduplicate and sort by time
  const candleData = useMemo(() => {
    // Deduplicate by time (keep last occurrence) and sort ascending
    const timeMap = new Map<number, Candle>();
    candles.forEach(candle => {
      const time = Math.floor(candle.time / 1000);
      timeMap.set(time, candle);
    });

    return Array.from(timeMap.entries())
      .sort((a, b) => a[0] - b[0])
      .map(([time, candle]): CandlestickData<Time> => ({
        time: time as Time,
        open: candle.open,
        high: candle.high,
        low: candle.low,
        close: candle.close,
      }));
  }, [candles]);

  const volumeData = useMemo(() => {
    // Deduplicate by time (keep last occurrence) and sort ascending
    const timeMap = new Map<number, Candle>();
    candles.forEach(c => {
      const time = Math.floor(c.time / 1000);
      timeMap.set(time, c);
    });

    return Array.from(timeMap.entries())
      .sort((a, b) => a[0] - b[0])
      .map(([time, c]) => ({
        time: time as Time,
        value: c.volume,
        color: c.close >= c.open ? 'rgba(38, 166, 154, 0.5)' : 'rgba(239, 83, 80, 0.5)',
      }));
  }, [candles]);

  // Helper: get sorted unique candles
  const sortedCandles = useMemo(() => {
    const timeMap = new Map<number, Candle>();
    candles.forEach(c => {
      const time = Math.floor(c.time / 1000);
      timeMap.set(time, c);
    });
    return Array.from(timeMap.entries())
      .sort((a, b) => a[0] - b[0])
      .map(([time, c]) => ({ ...c, time: time * 1000 }));
  }, [candles]);

  // Calculate EMAs for display
  const emaData = useMemo(() => {
    if (sortedCandles.length < 50) return { ema20: [], ema50: [] };

    const ema20: LineData<Time>[] = [];
    const ema50: LineData<Time>[] = [];

    let ema20Value = sortedCandles.slice(0, 20).reduce((s, c) => s + c.close, 0) / 20;
    let ema50Value = sortedCandles.slice(0, 50).reduce((s, c) => s + c.close, 0) / 50;

    const k20 = 2 / (20 + 1);
    const k50 = 2 / (50 + 1);

    for (let i = 20; i < sortedCandles.length; i++) {
      ema20Value = sortedCandles[i].close * k20 + ema20Value * (1 - k20);
      ema20.push({ time: Math.floor(sortedCandles[i].time / 1000) as Time, value: ema20Value });
    }

    for (let i = 50; i < sortedCandles.length; i++) {
      ema50Value = sortedCandles[i].close * k50 + ema50Value * (1 - k50);
      ema50.push({ time: Math.floor(sortedCandles[i].time / 1000) as Time, value: ema50Value });
    }

    return { ema20, ema50 };
  }, [sortedCandles]);

  // Calculate trade markers for the chart
  const tradeMarkers = useMemo((): SeriesMarker<Time>[] => {
    if (!trades || trades.length === 0) return [];

    const markers: SeriesMarker<Time>[] = [];

    trades.forEach((trade) => {
      // Entry marker
      if (trade.entryTime) {
        const entryTime = Math.floor(new Date(trade.entryTime).getTime() / 1000) as Time;
        markers.push({
          time: entryTime,
          position: trade.side === 'LONG' ? 'belowBar' : 'aboveBar',
          color: trade.side === 'LONG' ? '#26a69a' : '#ef5350',
          shape: trade.side === 'LONG' ? 'arrowUp' : 'arrowDown',
          text: trade.side === 'LONG' ? 'BUY' : 'SELL',
        });
      }

      // Exit marker (if trade is closed)
      if (trade.exitTime) {
        const exitTime = Math.floor(new Date(trade.exitTime).getTime() / 1000) as Time;
        markers.push({
          time: exitTime,
          position: 'inBar',
          color: trade.pnl >= 0 ? '#26a69a' : '#ef5350',
          shape: 'circle',
          text: trade.pnl >= 0 ? `+${trade.pnl.toFixed(2)}` : `${trade.pnl.toFixed(2)}`,
        });
      }
    });

    // Sort markers by time
    return markers.sort((a, b) => (a.time as number) - (b.time as number));
  }, [trades]);

  // Calculate Bollinger Bands
  const bbData = useMemo(() => {
    if (sortedCandles.length < 20) return { upper: [], middle: [], lower: [] };

    const period = 20;
    const stdDev = 2;
    const upper: LineData<Time>[] = [];
    const middle: LineData<Time>[] = [];
    const lower: LineData<Time>[] = [];

    for (let i = period - 1; i < sortedCandles.length; i++) {
      const slice = sortedCandles.slice(i - period + 1, i + 1);
      const sma = slice.reduce((s, c) => s + c.close, 0) / period;
      const variance = slice.reduce((s, c) => s + Math.pow(c.close - sma, 2), 0) / period;
      const std = Math.sqrt(variance);

      const time = Math.floor(sortedCandles[i].time / 1000) as Time;
      middle.push({ time, value: sma });
      upper.push({ time, value: sma + stdDev * std });
      lower.push({ time, value: sma - stdDev * std });
    }

    return { upper, middle, lower };
  }, [sortedCandles]);

  const formatPrice = useCallback((price: number) => {
    if (price >= 1000) return price.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (price >= 1) return price.toFixed(2);
    return price.toFixed(6);
  }, []);

  // Initialize chart
  useEffect(() => {
    if (!chartContainerRef.current) return;

    // TradingView exact dark theme colors
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#131722' },
        textColor: '#787b86',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Trebuchet MS", Roboto, Ubuntu, sans-serif',
        fontSize: 12,
      },
      grid: {
        vertLines: { color: '#1e222d' },
        horzLines: { color: '#1e222d' },
      },
      crosshair: {
        mode: CrosshairMode.Normal,
        vertLine: {
          color: '#758696',
          width: 1,
          style: 0,
          labelBackgroundColor: '#2962ff',
        },
        horzLine: {
          color: '#758696',
          width: 1,
          style: 0,
          labelBackgroundColor: '#2962ff',
        },
      },
      rightPriceScale: {
        borderColor: '#2a2e39',
        scaleMargins: { top: 0.1, bottom: 0.2 },
        borderVisible: true,
      },
      timeScale: {
        borderColor: '#2a2e39',
        timeVisible: true,
        secondsVisible: false,
        barSpacing: 8,
        minBarSpacing: 4,
        borderVisible: true,
        rightOffset: 5,
      },
      handleScroll: { vertTouchDrag: false },
      handleScale: { axisPressedMouseMove: true },
    });

    chartRef.current = chart;

    // Candlestick series - TradingView exact colors
    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderUpColor: '#26a69a',
      borderDownColor: '#ef5350',
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    });
    candleSeriesRef.current = candleSeries;

    // Create markers plugin for trade entry/exit markers
    const markersPlugin = createSeriesMarkers(candleSeries, []);
    markersPluginRef.current = markersPlugin;

    // Volume series
    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceFormat: { type: 'volume' },
      priceScaleId: 'volume',
    });
    volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.85, bottom: 0 },
    });
    volumeSeriesRef.current = volumeSeries;

    // EMA 20 - TradingView orange
    const ema20 = chart.addSeries(LineSeries, {
      color: '#ff9800',
      lineWidth: 1,
      priceScaleId: 'right',
      crosshairMarkerVisible: false,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    ema20Ref.current = ema20;

    // EMA 50 - TradingView purple
    const ema50 = chart.addSeries(LineSeries, {
      color: '#9c27b0',
      lineWidth: 1,
      priceScaleId: 'right',
      crosshairMarkerVisible: false,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    ema50Ref.current = ema50;

    // Bollinger Bands - TradingView blue
    const bbUpper = chart.addSeries(LineSeries, {
      color: 'rgba(33, 150, 243, 0.6)',
      lineWidth: 1,
      lineStyle: 0,
      priceScaleId: 'right',
      crosshairMarkerVisible: false,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    bbUpperRef.current = bbUpper;

    const bbMiddle = chart.addSeries(LineSeries, {
      color: 'rgba(33, 150, 243, 0.4)',
      lineWidth: 1,
      lineStyle: 2,
      priceScaleId: 'right',
      crosshairMarkerVisible: false,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    bbMiddleRef.current = bbMiddle;

    const bbLower = chart.addSeries(LineSeries, {
      color: 'rgba(33, 150, 243, 0.6)',
      lineWidth: 1,
      lineStyle: 0,
      priceScaleId: 'right',
      crosshairMarkerVisible: false,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    bbLowerRef.current = bbLower;

    // Crosshair move handler - only set hovered data when hovering
    chart.subscribeCrosshairMove((param) => {
      if (!param.time || !param.seriesData) {
        // Mouse left chart - clear hover data (will show last candle)
        setHoveredOHLC(null);
        return;
      }

      const data = param.seriesData.get(candleSeries) as CandlestickData<Time> | undefined;
      if (data) {
        setHoveredOHLC({
          time: new Date((param.time as number) * 1000).toLocaleString(),
          open: data.open,
          high: data.high,
          low: data.low,
          close: data.close,
          volume: 0,
        });
      }
    });

    // Resize handler - check for valid dimensions to prevent warnings on unmount
    const handleResize = () => {
      if (chartContainerRef.current && chartRef.current) {
        const width = chartContainerRef.current.clientWidth;
        const height = chartContainerRef.current.clientHeight;
        // Only apply if dimensions are valid (prevents -1 or 0 warnings during unmount)
        if (width > 0 && height > 0) {
          chartRef.current.applyOptions({ width, height });
        }
      }
    };

    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(chartContainerRef.current);

    // Initial resize
    handleResize();

    return () => {
      resizeObserver.disconnect();
      markersPluginRef.current = null;
      chart.remove();
      chartRef.current = null;
      dataLoadedRef.current = false;
    };
  }, []); // Empty deps - only create chart once

  // Update visibility
  useEffect(() => {
    volumeSeriesRef.current?.applyOptions({ visible: showVolume });
  }, [showVolume]);

  useEffect(() => {
    ema20Ref.current?.applyOptions({ visible: showEMA });
    ema50Ref.current?.applyOptions({ visible: showEMA });
  }, [showEMA]);

  useEffect(() => {
    bbUpperRef.current?.applyOptions({ visible: showBB });
    bbMiddleRef.current?.applyOptions({ visible: showBB });
    bbLowerRef.current?.applyOptions({ visible: showBB });
  }, [showBB]);

  // Update chart data
  useEffect(() => {
    if (!candleSeriesRef.current || candleData.length === 0) return;

    candleSeriesRef.current.setData(candleData);

    if (volumeSeriesRef.current) {
      volumeSeriesRef.current.setData(volumeData);
    }

    // Only fit content on first load
    if (!dataLoadedRef.current) {
      chartRef.current?.timeScale().fitContent();
      dataLoadedRef.current = true;
    }
  }, [candleData, volumeData]);

  // Update EMA data
  useEffect(() => {
    if (ema20Ref.current && emaData.ema20.length > 0) {
      ema20Ref.current.setData(emaData.ema20);
    }
    if (ema50Ref.current && emaData.ema50.length > 0) {
      ema50Ref.current.setData(emaData.ema50);
    }
  }, [emaData]);

  // Update BB data
  useEffect(() => {
    if (bbUpperRef.current && bbData.upper.length > 0) {
      bbUpperRef.current.setData(bbData.upper);
    }
    if (bbMiddleRef.current && bbData.middle.length > 0) {
      bbMiddleRef.current.setData(bbData.middle);
    }
    if (bbLowerRef.current && bbData.lower.length > 0) {
      bbLowerRef.current.setData(bbData.lower);
    }
  }, [bbData]);

  // Update trade markers on chart
  useEffect(() => {
    if (markersPluginRef.current) {
      markersPluginRef.current.setMarkers(tradeMarkers);
    }
  }, [tradeMarkers]);

  const lastCandle = sortedCandles[sortedCandles.length - 1];
  const prevCandle = sortedCandles[sortedCandles.length - 2];
  const priceChange = lastCandle && prevCandle
    ? ((lastCandle.close - prevCandle.close) / prevCandle.close) * 100
    : 0;

  // OHLC to display - hover data or last candle (always visible)
  const displayOHLC = hoveredOHLC || (lastCandle ? {
    open: lastCandle.open,
    high: lastCandle.high,
    low: lastCandle.low,
    close: lastCandle.close,
  } : null);

  return (
    <div
      ref={containerRef}
      style={{
        height: typeof height === 'string' ? '100%' : `${height}px`,
        display: 'flex',
        flexDirection: 'column',
        background: '#131722',
        borderRadius: '4px',
        overflow: 'hidden',
        border: '1px solid #2a2e39',
      }}
    >
      {/* Chart Header - TradingView style */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '8px 12px',
        borderBottom: '1px solid #2a2e39',
        flexShrink: 0,
        background: '#1e222d',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'baseline', gap: '12px' }}>
              <span style={{ fontSize: '15px', fontWeight: 600, color: '#d1d4dc' }}>
                {symbol}
              </span>
              {lastCandle && (
                <>
                  <span style={{ fontSize: '18px', fontWeight: 600, color: '#d1d4dc' }}>
                    ${formatPrice(lastCandle.close)}
                  </span>
                  <span
                    style={{
                      fontSize: '13px',
                      fontWeight: 500,
                      color: priceChange >= 0 ? '#26a69a' : '#ef5350',
                    }}
                  >
                    {priceChange >= 0 ? '+' : ''}{priceChange.toFixed(2)}%
                  </span>
                </>
              )}
            </div>
            {displayOHLC && (
              <div style={{ fontSize: '11px', color: '#787b86', marginTop: '2px' }}>
                O: <span style={{ color: displayOHLC.close >= displayOHLC.open ? '#26a69a' : '#ef5350' }}>{formatPrice(displayOHLC.open)}</span>
                {' '}H: <span style={{ color: displayOHLC.close >= displayOHLC.open ? '#26a69a' : '#ef5350' }}>{formatPrice(displayOHLC.high)}</span>
                {' '}L: <span style={{ color: displayOHLC.close >= displayOHLC.open ? '#26a69a' : '#ef5350' }}>{formatPrice(displayOHLC.low)}</span>
                {' '}C: <span style={{ color: displayOHLC.close >= displayOHLC.open ? '#26a69a' : '#ef5350' }}>{formatPrice(displayOHLC.close)}</span>
              </div>
            )}
          </div>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          {/* Timeframe selector - TradingView style */}
          <div style={{ display: 'flex', gap: '1px' }}>
            {TIMEFRAMES.map((tf) => (
              <button
                key={tf}
                onClick={() => onTimeframeChange?.(tf)}
                style={{
                  padding: '4px 8px',
                  fontSize: '12px',
                  fontWeight: 400,
                  border: 'none',
                  borderRadius: '3px',
                  cursor: 'pointer',
                  transition: 'all 0.1s',
                  background: currentTimeframe === tf ? '#2962ff' : 'transparent',
                  color: currentTimeframe === tf ? '#fff' : '#787b86',
                }}
              >
                {tf}
              </button>
            ))}
          </div>

          <div style={{ width: '1px', height: '16px', background: '#363a45', margin: '0 8px' }} />

          {/* Indicator toggles - TradingView style */}
          <div style={{ display: 'flex', gap: '2px' }}>
            <button
              onClick={() => setShowVolume(!showVolume)}
              style={{
                padding: '4px 8px',
                fontSize: '11px',
                fontWeight: 400,
                border: 'none',
                borderRadius: '3px',
                cursor: 'pointer',
                background: showVolume ? 'rgba(41, 98, 255, 0.3)' : 'transparent',
                color: showVolume ? '#2962ff' : '#787b86',
              }}
            >
              Vol
            </button>
            <button
              onClick={() => setShowEMA(!showEMA)}
              style={{
                padding: '4px 8px',
                fontSize: '11px',
                fontWeight: 400,
                border: 'none',
                borderRadius: '3px',
                cursor: 'pointer',
                background: showEMA ? 'rgba(255, 152, 0, 0.2)' : 'transparent',
                color: showEMA ? '#ff9800' : '#787b86',
              }}
            >
              EMA
            </button>
            <button
              onClick={() => setShowBB(!showBB)}
              style={{
                padding: '4px 8px',
                fontSize: '11px',
                fontWeight: 400,
                border: 'none',
                borderRadius: '3px',
                cursor: 'pointer',
                background: showBB ? 'rgba(33, 150, 243, 0.2)' : 'transparent',
                color: showBB ? '#2196f3' : '#787b86',
              }}
            >
              BB
            </button>
          </div>
        </div>
      </div>

      {/* Indicator Legend - TradingView style */}
      {(showEMA || showBB) && (
        <div style={{
          display: 'flex',
          gap: '12px',
          padding: '4px 12px',
          fontSize: '11px',
          borderBottom: '1px solid #2a2e39',
          color: '#787b86',
          flexShrink: 0,
          background: '#131722',
        }}>
          {showEMA && (
            <>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <div style={{ width: '10px', height: '2px', background: '#ff9800', borderRadius: '1px' }} />
                <span>EMA 20</span>
                {emaData.ema20.length > 0 && (
                  <span style={{ color: '#ff9800' }}>
                    {formatPrice(emaData.ema20[emaData.ema20.length - 1]?.value || 0)}
                  </span>
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                <div style={{ width: '10px', height: '2px', background: '#9c27b0', borderRadius: '1px' }} />
                <span>EMA 50</span>
                {emaData.ema50.length > 0 && (
                  <span style={{ color: '#9c27b0' }}>
                    {formatPrice(emaData.ema50[emaData.ema50.length - 1]?.value || 0)}
                  </span>
                )}
              </div>
            </>
          )}
          {showBB && bbData.upper.length > 0 && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
              <div style={{ width: '10px', height: '2px', background: '#2196f3', borderRadius: '1px' }} />
              <span>BB(20,2)</span>
              <span style={{ color: '#2196f3' }}>
                {formatPrice(bbData.lower[bbData.lower.length - 1]?.value || 0)} - {formatPrice(bbData.upper[bbData.upper.length - 1]?.value || 0)}
              </span>
            </div>
          )}
        </div>
      )}

      {/* Chart Container */}
      <div
        ref={chartContainerRef}
        style={{
          flex: 1,
          minHeight: 0,
        }}
      />

      {/* Footer - TradingView style */}
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '4px 12px',
        fontSize: '10px',
        color: '#787b86',
        borderTop: '1px solid #2a2e39',
        flexShrink: 0,
        background: '#1e222d',
      }}>
        <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
          <svg width="12" height="12" viewBox="0 0 36 28" fill="#2962ff">
            <path d="M14 22H7V11H0V4h14v18zM28 22h-8l7.5-18h8L28 22z"/>
          </svg>
          TradingView
        </span>
        <span>{candles.length} candles</span>
      </div>
    </div>
  );
}
