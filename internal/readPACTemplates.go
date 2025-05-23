package internal

/**
 * this file reads in the PAC files
 * convert the templates to an actual PAC is done when creating the LookupElement
 */

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

type pacTemplate struct {
	Filename string `json:"Filename"`
	content  string
}

func readTemplateFiles(relPacDir string) ([]*pacTemplate, error, int) {
	absPACPath, err := filepath.Abs(relPacDir)
	if err != nil {
		log.Errorf("Invalid Filepath for PACs found: \"%s\": %s", absPACPath, err.Error())
		return make([]*pacTemplate, 0), err, 1
	}
	files, err := utils.ListFiles(absPACPath)
	if err != nil {
		log.Errorf("Failed to List PAC Files in \"%s\": %s", absPACPath, err.Error())
		return make([]*pacTemplate, 0), err, 1
	}

	var templates []*pacTemplate
	problemCounter := 0

	for _, file := range files {
		template, err := readAndParse(absPACPath, file)
		if err != nil {
			log.Warnf("Unable to read PAC at \"%s\": %s", file, err.Error())
			problemCounter++
			continue
		}
		templates = append(templates, template)
	}

	return templates, nil, problemCounter
}

func readAndParse(basePath, file string) (*pacTemplate, error) {
	fullPath := filepath.Join(basePath, file)
	fileBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return &pacTemplate{
		Filename: utils.NormalizePath(file),
		content:  string(fileBytes),
	}, nil
}
