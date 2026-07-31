package internal

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func parseAdminACLs(value string) ([]netip.Prefix, error) {
	fields := strings.Split(value, ",")
	prefixes := make([]netip.Prefix, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(field)
		if err != nil {
			return nil, fmt.Errorf("%q is not an IP network", field)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	if len(prefixes) == 0 {
		return nil, fmt.Errorf("at least one network must be configured")
	}
	return prefixes, nil
}

func adminACLMiddleware(value string) fiber.Handler {
	prefixes, err := parseAdminACLs(value)
	return func(c *fiber.Ctx) error {
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "invalid admin ACL configuration")
		}
		if adminSourceAllowed(c.IP(), prefixes) {
			return c.Next()
		}
		return fiber.NewError(fiber.StatusForbidden, "admin access denied")
	}
}

func adminSourceAllowed(value string, prefixes []netip.Prefix) bool {
	source, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	source = source.Unmap()
	for _, prefix := range prefixes {
		if prefix.Contains(source) {
			return true
		}
	}
	return false
}
