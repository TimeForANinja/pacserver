package utils

import (
	"path/filepath"
	"strings"
)

// NormalizePath ensures all path separators are platform-specific.
func NormalizePath(p string) string {
	// replace all windows separators with unix, so all seps should be slashes
	unixSlashes := strings.ReplaceAll(p, "\\", "/")
	// Use filepath.Clean to normalize the path separators and resolve redundancies
	return filepath.Clean(filepath.FromSlash(unixSlashes))
}
