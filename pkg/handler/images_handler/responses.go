package image

import (
	"encoding/json"
	"net/http"
)

func responseError(w http.ResponseWriter, err error, statusCode int) {
	bts, err := json.Marshal(err)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
	}

	w.WriteHeader(statusCode)
	_, _ = w.Write(bts)
}

func responseJSON(w http.ResponseWriter, v any) {
	bts, err := json.Marshal(v)
	if err != nil {
		responseError(w, err, http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bts)
}
