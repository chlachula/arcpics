package arcpics

import (
	"bytes"
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
// ATTENTION!!
// ArcpicsFS work fine with fs.WalkDir unless there are any file operations
// Then use filepath.WalkDir(arcpicsFS.Dir,...

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
func jDirIsEqual(a, b JdirType) bool {
	if a.Description != b.Description {
		return false
	}
	if a.Location != b.Location {
		return false
	}
	if len(a.Files) != len(b.Files) {
		return false
	}
	for i, af := range a.Files {
		if af.Comment != b.Files[i].Comment {
			return false
		}
		if af.Name != b.Files[i].Name {
			return false
		}
		if af.Size != b.Files[i].Size {
			return false
		}
		if af.Time != b.Files[i].Time {
			fmt.Printf("jDirIsEqual %d. file a.time=%s, b.time=%s", i, af.Time, b.Files[i].Time)
			return false
		}
	}
	return true
}
func ArcpicsFilesUpdate(dir string) error {
	countDir := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			countDir++
			var jDir JdirType
			fjson := filepath.Join(path, jsonFilePrefix)

			if jDir, err = makeJdir(path); err != nil {
				return err
			}

			if !fileExists(fjson) {
				if err = CreateDirJson(fjson, jDir); err != nil {
					return err
				} else {
					fmt.Printf("Arcpics - created: %s \n", fjson)
				}
			} else {
				currentJDir, err := readJsonDirData(fjson)
				if err != nil {
					fmt.Printf("error reading file %s\n %s\n", fjson, err.Error())
				}
				if !jDirIsEqual(currentJDir, jDir) {
					if err = UpdateDirJson(fjson, jDir); err != nil {
						return err
					} else {
						fmt.Printf("Arcpics - updated: %s \n", fjson)
					}
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
func PurgeJson__(dir string) error {
	doExt_toBeRemoved := ".json--"
	count := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			if strings.HasPrefix(ext, doExt_toBeRemoved) {
				if err := os.Remove(path); err == nil {
					fmt.Printf("Removed: %s\n", path)
					count++
				} else {
					fmt.Printf("error removing file %s\n %s\n", path, err.Error())
				}
			}
		}
		return nil
	})
	fmt.Printf("Dir root: %s\nPurge count: %d\n", dir, count)
	return nil
}
func readJsonDirData(fname string) (JdirType, error) {
	var userData JdirType
	fileBytes, _ := os.ReadFile(fname)
	err := json.Unmarshal(fileBytes, &userData)
	return userData, err
}
func makeJdir(d string) (JdirType, error) {
	var jd JdirType
	jd.Files = make([]JfileType, 0)
	jd.Description = "here could be a description from file " + jsonUserData
	jd.Location = "here could be a description from file " + jsonUserData
	userFile := filepath.Join(d, jsonUserData)
	if fileExists(userFile) {
		userData, err := readJsonDirData(userFile)
		if err == nil {
			jd.Description = userData.Description
			jd.Location = userData.Location
		} else {
			fmt.Printf("error in the file %s\n %s\n", userFile, err.Error())
		}
	}

	var files []fs.DirEntry
	var err error
	if files, err = filesInDir(d); err != nil {
		return jd, err
	}

	for _, f := range files {
		info, _ := f.Info()
		var file JfileType
		file.Name = info.Name()
		file.Size = fmt.Sprintf("%d", info.Size())
		file.Time = info.ModTime().Format(timeStampJsonFormat)
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
			a := !strings.HasPrefix(file.Name(), jsonFilePrefix)
			b := !strings.HasPrefix(file.Name(), jsonUserData)
			if a && b {
				onlyFiles = append(onlyFiles, file)
			}
		}
	}
	return onlyFiles, nil
}

func prettyprint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}
func CreateDirJson(jfname string, jDir JdirType) error {
	jsonBytes, err := json.Marshal(jDir)
	if err != nil {
		return err
	}
	if jsonBytes, err = prettyprint(jsonBytes); err != nil {
		return err
	}
	if err := os.WriteFile(jfname, jsonBytes, 0666); err != nil {
		return err
	}
	return nil
}
func UpdateDirJson(fjson string, jDir JdirType) error {
	var fjson_FileInfo os.FileInfo
	var err error
	if fjson_FileInfo, err = os.Stat(fjson); err != nil {
		return nil // there is now current file
	}
	timeExtension := fjson_FileInfo.ModTime().Format("--2006-01-02_15-04-05") // Mon Jan 2 15:04:05 -0700 MST 2006
	fjson_time := fjson + timeExtension
	if err := os.Rename(fjson, fjson_time); err != nil {
		return err
	}
	if err = CreateDirJson(fjson, jDir); err != nil {
		return err
	}
	return nil
}
