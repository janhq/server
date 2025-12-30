/**
 * Generates interactive API documentation from OpenAPI spec using Fumadocs.
 *
 * Process:
 * 1. Cleans all generated tag directories in content/docs/api-reference
 *    - Deletes all subdirectories (generated content)
 *    - Preserves root-level .mdx files and meta.json (manual documentation)
 * 2. Reads OpenAPI schema from api/server.yaml
 * 3. Generates one MDX file per API operation with custom slug logic:
 *    - Strips version segments (e.g., /v1/)
 *    - Removes trailing path parameters (e.g., {id})
 *    - Drops tag prefixes from paths to avoid redundancy
 *    - Organizes by: {tag}/{resource-path}/{http-method}.mdx
 * 4. Creates files with <APIPage> components for interactive API playground
 * 5. Auto-generates meta.json files for each tag directory:
 *    - Scans generated tag directories
 *    - Discovers all endpoint .mdx files
 *    - Creates navigation metadata with proper tag titles from OpenAPI spec
 *    - Sorts pages by HTTP method (GET, POST, PATCH, DELETE) then alphabetically
 *
 * This approach ensures docs always match the OpenAPI spec with no manual maintenance.
 */
import { load } from 'js-yaml';
import { existsSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import path from 'node:path';

type GenerateFiles = (typeof import('fumadocs-openapi'))['generateFiles'];

const loadGenerateFiles = async () => {
  const { generateFiles } = await import('fumadocs-openapi');
  return generateFiles as GenerateFiles;
};

const OUTPUT_ROOT = './content/docs/api-reference';
const VERSION_SEGMENT_REGEX = /^v\d+$/i;

const toPosixPath = (value: string) => value.replace(/\\/g, '/');

/**
 * Converts a string into a URL-friendly slug
 * Removes special characters, replaces spaces with hyphens, and lowercases
 */
const slugifySegment = (value: string) =>
  value
    .trim()
    .replace(/[{}]/g, '')
    .replace(/[^a-zA-Z0-9]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .toLowerCase();

/**
 * Strips version prefix (e.g., /v1/) from API route path and returns path segments
 */
const stripVersionAndParams = (routePath: string) => {
  const rawSegments = routePath.split('/').filter(Boolean);
  if (rawSegments.length > 0 && VERSION_SEGMENT_REGEX.test(rawSegments[0])) {
    rawSegments.shift();
  }

  return rawSegments;
};

/**
 * Removes tag prefix from path segments to avoid redundancy in URLs
 * Example: ['auth', 'google', 'login'] with tag 'auth' -> ['google', 'login']
 */
const dropTagPrefix = (segments: string[], tagSlug: string) => {
  let combined = '';
  let prefixEnd = -1;

  for (let i = 0; i < segments.length; i += 1) {
    combined = combined ? `${combined}-${segments[i]}` : segments[i];
    if (combined === tagSlug) {
      prefixEnd = i;
    }
  }

  if (prefixEnd >= 0) {
    const sliced = segments.slice(prefixEnd + 1);
    if (sliced.length > 0) {
      return sliced;
    }
  }

  if (segments.length > 1 && tagSlug.startsWith(segments[0])) {
    return segments.slice(1);
  }

  return segments;
};

/**
 * Builds the file path slug for an API operation
 * Combines tag, resource path, and HTTP method into a clean URL structure
 * Example: /v1/auth/google/login GET -> authentication/google/login/get
 */
const buildOperationSlug = (entry: any): string => {
  if (entry.type !== 'operation') {
    return entry.item.method.toLowerCase();
  }

  const tags: string[] = entry.item.tags ?? [];
  const tagSlug = slugifySegment(tags[0] ?? 'general') || 'general';

  const rawSegments = stripVersionAndParams(entry.item.path);

  const cleanedSegments: string[] = rawSegments
    .map((segment, index) => {
      const isParam = segment.startsWith('{') && segment.endsWith('}');
      const isLast = index === rawSegments.length - 1;

      if (isParam && isLast) {
        return undefined;
      }

      const slug = slugifySegment(segment);
      return slug || undefined;
    })
    .filter((segment): segment is string => Boolean(segment));

  let resourceSegments = cleanedSegments;

  if (resourceSegments.length === 0) {
    resourceSegments = ['root'];
  }

  resourceSegments = dropTagPrefix(resourceSegments, tagSlug);

  if (resourceSegments.length === 0) {
    resourceSegments = cleanedSegments.length > 0 ? [cleanedSegments.at(-1)!] : ['root'];
  }

  const methodSlug = slugifySegment(entry.item.method);
  const finalSegments = [tagSlug, ...resourceSegments, methodSlug].filter(Boolean);

  return finalSegments.join('/');
};

/**
 * Removes all previously generated tag directories to ensure a clean slate
 * Deletes all subdirectories in api-reference/ while preserving root-level files
 * Root-level files (introduction.mdx, meta.json, etc.) are manual documentation
 */
const cleanGeneratedDirectories = () => {
  const outputPath = path.resolve(OUTPUT_ROOT);

  if (!existsSync(outputPath)) {
    return;
  }

  const entries = readdirSync(outputPath);

  for (const entry of entries) {
    const fullPath = path.join(outputPath, entry);
    const stat = statSync(fullPath);

    // Delete all subdirectories (generated content)
    // Preserve root-level files (manual documentation)
    if (stat.isDirectory()) {
      rmSync(fullPath, { recursive: true, force: true });
      console.log(`  🗑️  Removed ${entry}/`);
    }
  }
};

/**
 * Extract tag names from OpenAPI spec and build slug -> title mapping
 */
const extractTagTitles = (schemaPath: string): Map<string, string> => {
  const yaml = readFileSync(schemaPath, 'utf-8');
  const schema = load(yaml) as any;
  const tagMap = new Map<string, string>();

  // Extract tags from all operations
  if (schema.paths) {
    for (const pathObj of Object.values(schema.paths)) {
      for (const operation of Object.values(pathObj as any)) {
        const op = operation as any;
        if (op.tags && Array.isArray(op.tags)) {
          for (const tag of op.tags) {
            const slug = slugifySegment(tag);
            if (slug && !tagMap.has(slug)) {
              tagMap.set(slug, tag);
            }
          }
        }
      }
    }
  }

  return tagMap;
};

/**
 * Recursively find all .mdx files in a directory
 */
const findMdxFiles = (dir: string, baseDir: string): string[] => {
  const files: string[] = [];
  const entries = readdirSync(dir);

  for (const entry of entries) {
    const fullPath = path.join(dir, entry);
    const stat = statSync(fullPath);

    if (stat.isDirectory()) {
      files.push(...findMdxFiles(fullPath, baseDir));
    } else if (entry.endsWith('.mdx')) {
      // Convert to relative path (always POSIX) and remove .mdx extension
      const relativePath = path.relative(baseDir, fullPath);
      const pagePath = toPosixPath(relativePath).replace(/\.mdx$/, '');
      files.push(pagePath);
    }
  }

  return files;
};

/**
 * Sort pages by HTTP method order (GET, POST, PATCH, DELETE) and path
 */
const sortPages = (pages: string[]): string[] => {
  const methodOrder = ['get', 'post', 'patch', 'delete', 'put'];

  return pages.sort((a, b) => {
    const methodA = a.split('/').pop() || '';
    const methodB = b.split('/').pop() || '';

    const orderA = methodOrder.indexOf(methodA);
    const orderB = methodOrder.indexOf(methodB);

    // If methods are different and both are in the order list, sort by method
    if (orderA !== -1 && orderB !== -1 && orderA !== orderB) {
      return orderA - orderB;
    }

    // Otherwise sort alphabetically
    return a.localeCompare(b);
  });
};

/**
 * Generate meta.json files for each tag directory
 */
const generateTagMetaFiles = (tagTitles: Map<string, string>) => {
  const outputPath = path.resolve(OUTPUT_ROOT);

  // Find all directories in OUTPUT_ROOT (excluding root-level files)
  const entries = readdirSync(outputPath);

  for (const entry of entries) {
    const fullPath = path.join(outputPath, entry);
    const stat = statSync(fullPath);

    if (stat.isDirectory()) {
      // Find all .mdx files in this tag directory
      const pages = findMdxFiles(fullPath, fullPath);

      if (pages.length > 0) {
        // Sort pages
        const sortedPages = sortPages(pages);

        // Get tag title (default to capitalized slug if not found)
        const tagSlug = entry;
        const title =
          tagTitles.get(tagSlug) ||
          tagSlug
            .split('-')
            .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');

        // Create meta.json content
        const metaContent = {
          title,
          pages: sortedPages,
        };

        // Write meta.json
        const metaPath = path.join(fullPath, 'meta.json');
        writeFileSync(metaPath, JSON.stringify(metaContent, null, 2) + '\n');
        console.log(`  ✓ Generated ${entry}/meta.json (${sortedPages.length} pages)`);
      }
    }
  }
};

/**
 * This script uses Fumadocs' native OpenAPI generator
 * to create MDX files with interactive API playground components
 */
async function main() {
  console.log('🚀 Generating Fumadocs OpenAPI documentation...\n');

  try {
    const schemaPath = './api/server.yaml';

    console.log('🧹 Cleaning previously generated directories...');
    cleanGeneratedDirectories();

    console.log('\n📝 Generating MDX files from OpenAPI schema...');
    const generateFiles = await loadGenerateFiles();
    await generateFiles({
      input: [schemaPath], // Path to your OpenAPI schema
      output: OUTPUT_ROOT,
      per: 'operation', // Generate one file per API operation
      groupBy: 'none',
      name: (entry) => buildOperationSlug(entry),
    });
    console.log('✅ Interactive API documentation generated');

    // Extract tag titles and generate meta.json files
    console.log('\n📋 Generating meta.json files for tag directories...');
    const tagTitles = extractTagTitles(schemaPath);
    console.log('✅ Extracted tag titles from OpenAPI schema');
    generateTagMetaFiles(tagTitles);
    console.log('✅ Generated meta.json files for each tag directory');
    console.log(
      '\n✨ Complete! Files will use the <APIPage> component for interactive playgrounds',
    );
  } catch (error) {
    console.error('❌ Error generating OpenAPI docs:', error);
    process.exit(1);
  }
}

main();
