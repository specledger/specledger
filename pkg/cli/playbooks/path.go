package playbooks

import (
	"io/fs"
	"specledger/pkg/embedded"
)

var TemplatesFS = embedded.TemplatesFS

func PlaybooksDir() (string, error) {
	return "templates", nil
}

func WalkPlaybooks(fn func(path string, d fs.DirEntry, err error) error) error {
	return fs.WalkDir(TemplatesFS, ".", fn)
}

func ReadFile(path string) ([]byte, error) {
	return TemplatesFS.ReadFile(path)
}

func Stat(path string) (fs.FileInfo, error) {
	return fs.Stat(TemplatesFS, path)
}

func Exists(path string) bool {
	_, err := Stat(path)
	return err == nil
}

func AbsPath(relPath string) string {
	return relPath
}
