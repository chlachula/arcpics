package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/chlachula/arcpics"
)

var db *bolt.DB
var port = 8080

func help(msg string) {
	fmt.Print(arcpics.CmdHelp(msg))
}
func write_dirs_to_db(i int, errMsg string) int {
	var arcFS arcpics.ArcpicsFS
	var err error
	i = increaseAndCheckArgumentIndex(i, errMsg)
	arcFS, db, err = arcpics.AssignPicturesDirectoryWithDatabase(os.Args[i])
	exitIfErrorNotNil(err)
	err = arcpics.ArcpicsFiles2DB(db, arcFS)
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
		err = arcpics.ArcpicsFilesUpdate(arcFS)
	} else {
		err = arcpics.ArcpicsDatabaseUpdate(db, arcFS)
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
			keys := arcpics.ArcpicsAllKeys(db, arcpics.FILES_BUCKET)
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
		case "-m":
			i = increaseAndCheckArgumentIndex(i, "No dir to mount after -m")
			if err := arcpics.MountLabeledDirectory(os.Args[i]); err != nil {
				fmt.Println(err.Error())
			}
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
