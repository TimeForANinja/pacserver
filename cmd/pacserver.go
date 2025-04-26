package main

import (
	"flag"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/internal"
	"os"
	"syscall"
)

func main() {
	// Define command-line flags
	serveFlag := flag.Bool("serve", false, "Start the PAC server")
	testFlag := flag.Bool("test", false, "Validate configs and PACs without starting the server")
	reloadFlag := flag.Bool("reload", false, "Tell a running server to reload PACs and config")
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

	if *testFlag || *reloadFlag {
		// test and reload should both ensure that the zones and pacs are valid
		internal.GetConfig().IgnoreMinors = false
	}

	// Initialize caches (load PACs and zones)
	err = internal.InitCaches()
	if err != nil {
		log.Error("Unable to initialise Caches by loading PACs and Zones. Closing Server since we're unable to recover from this.")
		panic(err)
	}

	// Handle reload flag
	if *reloadFlag {
		err := reload()
		if err != nil {
			os.Exit(1)
		}
		return
	}

	// If test flag is provided, just validate and exit
	if *testFlag {
		internal.GetConfig().IgnoreMinors = false
		log.Info("Configuration and PACs validated successfully")
		return
	}

	// Start the server if serve flag is provided
	if *serveFlag {
		internal.LaunchServer()
		return
	}

	// If we get here, no valid action was specified
	fmt.Println("Please specify one of the following flags:")
	flag.PrintDefaults()
	os.Exit(1)
}

func reload() error {
	// config & eventlogger already init in main

	// Read the PID from the PID file
	pid, err := internal.ReadPidFile()
	if err != nil {
		log.Errorf("Failed to read PID file: %v", err)
		return err
	}

	// Send SIGHUP signal to the running server process
	log.Infof("Sending SIGHUP signal to process %d", pid)
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Errorf("Failed to find process with PID %d: %v", pid, err)
		return err
	}

	err = proc.Signal(syscall.SIGHUP)
	if err != nil {
		log.Errorf("Failed to send SIGHUP signal to process %d: %v", pid, err)
		return err
	}

	log.Infof("SIGHUP signal sent successfully to process %d", pid)
	return nil
}
