package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/timeforaninja/pacserver/pkg/acl"
)

func TestHTTPErrorHandlerWritesFiberError(t *testing.T) {
	app := newHTTPApp()
	app.Get("/forbidden", func(c *fiber.Ctx) error { return fiber.ErrForbidden })

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/forbidden", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
	}
}

// TestSingleListenerACLWiring builds the app exactly like LaunchServer (minus Listen) and
// verifies that the admin routes are gated by the source-network ACL while the PAC
// catch-all stays open to everyone.
func TestSingleListenerACLWiring(t *testing.T) {
	logDir := t.TempDir()
	configPath := filepath.Join(logDir, "config.yaml")
	// Fiber's app.Test serves from remote address 0.0.0.0, which the ACL below intentionally excludes.
	config := fmt.Sprintf(`ipMapFile: "../demo_files/zones.csv"
routeMapFile: "../demo_files/routes.csv"
pacRoot: "../demo_files/pacs"
defaultPACFile: "../demo_files/pacs/default.pac"
port: 18080
adminACLs:
  - "192.0.2.0/24"
adminSecret: "test-secret"
prometheusEnabled: true
accessLogFile: %q
eventLogFile: %q
`, filepath.Join(logDir, "access.log"), filepath.Join(logDir, "event.log"))
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}

	// Build the app in the same order as LaunchServer.
	app := newHTTPApp()
	installMiddlewares(app)
	trackPac := registerPrometheusMiddleware(app)
	aclHandler, err := acl.Middleware(GetConfig().AdminACLs)
	if err != nil {
		t.Fatal(err)
	}
	registerAdminRoutes(app, aclHandler)
	registerPrometheusEndpoint(app, aclHandler)
	registerCatchAllRoute(app, trackPac)

	for _, path := range []string{"/admin", "/pac/192.0.2.1", "/metrics"} {
		t.Run(path, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Fatalf("GET %s status = %d, want %d", path, resp.StatusCode, fiber.StatusForbidden)
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), "admin access denied") {
				t.Fatalf("GET %s body = %q, want the ACL denial message", path, body)
			}
		})
	}

	// The PAC catch-all must stay reachable for browsers; with the caches uninitialised it
	// answers 500 instead of being blocked by the admin ACL.
	t.Run("PAC catch-all stays public", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode == fiber.StatusForbidden {
			t.Fatal("GET / must not be blocked by the admin ACL")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "PAC unavailable") {
			t.Fatalf("GET / body = %q, want the PAC unavailable message", body)
		}
	})
}
