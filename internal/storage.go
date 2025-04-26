package internal

/**
 * this file is the initiator for our storage model
 *
 * it handles creating Lookup Trees and Lists when required,
 * starts regular tasks to refresh the data (if required),
 * and decides on how to react to minor problems (e.g. missing PACs, invalid zones, ...)
 */

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

var lookupTree *lookupTreeNode

// InitCaches does an initial fetch of all Zones and PAC Files
// this differs from the automated lookup in that it also errors out when minor problems are found
func InitCaches() error {
	config := GetConfig()
	problemCounter := updateLookupTree()
	if problemCounter > 0 {
		log.Errorf("There were %d minor problems while initialising caches. Please check the logs for details.", problemCounter)
		if !config.IgnoreMinors {
			return errors.New("zones or pac files includes errors - exiting")
		}
	}
	log.Info("Finished initial loading of IPMap and PACs - starting")

	// start a regular task to refresh the lookup tree
	if config.MaxCacheAge > 0 {
		go executeRegular(updateLookupTree)
	}

	return nil
}

func executeRegular(task func() int) {
	tick := time.Tick(time.Duration(GetConfig().MaxCacheAge) * time.Second)
	for range tick {
		log.Infof("Max Cache Age reached - Refreshing Lookup Tree")
		task()
	}
}

func updateLookupTree() int {
	config := GetConfig()
	// first we build a "flat" lookup element list
	// this maps IPMap to PAC
	table, minorProblems := buildLookupElementList(config.IPMapFile, config.PACRoot, config.ContactInfo)
	// then we build an optimized lookup tree to faster serve clients
	lookupTree = buildLookupTree(table)
	log.Infof("The following LookupTree was loaded:\n%s", stringifyLookupTree(lookupTree))
	return minorProblems
}
