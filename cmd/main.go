package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chlachula/arcpics"
)

var db *bolt.DB
var version string = "0.0.2"

var helpText = `=== arcpics: manage archived of pictures not only at external hard drives ===
ver %s
Usage arguments:
 -h help text
 -u picturesDirName       #update arcpics.json dir files
 -b picturesDirName       #update database about 
 -d databaseDirName       #database dir location other then default ~/.arcpics
 -c databaseDirName label #create label inside of database directory
 -a label                 #list all dirs on USB  with this label
 -s label query           #list specific resources
 -l                       #list all labels
 -p port                  #web port definition
 -w                       #start web - default port 8080

Examples:
-c %sArc-Pics Vacation-2023 #creates label file inside of directory %sArc-Pics
-u %sArc-Pics               #updates arcpics.json dir files inside of directories at %sArc-Pics
-b %sArc-Pics               #updates database %sVacation-2023.db
`

func help(msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	r := "/media/joe/USB32/"
	h := "~/.arcpics/"
	if runtime.GOOS == "windows" {
		r = "E:\\"
		h = "C:\\Users\\joe\\.arcpics\\"
	}
	fmt.Printf(helpText, version, r, r, r, r, r, h)
}
func update_dirs_or_db(i int, updateDirs bool, errMsg string) int {
	var arcFS arcpics.ArcpicsFS
	var err error
	i = increaseAndCheckArgumentIndex(i, errMsg)
	arcFS, db, err = arcpics.AssignPicturesDirectoryWithDatabase(os.Args[i])
	exitIfErrorNotNil(err)
	if updateDirs {
		err = arcpics.ArcpicsFilesUpdate(arcFS.Dir)
	} else {
		err = arcpics.ArcpicsDatabaseUpdate(db, arcFS.Dir)
	}
	exitIfErrorNotNil(err)
	db.Close()
	return i
}
func exitIfErrorNotNil(err error) {
	if err != nil {
		help(err.Error())
		os.Exit(1)
	}
}
func increaseAndCheckArgumentIndex(i int, errMsg string) int {
	i++
	if i >= len(os.Args) {
		help(errMsg)
		os.Exit(1)
	}
	return i
}
func main() {
	defer func(start time.Time) {
		fmt.Printf("Elapsed time %s\n", time.Since(start))
	}(time.Now())
	var updateDirs = true
	if len(os.Args) < 2 {
		help("No argument")
		os.Exit(1)
	}
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if !strings.HasPrefix(arg, "-") {
			help("Option '" + arg + "'does not start with char '-'")
			os.Exit(1)

		}
		switch arg {
		case "-h":
			help("")
			os.Exit(0)
		case "-u":
			i = update_dirs_or_db(i, updateDirs, "No picturesDirName after -u")
		case "-b":
			i = update_dirs_or_db(i, !updateDirs, "No picturesDirName after -b")
		case "-a":
			i = increaseAndCheckArgumentIndex(i, "No label after -a")
			db, err := arcpics.LabeledDatabase(os.Args[i])
			exitIfErrorNotNil(err)
			arcpics.ArcpicsAllKeys(db)
		case "-l":
			i++
			dbDir := arcpics.GetDatabaseDirName()
			labels, err := arcpics.GetLabelsInDbDir(dbDir)
			exitIfErrorNotNil(err)
			fmt.Printf("Labels in %s %v\n", dbDir, labels)
		default:
			help("Unknown option '" + arg + "'")
			os.Exit(1)
		}
	}

	println("done...")
}
