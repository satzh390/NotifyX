package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type WebhookRequest struct {
	URL       string                 `json:"url"`
	Method    string                 `json:"method"`
	Headers   map[string][]string    `json:"headers"`
	Body      map[string]interface{} `json:"body"`
	Timestamp time.Time              `json:"timestamp"`
}

type WebhookStore struct {
	mu       sync.RWMutex
	webhooks []WebhookRequest
}

var store = &WebhookStore{
	webhooks: make([]WebhookRequest, 0),
}

func main() {
	r := mux.NewRouter()

	// Catch-all webhook endpoint
	r.PathPrefix("/").HandlerFunc(handleWebhook).Methods("POST", "PUT", "PATCH")

	// Get all received webhooks
	r.HandleFunc("/webhooks", handleGetWebhooks).Methods("GET")

	// Clear all webhooks
	r.HandleFunc("/webhooks", handleClearWebhooks).Methods("DELETE")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	log.Println("Webhook receiver starting on :8888")
	if err := http.ListenAndServe(":8888", r); err != nil {
		log.Fatal("Webhook receiver error:", err)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body map[string]interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			// If not JSON, store as string
			body = map[string]interface{}{
				"raw": string(bodyBytes),
			}
		}
	}

	// Store the webhook
	store.mu.Lock()
	store.webhooks = append(store.webhooks, WebhookRequest{
		URL:       r.RequestURI,
		Method:    r.Method,
		Headers:   r.Header,
		Body:      body,
		Timestamp: time.Now(),
	})
	store.mu.Unlock()

	// Return success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func handleGetWebhooks(w http.ResponseWriter, r *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	json.NewEncoder(w).Encode(store.webhooks)
}

func handleClearWebhooks(w http.ResponseWriter, r *http.Request) {
	store.mu.Lock()
	store.webhooks = make([]WebhookRequest, 0)
	store.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}
