package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/IPLUT"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

var cache struct {
	sync.RWMutex
	// latest parsed tree
	lookupTree *lookupTreeNode
	routeLUT   map[string]*LookupEntry
	// static mappers for special pacs
	defaultPAC *LookupEntry
	// cached lists from the last tree parsing
	cachedIPMap []*IPMap
	cachedPACs  map[string]*PACTemplate
}

// InitCaches loads all file-backed data and populates the in-memory cache.
func InitCaches(cfg StorageConfig) error {
	if cfg == (StorageConfig{}) {
		return fmt.Errorf("storage config must not be empty")
	}

	// Build the lookup tree once at startup so request handling stays read-only.
	problems := UpdateLookupTree(cfg)
	if problems > 0 && !cfg.IgnoreMinors {
		return fmt.Errorf("zones or pac files includes errors - exiting")
	}
	return nil
}

// UpdateLookupTree reloads the backing files and rebuilds the lookup tree.
func UpdateLookupTree(cfg StorageConfig) int {
	log.Info("Reload - Initialized")

	// Snapshot the current cache so we can keep serving it if reload input is incomplete.
	cache.RLock()
	oldIPMaps := append([]*IPMap(nil), cache.cachedIPMap...)
	oldPACs := utils.MapClone(cache.cachedPACs)
	oldDefault := cache.defaultPAC
	oldTree := cache.lookupTree
	cache.RUnlock()

	problemCounter := 0

	// reload files from disk
	newIPMaps, err1, probs1 := ReadIPMaps(cfg.IPMapFile)
	var newRoutes []*RouteMap
	var routeErr error
	if strings.TrimSpace(cfg.RouteMapFile) != "" {
		newRoutes, routeErr, problemCounter = readRoutesForReload(cfg.RouteMapFile, problemCounter)
	}
	problemCounter += probs1
	newPACs, err2, probs2 := ReadPACTemplates(cfg.PACRoot)
	problemCounter += probs2
	log.Info("Reload - Loaded Config Files")

	// If both sources fail, reuse the existing cache
	if err1 != nil && err2 != nil {
		log.Errorf("Completely failed to load IPMap and PACs - keep serving cached data")
		return problemCounter
	}
	// If only one source fails, reuse the cached data for that source and the new data for the other
	if err1 != nil {
		log.Errorf("Completely failed to load IPMap - reusing cached zones")
		newIPMaps = oldIPMaps
		problemCounter++
	}
	if err2 != nil {
		log.Errorf("Completely failed to load PACs - reusing cached templates")
		newPACs = utils.MapToArray(oldPACs)
		problemCounter++
	}

	// correlate the ip-maps and pac-templates
	newPACIndex := utils.MapFromArray(newPACs, func(p *PACTemplate) string { return p.Filename })
	entries, keepPACs, probs3 := buildLookupEntries(newIPMaps, newPACIndex, oldPACs, cfg.ContactInfo)
	problemCounter += probs3
	log.Info("Reload - IP List and PACs merged")

	// load non-standard pacs
	defaultPAC, probs4 := loadSpecialEntry(cfg.DefaultPACFile, cfg.ContactInfo, oldDefault)
	problemCounter += probs4
	log.Info("Reload - Special-PACs build")

	// Rebuild the tree from the freshly parsed zones and templates.
	newTree := buildLookupTree(entries, defaultPAC, cfg.ContactInfo)
	newRouteLUT := make(map[string]*LookupEntry)
	if routeErr == nil {
		for _, route := range newRoutes {
			pac := newPACIndex[route.Filename]
			if pac == nil {
				problemCounter++
				continue
			}
			entry, err := NewRouteLookupEntry(route.Route, pac, cfg.ContactInfo)
			if err != nil {
				problemCounter++
				continue
			}
			newRouteLUT[strings.ToLower(route.Route)] = entry
		}
	} else {
		cache.RLock()
		newRouteLUT = cache.routeLUT
		cache.RUnlock()
	}
	if newTree == nil {
		log.Warn("Reload - LUT build failed, keeping previous tree")
		newTree = oldTree
	}
	log.Info("Reload - LUT build")

	// Preserve any templates that still need to be served even if they were not referenced by zones.
	mergedPACs := utils.MergeMaps(newPACIndex, keepPACs)

	// Swap the new tree into place atomically so readers never see a partially rebuilt cache.
	cache.Lock()
	cache.lookupTree = newTree
	cache.routeLUT = newRouteLUT
	if defaultPAC != nil {
		cache.defaultPAC = defaultPAC
	} else {
		cache.defaultPAC = oldDefault
	}
	cache.cachedIPMap = newIPMaps
	cache.cachedPACs = mergedPACs
	cache.Unlock()

	log.Infof("The following IPLUT was loaded:\n%s", IPLUT.Stringify(newTree))
	return problemCounter
}

func readRoutesForReload(filename string, problems int) ([]*RouteMap, error, int) {
	routes, err, routeProblems := ReadRoutes(filename)
	return routes, err, problems + routeProblems
}

func FindRoute(route string) *LookupEntry {
	cache.RLock()
	defer cache.RUnlock()
	return cache.routeLUT[strings.ToLower("/"+strings.TrimSpace(strings.Trim(route, "/")))]
}

// FindInLUT resolves the best matching lookup entry for the given IP.
func FindInLUT(ipStr string, networkBits int) (*LookupEntry, *IP.Net, []*LookupEntry) {
	// Read the current tree and fallback entry under lock, then release quickly.
	cache.RLock()
	tree := cache.lookupTree
	defaultPAC := cache.defaultPAC
	cache.RUnlock()

	// If the tree is unavailable, return the default PAC when possible so the server still answers.
	if tree == nil {
		if defaultPAC == nil {
			return nil, &IP.Net{}, nil
		}
		return defaultPAC, &IP.Net{}, []*LookupEntry{defaultPAC}
	}

	// validate the passed ip network data
	ipNet, err := IP.NewIPNetFromMixed(ipStr, networkBits)
	if err != nil {
		if defaultPAC == nil {
			return nil, &IP.Net{}, nil
		}
		return defaultPAC, &IP.Net{}, []*LookupEntry{defaultPAC}
	}

	// Walk the LUT and return the best matching node stack for debug output.
	node, stack := IPLUT.Find(tree, &ipNet)
	if node == nil {
		if defaultPAC == nil {
			return nil, &ipNet, nil
		}
		return defaultPAC, &ipNet, []*LookupEntry{defaultPAC}
	}

	resultStack := IPLUT.StackContent(stack)
	return node.Content, &ipNet, resultStack
}

// DefaultPAC returns the compiled default PAC entry.
func DefaultPAC() *LookupEntry {
	// Expose the cached default PAC without making callers know about the lock.
	cache.RLock()
	defer cache.RUnlock()
	return cache.defaultPAC
}
