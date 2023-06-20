package arcpics

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

var reHome = regexp.MustCompile(`(?m)^\/$`)
var reAbout = regexp.MustCompile(`(?m)\/about[\/]{0,1}$`)
var reMount = regexp.MustCompile(`(?m)\/mount[\/]{0,1}$`)
var reSearchStr = `(?m)\/search[\/]{0,1}(?P<Query>.*)$`
var reSearch = regexp.MustCompile(reSearchStr)
var reLabels = regexp.MustCompile(`(?m)^\/labels[\/]{0,1}$`)
var reLabelList = regexp.MustCompile(`(?m)\/label-list\/([a-zA-z0-9]+)$`)
var reLabelDir = regexp.MustCompile(`(?m)\/label-dir\/([a-zA-z0-9]+)\/(.*)$`)
var reLabelFileJpegStr = `(?m)\/label-dir\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)\/(?P<JpgFile>.*\.jpg)$`
var reLabelFileJpeg = regexp.MustCompile(reLabelFileJpegStr)

// var reLabelDirIndexStr = `(?m)\/label-dir-index\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)\/(?P<Index>\d+)`
// var reLabelDirIndex = regexp.MustCompile(reLabelDirIndexStr)
var reImageFileStr = `(?m)^\/image(?P<Fname>\/.+)$`
var reImageFile = regexp.MustCompile(reImageFileStr)
var reImageIndexStr = `(?m)\/imageindex\/(?P<Dir>.*)\/(?P<File>.*)(\?|)`
var reImageIndex = regexp.MustCompile(reImageIndexStr)

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
	case reSearch.MatchString(r.URL.Path):
		println("case reSearch:", r.URL.Path)
		pageSearch(w, r)
	case reLabelList.MatchString(r.URL.Path):
		println("case reLabelList:", r.URL.Path)
		pageLabelList(w, r)
	case reLabelFileJpeg.MatchString(r.URL.Path):
		println("case reLabelFileJpeg:", r.URL.Path)
		pageLabelFileJpeg(w, r)
	case reImageFile.MatchString(r.URL.Path):
		println("case reImageFile:", r.URL.Path)
		pageImageFile(w, r)
	case reImageIndex.MatchString(r.URL.Path):
		println("case reLabelDirIndex:", r.URL.Path)
		pageImageIndex(w, r)
	case reLabelDir.MatchString(r.URL.Path):
		println("case reLabelDir:", r.URL.Path)
		pageLabelDir(w, r)
	case reLabels.MatchString(r.URL.Path):
		pageLabels(w, r)
	case reAbout.MatchString(r.URL.Path):
		pageAbout(w, r)
	case reMount.MatchString(r.URL.Path):
		pageMount(w, r)
	default:
		w.Write([]byte("<a href='/'>home</a> Unrecognized URL Pattern r.URL.Path=" + r.URL.Path))
	}
}

func pageBeginning(title, jsFiles string) string {
	htmlPage := `<html><head>  <title>%s</title>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <link rel="icon" type="image/ico" href="https://raw.githubusercontent.com/chlachula/arcpics/main/cmd/favicon-Arc-bw-32.ico">	
<style>
* {	box-sizing: border-box;  }

  /* Tree unequal columns that floats next to each other */
  .column {  float: left;  padding-left:10px;padding-right:10px;padding-top:5x;padding-bottom:5px;}

  .left   {  width: 15%%;}
  .middle {  width: 10%%;}
  .right  {  width: 75%%;}
  .c3left   {  width: 15%%;}
  .c3middle {  width: 70%%;}
  .c3right  {  width: 15%%;}
  .c2left   {  width: 15%%;}
  .c2right  {  width: 85%%;}

  /* Clear floats after the columns */
  .row:after {  content: "";display: table;clear: both;}
</style>
<script>
  var globalLabel = "GlobalLabelNotSet"
  var mountWindow
  var myWindow;
  %s
function openWin(url, title) {
	var w = 1200;
	var h = 800;
	var left = (screen.width/2)-(w/2);
	var top = (screen.height/2)-(h/2);
	myWindow = window.open(url, title, 'toolbar=no, location=no, directories=no, status=no, menubar=no, scrollbars=no, resizable=no, copyhistory=no, width='+w+', height='+h+', top='+top+', left='+left);
}

function closeWin() {
  myWindow.close();
}

  function checkShowHide(chk, id) {
	var checkBox = document.getElementById(chk);
	var text = document.getElementById(id);
	if (checkBox.checked == true){
	   text.style.display = "block";
	} else {
	   text.style.display = "none";
	}
  }
  function clearInputById(id) {
	var x = document.getElementById(id);
	x.value = "";
  }
  function clearSearchInput() {
	clearInputById("id_labels");
	clearInputById("id_author");
	clearInputById("id_location");
	clearInputById("id_keywords");
	clearInputById("id_comment");
  }
  function addToInputValue(text, id) {
	var x = document.getElementById(id);
	x.value += text;
  }
  function toggleHideDisplay(myDIV) {
	var x = document.getElementById(myDIV);
	if (x.style.display === "none") {
	  x.style.display = "block";
    } else {
	  x.style.display = "none";
	}
  }
  function mountLabel(label) {
	mountWindow = window.open("/mount?label="+label, "_blank", "location=no,toolbar=no,scrollbars=yes,resizable=yes,menubar=no,top=300,left=500,width=800,height=400");
  }
  function mountClose() {
	window.close();
  }
  function mountLabelByBrowser(label) {
	console.log("mount label 1: "+label)
	//alert("mount label: "+label)
	globalLabel = label
	document.getElementById('fileselector').click()
  }
  function mountLabelSet(e) {
	var fileselector = document.getElementById('fileselector');
	console.log("mount label 2 "+globalLabel+":"+fileselector.value)
	alert("mount label 2 "+globalLabel+":"+fileselector.value)
  }
  function tBtnSave() {
	var eA = document.getElementById( 'tInfi'+'Author');
	var eL = document.getElementById( 'tInfi'+'Location');
	var eK = document.getElementById( 'tInfi'+'Keywords');
	var eC = document.getElementById( 'tInfi'+'Comment');
	var form = document.getElementById("UpdateDirForm");
	form.method = "PUT";
	alert("Author:"+eA.value+", Location:"+eL.value+", form action="+form.action+",method="+form.method)
	console.log("Author:"+eA.value+", Location:"+eL.value+", form action="+form.action)
	form.submit();
  }
  function tBtn(name) {
		var id = 'tInfs'+name;
	var es = document.getElementById(id);
	if (es.style.display == "none") {
		es.style.display = "inline"
	} else {
		es.style.display = "none"
	}

	id = 'tInfi'+name;
	var ei = document.getElementById(id);
	if (ei.style.display == "none") {
		ei.style.display = "inline"
	} else {
		ei.style.display = "none"
	}
	es.innerHTML = ei.value;
  }
 </script>
</head>
<body style="text-align:left"><a name="top"></a>
<input id="fileselector" type="file" onchange="mountLabelSet(event)" webkitdirectory directory multiple="false" style="display:none" />
`
	return fmt.Sprintf(htmlPage, title, jsFiles)
}
func webSearch0(labelsOnly bool) string {
	format := `&nbsp; <form action="/search" method="get" style="display: inline;">
	<span>
		%s
		%s
		<br/>
	    <button type="button" onclick="clearSearchInput()">clear</button>
		<input type="submit" value="search &#x1F50D;" title="search for pictures and files"/>
	</span>
	</form>`
	e := ""
	a1 := fmt.Sprintf(`Labels <input type="text" id="search" name="search" value="%s" size="100" /><br/>`, glob_searchLabels)
	a4 := fmt.Sprintf(`		Author <input type="text" id="id_author" name="na_author" value="%s" size="100" /><br/>
	Location <input type="text" id="id_location" name="na_location" value="%s" size="100" /><br/>
	Keywords <input type="text" id="id_keywords" name="na_keywords" value="%s" size="100" /><br/>
	Comment <input type="text" id="id_comment" name="na_comment" value="%s" size="100" />
`, e, e, e, e)
	s := fmt.Sprintf(format, a1, "")
	if !labelsOnly {
		s = fmt.Sprintf(format, a1, a4)
	}
	return s
}
func x60(s string) string {
	if len(s) > 60 {
		return s[:60]
	}
	return s
}
func webSearch(labels string) string {
	format := `<form action="/search" method="get">
	<div class="row">
	  <div class="column left">Labels</div><div class="column middle"><input type="checkbox" id="ch_labels" onclick="checkShowHide('ch_labels','list_labels')"></div>
	  <div class="column right"><input type="text" id="id_labels" name="na_labels" size="75%%" value="%s"></div>
    </div>
	<div id="list_labels" style="display:none">%s</div>
	
  
  <div id="i4" style="display:%s">
    <div class="row">
	  <div class="column left">Author</div><div class="column middle"><input type="checkbox" id="ch_author" onclick="checkShowHide('ch_author','list_author')"></div>
	  <div class="column right"><input type="text" id="id_author" name="na_author" size="75%%"></div>
    </div>
	<div id="list_author" style="display:none">%s</div>
    <div class="row">
	  <div class="column left">Location</div><div class="column middle"><input type="checkbox" id="ch_location" onclick="checkShowHide('ch_location','list_location')"></div>
	  <div class="column right"><input type="text" id="id_location" name="na_location" size="75%%"></div>
    </div>
	<div id="list_location" style="display:none">%s</div>
    <div class="row">
	  <div class="column left">Keywords</div><div class="column middle"><input type="checkbox" id="ch_keywords" onclick="checkShowHide('ch_keywords','list_keywords')"></div>
	  <div class="column right"><input type="text" id="id_keywords" name="na_keywords" size="75%%"></div>
    </div>
	<div id="list_keywords" style="display:none">%s</div>
    <div class="row">
	  <div class="column left">Comment</div><div class="column middle"><input type="checkbox" id="ch_comment" onclick="checkShowHide('ch_comment','list_comment')"></div>
	  <div class="column right"><input type="text" id="id_comment" name="na_comment" size="75%%"></div>
    </div>
	<div id="list_comment" style="display:none">%s</div>
	</div>
    <div class="row">
	  <div class="column left">&nbsp;</div><div class="column middle"><button type="button" onclick="clearSearchInput()">clear</button></div>
	  <div class="column right"><input type="submit" value="search &#x1F50D;" title="search for pictures and files"/></div>
    </div>
  </form>
  `
	occL := occurenciesLabels()
	listA := "none"
	listL := "none"
	listK := "none"
	listC := "none"
	label := labels
	fmt.Printf("%s-webSearch label='%s', labels=%s\n", time.Now().Format("15:04:05"), label, labels)
	db, err := LabeledDatabase(label)
	if err == nil {
		m := ArcpicsMostOcurrenceStrings(db)
		listA = occurenciesList("id_author", m["author"])
		listL = occurenciesList("id_location", m["location"])
		listK = occurenciesList("id_keywords", m["keywords"])
		listC = occurenciesList("id_comment", m["comment"])
		defer db.Close()
	}
	fmt.Printf("%s-lists:\nA:%s\n L:%s\n K:%s\n C:%s\n", time.Now().Format("15:04:05"), x60(listA), x60(listL), x60(listK), x60(listC))
	if labels == "" {
		return fmt.Sprintf(format, glob_searchLabels, occL, "none", listA, listL, listK, listC)
	} else {
		return fmt.Sprintf(format, glob_searchLabels, occL, "block", listA, listL, listK, listC)
	}
}
func webMenu(link string) string {
	items := []struct {
		L string
		N string
	}{
		{"/", "Home"},
		{"/labels", "Labels"},
		{"/search", "Search"},
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
func occurenciesLabels() string {
	dbDir := GetDatabaseDirName()
	labels, err := GetLabelsInDbDir(dbDir)
	if err != nil {
		return fmt.Sprintf(" %s<br/>\n", err.Error())
	}
	if len(labels) < 1 {
		return " No labels<br/>\n"
	} else {
		s := ""
		for _, k := range labels {
			s += fmt.Sprintf(" <button type=\"button\" onclick=\"addToInputValue('%s','id_labels')\">%s</button> ", k, k)
		}
		return s
	}
}
func occurenciesList(id string, m map[string]int) string {
	s := ""
	for _, k := range aMapKeysSortedByFreguency(m) {
		n := m[k]
		apostrofedK := k
		if strings.Contains(k, " ") {
			apostrofedK = "\\'" + k + "\\'"
		}
		s += fmt.Sprintf(" <button type=\"button\" onclick=\"addToInputValue('%s','%s')\">%s</button>:%d ", apostrofedK, id, k, n)
	}
	s += "\n"
	return s
}

func occurenciesArr(w http.ResponseWriter, name string) {
	nameColon := strings.ToUpper(name) + ":"
	fmt.Fprintf(w, "\n <button type=\"button\" onclick=\"addToInputValue('%s','id_labels')\"><b>%s:</b></button>\n", nameColon, name)
	dbDir := GetDatabaseDirName()
	labels, err := GetLabelsInDbDir(dbDir)
	if err != nil {
		fmt.Fprintf(w, " %s<br/>\n", err.Error())
		return
	}
	if len(labels) < 1 {
		fmt.Fprint(w, " No labels<br/>\n")
		return
	} else {
		for _, k := range labels {
			fmt.Fprintf(w, " <button type=\"button\" onclick=\"addToInputValue('%s', 'id_labels')\">%s</button> ", k, k)
		}
		fmt.Fprint(w, "\n<br/>\n<form action=\"./\" >")
		for _, k := range labels {
			fmt.Fprintf(w, ` <input type="radio" id="%s" name="select_label" value="%s">%s `, k, k, k)
		}
		fmt.Fprint(w, "\n</form>\n")

	}
	fmt.Fprint(w, "\n<br/><br/>\n")
}
func occurenciesMap(w http.ResponseWriter, name string, m map[string]int) {
	nameColon := strings.ToUpper(name) + ":"
	fmt.Fprintf(w, "\n <button type=\"button\" onclick=\"addToInputValue('%s','id_%s')\"><b>%s:</b></button>\n", nameColon, name, name)
	for _, k := range aMapKeysSortedByFreguency(m) {
		n := m[k]
		apostrofedK := k
		if strings.Contains(k, " ") {
			apostrofedK = "\\'" + k + "\\'"
		}
		fmt.Fprintf(w, " <button type=\"button\" onclick=\"addToInputValue('%s','id_%s')\">%s</button>:%d ", apostrofedK, name, k, n)
	}
	fmt.Fprint(w, "\n<br/><br/>\n")
}
func aMapKeysSortedByFreguency(wordsCount map[string]int) []string {
	keys := make([]string, 0, len(wordsCount))

	for key := range wordsCount {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return wordsCount[keys[i]] > wordsCount[keys[j]]
	})
	return keys
}
func getSearchLabels(query string) string {
	m := getParams(`(?m).*LABELS:(?P<Labels>[\w,]+)`, query)
	return m["Labels"]
}
func pageHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageBeginning("Arcpics home", ""))
	fmt.Fprint(w, webMenu("/"))
	fmt.Fprint(w, "<h1>Arcpics - labeled external archives management for pictures and another files</h1>\n")
	imgSrc := "https://lh3.googleusercontent.com/3O6QhFBHS9tb6U-Guk_cGUBYQTKmzBBd8Z3ldi7fK4d-gRiiQxYPvnCmKnDTB3IU86MbY0yw-jIaZhg62vaTajouUYlhqG0AzoSV-UJ74o4ewdWtB51CQKlgTaKQjTftDJiUV-tBa8j3cuPL4XlbHagABnBzTu4JdJ5P8FHgn7TqJSRYENYwAZIg60NQmi6OnGqDSElr30Aeghz7_aQ9mAqtb5c4aqxfiJSTuqbZQuwvtBDMifRCoS2WX9jgyc2W2hh7gFYGOn9xLKVvnns9q5xlg97QSYTmkpQW3DYe7QgQc7uKJe_6yQ-xG2Nw7F5k4yxsZiLJ9EA6vhmIL3OAuu9dbi9AxhYKlZ3yVhjncZajM2e3qdGoFFMnc8klBx191xqFyOEDpgX6c2YCmn_SJ-HcOmlo-v_1Uk5f7Mre4_-nK4BOlSHjmx7Ojur3ERpnLbWgQNJ09Sz6-uo58oBW-V8cIgwCes9fo97PBVzUQjBeoncZ2sa76sR_AQQfKOl2nnsQ2Ez6UI73r1S__A6cDRQhy4cFOIGg9P4FbUrx0UDJAoDswfD7h3w3tl6ASY2FB8ogyDrDEfnrh-XItzQL-VU21uuSiCaQYXTpetgXSAr-jifsTd4Xh5eJ2iiL3rCH21aJh_Gl_7pYfl2g6P82T0xdt-r2hXT6CFp5dvpBZGA1jQ1ZpoKwDN4J0n9NCBfnuZscn5Wopst3ABKF94NBMI-nV1bSye-zLOjNC0qUVQWreu8qr1yUy_FkKl6gKcVMGIv3G-rR1NNpD2LGyjln2PqDeviZFTxGCjfhDOjr1t75mGUAkgAk4iC5dYv-S1iCAB8rC1HeY0dT-IZltLljS5oTrBwsXIjyF2_syrnhyiD4srdX370mQoe5Jemoa2rlSKz13cvCJm0GWq2v5YOBzKwE_lVrz405r-x0oZYm58ZWioC4kQ=w1307-h980-s-no?authuser=0"
	fmt.Fprintf(w, "<img src=\"%s\" />\n", imgSrc)
}
func pageSearch(w http.ResponseWriter, r *http.Request) {
	glob_searchLabels = r.URL.Query().Get("na_labels")
	glob_searchAuthor = r.URL.Query().Get("na_author")
	glob_searchLocation = r.URL.Query().Get("na_location")
	glob_searchKeywords = r.URL.Query().Get("na_keywords")
	glob_searchComment = r.URL.Query().Get("na_comment")
	fmt.Fprint(w, pageBeginning("Arcpics search", ""))
	fmt.Fprint(w, webMenu("/search"))
	fmt.Fprint(w, webSearch(glob_searchLabels))
	fmt.Fprintf(w, "<h1>Results</h1>\n<code><b>Labels :</b>%s</code><br/>\n<code><b>Author :</b>%s</code><br/>\n<code><b>Location:</b>%s</code><br/>\n<code><b>Keywords:</b>%s</code><br/>\n<code><b>Comment :</b>%s</code><br/>\n", glob_searchLabels, glob_searchAuthor, glob_searchLocation, glob_searchKeywords, glob_searchComment)

	///////////////
	label := glob_searchLabels // for now
	var keys []string
	db, err := LabeledDatabase(label)
	if err == nil {
		keys = ArcpicsSearchFilesKeys(db, glob_searchAuthor, glob_searchLocation, glob_searchKeywords, glob_searchComment)
		defer db.Close()
	}
	for _, k := range keys {
		link := fmt.Sprintf(`<a href="/label-dir/%s/%s" title="%s">%s</a>`, label, k, k, k+"/")
		fmt.Fprintf(w, "<br/> %s\n", link)
	}

	///////////////
	/*
		fmt.Fprintf(w, "Query: %s<hr>\n", glob_searchLabels)

		occurenciesArr(w, "Labels")
		if reSearchLabels.MatchString(glob_searchLabels) {
			label := getSearchLabels(glob_searchLabels)
			db, err := LabeledDatabase(label)
			if err == nil {
				m := ArcpicsMostOcurrenceStrings(db)
				occurenciesMap(w, "Author", m["author"])
				occurenciesMap(w, "Location", m["location"])
				occurenciesMap(w, "Keywords", m["keywords"])
				occurenciesMap(w, "Comment", m["comment"])
			}
			defer db.Close()
		}
	*/
}
func pageAbout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, pageBeginning("About arcpics", ""))
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
	fmt.Fprint(w, pageBeginning("Arcpics Labels", ""))
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
		labelsString += "\n<pre>"
		for _, label := range labels {
			labelsString += fmt.Sprintf(`%-10s `, label)
			labelsString += fmt.Sprintf(`<a href="/label-list/%s">list</a>%s`, label, " ")
			labelsString += fmt.Sprintf(`<a href="/label-dir/%s/">dir</a>%s`, label, " ")
			labelsString += getSystemLabelSummary(label)
			labelsString += " " + LabelMounts.Html(label)
			labelsString += "\n"
		}
		labelsString += "</pre>\n"
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
		//keys = ArcpicsAllKeys(db, FILES_BUCKET)
		keys = ArcpicsAllKeys(db, SYSTEM_BUCKET)
	}
	defer db.Close()
	fmt.Fprint(w, pageBeginning("Arcpics Label "+label+" list", ""))
	fmt.Fprint(w, webMenu(""))
	lblfmt := "<h1>Arcpics Label %s list</h1>\n"
	fmt.Fprintf(w, lblfmt, label)
	//nodes := makeNodes(keys)
	for _, k := range keys {
		link := fmt.Sprintf(`<a href="/label-dir/%s/%s" title="%s">%s</a>`, label, k, k, k+"/")
		fmt.Fprintf(w, "<br/> %s\n", link)
	}
}

func pageImageFile(w http.ResponseWriter, r *http.Request) {
	params := getParams(reImageFileStr, r.URL.Path)
	fname := params["Fname"]
	if strings.Contains(fname, ":") && strings.HasPrefix(fname, "/") {
		fname = fname[1:]
	}
	f, err := os.Open(fname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>%d - File not found: %s - %s</h1>", http.StatusNotFound, fname, err.Error())
		return
	}
	defer f.Close()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	buf := make([]byte, 64*1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		if n > 0 {
			w.Write(buf)
		}
	}
}
func getLabelPathJdir(label, path string) (JdirType, error) {
	var jd JdirType
	if path == "" {
		path = "./"
	}
	db, err := LabeledDatabase(label)
	var val string
	if err == nil {
		val = GetDbValue(db, FILES_BUCKET, path)
	} else {
		return jd, err
	}
	defer db.Close()
	err = json.Unmarshal([]byte(val), &jd)
	return jd, err
}
func indexPrev(i, length int) int {
	i -= 1
	if i < 0 {
		i = length - 1
	}
	return i
}
func indexNext(i, length int) int {
	i += 1
	if i >= length {
		i = 0
	}
	return i
}
func pageImageIndex(w http.ResponseWriter, r *http.Request) {
	params := getParams(reImageIndexStr, r.URL.Path)
	dir := params["Dir"]
	file := params["File"]
	filesStr := r.URL.Query().Get("files")
	to_stash := r.URL.Query().Get("to_stash")
	if to_stash == "true" {
		copyToStash(dir, file)
	}
	files := strings.Split(filesStr, ",")
	prev := file
	next := file
	ok := false
	for i, f := range files {
		if f == file {
			prev = files[indexPrev(i, len(files))]
			next = files[indexNext(i, len(files))]
			ok = true
			break
		}
	}
	filesStr = "files=" + filesStr
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<h1>%d - File not found: %s</h1>", http.StatusNotFound, file)
		return
	}
	linkFmt := `/imageindex/%s/%s?%s`
	linkPrev := fmt.Sprintf(linkFmt, dir, prev, filesStr)
	linkCopy := fmt.Sprintf(linkFmt, dir, file, "to_stash=true&"+filesStr)
	linkNext := fmt.Sprintf(linkFmt, dir, next, filesStr)
	pageStr := `<html><body><map name="workmap">
	<area shape="rect" coords="0,0,200,800"   alt="prev" title="previous picture" href="%s">
	<area shape="circle" coords="600,50,50"   alt="copy" title="copy to stash" href="%s">
	<area shape="rect" coords="1000,0,1200,800" alt="next" title="next picture" href="%s">
  </map>
 <img src="/image/%s" usemap="#workmap" width="1200" height="800" >
 </body></html>
 `
	fmt.Fprintf(w, pageStr, linkPrev, linkCopy, linkNext, dir+"/"+file)
}

func pageLabelFileJpeg(w http.ResponseWriter, r *http.Request) {
	params := getParams(reLabelFileJpegStr, r.URL.Path)
	label := params["Label"]
	path := params["Path"]
	jpgFile := params["JpgFile"]

	ok := false
	jd, err := getLabelPathJdir(label, path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "<h1>Internal server error: %s</h1>", err.Error())
	} else {
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
				prevLink = fmt.Sprintf(fmtStr, jd.Dirs[i-1], "&#x21D0; - ")
			}
			if i+1 < len(jd.Dirs) {
				nextLink = fmt.Sprintf(fmtStr, jd.Dirs[i+1], " - &#x21D2;")
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
func tBtnSave() string {
	return `<input type="submit" value="Save" xxonclick="tBtnSave('')" title="Press to edit"/>`
}
func tBtn(name string) string {
	return fmt.Sprintf(`<input type="button" value="%s" onclick="tBtn('%s')" title="Press to edit"/>`, name, name)
}
func tInf(name, value string) string {
	return fmt.Sprintf(`<span id="tInfs%s">%s</span><input type="text" value="%s" id="tInfi%s" name="%s" style="display:none;background-color:#FFFACD;"/>`, name, value, value, name, name)
}
func infoTable(inf JinfoType, urlPath string) string {
	if inf.Author == "" && inf.Location == "" && inf.Keywords == "" && inf.Comment == "" {
		return "\n<!--no form, no table-->\n"
	}
	f := `<form id="UpdateDirForm" action="%s" method="post">
<table bgcolor="gray"><caption>Directory info &nbsp; &nbsp; %s</caption>
 <tbody>
 <tr bgcolor="white"><th>%s</th><th>%s</th><th>%s</th><th>%s</th></tr>
 <tr bgcolor="white"><td><span>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>
 </tbody>
</table>
</form>`
	A := "Author"
	L := "Location"
	K := "Keywords"
	C := "Comment"
	return fmt.Sprintf(f,
		urlPath,
		tBtnSave(),
		tBtn(A), tBtn(L), tBtn(K), tBtn(C),
		tInf(A, inf.Author), tInf(L, inf.Location), tInf(K, inf.Keywords), tInf(C, inf.Comment))
}
func jsConstFiles(jd JdirType) string {
	//	s := `const files = ["Saab", "Volvo", "BMW"];`
	s := "const files = ["
	comma := ""
	for _, f := range jd.Files {
		s += comma + "\"" + f.Name + "\""
		comma = ","
	}
	s += "];\n"
	return s
}

func pageLabelDir(w http.ResponseWriter, r *http.Request) {
	params := getParams(`\/label-dir\/(?P<Label>[a-zA-z0-9]+)\/(?P<Path>.*)`, r.URL.Path)
	label := params["Label"]
	mountDir := LabelMounts.Get(label)
	path := params["Path"]
	if path == "" {
		path = "./"
	}
	var inf JinfoType
	fmtErr := "<h2>Arcpics Label: %s</h2>\npath: %s <hr/>\nerror: %s\n"
	// parsing request body needs to be done before timeout
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, fmtErr, label, path, err.Error())
			return
		}
		inf.Author = r.FormValue("Author")
		inf.Location = r.FormValue("Location")
		inf.Keywords = r.FormValue("Keywords")
		inf.Comment = r.FormValue("Comment")
	}

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
	var jd JdirType

	if err = json.Unmarshal([]byte(val), &jd); err != nil {
		fmt.Fprint(w, pageBeginning("Arcpics Label "+label, ""))
		fmt.Fprint(w, webMenu(""))
		fmt.Fprintf(w, fmtErr, label, path, err.Error())
		return
	}
	fmt.Fprint(w, pageBeginning("Arcpics Label "+label, jsConstFiles(jd)))
	fmt.Fprint(w, webMenu(""))
	if r.Method == http.MethodPost {
		//update changed dir info into the same file info
		for _, f := range jd.Files {
			if jInfoIsEqual(jd.Info, f.Info) {
				f.Info = inf
			}
		}
		jd.Info = inf
		jd.ByUser = true
		if err = PutDbValueHttpReqDir(db, FILES_BUCKET, path, &jd); err != nil {
			fmt.Fprintf(w, fmtErr, label, path, err.Error())
			return
		}
	}

	linkPrev, linkNext := prevNextPathLinks(parentVal, lastDir(path))
	lblfmt := "<h1>Arcpics Label: %s</h1>\n%s(path: %s)%s<hr/>\n"
	fmt.Fprintf(w, lblfmt, label, linkPrev, path, linkNext)
	fmt.Fprint(w, infoTable(jd.Info, r.URL.Path))
	fmt.Fprint(w, "<button type=\"button\" onclick=\"toggleHideDisplay('idFiles')\">Hide/Display Files</button>")

	head := "<pre id=\"idFiles\">%33s %55s %45s %s\n"
	fmt.Fprintf(w, head, `<a href="?C=N;O=D">Name</a>`, `<a href="?C=M;O=A">Last modified</a>`, `<a href="?C=S;O=A">Size</a>`, `<a href="?C=D;O=A">Description</a>`)
	parentDir, _ := getParentDir(path)
	if path != "./" {
		fmt.Fprintf(w, "      <a href=\"/label-dir/%s/%s\">%s</a>\n", label, parentDir, "parent directory")
	}
	for _, f := range jd.Files {
		fixNameLink(w, label, path, f.Name, false)
		fmt.Fprintf(w, "%-26s%10s %s\n", f.Time, f.Size, f.Info.Comment)
	}
	for _, d := range jd.Dirs {
		fixNameLink(w, label, path+"/"+d, d, true)
	}

	//Thumbnail images
	fmt.Fprint(w, "\n</pre><hr/>\n")
	filesStr := ""
	for _, f := range jd.Files {
		if isJpegFile(f.Name) {
			if filesStr != "" {
				filesStr += ","
			}
			filesStr += f.Name
		}
	}
	for _, f := range jd.Files {
		if isJpegFile(f.Name) {
			title := f.Name + ": " + f.Info.Comment
			if mountDir != "" {
				title += " - click for full picture"
			}
			img := fmt.Sprintf(`<img src="/label-dir/%s/%s/%s" title="%s"/>`, label, path, f.Name, title)
			if mountDir != "" {
				//v1 fmt.Fprintf(w, `<a href="/image/%s/%s/%s">%s</a>%s`, mountDir, path, f.Name, img, "\n")
				//v2 title := "Title: " + f.Name + ": " + f.Info.Author + "|" + f.Info.Keywords + "|" + f.Info.Location + "|" + f.Info.Comment
				//v2 fmt.Fprintf(w, `<span onclick="openWin('/image/%s/%s/%s','%s')">%s</span>%s`, mountDir, path, f.Name, title, img, "\n")
				fmt.Fprintf(w, `<span onclick="openWin('/imageindex/%s/%s/%s?files=%s','%s')">%s</span>%s`, mountDir, path, f.Name, filesStr, title, img, "\n")

			} else {
				fmt.Fprintf(w, `%s%s`, img, "\n")
			}
		}
	}
	fmt.Fprint(w, "\n<br/>\n")
	fmt.Fprintf(w, "<hr/>Label: %s\n%s(path: %s)%s - %s\n", label, linkPrev, path, linkNext, " <a href=\"#top\">&#x21D1; top</a>")
}
func WinAvailableLetterDrives() (letters []string) {
	for ch := 'A'; ch <= 'Z'; ch++ {
		d := string(ch) + ":"
		if _, err := os.Stat(d + "\\"); err == nil {
			letters = append(letters, d)
		}
	}
	return letters
}

func RootMountDirs() []string {
	s := make([]string, 0)
	user := os.Getenv("USER")
	if runtime.GOOS == "windows" {
		s = WinAvailableLetterDrives()
	} else {
		s = append(s, "/")
		s = append(s, "/media/"+user)
	}
	s = append(s, os.Getenv("HOME"))
	return s
}

// func similar to getLabel maybe should be found common ground
func findLabel(dir string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), defaultNameDashLabelDot) {
			return f.Name()[len(defaultNameDashLabelDot):]
		}
	}
	return ""
}
func pageMount(w http.ResponseWriter, r *http.Request) {
	label := r.URL.Query().Get("label")
	localdir := r.URL.Query().Get("localdir")
	mountdir := r.URL.Query().Get("mountdir")
	fmt.Fprint(w, pageBeginning(fmt.Sprintf("Arcpics: Mount Label %s", label), ""))
	fmt.Fprintf(w, `<div class="column c3left"><input type="button" value="Close" onclick="mountClose();"/></div>
<div class="column c3middle">Manage root mount points for external labels %s</div>
<div class="column c3right"><input type="button" value="mount"/> </div>
<hr/>
`, label)

	fmt.Fprintf(w, `<div class="column c2left" style="background-color: lightgray">%s`, "\n")
	for _, d := range RootMountDirs() {
		fmt.Fprintf(w, `<a href="/mount?label=%s&localdir=%s">%s</a><br/>%s`, label, d, d, "\n")
	}
	fmt.Fprintf(w, `%s</div>`, "\n")

	fmt.Fprintf(w, `<div class="column c2right">%s`, "\n")
	if mountdir != "" {
		LabelMounts[label] = mountdir
		db, err := LabeledDatabase(label)
		if err == nil {
			insert2MountDirNow(db, mountdir)
			db.Close()
		}
		fmt.Fprintf(w, `<h2>Directory %s has been mounted to label %s</h2>`, mountdir, label)
		return
	}
	if localdir == "" {
		fmt.Fprintf(w, `</div>%s`, "\n")
		return
	}
	dirs, err := ioutil.ReadDir(localdir)
	if err != nil {
		fmt.Fprintf(w, `%s`, err.Error())
	} else {
		sep := "/"
		if runtime.GOOS == "windows" {
			sep = "\\"
		}
		pDir := localdir + sep + ".."
		fmt.Fprintf(w, `<a href="/mount?label=%s&localdir=%s">%s</a><br/>%s`, label, pDir, "..", "\n")
		for _, d := range dirs {
			if d.IsDir() {
				dir := localdir + sep + d.Name()
				foundLabel := findLabel(dir)
				if foundLabel == "" {
					fmt.Fprintf(w, `<a href="/mount?label=%s&localdir=%s">&#x1F4C1;</a> %s<br/>%s`, label, dir, d.Name(), "\n")
				} else {
					if label == foundLabel {
						fmt.Fprintf(w, `<span style="background-color: lightgreen">&#x1F4C1; %s - label: %s <a href="/mount?label=%s&mountdir=%s"><b>mount</b></a></span><br/>%s`, d.Name(), label, label, dir, "\n")
					} else {
						fmt.Fprintf(w, `<span style="background-color: lightgreen"><a href="/mount?label=%s&localdir=%s">&#x1F4C1;</a> %s - label: %s</span><br/>%s`, label, dir, d.Name(), foundLabel, "\n")
					}
				}
			}
		}
	}
	fmt.Fprintf(w, `%s</div>`, "\n")

}

// find for all labels if previously mounted directory is available or defined from CLI by -m option
// and record it into db
func ConnectLabelsToMountDirs() {
	dbDir := GetDatabaseDirName()
	labels, err := GetLabelsInDbDir(dbDir)
	if err != nil {
		fmt.Printf("ConnectLabelsToMountDirs: %s\n", err.Error())
		return
	}
	for _, label := range labels {
		dir := LabelMounts[label]
		db, err := LabeledDatabase(label)
		if err != nil {
			continue
		}
		if dir != "" {
			insert2MountDirNow(db, dir)
		} else {
			lastDir := lastMountedDir(db)
			if lastDir != "" {
				if DirExists(lastDir) {
					LabelMounts[label] = lastDir
					insert2MountDirNow(db, lastDir)
				}
			}
		}
		db.Close()

	}

}
func Web(port int) {
	ConnectLabelsToMountDirs()
	http.HandleFunc("/", route)

	colonPort := fmt.Sprintf(":%d", port)
	fmt.Printf("... listening at port %d", port)
	log.Fatal(http.ListenAndServe(colonPort, nil))

}
