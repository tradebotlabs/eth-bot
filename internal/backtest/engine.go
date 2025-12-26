package backtest

import (
	"fmt"
	"math"
	"time"

	"github.com/eth-trading/internal/indicators"
	"github.com/eth-trading/internal/strategy"
)

// Config holds backtest configuration
type Config struct {
	Symbol         string
	Timeframe      string
	StartDate      time.Time
	EndDate        time.Time
	InitialCapital float64
	Commission     float64
	Slippage       float64
	RiskPerTrade   float64
	Strategies     []strategy.Strategy
}

// Engine runs backtests
type Engine struct {
	config          *Config
	indicatorMgr    *indicators.Manager
	regimeDetector  *strategy.RegimeDetector
	scorer          *strategy.Scorer
}

// NewEngine creates a new backtest engine
func NewEngine(config *Config) *Engine {
	indicatorMgr := indicators.NewManager(indicators.DefaultConfig())
	regimeDetector := strategy.NewRegimeDetector(strategy.DefaultRegimeConfig(), indicatorMgr)
	scorer := strategy.NewScorer(strategy.DefaultScorerConfig())

	// Add strategies to scorer
	for _, strat := range config.Strategies {
		scorer.AddStrategy(strat)
	}

	return &Engine{
		config:         config,
		indicatorMgr:   indicatorMgr,
		regimeDetector: regimeDetector,
		scorer:         scorer,
	}
}

// Run executes the backtest
func (e *Engine) Run(data *HistoricalData) (*Result, error) {
	if data == nil || len(data.Candles) == 0 {
		return nil, fmt.Errorf("no historical data provided")
	}

	result := &Result{
		Config:         e.config,
		Metrics:        &Metrics{},
		EquityCurve:    []EquityPoint{},
		Trades:         []Trade{},
		MonthlyReturns: make(map[string]float64),
		StrategyStats:  make(map[string]StrategyStats),
		StartTime:      time.Now(),
	}

	portfolio := NewPortfolio(e.config.InitialCapital)

	// Minimum data needed for indicators
	minDataPoints := 100
	for _, strat := range e.config.Strategies {
		if strat.GetMinDataPoints() > minDataPoints {
			minDataPoints = strat.GetMinDataPoints()
		}
	}

	// Run through historical data
	for i := minDataPoints; i < len(data.Candles); i++ {
		candle := data.Candles[i]

		// Build market data for this point in time
		marketData := e.buildMarketData(data, i)

		// Update portfolio with current price
		portfolio.UpdatePrice(candle.Close)

		// Check exit conditions for open positions
		e.checkExits(portfolio, marketData, &result.Trades)

		// Get regime
		regime := e.regimeDetector.Detect(
			marketData.Opens,
			marketData.Highs,
			marketData.Lows,
			marketData.Closes,
			marketData.Volumes,
		)
		marketData.Regime = regime

		// Get combined score from all strategies
		score := e.scorer.Score(marketData, regime)

		// Enter new position if signal is strong enough
		if score.ShouldTrade && len(portfolio.Positions) == 0 {
			e.enterPosition(portfolio, marketData, score, &result.Trades)
		}

		// Record equity
		result.EquityCurve = append(result.EquityCurve, EquityPoint{
			Timestamp: candle.Timestamp,
			Equity:    portfolio.GetEquity(),
			Cash:      portfolio.Cash,
			Drawdown:  portfolio.GetDrawdown(),
		})
	}

	// Close any remaining positions
	if len(portfolio.Positions) > 0 {
		lastCandle := data.Candles[len(data.Candles)-1]
		for _, pos := range portfolio.Positions {
			trade := e.closePosition(portfolio, pos, lastCandle.Close, "backtest_end")
			result.Trades = append(result.Trades, trade)
		}
	}

	// Calculate metrics
	e.calculateMetrics(result, portfolio)

	result.EndTime = time.Now()
	result.ExecutionTime = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// buildMarketData creates MarketData from historical data up to index i
func (e *Engine) buildMarketData(data *HistoricalData, i int) *strategy.MarketData {
	// Extract data up to current point
	opens := make([]float64, i+1)
	highs := make([]float64, i+1)
	lows := make([]float64, i+1)
	closes := make([]float64, i+1)
	volumes := make([]float64, i+1)

	for j := 0; j <= i; j++ {
		opens[j] = data.Candles[j].Open
		highs[j] = data.Candles[j].High
		lows[j] = data.Candles[j].Low
		closes[j] = data.Candles[j].Close
		volumes[j] = data.Candles[j].Volume
	}

	// Calculate indicators
	analysis := e.indicatorMgr.Analyze(opens, highs, lows, closes, volumes)

	marketData := &strategy.MarketData{
		Symbol:       e.config.Symbol,
		Timeframe:    e.config.Timeframe,
		Timestamp:    data.Candles[i].Timestamp,
		Opens:        opens,
		Highs:        highs,
		Lows:         lows,
		Closes:       closes,
		Volumes:      volumes,
		Analysis:     analysis,
		CurrentPrice: data.Candles[i].Close,
		Bid:          data.Candles[i].Close,
		Ask:          data.Candles[i].Close,
	}

	return marketData
}

// enterPosition enters a new position based on signal
func (e *Engine) enterPosition(portfolio *Portfolio, data *strategy.MarketData, score strategy.CombinedScore, trades *[]Trade) {
	if score.BestSignal == nil {
		return
	}

	// Calculate position size based on risk
	entryPrice := e.applySlippage(data.CurrentPrice, score.Direction)
	stopLoss := score.BestSignal.StopLoss

	if stopLoss == 0 {
		// Fallback stop loss
		if score.Direction == strategy.DirectionLong {
			stopLoss = entryPrice * 0.98
		} else {
			stopLoss = entryPrice * 1.02
		}
	}

	// Calculate position size based on risk per trade
	riskPerShare := math.Abs(entryPrice - stopLoss)
	if riskPerShare == 0 {
		return
	}

	riskAmount := portfolio.GetEquity() * e.config.RiskPerTrade
	quantity := riskAmount / riskPerShare

	// Limit position size to available capital
	maxQuantity := (portfolio.Cash * 0.95) / entryPrice
	if quantity > maxQuantity {
		quantity = maxQuantity
	}

	if quantity <= 0 {
		return
	}

	// Calculate cost including commission
	cost := quantity * entryPrice
	commission := cost * e.config.Commission

	if cost+commission > portfolio.Cash {
		return
	}

	// Open position
	pos := &Position{
		ID:         int64(len(*trades) + 1),
		Symbol:     data.Symbol,
		Strategy:   score.BestSignal.Strategy,
		Direction:  score.Direction,
		EntryPrice: entryPrice,
		EntryTime:  data.Timestamp,
		Quantity:   quantity,
		StopLoss:   stopLoss,
		TakeProfit: score.BestSignal.TakeProfit,
		Commission: commission,
	}

	portfolio.OpenPosition(pos, cost+commission)
}

// checkExits checks if any positions should be exited
func (e *Engine) checkExits(portfolio *Portfolio, data *strategy.MarketData, trades *[]Trade) {
	var toClose []*Position

	for _, pos := range portfolio.Positions {
		shouldExit := false
		exitReason := ""

		// Check stop loss
		if pos.Direction == strategy.DirectionLong {
			if data.CurrentPrice <= pos.StopLoss {
				shouldExit = true
				exitReason = "stop_loss"
			} else if pos.TakeProfit > 0 && data.CurrentPrice >= pos.TakeProfit {
				shouldExit = true
				exitReason = "take_profit"
			}
		} else {
			if data.CurrentPrice >= pos.StopLoss {
				shouldExit = true
				exitReason = "stop_loss"
			} else if pos.TakeProfit > 0 && data.CurrentPrice <= pos.TakeProfit {
				shouldExit = true
				exitReason = "take_profit"
			}
		}

		// Check strategy exit signal
		if !shouldExit {
			for _, strat := range e.config.Strategies {
				if strat.Name() == pos.Strategy {
					stratPos := &strategy.Position{
						ID:         pos.ID,
						Symbol:     pos.Symbol,
						Direction:  pos.Direction,
						EntryPrice: pos.EntryPrice,
						Quantity:   pos.Quantity,
						CurrentPrice: data.CurrentPrice,
						StopLoss:   pos.StopLoss,
						TakeProfit: pos.TakeProfit,
						Strategy:   pos.Strategy,
						OpenTime:   pos.EntryTime,
					}
					exit, reason := strat.ShouldExit(data, stratPos)
					if exit {
						shouldExit = true
						exitReason = reason
					}
					break
				}
			}
		}

		if shouldExit {
			toClose = append(toClose, pos)
			trade := e.closePosition(portfolio, pos, data.CurrentPrice, exitReason)
			*trades = append(*trades, trade)
		}
	}

	// Remove closed positions
	for _, pos := range toClose {
		portfolio.ClosePosition(pos.ID)
	}
}

// closePosition closes a position and returns the trade record
func (e *Engine) closePosition(portfolio *Portfolio, pos *Position, exitPrice float64, exitReason string) Trade {
	exitPrice = e.applySlippage(exitPrice, -pos.Direction)

	// Calculate P&L
	var pnl float64
	if pos.Direction == strategy.DirectionLong {
		pnl = (exitPrice - pos.EntryPrice) * pos.Quantity
	} else {
		pnl = (pos.EntryPrice - exitPrice) * pos.Quantity
	}

	// Subtract commissions
	exitCommission := exitPrice * pos.Quantity * e.config.Commission
	netPnl := pnl - pos.Commission - exitCommission

	// Return cash to portfolio
	proceeds := pos.Quantity * exitPrice
	portfolio.Cash += proceeds - exitCommission

	returnPercent := netPnl / (pos.EntryPrice * pos.Quantity) * 100

	trade := Trade{
		ID:            pos.ID,
		Symbol:        pos.Symbol,
		Strategy:      pos.Strategy,
		Direction:     pos.Direction.String(),
		EntryTime:     pos.EntryTime,
		ExitTime:      time.Now(),
		EntryPrice:    pos.EntryPrice,
		ExitPrice:     exitPrice,
		Quantity:      pos.Quantity,
		NetProfit:     netPnl,
		ReturnPercent: returnPercent,
		ExitReason:    exitReason,
		Commission:    pos.Commission + exitCommission,
	}

	return trade
}

// applySlippage applies slippage to price
func (e *Engine) applySlippage(price float64, direction strategy.Direction) float64 {
	if e.config.Slippage == 0 {
		return price
	}

	if direction == strategy.DirectionLong {
		return price * (1 + e.config.Slippage)
	} else if direction == strategy.DirectionShort {
		return price * (1 - e.config.Slippage)
	}

	return price
}

// calculateMetrics calculates backtest metrics
func (e *Engine) calculateMetrics(result *Result, portfolio *Portfolio) {
	metrics := result.Metrics

	metrics.StartingCapital = e.config.InitialCapital
	metrics.EndingCapital = portfolio.GetEquity()
	metrics.NetProfit = metrics.EndingCapital - metrics.StartingCapital
	metrics.TotalReturn = metrics.NetProfit / metrics.StartingCapital

	// Trade statistics
	metrics.TotalTrades = len(result.Trades)
	if metrics.TotalTrades == 0 {
		return
	}

	var totalWin, totalLoss float64
	var holdingTimes []time.Duration

	for _, trade := range result.Trades {
		if trade.NetProfit > 0 {
			metrics.WinningTrades++
			totalWin += trade.NetProfit
			if trade.NetProfit > metrics.LargestWin {
				metrics.LargestWin = trade.NetProfit
			}
		} else {
			metrics.LosingTrades++
			totalLoss += math.Abs(trade.NetProfit)
			if trade.NetProfit < metrics.LargestLoss {
				metrics.LargestLoss = trade.NetProfit
			}
		}

		holdingTime := trade.ExitTime.Sub(trade.EntryTime)
		holdingTimes = append(holdingTimes, holdingTime)
	}

	metrics.WinRate = float64(metrics.WinningTrades) / float64(metrics.TotalTrades)

	if metrics.WinningTrades > 0 {
		metrics.AvgWin = totalWin / float64(metrics.WinningTrades)
	}
	if metrics.LosingTrades > 0 {
		metrics.AvgLoss = totalLoss / float64(metrics.LosingTrades)
	}

	if totalLoss > 0 {
		metrics.ProfitFactor = totalWin / totalLoss
	}

	// Expectancy
	metrics.Expectancy = (metrics.WinRate * metrics.AvgWin) - ((1 - metrics.WinRate) * metrics.AvgLoss)

	// Calculate drawdown from equity curve
	e.calculateDrawdown(result)

	// Calculate Sharpe and other ratios
	e.calculateRatios(result)

	// Calculate average holding time
	if len(holdingTimes) > 0 {
		var totalDuration time.Duration
		for _, d := range holdingTimes {
			totalDuration += d
		}
		avgDuration := totalDuration / time.Duration(len(holdingTimes))
		metrics.AvgHoldingTime = avgDuration.String()
	}

	// Strategy-specific stats
	e.calculateStrategyStats(result)
}

// calculateDrawdown calculates maximum drawdown from equity curve
func (e *Engine) calculateDrawdown(result *Result) {
	if len(result.EquityCurve) == 0 {
		return
	}

	peak := result.EquityCurve[0].Equity
	maxDD := 0.0

	for _, point := range result.EquityCurve {
		if point.Equity > peak {
			peak = point.Equity
		}

		dd := (peak - point.Equity) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}

	result.Metrics.MaxDrawdown = maxDD

	if maxDD > 0 {
		result.Metrics.RecoveryFactor = result.Metrics.NetProfit / (maxDD * e.config.InitialCapital)
	}
}

// calculateRatios calculates performance ratios
func (e *Engine) calculateRatios(result *Result) {
	if len(result.EquityCurve) < 2 {
		return
	}

	// Calculate returns for each period
	var returns []float64
	for i := 1; i < len(result.EquityCurve); i++ {
		prev := result.EquityCurve[i-1].Equity
		curr := result.EquityCurve[i].Equity
		ret := (curr - prev) / prev
		returns = append(returns, ret)
	}

	// Calculate annualized return
	days := result.Config.EndDate.Sub(result.Config.StartDate).Hours() / 24
	years := days / 365.0
	if years > 0 {
		result.Metrics.AnnualizedReturn = math.Pow(1+result.Metrics.TotalReturn, 1/years) - 1
	}

	// Calculate Sharpe ratio (assuming 0% risk-free rate)
	if len(returns) > 1 {
		mean := 0.0
		for _, r := range returns {
			mean += r
		}
		mean /= float64(len(returns))

		variance := 0.0
		for _, r := range returns {
			diff := r - mean
			variance += diff * diff
		}
		variance /= float64(len(returns))
		stdDev := math.Sqrt(variance)

		if stdDev > 0 {
			// Annualize Sharpe
			periodsPerYear := 365.0 * 24.0 / (days / float64(len(returns)))
			result.Metrics.SharpeRatio = (mean / stdDev) * math.Sqrt(periodsPerYear)
		}
	}

	// Calculate Sortino ratio (downside deviation)
	var downside []float64
	for _, r := range returns {
		if r < 0 {
			downside = append(downside, r*r)
		}
	}

	if len(downside) > 0 {
		downsideVar := 0.0
		for _, d := range downside {
			downsideVar += d
		}
		downsideVar /= float64(len(downside))
		downsideStd := math.Sqrt(downsideVar)

		if downsideStd > 0 {
			periodsPerYear := 365.0 * 24.0 / (days / float64(len(returns)))
			mean := result.Metrics.AnnualizedReturn / periodsPerYear
			result.Metrics.SortinoRatio = (mean / downsideStd) * math.Sqrt(periodsPerYear)
		}
	}

	// Calmar ratio
	if result.Metrics.MaxDrawdown > 0 {
		result.Metrics.CalmarRatio = result.Metrics.AnnualizedReturn / result.Metrics.MaxDrawdown
	}
}

// calculateStrategyStats calculates per-strategy statistics
func (e *Engine) calculateStrategyStats(result *Result) {
	strategyTrades := make(map[string][]Trade)

	for _, trade := range result.Trades {
		strategyTrades[trade.Strategy] = append(strategyTrades[trade.Strategy], trade)
	}

	for stratName, trades := range strategyTrades {
		stats := StrategyStats{
			Name:        stratName,
			TotalTrades: len(trades),
		}

		var wins, totalWin, totalLoss float64
		for _, trade := range trades {
			if trade.NetProfit > 0 {
				wins++
				totalWin += trade.NetProfit
			} else {
				totalLoss += math.Abs(trade.NetProfit)
			}
		}

		stats.WinRate = wins / float64(len(trades))
		if totalLoss > 0 {
			stats.ProfitFactor = totalWin / totalLoss
		}
		stats.NetProfit = totalWin - totalLoss
		stats.Contribution = stats.NetProfit / result.Metrics.NetProfit

		result.StrategyStats[stratName] = stats
	}
}
