package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// healthHandler returns service health status
func (g *gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "ok",
		"service":       "gateway",
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"ml_classifier": g.mlClassifierURL,
		"ml_generator":  g.mlGeneratorURL,
	})
}
