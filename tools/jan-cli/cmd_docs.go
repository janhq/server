package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Documentation management",
	Long:  `Sync and manage documentation for the Jan Platform.`,
}

var docsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync docs to platform content",
	Long: `Sync documentation from /docs to apps/platform/content/docs.

This command:
  1. Cleans existing docs (preserves api-reference/)
  2. Copies docs folders (architecture, configuration, conventions, guides, runbooks)
  3. Converts .md files to .mdx for fumadocs
  4. Renames README.mdx to index.mdx
  5. Creates proper meta.json files for fumadocs navigation
  6. Adds frontmatter with title to all MDX files`,
	RunE: runDocsSync,
}

func init() {
	docsCmd.AddCommand(docsSyncCmd)
}

// Meta represents a fumadocs meta.json file
type Meta struct {
	Title string   `json:"title"`
	Pages []string `json:"pages"`
}

// Section metadata configuration
var sectionMeta = map[string]Meta{
	"architecture": {
		Title: "Architecture",
		Pages: []string{"index", "system-design", "services", "data-flow", "security", "security-advanced", "observability", "test-flows"},
	},
	"configuration": {
		Title: "Configuration",
		Pages: []string{"index", "docker-compose", "kubernetes", "env-var-mapping", "precedence", "service-migration"},
	},
	"conventions": {
		Title: "Conventions",
		Pages: []string{"index", "conventions", "architecture-patterns", "design-patterns", "workflow"},
	},
	"guides": {
		Title: "Guides",
		Pages: []string{"index", "development", "deployment", "authentication", "monitoring", "monitoring-advanced", "testing", "troubleshooting", "jan-cli", "kong-plugins", "mcp-testing", "mcp-admin-interface", "conversation-management", "prompt-orchestration", "background-mode", "webhooks", "services-template", "user-settings-personalization", "user-management-todo"},
	},
	"runbooks": {
		Title: "Runbooks",
		Pages: []string{"index", "monitoring"},
	},
}

func runDocsSync(cmd *cobra.Command, args []string) error {
	// Get paths relative to project root (we're run from tools/jan-cli)
	sourceDir := filepath.Join("..", "..", "docs")
	destDir := filepath.Join("..", "..", "apps", "platform", "content", "docs")

	fmt.Printf("Syncing docs from %s to %s\n", sourceDir, destDir)

	// Step 1: Clean existing docs (except api-reference and root meta.json)
	fmt.Println("Cleaning existing docs...")
	if err := cleanDocsDir(destDir); err != nil {
		return fmt.Errorf("failed to clean docs dir: %w", err)
	}

	// Step 2: Copy directories (excluding api folder)
	dirs := []string{"architecture", "configuration", "conventions", "guides", "runbooks"}
	for _, dir := range dirs {
		srcPath := filepath.Join(sourceDir, dir)
		dstPath := filepath.Join(destDir, dir)
		if _, err := os.Stat(srcPath); err == nil {
			fmt.Printf("  Copying %s...\n", dir)
			if err := copyDirDocs(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %w", dir, err)
			}
		}
	}

	// Step 3: Copy and convert root .md files to .mdx
	rootFiles := []string{"quickstart.md", "repo-naming.md", "roadmap.md"}
	for _, file := range rootFiles {
		srcFile := filepath.Join(sourceDir, file)
		if _, err := os.Stat(srcFile); err == nil {
			dstFile := filepath.Join(destDir, strings.TrimSuffix(file, ".md")+".mdx")
			fmt.Printf("  Copying %s -> %s\n", file, filepath.Base(dstFile))
			if err := copyFile(srcFile, dstFile); err != nil {
				return fmt.Errorf("failed to copy %s: %w", file, err)
			}
		}
	}

	// Step 4: Rename all .md files to .mdx
	fmt.Println("Converting .md to .mdx...")
	if err := convertMdToMdx(destDir); err != nil {
		return fmt.Errorf("failed to convert md to mdx: %w", err)
	}

	// Step 5: Rename README.mdx to index.mdx
	fmt.Println("Converting README.mdx to index.mdx...")
	if err := renameReadmeToIndex(destDir); err != nil {
		return fmt.Errorf("failed to rename README to index: %w", err)
	}

	// Step 6: Create meta.json files
	fmt.Println("Creating meta.json files...")
	if err := createMetaFiles(destDir); err != nil {
		return fmt.Errorf("failed to create meta files: %w", err)
	}

	// Step 7: Add frontmatter to all MDX files
	fmt.Println("Adding frontmatter to MDX files...")
	if err := addFrontmatterToAllMdx(destDir); err != nil {
		return fmt.Errorf("failed to add frontmatter: %w", err)
	}

	// Step 8: Fix markdown links (remove .md extensions)
	fmt.Println("Fixing markdown links...")
	if err := fixMarkdownLinks(destDir); err != nil {
		return fmt.Errorf("failed to fix markdown links: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Docs synced successfully!")
	return nil
}

func cleanDocsDir(destDir string) error {
	entries, err := os.ReadDir(destDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(destDir, 0755)
		}
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		// Preserve api-reference and root meta.json
		if name == "api-reference" || name == "meta.json" {
			continue
		}
		path := filepath.Join(destDir, name)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func copyDirDocs(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

func convertMdToMdx(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".md") {
			newPath := strings.TrimSuffix(path, ".md") + ".mdx"
			if err := os.Rename(path, newPath); err != nil {
				return err
			}
		}
		return nil
	})
}

func renameReadmeToIndex(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) == "README.mdx" {
			newPath := filepath.Join(filepath.Dir(path), "index.mdx")
			if err := os.Rename(path, newPath); err != nil {
				return err
			}
		}
		return nil
	})
}

func createMetaFiles(destDir string) error {
	for section, meta := range sectionMeta {
		sectionDir := filepath.Join(destDir, section)
		if _, err := os.Stat(sectionDir); os.IsNotExist(err) {
			continue
		}

		metaPath := filepath.Join(sectionDir, "meta.json")
		data, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(metaPath, data, 0644); err != nil {
			return err
		}
		fmt.Printf("  Created %s/meta.json\n", section)
	}
	return nil
}

// extractTitleFromMarkdown extracts the first H1 heading from markdown content
func extractTitleFromMarkdown(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	h1Regex := regexp.MustCompile(`^#\s+(.+)$`)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if matches := h1Regex.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

// addFrontmatterToMdx adds frontmatter with title to an MDX file if it doesn't have it
func addFrontmatterToMdx(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	
	// Check if file already has frontmatter
	if strings.HasPrefix(contentStr, "---") {
		return nil // Already has frontmatter
	}

	// Extract title from content
	title := extractTitleFromMarkdown(contentStr)
	if title == "" {
		// Use filename as fallback
		base := filepath.Base(filePath)
		title = strings.TrimSuffix(base, filepath.Ext(base))
		title = strings.ReplaceAll(title, "-", " ")
		// Capitalize first letter of each word
		words := strings.Fields(title)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		title = strings.Join(words, " ")
	}

	// Create frontmatter
	frontmatter := fmt.Sprintf("---\ntitle: \"%s\"\n---\n\n", title)
	newContent := frontmatter + contentStr

	// Write back to file
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// addFrontmatterToAllMdx recursively adds frontmatter to all MDX files
func addFrontmatterToAllMdx(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".mdx") {
			if err := addFrontmatterToMdx(path); err != nil {
				return fmt.Errorf("failed to add frontmatter to %s: %w", path, err)
			}
		}
		return nil
	})
}

// fixMarkdownLinks removes .md extensions from relative markdown links in MDX files
func fixMarkdownLinks(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".mdx") {
			if err := fixLinksInFile(path); err != nil {
				return fmt.Errorf("failed to fix links in %s: %w", path, err)
			}
		}
		return nil
	})
}

// fixLinksInFile processes a single MDX file to remove .md extensions from links
func fixLinksInFile(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Pattern 1: Convert README.md links to index (e.g., [text](README.md) -> [text](index))
	readmePattern := regexp.MustCompile(`\]\(([^)]*/)README\.md\)`)
	contentStr = readmePattern.ReplaceAllString(contentStr, `]($1)`)

	// Pattern 2: Convert root README.md to empty/index (e.g., [text](README.md) -> [text](index))
	rootReadmePattern := regexp.MustCompile(`\]\(README\.md\)`)
	contentStr = rootReadmePattern.ReplaceAllString(contentStr, `](index)`)

	// Pattern 3: Remove .md extension from links with fragments (e.g., [text](file.md#section) -> [text](file#section))
	mdFragmentPattern := regexp.MustCompile(`\]\(([^):/]+)\.md(#[^)]+)\)`)
	contentStr = mdFragmentPattern.ReplaceAllString(contentStr, `]($1$2)`)

	// Pattern 4: Remove .md extension from path links with fragments (e.g., [text](path/file.md#section) -> [text](path/file#section))
	pathMdFragmentPattern := regexp.MustCompile(`\]\(([^):/]+/[^)]+)\.md(#[^)]+)\)`)
	contentStr = pathMdFragmentPattern.ReplaceAllString(contentStr, `]($1$2)`)

	// Pattern 5: Remove .md extension from relative links and add ./ prefix (e.g., [text](file.md) -> [text](./file))
	// This matches markdown links that end in .md but aren't external URLs or paths
	mdPattern := regexp.MustCompile(`\]\(([^):/\.]+)\.md\)`)
	contentStr = mdPattern.ReplaceAllString(contentStr, `](./$1)`)

	// Pattern 6: Add ./ prefix to simple relative links without extension (e.g., [text](file) -> [text](./file))
	// Only match links that don't already have ./, /, http://, https://, or #
	simpleRelativePattern := regexp.MustCompile(`\]\(([^):/\.#]+)\)`)
	contentStr = simpleRelativePattern.ReplaceAllString(contentStr, `](./$1)`)

	// Pattern 7: Remove .md extension from relative path links (e.g., [text](path/file.md) -> [text](path/file))
	pathMdPattern := regexp.MustCompile(`\]\(([^):/]+/[^)]+)\.md\)`)
	contentStr = pathMdPattern.ReplaceAllString(contentStr, `]($1)`)

	// Write back to file
	return os.WriteFile(filePath, []byte(contentStr), 0644)
}
