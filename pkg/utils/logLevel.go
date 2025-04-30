package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"slices"
	"strings"
)

func GetLoglevel(level string) log.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return log.LevelDebug
	case "INFO":
		return log.LevelInfo
	case "WARN":
		return log.LevelWarn
	case "ERROR":
		return log.LevelError
	}
	return log.LevelInfo
}

func ValidateLogLevel(level string) error {
	knownLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if !slices.Contains(knownLevels, strings.ToUpper(level)) {
		return fmt.Errorf("loglevel must be one of %v", knownLevels)
	}
	return nil
}
