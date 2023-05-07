package arcpics

/*
line:

	duplicate Windows — Shift + Alt + Up/Down
	select    Windows/Ubuntu — Ctrl + L
	delete    Windows/Ubuntu — Ctrl + Shift + K
	move Windows/Ubuntu — Alt + Up/Down arrow
*/
var Version string = "0.0.3"
var defaultName = "arcpics"                              // arcpics
var defaultNameDash = defaultName + "-"                  // arcpics-
var defaultNameDashLabel = defaultNameDash + "label"     // arcpics-label
var defaultNameDashLabelDot = defaultNameDashLabel + "." // arcpics-label.
var dotDefaultName = "." + defaultName                   // .arcpics
var defaultPicturesDirName = "Arc-Pics"

var SYSTEM_BUCKET = []byte("SYSTEM")
var FILES_BUCKET = []byte("FILES")

var INIT_LABEL_KEY = "ARC-PICS-LABEL-KEY"

var defaultNameJson = defaultName + ".json"                          // arcpics.json
var defaultNameDashUserDataJson = defaultNameDash + "user-data.json" // arcpics-user-data.json
var timeStampJsonFormat = "2006-01-02_15:04:05.99"

var ErrSkippedByUser = "error - skipped by user"
var Verbose bool = false

type JfileType = struct {
	Name      string
	Size      string
	Time      string //
	Comment   string `json:",omitempty"`
	Thumbnail []byte `json:",omitempty"`
}
type JdirType = struct {
	Description string      `json:",omitempty"`
	MostComment string      `json:",omitempty"`
	Location    string      `json:",omitempty"`
	Skip        []string    `json:",omitempty"`
	Files       []JfileType `json:",omitempty"`
	Dirs        []string    `json:",omitempty"`
}

var HelpTextFmt = `=== arcpics: manage archived of pictures not only at external hard drives ===
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
