package slnx

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func WriteFile(filePath string, solution *Solution) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil { // #nosec G301 -- standard directory permissions
		return fmt.Errorf("create output directory: %w", err)
	}

	content, err := Marshal(solution)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filePath, content, 0o644); err != nil { // #nosec G306 -- output file permissions
		return fmt.Errorf("write .slnx file: %w", err)
	}
	return nil
}

func Marshal(solution *Solution) ([]byte, error) {
	if solution == nil {
		return nil, fmt.Errorf("solution cannot be nil")
	}

	canonicalize(solution)

	xmlValue := xmlSolution{Folders: make([]xmlFolder, 0, len(solution.Folders))}
	for _, folder := range solution.Folders {
		xmlFolderValue := xmlFolder{
			Name:     normalizeSeparators(folder.Name),
			Files:    make([]xmlFile, 0, len(folder.Files)),
			Projects: make([]xmlProject, 0, len(folder.Projects)),
		}

		for _, file := range folder.Files {
			xmlFolderValue.Files = append(xmlFolderValue.Files, xmlFile{Path: normalizeSeparators(file.Path)})
		}
		for _, project := range folder.Projects {
			xmlFolderValue.Projects = append(xmlFolderValue.Projects, xmlProject{
				Path:     normalizeSeparators(project.Path),
				ID:       project.ID,
				InnerXML: project.InnerXML,
			})
		}

		xmlValue.Folders = append(xmlValue.Folders, xmlFolderValue)
	}

	content, err := xml.MarshalIndent(xmlValue, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal .slnx xml: %w", err)
	}
	content = append(content, '\n')
	return content, nil
}
