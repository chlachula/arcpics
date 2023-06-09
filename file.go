package arcpics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chlachula/arcpics/jpeg"
)

func OpenArcpicsFS(dir string) (ArcpicsFS, error) {
	var a ArcpicsFS
	a.Dir = dir
	label, err := getLabel(dir)
	if err != nil {
		return a, err
	}
	a.Label = label
	return a, nil
}

func (afs ArcpicsFS) Open(name string) (fs.File, error) {
	f, err := os.Open(filepath.Join(afs.Dir, name))
	if err != nil {
		return f, err
	}
	return f, nil
}
func (afs ArcpicsFS) DirPaths() ([]string, error) {
	paths := make([]string, 0)
	fs.WalkDir(afs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			paths = append(paths, path)
			if Verbose {
				fmt.Printf("DirPaths-: %3db  %s\n", len(path), path)
			}
		}
		return nil
	})
	return paths, nil
}

func (afs ArcpicsFS) DirPathsUpdate() error {
	paths, err := afs.DirPaths()
	if err != nil {
		return err
	}
	for i, path := range paths {
		dir := filepath.Join(afs.Dir, path)
		if Verbose {
			fmt.Printf("%2d. path: %s\n", i, dir)
		}
	}
	return nil
}
func getLabel(archiveDir string) (string, error) {
	nameStart := defaultNameDashLabelDot
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

// find all label files with prefix defaultNameDashLabelDot in root subdirectories
func getPathLabels(root string) []string {
	s := make([]string, 0)
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() {
			if strings.HasPrefix(d.Name(), defaultNameDashLabelDot) {
				s = append(s, path)
			}
		}
		return nil
	})

	return s
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
			if Verbose {
				fmt.Printf("dir  #%3d - %s\n", countDir, path)
			}
		} else {
			countFiles++
			if Verbose {
				fmt.Printf("    file #%3d - %s\n", countFiles, path)
			}
		}
		return nil
	})
	return countDir, countFiles
}
func DirCount(fsys fs.FS) (countDir int, totalPathLength int) {
	total := 0
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			countDir++
			total += len(path)
			if Verbose {
				fmt.Printf("DirCount:  #%3d %3d - %s\n", countDir, len(path), path)
			}
		}
		return nil
	})
	return countDir, total
}
func jInfoIsEqual(a, b JinfoType) bool {
	if a.Author != b.Author {
		return false
	}
	if a.Location != b.Location {
		return false
	}
	if a.Location != b.Location {
		return false
	}
	if a.Keywords != b.Keywords {
		return false
	}
	if a.Comment != b.Comment {
		return false
	}
	return true
}
func jDirIsEqual(a, b JdirType) bool {
	if !jInfoIsEqual(a.Info, b.Info) {
		return false
	}
	if !jInfoIsEqual(a.Info, b.Info) {
		return false
	}
	if len(a.Files) != len(b.Files) {
		return false
	}
	for i, af := range a.Files {
		if !jInfoIsEqual(af.Info, b.Files[i].Info) {
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
func byUserData_thisDirShouldBeSkipped(dir string, name string) bool {
	userData, err := readJsonDirData(filepath.Join(dir, defaultNameDashUserDataJson))
	if err != nil {
		return false //file not found or not readable
	}
	if skipFile(userData.Skip, name) {
		return true
	}
	return false
}

// Updating directory tree json files according to dir content
func ArcpicsFilesUpdate(arcFS ArcpicsFS) error {
	startTime := time.Now()
	countDir := 0
	countCreate := 0
	countUpdate := 0
	changedDirs := make([]string, 0)
	filepath.WalkDir(arcFS.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println(err.Error(), "fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			parentDir := filepath.Dir(path)
			if byUserData_thisDirShouldBeSkipped(parentDir, d.Name()) {
				return fs.SkipDir
			}

			countDir++
			var jDir JdirType
			jDirTimeStart := time.Now()
			fjson := filepath.Join(path, defaultNameJson)

			if jDir, err = makeJdir(path); err != nil {
				return err
			}

			if !FileExists(fjson) {
				if err = CreateDirJson(fjson, jDir); err != nil {
					return err
				} else {
					if Verbose {
						fmt.Printf("Arcpics - created: %4df %4s %s\n", len(jDir.Files), time.Since(jDirTimeStart).Truncate(time.Second), fjson)
					}
					changedDirs = append(changedDirs, path)
					countCreate++
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
						fmt.Printf("Arcpics - updated: %4df %4s %s\n", len(jDir.Files), time.Since(jDirTimeStart).Truncate(time.Second), fjson)
						changedDirs = append(changedDirs, path)
						countUpdate++
					}
				}
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	//fmt.Printf("new or updated dirs: %v\n", changedDirs)
	if Verbose {
		fmt.Printf("ArcpicsFilesUpdate: directories: %d, new: %d, updated: %d, elapsed time: %s\n", countDir, countCreate, countUpdate, time.Since(startTime))
	}
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
	if Verbose {
		fmt.Printf("Dir root: %s\nPurge count: %d\n", dir, count)
	}
	return nil
}
func readJsonDirData(fname string) (JdirType, error) {
	var userData JdirType
	fileBytes, _ := os.ReadFile(fname)
	err := json.Unmarshal(fileBytes, &userData)
	return userData, err
}
func updateByJpegValues(file *JfileType, fname string) {
	var j jpeg.JpegReader
	j.Open(fname, false) // verbose=false
	j.Decode()
	s := strings.Split(j.Comment, "|")
	file.Info.Comment = j.Comment //v1|UTF-8|(c) Josef Chlachula|Rochester MN|home,astro|Northern lights
	if len(s) > 2 {
		file.Info.Author = s[2]
	}
	if len(s) > 3 {
		file.Info.Location = s[3]
	}
	if len(s) > 4 {
		file.Info.Keywords = s[4]
	}
	if len(s) > 5 {
		file.Info.Comment = s[5]
	}
	file.Thumbnail = j.Thumbnail
	file.ThumbSrc = j.ThumbSrc
}
func skipFile(skipFiles []string, fileName string) bool {
	for _, skipFile := range skipFiles {
		if skipFile == fileName {
			return true
		}
	}
	return false
}

/*
	func skipDirByUser(dir string, skipFiles []string) error {
		f, _ := os.Open(dir)
		name := f.Name()
		fmt.Printf("TEST skipDirByUser dir=%s name=%s\n", dir, name)
		if skipFile(skipFiles, name) {
			return fs.SkipDir
			//fmt.Errorf(ErrSkippedByUser)
		}
		return nil
	}
*/
func isJpegFile(fileName string) bool {
	n := strings.ToLower(fileName)
	return strings.HasSuffix(n, "jpg") || strings.HasSuffix(n, "jpeg")
}
func mostOccurringString(counter map[string]int) (string, string) {
	vMax := 0
	kMax := ""
	sum := 0
	for k, v := range counter {
		sum += v
		if v > vMax {
			vMax = v
			kMax = k
		}
	}
	return kMax, fmt.Sprintf("%d/%d", vMax, sum)
}

func makeJdir(dir string) (JdirType, error) {
	var jd JdirType
	var userData JdirType
	var err error
	jd.Files = make([]JfileType, 0)

	var files []fs.DirEntry
	if files, err = filesInDir(dir); err != nil {
		return jd, err
	}
	// find most occuring author, location, keywords, comment
	authorCounter := map[string]int{}
	locationCounter := map[string]int{}
	keywordsCounter := map[string]int{}
	commentCounter := map[string]int{}
	for _, f := range files {
		info, _ := f.Info()
		var file JfileType
		file.Name = info.Name()
		if !skipFile(userData.Skip, file.Name) {
			file.Size = fmt.Sprintf("%d", info.Size())
			file.Time = info.ModTime().Format(timeStampJsonFormat)
			if isJpegFile(file.Name) {
				updateByJpegValues(&file, filepath.Join(dir, file.Name))
			}
			authorCounter[file.Info.Author]++
			locationCounter[file.Info.Location]++
			keywordsCounter[file.Info.Keywords]++
			commentCounter[file.Info.Comment]++
			jd.Files = append(jd.Files, file)
		}
	}

	var subdirs []string
	if subdirs, err = subdirsInDir(dir); err != nil {
		return jd, err
	}
	for _, d := range subdirs {
		if !skipFile(userData.Skip, d) {
			jd.Dirs = append(jd.Dirs, d)
		}
	}

	userFile := filepath.Join(dir, defaultNameDashUserDataJson)
	if FileExists(userFile) {
		userData, err = readJsonDirData(userFile)
		if err == nil {
			jd.ByUser = true
			jd.Info.Author = userData.Info.Author
			jd.Info.Location = userData.Info.Location
			jd.Info.Keywords = userData.Info.Keywords
			jd.Info.Comment = userData.Info.Comment
		} else {
			fmt.Printf("error in the file %s\n %s\n", userFile, err.Error())
		}
	} else {
		jd.Info.Author, _ = mostOccurringString(authorCounter)
		jd.Info.Location, _ = mostOccurringString(locationCounter)
		jd.Info.Keywords, _ = mostOccurringString(keywordsCounter)
		jd.Info.Comment, _ = mostOccurringString(commentCounter)
	}
	if jd.UpdateTime == "" {
		jd.UpdateTime = time.Now().Format(time.DateTime)
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
			a := !strings.HasPrefix(file.Name(), defaultNameJson)
			b := !strings.HasPrefix(file.Name(), defaultNameDashUserDataJson)
			if a && b {
				onlyFiles = append(onlyFiles, file)
			}
		}
	}
	return onlyFiles, nil
}
func subdirsInDir(d string) ([]string, error) {
	var files []fs.DirEntry
	var err error
	if files, err = os.ReadDir(d); err != nil {
		return nil, err
	}
	dirs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}
func GetLabelsInDbDir(d string) ([]string, error) {
	var files []fs.DirEntry
	var err error
	if files, err = os.ReadDir(d); err != nil {
		return nil, err
	}
	labels := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			a := strings.HasPrefix(file.Name(), defaultNameDash)
			b := filepath.Ext(file.Name()) == ".db"
			if a && b {
				label := file.Name()
				label = label[len(defaultNameDash):]
				label = label[:len(label)-3]
				labels = append(labels, label)
			}
		}
	}
	return labels, nil
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

// CopyDirFromTo copies the content of Src to Dst.
func CopyDirFromTo(fromSrc, toDst string) error {

	return filepath.Walk(fromSrc, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(toDst, strings.TrimPrefix(path, fromSrc))

		if info.IsDir() {
			os.MkdirAll(outpath, info.Mode())
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, _ := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer fh.Close()

		// make it the same
		fh.Chmod(info.Mode())

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}

func readEntireFileToBytes(fname string) ([]byte, error) {
	var bytes []byte
	reader, err := os.Open(fname)
	if err != nil {
		return bytes, err
	}
	bytes, err = io.ReadAll(reader)
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}
func copyToStash(dir, file string) {
	dbDir := GetDatabaseDirName()
	stashDir := filepath.Join(dbDir, stashName)
	_ = os.Mkdir(stashDir, 0755)
	copy(filepath.Join(dir, file), filepath.Join(stashDir, file))
}
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
func CmdHelp(msg string) string {
	s := ""
	if msg != "" {
		s = s + msg + "\n"
	}
	r := "/media/joe/USB32/"
	h := "~/.arcpics/"
	if runtime.GOOS == "windows" {
		r = "E:\\"
		h = "C:\\Users\\joe\\.arcpics\\"
	}
	s = s + fmt.Sprintf(HelpTextFmt, Version, r, r, r, h, r, r, r, r, h, r)
	return s
}
func extractDir(path string) (string, string) {
	i := strings.Index(path, "/")
	return path[:i], path[i+1:]
}
func makeNodes(arr []string) []Node {
	nodes := make([]Node, 0)
	prevdir := ""
	arr2 := make([]string, 0)
	for _, f := range arr {
		if !strings.Contains(f, "/") {
			var node Node
			node.Name = f
			nodes = append(nodes, node)
		} else {
			dir, f2 := extractDir(f)
			if prevdir != dir {
				if prevdir != "" {
					var node Node
					node.Name = prevdir
					node.Nodes = makeNodes(arr2)
					nodes = append(nodes, node)
					arr2 = make([]string, 0)
				}
			}
			arr2 = append(arr2, f2)
			prevdir = dir
		}
	}
	if len(arr2) > 0 {
		var node Node
		node.Name = prevdir
		node.Nodes = makeNodes(arr2)
		nodes = append(nodes, node)
	}
	return nodes
}
