package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
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
