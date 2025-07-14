package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aimrintech/x-backend/constants"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func getUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(constants.USER_ID_KEY).(string)
	if !ok {
		return "", errors.New("unauthorized")
	}
	return userID, nil
}
