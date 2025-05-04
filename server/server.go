package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"cs4049-assignment2-A-i200438_i200650/kvstore"
	"cs4049-assignment2-A-i200438_i200650/raft"
)

type Server struct {
	kvStore *kvstore.KVStore
	kvRaft  *raft.KVRaft
	mu      sync.Mutex
}

func NewServer(kv *kvstore.KVStore, kvRaft *raft.KVRaft) *Server {
	return &Server{kvStore: kv, kvRaft: kvRaft}
}

func (s *Server) HandlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Use KVRaft for consensus
	success := s.kvRaft.Put(req.Key, req.Value)
	
	w.Header().Set("Content-Type", "application/json")
	
	if success {
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"status": "success"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"status": "error", "message": "Failed to reach consensus"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func (s *Server) HandleAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	// Use KVRaft for consensus
	success := s.kvRaft.Append(req.Key, req.Value)
	
	w.Header().Set("Content-Type", "application/json")
	
	if success {
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"status": "success"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		response := map[string]string{"status": "error", "message": "Failed to reach consensus"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}
	
	// Use KVRaft for consistent reads
	value, exists := s.kvRaft.Get(key)
	
	if !exists {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]string{"value": value}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) StartServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/put", s.HandlePut)
	mux.HandleFunc("/append", s.HandleAppend)
	mux.HandleFunc("/get", s.HandleGet)
	
	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	fmt.Printf("Server running on port %d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}