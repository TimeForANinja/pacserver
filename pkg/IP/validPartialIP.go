package IP

import (
	"regexp"
	"strings"
)

var partialIPRegex = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){0,3}$`)

func IsValidPartialIP(ip string) bool {
	// Regular expression to validate IP octets
	// it matches 1-4 octets (as long as they don't end with a dot)
	// e.g. 10.0.4 would match as a valid partial IP
	return partialIPRegex.MatchString(ip)
}

// PadPartialIP ensures that the IP is always (at least) 4 octets long
func PadPartialIP(ip string) string {
	octets := strings.Split(ip, ".")
	for len(octets) < 4 {
		octets = append(octets, "0")
	}
	return strings.Join(octets, ".")
}
