package utils

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

// A Mocked os.DirEntry object, representing an item in a read directory
type fakeDirEntry struct {
	name string
	dir  bool
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.dir }
func (f fakeDirEntry) Type() fs.FileMode          { return 0 }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func Test_listFiles(t *testing.T) {
	t.Parallel()

	errorForcedFailure := errors.New("forced failure")
	errorUnknownDirectory := errors.New("unknown directory")
	tests := []struct {
		name    string
		root    string
		stack   string
		reader  readDirFunc
		want    []string
		wantErr error
	}{
		{
			name:  "SimpleFileStructure",
			root:  "",
			stack: "",
			reader: func(dirname string) ([]fs.DirEntry, error) {
				return []fs.DirEntry{
					&fakeDirEntry{"file1.txt", false},
					&fakeDirEntry{"file2.txt", false},
				}, nil
			},
			want: []string{"file1.txt", "file2.txt"},
		},
		{
			name:  "NestedDirectories",
			root:  "",
			stack: "",
			reader: func(dirname string) ([]fs.DirEntry, error) {
				switch dirname {
				case "":
					return []fs.DirEntry{&fakeDirEntry{"subDir", true}}, nil
				case "subDir":
					return []fs.DirEntry{&fakeDirEntry{"file.txt", false}}, nil
				default:
					return nil, errorUnknownDirectory
				}
			},
			want: []string{filepath.Join("subDir", "file.txt")},
		},
		{
			name:  "RootSet",
			root:  "root-dir",
			stack: "",
			reader: func(dirname string) ([]fs.DirEntry, error) {
				switch filepath.ToSlash(dirname) {
				case "root-dir":
					return []fs.DirEntry{&fakeDirEntry{"subDir", true}}, nil
				case "root-dir/subDir":
					return []fs.DirEntry{&fakeDirEntry{"file.txt", false}}, nil
				default:
					return nil, errorUnknownDirectory
				}
			},
			want: []string{filepath.Join("subDir", "file.txt")},
		},
		{
			name:  "ErrorOnReadDir",
			root:  "",
			stack: "",
			reader: func(dirname string) ([]fs.DirEntry, error) {
				return nil, errorForcedFailure
			},
			wantErr: errorForcedFailure,
		},
		{
			name:  "ErrorOnReadSubDir",
			root:  "",
			stack: "",
			reader: func(dirname string) ([]fs.DirEntry, error) {
				switch dirname {
				case "":
					return []fs.DirEntry{&fakeDirEntry{"subDir", true}}, nil
				default:
					return nil, errorForcedFailure
				}
			},
			wantErr: errorForcedFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _listFiles(tt.reader, tt.root, tt.stack)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error %v, got %v", tt.wantErr, err)
			}

			if len(got) != len(tt.want) {
				t.Errorf("Expected result length %v, got %v", len(tt.want), len(got))
			} else {
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("Expected result %v at index %d, got %v", tt.want[i], i, v)
					}
				}
			}
		})
	}
}
