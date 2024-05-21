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
			name: "Does not Error when only feed with dummy root",
			tree: &lookupTreeNode{data: buildInRootElement, children: []*lookupTreeNode{}},
			ip: &IP.Net{
				// 192.168.0.0/32
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: nil,
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
			got := findInTree(tc.tree, tc.ip)

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
