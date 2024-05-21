package IP

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type IP struct {
	Value uint32 `json:"value"`
}

var ErrInvalidIPFormat = errors.New("invalid ip format")

func newIP(srcIP string) (IP, error) {
	if !IsValidIP(srcIP) {
		return IP{}, ErrInvalidIPFormat
	}

	parts := strings.Split(srcIP, ".")
	var ip uint32
	for _, part := range parts {
		p, err := strconv.Atoi(part)
		if err != nil {
			// this error should be impossible due to prior validation
			return IP{}, err
		}
		ip = ip<<8 | uint32(p)
	}

	return IP{Value: ip}, nil
}

func (ip IP) toString() string {
	// Extract each byte from the uint32 IP value
	byte1 := ip.Value >> 24 & 0xFF
	byte2 := ip.Value >> 16 & 0xFF
	byte3 := ip.Value >> 8 & 0xFF
	byte4 := ip.Value & 0xFF

	// Format the bytes into the standard IP address format
	return fmt.Sprintf("%d.%d.%d.%d", byte1, byte2, byte3, byte4)
}
