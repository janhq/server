package testhelpers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WaitForHealth waits for service to be healthy.
func WaitForHealth(gatewayURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := CheckHealth(gatewayURL); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("health check timeout after %v", timeout)
}

// CheckHealth performs a single health check.
func CheckHealth(gatewayURL string) error {
	healthURL := strings.TrimSuffix(gatewayURL, "/") + "/healthz"
	resp, err := http.Get(healthURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unhealthy: %d", resp.StatusCode)
	}
	return nil
}
