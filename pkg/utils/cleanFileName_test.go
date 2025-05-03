package utils

import (
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Path with backslashes",
			path: "dir1\\dir2\\file.txt",
			want: "dir1/dir2/file.txt",
		},
		{
			name: "Path with forward slashes",
			path: "dir1/dir2/file.txt",
			want: "dir1/dir2/file.txt",
		},
		{
			name: "Path with mixed slashes",
			path: "dir1\\dir2/dir3\\file.txt",
			want: "dir1/dir2/dir3/file.txt",
		},
		{
			name: "Path with redundant separators",
			path: "dir1//dir2\\\\dir3/file.txt",
			want: "dir1/dir2/dir3/file.txt",
		},
		{
			name: "Path with dot notation",
			path: "dir1/./dir2/../dir3/file.txt",
			want: "dir1/dir3/file.txt",
		},
		{
			name: "Absolute path",
			path: "/dir1/dir2/file.txt",
			want: "/dir1/dir2/file.txt",
		},
		{
			name: "Windows absolute path",
			path: "C:\\dir1\\dir2\\file.txt",
			want: "C:/dir1/dir2/file.txt",
		},
		{
			name: "Path with trailing slash",
			path: "dir1/dir2/",
			want: "dir1/dir2",
		},
		{
			name: "Empty path",
			path: "",
			want: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePath(tt.path)
			// replace slash with OS-specific separator
			want := filepath.FromSlash(tt.want)
			if got != want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.path, got, want)
			}
		})
	}
}
