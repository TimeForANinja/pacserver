package internal

import (
	"fmt"
	"testing"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

// helper method to create an ipNet, that can not error
func forceIPNet(ip string, net int) IP.Net {
	ipNet, err := IP.NewIPNetFromMixed(ip, net)
	if err != nil {
		panic(err)
	}
	return ipNet
}

func TestFindInTree(t *testing.T) {
	buildInRootElement := &LookupElement{
		PAC: nil,
		IPMap: &ipMap{
			IPNet:    forceIPNet("0.0.0.0", 0),
			Filename: "root",
		},
	}
	globalElement := &LookupElement{
		PAC: &pacTemplate{},
		IPMap: &ipMap{
			IPNet:    forceIPNet("0.0.0.0", 0),
			Filename: "root",
		},
	}
	child1Element := &LookupElement{
		PAC: &pacTemplate{},
		IPMap: &ipMap{
			IPNet:    forceIPNet("192.168.0.0", 16),
			Filename: "child",
		},
	}
	child2Element := &LookupElement{
		PAC: &pacTemplate{},
		IPMap: &ipMap{
			IPNet:    forceIPNet("192.168.0.0", 24),
			Filename: "child-child",
		},
	}
	demoTree := &lookupTreeNode{
		data: buildInRootElement,
		children: []*lookupTreeNode{
			{
				data: globalElement,
				children: []*lookupTreeNode{
					{
						data: child1Element,
						children: []*lookupTreeNode{
							{data: child2Element, children: []*lookupTreeNode{}},
						},
					},
				},
			},
		},
	}

	// Define the test cases
	tests := []struct {
		name string
		tree *lookupTreeNode
		ip   *IP.Net
		want *LookupElement
	}{
		{
			name: "Does not Error when only feed with dummy root, but defaults to root-pac",
			tree: &lookupTreeNode{data: buildInRootElement, children: []*lookupTreeNode{}},
			ip: &IP.Net{
				// 192.168.0.0/32
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: buildInRootElement,
		},
		{
			name: "Unknown Element defaults to root",
			tree: demoTree,
			ip: &IP.Net{
				// 0.0.0.0/32
				NetworkAddress: IP.IP{Value: 0},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: globalElement,
		},
		{
			name: "Most specific Node gets matched",
			tree: demoTree,
			ip: &IP.Net{
				// 192.168.0.0/32
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: child2Element,
		},
		{
			name: "Works for searching networks",
			tree: demoTree,
			ip: &IP.Net{
				// 192.168.0.0/16
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 16, Mask: IP.Mask16},
			},
			want: child1Element,
		},
	}

	// Run the test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := findInTree(tc.tree, tc.ip)

			if (got == nil && tc.want != nil) || got != tc.want {
				t.Errorf("findInTree() = %v, want %v", got, tc.want)
			}
		})
	}
}

// helper method to compare two trees based on the IPMap.Filename property
func simpleTreeCompare(root1, root2 *lookupTreeNode) bool {
	if root1.data.IPMap.Filename != root2.data.IPMap.Filename {
		return false
	}
	if len(root1.children) != len(root2.children) {
		return false
	}
	for idx := range root1.children {
		if !simpleTreeCompare(root1.children[idx], root2.children[idx]) {
			return false
		}
	}
	return true
}

func TestBuildLookupTree(t *testing.T) {
	// Prepare the test cases in a table driven format
	testCases := []struct {
		Name     string
		Input    []*LookupElement
		Expected *lookupTreeNode
	}{
		{
			Name:  "Empty Input",
			Input: []*LookupElement{},
			Expected: &lookupTreeNode{
				data: &LookupElement{
					IPMap: &ipMap{
						Filename: "",
					},
				},
			},
		},
		{
			Name: "Overwrites build-in with explicit default",
			Input: []*LookupElement{
				{IPMap: &ipMap{IPNet: forceIPNet("0.0.0.0", 0), Filename: "new global"}, PAC: &pacTemplate{}},
			},
			Expected: &lookupTreeNode{
				data:     &LookupElement{IPMap: &ipMap{Filename: "new global"}},
				children: []*lookupTreeNode{},
			},
		},
		{
			Name: "Nested Networks",
			Input: []*LookupElement{
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "Node 1"}, PAC: &pacTemplate{}},
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 24), Filename: "Node 2"}, PAC: &pacTemplate{}},
			},
			Expected: &lookupTreeNode{
				data: &LookupElement{
					IPMap: &ipMap{Filename: ""},
				},
				children: []*lookupTreeNode{
					{
						data: &LookupElement{
							IPMap: &ipMap{Filename: "Node 1"},
						},
						children: []*lookupTreeNode{
							{
								data:     &LookupElement{IPMap: &ipMap{Filename: "Node 2"}},
								children: []*lookupTreeNode{},
							},
						},
					},
				},
			},
		},
		{
			Name: "Nested (unordered) Networks",
			Input: []*LookupElement{
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 24), Filename: "Node 2"}, PAC: &pacTemplate{}},
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "Node 1"}, PAC: &pacTemplate{}},
			},
			Expected: &lookupTreeNode{
				data: &LookupElement{
					IPMap: &ipMap{Filename: ""},
				},
				children: []*lookupTreeNode{
					{
						data: &LookupElement{
							IPMap: &ipMap{Filename: "Node 1"},
						},
						children: []*lookupTreeNode{
							{
								data:     &LookupElement{IPMap: &ipMap{Filename: "Node 2"}},
								children: []*lookupTreeNode{},
							},
						},
					},
				},
			},
		},
		{
			Name: "Duplicate Networks (get removed by simplify)",
			Input: []*LookupElement{
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "Node 1"}, PAC: &pacTemplate{}},
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "Node 2"}, PAC: &pacTemplate{}},
				{IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 24), Filename: "Node 3"}, PAC: &pacTemplate{}},
			},
			Expected: &lookupTreeNode{
				data: &LookupElement{
					IPMap: &ipMap{Filename: ""},
				},
				children: []*lookupTreeNode{
					{
						data: &LookupElement{
							IPMap: &ipMap{Filename: "Node 1"},
						},
						children: []*lookupTreeNode{
							{
								data:     &LookupElement{IPMap: &ipMap{Filename: "Node 3"}},
								children: []*lookupTreeNode{},
							},
						},
					},
				},
			},
		},
	}

	// prepare by defining required global objects
	rootPAC = &LookupElement{PAC: &pacTemplate{}}
	confStorage = &Config{}

	// Execute all the test cases
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// call the function to test with the test case input and get the output
			actualOutput := buildLookupTree(testCase.Input)

			if !simpleTreeCompare(testCase.Expected, actualOutput) {
				t.Error("Tree differs from expected Tree")
				fmt.Println("Got", stringifyLookupTree(actualOutput))
				fmt.Println("Expected", stringifyLookupTree(testCase.Expected))
			}
		})
	}
}

func TestSimplifyTreeSorting(t *testing.T) {
	root := &lookupTreeNode{
		data: &LookupElement{
			IPMap: &ipMap{
				IPNet:    forceIPNet("0.0.0.0", 0),
				Filename: "root",
			},
		},
		children: []*lookupTreeNode{
			{data: &LookupElement{
				IPMap: &ipMap{IPNet: forceIPNet("192.168.2.0", 24), Filename: "third"},
			}},
			{data: &LookupElement{
				IPMap: &ipMap{IPNet: forceIPNet("192.168.1.0", 24), Filename: "second"},
			}},
			{data: &LookupElement{
				IPMap: &ipMap{IPNet: forceIPNet("10.0.0.0", 8), Filename: "first"},
			}},
		},
	}

	simplifyTree(root)

	// Verify sorting
	if len(root.children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(root.children))
	}

	expectedOrder := []string{"first", "second", "third"}
	for i, expected := range expectedOrder {
		if root.children[i].data.IPMap.Filename != expected {
			t.Errorf("Child at position %d: expected %s, got %s",
				i, expected, root.children[i].data.IPMap.Filename)
		}
	}
}

func TestInsertTreeElement(t *testing.T) {
	t.Run("Safe removal when iterating backwards", func(t *testing.T) {
		// Create a root node with multiple children
		root := &lookupTreeNode{
			data: &LookupElement{
				IPMap: &ipMap{
					IPNet:    forceIPNet("0.0.0.0", 0),
					Filename: "root",
				},
			},
			children: []*lookupTreeNode{
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("192.168.1.0", 24), Filename: "child1"},
					},
				},
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("192.168.2.0", 24), Filename: "child2"},
					},
				},
			},
		}

		// Insert a new element that should contain both existing children
		// This will force both "old" children to be removed from the "root" node
		// and pushed under the "partent" node
		newElem := &LookupElement{
			IPMap: &ipMap{
				IPNet:    forceIPNet("192.168.0.0", 16),
				Filename: "parent",
			},
		}

		insertTreeElement(root, newElem)

		// Verify that both children were properly moved
		if len(root.children) != 1 {
			t.Errorf("Expected root to have 1 child, got %d", len(root.children))
		}
		if root.children[0].data.IPMap.Filename != "parent" {
			t.Errorf("Expected root.child to be 'parent', got %s", root.children[0].data.PAC.Filename)
		}
		if len(root.children[0].children) != 2 {
			t.Errorf("Expected new node to have 2 children, got %d", len(root.children[0].children))
		}
	})

	t.Run("Safe removal of last element", func(t *testing.T) {
		// Create a root node with single child
		root := &lookupTreeNode{
			data: &LookupElement{
				IPMap: &ipMap{IPNet: forceIPNet("0.0.0.0", 0), Filename: "root"},
			},
			children: []*lookupTreeNode{
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("192.168.1.0", 24), Filename: "lastChild"},
					},
				},
			},
		}

		// Insert a new element that should contain the existing child
		// the existing child is at last position, to check for index-out-of-bounds problems
		newElem := &LookupElement{
			IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "parent"},
		}

		insertTreeElement(root, newElem)

		// Verify that the child was properly moved
		if len(root.children) != 1 {
			t.Errorf("Expected root to have 1 child, got %d", len(root.children))
		}
		if len(root.children[0].children) != 1 {
			t.Errorf("Expected new node to have 1 child, got %d", len(root.children[0].children))
		}
		if root.children[0].children[0].data.IPMap.Filename != "lastChild" {
			t.Errorf("Expected child to be 'lastChild', got %s", root.children[0].children[0].data.IPMap.Filename)
		}
	})

	t.Run("Safe removal of middle element", func(t *testing.T) {
		// Create a root node with three children
		root := &lookupTreeNode{
			data: &LookupElement{
				IPMap: &ipMap{IPNet: forceIPNet("0.0.0.0", 0), Filename: "root"},
			},
			children: []*lookupTreeNode{
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("10.0.0.0", 8), Filename: "first"},
					},
				},
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("192.168.1.0", 24), Filename: "middle"},
					},
				},
				{
					data: &LookupElement{
						IPMap: &ipMap{IPNet: forceIPNet("172.16.0.0", 12), Filename: "last"},
					},
				},
			},
		}

		// Insert a new element that should contain only the middle child
		newElem := &LookupElement{
			IPMap: &ipMap{IPNet: forceIPNet("192.168.0.0", 16), Filename: "parent"},
		}

		insertTreeElement(root, newElem)

		// Verify the structure
		if len(root.children) != 3 {
			t.Errorf("Expected root to have 3 children, got %d", len(root.children))
		}

		// Find the new parent node (it should be in sorted order)
		var parentNode *lookupTreeNode
		for _, child := range root.children {
			if child.data.IPMap.Filename == "parent" {
				parentNode = child
				break
			}
		}

		if parentNode == nil {
			t.Fatal("Expected to find 'parent' node")
		}

		// Verify the parent has the middle child
		if len(parentNode.children) != 1 {
			t.Errorf("Expected parent to have 1 child, got %d", len(parentNode.children))
		}

		if parentNode.children[0].data.IPMap.Filename != "middle" {
			t.Errorf("Expected child to be 'middle', got %s", parentNode.children[0].data.IPMap.Filename)
		}

		// Verify other children remained in root
		var foundFirst, foundLast bool
		for _, child := range root.children {
			if child.data.IPMap.Filename == "first" {
				foundFirst = true
			}
			if child.data.IPMap.Filename == "last" {
				foundLast = true
			}
		}

		if !foundFirst {
			t.Error("First child missing from root")
		}
		if !foundLast {
			t.Error("Last child missing from root")
		}
	})
}
