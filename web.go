package arcpics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var mutex = &sync.Mutex{}

var reHome = regexp.MustCompile(`(?m)^\/$`)
var reAbout = regexp.MustCompile(`(?m)\/about[\/]{0,1}$`)
var reLabels = regexp.MustCompile(`(?m)^\/labels[\/]{0,1}$`)
var reLabelList = regexp.MustCompile(`(?m)\/label-list\/([a-zA-z0-9]+)$`)
var reLabelDir = regexp.MustCompile(`(?m)\/label-dir\/([a-zA-z0-9]+)\/(.*)$`)
var reLabelFileJpegStr = `(?m)\/label-dir\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)\/(?P<JpgFile>.*\.jpg)$`
var reLabelFileJpeg = regexp.MustCompile(reLabelFileJpegStr)

/**
 * Parses url with the given regular expression and returns the
 * group values defined in the expression.
 *
 */
func getParams(regEx, url string) (paramsMap map[string]string) {
	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func route(w http.ResponseWriter, r *http.Request) {
	println("route path:", r.URL.Path)
	switch {
	case reHome.MatchString(r.URL.Path):
		pageHome(w, r)
	case reLabelList.MatchString(r.URL.Path):
		println("case reLabelList:", r.URL.Path)
		pageLabelList(w, r)
	case reLabelFileJpeg.MatchString(r.URL.Path):
		println("case reLabelFileJpeg:", r.URL.Path)
		pageLabelFileJpeg(w, r)
	case reLabelDir.MatchString(r.URL.Path):
		println("case reLabelDir:", r.URL.Path)
		pageLabelDir(w, r)
	case reLabels.MatchString(r.URL.Path):
		pageLabels(w, r)
	case reAbout.MatchString(r.URL.Path):
		pageAbout(w, r)
	default:
		w.Write([]byte("<a href='/'>home</a> Unrecognized URL Pattern r.URL.Path=" + r.URL.Path))
	}
}

func pageBeginning(title string) string {
	htmlPage := `<html><head><title>%s</title>
<style>
</style>
</head>
<body style="text-align:left">
`
	return fmt.Sprintf(htmlPage, title)
}
func webMenu(link string) string {
	items := []struct {
		L string
		N string
	}{
		{"/", "Home"},
		{"/labels", "Labels"},
		{"/about", "About"},
	}
	s := ""
	for _, it := range items {
		s += ` <span class="mItem">`
		if link == it.L {
			s += it.N
		} else {
			s += fmt.Sprintf(`<a href="%s">%s</a>`, it.L, it.N)
		}
		s += "</span> "
	}
	s += "\n<hr/>\n"
	return s
}

func pageHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageBeginning("Arcpics home"))
	fmt.Fprint(w, webMenu("/"))
	fmt.Fprint(w, "<h1>Home</h1>")
}
func pageAbout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageBeginning("About arcpics"))
	fmt.Fprint(w, webMenu("/about"))
	fmt.Fprint(w, "<h1>About Arcpics</h1>")
	fmt.Fprint(w, "Here is the description ...")
}
func pageLabels(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageBeginning("Arcpics Labels"))
	fmt.Fprint(w, webMenu("/labels"))
	mutex.Lock()
	dbDir := GetDatabaseDirName()
	labels, err := GetLabelsInDbDir(dbDir)
	labelsString := ""
	if err != nil {
		labelsString = err.Error()
	}
	if len(labels) < 1 {
		labelsString += " no labels"
	} else {
		labelsString += "\n"
		for _, label := range labels {
			labelsString += fmt.Sprintf(`<br/>%s `, label)
			labelsString += fmt.Sprintf(`<a href="/label-list/%s">list</a> %s `, label, "\n")
			labelsString += fmt.Sprintf(`<a href="/label-dir/%s/">dir</a>%s`, label, "\n")
		}
		labelsString += "\n"
	}
	htmlPage := `<h1>Arcpics Labels:</h1>%s<hr/><h5>Label dababases are located inside of %s`
	fmt.Fprintf(w, htmlPage, labelsString, dbDir)
	mutex.Unlock()
}

func pageLabelList(w http.ResponseWriter, r *http.Request) {
	params := getParams(`\/label-list\/(?P<Label>[a-zA-z0-9]+)$`, r.URL.Path)
	label := params["Label"]
	var keys []string
	db, err := LabeledDatabase(label)
	if err == nil {
		keys = ArcpicsAllKeys(db)
	}
	defer db.Close()
	fmt.Fprint(w, pageBeginning("Arcpics Label "+label+" list"))
	fmt.Fprint(w, webMenu(""))
	lblfmt := "<h1>Arcpics Label %s list</h1>\n"
	fmt.Fprintf(w, lblfmt, label)
	for _, k := range keys {
		fmt.Fprintf(w, "<br/>%s\n", k)
	}
}

func pageLabelFileJpeg(w http.ResponseWriter, r *http.Request) {
	params := getParams(reLabelFileJpegStr, r.URL.Path)
	label := params["Label"]
	path := params["Path"]
	jpgFile := params["JpgFile"]
	if path == "" {
		path = "./"
	}
	db, err := LabeledDatabase(label)
	var val string
	if err == nil {
		val = GetDbValue(db, FILES_BUCKET, path)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "<h1>Internal server error: %s</h1>", err.Error())
		return
	}
	defer db.Close()
	var jd JdirType
	err = json.Unmarshal([]byte(val), &jd)
	ok := false
	if err == nil {
		for _, f := range jd.Files {
			if jpgFile == f.Name {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(f.Thumbnail)
				ok = true
				break
			}
		}
	}
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>%d - File not found: %s</h1>", http.StatusNotFound, jpgFile)
	}
}

func pageLabelDir(w http.ResponseWriter, r *http.Request) {
	params := getParams(`\/label-dir\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)`, r.URL.Path)
	label := params["Label"]
	path := params["Path"]
	if path == "" {
		path = "./"
	}
	db, err := LabeledDatabase(label)
	var val string
	if err == nil {
		val = GetDbValue(db, FILES_BUCKET, path)
	} else {
		val = err.Error()
	}
	defer db.Close()
	fmt.Fprint(w, pageBeginning("Arcpics Label "+label))
	fmt.Fprint(w, webMenu(""))
	lblfmt := "<h1>Arcpics Label: %s</h1>\npath: %s <hr/>\nvalue: \n<pre>\n%s\n</pre>"
	fmt.Fprintf(w, lblfmt, label, path, val)

	head := "<pre>%33s %55s %45s %s\n"
	fmt.Fprintf(w, head, `<a href="?C=N;O=D">Name</a>`, `<a href="?C=M;O=A">Last modified</a>`, `<a href="?C=S;O=A">Size</a>`, `<a href="?C=D;O=A">Description</a>`)
	var jd JdirType
	err = json.Unmarshal([]byte(val), &jd)
	if err == nil {
		for _, f := range jd.Files {
			name := f.Name
			if strings.HasSuffix(name, ".jpg") {
				name = fmt.Sprintf(`<a href="/label-dir/%s/%s/%s">%s</a>`, label, path, name, name)
			}
			for count := len(f.Name); count < 24; count++ {
				name = name + " "
			}
			fmt.Fprintf(w, "      %-24s%-26s%10s %s\n", name, f.Time, f.Size, f.Comment)
		}
		parentDir, _ := getParentDir(path)
		if path != "./" {
			fmt.Fprintf(w, "      <a href=\"/label-dir/%s/%s\">%s</a>\n", label, parentDir, "parent directory")
		}
		for _, d := range jd.Dirs {
			fmt.Fprintf(w, "      <a href=\"/label-dir/%s/%s\">%s</a>\n", label, path+"/"+d, d)
		}
	}
}

func Web(port int) {
	http.HandleFunc("/", route)

	colonPort := fmt.Sprintf(":%d", port)
	fmt.Printf("... listening at port %d", port)
	log.Fatal(http.ListenAndServe(colonPort, nil))

}
