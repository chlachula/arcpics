package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chlachula/arcpics"
)

var db *bolt.DB

/*
	func directoriesWalk(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			fmt.Printf("dir: %v\n", path)
		} else {
			if db == nil {
				panic(fmt.Errorf("db is nil"))
			}
			if getDbValue(db, FILES_BUCKET, path) == "" {
				fmt.Printf("     %-20s %10d %10s %10s\n", info.Name(), info.Size(), info.Mode(), info.ModTime())
				value := info.Name() + "-" + info.ModTime().String()
				err := putDbValue(db, FILES_BUCKET, path, value)
				if err != nil {
					panic(err)
				}
				fmt.Printf("     added value %s\n", value)
			}
		}
		return nil
	}
*/
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
	defer func(start time.Time) {
		fmt.Printf("Elapsed time %s\n", time.Since(start))
	}(time.Now())
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-h") {
		help("")
		os.Exit(0)
	}
	var picturesDirName string
	var err error
	picturesDirName, db, err = arcpics.AssignPicturesDirectoryWithDatabase(os.Args)
	if err != nil {
		help(err.Error())
		os.Exit(1)
	}
	arcFS, err := arcpics.ArcpicsFS(picturesDirName)
	if err != nil {
		fmt.Println("error: " + err.Error())
		os.Exit(1)
	}
	fmt.Println("arcFS.Dir=", arcFS.Dir)
	arcpics.ArcpicsFilesUpdate(arcFS.Dir)
	/*
		err = filepath.Walk(picturesDirName, directoriesWalk)
		if err != nil {
			help(err.Error())
			os.Exit(1)
		}
	*/

	defer db.Close()

	println("done...")
}
