package testhelpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetDefaultModel fetches the first available model ID.
func GetDefaultModel(gatewayURL, accessToken string) (string, error) {
	modelsURL := strings.TrimSuffix(gatewayURL, "/") + "/v1/models"
	req, err := http.NewRequest(http.MethodGet, modelsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("get models failed: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	if len(payload.Data) == 0 {
		return "", fmt.Errorf("no models available")
	}

	return payload.Data[0].ID, nil
}

// GetModelEncoded returns URL-encoded model ID.
func GetModelEncoded(modelID string) string {
	return url.QueryEscape(modelID)
}
