package internal

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

func readTemplateFiles(rel_pac_dir string) ([]*pacTemplate, error) {
	abs_pac_path, err := filepath.Abs(rel_pac_dir)
	if err != nil {
		log.Errorf("Invalid Filepath for PACs found: \"%s\" %s", err.Error())
		return nil, err
	}
	files, err := utils.ListFiles(abs_pac_path)
	if err != nil {
		log.Errorf("Failed to List PAC Files in \"%s\": %s", abs_pac_path, err.Error())
		return nil, err
	}

	var templates []*pacTemplate

	for _, file := range files {
		fullpath := filepath.Join(abs_pac_path, file)
		fileBytes, err := os.ReadFile(fullpath)
		if err != nil {
			log.Warnf("Unable to read PAC at \"%s\": %s", fullpath, err.Error())
			continue
		}

		templates = append(templates, &pacTemplate{
			Filename: file,
			content:  string(fileBytes),
		})
	}

	return templates, nil
}
