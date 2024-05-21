package internal

import (
	"time"

	"github.com/gofiber/fiber/v2/log"
)

var cachedIPMaps []*ipMap
var cachedPACs []*pacTemplate

var lookupTree *lookupTreeNode

// InitCaches does an initial fetch of all Zones and PAC Files
// this allows checking all conditions that would completely break the application
// directly after start
func InitCaches() error {
	var err error
	config := getConfig()

	cachedIPMaps, err = readIPMap(config.IPMapFile)
	if err != nil {
		// Logging is already done inside the Providers
		return err
	}

	cachedPACs, err = readTemplateFiles(config.PACRoot)
	if err != nil {
		// Logging is already done inside the Providers
		return err
	}

	// initial build of the lookup tree
	updateLookupTree()
	// start a regular task to refresh the lookup tree
	if config.DoAutoRefresh {
		go executeRegular(updateLookupTree)
	}

	return nil
}

func executeRegular(task func()) {
	tick := time.Tick(time.Duration(getConfig().MaxCacheAge) * time.Second)
	for range tick {
		log.Infof("Max Cache Age reached - Refreshing Lookup Tree")
		task()
	}
}

func updateLookupTree() {
	config := getConfig()
	table := buildLookupElements(config.IPMapFile, config.PACRoot, conf.ContactInfo)
	tree := buildLookupTree(table)
	log.Infof("The following LookupTree was loaded:\n%s", stringifyLookupTree(tree))
	lookupTree = tree
}
