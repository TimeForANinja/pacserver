package main

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/internal"
)

func main() {
	err := internal.LoadConfig("config.yml")
	if err != nil {
		log.Error("Unable to load \"config.yml\". Exiting.")
		panic(err)
	}

	internal.InitEventLogger()

	err = internal.InitCaches()
	if err != nil {
		log.Error("Unable to initialise Caches by loading PACs and Zones. Closing Server since we're unable to recover from this.")
		panic(err)
	}

	internal.LaunchServer()
}
