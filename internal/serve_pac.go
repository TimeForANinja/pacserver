package internal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/internal/storage"
	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/IPLUT"
)

// serveLookupRequest resolves the request IP against the LUT and returns the matching PAC.
func serveLookupRequest(c *fiber.Ctx, trackPac func(pac *storage.LookupEntry), forceDebug bool) error {
	if c == nil {
		return fiber.NewError(fiber.StatusBadRequest, "missing request context")
	}

	// 1. Extract all request facts without making a routing decision.
	features := extractRequestFeatures(c, forceDebug)
	// 2. Choose exactly one lookup key. IP and path deliberately remain separate.
	ipStr, path := requestVerdict(features)
	// 3. Fetch the PAC from the corresponding lookup table.
	var pac *storage.LookupEntry
	var ipNet *IP.Net
	var stackTrace []*storage.LookupEntry
	if path != "" {
		pac = storage.FindRoute(path)
		ipNet = &IP.Net{}
	} else {
		pac, ipNet, stackTrace = storage.FindInLUT(ipStr, features.RouteNetworkBits)
	}
	// 4. Record the inputs, decision, and result before writing the response.
	pacName := ""
	if pac != nil && pac.PAC != nil {
		pacName = pac.PAC.Filename
	}
	log.Debugf("PAC lookup features=%+v verdict_ip=%q verdict_path=%q pac=%q", features, ipStr, path, pacName)
	// 5. Serve the selected PAC.
	return servePAC(c, pac, stackTrace, ipNet, ipStr, features.RouteNetworkBits, trackPac, features.Debug)
}

type requestFeatures struct {
	SourceIP         string
	XForwardedForIP  string
	RouteIP          string
	RouteNetworkBits int
	Debug            bool
	RawRoute         string
}

func extractRequestFeatures(c *fiber.Ctx, forceDebug bool) requestFeatures {
	f := requestFeatures{RouteNetworkBits: 32, Debug: forceDebug}
	if c == nil {
		return f
	}
	f.RawRoute = c.Path()
	if source := strings.TrimSpace(c.IP()); IP.IsValidIP(source) {
		f.SourceIP = source
	}
	f.XForwardedForIP = extractXForwardedFor(c)
	if ip, bits, ok := extractURLIP(c); ok {
		f.RouteIP, f.RouteNetworkBits = ip, bits
	}
	for key := range c.Queries() {
		if strings.EqualFold(key, "debug") {
			f.Debug = true
			break
		}
	}
	return f
}

// requestVerdict returns either an IP or a path, never both.
func requestVerdict(f requestFeatures) (string, string) {
	if f.RouteIP != "" {
		return f.RouteIP, ""
	}
	if route := "/" + strings.Trim(strings.TrimSpace(f.RawRoute), "/"); route != "/" {
		return "", route
	}
	if f.XForwardedForIP != "" {
		return f.XForwardedForIP, ""
	}
	return f.SourceIP, ""
}

// servePAC writes the resolved PAC response and optionally returns debug metadata.
func servePAC(
	c *fiber.Ctx,
	pac *storage.LookupEntry,
	stackTrace []*storage.LookupEntry,
	ipNet *IP.Net,
	ipStr string,
	networkBits int,
	trackPac func(pac *storage.LookupEntry),
	forceDebug bool,
) error {
	if c == nil {
		return fiber.NewError(fiber.StatusBadRequest, "missing request context")
	}

	setAccessLogFields(c, ipStr)
	setAccessLogPAC(c, "")
	if pac != nil && pac.PAC != nil {
		setAccessLogPAC(c, pac.PAC.Filename)
	}

	// Count the served PAC before formatting the response so metrics match the actual reply.
	if trackPac != nil {
		trackPac(pac)
	}

	if pac == nil {
		LogUnexpectedError(fmt.Sprintf("missing PAC for %s/%d", ipStr, networkBits), fmt.Errorf("lookup returned nil"))
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.Status(fiber.StatusInternalServerError).SendString("PAC unavailable")
	}

	// Debug output is opt-in via query parameter so normal responses stay lightweight.
	hasDebug := false
	for key := range c.Queries() {
		if strings.EqualFold(key, "debug") {
			hasDebug = true
			break
		}
	}

	if hasDebug || forceDebug {
		// In debug mode, return the resolved request, the matching path, and the final PAC body.
		matchedIP := ""
		if ipNet != nil {
			matchedIP = ipNet.ToString()
		}
		pacMeta, err := json.MarshalIndent(fiber.Map{
			"requested_ip": fmt.Sprintf("%s/%d", ipStr, networkBits),
			"matched_ip":   matchedIP,
			"matched_rule": pac.Stringify(),
		}, "", "\t")
		if err != nil {
			log.Errorf("Error marshaling debug JSON: %v", err)
			return err
		}

		treeMeta := IPLUT.StringifyStack(stackTrace)

		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.SendString(strings.Join([]string{
			string(pacMeta),
			treeMeta,
			pac.Variant,
		}, "\n\n---------------------------------------\n\n"))
	}

	c.Set("content-type", "application/x-ns-proxy-autoconfig")
	return c.SendString(pac.Variant)
}

// extractIP determines the client IP from the request path, forwarded headers, or socket address.
func extractIP(c *fiber.Ctx) (string, int) {
	if c == nil {
		return "", 32
	}

	// URL parameters take priority because they are explicit and easy to test.
	if ipStr, bits, ok := extractURLIP(c); ok {
		return ipStr, bits
	}

	// If the URL does not specify an IP, fall back to the forwarded client address.
	log.Debug("Headers:  ", c.GetReqHeaders())
	if ipStr := extractXForwardedFor(c); ipStr != "" {
		return ipStr, 32
	}

	// As a last resort, use the direct remote address from the connection.
	if ipStr := strings.TrimSpace(c.IP()); IP.IsValidIP(ipStr) {
		return ipStr, 32
	}

	return "", 32
}

// extractURLIP parses an explicit IP or IP/CIDR path segment from the request URL.
func extractURLIP(c *fiber.Ctx) (string, int, bool) {
	if c == nil {
		return "", 32, false
	}

	// The route can include an explicit ip/netmask input
	path := strings.TrimSpace(strings.Trim(c.Path(), "/"))
	if path == "" {
		return "", 32, false
	}
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return "", 32, false
	}

	// The first segment is always the partial IP.
	ipStr := strings.TrimSpace(segments[0])
	if !IP.IsValidPartialIP(ipStr) {
		return "", 32, false
	}

	// If a second segment exists, treat it as the explicit CIDR mask.
	networkBits := len(strings.Split(ipStr, ".")) * 8
	if len(segments) > 1 {
		if cidr, err := strconv.Atoi(strings.TrimSpace(segments[1])); err == nil {
			networkBits = cidr
		}
	}

	return IP.PadPartialIP(ipStr), networkBits, true
}

// extractXForwardedFor returns the first valid client IP from the forwarded header chain.
func extractXForwardedFor(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	// Use only the first hop from X-Forwarded-For, since that is the client we care about.
	xff := strings.TrimSpace(c.Get("X-Forwarded-For"))
	log.Debug("xff:     ", xff)
	if xff == "" {
		return ""
	}

	// the xff is a list of elements, usually separated by "," but we support a few others
	separators := []string{",", "-", ";"}
	for _, sep := range separators {
		firstElement := strings.TrimSpace(strings.Split(xff, sep)[0])
		log.Debug("element: ", firstElement)
		if IP.IsValidIP(firstElement) {
			return firstElement
		}
	}

	return ""
}
