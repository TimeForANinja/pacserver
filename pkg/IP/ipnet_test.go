package IP

import (
	"reflect"
	"testing"
)

func TestNewIPNetFromStr(t *testing.T) {
	tests := []struct {
		name      string
		ipStr     string
		cidrStr   string
		wantIPNet Net
		wantErr   bool
	}{
		{
			name:    "valid IP and CIDR",
			ipStr:   "192.168.0.0",
			cidrStr: "24",
			wantIPNet: Net{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			wantErr: false,
		},
		{
			name:    "properly normalizes ip to network address",
			ipStr:   "192.168.0.127",
			cidrStr: "24",
			wantIPNet: Net{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ipStr:   "192.168.0.abc",
			cidrStr: "24",
			wantErr: true,
		},
		{
			name:    "invalid CIDR",
			ipStr:   "192.168.0.1",
			cidrStr: "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIPNet, err := NewIPNetFromStr(tt.ipStr, tt.cidrStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIPNetFromStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotIPNet, tt.wantIPNet) {
				t.Errorf("NewIPNetFromStr() = %v, want %v", gotIPNet, tt.wantIPNet)
			}
		})
	}
}

func TestNewIPNetFromMixed(t *testing.T) {
	type args struct {
		ipStr   string
		cidrStr int
	}
	tests := []struct {
		name    string
		args    args
		want    Net
		wantErr bool
	}{
		{
			name: "ValidIPv4",
			args: args{
				ipStr:   "192.168.0.0",
				cidrStr: 24,
			},
			want: Net{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			wantErr: false,
		},
		{
			name: "properly normalizes ip to network address",
			args: args{
				ipStr:   "192.168.0.127",
				cidrStr: 24,
			},
			want: Net{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			wantErr: false,
		},
		{
			name: "IPv6ShouldError",
			args: args{
				ipStr:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				cidrStr: 64,
			},
			want:    Net{},
			wantErr: true,
		},
		{
			name: "InvalidIP",
			args: args{
				ipStr:   "300.168.0.1",
				cidrStr: 24,
			},
			want:    Net{},
			wantErr: true,
		},
		{
			name: "InvalidCIDR",
			args: args{
				ipStr:   "192.168.0.1",
				cidrStr: 34,
			},
			want:    Net{},
			wantErr: true,
		},
		{
			name: "InvalidIPAndCIDR",
			args: args{
				ipStr:   "300.168.0.1",
				cidrStr: 34,
			},
			want:    Net{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIPNetFromMixed(tt.args.ipStr, tt.args.cidrStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIPNetFromMixed() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIPNetFromMixed() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestIsSubnetOf(t *testing.T) {
	tests := []struct {
		name        string
		net1IP      string
		net1CIDR    int
		net2IP      string
		net2CIDR    int
		expectedRes bool
	}{
		{
			name:        "subnetAllEqual",
			net1IP:      "10.0.0.0",
			net1CIDR:    24,
			net2IP:      "10.0.0.0",
			net2CIDR:    24,
			expectedRes: true,
		},
		{
			name:        "subnetWithinAndSameStart",
			net1IP:      "10.0.0.0",
			net1CIDR:    26,
			net2IP:      "10.0.0.0",
			net2CIDR:    24,
			expectedRes: true,
		},
		{
			name:        "subnetWithinAndOffsetStart",
			net1IP:      "10.0.0.64",
			net1CIDR:    26,
			net2IP:      "10.0.0.0",
			net2CIDR:    24,
			expectedRes: true,
		},
		{
			name:        "subnetOutside",
			net1IP:      "192.168.0.0",
			net1CIDR:    24,
			net2IP:      "10.0.0.0",
			net2CIDR:    24,
			expectedRes: false,
		},
		{
			name:        "subnetBigger",
			net1IP:      "10.0.0.0",
			net1CIDR:    22,
			net2IP:      "10.0.0.0",
			net2CIDR:    24,
			expectedRes: false,
		},
		{
			name:        "subnetDifferentNet",
			net1IP:      "10.0.0.0",
			net1CIDR:    22,
			net2IP:      "10.10.10.0",
			net2CIDR:    24,
			expectedRes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			net1, err1 := NewIPNetFromMixed(tt.net1IP, tt.net1CIDR)
			net2, err2 := NewIPNetFromMixed(tt.net2IP, tt.net2CIDR)
			if err1 != nil || err2 != nil {
				t.Errorf("Error in creating IPNet: net1:%v net2:%v", err1, err2)
				return
			}

			res := net1.IsSubnetOf(net2)
			if res != tt.expectedRes {
				t.Errorf("expected %v, got %v", tt.expectedRes, res)
			}
		})
	}
}

func TestIncludesIP(t *testing.T) {
	tests := []struct {
		name     string
		network  string
		cidr     string
		testIP   string
		expected bool
	}{
		{"IP is included in subnet", "192.168.0.0", "24", "192.168.0.10", true},
		{"IP is not included in subnet", "192.168.0.0", "24", "192.168.1.10", false},
		{"Subnet boundary, lower edge", "192.168.1.0", "24", "192.168.1.0", true},
		{"Subnet boundary, upper edge", "192.168.1.0", "24", "192.168.1.255", true},
		{"Subnet all zeroes", "0.0.0.0", "0", "192.168.0.1", true},
		{"Subnet all ones", "255.255.255.255", "32", "192.168.0.1", false},
		{"CIDR is 32, IP equals subnet", "192.168.0.100", "32", "192.168.0.100", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			net, errNet := NewIPNetFromStr(test.network, test.cidr)
			ip, errIP := newIP(test.testIP)
			if errNet != nil || errIP != nil {
				t.Errorf("Error in creating IPNet: %v, Error in creating IP: %v", errNet, errIP)
			}

			result := net.includesIP(ip)
			if result != test.expected {
				t.Errorf("Expected %v but got %v for test case '%s'", test.expected, result, test.name)
			}
		})
	}
}

func TestIsIdentical(t *testing.T) {
	tests := []struct {
		name     string
		net1IP   string
		net1CIDR int
		net2IP   string
		net2CIDR int
		want     bool
	}{
		{
			name:     "IdenticalIPNets",
			net1IP:   "10.0.0.0",
			net1CIDR: 8,
			net2IP:   "10.0.0.0",
			net2CIDR: 8,
			want:     true,
		},
		{
			name: "DifferentNetworkAddress",
			// gets normalised when creating the IPNet Object
			net1IP:   "192.168.0.0",
			net1CIDR: 16,
			net2IP:   "192.168.50.0",
			net2CIDR: 16,
			want:     true,
		},
		{
			name:     "DifferentCIDR",
			net1IP:   "192.168.0.0",
			net1CIDR: 8,
			net2IP:   "192.168.0.0",
			net2CIDR: 16,
			want:     false,
		},
		{
			name:     "BothValuesDifferent",
			net1IP:   "192.168.0.0",
			net1CIDR: 8,
			net2IP:   "10.0.0.1",
			net2CIDR: 16,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			net1, err1 := NewIPNetFromMixed(tt.net1IP, tt.net1CIDR)
			net2, err2 := NewIPNetFromMixed(tt.net2IP, tt.net2CIDR)
			if err1 != nil || err2 != nil {
				t.Errorf("Error in creating IPNet: net1:%v net2:%v", err1, err2)
				return
			}

			if got := net1.IsIdentical(net2); got != tt.want {
				t.Errorf("IsIdentical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		ipNet    Net
		expected string
	}{
		{
			name: "Class A network",
			ipNet: Net{
				NetworkAddress: IP{Value: 167772160}, // 10.0.0.0
				CIDR:           CIDR{Value: 8, Mask: Mask8},
			},
			expected: "10.0.0.0/8",
		},
		{
			name: "Class B network",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235520}, // 192.168.0.0
				CIDR:           CIDR{Value: 16, Mask: Mask16},
			},
			expected: "192.168.0.0/16",
		},
		{
			name: "Class C network",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235520}, // 192.168.0.0
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			expected: "192.168.0.0/24",
		},
		{
			name: "Host address",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235521}, // 192.168.0.1
				CIDR:           CIDR{Value: 32, Mask: Mask32},
			},
			expected: "192.168.0.1/32",
		},
		{
			name: "Zero address",
			ipNet: Net{
				NetworkAddress: IP{Value: 0}, // 0.0.0.0
				CIDR:           CIDR{Value: 0, Mask: Mask0},
			},
			expected: "0.0.0.0/0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ipNet.ToString()
			if result != tt.expected {
				t.Errorf("ToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetRawCIDR(t *testing.T) {
	tests := []struct {
		name     string
		ipNet    Net
		expected uint8
	}{
		{
			name: "CIDR 8",
			ipNet: Net{
				NetworkAddress: IP{Value: 167772160}, // 10.0.0.0
				CIDR:           CIDR{Value: 8, Mask: Mask8},
			},
			expected: 8,
		},
		{
			name: "CIDR 16",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235520}, // 192.168.0.0
				CIDR:           CIDR{Value: 16, Mask: Mask16},
			},
			expected: 16,
		},
		{
			name: "CIDR 24",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235520}, // 192.168.0.0
				CIDR:           CIDR{Value: 24, Mask: Mask24},
			},
			expected: 24,
		},
		{
			name: "CIDR 32",
			ipNet: Net{
				NetworkAddress: IP{Value: 3232235521}, // 192.168.0.1
				CIDR:           CIDR{Value: 32, Mask: Mask32},
			},
			expected: 32,
		},
		{
			name: "CIDR 0",
			ipNet: Net{
				NetworkAddress: IP{Value: 0}, // 0.0.0.0
				CIDR:           CIDR{Value: 0, Mask: Mask0},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ipNet.GetRawCIDR()
			if result != tt.expected {
				t.Errorf("GetRawCIDR() = %v, want %v", result, tt.expected)
			}
		})
	}
}
