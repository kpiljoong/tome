package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/paths"
)

func Start(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/journal", handleJournal)
	mux.HandleFunc("/blob", handleBlob)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	logx.Info("🚀 Server started at http://localhost:%d", port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server.ListenAndServe()
}

func handleJournal(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	query := r.URL.Query().Get("query")

	if ns == "" || query == "" {
		http.Error(w, "Missing namespace or query", http.StatusBadRequest)
		return
	}
	if err := paths.ValidateNamespace(ns); err != nil {
		http.Error(w, fmt.Sprintf("Invalid namespace: %v", err), http.StatusBadRequest)
		return
	}

	entries, err := core.Search(ns, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error searching for files: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func handleBlob(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "Missing hash", http.StatusBadRequest)
		return
	}
	if err := paths.ValidateBlobHash(hash); err != nil {
		http.Error(w, fmt.Sprintf("Invalid blob hash: %v", err), http.StatusBadRequest)
		return
	}

	data, err := core.GetBlobByHash(hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read blob: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}
