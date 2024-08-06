package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/go-chi/chi/v5"
)

type CurrencyHandler struct {
	currencyService service.CurrencyServiceInterface
}

func NewCurrencyHandler(currencyService service.CurrencyServiceInterface) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}

func (h *CurrencyHandler) ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	from := strings.ToUpper(r.URL.Query().Get("from"))
	to := strings.ToUpper(r.URL.Query().Get("to"))
	amountStr := r.URL.Query().Get("amount")

	if from == "" || to == "" || (amountStr == "") {
		commons.RespondWithError(w, http.StatusBadRequest, "missing required parameters")
		return
	}
	if len(from) != 3 || len(to) != 3 {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid currency code, must be 3 characters long following ISO 4217")
		return
	}
	amount, err := parseAmount(amountStr)
	if err != nil {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	if amount < 0 {
		commons.RespondWithError(w, http.StatusBadRequest, "amount must be non-negative")
		return
	}

	result, err := h.currencyService.Convert(r.Context(), from, to, amount)
	if err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "conversion failed")
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
		Code string      `json:"code"`
		Rate interface{} `json:"rate_to_usd"`
	}

	if err := json.NewDecoder(r.Body).Decode(&currency); err != nil {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	if currency.Code == "" {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid currency code")
		return
	}
	if len(currency.Code) != 3 {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid currency code, must be 3 characters long following ISO 4217")
		return
	}

	user, ok := r.Context().Value("user").(model.User)
	if !ok {
		commons.RespondWithError(w, http.StatusInternalServerError, "user information not available")
		return
	}
	rate, err := parseRate(currency.Rate)
	if err != nil {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid rate: "+err.Error())
		return
	}

	if rate <= 0 {
		commons.RespondWithError(w, http.StatusBadRequest, "rate must be positive")
		return
	}
	newCurrency := &model.Currency{
		Code:      strings.ToUpper(currency.Code),
		Rate:      rate,
		CreatedBy: user.ID,
		UpdatedBy: user.ID,
	}

	if err := h.currencyService.AddCurrency(r.Context(), newCurrency); err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "failed to add currency")
		return
	}

	commons.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "currency added successfully"})
}

func (h *CurrencyHandler) RemoveCurrency(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))

	if code == "" {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid currency code")
		return
	}
	if len(code) != 3 {
		commons.RespondWithError(w, http.StatusBadRequest, "invalid currency code, must be 3 characters long following ISO 4217")
		return
	}
	if err := h.currencyService.RemoveCurrency(r.Context(), code); err != nil {
		commons.RespondWithError(w, http.StatusInternalServerError, "failed to remove currency")
		return
	}

	commons.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "currency removed successfully"})
}
func parseAmount(amountStr string) (float64, error) {
	amountStr = strings.Replace(amountStr, ",", ".", -1)
	return strconv.ParseFloat(amountStr, 64)
}
func parseRate(rate interface{}) (float64, error) {
	switch v := rate.(type) {
	case float64:
		return v, nil
	case string:
		v = strings.Replace(v, ",", ".", -1)
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported rate type")
	}
}
