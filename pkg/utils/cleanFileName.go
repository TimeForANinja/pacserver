package utils

import (
	"path/filepath"
	"strings"
)

// NormalizePath ensures all path separators are platform-specific.
func NormalizePath(p string) string {
	// Replace all '\' with '/' for uniform formatting
	unixFormatted := strings.ReplaceAll(p, "\\", "/")
	// Use filepath.Clean to normalize the path separators and resolve redundancies
	return filepath.Clean(filepath.FromSlash(unixFormatted))
}
