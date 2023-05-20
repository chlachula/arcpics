package arcpics

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
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

// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func insertLabelSummary(db *bolt.DB, label string, countDir int, countFiles int, countJpegs int, totalSize int, duration time.Duration) {
	total := ByteCountIEC(int64(totalSize))
	FORMAT := "directories:%5d, files:%7d, jpegs:%6d, total size:%10s, elapsed time: %-15s on %s"
	sum := fmt.Sprintf(FORMAT, countDir, countFiles, countJpegs, total, duration, time.Now().Format("2006-01-02_15:04"))
	key := fmt.Sprintf(LABEL_SUMMARY_fmt, label)
	insert2System_KeyValueStrings(db, key, sum)
}
func insertFrequencyWords(db *bolt.DB, label string, m map[string]FrequencyCounterType) {
	bytes, err := json.Marshal(m)
	if err == nil {
		insert2System_KeyValue(db, []byte(LABEL_FREQUENCY_WORDS), bytes)
	} else {
		fmt.Printf("error label %s insertFrequencyWords: %s\n", label, err.Error())
	}
}
func initFrequencyCounter() map[string]FrequencyCounterType {
	m := make(map[string]FrequencyCounterType)
	m["author"] = make(map[string]int)
	m["location"] = make(map[string]int)
	m["keywords"] = make(map[string]int)
	m["comment"] = make(map[string]int)
	return m
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
	countJpegs := 0
	totalSize := 0
	countCreate := 0
	countUpdate := 0
	changedDirs := make([]string, 0)
	m := initFrequencyCounter()
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
			var jd JdirType
			jDirTimeStart := time.Now()

			if jd, err = makeJdir(path); err != nil {
				return err
			}
			m["author"][jd.MostAuthor]++
			m["location"][jd.MostLocation]++
			for _, k := range getCsvKeywords(jd.MostKeywords) {
				m["keywords"][k]++
			}
			m["comment"][jd.MostComment]++

			countFiles += len(jd.Files)
			for _, f := range jd.Files {
				size, _ := strconv.Atoi(f.Size)
				totalSize += size
				if isJpegFile(f.Name) {
					countJpegs++
				}
			}
			verbose := true
			if verbose {
				fmt.Printf("\r%4dd %4df %4s %s  ", countDir, len(jd.Files), time.Since(jDirTimeStart).Truncate(time.Second), path)
			}
			changedDirs = append(changedDirs, path)
			countCreate++
			bytes, err := json.Marshal(jd)
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
	insertLabelSummary(db, arcFS.Label, countDir, countFiles, countJpegs, totalSize, time.Since(startTime))
	insertFrequencyWords(db, arcFS.Label, m)
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
	m := initFrequencyCounter()

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
			var jd JdirType
			json.Unmarshal(bytes, &jd)
			m["author"][jd.MostAuthor]++
			m["location"][jd.MostLocation]++
			for _, k := range getCsvKeywords(jd.MostKeywords) {
				m["keywords"][k]++
			}
			m["comment"][jd.MostComment]++

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
	insertLabelSummary(db, arcFS.Label, countDir, -1, -2, -3, time.Since(startTime))
	insertFrequencyWords(db, LABEL_FREQUENCY_WORDS, m)
	return nil
}

func ArcpicsAllKeys(db *bolt.DB, bucket []byte) []string {
	keys := make([]string, 0)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(bucket)

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
func getCsvKeywords(s string) []string {
	arr := strings.Split(s, ",")
	for i, a := range arr {
		arr[i] = strings.TrimSpace(a)
	}
	return arr
}
func ArcpicsMostOcurrenceStrings(db *bolt.DB) map[string]FrequencyCounterType {

	var bytes []byte
	var m map[string]FrequencyCounterType
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(SYSTEM_BUCKET)
		bytes = b.Get([]byte(LABEL_FREQUENCY_WORDS))
		return nil
	})
	json.Unmarshal(bytes, &m)
	return m
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
