package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestAdminACLsConfig(t *testing.T) {
	t.Run("unmarshals a list of networks", func(t *testing.T) {
		path := writeTempConfig(t, "adminACLs:\n  - \"192.0.2.0/24\"\n  - \"10.0.0.0/8\"\n")
		conf, err := loadConfigWithViper(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(conf.AdminACLs) != 2 || conf.AdminACLs[0] != "192.0.2.0/24" || conf.AdminACLs[1] != "10.0.0.0/8" {
			t.Fatalf("adminACLs = %v, want [192.0.2.0/24 10.0.0.0/8]", conf.AdminACLs)
		}
	})

	t.Run("defaults to loopback when omitted", func(t *testing.T) {
		path := writeTempConfig(t, "")
		conf, err := loadConfigWithViper(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(conf.AdminACLs) != 1 || conf.AdminACLs[0] != "127.0.0.1/32" {
			t.Fatalf("adminACLs = %v, want [127.0.0.1/32]", conf.AdminACLs)
		}
	})

	t.Run("converts a legacy scalar value to a one-entry list", func(t *testing.T) {
		path := writeTempConfig(t, "adminACLs: \"127.0.0.1/32\"\n")
		conf, err := loadConfigWithViper(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(conf.AdminACLs) != 1 || conf.AdminACLs[0] != "127.0.0.1/32" {
			t.Fatalf("adminACLs = %v, want [127.0.0.1/32]", conf.AdminACLs)
		}
	})

	t.Run("splits a legacy comma-separated string", func(t *testing.T) {
		path := writeTempConfig(t, "adminACLs: \"192.0.2.0/24,10.0.0.0/8\"\n")
		conf, err := loadConfigWithViper(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(conf.AdminACLs) != 2 || conf.AdminACLs[0] != "192.0.2.0/24" || conf.AdminACLs[1] != "10.0.0.0/8" {
			t.Fatalf("adminACLs = %v, want [192.0.2.0/24 10.0.0.0/8]", conf.AdminACLs)
		}
	})
}
