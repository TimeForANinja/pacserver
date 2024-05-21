package IP

import "regexp"

var ipRegex = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`)

func IsValidIP(ip string) bool {
	// Regular expression to validate IP octets
	// it matches 1-4 octets (as long as they don't end with a dot)
	// e.g. 10.0.4 would match as a valid partial IP
	return ipRegex.MatchString(ip)
}
