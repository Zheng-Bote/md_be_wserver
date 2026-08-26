package http

import (
	"encoding/json"
	"net/http"
	"log"
	
	"md_be_wserver/internal/db"
)

type UploadMeasurementsRequest struct {
	Measurements []db.Measurement `json:"measurements"`
}

func (s *Server) handleMeasurementsUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UploadMeasurementsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(req.Measurements) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok", "message":"No measurements to upload"}`))
		return
	}

	if err := s.Repo.SaveMeasurements(r.Context(), req.Measurements); err != nil {
		log.Printf("ERROR: Failed to save measurements: %v\n", err)
		http.Error(w, "Failed to save measurements", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully saved %d measurements.\n", len(req.Measurements))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok", "message":"Measurements uploaded successfully"}`))
}
