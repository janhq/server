//go:build tools
// +build tools

package tools

import (
	_ "github.com/swaggo/swag/cmd/swag"
)

// This file declares dependencies on build tools.
// Go modules will download these tools, but they won't be included in the binary.
