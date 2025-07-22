package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(map[string]string{"status": "ok"})
	if err != nil {
		slog.Error("failed to encode json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(buf.Bytes()); err != nil {
		slog.Error("failed to write json health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", health)

	slog.Info("starting server", "address", ":4000")

	if err := http.ListenAndServe(":4000", mux); err != nil {
		slog.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
