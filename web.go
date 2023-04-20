package arcpics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"
)

var mutex = &sync.Mutex{}

/* deleteme
https://stackoverflow.com/questions/30474313/how-to-use-regexp-get-url-pattern-in-golang
https://stackoverflow.com/questions/30483652/how-to-get-capturing-group-functionality-in-go-regular-expressions
*/

var reHome = regexp.MustCompile(`(?m)^\/$`)
var reAbout = regexp.MustCompile(`(?m)\/about[\/]{0,1}$`)
var reLabels = regexp.MustCompile(`(?m)^\/labels[\/]{0,1}$`)
var reLabel = regexp.MustCompile(`(?m)\/label\/([a-zA-z0-9]+)\/(.*)$`)

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
	println("path", r.URL.Path)
	switch {
	case reHome.MatchString(r.URL.Path):
		pageHome(w, r)
	case reLabel.MatchString(r.URL.Path):
		webLabel(w, r)
	case reLabels.MatchString(r.URL.Path):
		pageLabels(w, r)
	case reAbout.MatchString(r.URL.Path):
		pageAbout(w, r)
	default:
		w.Write([]byte("<a href='/'>home</a> Unrecognized URL Pattern r.URL.Path=" + r.URL.Path))
	}
}

func pageHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webBegin("Arcpics home"))
	fmt.Fprint(w, webMenu("/"))
	fmt.Fprint(w, "<h1>Home</h1>")
}
func pageAbout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webBegin("About arcpics"))
	fmt.Fprint(w, webMenu("/about"))
	fmt.Fprint(w, "<h1>About Arcpics</h1>")
	fmt.Fprint(w, "Here is the description ...")
}
func pageLabels(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, webBegin("Arcpics Labels"))
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
			labelsString += fmt.Sprintf(`<br/><a href="/label/%s/">%s</a>%s`, label, label, "\n")
		}
		labelsString += "\n"
	}
	htmlPage := `<h1>Arcpics Labels:</h1>%s<hr/><h5>Label dababases are located inside of %s`
	fmt.Fprintf(w, htmlPage, labelsString, dbDir)
	mutex.Unlock()
}
func webBegin(title string) string {
	htmlPage := `<html><head><title>%s</title>
<style>
</style>
</head>
<body style="text-align:center">
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

func webLabel(w http.ResponseWriter, r *http.Request) {
	htmlPage := `<html><title>Label %s</title>
<body style="text-align:center"> <a href="/">home</a><hr/>
<h1>Arcpics Label: %s</h1>
path: %s <hr/>
value: 
<pre>
%s
</pre>
`
	params := getParams(`\/label\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)`, r.URL.Path)
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
	fmt.Fprintf(w, htmlPage, label, label, path, val)
	println(label, " --------> ", "`"+path+"`", val)
	var jd JdirType
	err = json.Unmarshal([]byte(val), &jd)
	if err == nil {
		for _, d := range jd.Dirs {
			fmt.Fprintf(w, "<br/><a href=\"/label/%s/%s\">%s</a>\n", label, d, d)
		}
	}
}

func Web(port int) {
	http.HandleFunc("/", route)

	colonPort := fmt.Sprintf(":%d", port)
	fmt.Printf("... listening at port %d", port)
	log.Fatal(http.ListenAndServe(colonPort, nil))

}
