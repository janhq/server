package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
	"github.com/rs/zerolog/log"
)

type MemoryHandler struct {
	service *memory.Service
}

func NewMemoryHandler(service *memory.Service) *MemoryHandler {
	return &MemoryHandler{service: service}
}

// HandleLoad handles POST /v1/memory/load
func (h *MemoryHandler) HandleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req memory.MemoryLoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode load request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Str("query", req.Query).
		Msg("Memory load request received")

	// Load memories
	resp, err := h.service.Load(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load memories")
		http.Error(w, "Failed to load memories", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}

// HandleObserve handles POST /v1/memory/observe
func (h *MemoryHandler) HandleObserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req memory.MemoryObserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode observe request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if req.ConversationID == "" {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}
	if len(req.Messages) == 0 {
		http.Error(w, "messages are required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Str("conversation_id", req.ConversationID).
		Int("message_count", len(req.Messages)).
		Msg("Memory observe request received")

	// Observe and store
	if err := h.service.Observe(r.Context(), req); err != nil {
		log.Error().Err(err).Msg("Failed to observe memories")
		http.Error(w, "Failed to observe memories", http.StatusInternalServerError)
		return
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Memory observation completed",
	})
}

// HandleStats handles GET /v1/memory/stats
func (h *MemoryHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}

	projectID := r.URL.Query().Get("project_id")

	stats, err := h.service.GetMemoryStats(r.Context(), userID, projectID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get memory stats")
		http.Error(w, "Failed to get memory stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// HandleExport handles GET /v1/memory/export
func (h *MemoryHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}

	exportData, err := h.service.ExportMemory(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to export memory")
		http.Error(w, "Failed to export memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=memory_export.json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(exportData))
}

// HandleHealth handles GET /healthz
func (h *MemoryHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "memory-tools",
	})
}

// HandleUserUpsert handles POST /v1/memory/user/upsert
func (h *MemoryHandler) HandleUserUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req memory.UserMemoryUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode user upsert request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items are required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("user_id", req.UserID).
		Int("item_count", len(req.Items)).
		Msg("User memory upsert request received")

	// Upsert user memories
	ids, err := h.service.UpsertUserMemories(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upsert user memories")
		http.Error(w, "Failed to upsert user memories", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "User memories upserted successfully",
		"ids":     ids,
	})
}

// HandleProjectUpsert handles POST /v1/memory/project/upsert
func (h *MemoryHandler) HandleProjectUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req memory.ProjectFactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode project upsert request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ProjectID == "" {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}
	if len(req.Facts) == 0 {
		http.Error(w, "facts are required", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("project_id", req.ProjectID).
		Int("fact_count", len(req.Facts)).
		Msg("Project fact upsert request received")

	// Upsert project facts
	ids, err := h.service.UpsertProjectFacts(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upsert project facts")
		http.Error(w, "Failed to upsert project facts", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Project facts upserted successfully",
		"ids":     ids,
	})
}

// HandleDelete handles DELETE /v1/memory/delete
func (h *MemoryHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req memory.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to decode delete request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.IDs) == 0 {
		http.Error(w, "ids are required", http.StatusBadRequest)
		return
	}

	log.Info().
		Int("id_count", len(req.IDs)).
		Msg("Memory delete request received")

	// Delete memories
	deletedCount, err := h.service.DeleteMemories(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete memories")
		http.Error(w, "Failed to delete memories", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"message":       "Memories deleted successfully",
		"deleted_count": deletedCount,
	})
}
