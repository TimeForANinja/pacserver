package IP

import (
	"errors"
	"strconv"
)

var ErrCIDRNotAnInt = errors.New("invalid cidr format - not an integer")
var ErrCIDROutOfRange = errors.New("invalid cidr format - out of range")

var (
	Mask0  uint32 = 0b00000000_00000000_00000000_00000000
	Mask8  uint32 = 0b11111111_00000000_00000000_00000000
	Mask16 uint32 = 0b11111111_11111111_00000000_00000000
	Mask24 uint32 = 0b11111111_11111111_11111111_00000000
	Mask32 uint32 = 0b11111111_11111111_11111111_11111111
)

func isValidCIDR(cidr int) bool {
	return cidr >= 0 && cidr <= 32
}

type CIDR struct {
	Value uint8  `json:"value"`
	Mask  uint32 `json:"mask"`
}

func cidrToNetmask(cidr int) uint32 {
	// Calculate netmask by shifting bits
	return ^uint32(0) << (32 - cidr)
}

func NewCIDR(cidrInt int) (CIDR, error) {
	if !isValidCIDR(cidrInt) {
		return CIDR{}, ErrCIDROutOfRange
	}
	return CIDR{
		Value: uint8(cidrInt),
		Mask:  cidrToNetmask(cidrInt),
	}, nil
}

func NewCIDRFromString(cidrStr string) (CIDR, error) {
	// (try to) read cidr
	cidrInt, err := strconv.Atoi(cidrStr)
	if err != nil {
		return CIDR{}, ErrCIDRNotAnInt
	}
	return NewCIDR(cidrInt)
}
