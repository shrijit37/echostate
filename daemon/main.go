package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	addr              = ":9081"
	ollamaEndpoint    = "http://localhost:11434/api/generate"
	summaryThreshold  = 4000
	ollamaModel       = "llama3.2"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}

	state := NewState(pwd, ollamaEndpoint, ollamaModel, summaryThreshold)

	// Build initial state
	if err := state.Refresh(); err != nil {
		log.Printf("initial refresh: %v", err)
	}

	// Start file watcher
	watcher, err := NewWatcher(state, ".echo-state.md")
	if err != nil {
		log.Fatalf("watcher: %v", err)
	}
	defer watcher.Close()

	// HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := state.Get()
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("encode: %v", err)
		}
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("daemon listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
