from dataclasses import dataclass
from typing import List, Optional

from internal.ProviderIPMap import IPMap
from internal.LookupElement import LookupElement
from pkg.IP.ipnet import Net

@dataclass
class LookupTreeNode:
    data: 'LookupElement'
    children: List['LookupTreeNode'] = None

    def __post_init__(self):
        if self.children is None:
            self.children = []

def stringify_lookup_tree(root: LookupTreeNode) -> str:
    return _stringify_lookup_tree(root, 0)

def _stringify_lookup_tree(node: LookupTreeNode, level: int) -> str:
    result = "{indent}- {ip_net} | pac({filename})\n".format(
        indent=" " * level,
        ip_net=node.data.ip_map.ip_net.to_string(),
        filename=node.data.ip_map.filename
    )

    for child in node.children:
        result += _stringify_lookup_tree(child, level + 1)

    return result

def insert_tree_element(root: LookupTreeNode, elem: 'LookupElement') -> None:
    new_node = LookupTreeNode(data=elem)

    i = 0
    while i < len(root.children):
        child = root.children[i]
        # Check if the elem is a subnet of the child
        if elem.is_subnet_of(child.data):
            insert_tree_element(child, elem)
            return
        # Check if the child is a subnet of the elem
        # Or identical (which simply get stacked)
        if child.data.is_subnet_of(elem) or elem.is_identical_net(root.data):
            # push child into new node
            new_node.children.append(child)
            # remove child from root
            root.children.pop(i)
            continue
        i += 1
    # If no subnet relation found, add new_node as a child
    root.children.append(new_node)

def build_lookup_tree(elements: List['LookupElement']) -> LookupTreeNode:
    # build a "fake" root element
    # this massively simplifies code since we
    # a) always only have a single root element
    # b) can make sure that we never have to swap the root
    root_ip = Net.new_from_mixed("0.0.0.0", 0)  # Assuming this method exists
    root = LookupTreeNode(
        data=LookupElement(
            ip_map=IPMap(ip_net=root_ip, filename=""),
            pac=None
        )
    )

    # insert one element after another into our tree
    for elem in elements:
        insert_tree_element(root, elem)

    # check if the user provided a single root
    # and if so, replace our custom root by it
    if len(root.children) == 1 and root.children[0].data.get_raw_cidr() == 0:
        root = root.children[0]

    # remove "intermediate" nodes
    # that serve the same pac as their parent
    simplify_tree(root)

    return root

def simplify_tree(root: LookupTreeNode) -> None:
    simplified_children: List[LookupTreeNode] = []
    for child in root.children:
        # Recursively simplify child nodes
        simplify_tree(child)
        # If the simplified child is not identical to the root, keep it
        # The following are the two conditions we check
        a = child.data.is_identical_pac(root.data)
        b = child.data.is_identical_net(root.data)
        # xor
        if not (a and b):
            simplified_children.append(child)
        else:
            simplified_children.extend(child.children)

    # Replace the children with the simplified children
    root.children = simplified_children

def find_in_tree(root: LookupTreeNode, ip: Net) -> Optional['LookupElement']:
    for child in root.children:
        if ip.is_subnet_of(child.data.ip_map.ip_net):
            return find_in_tree(child, ip)

    # no child matched
    # check if it's our fake root - this would mean no rule matches the location
    if root.data.pac is None:
        return None
    return root.data
