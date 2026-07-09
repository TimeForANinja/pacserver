package IPLUT

import (
	"fmt"
	"testing"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

type testPayload string

func (v testPayload) IsIdentical(other any) bool {
	switch other := other.(type) {
	case testPayload:
		return v == other
	default:
		return false
	}
}

func (v testPayload) Stringify() string {
	return string(v)
}

func forceIPNet(ip string, cidr int) IP.Net {
	ipNet, err := IP.NewIPNetFromMixed(ip, cidr)
	if err != nil {
		panic(err)
	}
	return ipNet
}

func testNode(ip string, cidr int, content string) *Node[testPayload] {
	return NewNode(forceIPNet(ip, cidr), testPayload(content))
}

func TestFind(t *testing.T) {
	buildInRootElement := testNode("0.0.0.0", 0, "root")
	globalElement := testNode("0.0.0.0", 0, "global")
	child1Element := testNode("192.168.0.0", 16, "child")
	child2Element := testNode("192.168.0.0", 24, "child-child")
	demoTree := &Node[testPayload]{
		Net:     buildInRootElement.Net,
		Content: buildInRootElement.Content,
		Children: []*Node[testPayload]{
			{
				Net:     globalElement.Net,
				Content: globalElement.Content,
				Children: []*Node[testPayload]{
					{
						Net:     child1Element.Net,
						Content: child1Element.Content,
						Children: []*Node[testPayload]{
							{Net: child2Element.Net, Content: child2Element.Content, Children: []*Node[testPayload]{}},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name string
		tree *Node[testPayload]
		ip   *IP.Net
		want *Node[testPayload]
	}{
		{
			name: "defaults to root for empty tree",
			tree: &Node[testPayload]{Net: buildInRootElement.Net, Content: buildInRootElement.Content, Children: []*Node[testPayload]{}},
			ip: &IP.Net{
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: buildInRootElement,
		},
		{
			name: "unknown element defaults to root",
			tree: demoTree,
			ip: &IP.Net{
				NetworkAddress: IP.IP{Value: 0},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: globalElement,
		},
		{
			name: "most specific node wins",
			tree: demoTree,
			ip: &IP.Net{
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 32, Mask: IP.Mask32},
			},
			want: child2Element,
		},
		{
			name: "network search works",
			tree: demoTree,
			ip: &IP.Net{
				NetworkAddress: IP.IP{Value: 3232235520},
				CIDR:           IP.CIDR{Value: 16, Mask: IP.Mask16},
			},
			want: child1Element,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := Find(tc.tree, tc.ip)
			if got == nil || tc.want == nil {
				if got != tc.want {
					t.Fatalf("Find() = %v, want %v", got, tc.want)
				}
				return
			}

			if got.Content != tc.want.Content {
				t.Errorf("Find() = %v, want %v", got.Content, tc.want.Content)
			}
		})
	}
}

func simpleTreeCompare(root1, root2 *Node[testPayload]) bool {
	if root1 == nil || root2 == nil {
		return root1 == root2
	}
	if root1.Net != root2.Net {
		return false
	}
	if root1.Content != root2.Content {
		return false
	}
	if len(root1.Children) != len(root2.Children) {
		return false
	}
	for idx := range root1.Children {
		if !simpleTreeCompare(root1.Children[idx], root2.Children[idx]) {
			return false
		}
	}
	return true
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name     string
		root     *Node[testPayload]
		input    []*Node[testPayload]
		expected *Node[testPayload]
	}{
		{
			name:  "empty input keeps root",
			root:  testNode("0.0.0.0", 0, ""),
			input: []*Node[testPayload]{},
			expected: &Node[testPayload]{
				Net:     forceIPNet("0.0.0.0", 0),
				Content: testPayload(""),
			},
		},
		{
			name: "explicit root replaces fake root",
			root: testNode("0.0.0.0", 0, ""),
			input: []*Node[testPayload]{
				testNode("0.0.0.0", 0, "new global"),
			},
			expected: testNode("0.0.0.0", 0, "new global"),
		},
		{
			name: "nested networks",
			root: testNode("0.0.0.0", 0, ""),
			input: []*Node[testPayload]{
				testNode("192.168.0.0", 16, "Node 1"),
				testNode("192.168.0.0", 24, "Node 2"),
			},
			expected: &Node[testPayload]{
				Net:     forceIPNet("0.0.0.0", 0),
				Content: testPayload(""),
				Children: []*Node[testPayload]{
					{
						Net:     forceIPNet("192.168.0.0", 16),
						Content: testPayload("Node 1"),
						Children: []*Node[testPayload]{
							{
								Net:      forceIPNet("192.168.0.0", 24),
								Content:  testPayload("Node 2"),
								Children: []*Node[testPayload]{},
							},
						},
					},
				},
			},
		},
		{
			name: "duplicate networks collapse",
			root: testNode("0.0.0.0", 0, ""),
			input: []*Node[testPayload]{
				testNode("192.168.0.0", 16, "Node 1"),
				testNode("192.168.0.0", 16, "Node 2"),
				testNode("192.168.0.0", 24, "Node 3"),
			},
			expected: &Node[testPayload]{
				Net:     forceIPNet("0.0.0.0", 0),
				Content: testPayload(""),
				Children: []*Node[testPayload]{
					{
						Net:     forceIPNet("192.168.0.0", 16),
						Content: testPayload("Node 1"),
						Children: []*Node[testPayload]{
							{
								Net:     forceIPNet("192.168.0.0", 16),
								Content: testPayload("Node 2"),
								Children: []*Node[testPayload]{
									{
										Net:      forceIPNet("192.168.0.0", 24),
										Content:  testPayload("Node 3"),
										Children: []*Node[testPayload]{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Build(tc.root, tc.input)
			if !simpleTreeCompare(tc.expected, actual) {
				t.Error("Tree differs from expected Tree")
				fmt.Println("Got", Stringify(actual))
				fmt.Println("Expected", Stringify(tc.expected))
			}
		})
	}
}

func TestSimplifySorting(t *testing.T) {
	root := testNode("0.0.0.0", 0, "root")
	root.Children = []*Node[testPayload]{
		testNode("192.168.2.0", 24, "third"),
		testNode("192.168.1.0", 24, "second"),
		testNode("10.0.0.0", 8, "first"),
	}

	simplify(root)

	if len(root.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(root.Children))
	}

	expectedOrder := []string{"first", "second", "third"}
	for i, expected := range expectedOrder {
		if root.Children[i].Content != testPayload(expected) {
			t.Errorf("Child at position %d: expected %s, got %v", i, expected, root.Children[i].Content)
		}
	}
}

func TestInsert(t *testing.T) {
	t.Run("moves contained children under new parent", func(t *testing.T) {
		root := testNode("0.0.0.0", 0, "root")
		root.Children = []*Node[testPayload]{
			testNode("192.168.1.0", 24, "child1"),
			testNode("192.168.2.0", 24, "child2"),
		}

		Insert(root, testNode("192.168.0.0", 16, "parent"))

		if len(root.Children) != 1 {
			t.Fatalf("Expected root to have 1 child, got %d", len(root.Children))
		}
		if root.Children[0].Content != testPayload("parent") {
			t.Fatalf("Expected root.child to be parent, got %v", root.Children[0].Content)
		}
		if len(root.Children[0].Children) != 2 {
			t.Fatalf("Expected new node to have 2 children, got %d", len(root.Children[0].Children))
		}
	})

	t.Run("safe removal of last element", func(t *testing.T) {
		root := testNode("0.0.0.0", 0, "root")
		root.Children = []*Node[testPayload]{testNode("192.168.1.0", 24, "lastChild")}

		Insert(root, testNode("192.168.0.0", 16, "parent"))

		if len(root.Children) != 1 {
			t.Fatalf("Expected root to have 1 child, got %d", len(root.Children))
		}
		if len(root.Children[0].Children) != 1 {
			t.Fatalf("Expected new node to have 1 child, got %d", len(root.Children[0].Children))
		}
		if root.Children[0].Children[0].Content != testPayload("lastChild") {
			t.Fatalf("Expected child to be lastChild, got %v", root.Children[0].Children[0].Content)
		}
	})

	t.Run("safe removal of middle element", func(t *testing.T) {
		root := testNode("0.0.0.0", 0, "root")
		root.Children = []*Node[testPayload]{
			testNode("10.0.0.0", 8, "first"),
			testNode("192.168.1.0", 24, "middle"),
			testNode("172.16.0.0", 12, "last"),
		}

		Insert(root, testNode("192.168.0.0", 16, "parent"))

		if len(root.Children) != 3 {
			t.Fatalf("Expected root to have 3 children, got %d", len(root.Children))
		}

		var parentNode *Node[testPayload]
		for _, child := range root.Children {
			if child.Content == testPayload("parent") {
				parentNode = child
				break
			}
		}

		if parentNode == nil {
			t.Fatal("Expected to find parent node")
		}
		if len(parentNode.Children) != 1 {
			t.Fatalf("Expected parent to have 1 child, got %d", len(parentNode.Children))
		}
		if parentNode.Children[0].Content != testPayload("middle") {
			t.Fatalf("Expected child to be middle, got %v", parentNode.Children[0].Content)
		}

		var foundFirst, foundLast bool
		for _, child := range root.Children {
			if child.Content == testPayload("first") {
				foundFirst = true
			}
			if child.Content == testPayload("last") {
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
