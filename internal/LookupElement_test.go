package internal

import (
	"testing"
)

func TestIsIdenticalNet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		le1  LookupElement
		le2  LookupElement
		want bool
	}{
		{
			name: "Identical networks",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test2.pac",
				},
			},
			want: true,
		},
		{
			name: "Different networks - different IP",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("10.0.0.0", 24),
					Filename: "test2.pac",
				},
			},
			want: false,
		},
		{
			name: "Different networks - different CIDR",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 16),
					Filename: "test2.pac",
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.le1.isIdenticalNet(tt.le2)
			if got != tt.want {
				t.Errorf("isIdenticalNet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIdenticalPAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		le1  LookupElement
		le2  LookupElement
		want bool
	}{
		{
			name: "Identical PAC filenames",
			le1: LookupElement{
				PAC: &pacTemplate{
					Filename: "test.pac",
				},
			},
			le2: LookupElement{
				PAC: &pacTemplate{
					Filename: "test.pac",
				},
			},
			want: true,
		},
		{
			name: "Different PAC filenames",
			le1: LookupElement{
				PAC: &pacTemplate{
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				PAC: &pacTemplate{
					Filename: "test2.pac",
				},
			},
			want: false,
		},
		{
			name: "First PAC is nil",
			le1: LookupElement{
				PAC: nil,
			},
			le2: LookupElement{
				PAC: &pacTemplate{
					Filename: "test.pac",
				},
			},
			want: false,
		},
		{
			name: "Second PAC is nil",
			le1: LookupElement{
				PAC: &pacTemplate{
					Filename: "test.pac",
				},
			},
			le2: LookupElement{
				PAC: nil,
			},
			want: false,
		},
		{
			name: "Both PACs are nil",
			le1: LookupElement{
				PAC: nil,
			},
			le2: LookupElement{
				PAC: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.le1.isIdenticalPAC(tt.le2)
			if got != tt.want {
				t.Errorf("isIdenticalPAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSubnetOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		le1  LookupElement
		le2  LookupElement
		want bool
	}{
		{
			name: "le1 is subnet of le2",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 16),
					Filename: "test2.pac",
				},
			},
			want: true,
		},
		{
			name: "le1 is not subnet of le2 - different network",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("10.0.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 16),
					Filename: "test2.pac",
				},
			},
			want: false,
		},
		{
			name: "le1 is not subnet of le2 - le1 is broader",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 16),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test2.pac",
				},
			},
			want: false,
		},
		{
			name: "le1 is identical to le2",
			le1: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test1.pac",
				},
			},
			le2: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test2.pac",
				},
			},
			want: true, // IsSubnetOf returns true for identical networks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.le1.isSubnetOf(tt.le2)
			if got != tt.want {
				t.Errorf("isSubnetOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRawCIDR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		le   LookupElement
		want uint8
	}{
		{
			name: "CIDR 24",
			le: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 24),
					Filename: "test.pac",
				},
			},
			want: 24,
		},
		{
			name: "CIDR 16",
			le: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("192.168.0.0", 16),
					Filename: "test.pac",
				},
			},
			want: 16,
		},
		{
			name: "CIDR 8",
			le: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("10.0.0.0", 8),
					Filename: "test.pac",
				},
			},
			want: 8,
		},
		{
			name: "CIDR 0",
			le: LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("0.0.0.0", 0),
					Filename: "test.pac",
				},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.le.getRawCIDR()
			if got != tt.want {
				t.Errorf("getRawCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLookupElement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ipMap       *ipMap
		pac         *pacTemplate
		contactInfo string
		wantErr     bool
	}{
		{
			name: "Valid template",
			ipMap: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test.pac",
			},
			pac: &pacTemplate{
				Filename: "test.pac",
				content:  "// This is a test PAC file for {{ .Filename }} by {{ .Contact }}",
			},
			contactInfo: "Test Contact",
			wantErr:     false,
		},
		{
			name: "Invalid template",
			ipMap: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 24),
				Filename: "test.pac",
			},
			pac: &pacTemplate{
				Filename: "test.pac",
				content:  "// This is a test PAC file with {{ .InvalidVariable }}",
			},
			contactInfo: "Test Contact",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			le, err := NewLookupElement(tt.ipMap, tt.pac, tt.contactInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLookupElement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the LookupElement was created correctly
				if le.IPMap != tt.ipMap {
					t.Errorf("NewLookupElement() IPMap = %v, want %v", le.IPMap, tt.ipMap)
				}
				if le.PAC != tt.pac {
					t.Errorf("NewLookupElement() PAC = %v, want %v", le.PAC, tt.pac)
				}

				// Verify the template was filled correctly
				expectedVariant := "// This is a test PAC file for test.pac by Test Contact"
				if le.Variant != expectedVariant {
					t.Errorf("NewLookupElement() Variant = %v, want %v", le.Variant, expectedVariant)
				}
			}
		})
	}
}
