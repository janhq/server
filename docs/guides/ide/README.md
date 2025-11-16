# IDE Configuration Guides

IDE-specific configuration and setup guides for Jan Server development.

## Available IDE Guides

### Visual Studio Code
- **[VS Code Guide](vscode.md)** - Complete VS Code setup including:
 - Debug configurations for LLM API and MCP Tools
 - VS Code tasks for service management
 - Environment variable configuration
 - Provider configuration (YAML vs legacy)
 - Common workflows and troubleshooting
 - Configuration file reference

**Ready-to-use configuration files:**
- **[launch.json](launch.json)** - Copy to `.vscode/launch.json`
- **[tasks.json](tasks.json)** - Copy to `.vscode/tasks.json`

**Quick setup:**
```bash
# From project root
mkdir -p .vscode
cp docs/guides/ide/launch.json .vscode/launch.json
cp docs/guides/ide/tasks.json .vscode/tasks.json
# Restart VS Code and press F5
```

## Quick Start

### VS Code Users

1. Open Jan Server workspace in VS Code
2. Install recommended extensions (Go extension)
3. Press `F5` to start debugging
4. See [VS Code Guide](vscode.md) for complete setup

### Other IDEs

Configuration guides for other IDEs coming soon:
- IntelliJ IDEA / GoLand
- Vim/Neovim
- Emacs

**Currently using another IDE?** The core development workflow works with any editor:
```bash
make setup        # Initial setup (.env + docker/.env)
make dev-full     # Start infrastructure + services with host routing
./jan-cli.sh dev run llm-api   # Run service natively (macOS/Linux)
.\jan-cli.ps1 dev run llm-api  # Run service natively (Windows)
```

See [Development Guide](../development.md) for editor-agnostic workflows.

## Contributing

Using a different IDE? We welcome contributions for additional IDE configuration guides!

**What to include:**
- Debug configuration setup
- Build tasks
- Test runner integration 
- Environment variable management
- Hot reload setup
- Common workflows

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for guidelines.
