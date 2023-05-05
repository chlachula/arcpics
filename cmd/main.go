package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chlachula/arcpics"
)

var db *bolt.DB
var version string = "0.0.3"
var port = 8080

var helpText = `=== arcpics: manage archived of pictures not only at external hard drives ===
ver %s
Usage arguments:
 -h help text
 -a1 picturesDirName      #write arcpics.json dir files directly to DB in 1 step
 -af picturesDirName      #update arcpics.json dir files
 -ab picturesDirName      #update database 
 -d databaseDirName       #database dir location other then default ~/.arcpics
 -c databaseDirName label #create label inside of database directory
 -ll                      #list all labels
 -la label                #list all dirs on USB  with this label
 -lf label                #word frequency
 -ls label query          #list specific resources
 -v                       #verbose output
 -p port                  #web port definition
 -w                       #start web - default port 8080

Examples:
-c %sArc-Pics Vacation-2023 #creates label file inside of directory %sArc-Pics
-af %sArc-Pics               #updates arcpics.json dir files inside of directories at %sArc-Pics
-ab %sArc-Pics               #updates database %sVacation-2023.db
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
func write_dirs_to_db(i int, errMsg string) int {
	var arcFS arcpics.ArcpicsFS
	var err error
	i = increaseAndCheckArgumentIndex(i, errMsg)
	arcFS, db, err = arcpics.AssignPicturesDirectoryWithDatabase(os.Args[i])
	exitIfErrorNotNil(err)
	err = arcpics.ArcpicsFiles2DB(db, arcFS.Dir)
	exitIfErrorNotNil(err)
	db.Close()
	return i
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
		case "-a1":
			i = write_dirs_to_db(i, "No picturesDirName after -a1")
		case "-af":
			i = update_dirs_or_db(i, updateDirs, "No picturesDirName after -af")
		case "-ab":
			i = update_dirs_or_db(i, !updateDirs, "No picturesDirName after -ab")
		case "-c":
			i = increaseAndCheckArgumentIndex(i, "No picturesDirName after -c")
			dirName := os.Args[i]
			if !arcpics.DirExists(dirName) {
				exitIfErrorNotNil(fmt.Errorf("directory %s not found", dirName))
			}
			i = increaseAndCheckArgumentIndex(i, "No label -c")
			newLabel := os.Args[i]
			err := arcpics.CreateLabelFile(dirName, newLabel)
			exitIfErrorNotNil(err)
		case "-la":
			i = increaseAndCheckArgumentIndex(i, "No label after -a")
			db, err := arcpics.LabeledDatabase(os.Args[i])
			exitIfErrorNotNil(err)
			keys := arcpics.ArcpicsAllKeys(db)
			for _, k := range keys {
				fmt.Println(k)
			}
		case "-ll":
			i++
			dbDir := arcpics.GetDatabaseDirName()
			labels, err := arcpics.GetLabelsInDbDir(dbDir)
			exitIfErrorNotNil(err)
			fmt.Printf("Labels in %s %v\n", dbDir, labels)
		case "-ls":
			i = increaseAndCheckArgumentIndex(i, "No label after -s")
			db, err := arcpics.LabeledDatabase(os.Args[i])
			exitIfErrorNotNil(err)
			i = increaseAndCheckArgumentIndex(i, "No query after -s")
			arcpics.ArcpicsQuery(db, os.Args[i])
		case "-lf":
			i = increaseAndCheckArgumentIndex(i, "No label after -f")
			db, err := arcpics.LabeledDatabase(os.Args[i])
			exitIfErrorNotNil(err)
			arcpics.ArcpicsWordFrequency(db)
		case "-p":
			i = increaseAndCheckArgumentIndex(i, "No port after -p")
			p, err := strconv.Atoi(os.Args[i])
			exitIfErrorNotNil(err)
			minPort := 0
			maxPort := 65535
			if p < minPort || p > maxPort {
				err = fmt.Errorf("port %d out of expected interval %d..%d", p, minPort, maxPort)
				exitIfErrorNotNil(err)
			}
			port = p
		case "-v":
			arcpics.Verbose = true
		case "-w":
			arcpics.Web(port)
		default:
			help("Unknown option '" + arg + "'")
			os.Exit(1)
		}
	}

	println("done...")
}
