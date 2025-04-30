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
var wpadPAC *LookupElement
var rootPAC *LookupElement

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

func loadDefaults() int {
	config := GetConfig()

	problemCounter := 0
	log.Debugf("Trying to load default PAC (%s) and WPAD (%s)", config.DefaultPACFile, config.WPADFile)

	rawDefault, err1 := readAndParse(".", config.DefaultPACFile)
	if err1 == nil {
		newRootPAC, err2 := NewLookupElement(&ipMap{}, rawDefault, config.ContactInfo)
		if err2 == nil {
			// assign to cached root pac if successful
			rootPAC = &newRootPAC
		} else {
			problemCounter++
			log.Errorf("Failed to parse Default PAC File \"%s\": %s", config.DefaultPACFile, err2.Error())
		}
	} else {
		problemCounter++
		log.Errorf("Failed to read Default PAC File \"%s\": %s", config.DefaultPACFile, err1.Error())
	}

	rawWPAD, err1 := readAndParse(".", config.WPADFile)
	if err1 == nil {
		newWPAD, err2 := NewLookupElement(&ipMap{}, rawWPAD, config.ContactInfo)
		if err2 == nil {
			// assign to cached wpad if successful
			wpadPAC = &newWPAD
		} else {
			problemCounter++
			log.Errorf("Failed to parse WPAD File \"%s\": %s", config.WPADFile, err2.Error())
		}
	} else {
		problemCounter++
		log.Errorf("Failed to read WPAD File \"%s\": %s", config.WPADFile, err1.Error())
	}

	return problemCounter
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
	// reload default PACs
	minorProblems1 := loadDefaults()
	// first we build a "flat" lookup element list
	// this maps IPMap to PAC
	table, minorProblems2 := buildLookupElementList(config.IPMapFile, config.PACRoot, config.ContactInfo)
	// then we build an optimized lookup tree to faster serve clients
	lookupTree = buildLookupTree(table)
	log.Infof("The following LookupTree was loaded:\n%s", stringifyLookupTree(lookupTree))
	return minorProblems1 + minorProblems2
}
