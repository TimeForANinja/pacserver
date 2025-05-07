package internal

/**
 * this file reads in the zone file and parses it into a list of IP Maps
 */

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/timeforaninja/pacserver/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/IP"
)

type ipMap struct {
	IPNet    IP.Net `json:"IPNet"`
	Filename string `json:"Filename"`
	Comment  string `json:"Comment"`
}

func (x1 *ipMap) CompareForSort(x2 *ipMap) bool {
	// First compare by network address
	if x1.IPNet.NetworkAddress.Value != x2.IPNet.NetworkAddress.Value {
		return x1.IPNet.NetworkAddress.Value < x2.IPNet.NetworkAddress.Value
	}
	// If network addresses are equal, compare by CIDR (more specific networks come later)
	return x1.IPNet.CIDR.Value < x2.IPNet.CIDR.Value
}

func readIPMap(relPath string) ([]*ipMap, error, int) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Errorf("Invalid Filepath for IPMap found: \"%s\": %s", absPath, err.Error())
		return make([]*ipMap, 0), err, 1
	}
	file, err := os.Open(absPath)
	if err != nil {
		log.Errorf("Unable to open IPMap at \"%s\": %s", absPath, err.Error())
		return make([]*ipMap, 0), err, 1
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mappings []*ipMap

	problemCounter := 0
	lineCount := 0
	for scanner.Scan() {
		// read next line
		textLine := scanner.Text()
		lineCount++

		mapping, err := parseIPMapLine(textLine)
		if err != nil {
			log.Errorf("Failed to parse CSV Line %d: %s", lineCount, err.Error())
			problemCounter++
		}
		// mapping=nil and error=nil for skipping lines
		if mapping != nil {
			// if we made it this far then store the zone
			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil, problemCounter
}

func parseIPMapLine(line string) (*ipMap, error) {
	// Skip comment-lines
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
		return nil, nil
	}

	// Skip empty lines
	if len(line) == 0 {
		return nil, nil
	}

	// parse line as csv
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = ',' // set comma as the field delimiter
	fields, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to parse line as csv: %s", err.Error())
	}

	// trim whitespace around all fields
	for i, field := range fields {
		fields[i] = strings.TrimSpace(field)
	}

	// Ensure the CSV has exactly two fields
	if len(fields) < 3 || len(fields) > 4 {
		return nil, fmt.Errorf("invalid number of fields, expected 3-4 but got %d", len(fields))
	}

	ipNet, err := IP.NewIPNetFromStr(fields[0], fields[1])
	if err != nil {
		return nil, fmt.Errorf("unable to parse IP: %s", err.Error())
	}

	newMap := ipMap{
		IPNet:    ipNet,
		Filename: utils.NormalizePath(fields[2]),
	}
	// assign Comment if we found one
	if len(fields) == 4 {
		newMap.Comment = fields[3]
	}
	return &newMap, nil
}
