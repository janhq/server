package main

import (
	"net/http"
	"os"
	"time"

	"jan-server/services/mcp-tools/tools/vector-store-service/store"

	"github.com/gin-gonic/gin"
)

type config struct {
	Port string
}

func loadConfig() config {
	port := os.Getenv("VECTOR_STORE_PORT")
	if port == "" {
		port = "3015"
	}
	return config{Port: port}
}

type indexRequest struct {
	DocumentID string         `json:"document_id" binding:"required"`
	Text       string         `json:"text" binding:"required"`
	Metadata   map[string]any `json:"metadata"`
	Tags       []string       `json:"tags"`
}

type queryRequest struct {
	Text   string   `json:"text" binding:"required"`
	TopK   int      `json:"top_k"`
	Filter []string `json:"document_ids"`
}

func main() {
	cfg := loadConfig()
	memStore := store.NewMemoryStore()

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	router.POST("/documents", func(c *gin.Context) {
		var req indexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		doc := store.Document{
			ID:        req.DocumentID,
			Text:      req.Text,
			Tags:      req.Tags,
			Metadata:  req.Metadata,
			Embedding: store.BuildEmbedding(req.Text),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		memStore.Upsert(doc)

		c.JSON(http.StatusCreated, gin.H{
			"status":      "indexed",
			"document_id": doc.ID,
			"token_count": len(doc.Embedding),
			"indexed_at":  doc.UpdatedAt.Format(time.RFC3339),
		})
	})

	router.POST("/query", func(c *gin.Context) {
		var req queryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		topK := req.TopK
		if topK <= 0 {
			topK = 5
		}
		if topK > 20 {
			topK = 20
		}

		results := memStore.Query(store.BuildEmbedding(req.Text), topK, req.Filter)
		response := make([]map[string]any, 0, len(results))
		for _, result := range results {
			response = append(response, map[string]any{
				"document_id":  result.Document.ID,
				"score":        result.Score,
				"text_preview": previewText(result.Document.Text),
				"metadata":     result.Document.Metadata,
				"tags":         result.Document.Tags,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"query":   req.Text,
			"top_k":   topK,
			"count":   len(response),
			"results": response,
		})
	})

	addr := ":" + cfg.Port
	if err := router.Run(addr); err != nil {
		panic(err)
	}
}

func previewText(text string) string {
	runes := []rune(text)
	if len(runes) <= 240 {
		return text
	}
	return string(runes[:240]) + "…"
}
