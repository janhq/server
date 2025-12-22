# Test Suite Refactor TODO (jan-cli api-test; jan-server/tests/automation)

> **Status: FINALIZED** (2024-12-22)
> 
> This plan has been reviewed and approved. Ready for implementation.

---

## Implementation Plan

### Task 1: Create `pkg/testhelpers/` Package
**File:** `pkg/testhelpers/auth.go`
```go
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

// GuestLogin performs guest login and returns access token
func GuestLogin(gatewayURL string) (string, error) {
	loginURL := strings.TrimSuffix(gatewayURL, "/") + "/auth/guest-login"
	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader("{}"))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("guest login failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	token, _ := payload["access_token"].(string)
	return token, nil
}

// AdminLogin performs Keycloak admin login
func AdminLogin(keycloakURL, username, password string) (string, error) {
	tokenURL := strings.TrimSuffix(keycloakURL, "/") + "/realms/master/protocol/openid-connect/token"

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", "admin-cli")
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("admin login failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var payload map[string]interface{}
	json.Unmarshal(body, &payload)

	token, _ := payload["access_token"].(string)
	return token, nil
}
```

**File:** `pkg/testhelpers/models.go`
```go
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

// GetDefaultModel fetches the first available model ID
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

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("get models failed: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
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

// GetModelEncoded returns URL-encoded model ID
func GetModelEncoded(modelID string) string {
	return url.QueryEscape(modelID)
}
```

**File:** `pkg/testhelpers/health.go`
```go
package testhelpers

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WaitForHealth waits for service to be healthy
func WaitForHealth(gatewayURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := CheckHealth(gatewayURL); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("health check timeout after %v", timeout)
}

// CheckHealth performs a single health check
func CheckHealth(gatewayURL string) error {
	healthURL := strings.TrimSuffix(gatewayURL, "/") + "/healthz"
	resp, err := http.Get(healthURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unhealthy: %d", resp.StatusCode)
	}
	return nil
}
```

---

### Task 2: Update `cmd/jan-cli/cmd_apitest.go`

**Add new flag variables (after existing vars):**
```go
var (
	envVars    []string
	verbose    bool
	debug      bool
	reporters  []string
	timeout    int
	// New flags
	autoAuth   string
	autoModels bool
	envFile    string
	folder     string
	bail       bool
)
```

**Register flags in `init()`:**
```go
func init() {
	apiTestCmd.AddCommand(runApiTestCmd)

	runApiTestCmd.Flags().StringArrayVar(&envVars, "env-var", []string{}, "Environment variable (key=value)")
	runApiTestCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	runApiTestCmd.Flags().BoolVar(&debug, "debug", false, "Debug mode")
	runApiTestCmd.Flags().StringArrayVar(&reporters, "reporters", []string{"cli"}, "Reporters to use")
	runApiTestCmd.Flags().IntVar(&timeout, "timeout-request", 30000, "Request timeout in milliseconds")
	// New flags
	runApiTestCmd.Flags().StringVar(&autoAuth, "auto-auth", "", "Auto-login: 'guest' or 'admin'")
	runApiTestCmd.Flags().BoolVar(&autoModels, "auto-models", false, "Auto-fetch model IDs before running")
	runApiTestCmd.Flags().StringVar(&envFile, "env-file", "", "Load environment variables from file")
	runApiTestCmd.Flags().StringVar(&folder, "folder", "", "Run only requests in this folder")
	runApiTestCmd.Flags().BoolVar(&bail, "bail", false, "Stop on first failure")
}
```

**Add env-file loading in `runApiTest()` (after parsing --env-var):**
```go
// Load env file if specified
if envFile != "" {
	if err := loadEnvFile(envFile, envMap); err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}
}
```

**Add auto-auth before running tests (after loading collection):**
```go
// Auto-auth if requested
if autoAuth != "" {
	gatewayURL := envMap["gateway_url"]
	if gatewayURL == "" {
		gatewayURL = envMap["kong_url"]
	}
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8000"
	}

	if autoAuth == "admin" || autoAuth == "both" {
		kcURL := firstNonEmpty(envMap["keycloak_base_url"], "http://localhost:8085")
		kcUser := firstNonEmpty(envMap["keycloak_admin"], os.Getenv("KEYCLOAK_ADMIN"))
		kcPass := firstNonEmpty(envMap["keycloak_admin_password"], os.Getenv("KEYCLOAK_ADMIN_PASSWORD"))
		if token, err := testhelpers.AdminLogin(kcURL, kcUser, kcPass); err == nil {
			envMap["kc_admin_access_token"] = token
			envMap["access_token"] = token
		}
	}

	if token, err := testhelpers.GuestLogin(gatewayURL); err == nil {
		envMap["guest_access_token"] = token
		if envMap["access_token"] == "" {
			envMap["access_token"] = token
		}
	}
}

// Auto-models if requested
if autoModels && envMap["access_token"] != "" {
	gatewayURL := firstNonEmpty(envMap["gateway_url"], envMap["kong_url"], "http://localhost:8000")
	if model, err := testhelpers.GetDefaultModel(gatewayURL, envMap["access_token"]); err == nil {
		envMap["model_id"] = model
		envMap["model_id_encoded"] = testhelpers.GetModelEncoded(model)
		envMap["default_model_id"] = model
	}
}
```

**Add helper function for env file loading:**
```go
func loadEnvFile(path string, envMap map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Try JSON first
	var jsonEnv map[string]string
	if err := json.Unmarshal(data, &jsonEnv); err == nil {
		for k, v := range jsonEnv {
			envMap[k] = v
		}
		return nil
	}

	// Fall back to .env format
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			envMap[key] = value
		}
	}
	return nil
}
```

**Modify `runItem()` for folder filtering and bail:**
```go
func runItem(item PostmanItem, envMap map[string]string, prefix string, parentAuth *PostmanAuth) []TestResult {
	// ... existing code ...

	// If this item has nested items (folder), run them
	if len(item.Item) > 0 {
		// Folder filtering
		if folder != "" && item.Name != folder && prefix == "" {
			// Skip top-level folders that don't match
			if verbose {
				fmt.Printf("📁 Skipping %s (--folder=%s)\n", item.Name, folder)
			}
			return results
		}
		// ... rest of existing folder handling ...
	}

	// ... existing request handling ...

	// Bail on failure
	if bail && !result.Passed {
		fmt.Printf("\n❌ Bail: stopping after first failure\n")
		os.Exit(1)
	}

	results = append(results, result)
	return results
}
```

---

### Task 3: Create Environment File
**File:** `tests/automation/.env`
```bash
# Gateway URLs
gateway_url=http://localhost:8000
kong_url=http://localhost:8000

# Service URLs (overrides)
memory_url=http://localhost:8090
media_url=http://localhost:8000/media
mcp_url=http://localhost:8091
embedding_url=http://localhost:8091

# Keycloak (for admin auth)
keycloak_base_url=http://localhost:8085
keycloak_admin=admin
keycloak_admin_password=admin
```

---

### Task 4: Update Makefile
**File:** `Makefile` (add to existing targets)
```makefile
# =============================================================================
# API Test Targets
# =============================================================================

GATEWAY_URL ?= http://localhost:8000
TIMEOUT_MS ?= 30000
COLLECTIONS_DIR := tests/automation/collections
AUTH_MODE ?= guest

API_TEST_FLAGS := --env-file tests/automation/.env \
	--env-var gateway_url=$(GATEWAY_URL) \
	--auto-auth $(AUTH_MODE) \
	--auto-models \
	--timeout-request $(TIMEOUT_MS)

.PHONY: test-all test-auth test-conversation test-response test-model test-memory test-media test-mcp test-dev

test-all:
	jan-cli api-test run $(COLLECTIONS_DIR)/*.postman.json $(API_TEST_FLAGS)

test-auth:
	jan-cli api-test run $(COLLECTIONS_DIR)/auth.postman.json $(API_TEST_FLAGS)

test-conversation:
	jan-cli api-test run $(COLLECTIONS_DIR)/conversation.postman.json $(API_TEST_FLAGS)

test-response:
	jan-cli api-test run $(COLLECTIONS_DIR)/response.postman.json $(API_TEST_FLAGS)

test-model:
	jan-cli api-test run $(COLLECTIONS_DIR)/model.postman.json $(API_TEST_FLAGS) --auto-auth admin

test-memory:
	jan-cli api-test run $(COLLECTIONS_DIR)/memory.postman.json $(API_TEST_FLAGS)

test-media:
	jan-cli api-test run $(COLLECTIONS_DIR)/media.postman.json $(API_TEST_FLAGS)

test-mcp:
	jan-cli api-test run $(COLLECTIONS_DIR)/mcp-runtime.postman.json $(COLLECTIONS_DIR)/mcp-admin.postman.json $(API_TEST_FLAGS)

test-dev:
	jan-cli api-test run $(COLLECTIONS_DIR)/*.postman.json $(API_TEST_FLAGS) --bail
```

---

## jan-cli api-test Improvements

Current limitations in `cmd_apitest.go` that force workarounds in Postman JSON:

| Issue | Current Behavior | Proposed Fix |
|-------|------------------|--------------|
| No `--folder` flag | Must run entire collection | Add `--folder <name>` to run specific folders |
| No `--include`/`--exclude` | Can't filter by request name | Add glob/regex filters for request names |
| No explicit `--auth-token` | Must rely on auto-extraction or bootstrap | Add `--auth-token <token>` to inject directly |
| No shared environment file | Variables passed via `--env-var` flags | Add `--env-file <path>` to load `.env` or JSON |
| No `--bail` option | Runs all tests even on failure | Add `--bail` to stop on first failure |
| No JSON output | CLI output only | Add `--output-format json` for structured results |
| No built-in auth helpers | Auth bootstrap in every collection | Add `--auto-auth` to handle guest/admin login |
| No model discovery | Must hardcode or extract model IDs | Add `--auto-models` to fetch available models |

### Priority Improvements

**1. Add `--auto-auth` flag** (highest impact)
```bash
jan-cli api-test run collection.json --auto-auth
```
- Automatically perform guest login before running tests
- Sets `guest_access_token`, `access_token` automatically
- With `--auto-auth=admin`, also fetches admin token
- Eliminates need for `00-auth-bootstrap` folder in collections

**2. Add `--auto-models` flag** (high impact)
```bash
jan-cli api-test run collection.json --auto-models
```
- Fetch available models from `/v1/models` before running tests
- Sets `model_id`, `model_id_encoded`, `default_model_id` automatically
- Eliminates model discovery boilerplate in collections

**3. Add `--folder` flag** (high impact)
```bash
jan-cli api-test run collection.json --folder "02-operations"
```
- Enables running subsets without maintaining separate collections
- Reduces collection complexity significantly

**4. Add `--env-file` support** (high impact)
```bash
jan-cli api-test run collection.json --env-file tests/automation/.env
```
- Load variables from file instead of many `--env-var` flags
- Support both `.env` and JSON formats

**5. Add `--bail` flag** (medium impact)
```bash
jan-cli api-test run collection.json --bail
```
- Stop execution on first failure
- Faster feedback during development

**6. Add `--output-format json`** (low impact, nice-to-have)
```bash
jan-cli api-test run collection.json --output-format json > results.json
```
- Machine-readable output for CI integration
- Include request/response details on failure

### Built-in Helpers (new `pkg/testhelpers` package)

Create reusable helper functions that jan-cli can call internally:

```go
// pkg/testhelpers/auth.go
package testhelpers

// GuestLogin performs guest login and returns tokens
func GuestLogin(gatewayURL string) (accessToken string, err error)

// AdminLogin performs admin login via Keycloak
func AdminLogin(keycloakURL, username, password string) (accessToken string, err error)

// RefreshToken refreshes an existing token
func RefreshToken(gatewayURL, refreshToken string) (accessToken string, err error)
```

```go
// pkg/testhelpers/models.go
package testhelpers

// GetDefaultModel fetches the first available model
func GetDefaultModel(gatewayURL, accessToken string) (modelID string, err error)

// GetModels fetches all available models
func GetModels(gatewayURL, accessToken string) ([]Model, error)

// GetModelEncoded returns URL-encoded model ID
func GetModelEncoded(modelID string) string
```

```go
// pkg/testhelpers/health.go
package testhelpers

// WaitForHealth waits for service to be healthy
func WaitForHealth(gatewayURL string, timeout time.Duration) error

// CheckHealth performs a single health check
func CheckHealth(gatewayURL string) error
```

### Code Changes Required

In `cmd/jan-cli/cmd_apitest.go`:

```go
// Add new flags in init()
runApiTestCmd.Flags().StringVar(&folderFilter, "folder", "", "Run only requests in this folder")
runApiTestCmd.Flags().StringVar(&envFile, "env-file", "", "Load environment from file")
runApiTestCmd.Flags().StringVar(&autoAuth, "auto-auth", "", "Auto-login: 'guest' or 'admin'")
runApiTestCmd.Flags().BoolVar(&autoModels, "auto-models", false, "Auto-fetch model IDs")
runApiTestCmd.Flags().BoolVar(&bail, "bail", false, "Stop on first failure")
runApiTestCmd.Flags().StringVar(&outputFormat, "output-format", "cli", "Output format: cli or json")

// In runApiTest(), before running collection:
if autoAuth != "" {
    if autoAuth == "admin" {
        token, _ := testhelpers.AdminLogin(...)
        envMap["kc_admin_access_token"] = token
    }
    token, _ := testhelpers.GuestLogin(envMap["gateway_url"])
    envMap["guest_access_token"] = token
    envMap["access_token"] = token
}

if autoModels {
    model, _ := testhelpers.GetDefaultModel(envMap["gateway_url"], envMap["access_token"])
    envMap["model_id"] = model
    envMap["model_id_encoded"] = testhelpers.GetModelEncoded(model)
    envMap["default_model_id"] = model
}
```

---

## jan-cli api-test Constraints

- Runner parses collections directly and only pattern-matches a subset of Postman scripts (e.g., `pm.collectionVariables.set`, `pm.response.to.have.status`, `pm.expect(pm.response.code).to.eql`). No external JS libs, no `pm.sendRequest`, no helper imports.
- Auth tokens are auto-hoisted into `access_token` from `kc_admin_access_token`, `guest_access_token`, or `llm_api_token`; keep those names consistent.
- Allowed status codes come from inline expectations; everything else treats `<400` as pass.
- URL/body templating uses `{{var}}` and `${var}` from `--env-var` plus extracted variables.

## Current Pain Points

- Collections repeat guest/admin login, health checks, and bearer headers instead of relying on the runner's implicit `access_token` handling.
- Base URLs and variable names drift (`kong_url`, `base_url`, `embedding_url`, `mcp_tools_url`), making reuse brittle.
- Some scripts use helper-style JS the runner ignores, so assertions silently drop.

## Refactor Goals

1. **Single collection per domain** – no duplication, no split between "smoke" and "deep".
2. **Built-in auth via `--auto-auth`** – no auth bootstrap folders needed in collections.
3. **Built-in model discovery via `--auto-models`** – no model fetch requests needed.
4. **Consistent variable naming** – canonical names with backward-compatible mappings.
5. **Runner-compatible assertions only** – no unsupported JS patterns.

## Collection Layout

```
tests/automation/
├── .env                        # canonical variables (gateway_url, etc.)
├── collections/
│   ├── auth.postman.json       # user/API-key CRUD, token refresh
│   ├── conversation.postman.json   # create, append, list, branching, metadata, rating
│   ├── response.postman.json   # response generation flows
│   ├── model.postman.json      # providers, provider_models, catalogs, prompt templates
│   ├── memory.postman.json     # store, retrieve, search
│   ├── media.postman.json      # upload, download, metadata
│   ├── mcp-runtime.postman.json    # tool invocation, runtime flows
│   └── mcp-admin.postman.json  # tool registration, admin operations
└── README.md                   # usage documentation
```

Each collection contains only actual test operations—no auth bootstrap or model discovery (handled by `--auto-auth` and `--auto-models`).

## Folder Structure per Collection

1. `01-operations` – main test operations (create, read, update, delete).
2. `02-edge-cases` – negative tests, error handling.
3. `99-cleanup` (optional) – delete created entities.

Entities are created in operations and reused; IDs captured via `pm.collectionVariables.set()`.

## Variable Canon

| Canonical Name | Replaces | Purpose |
|----------------|----------|---------|
| `gateway_url` | `kong_url`, `base_url` | Primary API gateway |
| `mcp_url` | `mcp_tools_url` | MCP service override |
| `memory_url` | — | Memory service override |
| `media_url` | — | Media service override |
| `embedding_url` | — | Embedding service override |

During migration, pass both old and new names via `--env-var` for backward compatibility.

## Auth Header Pattern

- Collection-level bearer: `"token": "{{access_token}}"`.
- Override per-request only for negative test cases (401/403).

## Assertion Patterns (Runner-Compatible)

```javascript
// Status code assertions
pm.response.to.have.status(200);
pm.expect(pm.response.code).to.eql(201);

// Variable extraction
const data = pm.response.json();
pm.collectionVariables.set('conversation_id', data.id);
pm.collectionVariables.set('model_id_encoded', encodeURIComponent(data.id));
```

## Negative Test Patterns

| Scenario | Pattern |
|----------|---------|
| 401/403 Unauthorized | Override auth: `"auth": { "type": "noauth" }` |
| 404 Not Found | `pm.response.to.have.status(404)` |
| 400/422 Validation | Assert expected error structure and message |

## Folder Structure per Collection

1. `00-auth-bootstrap` – health + login, sets tokens.
2. `01-setup` – create entities (conversation, upload, etc.), capture IDs.
3. `02-operations` – main test operations using captured IDs.
4. `03-edge-cases` – negative tests, error handling.
5. `99-cleanup` (optional) – delete created entities.

Entities are created once and reused; no duplication within a collection.

## Quick Start

```bash
# Run all tests against local gateway (auto-handles auth + models)
make test-all

# Run specific domain
make test-conversation

# Run with custom gateway URL
make test-all GATEWAY_URL=http://localhost:8001

# Run with admin auth (for admin-only endpoints)
make test-auth AUTH_MODE=admin
```

## Makefile Targets

```makefile
# Defaults
GATEWAY_URL ?= http://localhost:8000
TIMEOUT_MS ?= 30000
COLLECTIONS_DIR := tests/automation/collections
AUTH_MODE ?= guest

# Common flags - auto-auth and auto-models eliminate bootstrap boilerplate
API_TEST_FLAGS := --env-file tests/automation/.env \
	--env-var gateway_url=$(GATEWAY_URL) \
	--auto-auth $(AUTH_MODE) \
	--auto-models \
	--timeout-request $(TIMEOUT_MS)

# Run all tests
test-all:
	jan-cli api-test run $(COLLECTIONS_DIR)/*.postman.json $(API_TEST_FLAGS)

# Run single domain
test-auth:
	jan-cli api-test run $(COLLECTIONS_DIR)/auth.postman.json $(API_TEST_FLAGS)

test-conversation:
	jan-cli api-test run $(COLLECTIONS_DIR)/conversation.postman.json $(API_TEST_FLAGS)

test-response:
	jan-cli api-test run $(COLLECTIONS_DIR)/response.postman.json $(API_TEST_FLAGS)

test-model:
	jan-cli api-test run $(COLLECTIONS_DIR)/model.postman.json $(API_TEST_FLAGS) --auto-auth admin

test-memory:
	jan-cli api-test run $(COLLECTIONS_DIR)/memory.postman.json $(API_TEST_FLAGS)

test-media:
	jan-cli api-test run $(COLLECTIONS_DIR)/media.postman.json $(API_TEST_FLAGS)

test-mcp:
	jan-cli api-test run $(COLLECTIONS_DIR)/mcp-runtime.postman.json \
		$(COLLECTIONS_DIR)/mcp-admin.postman.json $(API_TEST_FLAGS)

# Run specific folder within a collection
test-conversation-operations:
	jan-cli api-test run $(COLLECTIONS_DIR)/conversation.postman.json \
		--folder "01-operations" $(API_TEST_FLAGS)

# Fail fast mode for development
test-dev:
	jan-cli api-test run $(COLLECTIONS_DIR)/*.postman.json $(API_TEST_FLAGS) --bail
```

## Migration Steps

### Phase 1: jan-cli Improvements (enables simpler collections)
1. Create `pkg/testhelpers/` package with reusable auth, models, and health functions.
2. Add `--auto-auth` flag that uses `testhelpers.GuestLogin()` / `testhelpers.AdminLogin()`.
3. Add `--auto-models` flag that uses `testhelpers.GetDefaultModel()`.
4. Add `--env-file` support to load variables from file.
5. Add `--folder` flag to run specific folders within a collection.
6. Add `--bail` flag to stop on first failure.

### Phase 2: Collection Refactor (much simpler with new flags)
7. Create `tests/automation/.env` with canonical variable names.
8. Remove `00-auth-bootstrap` folders from all collections (handled by `--auto-auth`).
9. Remove model discovery requests (handled by `--auto-models`).
10. Consolidate each domain into a single collection with only actual test operations.
11. Rewrite assertions to runner-recognized patterns; remove helper-style JS.
12. Update Makefile targets to use new flags.
13. Add `tests/automation/README.md` documenting collection map and jan-cli invocations.
14. Run validation checklist and fix any gaps.

## Post-Migration Validation Checklist

### jan-cli Improvements
- [ ] `pkg/testhelpers/` package exists with `auth.go`, `models.go`, `health.go`.
- [ ] `jan-cli api-test run --help` shows `--auto-auth`, `--auto-models`, `--folder`, `--env-file`, `--bail` flags.
- [ ] `--auto-auth guest` performs guest login and sets tokens.
- [ ] `--auto-auth admin` performs admin login via Keycloak.
- [ ] `--auto-models` fetches and sets `model_id`, `model_id_encoded`, `default_model_id`.
- [ ] `--env-file` loads variables from file correctly.
- [ ] `--folder` runs only the specified folder.
- [ ] `--bail` stops on first failure.

### Collection Quality
- [ ] `make test-all` completes successfully.
- [ ] Each domain collection runs independently (`make test-<domain>` works in isolation).
- [ ] No auth bootstrap folders in collections (handled by `--auto-auth`).
- [ ] No model discovery requests in collections (handled by `--auto-models`).
- [ ] No `pm.sendRequest` or external imports in any collection.
- [ ] All assertions use runner-detectable patterns.
- [ ] Variable names follow canonical naming.
- [ ] `tests/automation/README.md` documents all targets and usage.

---

## Quick Reference: Files to Create/Modify

| Task | File | Action |
|------|------|--------|
| 1 | `pkg/testhelpers/auth.go` | Create |
| 1 | `pkg/testhelpers/models.go` | Create |
| 1 | `pkg/testhelpers/health.go` | Create |
| 2 | `cmd/jan-cli/cmd_apitest.go` | Modify |
| 3 | `tests/automation/.env` | Create |
| 4 | `Makefile` | Add targets |
| 5 | `tests/automation/collections/` | Refactor existing |
| 6 | `tests/automation/README.md` | Create |

## Execution Order

```
Phase 1: jan-cli improvements (do first)
├── [ ] 1. Create pkg/testhelpers/ package
├── [ ] 2. Update cmd/jan-cli/cmd_apitest.go with new flags
├── [ ] 3. Build and test: go build -o jan-cli ./cmd/jan-cli
└── [ ] 4. Verify: jan-cli api-test run --help

Phase 2: Test infrastructure
├── [ ] 5. Create tests/automation/.env
├── [ ] 6. Add Makefile targets
└── [ ] 7. Test: make test-auth (with existing collection)

Phase 3: Collection refactor
├── [ ] 8. Remove auth bootstrap folders from collections
├── [ ] 9. Remove model discovery requests
├── [ ] 10. Consolidate and clean up assertions
├── [ ] 11. Create tests/automation/README.md
└── [ ] 12. Run validation checklist
```
