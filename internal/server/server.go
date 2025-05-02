package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kpiljoong/tome/internal/core"
	"github.com/kpiljoong/tome/pkg/logx"
)

func Start(port int) error {
	http.HandleFunc("/journal", handleJournal)
	http.HandleFunc("/blob", handleBlob)

	addr := fmt.Sprintf(":%d", port)
	logx.Info("🚀 Server started at http://localhost:%d", port)
	return http.ListenAndServe(addr, nil)
}

func handleJournal(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	query := r.URL.Query().Get("query")

	if ns == "" || query == "" {
		http.Error(w, "Missing namespace or query", http.StatusBadRequest)
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

	data, err := core.GetBlobByHash(hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read blob: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}
