package arcpics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

type JfileType = struct {
	Name    string
	Time    string
	Comment string
}
type JdirType = struct {
	Description string
	Location    string
	Files       []JfileType
}

var SYSTEM_BUCKET = []byte("SYSTEM")
var FILES_BUCKET = []byte("FILES")
var INIT_LABEL_KEY = "ARC-PICS-LABEL-KEY"
var defaultPicturesDirName = "Arc-Pics"
var defaultDatabaseDirName = "DB"
var defaultArcpicsDbLabel = "arcpics-db-label."
var jsonFilePrefix = "sample.json"

func picturesAndDatabaseDirectories(args []string) (string, string) {
	picturesDirName := defaultPicturesDirName
	databaseDirName := defaultDatabaseDirName

	if len(args) < 2 {
		return picturesDirName, databaseDirName

	}
	if len(args) >= 2 {
		picturesDirName = args[1]

	}
	if len(args) > 2 {
		databaseDirName = args[2]
	}
	return picturesDirName, databaseDirName
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func insertSystemLabelValue(db *bolt.DB, value string) error {
	//return insertDbValue(db, SYSTEM_BUCKET, keyBytes)
	keyBytes := []byte(INIT_LABEL_KEY)
	err := db.Update(func(tx *bolt.Tx) error {
		foundBytes := tx.Bucket([]byte(SYSTEM_BUCKET)).Get(keyBytes)
		if len(foundBytes) == 0 {
			err := tx.Bucket([]byte(SYSTEM_BUCKET)).Put(keyBytes, []byte(value))
			if err != nil {
				return fmt.Errorf("bucket %s - Could not insert entry: %v", SYSTEM_BUCKET, err)
			}
		}
		return nil
	})
	return err
}
func insertNewBucket(db *bolt.DB, bucket []byte) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		panic(err)
	}
}

////////////////////////
// Exported functions //
////////////////////////

func DbLabel(archiveDir string) (string, error) {
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

func PutDbValue(db *bolt.DB, bucket []byte, key, value string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(key), []byte(value))
	})
	return err
}

func GetDbValue(db *bolt.DB, bucket []byte, key string) string {
	var valueBytes []byte
	_ = db.View(func(tx *bolt.Tx) error {
		valueBytes = tx.Bucket([]byte(bucket)).Get([]byte(key))
		return nil
	})
	return string(valueBytes)
}

func AssignPicturesDirectoryWithDatabase(args []string) (string, *bolt.DB, error) {
	picturesDirName, databaseDirName := picturesAndDatabaseDirectories(args)
	label, err := DbLabel(picturesDirName)
	if err != nil {
		return picturesDirName, nil, err
	}
	databaseName := filepath.Join(databaseDirName, "arcpics-"+label+".db")
	dbDidExist := fileExists(databaseName)

	// Open the database data file. It will be created if it doesn't exist.
	var db *bolt.DB
	db, err = bolt.Open(databaseName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	if dbDidExist {
		// check INIT key
		s := GetDbValue(db, SYSTEM_BUCKET, INIT_LABEL_KEY)
		if s != label {
			return picturesDirName, db, fmt.Errorf("init value  %s at DB %s doesn't match %s at dir %s", s, databaseName, label, picturesDirName)
		}
	} else {
		insertNewBucket(db, SYSTEM_BUCKET) // insert SYSTEM bucket just once

		// insert init key into SYSTEM bucket just once
		err = insertSystemLabelValue(db, label)
		if err != nil {
			panic(err)
		}
		insertNewBucket(db, FILES_BUCKET) // insert FILES bucket just once

	}
	return picturesDirName, db, nil
}

func CreateDirJson(d string, dirFiles JdirType) error {
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
	var fjson_FileInfo os.FileInfo
	var err error
	if fjson_FileInfo, err = os.Stat(fjson); err != nil {
		return nil // there is now current file
	}
	timeExtension := fjson_FileInfo.ModTime().Format("--2006-01-02_15-04-05") // Mon Jan 2 15:04:05 -0700 MST 2006
	fjson_time := filepath.Join(d, jsonFilePrefix+timeExtension)
	if err := os.Rename(fjson, fjson_time); err != nil {
		return err
	}
	if err := os.Rename(fjson_new, fjson); err != nil {
		return err
	}
	return nil
}
