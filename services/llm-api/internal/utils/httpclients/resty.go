package httpclients

import (
	"context"
	"time"

	"jan-server/services/llm-api/internal/infrastructure/logger"

	"resty.dev/v3"
)

type RequestId struct{}
type HttpClientStartsAt struct{}
type HttpClientRequestBody struct{}

func NewClient(clientName string) *resty.Client {
	client := resty.New()
	client.AddRequestMiddleware(func(c *resty.Client, r *resty.Request) error {
		start := time.Now()
		ctx := context.WithValue(r.Context(), HttpClientStartsAt{}, start)
		ctx = context.WithValue(ctx, HttpClientRequestBody{}, r.Body)
		r.SetContext(ctx)
		return nil
	})
	client.AddResponseMiddleware(func(c *resty.Client, r *resty.Response) error {
		log := logger.GetLogger()
		requestID := r.Request.Context().Value(RequestId{})
		startTime, _ := r.Request.Context().Value(HttpClientStartsAt{}).(time.Time)
		requestBody := r.Request.Context().Value(HttpClientRequestBody{})
		latency := time.Since(startTime)
		var responseBody any
		if !r.Request.DoNotParseResponse {
			responseBody = r.Result()
		}

		requestIDStr := ""
		if reqID, ok := requestID.(string); ok {
			requestIDStr = reqID
		}

		log.Info().
			Str("request_id", requestIDStr).
			Str("client", clientName).
			Int("status", r.StatusCode()).
			Str("method", r.Request.RawRequest.Method).
			Str("path", r.Request.RawRequest.URL.Path).
			Str("query", r.Request.RawRequest.URL.RawQuery).
			Interface("req_body", requestBody).
			Interface("resp_body", responseBody).
			Dur("latency", latency).
			Msg("HTTP client request")
		return nil
	})
	return client
}
