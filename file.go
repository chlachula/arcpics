package arcpics

import (
	"encoding/json"
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
	Dir   string
	Label string
}

func ArcpicsFS(dir string) (arcpicsFS, error) {
	var a arcpicsFS
	a.Dir = dir
	label, err := getLabel(dir)
	if err != nil {
		return a, err
	}
	a.Label = label
	return a, nil
}

func (afs arcpicsFS) Open(name string) (fs.File, error) {
	f, err := os.Open(filepath.Join(afs.Dir, name))
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
func DirFilesCount(fsys fs.FS) (int, int) {
	countDir := 0
	countFiles := 0
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			countDir++
			fmt.Printf("dir  #%3d - %s\n", countDir, path)
		} else {
			countFiles++
			fmt.Printf("    file #%3d - %s\n", countFiles, path)

		}
		return nil
	})
	return countDir, countFiles
}
func DirCount(fsys fs.FS) (countDir int) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			countDir++
			fmt.Printf("DirCount:  #%3d - %s\n", countDir, path)
		}
		return nil
	})
	return countDir
}

// func ArcpicsFilesUpdate(fsys fs.FS) error {
func ArcpicsFilesUpdate(dir string) error {
	countDir := 0
	//countFiles := 0
	//dir := ""
	//entries := make([]fs.DirEntry, 0)
	//fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			countDir++
			var jDir JdirType
			//absPath, _ := filepath.Abs(path)
			fjson := filepath.Join(path, jsonFilePrefix)
			fmt.Printf("ArcpicsFilesUpdate: %s \n", fjson)
			//_ = jDir

			if jDir, err = makeJdir(path); err != nil {
				return err
			}

			if !fileExists(fjson) {
				if err = CreateDirJson(fjson, jDir); err != nil {
					return err
				}
				if err = CreateDirJson(fjson, jDir); err != nil {
					return err
				}
			} else {
				if err = UpdateDirJson(fjson, jDir); err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}

		}
		return nil
	})

	fmt.Printf("ArcpicsFilesUpdate: %d directories\n", countDir)
	return nil
}
func makeJdir(d string) (JdirType, error) {
	var jd JdirType
	jd.Files = make([]JfileType, 0)
	jd.Description = "a description..."
	jd.Location = "a secret location..."
	var files []fs.DirEntry
	var err error
	if files, err = filesInDir(d); err != nil {
		return jd, err
	}

	for _, f := range files {
		info, _ := f.Info()
		var file JfileType
		file.Name = info.Name()
		file.Time = info.ModTime().Format("2006-01-02_15:04:05")
		file.Comment = "my own comment, OK?"
		jd.Files = append(jd.Files, file)
	}
	return jd, nil
}

func filesInDir(d string) ([]fs.DirEntry, error) {
	var files []fs.DirEntry
	var err error
	if files, err = os.ReadDir(d); err != nil {
		return nil, err
	}
	onlyFiles := make([]fs.DirEntry, 0)
	for _, file := range files {
		if !file.IsDir() {
			if !strings.HasPrefix(file.Name(), jsonFilePrefix) {
				onlyFiles = append(onlyFiles, file)
			}
		}
	}

	//for _, f := range onlyFiles {
	//	info, _ := f.Info()
	//	fmt.Printf("    fs.DirEntry: %v %v\n", f.Name(), info.ModTime())
	//}
	//fmt.Printf("----\n")

	return onlyFiles, nil
}

func CreateDirJson(jfname string, jDir JdirType) error {
	jsonBytes, err := json.Marshal(jDir)
	if err != nil {
		return err
	}
	f, err := os.Create(jfname)
	if err != nil {
		return err
	}
	_, err = f.Write(jsonBytes) //lenght ommited
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
func UpdateDirJson(fjson string, jDir JdirType) error {
	var fjson_FileInfo os.FileInfo
	var err error
	if fjson_FileInfo, err = os.Stat(fjson); err != nil {
		return nil // there is now current file
	}
	timeExtension := fjson_FileInfo.ModTime().Format("--2006-01-_15-04-05") // Mon Jan 2 15:04:05 -0700 MST 2006
	fjson_time := fjson + timeExtension
	if err := os.Rename(fjson, fjson_time); err != nil {
		return err
	}
	if err = CreateDirJson(fjson, jDir); err != nil {
		return err
	}
	return nil
}
