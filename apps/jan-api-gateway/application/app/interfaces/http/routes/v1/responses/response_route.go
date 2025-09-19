package responses

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"menlo.ai/jan-api-gateway/app/domain/apikey"
	"menlo.ai/jan-api-gateway/app/domain/auth"
	"menlo.ai/jan-api-gateway/app/domain/common"
	"menlo.ai/jan-api-gateway/app/domain/conversation"
	"menlo.ai/jan-api-gateway/app/domain/response"
	"menlo.ai/jan-api-gateway/app/domain/user"
	responsetypes "menlo.ai/jan-api-gateway/app/interfaces/http/responses"
	"menlo.ai/jan-api-gateway/app/utils/ptr"
)

// moved: ResponseCreationResult now lives in response_model_handler.go

// ResponseRoute represents the response API routes
type ResponseRoute struct {
	userService                 *user.UserService
	authService                 *auth.AuthService
	apikeyService               *apikey.ApiKeyService
	conversationService         *conversation.ConversationService
	responseService             *response.ResponseService
	responseStreamingHandler    *ResponseStreamingHandler
	responseNonStreamingHandler *ResponseNonStreamingHandler
	responseModelHandler        *ResponseModelHandler
}

// NewResponseRoute creates a new ResponseRoute instance
func NewResponseRoute(userService *user.UserService, authService *auth.AuthService, apikeyService *apikey.ApiKeyService, conversationService *conversation.ConversationService, responseService *response.ResponseService, responseStreamingHandler *ResponseStreamingHandler, responseNonStreamingHandler *ResponseNonStreamingHandler, responseModelHandler *ResponseModelHandler) *ResponseRoute {
	return &ResponseRoute{
		userService:                 userService,
		authService:                 authService,
		apikeyService:               apikeyService,
		conversationService:         conversationService,
		responseService:             responseService,
		responseStreamingHandler:    responseStreamingHandler,
		responseNonStreamingHandler: responseNonStreamingHandler,
		responseModelHandler:        responseModelHandler,
	}
}

// RegisterRouter registers the response routes
func (responseRoute *ResponseRoute) RegisterRouter(router gin.IRouter) {
	responseRouter := router.Group("/responses")
	responseRoute.registerRoutes(responseRouter)
}

// registerRoutes registers all response routes
func (responseRoute *ResponseRoute) registerRoutes(router *gin.RouterGroup) {
	// Apply middleware to the entire group
	responseGroup := router.Group("",
		responseRoute.authService.AppUserAuthMiddleware(),
		responseRoute.authService.RegisteredUserMiddleware(),
	)

	responseGroup.POST("", responseRoute.CreateResponse)

	// Apply response middleware for routes that need response context
	responseMiddleWare := responseRoute.responseService.GetResponseMiddleWare()
	responseGroup.GET(fmt.Sprintf("/:%s", string(response.ResponseContextKeyPublicID)), responseMiddleWare, responseRoute.GetResponseHandler)
	responseGroup.DELETE(fmt.Sprintf("/:%s", string(response.ResponseContextKeyPublicID)), responseMiddleWare, responseRoute.DeleteResponseHandler)
	responseGroup.POST(fmt.Sprintf("/:%s/cancel", string(response.ResponseContextKeyPublicID)), responseMiddleWare, responseRoute.CancelResponseHandler)
	responseGroup.GET(fmt.Sprintf("/:%s/input_items", string(response.ResponseContextKeyPublicID)), responseMiddleWare, responseRoute.ListInputItemsHandler)
}

// CreateResponse creates a new response from LLM
// @Summary Create a response
// @Description Creates a new LLM response for the given input. Supports multiple input types including text, images, files, web search, and more.
// @Description
// @Description **Supported Input Types:**
// @Description - `text`: Plain text input
// @Description - `image`: Image input (URL or base64)
// @Description - `file`: File input by file ID
// @Description - `web_search`: Web search input
// @Description - `file_search`: File search input
// @Description - `streaming`: Streaming input
// @Description - `function_calls`: Function calls input
// @Description - `reasoning`: Reasoning input
// @Description
// @Description **Example Request:**
// @Description ```json
// @Description {
// @Description   "model": "gpt-4",
// @Description   "input": {
// @Description     "type": "text",
// @Description     "text": "Hello, how are you?"
// @Description   },
// @Description   "max_tokens": 100,
// @Description   "temperature": 0.7,
// @Description   "stream": false,
// @Description   "background": false
// @Description }
// @Description ```
// @Description
// @Description **Response Format:**
// @Description The response uses embedded structure where all fields are at the top level:
// @Description - `jan_status`: Jan API status code (optional)
// @Description - `id`: Response identifier
// @Description - `object`: Object type ("response")
// @Description - `created`: Unix timestamp
// @Description - `model`: Model used
// @Description - `status`: Response status
// @Description - `input`: Input data
// @Description - `output`: Generated output
// @Description
// @Description **Example Response:**
// @Description ```json
// @Description {
// @Description   "jan_status": "000000",
// @Description   "id": "resp_1234567890",
// @Description   "object": "response",
// @Description   "created": 1234567890,
// @Description   "model": "gpt-4",
// @Description   "status": "completed",
// @Description   "input": {
// @Description     "type": "text",
// @Description     "text": "Hello, how are you?"
// @Description   },
// @Description   "output": {
// @Description     "type": "text",
// @Description     "text": {
// @Description       "value": "I'm doing well, thank you!"
// @Description     }
// @Description   }
// @Description }
// @Description ```
// @Description
// @Description **Response Status:**
// @Description - `completed`: Response generation finished successfully
// @Description - `processing`: Response is being generated
// @Description - `failed`: Response generation failed
// @Description - `cancelled`: Response was cancelled
// @Tags Jan, Jan-Responses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body requesttypes.CreateResponseRequest true "Request payload containing model, input, and generation parameters"
// @Success 200 {object} responses.Response "Created response"
// @Success 202 {object} responses.Response "Response accepted for background processing"
// @Failure 400 {object} responsetypes.ErrorResponse "Invalid request payload"
// @Failure 401 {object} responsetypes.ErrorResponse "Unauthorized"
// @Failure 422 {object} responsetypes.ErrorResponse "Validation error"
// @Failure 429 {object} responsetypes.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} responsetypes.ErrorResponse "Internal server error"
// @Router /v1/responses [post]
func (responseRoute *ResponseRoute) CreateResponse(reqCtx *gin.Context) {
	user, _ := auth.GetUserFromContext(reqCtx)
	userID := user.ID

	// Parse and validate the request
	domainRequest, err := responseRoute.responseModelHandler.parseAndValidateRequest(reqCtx)
	if err != nil {
		return // Error response already sent
	}

	// Validate the request
	success, validationErr := ValidateCreateResponseRequest(domainRequest)
	if !success {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code:  validationErr.GetCode(),
			Error: validationErr.Error(),
		})
		return
	}

	responseRoute.responseModelHandler.ProcessResponseRequest(reqCtx, domainRequest, userID)
}

// moved handleResponseCreation to response_model_handler.go

// GetResponse retrieves a response by ID
// @Summary Get a response
// @Description Retrieves an LLM response by its ID. Returns the complete response object with embedded structure where all fields are at the top level.
// @Description
// @Description **Response Format:**
// @Description The response uses embedded structure where all fields are at the top level:
// @Description - `jan_status`: Jan API status code (optional)
// @Description - `id`: Response identifier
// @Description - `object`: Object type ("response")
// @Description - `created`: Unix timestamp
// @Description - `model`: Model used
// @Description - `status`: Response status
// @Description - `input`: Input data
// @Description - `output`: Generated output
// @Tags Jan, Jan-Responses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param response_id path string true "Unique identifier of the response"
// @Success 200 {object} responses.Response "Response details"
// @Failure 400 {object} responsetypes.ErrorResponse "Invalid request"
// @Failure 401 {object} responsetypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} responsetypes.ErrorResponse "Access denied"
// @Failure 404 {object} responsetypes.ErrorResponse "Response not found"
// @Failure 500 {object} responsetypes.ErrorResponse "Internal server error"
// @Router /v1/responses/{response_id} [get]
func (responseRoute *ResponseRoute) GetResponse(reqCtx *gin.Context) {
	resp, ok := response.GetResponseFromContext(reqCtx)
	if !ok {
		return
	}
	// Convert domain response to API response using the service
	apiResponse := responseRoute.responseService.ConvertDomainResponseToAPIResponse(resp)
	reqCtx.JSON(http.StatusOK, apiResponse)
}

// DeleteResponse deletes a response by ID
// @Summary Delete a response
// @Description Deletes an LLM response by its ID. Returns the deleted response object with embedded structure where all fields are at the top level.
// @Description
// @Description **Response Format:**
// @Description The response uses embedded structure where all fields are at the top level:
// @Description - `jan_status`: Jan API status code (optional)
// @Description - `id`: Response identifier
// @Description - `object`: Object type ("response")
// @Description - `created`: Unix timestamp
// @Description - `model`: Model used
// @Description - `status`: Response status (will be "cancelled")
// @Description - `input`: Input data
// @Description - `cancelled_at`: Cancellation timestamp
// @Tags Jan, Jan-Responses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param response_id path string true "Unique identifier of the response"
// @Success 200 {object} responses.Response "Deleted response"
// @Failure 400 {object} responsetypes.ErrorResponse "Invalid request"
// @Failure 401 {object} responsetypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} responsetypes.ErrorResponse "Access denied"
// @Failure 404 {object} responsetypes.ErrorResponse "Response not found"
// @Failure 500 {object} responsetypes.ErrorResponse "Internal server error"
// @Router /v1/responses/{response_id} [delete]
func (responseRoute *ResponseRoute) DeleteResponse(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	resp, ok := response.GetResponseFromContext(reqCtx)
	if !ok {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "k1l2m3n4-o5p6-7890-klmn-123456789012",
		})
		return
	}

	success, err := responseRoute.responseService.DeleteResponse(ctx, resp.ID)
	if !success {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responsetypes.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.Error(),
		})
		return
	}
	// Convert domain response to API response using the service
	apiResponse := responseRoute.responseService.ConvertDomainResponseToAPIResponse(resp)
	reqCtx.JSON(http.StatusOK, apiResponse)
}

// CancelResponse cancels a running response
// @Summary Cancel a response
// @Description Cancels a running LLM response that was created with background=true. Only responses that are currently processing can be cancelled.
// @Description
// @Description **Response Format:**
// @Description The response uses embedded structure where all fields are at the top level:
// @Description - `jan_status`: Jan API status code (optional)
// @Description - `id`: Response identifier
// @Description - `object`: Object type ("response")
// @Description - `created`: Unix timestamp
// @Description - `model`: Model used
// @Description - `status`: Response status (will be "cancelled")
// @Description - `input`: Input data
// @Description - `cancelled_at`: Cancellation timestamp
// @Tags Jan, Jan-Responses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param response_id path string true "Unique identifier of the response to cancel"
// @Success 200 {object} responses.Response "Response cancelled successfully"
// @Failure 400 {object} responsetypes.ErrorResponse "Invalid request or response cannot be cancelled"
// @Failure 401 {object} responsetypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} responsetypes.ErrorResponse "Access denied"
// @Failure 404 {object} responsetypes.ErrorResponse "Response not found"
// @Failure 500 {object} responsetypes.ErrorResponse "Internal server error"
// @Router /v1/responses/{response_id}/cancel [post]
func (responseRoute *ResponseRoute) CancelResponse(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	resp, ok := response.GetResponseFromContext(reqCtx)
	if !ok {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "m3n4o5p6-q7r8-9012-mnop-345678901234",
		})
		return
	}

	// TODO
	// Cancel the stream if it is streaming in go routine and update response status in go routine
	success, err := responseRoute.responseService.UpdateResponseStatus(ctx, resp.ID, response.ResponseStatusCancelled)
	if !success {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.Error(),
		})
		return
	}

	// Reload the response to get updated status
	updatedResp, err := responseRoute.responseService.GetResponseByPublicID(ctx, resp.PublicID)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responsetypes.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.Error(),
		})
		return
	}
	// Convert domain response to API response using the service
	apiResponse := responseRoute.responseService.ConvertDomainResponseToAPIResponse(updatedResp)
	reqCtx.JSON(http.StatusOK, apiResponse)
}

// ListInputItems lists input items for a response
// @Summary List input items
// @Description Retrieves a paginated list of input items for a response. Supports cursor-based pagination for efficient retrieval of large datasets.
// @Description
// @Description **Response Format:**
// @Description The response uses embedded structure where all fields are at the top level:
// @Description - `jan_status`: Jan API status code (optional)
// @Description - `first_id`: First item ID for pagination (optional)
// @Description - `last_id`: Last item ID for pagination (optional)
// @Description - `has_more`: Whether more items are available (optional)
// @Description - `id`: Input item identifier
// @Description - `object`: Object type ("input_item")
// @Description - `created`: Unix timestamp
// @Description - `type`: Input type
// @Description - `text`: Text content (for text type)
// @Description - `image`: Image content (for image type)
// @Description - `file`: File content (for file type)
// @Description
// @Description **Example Response:**
// @Description ```json
// @Description {
// @Description   "jan_status": "000000",
// @Description   "first_id": "input_123",
// @Description   "last_id": "input_456",
// @Description   "has_more": false,
// @Description   "id": "input_1234567890",
// @Description   "object": "input_item",
// @Description   "created": 1234567890,
// @Description   "type": "text",
// @Description   "text": "Hello, world!"
// @Description }
// @Description ```
// @Tags Jan, Jan-Responses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param response_id path string true "Unique identifier of the response"
// @Param limit query int false "Maximum number of items to return (default: 20, max: 100)"
// @Param after query string false "Cursor for pagination - return items after this ID"
// @Param before query string false "Cursor for pagination - return items before this ID"
// @Success 200 {object} responsetypes.ListInputItemsResponse "List of input items"
// @Failure 400 {object} responsetypes.ErrorResponse "Invalid request or pagination parameters"
// @Failure 401 {object} responsetypes.ErrorResponse "Unauthorized"
// @Failure 403 {object} responsetypes.ErrorResponse "Access denied"
// @Failure 404 {object} responsetypes.ErrorResponse "Response not found"
// @Failure 500 {object} responsetypes.ErrorResponse "Internal server error"
// @Router /v1/responses/{response_id}/input_items [get]
func (responseRoute *ResponseRoute) ListInputItems(reqCtx *gin.Context) {
	ctx := reqCtx.Request.Context()
	resp, ok := response.GetResponseFromContext(reqCtx)
	if !ok {
		reqCtx.AbortWithStatusJSON(http.StatusBadRequest, responsetypes.ErrorResponse{
			Code: "p6q7r8s9-t0u1-2345-pqrs-678901234567",
		})
		return
	}

	// Get items for this response using the response service
	items, err := responseRoute.responseService.GetItemsForResponse(ctx, resp.ID, nil)
	if err != nil {
		reqCtx.AbortWithStatusJSON(http.StatusInternalServerError, responsetypes.ErrorResponse{
			Code:  err.GetCode(),
			Error: err.Error(),
		})
		return
	}

	var firstId *string
	var lastId *string
	if len(items) > 0 {
		firstId = &items[0].PublicID
		lastId = &items[len(items)-1].PublicID
	}

	// Convert conversation items to input items using the service
	inputItems := make([]responsetypes.InputItem, 0, len(items))
	for _, item := range items {
		inputItem := responseRoute.responseService.ConvertConversationItemToInputItem(item)
		inputItems = append(inputItems, inputItem)
	}

	reqCtx.JSON(http.StatusOK, responsetypes.ListInputItemsResponse{
		Object:  "list",
		Data:    inputItems,
		FirstID: firstId,
		LastID:  lastId,
		HasMore: false, // For now, we'll return all items without pagination
	})
}

// moved createResponse to response_model_handler.go

// GetResponseHandler handles the business logic for getting a response
func (h *ResponseRoute) GetResponseHandler(reqCtx *gin.Context) {
	// Get response from middleware context
	responseEntity, ok := GetResponseFromContext(reqCtx)
	if !ok {
		// Fallback: load by public ID if middleware didn't set context
		publicID := reqCtx.Param(string(response.ResponseContextKeyPublicID))
		if publicID != "" {
			if resp, err := h.responseService.GetResponseByPublicID(reqCtx, publicID); err == nil && resp != nil {
				if user, okUser := auth.GetUserFromContext(reqCtx); okUser && resp.UserID == user.ID {
					responseEntity = resp
					ok = true
				}
			}
		}
	}
	if !ok {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "response not found in context")
		return
	}

	result, err := h.getResponse(responseEntity)
	if err != nil {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, err.GetCode(), err.Error())
		return
	}

	h.sendSuccessResponse(reqCtx, result)
}

// DeleteResponseHandler handles the business logic for deleting a response
func (h *ResponseRoute) DeleteResponseHandler(reqCtx *gin.Context) {
	// Get response from middleware context
	responseEntity, ok := GetResponseFromContext(reqCtx)
	if !ok {
		publicID := reqCtx.Param(string(response.ResponseContextKeyPublicID))
		if publicID != "" {
			if resp, err := h.responseService.GetResponseByPublicID(reqCtx, publicID); err == nil && resp != nil {
				if user, okUser := auth.GetUserFromContext(reqCtx); okUser && resp.UserID == user.ID {
					responseEntity = resp
					ok = true
				}
			}
		}
	}
	if !ok {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, "b2c3d4e5-f6g7-8901-bcde-f23456789012", "response not found in context")
		return
	}

	result, err := h.deleteResponse(reqCtx, responseEntity)
	if err != nil {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, err.GetCode(), err.Error())
		return
	}

	h.sendSuccessResponse(reqCtx, result)
}

// CancelResponseHandler handles the business logic for cancelling a response
func (h *ResponseRoute) CancelResponseHandler(reqCtx *gin.Context) {
	// Get response from middleware context
	responseEntity, ok := GetResponseFromContext(reqCtx)
	if !ok {
		publicID := reqCtx.Param(string(response.ResponseContextKeyPublicID))
		if publicID != "" {
			if resp, err := h.responseService.GetResponseByPublicID(reqCtx, publicID); err == nil && resp != nil {
				if user, okUser := auth.GetUserFromContext(reqCtx); okUser && resp.UserID == user.ID {
					responseEntity = resp
					ok = true
				}
			}
		}
	}
	if !ok {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, "d4e5f6g7-h8i9-0123-defg-456789012345", "response not found in context")
		return
	}

	result, err := h.cancelResponse(responseEntity)
	if err != nil {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, err.GetCode(), err.Error())
		return
	}

	h.sendSuccessResponse(reqCtx, result)
}

// ListInputItemsHandler handles the business logic for listing input items
func (h *ResponseRoute) ListInputItemsHandler(reqCtx *gin.Context) {
	// Get response from middleware context
	responseEntity, ok := GetResponseFromContext(reqCtx)
	if !ok {
		publicID := reqCtx.Param(string(response.ResponseContextKeyPublicID))
		if publicID != "" {
			if resp, err := h.responseService.GetResponseByPublicID(reqCtx, publicID); err == nil && resp != nil {
				if user, okUser := auth.GetUserFromContext(reqCtx); okUser && resp.UserID == user.ID {
					responseEntity = resp
					ok = true
				}
			}
		}
	}
	if !ok {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, "e5f6g7h8-i9j0-1234-efgh-567890123456", "response not found in context")
		return
	}

	result, err := h.listInputItems(reqCtx, responseEntity)
	if err != nil {
		h.sendErrorResponse(reqCtx, http.StatusBadRequest, err.GetCode(), err.Error())
		return
	}

	reqCtx.JSON(http.StatusOK, result)
}

// getResponse performs the business logic for getting a response
func (h *ResponseRoute) getResponse(responseEntity *response.Response) (responsetypes.Response, *common.Error) {
	// Convert domain response to API response using domain service
	apiResponse := h.responseService.ConvertDomainResponseToAPIResponse(responseEntity)
	return apiResponse, nil
}

// deleteResponse performs the business logic for deleting a response
func (h *ResponseRoute) deleteResponse(reqCtx *gin.Context, responseEntity *response.Response) (responsetypes.Response, *common.Error) {
	// Delete the response from database
	success, err := h.responseService.DeleteResponse(reqCtx, responseEntity.ID)
	if !success {
		return responsetypes.Response{}, err
	}

	// Return the deleted response data
	deletedResponse := responsetypes.Response{
		ID:          responseEntity.PublicID,
		Object:      "response",
		Created:     responseEntity.CreatedAt.Unix(),
		Model:       responseEntity.Model,
		Status:      responsetypes.ResponseStatusCancelled,
		CancelledAt: ptr.ToInt64(time.Now().Unix()),
	}

	return deletedResponse, nil
}

// cancelResponse performs the business logic for cancelling a response
func (h *ResponseRoute) cancelResponse(responseEntity *response.Response) (responsetypes.Response, *common.Error) {
	// TODO: Implement actual cancellation logic
	// For now, return the response with cancelled status
	mockResponse := responsetypes.Response{
		ID:          responseEntity.PublicID,
		Object:      "response",
		Created:     responseEntity.CreatedAt.Unix(),
		Model:       responseEntity.Model,
		Status:      responsetypes.ResponseStatusCancelled,
		CancelledAt: ptr.ToInt64(time.Now().Unix()),
	}

	return mockResponse, nil
}

// listInputItems performs the business logic for listing input items
func (h *ResponseRoute) listInputItems(reqCtx *gin.Context, responseEntity *response.Response) (responsetypes.OpenAIListResponse[responsetypes.InputItem], *common.Error) {
	// Parse pagination parameters
	limit := 20 // default limit
	if limitStr := reqCtx.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get input items for the response (only user role messages)
	userRole := conversation.ItemRole("user")
	items, err := h.responseService.GetItemsForResponse(reqCtx, responseEntity.ID, &userRole)
	if err != nil {
		return responsetypes.OpenAIListResponse[responsetypes.InputItem]{}, err
	}

	// Convert conversation items to input items using domain service
	inputItems := make([]responsetypes.InputItem, 0, len(items))
	for _, item := range items {
		inputItem := h.responseService.ConvertConversationItemToInputItem(item)
		inputItems = append(inputItems, inputItem)
	}

	// Apply pagination (simple implementation - in production you'd want cursor-based pagination)
	after := reqCtx.Query("after")
	before := reqCtx.Query("before")

	var paginatedItems []responsetypes.InputItem
	var hasMore bool

	if after != "" {
		// Find items after the specified ID
		found := false
		for _, item := range inputItems {
			if found {
				paginatedItems = append(paginatedItems, item)
				if len(paginatedItems) >= limit {
					break
				}
			}
			if item.ID == after {
				found = true
			}
		}
	} else if before != "" {
		// Find items before the specified ID
		for _, item := range inputItems {
			if item.ID == before {
				break
			}
			paginatedItems = append(paginatedItems, item)
			if len(paginatedItems) >= limit {
				break
			}
		}
	} else {
		// No pagination, return first N items
		if len(inputItems) > limit {
			paginatedItems = inputItems[:limit]
			hasMore = true
		} else {
			paginatedItems = inputItems
		}
	}

	// Set pagination metadata
	var firstID, lastID *string
	if len(paginatedItems) > 0 {
		firstID = &paginatedItems[0].ID
		lastID = &paginatedItems[len(paginatedItems)-1].ID
	}

	status := responsetypes.ResponseCodeOk
	objectType := responsetypes.ObjectTypeList

	return responsetypes.OpenAIListResponse[responsetypes.InputItem]{
		JanStatus: &status,
		Object:    &objectType,
		HasMore:   &hasMore,
		FirstID:   firstID,
		LastID:    lastID,
		T:         paginatedItems,
	}, nil
}

// sendErrorResponse sends a standardized error response
func (h *ResponseRoute) sendErrorResponse(reqCtx *gin.Context, statusCode int, errorCode, errorMessage string) {
	reqCtx.AbortWithStatusJSON(statusCode, responsetypes.ErrorResponse{
		Code:  errorCode,
		Error: errorMessage,
	})
}

// sendSuccessResponse sends a standardized success response
func (h *ResponseRoute) sendSuccessResponse(reqCtx *gin.Context, data any) {
	reqCtx.JSON(http.StatusOK, data.(responsetypes.Response))
}

// GetResponseFromContext extracts response from gin context
func GetResponseFromContext(reqCtx *gin.Context) (*response.Response, bool) {
	responseEntity, exists := reqCtx.Get(string(response.ResponseContextEntity))
	if !exists {
		return nil, false
	}
	response, ok := responseEntity.(*response.Response)
	return response, ok
}
