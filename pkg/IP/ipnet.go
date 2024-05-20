package IP

import "strconv"

type IPNet struct {
	NetworkAddress IP   `json:"network_address"`
	CIDR           CIDR `json:"cidr"`
}

func newIPNet(ip IP, cidr CIDR) IPNet {
	return IPNet{
		NetworkAddress: IP{
			// calculate net address
			Value: ip.Value & cidr.Mask,
		},
		CIDR: cidr,
	}
}

func NewIPNetFromStr(ip_str string, cidr_str string) (IPNet, error) {
	ip, err := newIP(ip_str)
	if err != nil {
		return IPNet{}, err
	}
	cidr, err := NewCIDRFromString(cidr_str)
	if err != nil {
		return IPNet{}, err
	}
	return newIPNet(ip, cidr), nil
}

func NewIPNetFromMixed(ip_str string, cidr_str int) (IPNet, error) {
	ip, err := newIP(ip_str)
	if err != nil {
		return IPNet{}, err
	}
	cidr, err := NewCIDR(cidr_str)
	if err != nil {
		return IPNet{}, err
	}
	return newIPNet(ip, cidr), nil
}

func (net1 IPNet) ToString() string {
	return net1.NetworkAddress.toString() + "/" + strconv.Itoa(int(net1.GetRawCIDR()))
}

func (net1 IPNet) GetRawCIDR() uint8 {
	return net1.CIDR.Value
}

func (net1 IPNet) IsSubnetOf(net2 IPNet) bool {
	return net2.includesIP(net1.NetworkAddress) && net1.CIDR.Value >= net2.CIDR.Value
}

func (net IPNet) includesIP(ip IP) bool {
	return ip.Value&net.CIDR.Mask == net.NetworkAddress.Value
}

func (net1 IPNet) IsIdentical(net2 IPNet) bool {
	return net1.NetworkAddress.Value == net2.NetworkAddress.Value && net1.CIDR.Value == net2.CIDR.Value
}
