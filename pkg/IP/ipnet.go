package IP

import "strconv"

type Net struct {
	NetworkAddress IP   `json:"network_address"`
	CIDR           CIDR `json:"cidr"`
}

func newIPNet(ip IP, cidr CIDR) Net {
	return Net{
		NetworkAddress: IP{
			// calculate net address
			Value: ip.Value & cidr.Mask,
		},
		CIDR: cidr,
	}
}

func NewIPNetFromStr(ipStr string, cidrStr string) (Net, error) {
	ip, err := newIP(ipStr)
	if err != nil {
		return Net{}, err
	}
	cidr, err := NewCIDRFromString(cidrStr)
	if err != nil {
		return Net{}, err
	}
	return newIPNet(ip, cidr), nil
}

func NewIPNetFromMixed(ipStr string, cidrStr int) (Net, error) {
	ip, err := newIP(ipStr)
	if err != nil {
		return Net{}, err
	}
	cidr, err := NewCIDR(cidrStr)
	if err != nil {
		return Net{}, err
	}
	return newIPNet(ip, cidr), nil
}

func (net1 Net) ToString() string {
	return net1.NetworkAddress.toString() + "/" + strconv.Itoa(int(net1.GetRawCIDR()))
}

func (net1 Net) GetRawCIDR() uint8 {
	return net1.CIDR.Value
}

func (net1 Net) IsSubnetOf(net2 Net) bool {
	return net2.includesIP(net1.NetworkAddress) && net1.CIDR.Value >= net2.CIDR.Value
}

func (net1 Net) includesIP(ip IP) bool {
	return ip.Value&net1.CIDR.Mask == net1.NetworkAddress.Value
}

func (net1 Net) IsIdentical(net2 Net) bool {
	return net1.NetworkAddress.Value == net2.NetworkAddress.Value && net1.CIDR.Value == net2.CIDR.Value
}
