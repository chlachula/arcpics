package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/chlachula/arcpics/jpeg"
)

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No directory argument")
		os.Exit(1)
	}
	dir := os.Args[1]
	defer func(start time.Time) {
		fmt.Printf("Elapsed time %s\n", time.Since(start))
	}(time.Now())
	files := find(dir, ".jpg")
	for _, fname := range files {
		var j jpeg.JpegReader
		verbose := false
		j.Open(fname, verbose)
		fname = "blabla"
		if err := j.Decode(); err != nil {
			fmt.Printf("Error for filename %s\n%s\n", j.Filename, err.Error())
		}
		fmt.Printf("%s %dx%d %s\n", j.Filename, j.ImageWidth, j.ImageHeight, j.Comment)
	}
}
