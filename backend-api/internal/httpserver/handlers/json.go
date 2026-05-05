package handlers

import (
	"demo-server/internal/httpserver/models"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, response models.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
