package arcpics

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func CrawlDir() {

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
			onlyFiles = append(onlyFiles, file)
		}
	}

	for _, f := range onlyFiles {
		info, _ := f.Info()
		fmt.Printf("    fs.DirEntry: %v %v\n", f.Name(), info.ModTime())
	}
	fmt.Printf("----\n")

	return onlyFiles, nil
}

func FilesInDirXXX(d string) (JdirType, error) {
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
		file.Comment = "my comment, realy"
		jd.Files = append(jd.Files, file)
	}
	return jd, nil
}

func CreateDirJson(d string, dirFiles JdirType) error {
	_, _ = filesInDir(d)
	jsonBytes, err := json.Marshal(dirFiles)
	if err != nil {
		return err
	}
	fname := filepath.Join(d, jsonFilePrefix+"-new")
	f, err := os.Create(fname)
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
func UpdateDirJson(d string) error {
	fjson_new := filepath.Join(d, jsonFilePrefix+"-new")
	if _, err := os.Stat(fjson_new); err != nil {
		return nil // there is no new file
	}
	fjson := filepath.Join(d, jsonFilePrefix)
	//fjson := filepath.Join(d, jsonlePrefix)
	var fjson_FileInfo os.FileInfo
	var err error
	if fjson_FileInfo, err = os.Stat(fjson); err != nil {
		return nil // there is now current file
	}
	timeExtension := fjson_FileInfo.ModTime().Format("--2006-01-_15-04-05") // Mon Jan 2 15:04:05 -0700 MST 2006
	fjson_time := filepath.Join(d, jsonFilePrefix+timeExtension)
	if err := os.Rename(fjson, fjson_time); err != nil {
		return err
	}
	if err := os.Rename(fjson_new, fjson); err != nil {
		return err
	}
	return nil
}
