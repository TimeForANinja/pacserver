package internal

import (
	"bufio"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/IP"
)

type ipMap struct {
	IPNet    IP.Net `json:"IPNet"`
	Filename string `json:"Filename"`
}

func readIPMap(relPath string) ([]*ipMap, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Errorf("Invalid Filepath for IPMap found: \"%s\": %s", absPath, err.Error())
		return nil, err
	}
	file, err := os.Open(absPath)
	if err != nil {
		log.Errorf("Unable to open IPMap at \"%s\": %s", absPath, err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mappings []*ipMap

	lineCount := -1
	for scanner.Scan() {
		// read next line
		textLine := scanner.Text()
		lineCount++

		// Skip comment-lines
		if strings.HasPrefix(textLine, "//") || strings.HasPrefix(textLine, "#") {
			continue
		}

		// parse line as csv
		r := csv.NewReader(strings.NewReader(textLine))
		r.Comma = ',' // set comma as the field delimiter
		fields, err := r.Read()
		if err != nil {
			log.Warnf("Unable to Parse CSV Line %d: %s", lineCount, err.Error())
			continue
		}

		// trim whitespace around all fields
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}

		// Ensure the CSV has exactly two fields
		if len(fields) != 3 {
			log.Warnf("Invalid number of fields on line %d, expected 3 buz got %d", lineCount, len(fields))
			continue
		}

		ipNet, err := IP.NewIPNetFromStr(fields[0], fields[1])
		if err != nil {
			log.Warnf("Unable to parse IP From Line %d: %s", lineCount, err.Error())
			continue
		}

		mapping := &ipMap{
			IPNet:    ipNet,
			Filename: fields[2],
		}

		// if we made it this far then store the zone
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}
