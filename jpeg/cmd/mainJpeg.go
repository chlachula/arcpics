package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chlachula/arcpics/jpeg"
)

var version string = "0.0.1"
var ThumbnailPrefix = "Thumbnail-"
var verbose bool = false
var createThumbnails bool = false
var dir string

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			if !strings.HasPrefix(d.Name(), ThumbnailPrefix) {
				a = append(a, s)
			}
		}
		return nil
	})
	return a
}

func createFileFromBytes(fname string, data []byte) {
	// create and open a file
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		panic(err)
	}
}
func thumbnailName(fname string) string {
	separator := "/"
	s := strings.Split(fname, separator)
	size := len(s)
	s[size-1] = ThumbnailPrefix + s[size-1]
	fname2 := ""
	sep := ""
	for _, a := range s {
		fname2 = fname2 + sep + a
		sep = separator
	}
	return fname2
}

var helpText = `=== jpegshow: display jpeg comments, size and create extracted thumbnails ===
ver %s
Usage arguments:
 -h help text
 -d DirName  #folder with .jpg files
 -t          #create extracted thumbnails files
 -v          #verbose output

Examples:
-d ~/myjpegs
-v -t -d /home/joe/myjpegs
`

func help(msg string) {
	if msg != "" {
		fmt.Println(msg)
	}
	fmt.Printf(helpText, version)

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
	if len(os.Args) < 2 {
		help("Not enough arguments")
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
		case "-d":
			i = increaseAndCheckArgumentIndex(i, "No directory after -d")
			dir = os.Args[i]
		case "-t":
			createThumbnails = true
		case "-v":
			verbose = true
		default:
			help("Unknown option '" + arg + "'")
			os.Exit(1)
		}
	}

	defer func(start time.Time) {
		fmt.Printf("Elapsed time %s\n", time.Since(start))
	}(time.Now())
	files := find(dir, ".jpg")
	infoCount := 0
	for _, fname := range files {
		var j jpeg.JpegReader
		j.Open(fname, verbose)
		fname = "blabla"
		if err := j.Decode(); err != nil {
			fmt.Printf("Error for filename %s\n%s\n", j.Filename, err.Error())
		}
		fmt.Printf("\n%s %dx%d %s\n", j.Filename, j.ImageWidth, j.ImageHeight, j.Comment)
		if createThumbnails {
			if len(j.Thumbnail) > 0 {
				tname := thumbnailName(j.Filename)
				createFileFromBytes(tname, j.Thumbnail)
				fmt.Printf("Created %s\n", tname)
			} else {
				fmt.Printf("INFO: no thumbnail for %s\n", j.Filename)
				infoCount++
			}
		}
	}
	if infoCount > 0 {
		fmt.Printf("There were %d INFO messages\n", infoCount)
	}
}
