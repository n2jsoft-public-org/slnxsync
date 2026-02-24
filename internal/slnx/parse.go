package slnx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
)

type xmlSolution struct {
	XMLName xml.Name    `xml:"Solution"`
	Folders []xmlFolder `xml:"Folder"`
}

type xmlFolder struct {
	Name     string       `xml:"Name,attr"`
	Files    []xmlFile    `xml:"File"`
	Projects []xmlProject `xml:"Project"`
}

type xmlFile struct {
	Path string `xml:"Path,attr"`
}

type xmlProject struct {
	Path     string `xml:"Path,attr"`
	ID       string `xml:"Id,attr,omitempty"`
	InnerXML string `xml:",innerxml"`
}

func ParseFile(filePath string) (*Solution, error) {
	content, err := os.ReadFile(filePath) // #nosec G304 -- filePath from config
	if err != nil {
		return nil, fmt.Errorf("read .slnx file: %w", err)
	}
	return Parse(content)
}

func Parse(content []byte) (*Solution, error) {
	var parsed xmlSolution
	decoder := xml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode .slnx xml: %w", err)
	}

	solution := &Solution{Folders: make([]Folder, 0, len(parsed.Folders))}
	for _, xmlFolderValue := range parsed.Folders {
		folder := Folder{
			Name:     normalizeSeparators(xmlFolderValue.Name),
			Files:    make([]File, 0, len(xmlFolderValue.Files)),
			Projects: make([]Project, 0, len(xmlFolderValue.Projects)),
		}

		for _, xmlFileValue := range xmlFolderValue.Files {
			folder.Files = append(folder.Files, File{Path: normalizeSeparators(xmlFileValue.Path)})
		}
		for _, xmlProjectValue := range xmlFolderValue.Projects {
			folder.Projects = append(folder.Projects, Project{
				Path:     normalizeSeparators(xmlProjectValue.Path),
				ID:       strings.TrimSpace(xmlProjectValue.ID),
				InnerXML: strings.TrimSpace(xmlProjectValue.InnerXML),
			})
		}

		solution.Folders = append(solution.Folders, folder)
	}

	canonicalize(solution)
	return solution, nil
}

func canonicalize(solution *Solution) {
	for folderIdx := range solution.Folders {
		sort.SliceStable(solution.Folders[folderIdx].Files, func(i, j int) bool {
			return solution.Folders[folderIdx].Files[i].Path < solution.Folders[folderIdx].Files[j].Path
		})
		sort.SliceStable(solution.Folders[folderIdx].Projects, func(i, j int) bool {
			return solution.Folders[folderIdx].Projects[i].Path < solution.Folders[folderIdx].Projects[j].Path
		})
	}

	sort.SliceStable(solution.Folders, func(i, j int) bool {
		return solution.Folders[i].Name < solution.Folders[j].Name
	})
}
