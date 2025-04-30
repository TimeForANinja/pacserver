package internal

import (
	"testing"
)

func TestCompareForSort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		x1   *ipMap
		x2   *ipMap
		want bool
	}{
		{
			name: "Different network addresses - x1 < x2",
			x1: &ipMap{
				IPNet:    forceIPNet("10.0.0.0", 8),
				Filename: "test1.pac",
			},
			x2: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 16),
				Filename: "test2.pac",
			},
			want: true,
		},
		{
			name: "Different network addresses - x1 > x2",
			x1: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 16),
				Filename: "test1.pac",
			},
			x2: &ipMap{
				IPNet:    forceIPNet("10.0.0.0", 8),
				Filename: "test2.pac",
			},
			want: false,
		},
		{
			name: "Same network address, different CIDR - x1 < x2",
			x1: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 16),
				Filename: "test1.pac",
			},
			x2: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test2.pac",
			},
			want: true,
		},
		{
			name: "Same network address, different CIDR - x1 > x2",
			x1: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test1.pac",
			},
			x2: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 16),
				Filename: "test2.pac",
			},
			want: false,
		},
		{
			name: "Same network address and CIDR",
			x1: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test1.pac",
			},
			x2: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test2.pac",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.x1.CompareForSort(tt.x2)
			if got != tt.want {
				t.Errorf("CompareForSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIPMapLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		line    string
		want    *ipMap
		wantErr bool
	}{
		{
			name:    "Comment line with //",
			line:    "// This is a comment",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Comment line with #",
			line:    "# This is a comment",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Empty line",
			line:    "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Valid line",
			line:    "192.168.0.0,24,test.pac",
			want:    &ipMap{IPNet: forceIPNet("192.168.0.0", 24), Filename: "test.pac"},
			wantErr: false,
		},
		{
			name:    "Valid line with whitespace",
			line:    " 192.168.0.0 , 24 , test.pac ",
			want:    &ipMap{IPNet: forceIPNet("192.168.0.0", 24), Filename: "test.pac"},
			wantErr: false,
		},
		{
			name:    "Invalid number of fields (<3)",
			line:    "192.168.0.0,24",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid number of fields (<3)",
			line:    "192.168.0.0,24,example.com,a comment",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid IP address",
			line:    "invalid,24,test.pac",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid CIDR",
			line:    "192.168.0.0,invalid,test.pac",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIPMapLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIPMapLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil && got == nil {
				// Both are nil, which is expected
				return
			}

			if tt.want == nil || got == nil {
				t.Errorf("parseIPMapLine() = %v, want %v", got, tt.want)
				return
			}

			// Compare the IPNet fields
			if !got.IPNet.IsIdentical(tt.want.IPNet) {
				t.Errorf("parseIPMapLine() IPNet = %v, want %v", got.IPNet, tt.want.IPNet)
			}

			// Compare the Filename fields
			if got.Filename != tt.want.Filename {
				t.Errorf("parseIPMapLine() Filename = %v, want %v", got.Filename, tt.want.Filename)
			}
		})
	}
}
