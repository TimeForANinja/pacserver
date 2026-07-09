package internal

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	accessLogLocalSettledIP = "access_log_settled_ip"
	accessLogLocalXFF       = "access_log_xff"
	accessLogLocalPAC       = "access_log_pac"
)

var accessLogMu sync.Mutex

type accessLogEntry struct {
	Timestamp     string `json:"timestamp"`
	Method        string `json:"method"`
	Path          string `json:"path"`
	Status        int    `json:"status"`
	LatencyMS     int64  `json:"latency_ms"`
	SettledIP     string `json:"settled_ip"`
	XForwardedFor string `json:"xff,omitempty"`
	ServedPAC     string `json:"served_pac,omitempty"`
}

// accessLogMiddleware records one JSON-line access log entry after each request completes.
func accessLogMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		writeAccessLogEntry(c, start)
		return err
	}
}

// setAccessLogFields stores the resolved client IP and raw X-Forwarded-For header on the request context.
func setAccessLogFields(c *fiber.Ctx, settledIP string) {
	if c == nil {
		return
	}

	c.Locals(accessLogLocalSettledIP, settledIP)
	c.Locals(accessLogLocalXFF, c.Get(fiber.HeaderXForwardedFor))
}

// setAccessLogPAC stores the PAC filename that will be returned for the current request.
func setAccessLogPAC(c *fiber.Ctx, pac string) {
	if c == nil {
		return
	}

	c.Locals(accessLogLocalPAC, pac)
}

// writeAccessLogEntry serializes the request metadata and appends it to the access log.
func writeAccessLogEntry(c *fiber.Ctx, start time.Time) {
	if c == nil {
		return
	}

	entry := buildAccessLogEntry(c, start)

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	line = append(line, '\n')

	accessLogMu.Lock()
	defer accessLogMu.Unlock()

	_, _ = getAccessLogger().Write(line)
}

// buildAccessLogEntry collects the request and response details that belong in the access log.
func buildAccessLogEntry(c *fiber.Ctx, start time.Time) accessLogEntry {
	entry := accessLogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
		Method:        c.Method(),
		Path:          c.Path(),
		Status:        c.Response().StatusCode(),
		LatencyMS:     time.Since(start).Milliseconds(),
		SettledIP:     localString(c, accessLogLocalSettledIP),
		XForwardedFor: localString(c, accessLogLocalXFF),
		ServedPAC:     localString(c, accessLogLocalPAC),
	}

	if entry.SettledIP == "" {
		entry.SettledIP = c.IP()
	}

	return entry
}

// localString returns a string request-local value when one has been stored.
func localString(c *fiber.Ctx, key string) string {
	if c == nil {
		return ""
	}

	switch v := c.Locals(key).(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
