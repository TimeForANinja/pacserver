package storage

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2/log"
	"github.com/timeforaninja/pacserver/pkg/utils"
)

// ReadPACTemplates loads all PAC templates from a directory.
func ReadPACTemplates(relPacDir string) ([]*PACTemplate, error, int) {
	// Resolve the directory first so we can report a single path in all downstream errors.
	absPACPath, err := filepath.Abs(relPacDir)
	if err != nil {
		log.Errorf("Invalid Filepath for PACs found: %q: %s", relPacDir, err.Error())
		return []*PACTemplate{}, err, 1
	}

	// Let the helper enumerate only actual files; the storage layer should stay file-focused.
	files, err := utils.ListFiles(absPACPath)
	if err != nil {
		log.Errorf("Failed to List PAC Files in %q: %s", absPACPath, err.Error())
		return []*PACTemplate{}, err, 1
	}

	templates := make([]*PACTemplate, 0, len(files))
	problemCounter := 0
	for _, file := range files {
		// Load each template independently so one bad PAC does not stop the rest.
		template, err := readPACTemplate(filepath.Join(absPACPath, file), file)
		if err != nil {
			log.Warnf("Unable to read PAC at %q: %s", file, err.Error())
			problemCounter++
			continue
		}
		templates = append(templates, template)
	}

	return templates, nil, problemCounter
}

func readPACTemplate(fullPath, filename string) (*PACTemplate, error) {
	// PAC files are served as plain text, so a raw read is enough here.
	fileBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	return &PACTemplate{
		Filename: utils.NormalizePath(filename),
		Content:  string(fileBytes),
	}, nil
}
