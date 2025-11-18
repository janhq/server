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
}

func runSetupAndRun(cmd *cobra.Command, args []string) error {
	skipPrompts, _ := cmd.Flags().GetBool("skip-prompts")

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

			if response != "y" && response != "yes" {
				fmt.Println("Using existing .env file...")
			} else {
				if err := promptForEnvVars(envPath); err != nil {
					return fmt.Errorf("failed to update .env: %w", err)
				}
			}
		} else {
			fmt.Println("📝 Creating .env file...")
			// Copy template
			if err := copyEnvTemplate(envPath); err != nil {
				return fmt.Errorf("failed to copy .env template: %w", err)
			}

			if err := promptForEnvVars(envPath); err != nil {
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

	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("⚙️  Running setup...")
	fmt.Println()

	// Run dev setup
	if err := execCommand("make", "setup"); err != nil {
		return fmt.Errorf("setup failed: %w", err)
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

func promptForEnvVars(envPath string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("=== Configuration Wizard ===")
	fmt.Println()

	// Read current .env
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
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

	// 3. Media API Configuration
	fmt.Println()
	fmt.Println("🖼️  Media API Setup")

	// Media API with local storage only works when using local vLLM
	if !useLocalVLLM {
		fmt.Println("Note: Media API with local storage requires local vLLM deployment")
		fmt.Println("      Only S3 storage is available with remote API providers")
		fmt.Print("Enable Media API with S3 storage? (y/N): ")

		mediaChoice, _ := reader.ReadString('\n')
		mediaChoice = strings.TrimSpace(strings.ToLower(mediaChoice))

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

	// Set COMPOSE_PROFILES based on enabled services
	if len(profiles) > 0 {
		updates["COMPOSE_PROFILES"] = strings.Join(profiles, ",")
	}

	// Store provider choice for return value (used later for conditional output)
	if useLocalVLLM {
		updates["_USING_LOCAL_VLLM"] = "true"
	}

	if len(updates) > 0 {
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip comments and empty lines
			if strings.HasPrefix(trimmed, "#") || trimmed == "" {
				continue
			}

			// Check each update
			for key, value := range updates {
				if strings.HasPrefix(trimmed, key+"=") {
					lines[i] = fmt.Sprintf("%s=%s", key, value)
					delete(updates, key) // Mark as applied
				}
			}
		}

		// Add any remaining updates that weren't found
		for key, value := range updates {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}

		// Write back
		newContent := strings.Join(lines, "\n")
		if err := os.WriteFile(envPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("write .env: %w", err)
		}

		fmt.Println("✓ Configuration saved to .env")
	} else {
		fmt.Println("✓ No changes made")
	}

	// Check if using local vLLM (look in updates or re-read from env)
	data, _ = os.ReadFile(envPath)
	if strings.Contains(string(data), "COMPOSE_PROFILES=full") {
		os.Setenv("_USING_LOCAL_VLLM", "true")
	}

	return nil
}

func containsPrefix(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}
