package slnx

import "strings"

type Solution struct {
	Folders []Folder
}

type Folder struct {
	Name     string
	Files    []File
	Projects []Project
}

type File struct {
	Path string
}

type Project struct {
	Path     string
	ID       string
	InnerXML string
}

func normalizeSeparators(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}
