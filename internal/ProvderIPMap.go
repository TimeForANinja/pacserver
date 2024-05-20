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
	IPNet     IP.IPNet `json:"IPNet"`
	Filename  string   `json:"Filename"`
	Hostnames []string `json:"Hostnames"`
}

func readIPMap(rel_path string) ([]*ipMap, error) {
	abs_path, err := filepath.Abs(rel_path)
	if err != nil {
		log.Errorf("Invalid Filepath for IPMap found: \"%s\" %s", err.Error())
		return nil, err
	}
	file, err := os.Open(abs_path)
	if err != nil {
		log.Errorf("Unable to open IPMap at \"%s\": %s", abs_path, err.Error())
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var mappings []*ipMap

	lineCount := -1
	for scanner.Scan() {
		// read next line
		txtline := scanner.Text()
		lineCount++
		// Skip comment-lines
		if strings.HasPrefix(txtline, "//") || strings.HasPrefix(txtline, "#") {
			continue
		}
		// parse line as csv
		r := csv.NewReader(strings.NewReader(txtline))
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

		ipnet, err := IP.NewIPNetFromStr(fields[0], fields[1])
		if err != nil {
			log.Warnf("Unable to parse IP From Line %d: %s", lineCount, err.Error())
			continue
		}

		// TODO: check syntax of other fields
		mapping := &ipMap{
			IPNet:     ipnet,
			Filename:  fields[2],
			Hostnames: make([]string, 0),
		}

		// append all hostnames
		mapping.Hostnames = append(mapping.Hostnames, fields[3:]...)
		if len(mapping.Hostnames) == 0 {
			log.Warnf("Zone on Line %d did not provide any Proxy Hosts", lineCount)
			continue
		}

		// if we made it this far then store the zone
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}
