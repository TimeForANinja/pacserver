package internal

import (
	"reflect"
	"testing"
)

func TestMatchIPMapToPac(t *testing.T) {
	t.Parallel()

	// Create test data
	newPAC1 := &pacTemplate{
		Filename: "test1.pac",
		content:  "// This is test1.pac by {{ .Contact }}",
	}
	newPAC2 := &pacTemplate{
		Filename: "test2.pac",
		content:  "// This is test2.pac by {{ .Contact }}",
	}
	oldPAC3 := &pacTemplate{
		Filename: "test3.pac",
		content:  "// This is test3.pac by {{ .Contact }}",
	}
	
	ipMap1 := &ipMap{
		IPNet:    forceIPNet("192.168.0.0", 24),
		Filename: "test1.pac",
	}
	ipMap2 := &ipMap{
		IPNet:    forceIPNet("10.0.0.0", 8),
		Filename: "test2.pac",
	}
	ipMap3 := &ipMap{
		IPNet:    forceIPNet("172.16.0.0", 12),
		Filename: "test3.pac",
	}
	ipMap4 := &ipMap{
		IPNet:    forceIPNet("8.8.8.0", 24),
		Filename: "test4.pac", // This one doesn't exist in either newPACs or oldPACs
	}

	tests := []struct {
		name           string
		newPACs        []*pacTemplate
		oldPACs        []*pacTemplate
		newIPMaps      []*ipMap
		contact        string
		wantElements   int
		wantKeepPACs   int
		wantProbCount  int
	}{
		{
			name:           "All PACs found in newPACs",
			newPACs:        []*pacTemplate{newPAC1, newPAC2},
			oldPACs:        []*pacTemplate{oldPAC3},
			newIPMaps:      []*ipMap{ipMap1, ipMap2},
			contact:        "Test Contact",
			wantElements:   2,
			wantKeepPACs:   0,
			wantProbCount:  0,
		},
		{
			name:           "Some PACs found in oldPACs",
			newPACs:        []*pacTemplate{newPAC1},
			oldPACs:        []*pacTemplate{oldPAC3},
			newIPMaps:      []*ipMap{ipMap1, ipMap3},
			contact:        "Test Contact",
			wantElements:   2,
			wantKeepPACs:   1,
			wantProbCount:  1, // One warning for using cached PAC
		},
		{
			name:           "Some PACs not found at all",
			newPACs:        []*pacTemplate{newPAC1},
			oldPACs:        []*pacTemplate{},
			newIPMaps:      []*ipMap{ipMap1, ipMap4},
			contact:        "Test Contact",
			wantElements:   1,
			wantKeepPACs:   0,
			wantProbCount:  1, // One warning for missing PAC
		},
		{
			name:           "No PACs found",
			newPACs:        []*pacTemplate{},
			oldPACs:        []*pacTemplate{},
			newIPMaps:      []*ipMap{ipMap1, ipMap2},
			contact:        "Test Contact",
			wantElements:   0,
			wantKeepPACs:   0,
			wantProbCount:  2, // Two warnings for missing PACs
		},
		{
			name:           "Empty inputs",
			newPACs:        []*pacTemplate{},
			oldPACs:        []*pacTemplate{},
			newIPMaps:      []*ipMap{},
			contact:        "Test Contact",
			wantElements:   0,
			wantKeepPACs:   0,
			wantProbCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, keepPACs, probCount := matchIPMapToPac(tt.newPACs, tt.oldPACs, tt.newIPMaps, tt.contact)
			
			// Check the number of elements
			if len(elements) != tt.wantElements {
				t.Errorf("matchIPMapToPac() returned %d elements, want %d", len(elements), tt.wantElements)
			}
			
			// Check the number of PACs to keep
			if len(keepPACs) != tt.wantKeepPACs {
				t.Errorf("matchIPMapToPac() returned %d keepPACs, want %d", len(keepPACs), tt.wantKeepPACs)
			}
			
			// Check the problem counter
			if probCount != tt.wantProbCount {
				t.Errorf("matchIPMapToPac() returned problem count %d, want %d", probCount, tt.wantProbCount)
			}
			
			// For the "Some PACs found in oldPACs" test, verify that the correct PAC is kept
			if tt.name == "Some PACs found in oldPACs" {
				if len(keepPACs) > 0 {
					found := false
					for _, pac := range keepPACs {
						if pac.Filename == "test3.pac" {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("matchIPMapToPac() did not keep the expected PAC 'test3.pac'")
					}
				}
			}
			
			// For the "All PACs found in newPACs" test, verify that the elements have the correct IPMaps and PACs
			if tt.name == "All PACs found in newPACs" {
				if len(elements) == 2 {
					// Create a map of IPMap.Filename to element for easier lookup
					elementMap := make(map[string]*LookupElement)
					for _, elem := range elements {
						elementMap[elem.IPMap.Filename] = elem
					}
					
					// Check that both elements are present with the correct IPMaps and PACs
					if elem, ok := elementMap["test1.pac"]; !ok {
						t.Errorf("matchIPMapToPac() did not return an element for 'test1.pac'")
					} else {
						if !reflect.DeepEqual(elem.IPMap, ipMap1) {
							t.Errorf("Element for 'test1.pac' has incorrect IPMap")
						}
						if elem.PAC.Filename != "test1.pac" {
							t.Errorf("Element for 'test1.pac' has incorrect PAC filename: %s", elem.PAC.Filename)
						}
					}
					
					if elem, ok := elementMap["test2.pac"]; !ok {
						t.Errorf("matchIPMapToPac() did not return an element for 'test2.pac'")
					} else {
						if !reflect.DeepEqual(elem.IPMap, ipMap2) {
							t.Errorf("Element for 'test2.pac' has incorrect IPMap")
						}
						if elem.PAC.Filename != "test2.pac" {
							t.Errorf("Element for 'test2.pac' has incorrect PAC filename: %s", elem.PAC.Filename)
						}
					}
				}
			}
		})
	}
}