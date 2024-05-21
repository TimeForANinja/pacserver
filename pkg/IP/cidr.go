package IP

import (
	"errors"
	"strconv"
)

var ErrCIDRNotAnInt = errors.New("invalid cidr format - not an integer")
var ErrCIDROutOfRange = errors.New("invalid cidr format - out of range")

var (
	MASK_0  uint32 = 0b00000000_00000000_00000000_00000000
	MASK_8  uint32 = 0b11111111_00000000_00000000_00000000
	MASK_16 uint32 = 0b11111111_11111111_00000000_00000000
	MASK_24 uint32 = 0b11111111_11111111_11111111_00000000
	MASK_32 uint32 = 0b11111111_11111111_11111111_11111111
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

func NewCIDR(cidr_int int) (CIDR, error) {
	if !isValidCIDR(cidr_int) {
		return CIDR{}, ErrCIDROutOfRange
	}
	return CIDR{
		Value: uint8(cidr_int),
		Mask:  cidrToNetmask(cidr_int),
	}, nil
}

func NewCIDRFromString(cidr_str string) (CIDR, error) {
	// (try to) read cidr
	cidr_int, err := strconv.Atoi(cidr_str)
	if err != nil {
		return CIDR{}, ErrCIDRNotAnInt
	}
	return NewCIDR(cidr_int)
}
