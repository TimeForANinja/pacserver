package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestBuildAccessLogEntry(t *testing.T) {
	app := fiber.New()
	var entry accessLogEntry
	app.Get("/wpad.dat", func(c *fiber.Ctx) error {
		c.Locals(accessLogLocalSettledIP, "10.0.0.42")
		c.Locals(accessLogLocalXFF, "203.0.113.10, 198.51.100.7")
		c.Locals(accessLogLocalPAC, "corp/main.pac")
		c.Status(http.StatusNoContent)
		entry = buildAccessLogEntry(c, time.Unix(0, 0))
		return c.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/wpad.dat", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app test failed: %v", err)
	}
	_ = resp.Body.Close()

	if entry.Method != http.MethodGet {
		t.Fatalf("method = %q, want %q", entry.Method, http.MethodGet)
	}
	if entry.Path != "/wpad.dat" {
		t.Fatalf("path = %q, want %q", entry.Path, "/wpad.dat")
	}
	if entry.Status != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", entry.Status, http.StatusNoContent)
	}
	if entry.SettledIP != "10.0.0.42" {
		t.Fatalf("settled ip = %q, want %q", entry.SettledIP, "10.0.0.42")
	}
	if entry.XForwardedFor != "203.0.113.10, 198.51.100.7" {
		t.Fatalf("xff = %q, want %q", entry.XForwardedFor, "203.0.113.10, 198.51.100.7")
	}
	if entry.ServedPAC != "corp/main.pac" {
		t.Fatalf("served pac = %q, want %q", entry.ServedPAC, "corp/main.pac")
	}
	if entry.Timestamp == "" {
		t.Fatal("timestamp must not be empty")
	}
}
