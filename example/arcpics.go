package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

var SYSTEM_BUCKET = []byte("SYSTEM")
var FILES_BUCKET = []byte("FILES")
var INIT_LABEL_KEY = "ARC-PICS-LABEL-KEY"

func dbLabel(archiveDir string) (string, error) {
	NAME_START := "arcpics-db-label."
	label := ""
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), NAME_START) {
			count++
			label = f.Name()[len(NAME_START):]
		}
	}
	if count == 0 {
		return label, fmt.Errorf("there is no file %s* at directory %s - should be e.g. %s001", NAME_START, archiveDir, NAME_START)
	} else if count > 1 {
		return label, fmt.Errorf("unexpected number files %s* at directory %s", NAME_START, archiveDir)
	}
	if len(label) < 1 {
		return label, fmt.Errorf("there is not at least one character label after dot 'arcpics-db-label.'")
	}
	return label, nil
}
func picturesAndDatabaseDirectories(args []string) (string, string) {
	picturesDirName := "Arc-Pics"
	databaseDirName := "DB"

	if len(args) < 1 {
		return picturesDirName, databaseDirName

	} else if len(args) > 1 {
		picturesDirName = args[0]

	} else if len(args) >= 2 {
		databaseDirName = args[1]
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
func putDbValue(db *bolt.DB, bucket []byte, key, value string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).Put([]byte(key), []byte(value))
	})
	return err
}
func getDbValue(db *bolt.DB, bucket []byte, key string) string {
	var valueBytes []byte
	_ = db.View(func(tx *bolt.Tx) error {
		valueBytes = tx.Bucket([]byte(bucket)).Get([]byte(key))
		return nil
	})
	return string(valueBytes)
}

func AssignPicturesDirectoryWithDatabase(args []string) (string, *bolt.DB, error) {
	picturesDirName, databaseDirName := picturesAndDatabaseDirectories(os.Args[1:])
	label, err := dbLabel(picturesDirName)
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
		s := getDbValue(db, SYSTEM_BUCKET, INIT_LABEL_KEY)
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
