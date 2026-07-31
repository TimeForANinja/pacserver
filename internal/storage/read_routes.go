package storage

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/timeforaninja/pacserver/pkg/utils"
)

// ReadRoutes loads exact URL-path to PAC mappings from a CSV file.
func ReadRoutes(filename string) ([]*RouteMap, error, int) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err, 1
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	var result []*RouteMap
	problems := 0
	for {
		row, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return result, err, problems + 1
		}
		if len(row) == 0 || strings.TrimSpace(row[0]) == "" || strings.HasPrefix(strings.TrimSpace(row[0]), "#") || strings.HasPrefix(strings.TrimSpace(row[0]), "//") {
			continue
		}
		if len(row) < 2 || len(row) > 3 {
			problems++
			continue
		}
		route := "/" + strings.Trim(strings.TrimSpace(row[0]), "/")
		if route == "/" {
			problems++
			continue
		}
		m := &RouteMap{Route: route, Filename: utils.NormalizePath(strings.TrimSpace(row[1]))}
		if len(row) == 3 {
			m.Comment = strings.TrimSpace(row[2])
		}
		if m.Filename == "" {
			problems++
			continue
		}
		result = append(result, m)
	}
	return result, nil, problems
}
