package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

func resetTestCache(t *testing.T) {
	t.Helper()

	cache.Lock()
	defer cache.Unlock()

	cache.lookupTree = nil
	cache.defaultPAC = nil
	cache.cachedIPMap = nil
	cache.cachedPACs = nil
}

func mustIPNet(t *testing.T, ip string, cidr int) IP.Net {
	t.Helper()

	ipNet, err := IP.NewIPNetFromMixed(ip, cidr)
	if err != nil {
		t.Fatalf("failed to create IP net %s/%d: %v", ip, cidr, err)
	}
	return ipNet
}

func mustEntry(t *testing.T, ip string, cidr int, filename, content string) *LookupEntry {
	t.Helper()

	ipNet := mustIPNet(t, ip, cidr)
	entry, err := NewLookupEntry(
		&IPMap{IPNet: ipNet, Filename: filename},
		&PACTemplate{Filename: filename, Content: content},
		"Help Desk",
	)
	if err != nil {
		t.Fatalf("failed to build lookup entry for %s: %v", filename, err)
	}
	return entry
}

func TestParseIPMapLineBlackbox(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantNil     bool
		wantErr     bool
		wantFile    string
		wantCIDR    uint8
		wantComment string
	}{
		{
			name:    "Comment line with //",
			line:    "// This is a comment",
			wantNil: true,
		},
		{
			name:    "Comment line with #",
			line:    "# This is a comment",
			wantNil: true,
		},
		{
			name:    "Empty line",
			line:    "",
			wantNil: true,
		},
		{
			name:     "Valid line",
			line:     "192.168.0.0,24,test.pac",
			wantFile: "test.pac",
			wantCIDR: 24,
		},
		{
			name:        "Valid line with whitespace and comment",
			line:        " 192.168.0.0 , 24 , test.pac , local network ",
			wantFile:    "test.pac",
			wantCIDR:    24,
			wantComment: "local network",
		},
		{
			name:    "Invalid number of fields",
			line:    "192.168.0.0,24",
			wantErr: true,
		},
		{
			name:    "Invalid IP address",
			line:    "invalid,24,test.pac",
			wantErr: true,
		},
		{
			name:    "Invalid CIDR",
			line:    "192.168.0.0,invalid,test.pac",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIPMapLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseIPMapLine() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if got != nil {
					t.Fatalf("parseIPMapLine() = %#v, want nil on error", got)
				}
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Fatalf("parseIPMapLine() = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("parseIPMapLine() returned nil map")
			}
			if got.Filename != tt.wantFile {
				t.Fatalf("parseIPMapLine() filename = %q, want %q", got.Filename, tt.wantFile)
			}
			if got.IPNet.GetRawCIDR() != tt.wantCIDR {
				t.Fatalf("parseIPMapLine() cidr = %d, want %d", got.IPNet.GetRawCIDR(), tt.wantCIDR)
			}
			if got.Comment != tt.wantComment {
				t.Fatalf("parseIPMapLine() comment = %q, want %q", got.Comment, tt.wantComment)
			}
		})
	}
}

func TestReadIPMapsBlackbox(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zones.csv")
	data := strings.Join([]string{
		"# comment line",
		"192.168.0.0,24,lan.pac, local network",
		"",
		"10.0.0.0,8,corp.pac",
	}, "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	maps, err, problems := ReadIPMaps(path)
	if err != nil {
		t.Fatalf("ReadIPMaps returned error: %v", err)
	}
	if problems != 0 {
		t.Fatalf("ReadIPMaps returned %d problems, want 0", problems)
	}
	if len(maps) != 2 {
		t.Fatalf("ReadIPMaps returned %d maps, want 2", len(maps))
	}
	if maps[0].Filename != "lan.pac" || maps[1].Filename != "corp.pac" {
		t.Fatalf("unexpected filenames: %#v", maps)
	}
	if maps[0].Comment != "local network" {
		t.Fatalf("unexpected comment: %#v", maps[0])
	}
}

func TestBuildLookupEntriesBlackbox(t *testing.T) {
	newPAC1 := &PACTemplate{
		Filename: "test1.pac",
		Content:  "// This is test1.pac by {{ .Contact }}",
	}
	newPAC2 := &PACTemplate{
		Filename: "test2.pac",
		Content:  "// This is test2.pac by {{ .Contact }}",
	}
	oldPAC3 := &PACTemplate{
		Filename: "test3.pac",
		Content:  "// This is test3.pac by {{ .Contact }}",
	}

	ipMap1 := &IPMap{IPNet: mustIPNet(t, "192.168.0.0", 24), Filename: "test1.pac"}
	ipMap2 := &IPMap{IPNet: mustIPNet(t, "10.0.0.0", 8), Filename: "test2.pac"}
	ipMap3 := &IPMap{IPNet: mustIPNet(t, "172.16.0.0", 12), Filename: "test3.pac"}
	ipMap4 := &IPMap{IPNet: mustIPNet(t, "8.8.8.0", 24), Filename: "test4.pac"}

	tests := []struct {
		name          string
		newPACs       []*PACTemplate
		oldPACs       []*PACTemplate
		newIPMaps     []*IPMap
		wantElements  int
		wantKeepPACs  int
		wantProbCount int
		wantKeepFile  string
	}{
		{
			name:          "All PACs found in newPACs",
			newPACs:       []*PACTemplate{newPAC1, newPAC2},
			oldPACs:       []*PACTemplate{oldPAC3},
			newIPMaps:     []*IPMap{ipMap1, ipMap2},
			wantElements:  2,
			wantKeepPACs:  0,
			wantProbCount: 0,
		},
		{
			name:          "Some PACs found in oldPACs",
			newPACs:       []*PACTemplate{newPAC1},
			oldPACs:       []*PACTemplate{oldPAC3},
			newIPMaps:     []*IPMap{ipMap1, ipMap3},
			wantElements:  2,
			wantKeepPACs:  1,
			wantProbCount: 1,
			wantKeepFile:  "test3.pac",
		},
		{
			name:          "Some PACs not found at all",
			newPACs:       []*PACTemplate{newPAC1},
			oldPACs:       []*PACTemplate{},
			newIPMaps:     []*IPMap{ipMap1, ipMap4},
			wantElements:  1,
			wantKeepPACs:  0,
			wantProbCount: 1,
		},
		{
			name:          "No PACs found",
			newPACs:       []*PACTemplate{},
			oldPACs:       []*PACTemplate{},
			newIPMaps:     []*IPMap{ipMap1, ipMap2},
			wantElements:  0,
			wantKeepPACs:  0,
			wantProbCount: 2,
		},
		{
			name:          "Empty inputs",
			newPACs:       []*PACTemplate{},
			oldPACs:       []*PACTemplate{},
			newIPMaps:     []*IPMap{},
			wantElements:  0,
			wantKeepPACs:  0,
			wantProbCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, keepPACs, probCount := buildLookupEntries(
				tt.newIPMaps,
				mapFromTemplates(tt.newPACs),
				mapFromTemplates(tt.oldPACs),
				"Test Contact",
			)

			if len(elements) != tt.wantElements {
				t.Fatalf("buildLookupEntries() returned %d elements, want %d", len(elements), tt.wantElements)
			}
			if len(keepPACs) != tt.wantKeepPACs {
				t.Fatalf("buildLookupEntries() returned %d keepPACs, want %d", len(keepPACs), tt.wantKeepPACs)
			}
			if probCount != tt.wantProbCount {
				t.Fatalf("buildLookupEntries() returned problem count %d, want %d", probCount, tt.wantProbCount)
			}
			if tt.wantKeepFile != "" {
				if _, ok := keepPACs[tt.wantKeepFile]; !ok {
					t.Fatalf("buildLookupEntries() did not keep %q", tt.wantKeepFile)
				}
			}
			for _, entry := range elements {
				if entry == nil || entry.IPMap == nil || entry.PAC == nil {
					t.Fatalf("buildLookupEntries() returned incomplete entry: %#v", entry)
				}
				if !strings.Contains(entry.Variant, "Test Contact") {
					t.Fatalf("expected rendered PAC variant to contain contact info, got %q", entry.Variant)
				}
			}
		})
	}
}

func TestBuildLookupTreeAndFindInLUTBlackbox(t *testing.T) {
	tests := []struct {
		name             string
		entries          []*LookupEntry
		queryIP          string
		queryCIDR        int
		wantOneOf        []string
		wantExact        string
		wantStackAtLeast int
	}{
		{
			name:             "No elements at all",
			entries:          []*LookupEntry{},
			queryIP:          "192.168.1.1",
			queryCIDR:        32,
			wantExact:        "fallback.pac",
			wantStackAtLeast: 1,
		},
		{
			name: "Explicit default overwrites built-in root",
			entries: []*LookupEntry{
				mustEntry(t, "0.0.0.0", 0, "custom-root.pac", "function FindProxyForURL(url, host) { return 'DIRECT'; }"),
			},
			queryIP:          "10.0.0.1",
			queryCIDR:        32,
			wantExact:        "custom-root.pac",
			wantStackAtLeast: 1,
		},
		{
			name: "Identical IPNet for two objects",
			entries: []*LookupEntry{
				mustEntry(t, "192.168.0.0", 24, "first.pac", "// first"),
				mustEntry(t, "192.168.0.0", 24, "second.pac", "// second"),
			},
			queryIP:          "192.168.0.1",
			queryCIDR:        32,
			wantOneOf:        []string{"first.pac", "second.pac"},
			wantStackAtLeast: 1,
		},
		{
			name: "Duplicate IPNet entries keep lookup stable",
			entries: []*LookupEntry{
				mustEntry(t, "192.168.0.0", 16, "first.pac", "// first"),
				mustEntry(t, "192.168.0.0", 16, "second.pac", "// second"),
				mustEntry(t, "192.168.0.0", 24, "child.pac", "// child"),
			},
			queryIP:          "192.168.0.1",
			queryCIDR:        32,
			wantOneOf:        []string{"first.pac", "second.pac", "child.pac"},
			wantStackAtLeast: 1,
		},
		{
			name: "Most specific node wins",
			entries: []*LookupEntry{
				mustEntry(t, "192.168.0.0", 16, "parent.pac", "// parent"),
				mustEntry(t, "192.168.0.0", 24, "child.pac", "// child"),
			},
			queryIP:          "192.168.0.1",
			queryCIDR:        32,
			wantExact:        "child.pac",
			wantStackAtLeast: 2,
		},
		{
			name: "Invalid IP input falls back to root",
			entries: []*LookupEntry{
				mustEntry(t, "10.0.0.0", 8, "network.pac", "// network"),
			},
			queryIP:          "not-an-ip",
			queryCIDR:        32,
			wantExact:        "fallback.pac",
			wantStackAtLeast: 1,
		},
		{
			name: "Invalid CIDR input falls back to root",
			entries: []*LookupEntry{
				mustEntry(t, "10.0.0.0", 8, "network.pac", "// network"),
			},
			queryIP:          "10.0.0.1",
			queryCIDR:        33,
			wantExact:        "fallback.pac",
			wantStackAtLeast: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTestCache(t)

			tree := buildLookupTree(tt.entries, nil, "Test Contact")
			if tree == nil {
				t.Fatal("buildLookupTree returned nil")
			}

			cache.Lock()
			cache.lookupTree = tree
			cache.defaultPAC = tree.Content
			cache.Unlock()

			entry, _, stack := FindInLUT(tt.queryIP, tt.queryCIDR)
			if entry == nil {
				t.Fatal("FindInLUT returned nil entry")
			}
			if len(stack) < tt.wantStackAtLeast {
				t.Fatalf("FindInLUT returned stack length %d, want at least %d", len(stack), tt.wantStackAtLeast)
			}

			gotFile := entry.IPMap.Filename
			if tt.wantExact != "" && gotFile != tt.wantExact {
				t.Fatalf("FindInLUT returned %q, want %q", gotFile, tt.wantExact)
			}
			if len(tt.wantOneOf) > 0 && !contains(tt.wantOneOf, gotFile) {
				t.Fatalf("FindInLUT returned %q, want one of %v", gotFile, tt.wantOneOf)
			}
		})
	}
}

func mapFromTemplates(arr []*PACTemplate) map[string]*PACTemplate {
	res := make(map[string]*PACTemplate, len(arr))
	for _, item := range arr {
		if item == nil {
			continue
		}
		res[item.Filename] = item
	}
	return res
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestBuildLookupTreeBlackboxStructure(t *testing.T) {
	resetTestCache(t)

	tree := buildLookupTree([]*LookupEntry{
		mustEntry(t, "10.0.0.0", 8, "first", "// first"),
		mustEntry(t, "192.168.1.0", 24, "second", "// second"),
		mustEntry(t, "192.168.2.0", 24, "third", "// third"),
	}, nil, "Test Contact")

	if tree == nil {
		t.Fatal("buildLookupTree returned nil")
	}
	if tree.Content == nil {
		t.Fatal("buildLookupTree returned nil content")
	}

	got := []string{}
	for _, child := range tree.Children {
		if child != nil && child.Content != nil && child.Content.IPMap != nil {
			got = append(got, child.Content.IPMap.Filename)
		}
	}
	want := []string{"first", "second", "third"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildLookupTree child order = %v, want %v", got, want)
	}
}
