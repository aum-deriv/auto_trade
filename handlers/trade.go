package handlers

import (
	"auto_trade/trading"
	"encoding/json"
	"net/http"
)

type BuyRequest struct {
	Symbol string `json:"symbol"`
}

type SellRequest struct {
	TradeID string `json:"trade_id"`
}

type TradeResponse struct {
	TradeID string  `json:"trade_id"`
	Symbol  string  `json:"symbol"`
	Price   float64 `json:"price"`
	Status  string  `json:"status"`
}

type TradeHandler struct {
	TradeManager *trading.TradeManager
}

func NewTradeHandler() *TradeHandler {
	return &TradeHandler{}
}

func (h *TradeHandler) Sell(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SellRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	trade, err := h.TradeManager.CloseTrade(req.TradeID, 0) // Price will be updated in response
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := TradeResponse{
		TradeID: trade.ID,
		Symbol:  trade.Symbol,
		Price:   trade.Price,
		Status:  trade.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TradeHandler) Buy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real system, you would get the current price from a price feed
	// For now, we'll use a hardcoded price based on the symbol
	var currentPrice float64
	switch req.Symbol {
	case "BTC/USD":
		currentPrice = 40000.0
	case "ETH/USD":
		currentPrice = 2500.0
	case "SOL/USD":
		currentPrice = 100.0
	default:
		http.Error(w, "Invalid symbol", http.StatusBadRequest)
		return
	}

	trade, err := h.TradeManager.PlaceBuyOrder(req.Symbol, currentPrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TradeResponse{
		TradeID: trade.ID,
		Symbol:  trade.Symbol,
		Price:   trade.Price,
		Status:  trade.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
