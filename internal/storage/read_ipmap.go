package storage

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/IP"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

// ReadIPMaps loads and parses the zone CSV file into structured rows.
func ReadIPMaps(relPath string) ([]*IPMap, error, int) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Errorf("Invalid Filepath for IPMap found: %q: %s", relPath, err.Error())
		return []*IPMap{}, err, 1
	}

	file, err := os.Open(absPath)
	if err != nil {
		log.Errorf("Unable to open IPMap at %q: %s", absPath, err.Error())
		return []*IPMap{}, err, 1
	}
	defer file.Close()

	return parseIPMapCSV(file)
}

func parseIPMapCSV(r io.Reader) ([]*IPMap, error, int) {
	// Scan line by line so one bad row does not block the rest of the file.
	scanner := bufio.NewScanner(r)
	mappings := make([]*IPMap, 0)
	problemCounter := 0
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		mapping, err := parseIPMapLine(scanner.Text())
		if err != nil {
			log.Errorf("Failed to parse CSV Line %d: %s", lineCount, err.Error())
			problemCounter++
			continue
		}
		if mapping != nil {
			mappings = append(mappings, mapping)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Failed to read IPMap CSV: %s", err.Error())
		return mappings, err, problemCounter + 1
	}

	return mappings, nil, problemCounter
}

func parseIPMapLine(line string) (*IPMap, error) {
	// Skip comments and blank lines because the CSV is also used as a human-edited file.
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return nil, nil
	}

	// Let the CSV parser handle quoting so filenames and comments can stay readable.
	fields, err := csv.NewReader(strings.NewReader(line)).Read()
	if err != nil {
		return nil, fmt.Errorf("unable to parse line as csv: %w", err)
	}

	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	if len(fields) < 3 || len(fields) > 4 {
		return nil, fmt.Errorf("invalid number of fields, expected 3-4 but got %d", len(fields))
	}

	// Convert the textual range into a normalized IP network for later tree construction.
	ipNet, err := IP.NewIPNetFromStr(fields[0], fields[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse IP: %w", err)
	}

	newMap := &IPMap{
		IPNet:    ipNet,
		Filename: utils.NormalizePath(fields[2]),
	}
	if len(fields) == 4 {
		newMap.Comment = fields[3]
	}

	return newMap, nil
}
