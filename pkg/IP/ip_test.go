package IP

import (
	"testing"
)

func TestNewIP(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expectErr bool
		expected  IP
	}{
		{
			name:      "InvalidFormat",
			input:     "300.155.22.47",
			expectErr: true,
		},
		{
			name:      "ValidIP",
			input:     "192.168.1.1",
			expectErr: false,
			expected:  IP{3232235777},
		},
		{
			name:      "EmptyString",
			input:     "",
			expectErr: true,
		},
		{
			name:      "NonNumericParts",
			input:     "192.abc.1.d",
			expectErr: true,
		},
		{
			name:      "PartialIP",
			input:     "192.168.2",
			expectErr: true,
		},
		{
			name:      "IPv6ShouldError",
			input:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newIP(tc.input)
			if (err != nil) != tc.expectErr {
				t.Fatalf("newIP() error = %v, expectErr %v", err, tc.expectErr)
				return
			}
			if got.Value != tc.expected.Value {
				t.Errorf("newIP() got = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestIP_toString(t *testing.T) {
	tests := []struct {
		name string
		ip   IP
		want string
	}{
		{
			name: "AllZeros",
			ip:   IP{Value: 0},
			want: "0.0.0.0",
		},
		{
			name: "AllOnes",
			ip:   IP{Value: 4294967295},
			want: "255.255.255.255",
		},
		{
			name: "FirstByte255",
			ip:   IP{Value: 4278190080},
			want: "255.0.0.0",
		},
		{
			name: "SecondByte255",
			ip:   IP{Value: 16711680},
			want: "0.255.0.0",
		},
		{
			name: "ThirdByte255",
			ip:   IP{Value: 65280},
			want: "0.0.255.0",
		},
		{
			name: "FourthByte255",
			ip:   IP{Value: 255},
			want: "0.0.0.255",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ip.toString(); got != tt.want {
				t.Errorf("IP.toString() = %v, want %v", got, tt.want)
			}
		})
	}
}
