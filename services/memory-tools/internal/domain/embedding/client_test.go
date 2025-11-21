package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBGE_M3_Client_Embed(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embed" {
			t.Errorf("Expected path /embed, got %s", r.URL.Path)
		}

		var req EmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		// Return mock embeddings
		inputs, ok := req.Inputs.([]interface{})
		if !ok {
			// Single string input
			inputs = []interface{}{req.Inputs}
		}

		embeddings := make([][]float32, len(inputs))
		for i := range embeddings {
			embeddings[i] = make([]float32, 1024)
			// Fill with mock values
			for j := range embeddings[i] {
				embeddings[i][j] = float32(i+j) / 1024.0
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	// Create client with noop cache
	cacheConfig := CacheConfig{
		Type: "noop",
	}
	client, err := NewBGE_M3_Client(server.URL, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test single embedding
	ctx := context.Background()
	embeddings, err := client.Embed(ctx, []string{"test query"})
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	if len(embeddings) != 1 {
		t.Errorf("Expected 1 embedding, got %d", len(embeddings))
	}

	if len(embeddings[0]) != 1024 {
		t.Errorf("Expected 1024 dimensions, got %d", len(embeddings[0]))
	}
}

func TestBGE_M3_Client_BatchEmbed(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req EmbedRequest
		json.NewDecoder(r.Body).Decode(&req)

		inputs := req.Inputs.([]interface{})
		embeddings := make([][]float32, len(inputs))
		for i := range embeddings {
			embeddings[i] = make([]float32, 1024)
			for j := range embeddings[i] {
				embeddings[i][j] = float32(i+j) / 1024.0
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	cacheConfig := CacheConfig{Type: "noop"}
	client, err := NewBGE_M3_Client(server.URL, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test batch embedding
	ctx := context.Background()
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := client.Embed(ctx, texts)
	if err != nil {
		t.Fatalf("Batch embed failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) != 1024 {
			t.Errorf("Embedding %d: expected 1024 dimensions, got %d", i, len(emb))
		}
	}
}

func TestBGE_M3_Client_CacheHit(t *testing.T) {
	callCount := 0

	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		var req EmbedRequest
		json.NewDecoder(r.Body).Decode(&req)

		inputs := []interface{}{req.Inputs}
		embeddings := make([][]float32, 1)
		embeddings[0] = make([]float32, 1024)
		for j := range embeddings[0] {
			embeddings[0][j] = 0.5
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	// Create client with memory cache
	cacheConfig := CacheConfig{
		Type:    "memory",
		MaxSize: 100,
		TTL:     1 * time.Hour,
	}
	client, err := NewBGE_M3_Client(server.URL, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// First call - cache miss
	_, err = client.Embed(ctx, []string{"test query"})
	if err != nil {
		t.Fatalf("First embed failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 API call after first request, got %d", callCount)
	}

	// Second call - cache hit
	_, err = client.Embed(ctx, []string{"test query"})
	if err != nil {
		t.Fatalf("Second embed failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 API call after cache hit, got %d", callCount)
	}
}

func TestBGE_M3_Client_ValidateServer(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/info":
			info := ModelInfo{ModelID: "BAAI/bge-m3"}
			json.NewEncoder(w).Encode(info)
		case "/embed":
			embeddings := [][]float32{make([]float32, 1024)}
			json.NewEncoder(w).Encode(embeddings)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cacheConfig := CacheConfig{Type: "noop"}
	client, err := NewBGE_M3_Client(server.URL, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	if err := client.ValidateServer(ctx); err != nil {
		t.Errorf("ValidateServer failed: %v", err)
	}
}

func TestMemoryCache(t *testing.T) {
	cache, err := NewMemoryCache(10, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}

	// Test set and get
	embedding := []float32{0.1, 0.2, 0.3}
	cache.Set("key1", embedding, 1*time.Second)

	retrieved, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find cached item")
	}

	if len(retrieved) != len(embedding) {
		t.Errorf("Expected %d elements, got %d", len(embedding), len(retrieved))
	}

	// Test expiration
	cache.Set("key2", embedding, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	_, found = cache.Get("key2")
	if found {
		t.Error("Expected expired item to not be found")
	}
}

func TestNoOpsCache(t *testing.T) {
	cache := NewNoOpsCache()

	// Set should do nothing
	embedding := []float32{0.1, 0.2, 0.3}
	cache.Set("key1", embedding, 1*time.Hour)

	// Get should always return not found
	_, found := cache.Get("key1")
	if found {
		t.Error("NoOps cache should never return found")
	}
}

func BenchmarkEmbed_Single(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		embeddings := [][]float32{make([]float32, 1024)}
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	cacheConfig := CacheConfig{Type: "noop"}
	client, _ := NewBGE_M3_Client(server.URL, cacheConfig)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Embed(ctx, []string{"test query"})
	}
}

func BenchmarkEmbed_Batch32(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		embeddings := make([][]float32, 32)
		for i := range embeddings {
			embeddings[i] = make([]float32, 1024)
		}
		json.NewEncoder(w).Encode(embeddings)
	}))
	defer server.Close()

	cacheConfig := CacheConfig{Type: "noop"}
	client, _ := NewBGE_M3_Client(server.URL, cacheConfig)
	ctx := context.Background()

	texts := make([]string, 32)
	for i := range texts {
		texts[i] = "test query"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Embed(ctx, texts)
	}
}
