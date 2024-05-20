package IP

import "regexp"

var partialIPRegex *regexp.Regexp = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){0,3}$`)

func IsValidPartialIP(ip string) bool {
	// Regular expression to validate IP octets
	// it matches 1-4 octests (as long as they don't end with a dot)
	// e.g. 10.0.4 would match as a valid partial IP
	return partialIPRegex.MatchString(ip)
}
