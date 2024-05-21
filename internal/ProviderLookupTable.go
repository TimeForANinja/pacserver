package internal

import (
	"github.com/gofiber/fiber/v2/log"
)

func buildLookupElements(ipMapFile, pacRoot, contactInfo string) []*LookupElement {
	// store current cached PACs
	// they can be useful when calculating LookupElements
	// if some pac has been partially deleted by accident
	oldPACs := cachedPACs

	// read new PACs / Zones
	newIPMaps, err1 := readIPMap(ipMapFile)
	newPACs, err2 := readTemplateFiles(pacRoot)

	// check if the loading worked
	// if not print error and try to use cached version
	// if yes then update the cache
	if err1 != nil && err2 != nil {
		log.Errorf("Completely failed to load IPMap and PACs - keep serving cached data")
		// no need to recalculate Tree since nothing can change
		return nil
	} else if err1 != nil {
		log.Errorf("Completely failed to load IPMap - loading new? PACs with cached Zones")
		newIPMaps = cachedIPMaps
		cachedPACs = newPACs
	} else if err2 != nil {
		log.Errorf("Completely failed to load PACs - loading new? Zones with cached PACs")
		newPACs = oldPACs
		cachedIPMaps = newIPMaps
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

			// after checking the cache, write log
			if match != nil {
				log.Warnf("Unknown PAC %s, using available Cached Version", ipm.Filename)
				// keep the old pac in the cache for the next check
				newPACs = append(newPACs, match)
			} else {
				log.Warnf("Unknown PAC %s, no Cached Version available, skipping Zone %s", ipm.Filename, ipm.IPNet.ToString())
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
				continue
			}
			res = append(res, &le)
		}
	}
	return res
}
