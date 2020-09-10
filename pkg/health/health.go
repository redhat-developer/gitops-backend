package health

import (
	"encoding/json"
	"log"
	"net/http"
)

var GitRevision = "unknown"

// Handler returns the value of GitRevision in simple struct.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		Version string `json:"version"`
	}{Version: GitRevision}); err != nil {
		log.Printf("failed to encode Health response: %s", err)
	}
}
