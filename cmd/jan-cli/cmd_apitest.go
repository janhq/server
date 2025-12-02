package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var apiTestCmd = &cobra.Command{
	Use:   "api-test",
	Short: "Run API tests from Postman collections",
	Long: `Run API integration tests using Postman collection JSON files.

This is a lightweight cli api test that supports the essential
features needed for Jan Server testing: running collections, setting 
environment variables, and reporting results.

Examples:
  jan-cli api-test run tests/automation/auth-postman-scripts.json
  jan-cli api-test run tests/automation/auth-postman-scripts.json \
    --env-var "kong_url=http://localhost:8000" \
    --env-var "keycloak_admin=admin" \
    --verbose`,
}

var runApiTestCmd = &cobra.Command{
	Use:   "run [collection-file]",
	Short: "Run a Postman collection",
	Long:  `Execute all requests in a Postman collection file and report results.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runApiTest,
}

var (
	envVars   []string
	verbose   bool
	debug     bool
	reporters []string
	timeout   int
)

func init() {
	apiTestCmd.AddCommand(runApiTestCmd)

	runApiTestCmd.Flags().StringArrayVar(&envVars, "env-var", []string{}, "Environment variable (key=value)")
	runApiTestCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	runApiTestCmd.Flags().BoolVar(&debug, "debug", false, "Debug mode - print full request and response details")
	runApiTestCmd.Flags().StringArrayVar(&reporters, "reporters", []string{"cli"}, "Reporters to use")
	runApiTestCmd.Flags().IntVar(&timeout, "timeout-request", 30000, "Request timeout in milliseconds")
}

type PostmanCollection struct {
	Info struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	} `json:"info"`
	Item  []PostmanItem  `json:"item"`
	Event []PostmanEvent `json:"event,omitempty"`
}

type PostmanItem struct {
	Name    string          `json:"name"`
	Request *PostmanRequest `json:"request,omitempty"`
	Item    []PostmanItem   `json:"item,omitempty"`
	Event   []PostmanEvent  `json:"event,omitempty"`
}

type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header"`
	Body   *PostmanBody    `json:"body,omitempty"`
	URL    interface{}     `json:"url"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanBody struct {
	Mode       string            `json:"mode"`
	Raw        string            `json:"raw,omitempty"`
	Urlencoded []PostmanFormData `json:"urlencoded,omitempty"`
}

type PostmanFormData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanEvent struct {
	Listen string        `json:"listen"`
	Script PostmanScript `json:"script"`
}

type PostmanScript struct {
	Type string   `json:"type"`
	Exec []string `json:"exec"`
}

type TestResult struct {
	Name     string
	Passed   bool
	Duration time.Duration
	Error    string
}

func runApiTest(cmd *cobra.Command, args []string) error {
	collectionFile := args[0]

	// Parse environment variables
	envMap := make(map[string]string)
	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Load collection
	data, err := os.ReadFile(collectionFile)
	if err != nil {
		return fmt.Errorf("failed to read collection file: %w", err)
	}

	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse collection: %w", err)
	}

	fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Jan API Test Runner                                                 │\n")
	fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n\n")
	fmt.Printf("→ %s\n\n", collection.Info.Name)

	// Process collection-level prerequest scripts
	processCollectionEvents(collection.Event, envMap)

	// Run tests
	results := []TestResult{}
	totalStart := time.Now()

	for _, item := range collection.Item {
		itemResults := runItem(item, envMap, "")
		results = append(results, itemResults...)
	}

	totalDuration := time.Since(totalStart)

	// Report results
	printResults(results, totalDuration)

	// Check for failures
	for _, result := range results {
		if !result.Passed {
			return fmt.Errorf("tests failed")
		}
	}

	return nil
}

func runItem(item PostmanItem, envMap map[string]string, prefix string) []TestResult {
	results := []TestResult{}

	// If this item has nested items (folder), run them
	if len(item.Item) > 0 {
		if verbose {
			fmt.Printf("\n📁 %s\n", item.Name)
		}
		for _, subItem := range item.Item {
			subResults := runItem(subItem, envMap, prefix+"  ")
			results = append(results, subResults...)
		}
		return results
	}

	// This is a request item
	if item.Request == nil {
		return results
	}

	result := TestResult{
		Name:   item.Name,
		Passed: true,
	}

	start := time.Now()

	// Build URL
	urlStr := buildURL(item.Request.URL, envMap)

	if verbose {
		fmt.Printf("%s→ %s %s\n", prefix, item.Request.Method, urlStr)
	}

	// Create request
	var bodyReader io.Reader
	var bodyContent string
	if item.Request.Body != nil {
		if item.Request.Body.Mode == "raw" {
			body := replaceVariables(item.Request.Body.Raw, envMap)
			bodyContent = body
			bodyReader = strings.NewReader(body)
		} else if item.Request.Body.Mode == "urlencoded" {
			formData := url.Values{}
			for _, param := range item.Request.Body.Urlencoded {
				key := replaceVariables(param.Key, envMap)
				value := replaceVariables(param.Value, envMap)
				formData.Set(key, value)
			}
			bodyContent = formData.Encode()
			bodyReader = strings.NewReader(bodyContent)
		}
	}

	req, err := http.NewRequest(item.Request.Method, urlStr, bodyReader)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(start)
		results = append(results, result)
		return results
	}

	// Set headers
	for _, header := range item.Request.Header {
		value := replaceVariables(header.Value, envMap)
		req.Header.Set(header.Key, value)
	}

	// Debug: Print full request
	if debug {
		fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", prefix)
		fmt.Printf("%s🔍 REQUEST DEBUG: %s\n", prefix, item.Name)
		fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", prefix)
		fmt.Printf("%s%s %s\n", prefix, item.Request.Method, urlStr)
		fmt.Printf("%s\n%sHeaders:\n", prefix, prefix)
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Printf("%s  %s: %s\n", prefix, key, value)
			}
		}
		if bodyContent != "" {
			fmt.Printf("%s\n%sBody:\n%s%s\n", prefix, prefix, prefix, bodyContent)
		}
		fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", prefix)
	}

	// Execute request
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("Request failed: %v", err)
		result.Duration = time.Since(start)
		results = append(results, result)
		return results
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)

	result.Duration = time.Since(start)

	// Debug: Print full response
	if debug {
		fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", prefix)
		fmt.Printf("%s🔍 RESPONSE DEBUG: %s\n", prefix, item.Name)
		fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n", prefix)
		fmt.Printf("%sStatus: %d %s\n", prefix, resp.StatusCode, http.StatusText(resp.StatusCode))
		fmt.Printf("%sDuration: %dms\n", prefix, result.Duration.Milliseconds())
		fmt.Printf("%s\n%sHeaders:\n", prefix, prefix)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("%s  %s: %s\n", prefix, key, value)
			}
		}
		if len(respBody) > 0 {
			fmt.Printf("%s\n%sBody:\n%s%s\n", prefix, prefix, prefix, string(respBody))
		}
		fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n", prefix)
	}

	if verbose {
		fmt.Printf("%s  ← %d %s (%dms)\n", prefix, resp.StatusCode, http.StatusText(resp.StatusCode), result.Duration.Milliseconds())
	}

	// Simple test: check if status is 2xx or 3xx (success range)
	if resp.StatusCode >= 400 {
		result.Passed = false
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))
	} else {
		// Extract variables from test scripts if the request succeeded
		extractVariablesFromScripts(item, respBody, resp, envMap)
	}

	results = append(results, result)
	return results
}

func buildURL(urlInterface interface{}, envMap map[string]string) string {
	switch v := urlInterface.(type) {
	case string:
		return replaceVariables(v, envMap)
	case map[string]interface{}:
		// Handle Postman URL object format
		if raw, ok := v["raw"].(string); ok {
			url := replaceVariables(raw, envMap)

			// Handle path variables (e.g., :id, :public_id)
			if variables, ok := v["variable"].([]interface{}); ok {
				for _, varInterface := range variables {
					if varMap, ok := varInterface.(map[string]interface{}); ok {
						if key, ok := varMap["key"].(string); ok {
							if value, ok := varMap["value"].(string); ok {
								// Replace :key with actual value
								replacedValue := replaceVariables(value, envMap)
								url = strings.ReplaceAll(url, ":"+key, replacedValue)
							}
						}
					}
				}
			}

			return url
		}
		// Also try "url" field
		if urlStr, ok := v["url"].(string); ok {
			return replaceVariables(urlStr, envMap)
		}
	}
	return fmt.Sprintf("%v", urlInterface)
}

func replaceVariables(text string, envMap map[string]string) string {
	result := text
	for key, value := range envMap {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
		result = strings.ReplaceAll(result, "${"+key+"}", value)
	}
	return result
}

// processCollectionEvents processes collection-level prerequest scripts to initialize variables
func processCollectionEvents(events []PostmanEvent, envMap map[string]string) {
	for _, event := range events {
		if event.Listen != "prerequest" {
			continue
		}

		script := strings.Join(event.Script.Exec, "\n")
		lines := strings.Split(script, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Handle test_user_username
			if strings.Contains(line, "pm.collectionVariables.set('test_user_username'") {
				if _, exists := envMap["test_user_username"]; !exists {
					envMap["test_user_username"] = fmt.Sprintf("automation-user-%d", time.Now().UnixNano())
				}
			}

			// Handle test_user_password
			if strings.Contains(line, "pm.collectionVariables.set('test_user_password'") {
				if _, exists := envMap["test_user_password"]; !exists {
					envMap["test_user_password"] = fmt.Sprintf("Passw0rd!%d", time.Now().UnixNano()%10000)
				}
			}

			// Handle test_user_email
			if strings.Contains(line, "pm.collectionVariables.set('test_user_email'") {
				if _, exists := envMap["test_user_email"]; !exists {
					if username, ok := envMap["test_user_username"]; ok {
						envMap["test_user_email"] = username + "@example.com"
					}
				}
			}

			// Handle test_user_pid
			if strings.Contains(line, "pm.collectionVariables.set('test_user_pid'") {
				if _, exists := envMap["test_user_pid"]; !exists {
					if username, ok := envMap["test_user_username"]; ok {
						envMap["test_user_pid"] = username
					}
				}
			}

			// Handle collection_timestamp
			if strings.Contains(line, "pm.collectionVariables.set('collection_timestamp'") {
				if _, exists := envMap["collection_timestamp"]; !exists {
					envMap["collection_timestamp"] = time.Now().Format(time.RFC3339)
				}
			}
		}
	}
}

// extractVariablesFromScripts parses test scripts and extracts variables
func extractVariablesFromScripts(item PostmanItem, respBody []byte, resp *http.Response, envMap map[string]string) {
	if len(item.Event) == 0 {
		return
	}

	// Parse response body as JSON if possible
	var responseData map[string]interface{}
	json.Unmarshal(respBody, &responseData) // Ignore error - not all responses are JSON

	// Process each event script
	for _, event := range item.Event {
		if event.Listen != "test" {
			continue
		}

		// Join script lines
		script := strings.Join(event.Script.Exec, "\n")

		// Check for Location header extraction in the script
		locationExtracted := false
		if strings.Contains(script, "pm.response.headers.get('Location')") &&
			strings.Contains(script, "pm.collectionVariables.set('test_user_id'") {
			// Extract user ID from Location header
			if resp.StatusCode == 201 || resp.StatusCode == 204 {
				location := resp.Header.Get("Location")
				if location != "" {
					// Extract ID from location (last segment of path)
					lastSlash := strings.LastIndex(location, "/")
					if lastSlash >= 0 && lastSlash < len(location)-1 {
						userID := location[lastSlash+1:]
						envMap["test_user_id"] = userID
						envMap["teardown_user_id"] = userID
						locationExtracted = true
					}
				}
			}
		}

		// Simple pattern matching for pm.collectionVariables.set calls
		// Pattern: pm.collectionVariables.set('varname', data.field)
		lines := strings.Split(script, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Look for pm.collectionVariables.set
			if strings.Contains(line, "pm.collectionVariables.set") {
				// Extract variable name and source field
				// Example: pm.collectionVariables.set('kc_admin_access_token', data.access_token);
				varName, jsonPath := extractVarSetPattern(line)

				// Skip if we already extracted this variable from Location header
				if varName == "test_user_id" && locationExtracted {
					continue
				}

				if varName != "" && jsonPath != "" {
					// Extract value from response data
					if value := extractJSONValue(responseData, jsonPath); value != "" {
						envMap[varName] = value
					}
				}
			}
		}
	}
}

// extractVarSetPattern extracts variable name and JSON path from pm.collectionVariables.set line
func extractVarSetPattern(line string) (varName string, jsonPath string) {
	// Remove semicolons and clean up
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSpace(line)

	// Find the pattern: pm.collectionVariables.set('varname', source)
	if idx := strings.Index(line, "pm.collectionVariables.set("); idx >= 0 {
		// Extract the arguments
		argsStart := idx + len("pm.collectionVariables.set(")
		argsEnd := strings.LastIndex(line, ")")
		if argsEnd > argsStart {
			args := line[argsStart:argsEnd]
			// Split by comma
			parts := strings.SplitN(args, ",", 2)
			if len(parts) == 2 {
				// Extract variable name (remove quotes)
				varName = strings.Trim(strings.TrimSpace(parts[0]), "'\"")
				// Extract JSON path (e.g., "data.access_token" or "body.id")
				jsonPath = strings.TrimSpace(parts[1])
				// Remove common prefixes
				jsonPath = strings.TrimPrefix(jsonPath, "data.")
				jsonPath = strings.TrimPrefix(jsonPath, "body.")
				jsonPath = strings.TrimPrefix(jsonPath, "responseData.")
				jsonPath = strings.Trim(jsonPath, "'\"")
			}
		}
	}
	return
}

// extractJSONValue extracts a value from JSON response using dot notation
func extractJSONValue(data map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return ""
		}
	}

	// Convert to string
	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func printResults(results []TestResult, totalDuration time.Duration) {
	passed := 0
	failed := 0

	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ Test Results                                                        │\n")
	fmt.Printf("└─────────────────────────────────────────────────────────────────────┘\n\n")

	// Print all test results with visual indicators
	for _, result := range results {
		if result.Passed {
			fmt.Printf("  ✓✓✓ %s (%dms)\n", result.Name, result.Duration.Milliseconds())
		} else {
			fmt.Printf("  ✗✗✗ %s (%dms)\n", result.Name, result.Duration.Milliseconds())
			if result.Error != "" {
				fmt.Printf("      %s\n", result.Error)
			}
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total:    %d tests\n", len(results))
	fmt.Printf("  Passed:   %d ✓✓✓\n", passed)
	if failed > 0 {
		fmt.Printf("  Failed:   %d ✗✗✗\n", failed)
	}
	fmt.Printf("  Duration: %dms\n\n", totalDuration.Milliseconds())

	if failed == 0 {
		fmt.Printf("✓✓✓ All tests passed!\n\n")
	} else {
		fmt.Printf("✗✗✗ Some tests failed\n\n")
	}
}
