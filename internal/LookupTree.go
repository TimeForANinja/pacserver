package internal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

type lookupTreeNode struct {
	data     *LookupElement
	children []*lookupTreeNode
}

func stringifyLookupTree(root *lookupTreeNode) string {
	return _stringifyLookupTree(root, 0)
}

func _stringifyLookupTree(node *lookupTreeNode, level int) string {
	str := fmt.Sprintf(
		"%s - %s | pac(%s)\n",
		strings.Repeat(" ", level),
		node.data.IPMap.IPNet.ToString(),
		node.data.IPMap.Filename,
	)

	for _, c := range node.children {
		str += _stringifyLookupTree(c, level+1)
	}

	return str
}

func insertTreeElement(root *lookupTreeNode, elem *LookupElement) {
	newNode := &lookupTreeNode{data: elem, children: []*lookupTreeNode{}}

	// Iterate backwards
	// in case we need to remove an element this does not screw up the counter
	for i := len(root.children) - 1; i >= 0; i-- {
		child := root.children[i]
		// Check if the elem is a subnet of the child
		if elem.isSubnetOf(*child.data) {
			insertTreeElement(child, elem)
			return
		}
		// Check if the child is a subnet of the elem
		// Or identical (which simply get stacked)
		if child.data.isSubnetOf(*elem) || elem.isIdenticalNet(*root.data) {
			// push child into new node
			newNode.children = append(newNode.children, child)
			// remove child from root
			// watch out for doing i+1 on the last element
			if i == len(root.children)-1 {
				root.children = root.children[:i]
			} else {
				root.children = append(root.children[:i], root.children[i+1:]...)
			}
		}
	}
	// If no subnet relation found, add newNode as a child
	root.children = append(root.children, newNode)
}

func buildLookupTree(elements []*LookupElement) *lookupTreeNode {
	// build a "fake" root element
	// this massively simplifies code since we
	// a) always only have a single root element
	// b) can make sure that we never have to swap the root
	rootIP, _ := IP.NewIPNetFromMixed("0.0.0.0", 0)
	var root = &lookupTreeNode{
		data: &LookupElement{
			IPMap: &ipMap{
				IPNet: rootIP,
			},
		},
	}

	// insert one element after another into our tree
	for _, elem := range elements {
		insertTreeElement(root, elem)
	}

	// check if the user provided a single root
	// and if so, replace our custom root by it
	if len(root.children) == 1 && root.children[0].data.getRawCIDR() == 0 {
		root = root.children[0]
	}

	// remove "intermediate" nodes
	// that serve the same pac as their parent
	simplifyTree(root)

	return root
}

func simplifyTree(root *lookupTreeNode) {
	var simplifiedChildren []*lookupTreeNode
	for _, child := range root.children {
		// Recursively simplify child nodes
		simplifyTree(child)
		// If the simplified child is not identical to the root, keep it
		// The following are the two conditions we check
		a := child.data.isIdenticalPAC(*root.data)
		b := child.data.isIdenticalNet(*root.data)
		// xor
		if !(a && b) {
			simplifiedChildren = append(simplifiedChildren, child)
		} else {
			simplifiedChildren = append(simplifiedChildren, child.children...)
		}
	}
	// Sort the Children
	sort.Slice(simplifiedChildren, func(i, j int) bool {
		return simplifiedChildren[i].data.IPMap.CompareForSort(simplifiedChildren[j].data.IPMap)
	})
	// Replace the children with the simplified children
	root.children = simplifiedChildren
}

func findInTree(root *lookupTreeNode, ip *IP.Net) *LookupElement {
	// Check for nil root to prevent panic
	if root == nil {
		return nil
	}
	if root.data == nil || root.data.IPMap == nil {
		return nil
	}

	for _, c := range root.children {
		// Check for nil child data or IPMap to prevent panic
		if c == nil || c.data == nil || c.data.IPMap == nil {
			continue
		}

		if ip.IsSubnetOf(c.data.IPMap.IPNet) {
			// Use a recursive call to check if we have more detailed children
			return findInTree(c, ip)
		}
	}

	// no child matched
	// check if it's our dummy root - this would mean no rule matches the location
	if root.data.PAC == nil {
		return nil
	}
	return root.data
}
