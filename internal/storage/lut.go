package storage

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/IPLUT"
)

type lookupTreeNode = IPLUT.Node[*LookupEntry]

// buildLookupTree converts the loaded lookup entries into the in-memory IPLUT tree.
func buildLookupTree(entries []*LookupEntry, defaultPAC *LookupEntry, contactInfo string) *lookupTreeNode {
	// Use the default PAC as the root so every lookup tree has a stable fallback.
	root := buildRootEntry(defaultPAC, contactInfo)
	if root == nil || root.IPMap == nil {
		return nil
	}

	// Convert the flat entry list into IPLUT nodes before handing it to the tree builder.
	rootNode := IPLUT.NewNode(root.IPMap.IPNet, root)
	nodes := make([]*lookupTreeNode, 0, len(entries))
	for _, entry := range entries {
		if entry == nil || entry.IPMap == nil {
			continue
		}
		nodes = append(nodes, IPLUT.NewNode(entry.IPMap.IPNet, entry))
	}

	return IPLUT.Build(rootNode, nodes)
}

// buildRootEntry returns the root PAC entry or builds a fallback one when needed.
func buildRootEntry(defaultPAC *LookupEntry, contactInfo string) *LookupEntry {
	// Reuse the compiled default PAC when one is already available.
	if defaultPAC != nil {
		return defaultPAC
	}

	// fall back to a synthetic root entry so the tree can still be built.
	rootIP, _ := IP.NewIPNetFromMixed("0.0.0.0", 0)
	rootPAC := &PACTemplate{
		Filename: "fallback.pac",
		Content: `// Fallback Root-PAC
function FindProxyForURL(url, host) {
  return "DIRECT"
}`,
	}
	entry, err := NewLookupEntry(&IPMap{
		IPNet:    rootIP,
		Filename: rootPAC.Filename,
	}, rootPAC, contactInfo)
	if err != nil {
		log.Warnf("Failed to build fallback root entry: %v", err)
		return nil
	}
	return entry
}

// buildLookupEntries resolves every zone against the freshly loaded PAC templates, falling back to cache when needed.
func buildLookupEntries(newIPMaps []*IPMap, newPACs map[string]*PACTemplate, oldPACs map[string]*PACTemplate, contact string) ([]*LookupEntry, map[string]*PACTemplate, int) {
	problemCounter := 0
	res := make([]*LookupEntry, 0, len(newIPMaps))
	keepPACs := make(map[string]*PACTemplate)

	// Resolve every zone against the freshly loaded PAC templates, falling back to cache when needed.
	for _, ipm := range newIPMaps {
		if ipm == nil {
			continue
		}

		pac := newPACs[ipm.Filename]
		if pac == nil {
			pac = oldPACs[ipm.Filename]
			if pac != nil {
				log.Warnf("Unknown PAC %s, using available cached version", ipm.Filename)
				// store in keepPACs to force keeping it cached
				keepPACs[pac.Filename] = pac
				problemCounter++
			} else {
				log.Warnf("Unknown PAC %s, skipping zone %s", ipm.Filename, ipm.IPNet.ToString())
				problemCounter++
				continue
			}
		}

		entry, err := NewLookupEntry(ipm, pac, contact)
		if err != nil {
			log.Warnf("Failed to compile template %s for zone %s: %s", pac.Filename, ipm.IPNet.ToString(), err.Error())
			problemCounter++
			continue
		}
		res = append(res, entry)
	}

	return res, keepPACs, problemCounter
}

// loadSpecialEntry compiles the default or WPAD PAC file and reuses the cached copy on failure.
func loadSpecialEntry(path, contactInfo string, fallback *LookupEntry) (*LookupEntry, int) {
	// The special PACs are compiled separately because they are not part of the zone table.
	absPath, err := filepath.Abs(path)
	if err != nil {
		if fallback != nil {
			log.Warnf("Failed to resolve special PAC %q, keeping cached copy: %s", path, err.Error())
			return fallback, 1
		}
		log.Errorf("Failed to resolve special PAC %q: %s", path, err.Error())
		return nil, 1
	}

	pac, err := readPACTemplate(absPath, path)
	if err != nil {
		if fallback != nil {
			log.Warnf("Failed to load special PAC %q, keeping cached copy: %s", path, err.Error())
			return fallback, 1
		}
		log.Errorf("Failed to load special PAC %q: %s", path, err.Error())
		return nil, 1
	}

	rootIP, _ := IP.NewIPNetFromMixed("0.0.0.0", 0)
	entry, err := NewLookupEntry(&IPMap{
		IPNet:    rootIP,
		Filename: pac.Filename,
	}, pac, contactInfo)
	if err != nil {
		log.Errorf("Failed to compile special PAC %q: %s", path, err.Error())
		if fallback != nil {
			return fallback, 1
		}
		return nil, 1
	}

	return entry, 0
}
