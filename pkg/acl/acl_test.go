package acl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestParse(t *testing.T) {
	t.Run("parses a list of networks", func(t *testing.T) {
		prefixes, err := Parse([]string{"192.0.2.0/24", " 10.0.0.5/32 "})
		if err != nil {
			t.Fatal(err)
		}
		if len(prefixes) != 2 {
			t.Fatalf("parsed %d prefixes, want 2", len(prefixes))
		}
	})

	t.Run("rejects an empty list", func(t *testing.T) {
		if _, err := Parse(nil); err == nil {
			t.Fatal("expected an error for an empty network list")
		}
	})

	t.Run("rejects an invalid network", func(t *testing.T) {
		if _, err := Parse([]string{"not-a-network"}); err == nil {
			t.Fatal("expected an error for an invalid network")
		}
	})
}

func TestSourceAllowed(t *testing.T) {
	prefixes, err := Parse([]string{"192.0.2.0/24", "10.0.0.5/32"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		source string
		want   bool
	}{
		{"192.0.2.42", true}, {"10.0.0.5", true}, {"10.0.0.6", false},
	} {
		if got := sourceAllowed(tt.source, prefixes); got != tt.want {
			t.Errorf("source %s: allowed %v, want %v", tt.source, got, tt.want)
		}
	}
}

func TestMiddleware(t *testing.T) {
	t.Run("rejects an empty list", func(t *testing.T) {
		if _, err := Middleware(nil); err == nil {
			t.Fatal("expected an error for an empty network list")
		}
	})

	t.Run("rejects an invalid network", func(t *testing.T) {
		if _, err := Middleware([]string{"not-a-network"}); err == nil {
			t.Fatal("expected an error for an invalid network")
		}
	})

	t.Run("allows listed sources and denies the rest", func(t *testing.T) {
		handler, err := Middleware([]string{"192.0.2.0/24", "10.0.0.5/32"})
		if err != nil {
			t.Fatal(err)
		}
		// fiber's app.Test cannot set the socket remote address, so the source IP is
		// simulated through X-Forwarded-For, which c.IP() returns when configured.
		app := fiber.New(fiber.Config{ProxyHeader: fiber.HeaderXForwardedFor})
		app.Get("/protected", handler, func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})
		for _, tt := range []struct {
			source     string
			wantStatus int
		}{
			{"192.0.2.42", fiber.StatusOK},
			{"10.0.0.5", fiber.StatusOK},
			{"10.0.0.6", fiber.StatusForbidden},
		} {
			t.Run(tt.source, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set(fiber.HeaderXForwardedFor, tt.source)
				resp, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != tt.wantStatus {
					t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
				}
			})
		}
	})
}
