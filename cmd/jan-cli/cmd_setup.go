package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var setupAndRunCmd = &cobra.Command{
	Use:   "setup-and-run",
	Short: "Interactive setup and run Jan Server",
	Long:  `Interactively configure environment variables and start Jan Server with all services.`,
	RunE:  runSetupAndRun,
}

func init() {
	setupAndRunCmd.Flags().Bool("skip-prompts", false, "Skip interactive prompts and use existing .env")
	setupAndRunCmd.Flags().Bool("with-memory-tools", false, "Enable memory tools profile and defaults during setup")
}

func runSetupAndRun(cmd *cobra.Command, args []string) error {
	skipPrompts, _ := cmd.Flags().GetBool("skip-prompts")
	enableMemory, _ := cmd.Flags().GetBool("with-memory-tools")

	fmt.Println("🚀 Jan Server Setup and Run")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Check if .env exists
	envPath := ".env"
	envExists := false
	if _, err := os.Stat(envPath); err == nil {
		envExists = true
	}

	if !skipPrompts {
		// Create or update .env file
		if envExists {
			fmt.Println("✓ Found existing .env file")
			fmt.Print("Do you want to update it? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			// Default is No for updating existing config
			if response != "y" && response != "yes" {
				fmt.Println("Using existing .env file...")
			} else {
				if err := promptForEnvVars(envPath, enableMemory); err != nil {
					return fmt.Errorf("failed to update .env: %w", err)
				}
			}
		} else {
			fmt.Println("📝 Creating .env file...")
			// Copy template
			if err := copyEnvTemplate(envPath); err != nil {
				return fmt.Errorf("failed to copy .env template: %w", err)
			}

			if err := promptForEnvVars(envPath, enableMemory); err != nil {
				return fmt.Errorf("failed to configure .env: %w", err)
			}
		}
	} else if !envExists {
		// Skip prompts but no .env exists - copy template
		fmt.Println("📝 Creating .env from template...")
		if err := copyEnvTemplate(envPath); err != nil {
			return fmt.Errorf("failed to copy .env template: %w", err)
		}
	}

	if skipPrompts && enableMemory {
		if err := applyMemoryDefaults(envPath); err != nil {
			return fmt.Errorf("failed to enable memory tools defaults: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("⚙️  Running setup...")
	fmt.Println()

	// Run dev setup
	if err := execCommand("make", "setup"); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Ask about monitoring setup
	if !skipPrompts {
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println("📊 Monitoring Stack Setup (Optional)")
		fmt.Println()
		fmt.Println("Would you like to set up the monitoring stack?")
		fmt.Println("This includes:")
		fmt.Println("  • Prometheus (metrics)")
		fmt.Println("  • Grafana (dashboards)")
		fmt.Println("  • Jaeger (distributed tracing)")
		fmt.Println("  • OpenTelemetry Collector")
		fmt.Println()
		fmt.Println()
		fmt.Print("Set up monitoring? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		monitorResponse, _ := reader.ReadString('\n')
		monitorResponse = strings.TrimSpace(strings.ToLower(monitorResponse))

		// Default is No for monitoring (optional feature)
		if monitorResponse == "y" || monitorResponse == "yes" {
			fmt.Println()
			fmt.Println("🔧 Installing monitoring dependencies...")

			// Enable tracing in .env
			if err := updateEnvVariable(envPath, "OTEL_ENABLED", "true"); err != nil {
				fmt.Println("⚠️  Warning: Failed to enable OTEL_ENABLED in .env")
			} else {
				fmt.Println("✓ Enabled telemetry collection (OTEL_ENABLED=true)")
			}

			if err := execCommand("make", "monitor-up"); err != nil {
				fmt.Println("⚠️  Warning: Failed to start monitoring stack")
				fmt.Println("You can set it up later with: jan-cli monitor setup")
			} else {
				fmt.Println("✓ Monitoring stack started successfully!")
				fmt.Println()
				fmt.Println("Access monitoring dashboards:")
				fmt.Println("  • Grafana:    http://localhost:3331 (admin/admin)")
				fmt.Println("  • Prometheus: http://localhost:9090")
				fmt.Println("  • Jaeger:     http://localhost:16686")
			}
		} else {
			fmt.Println("⏭️  Skipping monitoring setup")
			fmt.Println("You can set it up later with: jan-cli monitor setup")
		}
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("🐳 Starting Docker services...")
	fmt.Println("This may take 1-2 minutes on first run...")
	fmt.Println()

	// Start services
	if err := execCommand("make", "up-full"); err != nil {
		// Docker compose up -d returns non-zero if services are already running
		// Check if it's actually an error or just a warning
		fmt.Println()
		fmt.Println("Note: Some services may already be running")
	}

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("✅ Jan Server is starting!")
	fmt.Println()
	fmt.Println("Waiting for services to be ready (30 seconds)...")

	// Wait for services to start - cross-platform
	if isWindows() {
		execCommandSilent("powershell", "-Command", "Start-Sleep -Seconds 30")
	} else {
		execCommandSilent("sleep", "30")
	}

	fmt.Println()
	fmt.Println("Access your services:")
	fmt.Println("  • API Gateway:      http://localhost:8000")
	fmt.Println("  • API Docs:         http://localhost:8000/v1/swagger/")
	fmt.Println("  • LLM API:          http://localhost:8080")
	fmt.Println("  • Keycloak:         http://localhost:8085 (admin/admin)")

	// Only show vLLM if using local provider
	if os.Getenv("_USING_LOCAL_VLLM") == "true" {
		fmt.Println("  • vLLM (Local):     http://localhost:8101")
	}

	fmt.Println()
	fmt.Println("Get started:")
	fmt.Println("  1. Get a token:     curl -X POST http://localhost:8000/llm/auth/guest-login")
	fmt.Println("  2. Health check:    make health-check")
	fmt.Println("  3. View logs:       make logs-llm-api")
	fmt.Println("  4. Stop services:   make down")
	fmt.Println()

	return nil
}

func copyEnvTemplate(destPath string) error {
	templatePath := ".env.template"

	// Read template
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	return nil
}

func promptForEnvVars(envPath string, defaultEnableMemory bool) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("=== Configuration Wizard ===")
	fmt.Println()

	// Read current .env
	if _, err := os.ReadFile(envPath); err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	updates := make(map[string]string)

	// 1. LLM Provider Configuration
	fmt.Println("📦 LLM Provider Setup")
	fmt.Println("Choose your LLM provider:")
	fmt.Println("  1. Local vLLM (requires GPU, uses HuggingFace token)")
	fmt.Println("  2. Remote API endpoint (OpenAI-compatible)")
	fmt.Print("Enter choice [1/2] (default: 1): ")

	providerChoice, _ := reader.ReadString('\n')
	providerChoice = strings.TrimSpace(providerChoice)
	if providerChoice == "" {
		providerChoice = "1"
	}

	// Track which services to enable
	useLocalVLLM := false
	profiles := []string{"infra", "api", "mcp"} // Always include core services

	if providerChoice == "1" {
		// Local vLLM setup
		fmt.Println()
		fmt.Print("HF_TOKEN (get from https://huggingface.co/settings/tokens): ")
		hfToken, _ := reader.ReadString('\n')
		hfToken = strings.TrimSpace(hfToken)
		if hfToken != "" {
			updates["HF_TOKEN"] = hfToken
		}

		// Disable remote provider, enable local
		updates["JAN_PROVIDER_CONFIGS"] = "true"
		updates["JAN_DEFAULT_NODE_SETUP"] = "false"
		profiles = append(profiles, "full") // Add vLLM
		useLocalVLLM = true

		fmt.Println("✓ Will use local vLLM with model: Qwen/Qwen2.5-0.5B-Instruct")
	} else {
		// Remote provider setup
		fmt.Println()
		fmt.Print("Remote API URL (e.g., https://api.openai.com/v1): ")
		remoteURL, _ := reader.ReadString('\n')
		remoteURL = strings.TrimSpace(remoteURL)

		fmt.Print("API Key (press Enter if no key required): ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		if remoteURL != "" {
			updates["JAN_DEFAULT_NODE_SETUP"] = "true"
			updates["JAN_DEFAULT_NODE_URL"] = remoteURL
			updates["JAN_DEFAULT_NODE_API_KEY"] = apiKey
			updates["JAN_PROVIDER_CONFIGS"] = "false"
			updates["HF_TOKEN"] = "not_required_for_remote_provider"
			// Note: infra, api, mcp already in profiles
			fmt.Println("✓ Will use remote provider:", remoteURL)
		}
	}

	// 2. MCP Search Tool Configuration
	fmt.Println()
	fmt.Println("🔍 MCP Search Tool Setup")
	fmt.Println("Choose search provider for MCP tools:")
	fmt.Println("  1. Serper (requires API key, Google search)")
	fmt.Println("  2. SearXNG (local, no API key needed)")
	fmt.Println("  3. None (disable MCP search, but Vector Store still available)")
	fmt.Print("Enter choice [1/2/3] (default: 1): ")

	searchChoice, _ := reader.ReadString('\n')
	searchChoice = strings.TrimSpace(searchChoice)
	if searchChoice == "" {
		searchChoice = "1"
	}

	// MCP profile is already included in the profiles list

	switch searchChoice {
	case "1":
		fmt.Println()
		fmt.Print("SERPER_API_KEY (get from https://serper.dev): ")
		serperKey, _ := reader.ReadString('\n')
		serperKey = strings.TrimSpace(serperKey)
		if serperKey != "" {
			updates["SERPER_API_KEY"] = serperKey
			updates["SEARCH_ENGINE"] = "serper"
			fmt.Println("✓ Will use Serper for search")
		} else {
			fmt.Println("⚠️  No API key provided, falling back to SearXNG")
			updates["SEARCH_ENGINE"] = "searxng"
		}
	case "2":
		updates["SEARCH_ENGINE"] = "searxng"
		updates["SERPER_API_KEY"] = "not_required_for_searxng"
		fmt.Println("✓ Will use SearXNG (local) for search")
	case "3":
		updates["SEARCH_ENGINE"] = "none"
		updates["SERPER_API_KEY"] = "mcp_search_disabled"
		fmt.Println("✓ MCP search disabled (Vector Store still available)")
	}

	// 3. Memory Tools Configuration
	fmt.Println()
	fmt.Println("🧠 Memory Tools Setup")
	fmt.Println("Enable memory tools for long-term context and retrieval.")
	memoryPromptDefault := "Y/n"
	if !defaultEnableMemory {
		memoryPromptDefault = "y/N"
	}
	fmt.Printf("Enable memory tools? (%s): ", memoryPromptDefault)

	memoryChoice, _ := reader.ReadString('\n')
	memoryChoice = strings.TrimSpace(strings.ToLower(memoryChoice))

	// Default based on defaultEnableMemory flag (Y/n or y/N)
	enableMemory := defaultEnableMemory
	if memoryChoice != "" {
		enableMemory = memoryChoice != "n" && memoryChoice != "no"
	}

	externalEmbedding := false
	useRedis := false
	if enableMemory {
		externalEmbedding, useRedis = configureMemoryOptions(reader, updates)
	}
	applyMemorySettings(updates, &profiles, enableMemory, externalEmbedding, useRedis)

	// 4. Media API Configuration
	fmt.Println()
	fmt.Println("🖼️  Media API Setup")

	// Media API with local storage only works when using local vLLM
	if !useLocalVLLM {
		fmt.Println("Note: Media API with local storage requires local vLLM deployment")
		fmt.Println("      Only S3 storage is available with remote API providers")
		fmt.Print("Enable Media API with S3 storage? (y/N): ")

		mediaChoice, _ := reader.ReadString('\n')
		mediaChoice = strings.TrimSpace(strings.ToLower(mediaChoice))

		// Default is No for S3 with remote provider (requires credentials)
		if mediaChoice == "y" || mediaChoice == "yes" {
			updates["MEDIA_API_ENABLED"] = "true"
			updates["MEDIA_STORAGE_BACKEND"] = "s3"

			fmt.Println()
			fmt.Println("S3-compatible storage configuration:")
			fmt.Println("(Press Enter to use default Menlo AI settings)")

			fmt.Print("S3 Endpoint URL (default: https://s3.menlo.ai): ")
			s3Endpoint, _ := reader.ReadString('\n')
			s3Endpoint = strings.TrimSpace(s3Endpoint)
			if s3Endpoint == "" {
				s3Endpoint = "https://s3.menlo.ai"
			}
			updates["MEDIA_S3_ENDPOINT"] = s3Endpoint

			fmt.Print("S3 Bucket name (default: platform-dev): ")
			s3Bucket, _ := reader.ReadString('\n')
			s3Bucket = strings.TrimSpace(s3Bucket)
			if s3Bucket == "" {
				s3Bucket = "platform-dev"
			}
			updates["MEDIA_S3_BUCKET"] = s3Bucket

			fmt.Print("S3 Access Key ID (default: 7N33WPTUI1KN99MFILQS): ")
			s3AccessKey, _ := reader.ReadString('\n')
			s3AccessKey = strings.TrimSpace(s3AccessKey)
			if s3AccessKey == "" {
				s3AccessKey = "7N33WPTUI1KN99MFILQS"
			}
			updates["MEDIA_S3_ACCESS_KEY_ID"] = s3AccessKey

			fmt.Print("S3 Secret Access Key (default: ppxQsHpnfDSewYZD065aGjQeEQ0nTFA7c2aHNPz5): ")
			s3SecretKey, _ := reader.ReadString('\n')
			s3SecretKey = strings.TrimSpace(s3SecretKey)
			if s3SecretKey == "" {
				s3SecretKey = "ppxQsHpnfDSewYZD065aGjQeEQ0nTFA7c2aHNPz5"
			}
			updates["MEDIA_S3_SECRET_ACCESS_KEY"] = s3SecretKey

			fmt.Print("S3 Region (default: us-west-2): ")
			s3Region, _ := reader.ReadString('\n')
			s3Region = strings.TrimSpace(s3Region)
			if s3Region == "" {
				s3Region = "us-west-2"
			}
			updates["MEDIA_S3_REGION"] = s3Region

			// Set media API URLs
			updates["MEDIA_API_URL"] = "http://media-api:8285"
			updates["MEDIA_RESOLVE_URL"] = "http://media-api:8285/v1/media/resolve"

			fmt.Println("✓ Media API enabled with S3 storage")
		} else {
			updates["MEDIA_API_ENABLED"] = "false"
			fmt.Println("✓ Media API disabled")
		}
	} else {
		// Local vLLM - offer both storage options
		fmt.Print("Enable Media API? (Y/n): ")

		mediaChoice, _ := reader.ReadString('\n')
		mediaChoice = strings.TrimSpace(strings.ToLower(mediaChoice))

		// Default is Yes for Media API with local vLLM
		if mediaChoice == "n" || mediaChoice == "no" {
			updates["MEDIA_API_ENABLED"] = "false"
			fmt.Println("✓ Media API disabled")
		} else {
			updates["MEDIA_API_ENABLED"] = "true"

			// Ask for storage backend
			fmt.Println()
			fmt.Println("Choose Media storage backend:")
			fmt.Println("  1. Local file system (default, stores files locally)")
			fmt.Println("  2. S3-compatible storage (requires credentials)")
			fmt.Print("Enter choice [1/2] (default: 1): ")

			storageChoice, _ := reader.ReadString('\n')
			storageChoice = strings.TrimSpace(storageChoice)
			if storageChoice == "" {
				storageChoice = "1"
			}

			if storageChoice == "2" {
				// S3 Configuration
				updates["MEDIA_STORAGE_BACKEND"] = "s3"

				fmt.Println()
				fmt.Println("S3-compatible storage configuration:")
				fmt.Println("(Press Enter to use default Menlo AI settings)")

				fmt.Print("S3 Endpoint URL (default: https://s3.menlo.ai): ")
				s3Endpoint, _ := reader.ReadString('\n')
				s3Endpoint = strings.TrimSpace(s3Endpoint)
				if s3Endpoint == "" {
					s3Endpoint = "https://s3.menlo.ai"
				}
				updates["MEDIA_S3_ENDPOINT"] = s3Endpoint

				fmt.Print("S3 Bucket name (default: platform-dev): ")
				s3Bucket, _ := reader.ReadString('\n')
				s3Bucket = strings.TrimSpace(s3Bucket)
				if s3Bucket == "" {
					s3Bucket = "platform-dev"
				}
				updates["MEDIA_S3_BUCKET"] = s3Bucket

				fmt.Print("S3 Access Key ID (default: 7N33WPTUI1KN99MFILQS): ")
				s3AccessKey, _ := reader.ReadString('\n')
				s3AccessKey = strings.TrimSpace(s3AccessKey)
				if s3AccessKey == "" {
					s3AccessKey = "7N33WPTUI1KN99MFILQS"
				}
				updates["MEDIA_S3_ACCESS_KEY_ID"] = s3AccessKey

				fmt.Print("S3 Secret Access Key (default: ppxQsHpnfDSewYZD065aGjQeEQ0nTFA7c2aHNPz5): ")
				s3SecretKey, _ := reader.ReadString('\n')
				s3SecretKey = strings.TrimSpace(s3SecretKey)
				if s3SecretKey == "" {
					s3SecretKey = "ppxQsHpnfDSewYZD065aGjQeEQ0nTFA7c2aHNPz5"
				}
				updates["MEDIA_S3_SECRET_ACCESS_KEY"] = s3SecretKey

				fmt.Print("S3 Region (default: us-west-2): ")
				s3Region, _ := reader.ReadString('\n')
				s3Region = strings.TrimSpace(s3Region)
				if s3Region == "" {
					s3Region = "us-west-2"
				}
				updates["MEDIA_S3_REGION"] = s3Region

				fmt.Println("✓ Media API enabled with S3 storage")
			} else {
				// Local file system storage (default)
				updates["MEDIA_STORAGE_BACKEND"] = "local"
				updates["MEDIA_LOCAL_STORAGE_PATH"] = "./media-data"
				updates["MEDIA_LOCAL_STORAGE_BASE_URL"] = "http://localhost:8285/v1/files"
				fmt.Println("✓ Media API enabled with local file system storage")
			}

			// Set media API URLs (common for both backends)
			updates["MEDIA_API_URL"] = "http://media-api:8285"
			updates["MEDIA_RESOLVE_URL"] = "http://media-api:8285/v1/media/resolve"
		}
	}

	// Apply all updates
	fmt.Println()

	// Ensure Keycloak URLs are properly set for browser access
	if _, exists := updates["KEYCLOAK_PUBLIC_URL"]; !exists {
		updates["KEYCLOAK_PUBLIC_URL"] = "http://localhost:8085"
	}
	if _, exists := updates["KEYCLOAK_ADMIN_URL"]; !exists {
		updates["KEYCLOAK_ADMIN_URL"] = "http://localhost:8085"
	}
	if _, exists := updates["KEYCLOAK_BASE_URL"]; !exists {
		updates["KEYCLOAK_BASE_URL"] = "http://keycloak:8085"
	}
	if _, exists := updates["ISSUER"]; !exists {
		updates["ISSUER"] = "http://localhost:8085/realms/jan"
	}

	// Set COMPOSE_PROFILES based on enabled services
	if len(profiles) > 0 {
		updates["COMPOSE_PROFILES"] = strings.Join(profiles, ",")
	}

	// Store provider choice for return value (used later for conditional output)
	if useLocalVLLM {
		updates["_USING_LOCAL_VLLM"] = "true"
	}

	if len(updates) > 0 {
		if err := applyEnvUpdates(envPath, updates); err != nil {
			return err
		}

		fmt.Println("✓ Configuration saved to .env")
	} else {
		fmt.Println("✓ No changes made")
	}

	// Check if using local vLLM (look in updates or re-read from env)
	data, _ := os.ReadFile(envPath)
	if strings.Contains(string(data), "COMPOSE_PROFILES=full") {
		os.Setenv("_USING_LOCAL_VLLM", "true")
	}

	return nil
}

func applyEnvUpdates(envPath string, updates map[string]string) error {
	if len(updates) == 0 {
		return nil
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	pending := make(map[string]string, len(updates))
	for key, value := range updates {
		pending[key] = value
	}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}

		for key, value := range pending {
			if strings.HasPrefix(trimmed, key+"=") {
				lines[i] = fmt.Sprintf("%s=%s", key, value)
				delete(pending, key)
			}
		}
	}

	for key, value := range pending {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	return nil
}

func applyMemoryDefaults(envPath string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	profiles := parseProfiles(strings.Split(string(data), "\n"))
	updates := make(map[string]string)
	setMemoryDefaults(updates, &profiles, false, false)
	if len(profiles) > 0 {
		updates["COMPOSE_PROFILES"] = strings.Join(profiles, ",")
	}

	return applyEnvUpdates(envPath, updates)
}

func applyMemorySettings(updates map[string]string, profiles *[]string, enable bool, externalEmbedding bool, useRedis bool) {
	if enable {
		setMemoryDefaults(updates, profiles, externalEmbedding, useRedis)
		fmt.Println("Memory tools enabled (profile: memory)")
	} else {
		updates["MEMORY_TOOLS_ENABLED"] = "false"
		fmt.Println("Memory tools disabled (enable later by editing .env)")
	}
}

func setMemoryDefaults(updates map[string]string, profiles *[]string, externalEmbedding bool, useRedis bool) {
	if profiles != nil {
		hasMemory := false
		hasMock := false
		hasRedis := false
		for _, profile := range *profiles {
			if profile == "memory" {
				hasMemory = true
			}
			if profile == "memory-mock" {
				hasMock = true
			}
			if profile == "memory-redis" {
				hasRedis = true
			}
		}
		if !hasMemory {
			*profiles = append(*profiles, "memory")
		}
		if !externalEmbedding && !hasMock {
			*profiles = append(*profiles, "memory-mock")
		}
		if useRedis && !hasRedis {
			*profiles = append(*profiles, "memory-redis")
		}
	}

	if _, exists := updates["MEMORY_TOOLS_PORT"]; !exists {
		updates["MEMORY_TOOLS_PORT"] = "8090"
	}

	if !externalEmbedding && updates["EMBEDDING_SERVICE_URL"] == "" {
		updates["EMBEDDING_SERVICE_URL"] = "http://bge-m3:8091"
	}

	updates["MEMORY_TOOLS_ENABLED"] = "true"
	updates["EMBEDDING_CACHE_TYPE"] = "memory"
	updates["PROMPT_ORCHESTRATION_MEMORY"] = "true"
}

func configureMemoryOptions(reader *bufio.Reader, updates map[string]string) (bool, bool) {
	fmt.Println()
	fmt.Println("Memory Embedding Service")
	fmt.Println("Use the built-in BGE-M3 mock (default) or point to your own embedding endpoint.")
	fmt.Print("Custom embedding service URL (leave blank for http://bge-m3:8091): ")
	customURL, _ := reader.ReadString('\n')
	customURL = strings.TrimSpace(customURL)
	external := false
	if customURL != "" {
		updates["EMBEDDING_SERVICE_URL"] = customURL
		external = true
	} else if _, exists := updates["EMBEDDING_SERVICE_URL"]; !exists {
		updates["EMBEDDING_SERVICE_URL"] = "http://bge-m3:8091"
	}

	fmt.Println()
	fmt.Println("Embedding Cache")
	fmt.Println("Choose Redis for shared cache or in-memory for simplicity.")
	fmt.Print("Use Redis cache? (y/N): ")
	cacheChoice, _ := reader.ReadString('\n')
	cacheChoice = strings.TrimSpace(strings.ToLower(cacheChoice))
	// Default is No for Redis (in-memory is simpler for getting started)
	useRedis := false
	if cacheChoice == "y" || cacheChoice == "yes" {
		updates["EMBEDDING_CACHE_TYPE"] = "redis"
		fmt.Print("Redis URL (default: redis://redis-memory:6379/3): ")
		redisURL, _ := reader.ReadString('\n')
		redisURL = strings.TrimSpace(redisURL)
		if redisURL == "" {
			redisURL = "redis://redis-memory:6379/3"
		}
		updates["EMBEDDING_CACHE_REDIS_URL"] = redisURL
		useRedis = true
	} else {
		updates["EMBEDDING_CACHE_TYPE"] = "memory"
	}

	return external, useRedis
}

func updateEnvVariable(envPath, key, value string) error {
	// Read current .env
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	found := false

	// Update existing line or add new one
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, key+"=") {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	// If not found, append
	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write .env: %w", err)
	}

	return nil
}

func parseProfiles(lines []string) []string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "COMPOSE_PROFILES=") {
			value := strings.TrimPrefix(trimmed, "COMPOSE_PROFILES=")
			if value != "" {
				return strings.Split(value, ",")
			}
		}
	}
	return []string{"infra", "api", "mcp"}
}
