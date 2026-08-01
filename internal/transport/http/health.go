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
	// Marshal first so a (theoretically impossible) encoding failure can't
	// leave a truncated body under an already-committed 200.
	body, err := json.Marshal(healthResponse{Status: "ok"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// A write failure here means the client hung up — nothing actionable remains.
	_, _ = w.Write(body)
}
