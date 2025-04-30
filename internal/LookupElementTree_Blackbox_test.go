package internal

import (
	"testing"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

// TestBuildAndFindCombined tests the combined functionality of buildLookupTree and findInTree
// focusing on edge cases as specified in the requirements
func TestBuildAndFindCombined(t *testing.T) {
	// Setup global variables needed for buildLookupTree
	setupTestEnvironment()

	// Test cases
	tests := []struct {
		name     string
		elements []*LookupElement
		findIP   *IP.Net
		setup    func() // Optional setup function for more complex test cases
	}{
		{
			name:     "No elements at all",
			elements: []*LookupElement{},
			findIP:   createIPNet("192.168.1.1", 32),
		},
		{
			name: "A 0.0.0.0/0 element that overwrites the root",
			elements: []*LookupElement{
				createLookupElement("0.0.0.0", 0, "custom-root.pac"),
			},
			findIP: createIPNet("10.0.0.1", 32),
		},
		{
			name: "Identical IPNet for two objects",
			elements: []*LookupElement{
				createLookupElement("192.168.0.0", 24, "first.pac"),
				createLookupElement("192.168.0.0", 24, "second.pac"),
			},
			findIP: createIPNet("192.168.0.1", 32),
		},
		{
			name: "Identical with cidr > 32",
			elements: []*LookupElement{
				// Create elements with invalid CIDR directly instead of using helper function
				{
					IPMap: &ipMap{
						IPNet: IP.Net{
							NetworkAddress: IP.IP{Value: 3232235520}, // 192.168.0.0
							CIDR:           IP.CIDR{Value: 33, Mask: 0}, // Invalid CIDR
						},
						Filename: "invalid-cidr.pac",
					},
					PAC: &pacTemplate{
						Filename: "invalid-cidr.pac",
						content:  "// Test PAC file",
					},
					Variant: "// Test PAC file",
				},
			},
			findIP: createIPNet("192.168.0.1", 32),
		},
		{
			name: "findInTree with 0.0.0.0/0",
			elements: []*LookupElement{
				createLookupElement("10.0.0.0", 8, "ten-net.pac"),
				createLookupElement("172.16.0.0", 12, "private-net.pac"),
			},
			findIP: createIPNet("0.0.0.0", 0),
		},
		{
			name: "findInTree with a network that has two identical IPNet elements",
			elements: []*LookupElement{
				createLookupElement("192.168.0.0", 24, "first.pac"),
				createLookupElement("192.168.0.0", 24, "second.pac"),
			},
			findIP: createIPNet("192.168.0.0", 24),
		},
		{
			name:     "findInTree with invalid IP.Net Network Address",
			elements: []*LookupElement{},
			findIP: &IP.Net{
				NetworkAddress: IP.IP{Value: 0xFFFFFFFF}, // Invalid value (255.255.255.255)
				CIDR:           IP.CIDR{Value: 24, Mask: IP.Mask24},
			},
		},
		{
			name:     "findInTree with invalid IP.Net CIDR",
			elements: []*LookupElement{},
			findIP: &IP.Net{
				NetworkAddress: IP.IP{Value: 3232235520}, // 192.168.0.0
				CIDR:           IP.CIDR{Value: 255, Mask: 0}, // Invalid CIDR
			},
		},
		{
			name:     "findInTree with an IP that is initialised with only default values",
			elements: []*LookupElement{},
			findIP:   &IP.Net{}, // Default/zero values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If there's a setup function, run it
			if tt.setup != nil {
				tt.setup()
			}

			// Build the tree
			tree := buildLookupTree(tt.elements)

			// Verify the tree is not nil
			if tree == nil {
				t.Fatal("buildLookupTree returned nil")
			}

			// Find in tree
			result, stack := findInTree(tree, tt.findIP)

			// Verify result is not nil
			if result == nil {
				t.Fatal("findInTree returned nil result")
			}

			// Verify stack is not nil and contains at least one element
			if stack == nil || len(stack) == 0 {
				t.Fatal("findInTree returned nil or empty stack")
			}

			// Additional test-specific assertions
			switch tt.name {
			case "No elements at all":
				// Should return the default root element
				if result.IPMap.Filename != GetConfig().DefaultPACFile {
					t.Errorf("Expected default PAC file, got %s", result.IPMap.Filename)
				}
			case "A 0.0.0.0/0 element that overwrites the root":
				// Should return the custom root element
				if result.IPMap.Filename != "custom-root.pac" {
					t.Errorf("Expected custom-root.pac, got %s", result.IPMap.Filename)
				}
			case "Identical IPNet for two objects":
				// Should return one of the elements
				// The actual implementation returns the second element when there are identical IPNets
				if result.IPMap.Filename != "first.pac" && result.IPMap.Filename != "second.pac" {
					t.Errorf("Expected either first.pac or second.pac, got %s", result.IPMap.Filename)
				}
			case "findInTree with 0.0.0.0/0":
				// Should return the root element
				if result.IPMap.IPNet.GetRawCIDR() != 0 {
					t.Errorf("Expected CIDR 0, got %d", result.IPMap.IPNet.GetRawCIDR())
				}
			case "findInTree with a network that has two identical IPNet elements":
				// Should return one of the elements
				// The actual implementation returns the second element when there are identical IPNets
				if result.IPMap.Filename != "first.pac" && result.IPMap.Filename != "second.pac" {
					t.Errorf("Expected either first.pac or second.pac, got %s", result.IPMap.Filename)
				}
			case "findInTree with invalid IP.Net Network Address", 
				 "findInTree with invalid IP.Net CIDR", 
				 "findInTree with an IP that is initialised with only default values":
				// Should return the root element
				if result.IPMap.IPNet.GetRawCIDR() != 0 {
					t.Errorf("Expected CIDR 0, got %d", result.IPMap.IPNet.GetRawCIDR())
				}
			}
		})
	}
}

// Helper function to create an IP.Net without error handling
func createIPNet(ip string, cidr int) *IP.Net {
	ipNet, err := IP.NewIPNetFromMixed(ip, cidr)
	if err != nil {
		// In a real test, we might want to handle this differently
		// but for these tests, we'll panic if we can't create a valid IP.Net
		panic(err)
	}
	return &ipNet
}

// Helper function to create a LookupElement for testing
func createLookupElement(ip string, cidr int, filename string) *LookupElement {
	ipNet := createIPNet(ip, cidr)

	element := &LookupElement{
		IPMap: &ipMap{
			IPNet:    *ipNet,
			Filename: filename,
		},
		PAC: &pacTemplate{
			Filename: filename,
			content:  "// Test PAC file",
		},
		Variant: "// Test PAC file",
	}

	return element
}

// Setup the test environment with necessary global variables
func setupTestEnvironment() {
	// Initialize rootPAC if needed
	if rootPAC == nil {
		rootPAC = &LookupElement{
			PAC: &pacTemplate{
				Filename: "default.pac",
				content:  "// Default PAC file",
			},
		}
	}

	// Initialize config if needed
	if confStorage == nil {
		confStorage = &Config{
			DefaultPACFile: "default.pac",
			ContactInfo:    "Test Contact",
		}
	}
}
