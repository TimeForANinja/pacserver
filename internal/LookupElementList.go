package internal

/**
 * LookupElementList is an intermediate data structure
 *
 * it is the basic list initially build from correlating the IPMap and PACFiles
 * and later gets converted to the LookupTree for faster serving
 *
 * it also does the caching of previous IPMap and PAC files,
 * which allows for more resilience to minor problems with config changes
 */

import (
	"github.com/gofiber/fiber/v2/log"
)

var (
	cachedIPMaps = make([]*ipMap, 0)
	cachedPACs   = make([]*pacTemplate, 0)
)

// buildLookupElementList reads the IPMap and PACFiles from disk
// and tries to convert them into a flat list of Lookup Elements
func buildLookupElementList(ipMapFile, pacRoot, contactInfo string) ([]*LookupElement, int) {
	problemCounter := 0
	// store current cached PACs
	// they can be useful when calculating LookupElements
	// if some pac has been partially deleted by accident
	oldPACs := cachedPACs

	// read new PACs / Zones
	newIPMaps, err1, probs1 := readIPMap(ipMapFile)
	problemCounter += probs1
	newPACs, err2, probs2 := readTemplateFiles(pacRoot)
	problemCounter += probs2

	// check if the loading worked
	// if not print error and try to use cached version
	// if yes then update the cache
	if err1 != nil && err2 != nil {
		log.Errorf("Completely failed to load IPMap and PACs - keep serving cached data")
		// no need to recalculate Tree since nothing can change
		return nil, 1
	} else if err1 != nil {
		log.Errorf("Completely failed to load IPMap - loading new? PACs with cached Zones")
		newIPMaps = cachedIPMaps
		cachedPACs = newPACs
		problemCounter++
	} else if err2 != nil {
		log.Errorf("Completely failed to load PACs - loading new? Zones with cached PACs")
		newPACs = oldPACs
		cachedIPMaps = newIPMaps
		problemCounter++
	} else {
		cachedIPMaps = newIPMaps
		cachedPACs = newPACs
	}

	// build new lookup elements
	res := make([]*LookupElement, 0)
	for _, ipm := range newIPMaps {
		// for each IPMap, (try to) find the corresponding pac
		var match *pacTemplate
		for _, p := range newPACs {
			if p.Filename == ipm.Filename {
				match = p
				break
			}
		}

		// did not find one, try checking the cached versions
		if match == nil {
			for _, p := range oldPACs {
				if p.Filename == ipm.Filename {
					match = p
					break
				}
			}

			// after checking the cache, write a log
			if match != nil {
				log.Warnf("Unknown PAC %s, using available Cached Version", ipm.Filename)
				// keep the old pac in the cache for the next check
				newPACs = append(newPACs, match)
				problemCounter++
			} else {
				log.Warnf("Unknown PAC %s, no Cached Version available, skipping Zone %s", ipm.Filename, ipm.IPNet.ToString())
				problemCounter++
			}
		}

		// if we found a match (after checking new and cached PACs)
		// then try to parse it
		if match != nil {
			le, err := NewLookupElement(ipm, match, contactInfo)
			if err != nil {
				// NewLookupElement only fails when the Template could not be filled with the variables
				// Log it, and recover by skipping this zone
				log.Warnf("Failed to compile Template %s for zone %s: %s", match.Filename, ipm.IPNet.ToString(), err.Error())
				problemCounter++
				continue
			}
			res = append(res, &le)
		}
	}
	return res, problemCounter
}
