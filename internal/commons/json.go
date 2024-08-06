package commons

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Lutefd/challenge-bravo/internal/logger"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		logger.Errorf("responding with %d error: %s", code, msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	RespondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
