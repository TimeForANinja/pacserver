package admin

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const adminSecretHeader = "X-Admin-Secret"
const adminSecretQuery = "secret"
const adminSecretCookie = "pacserver_admin_secret"

func getPresentedSecret(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	if headerSecret := strings.TrimSpace(c.Get(adminSecretHeader)); headerSecret != "" {
		return headerSecret
	}

	if querySecret := strings.TrimSpace(c.Query(adminSecretQuery)); querySecret != "" {
		return querySecret
	}

	return strings.TrimSpace(c.Cookies(adminSecretCookie))
}

func isAuthorized(secret string, c *fiber.Ctx) bool {
	return secret != "" && getPresentedSecret(c) == secret
}

// RegisterAdminReloadRoute registers the protected admin reload endpoint.
func RegisterAdminReloadRoute(app *fiber.App, secret string, reload func() error) {
	if app == nil {
		return
	}

	// Keep the admin endpoint in one place so the auth rule and reload flow stay consistent.
	app.Post("/admin/reload", func(c *fiber.Ctx) error {
		log.Infof("Admin reload requested from %s", c.IP())
		// Reject empty secrets up front so a misconfigured deployment cannot expose the route.
		if secret == "" {
			log.Warn("Admin reload rejected: admin secret is not configured")
			return fiber.ErrForbidden
		}
		// Require the caller to present the same shared secret as the server config.
		if !isAuthorized(secret, c) {
			log.Warnf("Admin reload rejected from %s: invalid secret", c.IP())
			return fiber.ErrForbidden
		}

		// Allow a registration-only route when the caller just wants the endpoint available.
		if reload == nil {
			return c.SendStatus(fiber.StatusNoContent)
		}

		// Translate reload failures into a clear HTTP error for the admin client.
		if err := reload(); err != nil {
			log.Errorf("Admin reload failed: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "reload failed")
		}

		log.Infof("Admin reload completed successfully from %s", c.IP())
		return c.SendStatus(fiber.StatusNoContent)
	})
}

// RegisterAdminUIRoute registers the protected admin web UI.
func RegisterAdminUIRoute(app *fiber.App, secret string, metricsPath string) {
	if app == nil {
		return
	}

	// Keep the UI behind the same secret as the reload endpoint so both entry points share one gate.
	app.Get("/admin", func(c *fiber.Ctx) error {
		log.Infof("Admin UI requested from %s", c.IP())
		if secret == "" {
			log.Warn("Admin UI rejected: admin secret is not configured")
			return fiber.ErrForbidden
		}

		// If the caller has not provided the token yet, serve a tiny prompt instead of a hard 403.
		if !isAuthorized(secret, c) {
			log.Warnf("Admin UI login prompt shown to %s", c.IP())
			page, err := renderAdminLoginPage()
			if err != nil {
				log.Errorf("Failed to render admin login page: %v", err)
				return fiber.NewError(fiber.StatusInternalServerError, "failed to render admin page")
			}

			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.SendString(page)
		}

		// Once the token is present, render the dashboard and let the page reuse that same secret for reload.
		log.Infof("Admin UI authenticated for %s", c.IP())
		page, err := renderAdminPage(metricsPath)
		if err != nil {
			log.Errorf("Failed to render admin page: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to render admin page")
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(page)
	})
}

// RegisterAdminLoginRoute registers the client-side login endpoint for the admin UI.
func RegisterAdminLoginRoute(app *fiber.App, secret string) {
	if app == nil {
		return
	}

	// Keep login separate from the dashboard so the browser can cache the secret as a cookie.
	app.Post("/admin/login", func(c *fiber.Ctx) error {
		log.Infof("Admin login requested from %s", c.IP())
		if secret == "" {
			log.Warn("Admin login rejected: admin secret is not configured")
			return fiber.ErrForbidden
		}
		if !isAuthorized(secret, c) {
			log.Warnf("Admin login rejected from %s: invalid secret", c.IP())
			return fiber.ErrForbidden
		}

		c.Cookie(&fiber.Cookie{
			Name:     adminSecretCookie,
			Value:    secret,
			Path:     "/",
			HTTPOnly: true,
			SameSite: "Lax",
		})

		log.Infof("Admin login completed successfully for %s", c.IP())
		return c.SendStatus(fiber.StatusNoContent)
	})
}

// TriggerAdminReload sends the reload request to a running server.
func TriggerAdminReload(port uint16, secret string) error {
	// Build the request exactly as the server expects it so the CLI stays tiny.
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/admin/reload", port), nil)
	if err != nil {
		log.Errorf("Admin reload request could not be created: %v\n%s", err, debug.Stack())
		return err
	}

	// Reuse the same shared-secret header that the server checks.
	req.Header.Set(adminSecretHeader, secret)

	// Keep the client simple and short-lived because this call is synchronous.
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Admin reload request failed to send: %v\n%s", err, debug.Stack())
		return err
	}
	defer resp.Body.Close()

	_, _ = io.ReadAll(resp.Body)
	// Treat anything other than 204 as a failed reload.
	if resp.StatusCode != http.StatusNoContent {
		log.Errorf("Admin reload request returned %s\n%s", resp.Status, debug.Stack())
		return fmt.Errorf("reload request failed with status %s", resp.Status)
	}

	return nil
}
