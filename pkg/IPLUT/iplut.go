package IPLUT

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

// Payload is the behavior IPLUT needs from stored values.
type Payload interface {
	IsIdentical(other any) bool
	Stringify() string
}

func isNil[T any](v T) bool {
	// Generic zero values need a reflection check because interfaces can hold typed nils.
	if any(v) == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

// Node stores one IP network and a generic payload.
type Node[T Payload] struct {
	Net      IP.Net
	Content  T
	Children []*Node[T]
}

// NewNode creates a tree node for the given network and payload.
func NewNode[T Payload](net IP.Net, content T) *Node[T] {
	return &Node[T]{
		Net:      net,
		Content:  content,
		Children: make([]*Node[T], 0),
	}
}

// Build creates a tree from the supplied nodes under the provided root.
func Build[T Payload](root *Node[T], elements []*Node[T]) *Node[T] {
	if root == nil {
		return nil
	}

	// Insert every node first; the tree is normalized only after the full set is present.
	for _, elem := range elements {
		Insert(root, elem)
	}

	// Collapse an artificial zero-CIDR parent when it is the only thing standing between root and data.
	if len(root.Children) == 1 && root.Children[0] != nil && root.Children[0].Net.GetRawCIDR() == 0 {
		root = root.Children[0]
	}

	// Remove redundant structure only after the tree is fully assembled.
	simplify(root)
	return root
}

// Find walks the tree and returns the most specific matching node plus its ancestry.
func Find[T Payload](root *Node[T], ip *IP.Net) (*Node[T], []*Node[T]) {
	if root == nil || ip == nil {
		return nil, nil
	}

	// Descend into the first matching child and keep the current node in the ancestry stack.
	for _, child := range root.Children {
		if child == nil {
			continue
		}

		if ip.IsSubnetOf(child.Net) {
			node, stack := Find(child, ip)
			stack = append([]*Node[T]{root}, stack...)
			return node, stack
		}
	}

	return root, []*Node[T]{root}
}

// Stringify renders the tree using the payload formatter.
func Stringify[T Payload](root *Node[T]) string {
	if root == nil {
		return ""
	}
	return stringifyNode(root, 0)
}

// stringifyNode renders one node and its children.
func stringifyNode[T Payload](node *Node[T], level int) string {
	if node == nil {
		return ""
	}

	render := ""
	if !isNil(node.Content) {
		render = node.Content.Stringify()
	}
	str := fmt.Sprintf("%s - %s\n", strings.Repeat("\t", level), render)
	for _, child := range node.Children {
		str += stringifyNode(child, level+1)
	}
	return str
}

// StringifyStack renders a stack using the configured node formatter.
func StringifyStack[T Payload](stack []T) string {
	if len(stack) == 0 {
		return ""
	}

	var b strings.Builder
	for level, node := range stack {
		if isNil(node) {
			continue
		}

		// Skip nil entries so partial debug stacks still render cleanly.
		fmt.Fprintf(&b, "%s - %s\n", strings.Repeat("\t", level), node.Stringify())
	}

	return b.String()
}

// Insert places a node into the correct place in the tree.
func Insert[T Payload](root *Node[T], elem *Node[T]) {
	if root == nil || elem == nil {
		return
	}

	// Copy the node so the caller can keep its original instance untouched.
	newNode := &Node[T]{
		Net:      elem.Net,
		Content:  elem.Content,
		Children: make([]*Node[T], 0),
	}

	// Walk backward so child promotion and sibling removal stay index-safe.
	for i := len(root.Children) - 1; i >= 0; i-- {
		child := root.Children[i]
		if child == nil {
			continue
		}

		if elem.Net.IsSubnetOf(child.Net) {
			Insert(child, elem)
			return
		}

		if child.Net.IsSubnetOf(elem.Net) || elem.Net.IsIdentical(root.Net) {
			// Promote the old child under the new parent when the new node is broader.
			newNode.Children = append(newNode.Children, child)
			if i == len(root.Children)-1 {
				root.Children = root.Children[:i]
			} else {
				root.Children = append(root.Children[:i], root.Children[i+1:]...)
			}
		}
	}

	root.Children = append(root.Children, newNode)
}

// simplify removes intermediate nodes that carry the same payload as their parent.
func simplify[T Payload](root *Node[T]) {
	if root == nil {
		return
	}

	var simplifiedChildren []*Node[T]
	for _, child := range root.Children {
		if child == nil {
			continue
		}

		// Normalize each subtree before comparing it with the current node.
		simplify(child)

		sameContent := !isNil(child.Content) && !isNil(root.Content) && child.Content.IsIdentical(root.Content)
		sameNet := child.Net.IsIdentical(root.Net)

		if !(sameContent && sameNet) {
			simplifiedChildren = append(simplifiedChildren, child)
			continue
		}

		// Flatten nodes that add no new information beyond their parent.
		simplifiedChildren = append(simplifiedChildren, child.Children...)
	}

	// Keep siblings sorted so tree rendering and lookup behavior remain deterministic.
	sort.SliceStable(simplifiedChildren, func(i, j int) bool {
		left := simplifiedChildren[i]
		right := simplifiedChildren[j]
		if left == nil || right == nil {
			return left != nil
		}
		if left.Net.NetworkAddress.Value != right.Net.NetworkAddress.Value {
			return left.Net.NetworkAddress.Value < right.Net.NetworkAddress.Value
		}
		return left.Net.CIDR.Value < right.Net.CIDR.Value
	})

	root.Children = simplifiedChildren
}

// StackContent unwrapps the content from a node stack.
func StackContent[T Payload](stack []*Node[T]) []T {
	if len(stack) == 0 {
		return []T{}
	}
	result := make([]T, 0, len(stack))
	for _, entry := range stack {
		if entry == nil || isNil(entry.Content) {
			continue
		}
		result = append(result, entry.Content)
	}
	return result
}
