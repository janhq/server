# Scripts Directory

This directory contains utility scripts for the Jan Platform.

## Environment Setup Scripts

### setup-env.sh / setup-env.bat

Interactive scripts to help you quickly set up your `.env.local` file.

**Usage (Linux/Mac):**
```bash
./scripts/setup-env.sh
```

**Usage (Windows):**
```cmd
scripts\setup-env.bat
```

These scripts will:
1. Copy `.env.example` to `.env.local`
2. Prompt you for API URL configuration
3. Backup existing `.env.local` if it exists

---

## test-import.ts

A simple test file to verify fumadocs-openapi import is working correctly.

**Usage:**
```bash
npx tsx scripts/test-import.ts
```

---

## Documentation Generation Scripts

### generate-fumadocs-openapi.ts

The main documentation generator that creates interactive API reference pages with live playground components.

### What it does

1. **Cleans previous generation**
   - Removes all tag subdirectories in `content/docs/api-reference/`
   - Preserves root-level manual documentation files (introduction.mdx, meta.json, etc.)

2. **Generates MDX files**
   - Reads OpenAPI schema from `api/server.yaml`
   - Creates one MDX file per API operation
   - Uses custom slug logic to organize files by tag/resource/method
   - Embeds interactive `<APIPage>` components for live API testing

3. **Auto-generates navigation**
   - Extracts tag names from OpenAPI spec
   - Creates `meta.json` files for each tag directory
   - Sorts endpoints by HTTP method (GET, POST, PATCH, DELETE) then alphabetically

### Usage

```bash
npm run generate-openapi
```

Or via the full workflow:

```bash
# Generate OpenAPI spec from backend source
npm run generate-openapi

# Process MDX and build documentation
npm run postinstall

# Start dev server
npm run dev
```

### File Structure

```
content/docs/api-reference/
├── introduction.mdx           # Manual (preserved)
├── debugging-requests.mdx     # Manual (preserved)
├── meta.json                  # Manual (preserved)
├── authentication/            # Auto-generated
│   ├── meta.json             # Auto-generated
│   ├── google/
│   │   ├── login/
│   │   │   └── get.mdx       # Auto-generated
│   │   └── callback/
│   │       └── post.mdx      # Auto-generated
│   └── me/
│       └── get.mdx           # Auto-generated
└── chat-completions/          # Auto-generated
    ├── meta.json             # Auto-generated
    └── completions/
        └── post.mdx          # Auto-generated
```

### Key Features

**Zero-maintenance scalability**

- Add new OpenAPI tags → automatically creates new directories
- Remove tags → automatically cleans up old directories
- No hardcoded tag lists to maintain

**Smart slug generation**

- Strips version prefixes (`/v1/auth/login` → `auth/login`)
- Removes trailing path parameters (`/users/{id}` → `/users`)
- Drops redundant tag prefixes from paths
- Converts to lowercase kebab-case

**Proper tag titles**

- Reads tag names directly from OpenAPI spec
- Maps slugified directories to human-readable titles
- Example: `admin-provider/` → "Admin Provider"

### Configuration

Edit these constants in the script if needed:

```typescript
const OUTPUT_ROOT = './content/docs/api-reference';
const VERSION_SEGMENT_REGEX = /^v\d+$/i;
```

### How Slug Generation Works

For an endpoint: `POST /v1/auth/google/callback`

1. Tag from OpenAPI: `Authentication` → slug: `authentication`
2. Strip version: `/auth/google/callback`
3. Split into segments: `['auth', 'google', 'callback']`
4. Drop tag prefix: `['google', 'callback']` (removes redundant 'auth')
5. Add method: `['authentication', 'google', 'callback', 'post']`
6. Result: `authentication/google/callback/post.mdx`

### Troubleshooting

**Missing meta.json files**

- Ensure the OpenAPI spec has proper tags on all operations
- Check console output for tag extraction results

**Wrong file paths**

- Verify OpenAPI paths follow `/v1/{resource}/{action}` pattern
- Check that tag names match expected slugified format

**Stale documentation**

- All subdirectories are deleted on each run
- If files persist, they may be in the wrong location (should be in subdirectories, not root)

## Development

**Dependencies:**

- `fumadocs-openapi` - Generates MDX files from OpenAPI
- `js-yaml` - Parses YAML OpenAPI specs
- Node.js `fs` - File system operations

**Adding new features:**

1. Update slug logic in `buildOperationSlug()`
2. Modify sorting in `sortPages()`
3. Customize meta.json structure in `generateTagMetaFiles()`

**Testing changes:**

```bash
# Clean and regenerate
rm -rf content/docs/api-reference/*/
npm run generate-openapi
```
