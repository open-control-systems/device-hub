package htcore

import (
	"net/http"
	"strconv"
)

// WriteText writes text to HTTP response.
func WriteText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(text)))

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(text)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
