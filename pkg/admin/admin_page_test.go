package admin

import (
	"strings"
	"testing"
)

func TestRenderAdminPage(t *testing.T) {
	page, err := renderAdminPage("/metrics")
	if err != nil {
		t.Fatalf("renderAdminPage returned error: %v", err)
	}

	if !strings.Contains(page, "Lookup debug") {
		t.Fatal("admin page is missing the lookup panel title")
	}
	if !strings.Contains(page, "Metrics") {
		t.Fatal("admin page is missing the metrics panel title")
	}
	if !strings.Contains(page, "/metrics") {
		t.Fatal("admin page does not include the prometheus path")
	}
	if !strings.Contains(page, "Reload") {
		t.Fatal("admin page is missing the reload button")
	}
	if !strings.Contains(page, "reloadStatus") {
		t.Fatal("admin page is missing the reload status area")
	}
}

func TestRenderAdminLoginPage(t *testing.T) {
	page, err := renderAdminLoginPage()
	if err != nil {
		t.Fatalf("renderAdminLoginPage returned error: %v", err)
	}

	if !strings.Contains(page, "Enter the admin token") {
		t.Fatal("admin login page is missing the prompt")
	}
}
