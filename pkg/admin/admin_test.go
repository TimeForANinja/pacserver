package admin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRegisterAdminUIRoute(t *testing.T) {
	t.Run("prompts for secret when missing", func(t *testing.T) {
		app := fiber.New()
		RegisterAdminUIRoute(app, "secret", "/metrics", 8082)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "Enter the admin token") {
			t.Fatal("admin ui response is missing the login prompt")
		}
	})

	t.Run("serves dashboard after login", func(t *testing.T) {
		app := fiber.New()
		RegisterAdminUIRoute(app, "secret", "/metrics", 8082)
		RegisterAdminLoginRoute(app, "secret")

		loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
		loginReq.Header.Set(adminSecretHeader, "secret")
		loginResp, err := app.Test(loginReq)
		if err != nil {
			t.Fatal(err)
		}
		if loginResp.StatusCode != fiber.StatusNoContent {
			t.Fatalf("status = %d, want %d", loginResp.StatusCode, fiber.StatusNoContent)
		}

		var cookie string
		for _, c := range loginResp.Cookies() {
			if c.Name == adminSecretCookie {
				cookie = c.String()
				break
			}
		}
		if cookie == "" {
			t.Fatal("login did not set the admin cookie")
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Cookie", cookie)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "Lookup debug") {
			t.Fatal("admin ui response is missing the dashboard content")
		}
	})
}

func TestRegisterAdminReloadRoute(t *testing.T) {
	t.Run("rejects missing secret", func(t *testing.T) {
		app := fiber.New()
		RegisterAdminReloadRoute(app, "secret", func() error {
			t.Fatal("reload callback should not run")
			return nil
		})

		req := httptest.NewRequest(http.MethodPost, "/admin/reload", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusForbidden)
		}
	})

	t.Run("runs reload with matching secret", func(t *testing.T) {
		app := fiber.New()
		called := false
		RegisterAdminReloadRoute(app, "secret", func() error {
			called = true
			return nil
		})

		req := httptest.NewRequest(http.MethodPost, "/admin/reload", nil)
		req.Header.Set(adminSecretHeader, "secret")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusNoContent {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
		}
		if !called {
			t.Fatal("reload callback was not called")
		}
	})

	t.Run("runs reload with auth cookie", func(t *testing.T) {
		app := fiber.New()
		called := false
		RegisterAdminReloadRoute(app, "secret", func() error {
			called = true
			return nil
		})

		req := httptest.NewRequest(http.MethodPost, "/admin/reload", nil)
		req.AddCookie(&http.Cookie{Name: adminSecretCookie, Value: "secret"})
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusNoContent {
			t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusNoContent)
		}
		if !called {
			t.Fatal("reload callback was not called")
		}
	})
}
