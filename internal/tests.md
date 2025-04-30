# Unit Tests in the Internal Module

This document provides a comprehensive list of all unit tests in the internal module, organized by test file with detailed test cases.

## LookupElement_test.go

Tests for the LookupElement struct and its methods.

| Test Case                                    | Tested Function    | Description of Input                                                    | Description of Expected Output                |
|----------------------------------------------|--------------------|-------------------------------------------------------------------------|-----------------------------------------------|
| Identical networks                           | `isIdenticalNet`   | Two LookupElements with same IP (192.168.0.0) and CIDR (24)             | Returns true                                  |
| Different networks - different IP            | `isIdenticalNet`   | Two LookupElements with different IPs (192.168.0.0 vs 10.0.0.0)         | Returns false                                 |
| Different networks - different CIDR          | `isIdenticalNet`   | Two LookupElements with same IP but different CIDRs (24 vs 16)          | Returns false                                 |
| Identical PAC filenames                      | `isIdenticalPAC`   | Two LookupElements with same PAC filename                               | Returns true                                  |
| Different PAC filenames                      | `isIdenticalPAC`   | Two LookupElements with different PAC filenames                         | Returns false                                 |
| First PAC is nil                             | `isIdenticalPAC`   | First LookupElement has nil PAC                                         | Returns false                                 |
| Second PAC is nil                            | `isIdenticalPAC`   | Second LookupElement has nil PAC                                        | Returns false                                 |
| Both PACs are nil                            | `isIdenticalPAC`   | Both LookupElements have nil PACs                                       | Returns false                                 |
| le1 is subnet of le2                         | `isSubnetOf`       | First network (192.168.0.0/24) is subnet of second (192.168.0.0/16)     | Returns true                                  |
| le1 is not subnet of le2 - different network | `isSubnetOf`       | Networks are completely different (10.0.0.0/24 vs 192.168.0.0/16)       | Returns false                                 |
| le1 is not subnet of le2 - le1 is broader    | `isSubnetOf`       | First network is broader than second (192.168.0.0/16 vs 192.168.0.0/24) | Returns false                                 |
| le1 is identical to le2                      | `isSubnetOf`       | Networks are identical (both 192.168.0.0/24)                            | Returns true                                  |
| CIDR 24                                      | `getRawCIDR`       | LookupElement with CIDR 24                                              | Returns 24                                    |
| CIDR 16                                      | `getRawCIDR`       | LookupElement with CIDR 16                                              | Returns 16                                    |
| CIDR 8                                       | `getRawCIDR`       | LookupElement with CIDR 8                                               | Returns 8                                     |
| CIDR 0                                       | `getRawCIDR`       | LookupElement with CIDR 0                                               | Returns 0                                     |
| Valid template                               | `NewLookupElement` | Valid template with proper variables                                    | Creates LookupElement with processed template |
| Invalid template                             | `NewLookupElement` | Template with invalid variable                                          | Returns error                                 |

## LookupElementList_test.go

Tests for the functions in LookupElementList.go.

| Test Case                  | Tested Function    | Description of Input                                     | Description of Expected Output                     |
|----------------------------|--------------------|----------------------------------------------------------|----------------------------------------------------|
| All PACs found in newPACs  | `matchIPMapToPac ` | Two IP maps with matching PACs in newPACs                | Returns two elements, no PACs to keep, no problems |
| Some PACs found in oldPACs | `matchIPMapToPac`  | Two IP maps, one matching PAC in newPACs, one in oldPACs | Returns two elements, one PAC to keep, one problem |
| Some PACs not found at all | `matchIPMapToPac`  | Two IP maps, one matching PAC in newPACs, one not found  | Returns one element, no PACs to keep, one problem  |
| No PACs found              | `matchIPMapToPac`  | Two IP maps, no matching PACs                            | Returns no elements, no PACs to keep, two problems |
| Empty inputs               | `matchIPMapToPac`  | Empty arrays for all inputs                              | Returns no elements, no PACs to keep, no problems  |

## LookupElementTree_test.go

Tests for the functions in LookupElementTree.go.

| Test Case                                                               | Tested Function       | Description of Input                                           | Description of Expected Output                                |
|-------------------------------------------------------------------------|-----------------------|----------------------------------------------------------------|---------------------------------------------------------------|
| Does not Error when only feed with dummy root, but defaults to root-pac | `findInTree`          | Tree with only root node, IP 192.168.0.0/32                    | Returns root element                                          |
| Unknown Element defaults to root                                        | `findInTree`          | Tree with multiple nodes, IP 0.0.0.0/32                        | Returns global element                                        |
| Most specific Node gets matched                                         | `findInTree`          | Tree with multiple nodes, IP 192.168.0.0/32                    | Returns most specific matching element (child2Element)        |
| Works for searching networks                                            | `findInTree`          | Tree with multiple nodes, IP 192.168.0.0/16                    | Returns matching network element (child1Element)              |
| Empty Input                                                             | `buildLookupTree`     | Empty array of LookupElements                                  | Returns tree with default root node                           |
| Overwrites build-in with explicit default                               | `buildLookupTree`     | Array with one global element (0.0.0.0/0)                      | Returns tree with custom global element                       |
| Nested Networks                                                         | `buildLookupTree`     | Array with two elements in hierarchical order                  | Returns properly nested tree structure                        |
| Nested (unordered) Networks                                             | `buildLookupTree`     | Array with two elements in reverse hierarchical order          | Returns properly nested tree structure (same as ordered case) |
| Duplicate Networks (get removed by simplify)                            | `buildLookupTree`     | Array with duplicate networks and a more specific one          | Returns tree with duplicates removed, keeping first element   |
| (No specific test case name)                                            | `simplifyTreeSorting` | Tree with unsorted children                                    | Returns tree with children sorted by network address          |
| Safe removal when iterating backwards                                   | `insertTreeElement`   | Root node with two children, inserting parent element          | Both children moved under new parent element                  |
| Safe removal of last element                                            | `insertTreeElement`   | Root node with one child, inserting parent element             | Child moved under new parent element                          |
| Safe removal of middle element                                          | `insertTreeElement`   | Root node with three children, inserting parent for middle one | Middle child moved under new parent, others remain at root    |

## LookupElementTree_Blackbox_test.go

End-to-end tests for the LookupElementTree functionality.

| Test Case                                                          | Tested Function            | Elements Input                                        | IP Input                      | Description of Expected Output                             |
|--------------------------------------------------------------------|----------------------------|------------------------------------------------------|-------------------------------|------------------------------------------------------------|
| No elements at all                                                 | `TestBuildAndFindCombined` | Empty array of elements                               | IP 192.168.1.1/32             | Returns default root element                               |
| A 0.0.0.0/0 element that overwrites the root                       | `TestBuildAndFindCombined` | Array with custom global element                      | IP 10.0.0.1/32                | Returns custom root element                                |
| Identical IPNet for two objects                                    | `TestBuildAndFindCombined` | Array with two elements having identical networks     | IP 192.168.0.1/32             | Returns one of the elements (implementation returns first) |
| Identical with cidr > 32                                           | `TestBuildAndFindCombined` | Element with invalid CIDR (33)                        | IP 192.168.0.1/32             | Handles invalid CIDR gracefully                            |
| findInTree with 0.0.0.0/0                                          | `TestBuildAndFindCombined` | Array with multiple elements                          | IP 0.0.0.0/0                  | Returns root element                                       |
| findInTree with a network that has two identical IPNet elements    | `TestBuildAndFindCombined` | Array with two elements having identical networks     | IP 192.168.0.0/24             | Returns one of the elements                                |
| findInTree with invalid IP.Net Network Address                     | `TestBuildAndFindCombined` | Empty array                                           | IP with invalid network address| Returns root element                                       |
| findInTree with invalid IP.Net CIDR                                | `TestBuildAndFindCombined` | Empty array                                           | IP with invalid CIDR          | Returns root element                                       |
| findInTree with an IP that is initialised with only default values | `TestBuildAndFindCombined` | Empty array                                           | IP with default values        | Returns root element                                       |

## readIPMap_test.go

Tests for the functions in readIPMap.go.

| Test Case                                      | Tested Function  | Description of Input                                                                | Description of Expected Output                     |
|------------------------------------------------|------------------|-------------------------------------------------------------------------------------|----------------------------------------------------|
| Different network addresses - x1 < x2          | `CompareForSort` | Two ipMaps with different network addresses (10.0.0.0/8 vs 192.168.0.0/16)          | Returns true (first comes before second)           |
| Different network addresses - x1 > x2          | `CompareForSort` | Two ipMaps with different network addresses (192.168.0.0/16 vs 10.0.0.0/8)          | Returns false (first comes after second)           |
| Same network address, different CIDR - x1 < x2 | `CompareForSort` | Two ipMaps with same network but different CIDRs (192.168.0.0/16 vs 192.168.0.0/24) | Returns true (broader network comes first)         |
| Same network address, different CIDR - x1 > x2 | `CompareForSort` | Two ipMaps with same network but different CIDRs (192.168.0.0/24 vs 192.168.0.0/16) | Returns false (more specific network comes second) |
| Same network address and CIDR                  | `CompareForSort` | Two ipMaps with identical networks (192.168.0.0/24)                                 | Returns false (neither comes before the other)     |
| Comment line with //                           | `parseIPMapLine` | Line starting with //                                                               | Returns nil, no error                              |
| Comment line with #                            | `parseIPMapLine` | Line starting with #                                                                | Returns nil, no error                              |
| Empty line                                     | `parseIPMapLine` | Empty string                                                                        | Returns nil, no error                              |
| Valid line                                     | `parseIPMapLine` | Valid line with IP, CIDR, and filename                                              | Returns correctly parsed ipMap                     |
| Valid line with whitespace                     | `parseIPMapLine` | Valid line with whitespace around values                                            | Returns correctly parsed ipMap with trimmed values |
| Invalid number of fields (<3)                  | `parseIPMapLine` | Line with only IP and CIDR                                                          | Returns error                                      |
| Invalid number of fields (>3)                  | `parseIPMapLine` | Line with extra fields                                                              | Returns error                                      |
| Invalid IP address                             | `parseIPMapLine` | Line with invalid IP                                                                | Returns error                                      |
| Invalid CIDR                                   | `parseIPMapLine` | Line with invalid CIDR                                                              | Returns error                                      |
