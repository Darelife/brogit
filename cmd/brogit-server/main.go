package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"brogit/pkg/api"
	"brogit/pkg/replay"
	"brogit/pkg/store"
)

var globalStore *store.Store

func handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := globalStore.Add(req.Entry); err != nil {
		log.Printf("Error adding to store: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.PushResponse{Status: "ok"})
}

func handleCommit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Replay logic
	entries, err := globalStore.GetAll()
	if err != nil {
		log.Printf("Error getting entries: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// We assume the server is running in the root of the repo for now, or use a specific path
	repoPath, _ := os.Getwd()
	replayer := replay.NewReplayer(repoPath)

	if err := replayer.ProcessCommands(entries); err != nil {
		log.Printf("Error processing commands: %v", err)
		http.Error(w, fmt.Sprintf("Error replaying commands: %v", err), http.StatusInternalServerError)
		return
	}

	// Clear store after successful replay
	if err := globalStore.Clear(); err != nil {
		log.Printf("Error clearing store: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "committed", "count": fmt.Sprintf("%d", len(entries))})
}

func main() {
	globalStore = store.NewStore("")

	http.HandleFunc("/api/push", handlePush)
	http.HandleFunc("/api/commit", handleCommit)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
