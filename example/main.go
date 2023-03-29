package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

func directoriesWalk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		fmt.Printf("dir: %v\n", path)
	} else {
		fmt.Printf("     %-20s %10d %10s %10s\n", info.Name(), info.Size(), info.Mode(), info.ModTime())
		if db == nil {
			panic(fmt.Errorf("db is nil"))
		}
		if getDbValue(db, FILES_BUCKET, path) != path {
			err := putDbValue(db, FILES_BUCKET, path, info.Name()+"-"+info.ModTime().String())
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
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
	var picturesDirName string
	var err error
	picturesDirName, db, err = AssignPicturesDirectoryWithDatabase(os.Args[1:])

	err = filepath.Walk(picturesDirName, directoriesWalk)
	if err != nil {
		help(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	println("done...")
}
