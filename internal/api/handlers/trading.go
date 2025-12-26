package handlers

import (
	"net/http"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// TradingHandler handles trading control endpoints
type TradingHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewTradingHandler creates a new trading handler
func NewTradingHandler(orch *orchestrator.Orchestrator) *TradingHandler {
	return &TradingHandler{orchestrator: orch}
}

// TradingStateResponse represents trading state response
type TradingStateResponse struct {
	State *orchestrator.TradingState `json:"state"`
}

// GetState returns current trading state
func (h *TradingHandler) GetState(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()
	return c.JSON(http.StatusOK, TradingStateResponse{State: state})
}

// Start starts the trading bot
func (h *TradingHandler) Start(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	if err := h.orchestrator.Start(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "started"})
}

// Stop stops the trading bot
func (h *TradingHandler) Stop(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	h.orchestrator.Stop()
	return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
}

// Pause pauses trading
func (h *TradingHandler) Pause(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	h.orchestrator.Pause()
	return c.JSON(http.StatusOK, map[string]string{"status": "paused"})
}

// Resume resumes trading
func (h *TradingHandler) Resume(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	h.orchestrator.Resume()
	return c.JSON(http.StatusOK, map[string]string{"status": "resumed"})
}

// ModeResponse represents trading mode response
type ModeResponse struct {
	Mode string `json:"mode"`
}

// ModeRequest represents trading mode request
type ModeRequest struct {
	Mode string `json:"mode"`
}

// GetMode returns current trading mode
func (h *TradingHandler) GetMode(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()
	return c.JSON(http.StatusOK, ModeResponse{Mode: state.Mode.String()})
}

// SetMode sets the trading mode
func (h *TradingHandler) SetMode(c echo.Context) error {
	var req ModeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Mode switching would require recreating executor
	// For now, just acknowledge the request
	return c.JSON(http.StatusOK, ModeResponse{Mode: req.Mode})
}
