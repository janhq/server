# @janhq/interfaces

Shared UI component library for Jan applications.

## Installation

This package is part of the Jan workspace and is automatically available to all workspace packages.

Add it to your `package.json`:

```json
{
  "dependencies": {
    "@janhq/interfaces": "workspace:*"
  }
}
```

## Usage

Import components directly from the package:

```tsx
import { Button } from "@janhq/interfaces/button";
import { Dialog } from "@janhq/interfaces/dialog";
import { Input } from "@janhq/interfaces/input";

function MyComponent() {
  return <Button variant="default">Click me</Button>;
}
```

## Available Components

All components are built with:

- **React 19** - Latest React features
- **Radix UI** - Accessible component primitives
- **Tailwind CSS** - Utility-first styling
- **TypeScript** - Full type safety

### Core Components

- `avatar` - User avatar with fallback
- `badge` - Status badges and labels
- `button` - Interactive button with variants
- `card` - Content containers
- `dialog` - Modal dialogs
- `drawer` - Side drawers
- `dropdown-menu` - Dropdown menus
- `input` - Text input fields
- `label` - Form labels
- `select` - Dropdown select
- `separator` - Visual dividers
- `sheet` - Slide-out panels
- `sidebar` - Navigation sidebar
- `skeleton` - Loading placeholders
- `switch` - Toggle switches
- `textarea` - Multi-line text input
- `tooltip` - Contextual tooltips

### Composed Components

- `button-group` - Grouped buttons
- `collapsible` - Collapsible content
- `command` - Command palette
- `dropdrawer` - Combined dropdown/drawer
- `empty` - Empty state displays
- `field` - Form field wrapper
- `hover-card` - Hover popover
- `input-group` - Grouped inputs
- `popover` - Floating content

### Icons

- `svgs/discord` - Discord icon
- `svgs/github` - GitHub icon
- `svgs/google` - Google icon
- `svgs/jan` - Jan logo

### Utilities

```tsx
import { cn } from "@janhq/interfaces/lib";

// Merge Tailwind classes
const className = cn("base-class", condition && "conditional-class");
```

## Development

```bash
# Type check
pnpm type-check

# Lint
pnpm lint
```

## Notes

- All components use the `cn()` utility from `./lib.ts` for className merging
- Components are designed to work with Tailwind CSS
- Peer dependencies: React 19+
