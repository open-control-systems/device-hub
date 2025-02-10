package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"
)

func telemetryHandler(w http.ResponseWriter, _ *http.Request) {
	var telemetryData = map[string]any{
		"timestamp":   -1,
		"temperature": 22.5,
		"humidity":    60,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(telemetryData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func registrationHandler(w http.ResponseWriter, _ *http.Request) {
	var registrationData = map[string]any{
		"timestamp": -1,
		"device_id": "0xABCD",
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(registrationData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/telemetry", telemetryHandler)
	mux.HandleFunc("/registration", registrationHandler)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	addr := listener.Addr().String()
	log.Printf("Server is running at http://%s", addr)

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
