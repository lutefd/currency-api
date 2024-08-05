package handler

import (
	"net/http"
	"strconv"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/service"
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
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
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
	// TODO: Implement adding a new currency
}

func (h *CurrencyHandler) RemoveCurrency(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement removing a currency
}
