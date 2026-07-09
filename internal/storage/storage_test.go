package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadIPMaps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zones.csv")
	data := "" +
		"# comment line\n" +
		"192.168.0.0,24,lan.pac, local network\n" +
		"\n" +
		"10.0.0.0,8,corp.pac\n"
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
}

func TestReadPACTemplates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "one.pac"), []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "two.pac"), []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}

	templates, err, problems := ReadPACTemplates(dir)
	if err != nil {
		t.Fatalf("ReadPACTemplates returned error: %v", err)
	}
	if problems != 0 {
		t.Fatalf("ReadPACTemplates returned %d problems, want 0", problems)
	}
	if len(templates) != 2 {
		t.Fatalf("ReadPACTemplates returned %d templates, want 2", len(templates))
	}
}

func TestInitCachesAndFindInLUT(t *testing.T) {
	root := t.TempDir()
	pacRoot := filepath.Join(root, "pacs")
	if err := os.MkdirAll(pacRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "default.pac"), []byte("DEFAULT {{ .Contact }}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "wpad.dat"), []byte("WPAD"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pacRoot, "lan.pac"), []byte("LAN {{ .Contact }}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "zones.csv"), []byte("192.168.0.0,24,lan.pac\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := StorageConfig{
		IPMapFile:      filepath.Join(root, "zones.csv"),
		PACRoot:        pacRoot,
		DefaultPACFile: filepath.Join(root, "default.pac"),
		WPADFile:       filepath.Join(root, "wpad.dat"),
		ContactInfo:    "Help Desk",
		IgnoreMinors:   false,
	}

	if err := InitCaches(cfg); err != nil {
		t.Fatalf("InitCaches returned error: %v", err)
	}

	if got := DefaultPAC(); got == nil || got.Variant == "" {
		t.Fatal("DefaultPAC was not loaded")
	}
	if got := WPAD(); got == nil || got.Variant == "" {
		t.Fatal("WPAD was not loaded")
	}

	entry, _, stack := FindInLUT("192.168.0.1", 32)
	if entry == nil {
		t.Fatal("FindInLUT returned nil entry")
	}
	if entry.IPMap == nil || entry.IPMap.Filename != "lan.pac" {
		t.Fatalf("FindInLUT returned wrong entry: %#v", entry)
	}
	if len(stack) == 0 {
		t.Fatal("FindInLUT returned empty stack")
	}
}
