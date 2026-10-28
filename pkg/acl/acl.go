package acl

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Parse converts a list of CIDR networks into masked prefixes.
// Blank entries are skipped and at least one valid network is required.
func Parse(networks []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(networks))
	for _, value := range networks {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("%q is not an IP network", value)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	if len(prefixes) == 0 {
		return nil, fmt.Errorf("at least one network must be configured")
	}
	return prefixes, nil
}

// Middleware builds a fiber handler that lets requests through only when the
// source IP is inside one of the configured networks. Any other request gets
// a 403. An error is returned for an empty or invalid network list so the
// caller can fail fast at startup instead of rejecting traffic per request.
func Middleware(networks []string) (fiber.Handler, error) {
	prefixes, err := Parse(networks)
	if err != nil {
		return nil, err
	}

	return func(c *fiber.Ctx) error {
		if sourceAllowed(c.IP(), prefixes) {
			return c.Next()
		}
		return fiber.NewError(fiber.StatusForbidden, "admin access denied")
	}, nil
}

// sourceAllowed reports whether the source IP matches any of the prefixes.
// IPv4 addresses embedded in IPv6 notation are normalised before the check.
func sourceAllowed(value string, prefixes []netip.Prefix) bool {
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
