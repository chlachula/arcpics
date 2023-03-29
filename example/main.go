package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

var SYSTEM_BUCKET = "SYSTEM"
var INIT_POSTFIX_KEY = "ARC-PICS-POSTFIX-KEY"

func volumeName(relativePath string) string {
	absPath, err := filepath.Abs(relativePath)
	if err == nil {
		return filepath.VolumeName(absPath)
	}
	return ""
}
func dbPostfix(archiveDir string) (string, error) {
	NAME_START := "arcpics-db-postfix."
	postfix := ""
	files, err := ioutil.ReadDir(archiveDir)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), NAME_START) {
			count++
			postfix = f.Name()[len(NAME_START):]
		}
	}
	if count == 0 {
		return postfix, fmt.Errorf("there is no file %s* at directory %s - should be e.g. %s001", NAME_START, archiveDir, NAME_START)
	} else if count > 1 {
		return postfix, fmt.Errorf("unexpected number files %s* at directory %s", NAME_START, archiveDir)
	}
	if len(postfix) < 1 {
		return postfix, fmt.Errorf("there is not at least one character postfix after dot 'arcpics-db-postfix.'")
	}
	return postfix, nil
}
func directoriesWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		fmt.Printf("dir: %v\n", path)
	} else {
		fmt.Printf("     %-20s %10d %10s %10s\n", info.Name(), info.Size(), info.Mode(), info.ModTime())
	}
	return nil
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
func insertPostfixValue(db *bolt.DB, value string) error {
	keyBytes := []byte(INIT_POSTFIX_KEY)
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

func getDbValue(db *bolt.DB, bucket string, key string) string {
	keyBytes := []byte(key)
	var valueBytes []byte
	_ = db.View(func(tx *bolt.Tx) error {
		valueBytes = tx.Bucket([]byte(bucket)).Get(keyBytes)
		return nil
	})
	return string(valueBytes)
}

var helpText = `=== arcpics: manage archived of pictures not only at external hard drives ===
Usage arguments:
 [-h] help text
 [picturesDirName] [databaseDirName]
Examples:
 -h
 ..no arguments - default relative directories: Arc-Pics  DB
 E:\Arc-Pics  C:\Users\Joe\DB
`

func help(msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Println(helpText)
}
func main() {
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-h") {
		help("")
		os.Exit(0)
	}
	picturesDirName, databaseDirName := picturesAndDatabaseDirectories(os.Args[1:])
	postFix, err := dbPostfix(picturesDirName)
	if err != nil {
		help(err.Error())
		os.Exit(1)
	}
	databaseName := filepath.Join(databaseDirName, "arcpics-"+postFix+".db")

	dbDidExist := fileExists(databaseName)
	// Open the database data file. It will be created if it doesn't exist.
	db, err := bolt.Open(databaseName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if dbDidExist {
		// check INIT key
		s := getDbValue(db, SYSTEM_BUCKET, INIT_POSTFIX_KEY)
		if s != postFix {
			help(fmt.Sprintf("INIT value  %s at DB %s doesn't match %s at dir %s", s, databaseName, postFix, picturesDirName))
			os.Exit(1)
		}
		println("DB ALREADY EXISTED")
	} else {
		// insert SYSTEM bucket just once
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(SYSTEM_BUCKET))
			if err != nil {
				return fmt.Errorf("could not create root bucket: %v", err)
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		// insert init key into SYSTEM bucket just once
		err = insertPostfixValue(db, postFix)
		if err != nil {
			panic(err)
		}
		err = db.Close()
		if err != nil {
			panic(err)
		}
		println("DB DID NOT EXISTED !!!!!")
	}

	fmt.Println("The Volume Name is: ", volumeName(picturesDirName))
	err = filepath.Walk(picturesDirName, directoriesWalk)
	if err != nil {
		log.Println(err)
	}
	println("done...")
}
