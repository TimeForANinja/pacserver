package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/internal"
	"github.com/timeforaninja/pacserver/internal/storage"
	"github.com/timeforaninja/pacserver/pkg/admin"
)

func main() {
	// Define command-line flags
	serveFlag := flag.Bool("serve", false, "Start the PAC server")
	testFlag := flag.Bool("test", false, "Validate configs and PACs without starting the server")
	reloadFlag := flag.Bool("reload", false, "Trigger a running server to reload its configuration")
	flag.Parse()

	// If no flags are provided, show usage
	if !*serveFlag && !*testFlag && !*reloadFlag {
		fmt.Println("Please specify one of the following flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load configuration
	err := internal.LoadConfig("config.yml")
	if err != nil {
		log.Error("Unable to load \"config.yml\". Exiting.")
		panic(err)
	}

	if *serveFlag {
		// Logging to file should only be done if we're actually serving
		// the test / reload command should send to stdout
		internal.InitEventLogger()
	}

	if *reloadFlag {
		err = admin.TriggerAdminReload(internal.GetConfig().Port, internal.GetConfig().AdminSecret)
		if err != nil {
			log.Error("Unable to trigger reload on the running server. Exiting.")
			panic(err)
		}
		log.Info("Reload request sent successfully")
		return
	}

	if *testFlag {
		// test and reload should both ensure that the zones and pacs are valid
		internal.GetConfig().IgnoreMinors = false
	}

	// Initialize caches (load PACs and zones)
	err = storage.InitCaches(internal.GetConfig().ToStorageConfig())
	if err != nil {
		log.Error("Unable to initialise Caches by loading PACs and Zones. Closing Server since we're unable to recover from this.")
		panic(err)
	}

	// If test flag is provided, we can exit since we already validated when populating the caches
	if *testFlag {
		log.Info("Configuration and PACs validated successfully")
		return
	}

	// Start the server if serve flag is provided
	if *serveFlag {
		internal.LaunchServer()
		return
	}
}
