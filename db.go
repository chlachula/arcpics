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

// Updating database according to the directory tree json files
func ArcpicsDatabaseUpdate(db *bolt.DB, dir string) error {
	rootDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	fmt.Printf("ArcpicsDatabaseUpdate rootDir='%s'\n", rootDir)
	startTime := time.Now()
	countDir := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
				return err
			}

			db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket(FILES_BUCKET)
				err := b.Put([]byte(relPath(dir, path)), bytes)
				if err != nil {
					fmt.Printf("TEST, b.Put  error %s\n", err.Error())
				}
				return err
			})
		}
		return nil
	})
	fmt.Printf("ArcpicsDatabaseUpdate: elapsed time: %s\n", time.Since(startTime))
	return nil
}

func ArcpicsAllKeys(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(FILES_BUCKET)

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			//fmt.Printf("key=%s, value=%s\n", k, v)
			_ = v
			fmt.Printf("ArcpicsAllKeys key=%s=\n", k)
		}

		return nil
	})
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
