package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/go-chi/chi/v5"
)

type CurrencyHandler struct {
	currencyService *service.CurrencyService
}

func NewCurrencyHandler(currencyService *service.CurrencyService) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}

func (h *CurrencyHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	from := strings.ToUpper(r.URL.Query().Get("from"))
	to := strings.ToUpper(r.URL.Query().Get("to"))
	amountStr := r.URL.Query().Get("amount")

	if from == "" || to == "" || amountStr == "" {
		commons.RespondWithError(w, http.StatusBadRequest, "Missing required parameters")
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		commons.RespondWithError(w, http.StatusBadRequest, "Invalid amount")
		return
	}

	result, err := h.currencyService.Convert(r.Context(), from, to, amount)
	if err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "Conversion failed")
		return
	}

	commons.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"from":   from,
		"to":     to,
		"amount": amount,
		"result": result,
	})
}

func (h *CurrencyHandler) AddCurrency(w http.ResponseWriter, r *http.Request) {
	var currency struct {
		Code string  `json:"code"`
		Rate float64 `json:"rate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&currency); err != nil {
		commons.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := h.currencyService.AddCurrency(r.Context(), currency.Code, currency.Rate); err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "Failed to add currency")
		return
	}

	commons.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Currency added successfully"})
}

func (h *CurrencyHandler) RemoveCurrency(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if err := h.currencyService.RemoveCurrency(r.Context(), code); err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "Failed to remove currency")
		return
	}

	commons.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Currency removed successfully"})
}
