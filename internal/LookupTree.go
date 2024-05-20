package internal

import (
	"fmt"
	"strings"

	"github.com/timeforaninja/pacserver/pkg/IP"
)

type lookupTreeNode struct {
	data     *lookupElement
	children []*lookupTreeNode
}

func stringifyLookupTree(root *lookupTreeNode) string {
	return _stringifyLookupTree(root, 0)
}

func _stringifyLookupTree(node *lookupTreeNode, level int) string {
	str := fmt.Sprintf(
		"%s - %s | pac(%s, %s)\n",
		strings.Repeat(" ", level),
		node.data.IPMap.IPNet.ToString(),
		node.data.IPMap.Filename,
		strings.Join(node.data.IPMap.Hostnames, ", "),
	)

	for _, c := range node.children {
		str += _stringifyLookupTree(c, level+1)
	}

	return str
}

func insertTreeElement(root *lookupTreeNode, elem *lookupElement) {
	newNode := &lookupTreeNode{data: elem, children: []*lookupTreeNode{}}

	for i, child := range root.children {
		// Check if the elem is a subnet of the child
		if elem.isSubnetOf(*child.data) {
			insertTreeElement(child, elem)
			return
		}
		// Check if the child is a subnet of the elem
		// Or identical (which simply get stacked)
		if child.data.isSubnetOf(*elem) || elem.isIdenticalNet(*root.data) {
			// push child into newnode
			newNode.children = append(newNode.children, child)
			// remove child from root
			root.children = append(root.children[:i], root.children[i+1:]...)
		}
	}
	// If no subnet relation found, add newNode as a child
	root.children = append(root.children, newNode)
}

func buildLookupTree(elements []*lookupElement) *lookupTreeNode {
	// build a "fake" root element
	// this massively simplifes code since we
	// a) always only have a single root element
	// b) can make sure that we never have to swap the root
	root_ip, _ := IP.NewIPNetFromMixed("0.0.0.0", 0)
	var root = &lookupTreeNode{
		data: &lookupElement{
			IPMap: &ipMap{
				IPNet: root_ip,
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
		if !child.data.isIdenticalPAC(*root.data) {
			simplifiedChildren = append(simplifiedChildren, child)
		}
	}
	// Replace the children with the simplified children
	root.children = simplifiedChildren
}

func findInTree(root *lookupTreeNode, ip *IP.IPNet) *lookupElement {
	for _, c := range root.children {
		if ip.IsSubnetOf(c.data.IPMap.IPNet) {
			return findInTree(c, ip)
		}
	}
	// check if it's our dummy root
	// this would mean no rule matches the location
	if root.data.PAC == nil {
		return nil
	}
	return root.data
}
