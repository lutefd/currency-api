package handler

import (
	"net/http"

	"github.com/Lutefd/challenge-bravo/internal/commons"
)

func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	commons.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
