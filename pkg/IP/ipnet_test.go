package IP

import (
	"reflect"
	"testing"
)

func TestNewIPNetFromStr(t *testing.T) {
	tests := []struct {
		name      string
		ip_str    string
		cidr_str  string
		wantIPNet IPNet
		wantErr   bool
	}{
		{
			name:     "valid IP and CIDR",
			ip_str:   "192.168.0.0",
			cidr_str: "24",
			wantIPNet: IPNet{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: MASK_24},
			},
			wantErr: false,
		},
		{
			name:     "properly normalizes ip to network address",
			ip_str:   "192.168.0.127",
			cidr_str: "24",
			wantIPNet: IPNet{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: MASK_24},
			},
			wantErr: false,
		},
		{
			name:     "invalid IP",
			ip_str:   "192.168.0.abc",
			cidr_str: "24",
			wantErr:  true,
		},
		{
			name:     "invalid CIDR",
			ip_str:   "192.168.0.1",
			cidr_str: "abc",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIPNet, err := NewIPNetFromStr(tt.ip_str, tt.cidr_str)
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
		ip_str   string
		cidr_str int
	}
	tests := []struct {
		name    string
		args    args
		want    IPNet
		wantErr bool
	}{
		{
			name: "ValidIPv4",
			args: args{
				ip_str:   "192.168.0.0",
				cidr_str: 24,
			},
			want: IPNet{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: MASK_24},
			},
			wantErr: false,
		},
		{
			name: "properly normalizes ip to network address",
			args: args{
				ip_str:   "192.168.0.127",
				cidr_str: 24,
			},
			want: IPNet{
				NetworkAddress: IP{Value: 3232235520},
				CIDR:           CIDR{Value: 24, Mask: MASK_24},
			},
			wantErr: false,
		},
		{
			name: "IPv6ShouldError",
			args: args{
				ip_str:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				cidr_str: 64,
			},
			want:    IPNet{},
			wantErr: true,
		},
		{
			name: "InvalidIP",
			args: args{
				ip_str:   "300.168.0.1",
				cidr_str: 24,
			},
			want:    IPNet{},
			wantErr: true,
		},
		{
			name: "InvalidCIDR",
			args: args{
				ip_str:   "192.168.0.1",
				cidr_str: 34,
			},
			want:    IPNet{},
			wantErr: true,
		},
		{
			name: "InvalidIPAndCIDR",
			args: args{
				ip_str:   "300.168.0.1",
				cidr_str: 34,
			},
			want:    IPNet{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIPNetFromMixed(tt.args.ip_str, tt.args.cidr_str)
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
