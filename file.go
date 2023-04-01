package arcpics

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// File system ArcpicsFS has to have at root special label file with name "arcpics-db-label"
// and at least one character long arbitrary extension.
// For example file "arcpics-db-label.a" has label value "a"
// or "arcpics-db-label.my1TB_hard_drive" has label value "my1TB_hard_drive"
//

var defaultArcpicsDbLabel = "arcpics-db-label."

type arcpicsFS struct {
	dir   string
	label string
	//fs    fs.FS
}

func ArcpicsFS(dir string) (arcpicsFS, error) {
	var a arcpicsFS
	a.dir = dir
	label, err := getLabel(dir)
	if err != nil {
		return a, err
	}
	a.label = label
	return a, nil
}

func (afs arcpicsFS) Open(name string) (fs.File, error) {
	f, err := os.Open(filepath.Join(afs.dir, name))
	if err != nil {
		return f, err
	}
	return f, nil
}
func getLabel(archiveDir string) (string, error) {
	nameStart := defaultArcpicsDbLabel
	label := ""
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), nameStart) {
			count++
			label = f.Name()[len(nameStart):]
		}
	}
	if count == 0 {
		return label, fmt.Errorf("there is no file %s* at directory %s - should be e.g. %s001", nameStart, archiveDir, nameStart)
	} else if count > 1 {
		return label, fmt.Errorf("unexpected number files %s* at directory %s", nameStart, archiveDir)
	}
	if len(label) < 1 {
		return label, fmt.Errorf("there is not at least one character label after dot 'arcpics-db-label.'")
	}
	return label, nil
}

func FilesCount(fsys fs.FS) (count int) {
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		count++
		return nil
	})
	return count
}
