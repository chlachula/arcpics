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
	htmlPage := `<html><head>  <title>%s</title>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <link rel="icon" type="image/ico" href="favicon.ico">	
<style>
</style>
<script>
 function toggleHideDisplay(myDIV) {
   var x = document.getElementById(myDIV);
   if (x.style.display === "none") {
     x.style.display = "block";
   } else {
     x.style.display = "none";
   }
 }
</script>
</head>
<body style="text-align:left"><a name="top"></a>
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
	fmt.Fprintf(w, `<b>Arcpics</b> is the program for management external picture archives, or any other files.
	<br/><br/>
	External archive can be at external harddrive, external USB flash stick, memory card, etc.
	Archived directory has to be labeled by special file with prefix <b>%s</b> and extension of this file is actually name label,
	for example <b>%sVacation-2023</b>. 
	Label should be written on medium itself, e.g. by marker or pasted tag, to help easie find it. 
	Archived directory is most the time the root directory on archived medium, 
	but it could be also at any subdirectory.
	Label file can be created with option -c, for example:
	<pre>
	arcpics -c /media/joe/USB32/Arc-Pics Vacation-2023
	arcpics -c E:\\Arc-Pics Vacation-2023                #on Windows
	</pre>
	In a next step can be created label database in home subdirectory %s with option -a1, e.g. following command would create file <b>%sVacation-2023.db</b>
	<pre>
	arcpics -a1 /media/joe/USB32/Arc-Pics Vacation-2023
	arcpics -a1 E:\\Arc-Pics Vacation-2023                #on Windows
	</pre>
	In the last step can be used any from label options -ll,  -la, lf, ls to display text information on command line terminal 
	or option -w to start local webserver on default port 8080 which can be modified with option -p.
	For example after following command browser should listen at http://localhost:8081
	<pre>
	arcpics -p 8081 -w
	</pre>
	<br/><br/>
	`, defaultNameDashLabelDot, defaultNameDashLabelDot, dotDefaultName, defaultNameDash)
	fmt.Fprintf(w, `<b>Arcpics command line help text</b>
	<pre>
	%s
	</pre>
	`, CmdHelp(""))
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
		link := fmt.Sprintf(`<a href="/label-dir/%s/%s" title="%s">%s</a>`, label, k, k, k+"/")
		fmt.Fprintf(w, "<br/> %s\n", link)
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
func fixNameLink(w http.ResponseWriter, label, path, name string, dir bool) {
	var linkName string
	fixName := name
	if len(name) >= 24 {
		fixName = name[:21] + "..&gt;"
	}
	if dir {
		fixName += "/"
	}
	if dir {
		linkName = fmt.Sprintf(`<a href="/label-dir/%s/%s" title="%s">%s</a>`, label, path, name, fixName)
	} else if isJpegFile(name) || dir {
		linkName = fmt.Sprintf(`<a href="/label-dir/%s/%s/%s" title="%s">%s</a>`, label, path, name, name, fixName)
	} else {
		linkName = fmt.Sprintf(`<span title="%s">%s</span>`, name, fixName)
	}
	for count := len(name); count < 24; count++ {
		linkName = linkName + " "
	}
	fmt.Fprintf(w, "      %-24s", linkName)
	if dir {
		fmt.Fprint(w, "\n")
	}
}
func prevNextPathLinks(val string, dirName string) (string, string) {
	fmtStr := "<a href=\"%s\">%s</a>"
	prevLink := "|"
	nextLink := "|"
	if dirName == "./" {
		return " ", " "
	}
	var jd JdirType
	if err := json.Unmarshal([]byte(val), &jd); err != nil {
		return "-", "-"
	}
	for i, d := range jd.Dirs {
		if d == dirName {
			if i-1 >= 0 {
				prevLink = fmt.Sprintf(fmtStr, jd.Dirs[i-1], "&lt;-")
			}
			if i+1 < len(jd.Dirs) {
				nextLink = fmt.Sprintf(fmtStr, jd.Dirs[i+1], "-&gt;")
			}
			break
		}
	}
	return prevLink, nextLink
}
func lastDir(path string) string {
	s := strings.Split(path, "/")
	if len(s) < 2 {
		return path
	}
	return s[len(s)-1]
}
func pageLabelDir(w http.ResponseWriter, r *http.Request) {
	params := getParams(`\/label-dir\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)`, r.URL.Path)
	label := params["Label"]
	path := params["Path"]
	if path == "" {
		path = "./"
	}
	//	myUrl, _ := url.Parse(r.RequestURI)
	//	urlParams, _ := url.ParseQuery(myUrl.RawQuery)
	db, err := LabeledDatabase(label)
	var val string
	var parentVal string
	if err == nil {
		val = GetDbValue(db, FILES_BUCKET, path)
		parentDir, _ := getParentDir(path)
		parentVal = GetDbValue(db, FILES_BUCKET, parentDir)
	} else {
		val = err.Error()
	}
	defer db.Close()
	fmt.Fprint(w, pageBeginning("Arcpics Label "+label))
	fmt.Fprint(w, webMenu(""))
	var jd JdirType
	if err = json.Unmarshal([]byte(val), &jd); err != nil {
		lblfmt := "<h2>Arcpics Label: %s</h2>\npath: %s <hr/>\nerror: %s\n"
		fmt.Fprintf(w, lblfmt, label, path, err.Error())
		return
	}

	lblfmt := "<h1>Arcpics Label: %s</h1>\n%s(path: %s)%s %s<hr/>\n"
	comments := ""
	if jd.MostComment != "" {
		comments = "most comments: " + jd.MostComment
	}
	linkPrev, linkNext := prevNextPathLinks(parentVal, lastDir(path))
	fmt.Fprintf(w, lblfmt, label, linkPrev, path, linkNext, comments)
	fmt.Fprint(w, "<button onclick=\"toggleHideDisplay('idFiles')\">Hide/Display Files</button>")

	head := "<pre id=\"idFiles\">%33s %55s %45s %s\n"
	fmt.Fprintf(w, head, `<a href="?C=N;O=D">Name</a>`, `<a href="?C=M;O=A">Last modified</a>`, `<a href="?C=S;O=A">Size</a>`, `<a href="?C=D;O=A">Description</a>`)
	parentDir, _ := getParentDir(path)
	if path != "./" {
		fmt.Fprintf(w, "      <a href=\"/label-dir/%s/%s\">%s</a>\n", label, parentDir, "parent directory")
	}
	for _, f := range jd.Files {
		fixNameLink(w, label, path, f.Name, false)
		fmt.Fprintf(w, "%-26s%10s %s\n", f.Time, f.Size, f.Comment)
	}
	for _, d := range jd.Dirs {
		fixNameLink(w, label, path+"/"+d, d, true)
	}

	//Thumbnail images
	fmt.Fprint(w, "\n</pre><hr/>\n")
	for _, f := range jd.Files {
		if isJpegFile(f.Name) {
			title := f.Name + ": " + f.Comment
			fmt.Fprintf(w, `<img src="/label-dir/%s/%s/%s" title="%s"/>%s`, label, path, f.Name, title, "\n")
		}
	}
	fmt.Fprint(w, "\n<br/>\n")
	fmt.Fprintf(w, lblfmt, label, linkPrev, path, linkNext, " <a href=\"#top\">top</a>")
}

func Web(port int) {
	http.HandleFunc("/", route)

	colonPort := fmt.Sprintf(":%d", port)
	fmt.Printf("... listening at port %d", port)
	log.Fatal(http.ListenAndServe(colonPort, nil))

}
