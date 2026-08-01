package tallyhttp

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

// healthHandler responds to GET /health with a simple 200 OK.
// Deliberately dumb — no DB check, no dependencies on anything else.
// This proves the process itself is alive and serving requests.
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
