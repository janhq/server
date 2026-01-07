package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"jan-server/services/media-api/internal/config"
	domain "jan-server/services/media-api/internal/domain/media"
	"jan-server/services/media-api/internal/interfaces/httpserver/responses"
	"jan-server/services/media-api/internal/utils/platformerrors"
)

// MediaHandler exposes media endpoints.
type MediaHandler struct {
	cfg     *config.Config
	service *domain.Service
	log     zerolog.Logger
}

func NewMediaHandler(cfg *config.Config, service *domain.Service, log zerolog.Logger) *MediaHandler {
	return &MediaHandler{
		cfg:     cfg,
		service: service,
		log:     log.With().Str("component", "media-handler").Logger(),
	}
}

type ingestResponse struct {
	ID      string `json:"id"`
	Mime    string `json:"mime"`
	Bytes   int64  `json:"bytes"`
	Deduped bool   `json:"deduped"`
	URL     string `json:"url"`
}

type resolveRequest struct {
	Payload JSONPayload `json:"payload" binding:"required"`
}

type resolveResponse struct {
	Payload JSONPayload `json:"payload"`
}

// JSONPayload is used to document arbitrary JSON blobs in swagger.
type JSONPayload = json.RawMessage

// Ingest godoc
// @Summary      Upload media
// @Description  Accepts data URLs or remote URLs and stores content privately.
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        request  body      domain.IngestRequest  true  "Media request"
// @Success      200      {object}  ingestResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media [post]
func (h *MediaHandler) Ingest(c *gin.Context) {
	var req domain.IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.HandleNewError(c, platformerrors.ErrorTypeValidation, "invalid request body", "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d")
		return
	}

	obj, dedup, err := h.service.Ingest(c.Request.Context(), req)
	if err != nil {
		h.log.Error().Err(err).Msg("ingest failed")
		responses.HandleError(c, err, "failed to ingest media")
		return
	}

	// Generate direct public URL for embedding in HTML
	directURL := h.buildDirectURL(obj.ID)

	c.JSON(http.StatusOK, ingestResponse{
		ID:      obj.ID,
		Mime:    obj.MimeType,
		Bytes:   obj.Bytes,
		Deduped: dedup,
		URL:     directURL,
	})
}

// Resolve godoc
// @Summary      Resolve jan_* placeholders
// @Description  Replaces pseudo data URLs with short-lived signed URLs.
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        request  body      resolveRequest  true  "Payload to resolve"
// @Success      200      {object}  resolveResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media/resolve [post]
func (h *MediaHandler) Resolve(c *gin.Context) {
	var req resolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.HandleNewError(c, platformerrors.ErrorTypeValidation, "invalid request body", "2b3c4d5e-6f7a-8b9c-0d1e-2f3a4b5c6d7e")
		return
	}

	out, err := h.service.ResolvePayload(c.Request.Context(), json.RawMessage(req.Payload))
	if err != nil {
		h.log.Error().Err(err).Msg("resolve failed")
		responses.HandleError(c, err, "failed to resolve media payload")
		return
	}

	c.JSON(http.StatusOK, resolveResponse{Payload: JSONPayload(out)})
}

// PrepareUpload godoc
// @Summary      Request presigned upload URL
// @Description  Generates a presigned upload URL and reserves a jan_id. Client uploads directly to S3 using the URL. Not available for local storage.
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        request  body      domain.PrepareUploadRequest  true  "Upload preparation request"
// @Success      200      {object}  domain.UploadPreparation
// @Failure      400      {object}  map[string]string
// @Failure      501      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media/prepare-upload [post]
func (h *MediaHandler) PrepareUpload(c *gin.Context) {
	// Check if storage backend supports presigned uploads
	if !h.service.SupportsPresignedUploads() {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error":   "presigned uploads not supported",
			"message": "use POST /v1/media/upload for direct upload with local storage",
		})
		return
	}

	var req domain.PrepareUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.HandleNewError(c, platformerrors.ErrorTypeValidation, "invalid request body", "3c4d5e6f-7a8b-9c0d-1e2f-3a4b5c6d7e8f")
		return
	}

	prep, err := h.service.PrepareUpload(c.Request.Context(), req.MimeType, req.UserID)
	if err != nil {
		h.log.Error().Err(err).Msg("prepare upload failed")
		responses.HandleError(c, err, "failed to prepare upload")
		return
	}

	c.JSON(http.StatusOK, prep)
}

// GetPresignedURL godoc
// @Summary      Get presigned download URL
// @Description  Returns a temporary signed URL for downloading media by jan_id.
// @Tags         media
// @Produce      json
// @Param        id   path      string  true  "Media ID (jan_xxx)"
// @Success      200  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media/{id}/presign [get]
func (h *MediaHandler) GetPresignedURL(c *gin.Context) {
	id := c.Param("id")

	url, err := h.service.Presign(c.Request.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Str("id", id).Msg("presign failed")
		responses.HandleError(c, err, "failed to generate presigned URL")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         id,
		"url":        url,
		"expires_in": int(h.cfg.S3PresignTTL.Seconds()),
	})
}

// Proxy godoc
// @Summary      Stream media bytes
// @Description  Streams the object through the media API without exposing storage URLs. If proxying is disabled, returns a presigned download URL instead.
// @Tags         media
// @Produce      json
// @Produce      octet-stream
// @Param        id   path      string  true  "Media ID"
// @Success      200  {object}  map[string]interface{} "Presigned URL response when proxying is disabled; otherwise binary stream"
// @Failure      404  {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media/{id} [get]
func (h *MediaHandler) Proxy(c *gin.Context) {
	id := c.Param("id")

	if !h.cfg.ProxyDownload {
		url, err := h.service.Presign(c.Request.Context(), id)
		if err != nil {
			h.log.Error().Err(err).Msg("presign failed")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"url": url})
		return
	}

	reader, mime, err := h.service.Download(c.Request.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Msg("download failed")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	if mime == "" {
		mime = "application/octet-stream"
	}

	c.Header("Content-Type", mime)
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		h.log.Error().Err(err).Msg("stream error")
	}
}

// DirectUpload godoc
// @Summary      Direct file upload
// @Description  Accepts multipart file upload for local storage. Alternative to presigned uploads.
// @Tags         media
// @Accept       multipart/form-data
// @Produce      json
// @Param        file      formData  file    true  "File to upload"
// @Param        user_id   formData  string  false "User ID"
// @Success      200       {object}  ingestResponse
// @Failure      400       {object}  map[string]string
// @Security     ApiKeyAuth
// @Router       /v1/media/upload [post]
func (h *MediaHandler) DirectUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	userID := c.Request.FormValue("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to read file")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	// Determine content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Create an ingest request with data URL
	dataURL := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)

	req := domain.IngestRequest{
		Source: domain.Source{
			Type:    "data_url",
			DataURL: dataURL,
		},
		Filename: header.Filename,
		UserID:   userID,
	}

	obj, dedup, err := h.service.Ingest(c.Request.Context(), req)
	if err != nil {
		h.log.Error().Err(err).Msg("ingest failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate direct public URL for embedding in HTML
	directURL := h.buildDirectURL(obj.ID)

	c.JSON(http.StatusOK, ingestResponse{
		ID:      obj.ID,
		Mime:    obj.MimeType,
		Bytes:   obj.Bytes,
		Deduped: dedup,
		URL:     directURL,
	})
}

// PublicServe godoc
// @Summary      Serve media publicly
// @Description  Streams the media file directly for use in HTML img src. This endpoint does not require authentication.
// @Tags         media
// @Produce      image/jpeg
// @Produce      image/png
// @Produce      image/webp
// @Produce      image/gif
// @Param        id   path      string  true  "Media ID (jan_xxx)"
// @Success      200  {file}    binary
// @Failure      404  {object}  map[string]string
// @Router       /api/media/{id} [get]
func (h *MediaHandler) PublicServe(c *gin.Context) {
	id := c.Param("id")

	reader, mime, err := h.service.Download(c.Request.Context(), id)
	if err != nil {
		h.log.Error().Err(err).Str("id", id).Msg("public serve failed")
		c.JSON(http.StatusNotFound, gin.H{"error": "media not found"})
		return
	}
	defer reader.Close()

	if mime == "" {
		mime = "application/octet-stream"
	}

	// Set cache headers for browser caching
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("Content-Type", mime)
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		h.log.Error().Err(err).Msg("stream error")
	}
}

// buildDirectURL constructs the public URL for direct media access
func (h *MediaHandler) buildDirectURL(id string) string {
	publicURL := h.cfg.PublicURL
	if publicURL == "" {
		// Fallback to localhost if not configured
		publicURL = "http://localhost:8000"
	}
	return fmt.Sprintf("%s/api/media/%s", strings.TrimSuffix(publicURL, "/"), id)
}
