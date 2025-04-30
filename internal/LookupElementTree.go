package internal

/**
 * LookupTree is the main structure to store and serve the PAC Mappings
 *
 * It sorts and nests the Elements based on the IP Network to allow for
 * efficient lookup of the best matching PAC
 */

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
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
		"%s - %s\n",
		strings.Repeat("\t", level),
		node.data._stringify(),
	)

	for _, c := range node.children {
		str += _stringifyLookupTree(c, level+1)
	}

	return str
}

func _stringifyLookupStack(stack []*LookupElement) string {
	str := ""
	for level, node := range stack {
		str += fmt.Sprintf(
			"%s - %s\n",
			strings.Repeat("\t", level),
			node._stringify(),
		)
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
	conf := GetConfig()
	// build a "fake" root element
	// this massively simplifies code since we
	// a) always only have a single root element
	// b) can make sure that we never have to swap the root
	rootIP, _ := IP.NewIPNetFromMixed("0.0.0.0", 0)
	rootElement, _ := NewLookupElement(&ipMap{
		IPNet:    rootIP,
		Filename: conf.DefaultPACFile,
	}, rootPAC.PAC, conf.ContactInfo)
	var root = &lookupTreeNode{
		data:     &rootElement,
		children: make([]*lookupTreeNode, 0),
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

func findInTree(root *lookupTreeNode, ip *IP.Net) (*LookupElement, []*LookupElement) {
	// Check for nil root to prevent panic
	log.Debug("findInTree", root, ip.ToString())

	for _, c := range root.children {
		// Check for nil child data or IPMap to prevent panic
		if c == nil || c.data == nil || c.data.IPMap == nil {
			continue
		}

		if ip.IsSubnetOf(c.data.IPMap.IPNet) {
			// Use a recursive call to check if we have more detailed children
			node, stack := findInTree(c, ip)
			// append to the front of the stack
			stack = append([]*LookupElement{root.data}, stack...)
			return node, stack
		}
	}

	// no child matched
	// we return our dummy root as "default"-pac
	return root.data, []*LookupElement{root.data}
}
