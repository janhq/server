package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
	"github.com/janhq/jan-server/services/memory-tools/internal/interfaces/httpserver/responses"
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
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodPost {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req memory.MemoryLoadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode load request")
		responses.Error(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.UserID == "" {
		responses.Error(w, r, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.Query == "" {
		responses.Error(w, r, http.StatusBadRequest, "query is required")
		return
	}

	logger.Info().
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Str("query", req.Query).
		Msg("Memory load request received")

	// Load memories
	resp, err := h.service.Load(r.Context(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to load memories")
		responses.Error(w, r, http.StatusInternalServerError, "failed to load memories")
		return
	}

	// Return response
	responses.JSON(w, r, http.StatusOK, resp)
}

// HandleObserve handles POST /v1/memory/observe
func (h *MemoryHandler) HandleObserve(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodPost {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req memory.MemoryObserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode observe request")
		responses.Error(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.UserID == "" {
		responses.Error(w, r, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.ConversationID == "" {
		responses.Error(w, r, http.StatusBadRequest, "conversation_id is required")
		return
	}
	if len(req.Messages) == 0 {
		responses.Error(w, r, http.StatusBadRequest, "messages are required")
		return
	}

	logger.Info().
		Str("user_id", req.UserID).
		Str("project_id", req.ProjectID).
		Str("conversation_id", req.ConversationID).
		Int("message_count", len(req.Messages)).
		Msg("Memory observe request received")

	// Observe and store
	if err := h.service.Observe(r.Context(), req); err != nil {
		logger.Error().Err(err).Msg("Failed to observe memories")
		responses.Error(w, r, http.StatusInternalServerError, "failed to observe memories")
		return
	}

	// Return success
	responses.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Memory observation completed",
	})
}

// HandleStats handles GET /v1/memory/stats
func (h *MemoryHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodGet {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		responses.Error(w, r, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	projectID := r.URL.Query().Get("project_id")

	stats, err := h.service.GetMemoryStats(r.Context(), userID, projectID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get memory stats")
		responses.Error(w, r, http.StatusInternalServerError, "failed to get memory stats")
		return
	}

	responses.JSON(w, r, http.StatusOK, stats)
}

// HandleExport handles GET /v1/memory/export
func (h *MemoryHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodGet {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		responses.Error(w, r, http.StatusBadRequest, "user_id query parameter is required")
		return
	}

	exportData, err := h.service.ExportMemory(r.Context(), userID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to export memory")
		responses.Error(w, r, http.StatusInternalServerError, "failed to export memory")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=memory_export.json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(exportData))
}

// HandleHealth handles GET /healthz
func (h *MemoryHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "memory-tools",
	})
}

// HandleUserUpsert handles POST /v1/memory/user/upsert
func (h *MemoryHandler) HandleUserUpsert(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodPost {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req memory.UserMemoryUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode user upsert request")
		responses.Error(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.UserID == "" {
		responses.Error(w, r, http.StatusBadRequest, "user_id is required")
		return
	}
	if len(req.Items) == 0 {
		responses.Error(w, r, http.StatusBadRequest, "items are required")
		return
	}

	logger.Info().
		Str("user_id", req.UserID).
		Int("item_count", len(req.Items)).
		Msg("User memory upsert request received")

	// Upsert user memories
	ids, err := h.service.UpsertUserMemories(r.Context(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert user memories")
		responses.Error(w, r, http.StatusInternalServerError, "failed to upsert user memories")
		return
	}

	// Return response
	responses.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "User memories upserted successfully",
		"ids":     ids,
	})
}

// HandleProjectUpsert handles POST /v1/memory/project/upsert
func (h *MemoryHandler) HandleProjectUpsert(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodPost {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req memory.ProjectFactUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode project upsert request")
		responses.Error(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.ProjectID == "" {
		responses.Error(w, r, http.StatusBadRequest, "project_id is required")
		return
	}
	if len(req.Facts) == 0 {
		responses.Error(w, r, http.StatusBadRequest, "facts are required")
		return
	}

	logger.Info().
		Str("project_id", req.ProjectID).
		Int("fact_count", len(req.Facts)).
		Msg("Project fact upsert request received")

	// Upsert project facts
	ids, err := h.service.UpsertProjectFacts(r.Context(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to upsert project facts")
		responses.Error(w, r, http.StatusInternalServerError, "failed to upsert project facts")
		return
	}

	// Return response
	responses.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Project facts upserted successfully",
		"ids":     ids,
	})
}

// HandleDelete handles DELETE /v1/memory/delete
func (h *MemoryHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	if logger == nil {
		logger = &log.Logger
	}

	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		responses.Error(w, r, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req memory.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error().Err(err).Msg("Failed to decode delete request")
		responses.Error(w, r, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if len(req.IDs) == 0 {
		responses.Error(w, r, http.StatusBadRequest, "ids are required")
		return
	}

	logger.Info().
		Int("id_count", len(req.IDs)).
		Msg("Memory delete request received")

	// Delete memories
	deletedCount, err := h.service.DeleteMemories(r.Context(), req)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to delete memories")
		responses.Error(w, r, http.StatusInternalServerError, "failed to delete memories")
		return
	}

	// Return response
	responses.JSON(w, r, http.StatusOK, map[string]interface{}{
		"status":        "success",
		"message":       "Memories deleted successfully",
		"deleted_count": deletedCount,
	})
}
