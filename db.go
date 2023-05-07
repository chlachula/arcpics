package arcpics

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func absRootPath(dir string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("error converting to abs path dir: '%s'", dir)
	}
	return strings.TrimSuffix(absDir, "/")
}

// Returns relative path to the root path
func relPath(root, path string) string {
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("error converting to abs path dir: '%s'", path)
	}
	path = strings.TrimSuffix(path, "/")
	lenRoot := len(root)
	if lenRoot < len(path) {
		return strings.Replace(path[lenRoot+1:], "\\", "/", -1)
	}
	if lenRoot == len(path) {
		return "./"
	}
	return strings.Replace(path, "\\", "/", -1)
}

func getParentDir(relPath string) (string, error) {
	if strings.HasPrefix(relPath, "/") {
		return "", fmt.Errorf("not relative path: '%s'", relPath)
	}
	i := strings.LastIndex(relPath, "/")
	if i <= 0 {
		return "./", nil // path abc
	}
	return relPath[:i], nil // path abc/def
}

func insertLabelSummary(db *bolt.DB, label string, countDir int, countFiles int, duration time.Duration) {
	sum := fmt.Sprintf("directories: %d, files: %d, elapsed time: %s", countDir, countFiles, duration)
	insertSystemLabelSummary(db, label, sum)
}

// Writting the directory tree json files to database
func ArcpicsFiles2DB(db *bolt.DB, arcFS ArcpicsFS) error {
	rootDir, err := filepath.Abs(arcFS.Dir)
	if err != nil {
		return err
	}
	if Verbose {
		fmt.Printf("ArcpicsFiles2DB rootDir='%s'\n", rootDir)
	}
	startTime := time.Now()
	countDir := 0
	countFiles := 0
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

			if jDir, err = makeJdir(path); err != nil {
				return err
			}
			countFiles += len(jDir.Files)
			//			if err != nil {
			//				return err
			//			}
			if Verbose {
				fmt.Printf("\rArcpicsFiles2DB: %4df %4s %s\n", len(jDir.Files), time.Since(jDirTimeStart).Truncate(time.Second), path)
			}
			changedDirs = append(changedDirs, path)
			countCreate++
			bytes, err := json.Marshal(jDir)
			if err != nil {
				return err
			}
			if bytes, err = prettyprint(bytes); err != nil {
				return err
			}

			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket(FILES_BUCKET)
				rp := relPath(rootDir, path)
				if Verbose {
					fmt.Printf("db.Add rel.path: %s\n", rp)
				}
				err := b.Put([]byte(rp), bytes)
				if err != nil {
					fmt.Printf(string(FILES_BUCKET)+" b.Put  error %s\n", err.Error())
				}
				return err
			})

		}
		return nil
	})
	//fmt.Printf("new or updated dirs: %v\n", changedDirs)
	if Verbose {
		fmt.Printf("ArcpicsFiles2DB: directories: %d, new: %d, updated: %d, elapsed time: %s\n", countDir, countCreate, countUpdate, time.Since(startTime))
	}
	insertLabelSummary(db, arcFS.Label, countDir, countFiles, time.Since(startTime))
	return nil
}

// Updating database according to the directory tree json files
func ArcpicsDatabaseUpdate(db *bolt.DB, arcFS ArcpicsFS) error {
	rootDir, err := filepath.Abs(arcFS.Dir)
	if err != nil {
		return err
	}
	if Verbose {
		fmt.Printf("ArcpicsDatabaseUpdate rootDir='%s'\n", rootDir)
	}
	startTime := time.Now()
	countDir := 0
	filepath.WalkDir(arcFS.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			println("fs.SkipDir", path)
			return fs.SkipDir
		}
		if d.IsDir() {
			countDir++
			fjson := filepath.Join(path, defaultNameJson)
			var bytes []byte
			var err error
			if bytes, err = readEntireFileToBytes(fjson); err != nil {
				fmt.Printf("Error reading file %s - %s \n", fjson, err.Error())
				return err
			}
			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket(FILES_BUCKET)
				rp := relPath(rootDir, path)
				if Verbose {
					fmt.Printf("db.Add rel.path: %s\n", rp)
				}
				err := b.Put([]byte(rp), bytes)
				if err != nil {
					fmt.Printf(string(FILES_BUCKET)+" b.Put  error %s\n", err.Error())
				}
				return err
			})
		}
		return nil
	})
	if countDir < 1 {
		fmt.Printf("ArcpicsDatabaseUpdate Warning: no %s files found at %s", defaultNameJson, arcFS.Dir)
	}
	if Verbose {
		fmt.Printf("ArcpicsDatabaseUpdate: %d files %s found, elapsed time: %s\n", countDir, defaultNameJson, time.Since(startTime))
	}
	insertLabelSummary(db, arcFS.Label, countDir, -1, time.Since(startTime))

	return nil
}

func ArcpicsAllKeys(db *bolt.DB) []string {
	keys := make([]string, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(FILES_BUCKET)

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			_ = v
			keys = append(keys, string(k))
		}

		return nil
	})
	return keys
}
func ArcpicsWordFrequency(db *bolt.DB) {
	counter := map[string]int{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FILES_BUCKET)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			var jDir JdirType
			err := json.Unmarshal(v, &jDir)
			if err == nil {
				s := strings.Replace(jDir.MostComment, "|", " ", -1)
				s = strings.Replace(s, ",", " ", -1)
				s = strings.Replace(s, ";", " ", -1)
				s = strings.Replace(s, ":", " ", -1)
				s = strings.Replace(s, "-", " ", -1)
				words := strings.Split(s, " ")
				for _, word := range words {
					counter[word]++
				}
			}
		}

		return nil
	})
	fmt.Printf("MostComment word frequency: %v\n", counter)
}

func ArcpicsQuery(db *bolt.DB, query string) {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(FILES_BUCKET)

		c := b.Cursor()
		counter := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			str := string(v)
			if strings.Contains(str, query) {
				counter++
				fmt.Printf("ArcpicsQuery %2d. %s\n", counter, k)
			}
		}

		return nil
	})
}

/* getValue
func ArcpicsValue(db *bolt.DB, path string) string {
	str := ""
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(FILES_BUCKET)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			if strings.Contains(string(k), path) {
				str = string(v)
				break
			}
		}
		return nil
	})
	return str
}
*/
