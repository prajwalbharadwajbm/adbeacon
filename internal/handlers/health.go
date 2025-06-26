package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/prajwalbharadwajbm/adbeacon/internal/logger"
)

// Health is required to check
//   - to ensure reliability and proper functioning
//   - ensuring application availability: this is useful if hosted on cloud
//     as azure function app or aws ec2 for this instance to monitor availability.
func Health(w http.ResponseWriter, r *http.Request) {

	envelope := map[string]any{
		"status": "available",
		"application-details": map[string]any{
			"version":     "1.0.0",
			"environment": config.AppConfigInstance.GeneralConfig.Env,
		},
		"dependencies": map[string]any{
			"database": map[string]any{
				"status": "not-implemented-yet!",
			},
		},
	}

	healthObj, err := json.Marshal(envelope)
	if err != nil {
		logger.Log.Error("failed to marshal health check response", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(healthObj)
}
