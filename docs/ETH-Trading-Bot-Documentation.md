# ETH Trading Bot Documentation

A comprehensive Ethereum algorithmic trading system built with **React + Go + Binance API** featuring multiple trading strategies, dynamic regime detection, and robust risk management.

---

## Table of Contents

1. [Overview](#overview)
2. [Trading Strategies](#trading-strategies)
3. [Market Conditions & Indicators](#market-conditions--indicators)
4. [Strategy Switching Logic](#strategy-switching-logic)
5. [Risk Management System](#risk-management-system)
6. [Data Storage Architecture](#data-storage-architecture)
7. [System Architecture](#system-architecture)
8. [Binance API Integration](#binance-api-integration)
9. [Implementation Guide](#implementation-guide)
10. [Configuration Reference](#configuration-reference)

---

## Overview

Ethereum (ETH) exhibits complex price behavior. Empirical studies show:
- **Trend/momentum strategies** tend to dominate in strong directional moves
- **Mean-reversion** strategies work best in choppy/sideways markets
- **Breakout strategies** have mixed edge without proper filtering

This system implements a **multi-strategy approach** that dynamically switches between strategies based on market regime detection.

### Key Features

- Real-time market data via Binance WebSocket
- Multiple concurrent trading strategies
- Automatic regime detection and strategy switching
- Comprehensive risk management (automatic + manual)
- Backtesting capabilities
- React dashboard for monitoring and control

---

## Trading Strategies

### 1. Trend Following / Momentum

**Description**: Captures profits during sustained directional moves.

| Attribute | Details |
|-----------|---------|
| **Methods** | Moving-average crossovers, Opening Range Breakout (ORB), MACD, ADX trend signal |
| **Favorable Regime** | Sustained bull or bear trends (ADX > 25) |
| **Key Indicators** | MA crossovers, ADX/DMI, momentum RSI, volume spikes |
| **Performance** | ORB on ETH (4.5% threshold) netted ~$19k profit on $10k trades (190% return) with max drawdown ~$3k |

**Entry Conditions**:
```
- Fast SMA > Slow SMA (bullish) OR Fast SMA < Slow SMA (bearish)
- ADX > 25 (confirming trend strength)
- Volume above 20-period average
```

**Exit Conditions**:
```
- Opposite MA crossover
- ADX drops below 20
- Trailing stop hit
```

---

### 2. Mean Reversion / Countertrend

**Description**: Profits from price returning to mean after overextension.

| Attribute | Details |
|-----------|---------|
| **Methods** | RSI oversold/overbought bounces, false-breakout reentry, Bollinger Band center reversion |
| **Favorable Regime** | Choppy/sideways markets, brief pullbacks in trends |
| **Key Indicators** | RSI extremes (<30/>70), Bollinger midline touches, VWAP reversion |
| **Performance** | Mean-reversion "false breakout" system earned ~$71k from 75 trades (2016-2024 backtest) |

**Entry Conditions**:
```
- RSI < 30 (oversold) for long OR RSI > 70 (overbought) for short
- Price at or beyond Bollinger Band (2 std dev)
- ADX < 20 (confirming range-bound market)
```

**Exit Conditions**:
```
- RSI returns to 50 (neutral)
- Price reaches Bollinger middle band
- Time-based exit (max hold period)
```

---

### 3. Breakout

**Description**: Captures momentum from price breaking key levels.

| Attribute | Details |
|-----------|---------|
| **Methods** | Opening-range breakouts, volatility breakouts (Bollinger/Keltner), chart-pattern breakouts |
| **Favorable Regime** | Volatile markets, trend initiations |
| **Key Indicators** | Price > recent highs, Bollinger band expansion, ATR spikes, volume spikes |
| **Performance** | Requires ADX filter; without filter ~50% win rate with no statistical edge |

**Entry Conditions**:
```
- Price breaks above/below N-period high/low
- ATR expansion (current ATR > 1.5x average ATR)
- Volume spike (> 2x average volume)
- ADX > 20 (filter for trending potential)
```

**Exit Conditions**:
```
- Price fails to follow through (closes back inside range)
- ATR contraction
- Fixed profit target (1.5-2x ATR)
```

---

### 4. Volatility-Based

**Description**: Trades based on volatility expansion/contraction cycles.

| Attribute | Details |
|-----------|---------|
| **Methods** | Bollinger squeeze, volatility filters, straddle trades |
| **Favorable Regime** | Regime shifts, jumpy markets (high ATR) |
| **Key Indicators** | ATR thresholds, Bollinger/Donchian band width, VIX-style measures |
| **Performance** | Can catch swift moves but requires tight stops; success depends on proper filtering |

**Entry Conditions**:
```
- Bollinger Band width at 6-month low (squeeze)
- Followed by band expansion > 20%
- Direction determined by first candle close outside bands
```

**Exit Conditions**:
```
- Volatility contraction (bands narrowing)
- Opposite band touch
- Trailing stop based on ATR
```

---

### 5. Statistical Arbitrage

**Description**: Market-neutral strategy exploiting price relationships.

| Attribute | Details |
|-----------|---------|
| **Methods** | Pairs trading (ETH/BTC), ETH/futures spreads, cointegration models |
| **Favorable Regime** | All regimes (market-neutral) - best when ETH decouples from reference |
| **Key Indicators** | Price ratios, z-scores, cointegration residuals, cross-exchange spreads |
| **Performance** | ETH/BTC extremes can be traded as market-neutral play |

**Entry Conditions**:
```
- ETH/BTC z-score > 2 (short ETH, long BTC) OR < -2 (long ETH, short BTC)
- Cointegration test confirms relationship stability
- Spread deviation > 2 standard deviations from mean
```

**Exit Conditions**:
```
- Z-score returns to 0 (mean)
- Cointegration breaks down
- Max holding period reached
```

---

## Market Conditions & Indicators

### Trend Strength Indicators

| Indicator | Bullish | Bearish | Neutral/Range |
|-----------|---------|---------|---------------|
| **ADX** | > 25 with +DI > -DI | > 25 with -DI > +DI | < 20 |
| **MA Slope** | Fast SMA > Slow SMA | Fast SMA < Slow SMA | MAs flat/intertwined |
| **Price Position** | Above 200 SMA | Below 200 SMA | Oscillating around SMA |

### Momentum/Oversold Indicators

| Indicator | Overbought | Oversold | Neutral |
|-----------|------------|----------|---------|
| **RSI (14)** | > 70 | < 30 | 30-70 |
| **RSI (14) Extreme** | > 80 | < 20 | 20-80 |
| **Stochastic** | > 80 | < 20 | 20-80 |

### Volatility Indicators

| Indicator | High Volatility | Low Volatility | Calculation |
|-----------|-----------------|----------------|-------------|
| **ATR (14)** | > 1.5x 50-period avg | < 0.75x 50-period avg | 14-period Average True Range |
| **Bollinger Width** | > 4% of price | < 2% of price | (Upper - Lower) / Middle |
| **Keltner Width** | Expanding | Contracting | Based on ATR channels |

### Volume Indicators

| Signal | Condition | Interpretation |
|--------|-----------|----------------|
| **Volume Spike** | > 2x 20-period average | Validates breakouts |
| **Volume Decline** | < 0.5x average | Exhaustion signal |
| **On-chain Volume** | Stablecoin minting increase | Anticipates ETH demand |

### Relative Performance

| Metric | Calculation | Usage |
|--------|-------------|-------|
| **ETH/BTC Ratio** | ETH price / BTC price | Mean reversion when extreme |
| **ETH vs Index** | ETH return - Crypto index return | Relative strength |

---

## Strategy Switching Logic

### Regime Classification Rules

```
┌─────────────────────────────────────────────────────────────────┐
│                    REGIME DETECTION FLOWCHART                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐                                               │
│  │ ADX > 30?    │                                               │
│  └──────┬───────┘                                               │
│         │                                                        │
│    YES  │   NO                                                   │
│    ▼    ▼                                                        │
│  ┌─────────────┐  ┌──────────────┐                              │
│  │ RSI > 50?   │  │ ADX < 20?    │                              │
│  └──────┬──────┘  └──────┬───────┘                              │
│         │                │                                       │
│   YES   │  NO      YES   │   NO                                  │
│    ▼    ▼           ▼    ▼                                       │
│ ┌──────┐ ┌──────┐ ┌─────────┐ ┌────────────┐                    │
│ │BULL  │ │BEAR  │ │SIDEWAYS │ │TRANSITIONAL│                    │
│ │TREND │ │TREND │ │RANGE    │ │            │                    │
│ └──────┘ └──────┘ └─────────┘ └────────────┘                    │
│                                                                  │
│  Check ATR for volatility overlay:                              │
│  - ATR > 1.5x avg → Add VOLATILE flag                           │
│  - ATR < 0.75x avg → Add LOW_VOL flag                           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Strategy Priority by Regime

| Regime | Primary Strategy | Secondary Strategy | Disabled |
|--------|------------------|-------------------|----------|
| **BULL_TREND** | Trend Following (Long) | Breakout (Long) | Mean Reversion Short |
| **BEAR_TREND** | Trend Following (Short) | Breakout (Short) | Mean Reversion Long |
| **SIDEWAYS** | Mean Reversion | Stat Arb | Trend Following |
| **VOLATILE** | Volatility Breakout | Trend Following | Mean Reversion |
| **LOW_VOL** | Stat Arb | Mean Reversion | Breakout |

### Scoring System

Each strategy receives a score based on current conditions:

```go
type StrategyScore struct {
    Strategy     string
    Score        float64
    Conditions   []string
}

// Example scoring
MomentumScore = (RSI_momentum * 0.3) + (ADX_strength * 0.3) + (MA_alignment * 0.2) + (Volume_confirmation * 0.2)
ReversionScore = (RSI_extreme * 0.4) + (BB_deviation * 0.3) + (Range_confirmation * 0.3)
BreakoutScore = (Price_breakout * 0.3) + (Volume_spike * 0.3) + (ATR_expansion * 0.2) + (ADX_filter * 0.2)
```

### Safety Filters

1. **Reversal trades**: Only if volume or RSI divergence confirms
2. **Breakout trades**: Only if RSI isn't already extreme (avoid chasing)
3. **Trend trades**: Require volume confirmation
4. **All trades**: Check correlation with BTC for regime confirmation

---

## Risk Management System

The risk management system operates on two levels: **Automatic** (algorithmic) and **Manual** (user-controlled).

### Automatic Risk Management

#### 1. Position Sizing

**Volatility-Adjusted Sizing (Risk Parity)**

```go
type PositionSizer struct {
    AccountBalance    float64
    RiskPerTrade      float64  // Percentage of account (e.g., 0.01 = 1%)
    MaxPositionSize   float64  // Maximum position as % of account
}

// Calculate position size based on volatility
func (ps *PositionSizer) CalculateSize(atr float64, entryPrice float64, stopDistance float64) float64 {
    // Risk amount in dollars
    riskAmount := ps.AccountBalance * ps.RiskPerTrade

    // Position size = Risk Amount / Stop Distance
    positionSize := riskAmount / stopDistance

    // Apply volatility adjustment (smaller size in high volatility)
    volatilityFactor := ps.BaseATR / atr  // BaseATR is calibrated average
    adjustedSize := positionSize * volatilityFactor

    // Cap at maximum position size
    maxSize := ps.AccountBalance * ps.MaxPositionSize / entryPrice
    return min(adjustedSize, maxSize)
}
```

**Kelly Criterion (Optional)**

```go
func KellySize(winRate float64, avgWin float64, avgLoss float64) float64 {
    // Kelly % = W - [(1-W) / R]
    // W = win rate, R = win/loss ratio
    R := avgWin / avgLoss
    kelly := winRate - ((1 - winRate) / R)

    // Use fractional Kelly (25-50%) for safety
    return kelly * 0.25
}
```

#### 2. Stop Loss System

| Stop Type | Description | Calculation |
|-----------|-------------|-------------|
| **Fixed ATR Stop** | Based on ATR multiple | Entry ± (ATR × multiplier) |
| **Percentage Stop** | Fixed % from entry | Entry × (1 ± stop%) |
| **Volatility Stop** | Adapts to market vol | Entry ± (ATR × vol_factor) |
| **Time Stop** | Exit after N periods | Max holding period |
| **Trailing Stop** | Follows price | Highest/Lowest - (ATR × multiplier) |

```go
type StopLossConfig struct {
    Type           string   // "fixed", "atr", "percentage", "trailing"
    ATRMultiplier  float64  // For ATR-based stops (default: 2.0)
    Percentage     float64  // For percentage stops (default: 0.02 = 2%)
    TrailingATR    float64  // For trailing stops (default: 1.5)
    TimeLimit      int      // Max bars to hold position
}

type StopLoss struct {
    InitialStop    float64
    CurrentStop    float64
    TrailingActive bool
}

func (sl *StopLoss) Update(currentPrice float64, atr float64, isLong bool) {
    if sl.TrailingActive {
        if isLong {
            newStop := currentPrice - (atr * sl.Config.TrailingATR)
            sl.CurrentStop = max(sl.CurrentStop, newStop)
        } else {
            newStop := currentPrice + (atr * sl.Config.TrailingATR)
            sl.CurrentStop = min(sl.CurrentStop, newStop)
        }
    }
}
```

#### 3. Take Profit System

| Target Type | Description | Calculation |
|-------------|-------------|-------------|
| **Fixed R:R** | Risk/Reward ratio | Stop distance × R multiple |
| **ATR Target** | Based on ATR | Entry ± (ATR × multiplier) |
| **Resistance/Support** | Technical levels | Nearest S/R level |
| **Partial Exits** | Scale out | 50% at 1R, 50% at 2R |

```go
type TakeProfitConfig struct {
    Targets []ProfitTarget
}

type ProfitTarget struct {
    Percentage   float64  // % of position to close
    ATRMultiple  float64  // Distance in ATR units
    RiskMultiple float64  // Distance as multiple of risk
}

// Example: Scale out at 1R, 2R, 3R
var DefaultTargets = []ProfitTarget{
    {Percentage: 0.33, RiskMultiple: 1.0},
    {Percentage: 0.33, RiskMultiple: 2.0},
    {Percentage: 0.34, RiskMultiple: 3.0},
}
```

#### 4. Drawdown Protection

```go
type DrawdownProtection struct {
    MaxDrawdown        float64  // Maximum allowed drawdown (e.g., 0.10 = 10%)
    DailyDrawdownLimit float64  // Max daily loss (e.g., 0.03 = 3%)
    WeeklyDrawdownLimit float64 // Max weekly loss (e.g., 0.07 = 7%)

    PeakEquity         float64
    CurrentEquity      float64
    DailyStartEquity   float64
    WeeklyStartEquity  float64
}

type DrawdownAction string
const (
    ActionNone         DrawdownAction = "none"
    ActionReduceSize   DrawdownAction = "reduce_size"
    ActionPauseTrading DrawdownAction = "pause_trading"
    ActionCloseAll     DrawdownAction = "close_all"
)

func (dp *DrawdownProtection) CheckDrawdown() DrawdownAction {
    currentDrawdown := (dp.PeakEquity - dp.CurrentEquity) / dp.PeakEquity
    dailyDrawdown := (dp.DailyStartEquity - dp.CurrentEquity) / dp.DailyStartEquity
    weeklyDrawdown := (dp.WeeklyStartEquity - dp.CurrentEquity) / dp.WeeklyStartEquity

    // Critical drawdown - close all positions
    if currentDrawdown >= dp.MaxDrawdown {
        return ActionCloseAll
    }

    // Daily limit hit - pause for the day
    if dailyDrawdown >= dp.DailyDrawdownLimit {
        return ActionPauseTrading
    }

    // Weekly limit approaching - reduce position sizes
    if weeklyDrawdown >= dp.WeeklyDrawdownLimit * 0.75 {
        return ActionReduceSize
    }

    return ActionNone
}
```

#### 5. Correlation & Exposure Management

```go
type ExposureManager struct {
    MaxLongExposure     float64  // Max % of account in long positions
    MaxShortExposure    float64  // Max % of account in short positions
    MaxTotalExposure    float64  // Max total exposure (long + short)
    MaxCorrelatedRisk   float64  // Max risk in correlated assets
}

func (em *ExposureManager) CanOpenPosition(
    side string,
    size float64,
    currentLong float64,
    currentShort float64,
) bool {
    if side == "long" {
        return (currentLong + size) <= em.MaxLongExposure
    }
    return (currentShort + size) <= em.MaxShortExposure
}
```

#### 6. Circuit Breakers

```go
type CircuitBreaker struct {
    MaxConsecutiveLosses  int      // Pause after N consecutive losses
    MaxLossesPerHour      int      // Max losses per hour
    MaxLossesPerDay       int      // Max losses per day
    VolatilityThreshold   float64  // Pause if volatility exceeds threshold

    ConsecutiveLosses     int
    HourlyLosses          int
    DailyLosses           int
    LastResetHour         time.Time
    LastResetDay          time.Time
}

type CircuitBreakerState string
const (
    StateActive     CircuitBreakerState = "active"
    StatePaused     CircuitBreakerState = "paused"
    StateEmergency  CircuitBreakerState = "emergency"
)

func (cb *CircuitBreaker) RecordTrade(isWin bool) CircuitBreakerState {
    cb.resetCountersIfNeeded()

    if !isWin {
        cb.ConsecutiveLosses++
        cb.HourlyLosses++
        cb.DailyLosses++
    } else {
        cb.ConsecutiveLosses = 0
    }

    if cb.ConsecutiveLosses >= cb.MaxConsecutiveLosses {
        return StatePaused
    }
    if cb.HourlyLosses >= cb.MaxLossesPerHour {
        return StatePaused
    }
    if cb.DailyLosses >= cb.MaxLossesPerDay {
        return StateEmergency
    }

    return StateActive
}
```

---

### Manual Risk Management

#### 1. User-Configurable Parameters

```go
type ManualRiskConfig struct {
    // Position Limits
    MaxPositionSizeUSD    float64  `json:"max_position_size_usd"`
    MaxPositionSizeETH    float64  `json:"max_position_size_eth"`
    MaxOpenPositions      int      `json:"max_open_positions"`

    // Risk Per Trade
    RiskPerTradePercent   float64  `json:"risk_per_trade_percent"`
    MaxRiskPerTradeUSD    float64  `json:"max_risk_per_trade_usd"`

    // Stop Loss Settings
    UseStopLoss           bool     `json:"use_stop_loss"`
    StopLossType          string   `json:"stop_loss_type"`
    StopLossPercent       float64  `json:"stop_loss_percent"`
    StopLossATRMultiple   float64  `json:"stop_loss_atr_multiple"`

    // Take Profit Settings
    UseTakeProfit         bool     `json:"use_take_profit"`
    TakeProfitPercent     float64  `json:"take_profit_percent"`
    TakeProfitATRMultiple float64  `json:"take_profit_atr_multiple"`
    UseScaleOut           bool     `json:"use_scale_out"`

    // Drawdown Limits
    MaxDailyLossPercent   float64  `json:"max_daily_loss_percent"`
    MaxWeeklyLossPercent  float64  `json:"max_weekly_loss_percent"`
    MaxTotalDrawdown      float64  `json:"max_total_drawdown"`

    // Trading Hours
    EnableTradingHours    bool     `json:"enable_trading_hours"`
    TradingStartHourUTC   int      `json:"trading_start_hour_utc"`
    TradingEndHourUTC     int      `json:"trading_end_hour_utc"`
    TradingDays           []int    `json:"trading_days"`  // 0=Sunday, 6=Saturday

    // Strategy Toggles
    EnabledStrategies     []string `json:"enabled_strategies"`

    // Emergency Controls
    PanicSellEnabled      bool     `json:"panic_sell_enabled"`
    PanicSellThreshold    float64  `json:"panic_sell_threshold"`  // % drop to trigger
}
```

#### 2. Manual Override Controls

```go
type ManualControls struct {
    // Instant Actions
    PauseAllTrading       bool     `json:"pause_all_trading"`
    CloseAllPositions     bool     `json:"close_all_positions"`
    CancelAllOrders       bool     `json:"cancel_all_orders"`

    // Position-Specific
    ClosePosition         string   `json:"close_position"`  // Position ID
    ModifyStopLoss        float64  `json:"modify_stop_loss"`
    ModifyTakeProfit      float64  `json:"modify_take_profit"`

    // Strategy-Specific
    DisableStrategy       string   `json:"disable_strategy"`
    EnableStrategy        string   `json:"enable_strategy"`

    // Risk Adjustments
    ReduceAllPositions    float64  `json:"reduce_all_positions"`  // % to reduce
    LockInProfits         float64  `json:"lock_in_profits"`  // Move stops to breakeven
}
```

#### 3. Alert & Notification System

```go
type AlertConfig struct {
    // Alert Channels
    EnableEmail           bool     `json:"enable_email"`
    EnableSMS             bool     `json:"enable_sms"`
    EnableTelegram        bool     `json:"enable_telegram"`
    EnableWebhook         bool     `json:"enable_webhook"`

    // Alert Thresholds
    AlertOnDrawdown       float64  `json:"alert_on_drawdown"`
    AlertOnProfit         float64  `json:"alert_on_profit"`
    AlertOnVolatility     float64  `json:"alert_on_volatility"`
    AlertOnSignal         bool     `json:"alert_on_signal"`
    AlertOnExecution      bool     `json:"alert_on_execution"`
    AlertOnError          bool     `json:"alert_on_error"`

    // Contact Details
    EmailAddress          string   `json:"email_address"`
    PhoneNumber           string   `json:"phone_number"`
    TelegramChatID        string   `json:"telegram_chat_id"`
    WebhookURL            string   `json:"webhook_url"`
}

type Alert struct {
    Type      string    `json:"type"`
    Severity  string    `json:"severity"`  // info, warning, critical
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}
```

#### 4. Manual Position Management UI Features

| Feature | Description | API Endpoint |
|---------|-------------|--------------|
| **One-Click Close** | Close position immediately | `POST /api/positions/{id}/close` |
| **Modify Stop** | Adjust stop loss level | `PUT /api/positions/{id}/stop` |
| **Modify Target** | Adjust take profit | `PUT /api/positions/{id}/target` |
| **Partial Close** | Close % of position | `POST /api/positions/{id}/partial-close` |
| **Add to Position** | Increase position size | `POST /api/positions/{id}/add` |
| **Move to Breakeven** | Set stop at entry | `POST /api/positions/{id}/breakeven` |
| **Trail Stop** | Enable trailing stop | `PUT /api/positions/{id}/trail` |

#### 5. Risk Dashboard Metrics

```go
type RiskDashboard struct {
    // Current State
    TotalEquity           float64  `json:"total_equity"`
    AvailableBalance      float64  `json:"available_balance"`
    TotalExposure         float64  `json:"total_exposure"`
    ExposurePercent       float64  `json:"exposure_percent"`

    // P&L Metrics
    DailyPnL              float64  `json:"daily_pnl"`
    DailyPnLPercent       float64  `json:"daily_pnl_percent"`
    WeeklyPnL             float64  `json:"weekly_pnl"`
    MonthlyPnL            float64  `json:"monthly_pnl"`

    // Drawdown Metrics
    CurrentDrawdown       float64  `json:"current_drawdown"`
    MaxDrawdown           float64  `json:"max_drawdown"`
    DaysInDrawdown        int      `json:"days_in_drawdown"`

    // Risk Metrics
    OpenRisk              float64  `json:"open_risk"`  // $ at risk in open positions
    RiskPercent           float64  `json:"risk_percent"`
    PositionsCount        int      `json:"positions_count"`
    WinRate               float64  `json:"win_rate"`
    ProfitFactor          float64  `json:"profit_factor"`
    SharpeRatio           float64  `json:"sharpe_ratio"`

    // Limits Status
    DailyLossRemaining    float64  `json:"daily_loss_remaining"`
    WeeklyLossRemaining   float64  `json:"weekly_loss_remaining"`
    CircuitBreakerStatus  string   `json:"circuit_breaker_status"`

    // Position Details
    Positions             []PositionRisk `json:"positions"`
}

type PositionRisk struct {
    ID                string   `json:"id"`
    Symbol            string   `json:"symbol"`
    Side              string   `json:"side"`
    Size              float64  `json:"size"`
    EntryPrice        float64  `json:"entry_price"`
    CurrentPrice      float64  `json:"current_price"`
    UnrealizedPnL     float64  `json:"unrealized_pnl"`
    StopLoss          float64  `json:"stop_loss"`
    TakeProfit        float64  `json:"take_profit"`
    RiskAmount        float64  `json:"risk_amount"`
    RiskRewardRatio   float64  `json:"risk_reward_ratio"`
    Strategy          string   `json:"strategy"`
    Duration          string   `json:"duration"`
}
```

---

## Data Storage Architecture

The system uses a **hybrid storage approach**:
- **SQLite Database**: Persistent storage for historical candles, trades, positions, and configuration
- **In-Memory Circular Queue**: Fast access to recent candles needed for indicator calculations

### Why This Architecture?

| Requirement | SQLite | Circular Queue |
|-------------|--------|----------------|
| **Persistence** | Survives restarts | Lost on restart |
| **Speed** | Disk I/O (~1-10ms) | Memory access (~1μs) |
| **Capacity** | Unlimited (disk) | Fixed size (RAM) |
| **Use Case** | Historical data, backtesting | Real-time indicator calculation |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DATA FLOW ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌──────────────┐                                                          │
│   │   Binance    │                                                          │
│   │  WebSocket   │                                                          │
│   └──────┬───────┘                                                          │
│          │ Real-time candles                                                 │
│          ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────────┐      │
│   │                      DATA INGESTION SERVICE                       │      │
│   └──────────────────────────┬───────────────────────────────────────┘      │
│                              │                                               │
│              ┌───────────────┴───────────────┐                              │
│              │                               │                               │
│              ▼                               ▼                               │
│   ┌──────────────────┐            ┌──────────────────┐                      │
│   │  CIRCULAR QUEUE  │            │   SQLite DB      │                      │
│   │   (In-Memory)    │            │  (Persistent)    │                      │
│   │                  │            │                  │                      │
│   │ • Last N candles │            │ • candles table  │                      │
│   │ • O(1) push/pop  │            │ • trades table   │                      │
│   │ • Fixed size     │            │ • positions table│                      │
│   │ • Thread-safe    │            │ • orders table   │                      │
│   └────────┬─────────┘            │ • config table   │                      │
│            │                      └──────────────────┘                      │
│            │ Fast read                     │                                 │
│            ▼                               │ Batch write (async)            │
│   ┌──────────────────┐                     │                                 │
│   │    INDICATOR     │◄────────────────────┘                                │
│   │     SERVICE      │     Historical queries                               │
│   └──────────────────┘     (backtesting)                                    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

### Circular Queue Implementation

The circular queue (ring buffer) stores the most recent N candles in memory for fast indicator calculations.

#### Structure

```go
package storage

import (
    "sync"
    "time"
)

// Candle represents OHLCV data
type Candle struct {
    OpenTime   time.Time `json:"open_time" db:"open_time"`
    CloseTime  time.Time `json:"close_time" db:"close_time"`
    Symbol     string    `json:"symbol" db:"symbol"`
    Timeframe  string    `json:"timeframe" db:"timeframe"`
    Open       float64   `json:"open" db:"open"`
    High       float64   `json:"high" db:"high"`
    Low        float64   `json:"low" db:"low"`
    Close      float64   `json:"close" db:"close"`
    Volume     float64   `json:"volume" db:"volume"`
    Trades     int       `json:"trades" db:"trades"`
    IsClosed   bool      `json:"is_closed" db:"is_closed"`
}

// CandleQueue is a thread-safe circular queue for candles
type CandleQueue struct {
    buffer   []Candle
    capacity int
    head     int      // Points to oldest element
    tail     int      // Points to next write position
    size     int      // Current number of elements
    mu       sync.RWMutex
}

// NewCandleQueue creates a new circular queue with given capacity
func NewCandleQueue(capacity int) *CandleQueue {
    return &CandleQueue{
        buffer:   make([]Candle, capacity),
        capacity: capacity,
        head:     0,
        tail:     0,
        size:     0,
    }
}

// Push adds a candle to the queue (overwrites oldest if full)
func (q *CandleQueue) Push(candle Candle) {
    q.mu.Lock()
    defer q.mu.Unlock()

    q.buffer[q.tail] = candle
    q.tail = (q.tail + 1) % q.capacity

    if q.size < q.capacity {
        q.size++
    } else {
        // Queue is full, move head forward (discard oldest)
        q.head = (q.head + 1) % q.capacity
    }
}

// GetAll returns all candles in order (oldest to newest)
func (q *CandleQueue) GetAll() []Candle {
    q.mu.RLock()
    defer q.mu.RUnlock()

    result := make([]Candle, q.size)
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        result[i] = q.buffer[idx]
    }
    return result
}

// GetLast returns the last N candles (newest first)
func (q *CandleQueue) GetLast(n int) []Candle {
    q.mu.RLock()
    defer q.mu.RUnlock()

    if n > q.size {
        n = q.size
    }

    result := make([]Candle, n)
    for i := 0; i < n; i++ {
        // Start from tail-1 (most recent) and go backwards
        idx := (q.tail - 1 - i + q.capacity) % q.capacity
        result[i] = q.buffer[idx]
    }
    return result
}

// GetLatest returns the most recent candle
func (q *CandleQueue) GetLatest() (Candle, bool) {
    q.mu.RLock()
    defer q.mu.RUnlock()

    if q.size == 0 {
        return Candle{}, false
    }

    idx := (q.tail - 1 + q.capacity) % q.capacity
    return q.buffer[idx], true
}

// UpdateLatest updates the most recent candle (for live candle updates)
func (q *CandleQueue) UpdateLatest(candle Candle) bool {
    q.mu.Lock()
    defer q.mu.Unlock()

    if q.size == 0 {
        return false
    }

    idx := (q.tail - 1 + q.capacity) % q.capacity
    q.buffer[idx] = candle
    return true
}

// Size returns the current number of elements
func (q *CandleQueue) Size() int {
    q.mu.RLock()
    defer q.mu.RUnlock()
    return q.size
}

// IsFull returns true if the queue is at capacity
func (q *CandleQueue) IsFull() bool {
    q.mu.RLock()
    defer q.mu.RUnlock()
    return q.size == q.capacity
}

// Clear removes all elements from the queue
func (q *CandleQueue) Clear() {
    q.mu.Lock()
    defer q.mu.Unlock()

    q.head = 0
    q.tail = 0
    q.size = 0
}

// GetWindow returns candles within a time window
func (q *CandleQueue) GetWindow(from, to time.Time) []Candle {
    q.mu.RLock()
    defer q.mu.RUnlock()

    var result []Candle
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        candle := q.buffer[idx]
        if candle.OpenTime.After(from) && candle.OpenTime.Before(to) {
            result = append(result, candle)
        }
    }
    return result
}

// GetCloses returns just the close prices (useful for indicators)
func (q *CandleQueue) GetCloses() []float64 {
    q.mu.RLock()
    defer q.mu.RUnlock()

    closes := make([]float64, q.size)
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        closes[i] = q.buffer[idx].Close
    }
    return closes
}

// GetHighs returns just the high prices
func (q *CandleQueue) GetHighs() []float64 {
    q.mu.RLock()
    defer q.mu.RUnlock()

    highs := make([]float64, q.size)
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        highs[i] = q.buffer[idx].High
    }
    return highs
}

// GetLows returns just the low prices
func (q *CandleQueue) GetLows() []float64 {
    q.mu.RLock()
    defer q.mu.RUnlock()

    lows := make([]float64, q.size)
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        lows[i] = q.buffer[idx].Low
    }
    return lows
}

// GetVolumes returns just the volumes
func (q *CandleQueue) GetVolumes() []float64 {
    q.mu.RLock()
    defer q.mu.RUnlock()

    volumes := make([]float64, q.size)
    for i := 0; i < q.size; i++ {
        idx := (q.head + i) % q.capacity
        volumes[i] = q.buffer[idx].Volume
    }
    return volumes
}
```

#### Multi-Timeframe Queue Manager

```go
// QueueManager manages multiple candle queues for different timeframes
type QueueManager struct {
    queues map[string]*CandleQueue  // key: "ETHUSDT_1m", "ETHUSDT_5m", etc.
    mu     sync.RWMutex
}

// NewQueueManager creates a new queue manager
func NewQueueManager() *QueueManager {
    return &QueueManager{
        queues: make(map[string]*CandleQueue),
    }
}

// GetOrCreate returns existing queue or creates new one
func (qm *QueueManager) GetOrCreate(symbol, timeframe string, capacity int) *CandleQueue {
    key := symbol + "_" + timeframe

    qm.mu.Lock()
    defer qm.mu.Unlock()

    if queue, exists := qm.queues[key]; exists {
        return queue
    }

    queue := NewCandleQueue(capacity)
    qm.queues[key] = queue
    return queue
}

// Get returns a queue if it exists
func (qm *QueueManager) Get(symbol, timeframe string) (*CandleQueue, bool) {
    key := symbol + "_" + timeframe

    qm.mu.RLock()
    defer qm.mu.RUnlock()

    queue, exists := qm.queues[key]
    return queue, exists
}

// Capacity recommendations by timeframe
var DefaultCapacities = map[string]int{
    "1m":  500,   // ~8 hours of 1m candles
    "5m":  300,   // ~25 hours of 5m candles
    "15m": 200,   // ~50 hours of 15m candles
    "1h":  200,   // ~8 days of 1h candles
    "4h":  150,   // ~25 days of 4h candles
    "1d":  100,   // ~100 days of daily candles
}
```

---

### SQLite Database Schema

#### Database Initialization

```go
package storage

import (
    "database/sql"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
    db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
    db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL")
    if err != nil {
        return nil, err
    }

    // Enable WAL mode for better concurrent read/write performance
    _, err = db.Exec("PRAGMA journal_mode=WAL")
    if err != nil {
        return nil, err
    }

    sqliteDB := &SQLiteDB{db: db}
    if err := sqliteDB.migrate(); err != nil {
        return nil, err
    }

    return sqliteDB, nil
}

func (s *SQLiteDB) migrate() error {
    migrations := []string{
        // Candles table
        `CREATE TABLE IF NOT EXISTS candles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            symbol TEXT NOT NULL,
            timeframe TEXT NOT NULL,
            open_time DATETIME NOT NULL,
            close_time DATETIME NOT NULL,
            open REAL NOT NULL,
            high REAL NOT NULL,
            low REAL NOT NULL,
            close REAL NOT NULL,
            volume REAL NOT NULL,
            trades INTEGER DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(symbol, timeframe, open_time)
        )`,

        // Index for fast candle queries
        `CREATE INDEX IF NOT EXISTS idx_candles_symbol_timeframe_time
         ON candles(symbol, timeframe, open_time DESC)`,

        // Trades/Executions table
        `CREATE TABLE IF NOT EXISTS trades (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            order_id TEXT UNIQUE NOT NULL,
            symbol TEXT NOT NULL,
            side TEXT NOT NULL,
            type TEXT NOT NULL,
            quantity REAL NOT NULL,
            price REAL NOT NULL,
            commission REAL DEFAULT 0,
            commission_asset TEXT,
            executed_at DATETIME NOT NULL,
            strategy TEXT,
            signal_strength REAL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

        // Positions table
        `CREATE TABLE IF NOT EXISTS positions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            symbol TEXT NOT NULL,
            side TEXT NOT NULL,
            entry_price REAL NOT NULL,
            quantity REAL NOT NULL,
            current_price REAL,
            unrealized_pnl REAL DEFAULT 0,
            realized_pnl REAL DEFAULT 0,
            stop_loss REAL,
            take_profit REAL,
            strategy TEXT,
            status TEXT DEFAULT 'open',
            opened_at DATETIME NOT NULL,
            closed_at DATETIME,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

        // Orders table
        `CREATE TABLE IF NOT EXISTS orders (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            order_id TEXT UNIQUE NOT NULL,
            client_order_id TEXT,
            symbol TEXT NOT NULL,
            side TEXT NOT NULL,
            type TEXT NOT NULL,
            quantity REAL NOT NULL,
            price REAL,
            stop_price REAL,
            status TEXT DEFAULT 'pending',
            filled_quantity REAL DEFAULT 0,
            avg_fill_price REAL,
            strategy TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

        // Account snapshots for equity tracking
        `CREATE TABLE IF NOT EXISTS account_snapshots (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            total_equity REAL NOT NULL,
            available_balance REAL NOT NULL,
            unrealized_pnl REAL DEFAULT 0,
            daily_pnl REAL DEFAULT 0,
            open_positions INTEGER DEFAULT 0,
            snapshot_time DATETIME NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

        // Strategy performance tracking
        `CREATE TABLE IF NOT EXISTS strategy_performance (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            strategy TEXT NOT NULL,
            date DATE NOT NULL,
            trades INTEGER DEFAULT 0,
            wins INTEGER DEFAULT 0,
            losses INTEGER DEFAULT 0,
            gross_profit REAL DEFAULT 0,
            gross_loss REAL DEFAULT 0,
            net_pnl REAL DEFAULT 0,
            max_drawdown REAL DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(strategy, date)
        )`,

        // Configuration table
        `CREATE TABLE IF NOT EXISTS config (
            key TEXT PRIMARY KEY,
            value TEXT NOT NULL,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,

        // Alerts/Notifications log
        `CREATE TABLE IF NOT EXISTS alerts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            type TEXT NOT NULL,
            severity TEXT NOT NULL,
            message TEXT NOT NULL,
            data TEXT,
            acknowledged BOOLEAN DEFAULT FALSE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
    }

    for _, migration := range migrations {
        if _, err := s.db.Exec(migration); err != nil {
            return err
        }
    }

    return nil
}
```

#### Candle Repository

```go
// CandleRepository handles candle persistence
type CandleRepository struct {
    db *SQLiteDB
}

func NewCandleRepository(db *SQLiteDB) *CandleRepository {
    return &CandleRepository{db: db}
}

// Insert adds a new candle (upsert)
func (r *CandleRepository) Insert(candle Candle) error {
    query := `
        INSERT INTO candles (symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(symbol, timeframe, open_time) DO UPDATE SET
            high = MAX(excluded.high, candles.high),
            low = MIN(excluded.low, candles.low),
            close = excluded.close,
            volume = excluded.volume,
            trades = excluded.trades
    `
    _, err := r.db.db.Exec(query,
        candle.Symbol, candle.Timeframe, candle.OpenTime, candle.CloseTime,
        candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Trades,
    )
    return err
}

// InsertBatch inserts multiple candles efficiently
func (r *CandleRepository) InsertBatch(candles []Candle) error {
    tx, err := r.db.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO candles (symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(symbol, timeframe, open_time) DO UPDATE SET
            high = MAX(excluded.high, candles.high),
            low = MIN(excluded.low, candles.low),
            close = excluded.close,
            volume = excluded.volume,
            trades = excluded.trades
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, candle := range candles {
        _, err := stmt.Exec(
            candle.Symbol, candle.Timeframe, candle.OpenTime, candle.CloseTime,
            candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Trades,
        )
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

// GetRange retrieves candles within a time range
func (r *CandleRepository) GetRange(symbol, timeframe string, from, to time.Time) ([]Candle, error) {
    query := `
        SELECT symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades
        FROM candles
        WHERE symbol = ? AND timeframe = ? AND open_time >= ? AND open_time <= ?
        ORDER BY open_time ASC
    `
    rows, err := r.db.db.Query(query, symbol, timeframe, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var candles []Candle
    for rows.Next() {
        var c Candle
        err := rows.Scan(&c.Symbol, &c.Timeframe, &c.OpenTime, &c.CloseTime,
            &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.Trades)
        if err != nil {
            return nil, err
        }
        candles = append(candles, c)
    }
    return candles, nil
}

// GetLast retrieves the last N candles
func (r *CandleRepository) GetLast(symbol, timeframe string, limit int) ([]Candle, error) {
    query := `
        SELECT symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades
        FROM candles
        WHERE symbol = ? AND timeframe = ?
        ORDER BY open_time DESC
        LIMIT ?
    `
    rows, err := r.db.db.Query(query, symbol, timeframe, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var candles []Candle
    for rows.Next() {
        var c Candle
        err := rows.Scan(&c.Symbol, &c.Timeframe, &c.OpenTime, &c.CloseTime,
            &c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.Trades)
        if err != nil {
            return nil, err
        }
        candles = append(candles, c)
    }

    // Reverse to get oldest first
    for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
        candles[i], candles[j] = candles[j], candles[i]
    }
    return candles, nil
}

// Cleanup removes old candles to prevent unbounded growth
func (r *CandleRepository) Cleanup(retentionDays int) error {
    cutoff := time.Now().AddDate(0, 0, -retentionDays)
    _, err := r.db.db.Exec(`DELETE FROM candles WHERE open_time < ?`, cutoff)
    return err
}
```

---

### Data Service (Combining Queue + SQLite)

```go
package storage

import (
    "context"
    "sync"
    "time"
)

// DataService coordinates between in-memory queue and SQLite
type DataService struct {
    queueManager    *QueueManager
    candleRepo      *CandleRepository
    persistInterval time.Duration
    pendingCandles  []Candle
    pendingMu       sync.Mutex
}

func NewDataService(db *SQLiteDB, persistInterval time.Duration) *DataService {
    return &DataService{
        queueManager:    NewQueueManager(),
        candleRepo:      NewCandleRepository(db),
        persistInterval: persistInterval,
        pendingCandles:  make([]Candle, 0, 100),
    }
}

// AddCandle adds a candle to both in-memory queue and persistence queue
func (ds *DataService) AddCandle(candle Candle) {
    // Add to in-memory queue for fast access
    capacity := DefaultCapacities[candle.Timeframe]
    if capacity == 0 {
        capacity = 200 // default
    }
    queue := ds.queueManager.GetOrCreate(candle.Symbol, candle.Timeframe, capacity)

    // If candle is not closed, update the latest candle
    if !candle.IsClosed {
        if latest, ok := queue.GetLatest(); ok && latest.OpenTime.Equal(candle.OpenTime) {
            queue.UpdateLatest(candle)
            return
        }
    }

    queue.Push(candle)

    // Queue for async persistence (only closed candles)
    if candle.IsClosed {
        ds.pendingMu.Lock()
        ds.pendingCandles = append(ds.pendingCandles, candle)
        ds.pendingMu.Unlock()
    }
}

// GetCandles returns candles from in-memory queue
func (ds *DataService) GetCandles(symbol, timeframe string) []Candle {
    queue, exists := ds.queueManager.Get(symbol, timeframe)
    if !exists {
        return nil
    }
    return queue.GetAll()
}

// GetCandlesForIndicators returns price arrays for indicator calculation
func (ds *DataService) GetCandlesForIndicators(symbol, timeframe string) (closes, highs, lows, volumes []float64) {
    queue, exists := ds.queueManager.Get(symbol, timeframe)
    if !exists {
        return nil, nil, nil, nil
    }
    return queue.GetCloses(), queue.GetHighs(), queue.GetLows(), queue.GetVolumes()
}

// LoadHistoricalCandles loads candles from SQLite into memory queue on startup
func (ds *DataService) LoadHistoricalCandles(symbol, timeframe string) error {
    capacity := DefaultCapacities[timeframe]
    if capacity == 0 {
        capacity = 200
    }

    // Load from SQLite
    candles, err := ds.candleRepo.GetLast(symbol, timeframe, capacity)
    if err != nil {
        return err
    }

    // Create queue and populate
    queue := ds.queueManager.GetOrCreate(symbol, timeframe, capacity)
    for _, candle := range candles {
        queue.Push(candle)
    }

    return nil
}

// StartPersistence starts the background persistence goroutine
func (ds *DataService) StartPersistence(ctx context.Context) {
    ticker := time.NewTicker(ds.persistInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // Final flush before exit
            ds.flushPendingCandles()
            return
        case <-ticker.C:
            ds.flushPendingCandles()
        }
    }
}

func (ds *DataService) flushPendingCandles() {
    ds.pendingMu.Lock()
    if len(ds.pendingCandles) == 0 {
        ds.pendingMu.Unlock()
        return
    }
    candles := ds.pendingCandles
    ds.pendingCandles = make([]Candle, 0, 100)
    ds.pendingMu.Unlock()

    // Batch insert to SQLite
    if err := ds.candleRepo.InsertBatch(candles); err != nil {
        // Log error, candles will be lost but queue still has them
        // Consider retry logic here
    }
}
```

---

### Database Configuration

```yaml
# configs/config.yaml

database:
  # SQLite database file path
  path: "./data/trading.db"

  # WAL mode for better concurrent performance
  journal_mode: "WAL"

  # Synchronous mode (NORMAL is good balance of safety/speed)
  synchronous: "NORMAL"

  # How often to flush pending candles to disk (seconds)
  persist_interval: 10

  # Retention period for historical candles (days)
  candle_retention_days: 365

  # Cleanup schedule (cron expression)
  cleanup_schedule: "0 0 * * *"  # Daily at midnight

# In-memory queue settings
candle_queues:
  # Capacity per timeframe
  capacities:
    "1m": 500    # ~8 hours
    "5m": 300    # ~25 hours
    "15m": 200   # ~50 hours
    "1h": 200    # ~8 days
    "4h": 150    # ~25 days
    "1d": 100    # ~100 days

  # Default capacity if timeframe not specified
  default_capacity: 200
```

---

### Performance Considerations

| Operation | In-Memory Queue | SQLite |
|-----------|-----------------|--------|
| **Read latest candle** | O(1) ~1μs | O(1) ~1-5ms |
| **Read last N candles** | O(N) ~10μs | O(N) ~5-20ms |
| **Write single candle** | O(1) ~1μs | O(1) ~1-10ms |
| **Batch write 100 candles** | N/A | ~20-50ms |
| **Memory usage (500 candles)** | ~50KB | N/A (disk) |

### Best Practices

1. **Use queue for real-time operations**: All indicator calculations should read from the in-memory queue
2. **Batch writes to SQLite**: Accumulate candles and write in batches to reduce I/O
3. **WAL mode**: Enables concurrent reads while writing
4. **Index properly**: Composite index on (symbol, timeframe, open_time) for fast queries
5. **Periodic cleanup**: Remove old data to prevent unbounded growth
6. **Load on startup**: Populate queues from SQLite when the bot starts

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SYSTEM ARCHITECTURE                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         REACT FRONTEND                               │    │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────────────┐   │    │
│  │  │ Dashboard │ │ Strategy  │ │   Risk    │ │    Manual         │   │    │
│  │  │  Charts   │ │  Config   │ │ Dashboard │ │    Controls       │   │    │
│  │  └───────────┘ └───────────┘ └───────────┘ └───────────────────┘   │    │
│  └─────────────────────────────────┬───────────────────────────────────┘    │
│                                    │ WebSocket / REST                        │
│  ┌─────────────────────────────────▼───────────────────────────────────┐    │
│  │                          GO BACKEND                                  │    │
│  │                                                                      │    │
│  │  ┌─────────────────────────────────────────────────────────────┐   │    │
│  │  │                    API LAYER (Gin/Echo)                      │   │    │
│  │  │  • REST endpoints for config, positions, orders             │   │    │
│  │  │  • WebSocket for real-time updates                          │   │    │
│  │  └─────────────────────────────────────────────────────────────┘   │    │
│  │                                                                      │    │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │    │
│  │  │   Data      │ │  Indicator  │ │  Strategy   │ │    Risk     │   │    │
│  │  │  Ingestion  │ │   Service   │ │   Engine    │ │   Manager   │   │    │
│  │  │             │ │             │ │             │ │             │   │    │
│  │  │ • Klines    │ │ • RSI       │ │ • Trend     │ │ • Position  │   │    │
│  │  │ • Trades    │ │ • ADX/DMI   │ │ • Reversion │ │   Sizing    │   │    │
│  │  │ • OrderBook │ │ • Bollinger │ │ • Breakout  │ │ • Stop Loss │   │    │
│  │  │ • WebSocket │ │ • ATR       │ │ • Volatility│ │ • Drawdown  │   │    │
│  │  │             │ │ • MACD      │ │ • StatArb   │ │ • Circuit   │   │    │
│  │  │             │ │ • Volume    │ │             │ │   Breakers  │   │    │
│  │  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘   │    │
│  │         │               │               │               │          │    │
│  │  ┌──────▼───────────────▼───────────────▼───────────────▼──────┐   │    │
│  │  │                  ORCHESTRATOR / SWITCHING CONTROLLER         │   │    │
│  │  │  • Regime detection                                          │   │    │
│  │  │  • Strategy scoring & selection                              │   │    │
│  │  │  • Signal aggregation                                        │   │    │
│  │  │  • Order generation                                          │   │    │
│  │  └─────────────────────────────┬────────────────────────────────┘   │    │
│  │                                │                                     │    │
│  │  ┌─────────────────────────────▼────────────────────────────────┐   │    │
│  │  │                    ORDER EXECUTOR                             │   │    │
│  │  │  • Order validation                                          │   │    │
│  │  │  • Binance API integration                                   │   │    │
│  │  │  • Execution tracking                                        │   │    │
│  │  └──────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────┬───────────────────────────────────┘    │
│                                    │                                         │
│  ┌─────────────────────────────────▼───────────────────────────────────┐    │
│  │                         BINANCE API                                  │    │
│  │  • REST: /api/v3/klines, /api/v3/ticker, /api/v3/order             │    │
│  │  • WebSocket: wss://stream.binance.com:9443/ws/ethusdt@kline_1m    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **Data Ingestion** | Fetch and stream market data from Binance |
| **Indicator Service** | Calculate technical indicators in real-time |
| **Strategy Engine** | Execute strategy logic, generate signals |
| **Risk Manager** | Enforce risk rules, position sizing, stops |
| **Orchestrator** | Coordinate strategies, regime detection |
| **Order Executor** | Place and manage orders on Binance |
| **API Layer** | Expose REST/WebSocket endpoints for frontend |

---

## Binance API Integration

### REST Endpoints Used

| Endpoint | Purpose | Rate Limit Weight |
|----------|---------|-------------------|
| `GET /api/v3/klines` | Historical candlesticks | 2 |
| `GET /api/v3/ticker/price` | Current price | 2 |
| `GET /api/v3/ticker/24hr` | 24h statistics | 2 |
| `GET /api/v3/depth` | Order book | 5-50 |
| `POST /api/v3/order` | Place order | 1 |
| `DELETE /api/v3/order` | Cancel order | 1 |
| `GET /api/v3/account` | Account info | 20 |

### WebSocket Streams

```go
// Kline stream for candlestick data
wss://stream.binance.com:9443/ws/ethusdt@kline_1m

// Trade stream for real-time trades
wss://stream.binance.com:9443/ws/ethusdt@trade

// Depth stream for order book
wss://stream.binance.com:9443/ws/ethusdt@depth20@100ms

// Combined stream
wss://stream.binance.com:9443/stream?streams=ethusdt@kline_1m/ethusdt@trade
```

### Sample Data Structures

```go
type Kline struct {
    OpenTime   int64   `json:"t"`
    Open       string  `json:"o"`
    High       string  `json:"h"`
    Low        string  `json:"l"`
    Close      string  `json:"c"`
    Volume     string  `json:"v"`
    CloseTime  int64   `json:"T"`
    Trades     int     `json:"n"`
}

type OrderRequest struct {
    Symbol      string  `json:"symbol"`
    Side        string  `json:"side"`      // BUY or SELL
    Type        string  `json:"type"`      // LIMIT, MARKET, STOP_LOSS, etc.
    Quantity    float64 `json:"quantity"`
    Price       float64 `json:"price,omitempty"`
    StopPrice   float64 `json:"stopPrice,omitempty"`
    TimeInForce string  `json:"timeInForce,omitempty"`
}
```

---

## Implementation Guide

### Project Structure

```
eth-trading/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers.go          # HTTP handlers
│   │   ├── websocket.go         # WebSocket handlers
│   │   └── router.go            # Route definitions
│   ├── binance/
│   │   ├── client.go            # Binance REST client
│   │   ├── websocket.go         # Binance WebSocket client
│   │   └── types.go             # Binance data types
│   ├── storage/                 # Data storage layer
│   │   ├── candle.go            # Candle model
│   │   ├── queue.go             # Circular queue implementation
│   │   ├── queue_manager.go     # Multi-timeframe queue manager
│   │   ├── sqlite.go            # SQLite database connection
│   │   ├── candle_repo.go       # Candle repository (CRUD)
│   │   ├── trade_repo.go        # Trade repository
│   │   ├── position_repo.go     # Position repository
│   │   ├── order_repo.go        # Order repository
│   │   └── data_service.go      # Coordinates queue + SQLite
│   ├── indicators/
│   │   ├── rsi.go               # RSI calculation
│   │   ├── adx.go               # ADX/DMI calculation
│   │   ├── bollinger.go         # Bollinger Bands
│   │   ├── atr.go               # Average True Range
│   │   ├── macd.go              # MACD
│   │   └── moving_average.go    # SMA, EMA
│   ├── strategy/
│   │   ├── interface.go         # Strategy interface
│   │   ├── trend.go             # Trend following
│   │   ├── mean_reversion.go    # Mean reversion
│   │   ├── breakout.go          # Breakout strategy
│   │   ├── volatility.go        # Volatility-based
│   │   └── stat_arb.go          # Statistical arbitrage
│   ├── risk/
│   │   ├── manager.go           # Risk manager
│   │   ├── position_sizer.go    # Position sizing
│   │   ├── stop_loss.go         # Stop loss logic
│   │   ├── take_profit.go       # Take profit logic
│   │   ├── drawdown.go          # Drawdown protection
│   │   └── circuit_breaker.go   # Circuit breakers
│   ├── orchestrator/
│   │   ├── engine.go            # Main orchestrator
│   │   ├── regime.go            # Regime detection
│   │   └── switcher.go          # Strategy switching
│   └── models/
│       ├── position.go          # Position model
│       ├── order.go             # Order model
│       └── signal.go            # Signal model
├── data/                        # SQLite database files
│   └── trading.db               # Main database
├── pkg/
│   └── utils/
│       └── math.go              # Math utilities
├── web/                         # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── services/
│   └── package.json
├── configs/
│   └── config.yaml              # Configuration file
├── docs/
│   └── ETH-Trading-Bot-Documentation.md
├── go.mod
├── go.sum
└── README.md
```

### Strategy Interface

```go
package strategy

type Signal struct {
    Action    string   // "buy", "sell", "hold"
    Strength  float64  // 0.0 to 1.0
    StopLoss  float64
    Target    float64
    Reason    string
}

type Strategy interface {
    Name() string
    CheckSignal(candles []Candle, indicators map[string]float64) Signal
    GetRequiredIndicators() []string
    IsActive() bool
    SetActive(active bool)
}
```

### Main Orchestrator Loop

```go
func (o *Orchestrator) Run(ctx context.Context) error {
    ticker := time.NewTicker(o.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // 1. Fetch latest data
            candles, err := o.dataService.GetLatestCandles()
            if err != nil {
                log.Error("Failed to fetch candles", err)
                continue
            }

            // 2. Calculate indicators
            indicators := o.indicatorService.Calculate(candles)

            // 3. Detect regime
            regime := o.regimeDetector.Detect(indicators)

            // 4. Get signals from active strategies
            signals := make(map[string]Signal)
            for _, strategy := range o.strategies {
                if o.switcher.IsStrategyActive(strategy.Name(), regime) {
                    signals[strategy.Name()] = strategy.CheckSignal(candles, indicators)
                }
            }

            // 5. Aggregate signals
            finalSignal := o.aggregateSignals(signals, regime)

            // 6. Apply risk management
            if finalSignal.Action != "hold" {
                finalSignal = o.riskManager.ValidateSignal(finalSignal)
            }

            // 7. Execute if valid
            if finalSignal.Action != "hold" {
                o.executor.Execute(finalSignal)
            }

            // 8. Update positions (stop loss, take profit)
            o.riskManager.UpdatePositions(candles[len(candles)-1])

            // 9. Broadcast to frontend
            o.broadcaster.Send(DashboardUpdate{
                Regime:     regime,
                Indicators: indicators,
                Signals:    signals,
                Positions:  o.riskManager.GetPositions(),
            })
        }
    }
}
```

---

## Configuration Reference

### Sample Configuration File

```yaml
# configs/config.yaml

server:
  port: 8080
  websocket_port: 8081

binance:
  api_key: "${BINANCE_API_KEY}"
  api_secret: "${BINANCE_API_SECRET}"
  base_url: "https://api.binance.com"
  websocket_url: "wss://stream.binance.com:9443"

trading:
  symbol: "ETHUSDT"
  timeframes:
    - "1m"
    - "5m"
    - "15m"
    - "1h"
  default_timeframe: "5m"

indicators:
  rsi:
    period: 14
    overbought: 70
    oversold: 30
  adx:
    period: 14
    trend_threshold: 25
  bollinger:
    period: 20
    std_dev: 2.0
  atr:
    period: 14
  macd:
    fast_period: 12
    slow_period: 26
    signal_period: 9
  moving_averages:
    fast: 10
    slow: 50
    trend: 200

strategies:
  trend:
    enabled: true
    weight: 0.3
  mean_reversion:
    enabled: true
    weight: 0.25
  breakout:
    enabled: true
    weight: 0.2
  volatility:
    enabled: true
    weight: 0.15
  stat_arb:
    enabled: false
    weight: 0.1

risk:
  # Position Sizing
  max_position_size_percent: 0.1     # 10% of account
  risk_per_trade_percent: 0.01       # 1% risk per trade

  # Stop Loss
  default_stop_type: "atr"
  stop_loss_atr_multiple: 2.0
  stop_loss_percent: 0.02            # 2% fallback
  use_trailing_stop: true
  trailing_stop_atr: 1.5

  # Take Profit
  use_take_profit: true
  take_profit_atr_multiple: 3.0
  use_scale_out: true
  scale_out_targets:
    - percent: 0.33
      risk_multiple: 1.0
    - percent: 0.33
      risk_multiple: 2.0
    - percent: 0.34
      risk_multiple: 3.0

  # Drawdown Protection
  max_daily_drawdown: 0.03           # 3%
  max_weekly_drawdown: 0.07          # 7%
  max_total_drawdown: 0.15           # 15%

  # Circuit Breakers
  max_consecutive_losses: 5
  max_losses_per_hour: 3
  max_losses_per_day: 7

  # Exposure Limits
  max_long_exposure: 0.5             # 50% of account
  max_short_exposure: 0.3            # 30% of account
  max_total_exposure: 0.6            # 60% of account

alerts:
  enabled: true
  telegram:
    enabled: true
    bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
  email:
    enabled: false
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    from: "${EMAIL_FROM}"
    to: "${EMAIL_TO}"

  thresholds:
    drawdown_warning: 0.05           # Alert at 5% drawdown
    profit_notification: 0.02        # Alert at 2% profit
    volatility_spike: 2.0            # Alert if ATR > 2x average

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: "logs/trading.log"
```

---

## Performance Benchmarks

| Strategy | Period | Trades | Return | Max Drawdown | Sharpe |
|----------|--------|--------|--------|--------------|--------|
| ORB (4.5% threshold) | 2017-2022 | 190 | +190% | ~$3k on $10k | - |
| Mean Reversion | 2016-2024 | 75 | ~$71k | - | - |
| Stablecoin Z-score | 2021 bear | - | +7% vs -26% B&H | - | - |
| ETH/USDC Pool Signal | 2021 bear | - | +12% vs -27% B&H | - | - |

---

## Sources & References

1. [Trading system: mean reverting strategy on Ethereum](https://en.cryptonomist.ch/2024/06/17/mean-reverting-trading-system-on-cryptocurrencies-a-strategy-to-exploit-the-false-breakouts-of-ethereum-eth/)
2. [Developing and Backtesting Winning ETH Trading Strategies](https://blog.amberdata.io/developing-and-backtesting-winning-eth-trading-strategies-report)
3. [Opening range breakout on Ethereum](https://en.cryptonomist.ch/2022/11/19/opening-range-breakout-ethereum-2/)
4. [Backtesting Results - Crypto Quant Models Guide](https://menthorq.com/guide/backtesting-results-crypto-quant-models/)
5. [Binance Spot API Documentation](https://developers.binance.com/docs/binance-spot-api-docs/rest-api/market-data-endpoints)

---

## Disclaimer

This documentation is for educational and research purposes only. Cryptocurrency trading involves substantial risk of loss. Past performance does not guarantee future results. Always conduct thorough backtesting and paper trading before deploying any trading system with real capital.
