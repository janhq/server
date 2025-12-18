package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"jan-server/services/realtime-api/internal/domain/session"
	"jan-server/services/realtime-api/internal/infrastructure/store"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/handlers"
	"jan-server/services/realtime-api/internal/utils/platformerrors"
)

// RegisterRealtimeRoutes registers the realtime session routes.
func RegisterRealtimeRoutes(router gin.IRoutes, handler *handlers.SessionHandler) {
	// Session management endpoints
	router.POST("/realtime/sessions", createSession(handler))

	// Extension endpoints
	router.GET("/realtime/sessions", listSessions(handler))
	router.GET("/realtime/sessions/:id", getSession(handler))
	router.DELETE("/realtime/sessions/:id", deleteSession(handler))
}

// createSession godoc
// @Summary      Create a realtime session
// @Description  Creates a new realtime session with LiveKit token. No request body required.
// @Tags         Realtime API
// @Accept       json
// @Produce      json
// @Success      201 {object} session.Session
// @Failure      401 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Security     BearerAuth
// @Router       /realtime/sessions [post]
func createSession(handler *handlers.SessionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := extractUserID(c)

		sess, err := handler.CreateSession(c.Request.Context(), &session.CreateSessionRequest{}, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				Error: &errorDetail{
					Message: err.Error(),
					Type:    "server_error",
				},
			})
			return
		}

		// POST response: include client_secret, room_id, user_id but NOT status
		c.JSON(http.StatusCreated, sess)
	}
}

// listSessions godoc
// @Summary      List realtime sessions
// @Description  Lists all active sessions for the current user
// @Tags         Realtime API
// @Produce      json
// @Success      200 {object} session.ListSessionsResponse
// @Failure      401 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Security     BearerAuth
// @Router       /realtime/sessions [get]
func listSessions(handler *handlers.SessionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := extractUserID(c)

		sessions, err := handler.ListUserSessions(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse{
				Error: &errorDetail{
					Message: err.Error(),
					Type:    "server_error",
				},
			})
			return
		}

		// GET response: include status, room_id, user_id but NOT client_secret
		result := make([]*session.Session, len(sessions))
		for i, s := range sessions {
			result[i] = &session.Session{
				ID:     s.ID,
				Object: s.Object,
				WsURL:  s.WsURL,
				RoomID: s.Room,
				UserID: s.UserID,
				Status: s.State,
			}
		}

		c.JSON(http.StatusOK, session.ListSessionsResponse{
			Object: "list",
			Data:   result,
		})
	}
}

// getSession godoc
// @Summary      Get a realtime session
// @Description  Retrieves a specific session by ID. Users can only access their own sessions.
// @Tags         Realtime API
// @Produce      json
// @Param        id path string true "Session ID"
// @Success      200 {object} session.Session
// @Failure      403 {object} errorResponse
// @Failure      404 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Security     BearerAuth
// @Router       /realtime/sessions/{id} [get]
func getSession(handler *handlers.SessionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := extractUserID(c)

		sess, err := handler.GetSession(c.Request.Context(), id)
		if err != nil {
			status, errType := mapError(err)
			c.JSON(status, errorResponse{
				Error: &errorDetail{
					Message: err.Error(),
					Type:    errType,
				},
			})
			return
		}

		// Authorization: verify session belongs to the authenticated user
		if sess.UserID != userID {
			c.JSON(http.StatusForbidden, errorResponse{
				Error: &errorDetail{
					Message: "access denied",
					Type:    "forbidden_error",
				},
			})
			return
		}

		// GET response: include status, room_id, user_id but NOT client_secret
		result := &session.Session{
			ID:     sess.ID,
			Object: sess.Object,
			WsURL:  sess.WsURL,
			RoomID: sess.Room,
			UserID: sess.UserID,
			Status: sess.State,
		}

		c.JSON(http.StatusOK, result)
	}
}

// deleteSession godoc
// @Summary      Delete a realtime session
// @Description  Ends a session and invalidates its token. Users can only delete their own sessions.
// @Tags         Realtime API
// @Produce      json
// @Param        id path string true "Session ID"
// @Success      200 {object} session.DeleteSessionResponse
// @Failure      403 {object} errorResponse
// @Failure      404 {object} errorResponse
// @Failure      500 {object} errorResponse
// @Security     BearerAuth
// @Router       /realtime/sessions/{id} [delete]
func deleteSession(handler *handlers.SessionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := extractUserID(c)

		// First, get the session to verify ownership
		sess, err := handler.GetSession(c.Request.Context(), id)
		if err != nil {
			status, errType := mapError(err)
			c.JSON(status, errorResponse{
				Error: &errorDetail{
					Message: err.Error(),
					Type:    errType,
				},
			})
			return
		}

		// Authorization: verify session belongs to the authenticated user
		if sess.UserID != userID {
			c.JSON(http.StatusForbidden, errorResponse{
				Error: &errorDetail{
					Message: "access denied",
					Type:    "forbidden_error",
				},
			})
			return
		}

		if err := handler.DeleteSession(c.Request.Context(), id); err != nil {
			status, errType := mapError(err)
			c.JSON(status, errorResponse{
				Error: &errorDetail{
					Message: err.Error(),
					Type:    errType,
				},
			})
			return
		}

		c.JSON(http.StatusOK, session.DeleteSessionResponse{
			ID:      id,
			Object:  "realtime.session.deleted",
			Deleted: true,
		})
	}
}

// Helper types and functions

type errorResponse struct {
	Error *errorDetail `json:"error"`
}

type errorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

func extractUserID(c *gin.Context) string {
	// Check for user_id set directly by middleware (for Kong API key auth)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok && id != "" {
			return id
		}
	}

	// Fall back to extracting from JWT token
	if token, exists := c.Get("auth_token"); exists {
		if jwtToken, ok := token.(*jwt.Token); ok {
			if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
				if sub, ok := claims["sub"].(string); ok {
					return sub
				}
			}
		}
	}
	return "anonymous"
}

func mapError(err error) (int, string) {
	if errors.Is(err, store.ErrSessionNotFound) {
		return http.StatusNotFound, "not_found_error"
	}
	if errors.Is(err, store.ErrSessionAlreadyExists) || errors.Is(err, store.ErrRoomAlreadyExists) {
		return http.StatusConflict, "conflict_error"
	}
	if platformerrors.IsErrorType(err, platformerrors.ErrorTypeNotFound) {
		return http.StatusNotFound, "not_found_error"
	}
	return http.StatusInternalServerError, "server_error"
}
