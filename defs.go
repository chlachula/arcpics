package arcpics

/*
line:

	duplicate Windows — Shift + Alt + Up/Down
	select    Windows/Ubuntu — Ctrl + L
	delete    Windows/Ubuntu — Ctrl + Shift + K
	move Windows/Ubuntu — Alt + Up/Down arrow
*/
const (
	Version                 string = "0.0.6"
	defaultName                    = "arcpics"                  // arcpics
	defaultNameDash                = defaultName + "-"          // arcpics-
	defaultNameDashLabel           = defaultNameDash + "label"  // arcpics-label
	defaultNameDashLabelDot        = defaultNameDashLabel + "." // arcpics-label.
	dotDefaultName                 = "." + defaultName          // .arcpics
	defaultPicturesDirName         = "Arc-Pics"
	stashName                      = "stash"

	INIT_LABEL_KEY        = "ARC-PICS-LABEL-KEY"
	LABEL_SUMMARY_fmt     = "LABEL-%s-SUMMARY"
	LABEL_FREQUENCY_WORDS = "LabelFrequencyWords"

	defaultNameJson             = defaultName + ".json"              // arcpics.json
	defaultNameDashUserDataJson = defaultNameDash + "user-data.json" // arcpics-user-data.json
	timeStampJsonFormat         = "2006-01-02_15:04:05.99"

	ErrSkippedByUser = "error - skipped by user"
)

var (
	Verbose             bool            = false
	SYSTEM_BUCKET                       = []byte("SYSTEM")
	FILES_BUCKET                        = []byte("FILES")
	MOUNTS_BUCKET                       = []byte("MOUNTS")
	LabelMounts         LabelMountsType = make(map[string]string)
	glob_searchLabels                   = ""
	glob_searchAuthor                   = ""
	glob_searchLocation                 = ""
	glob_searchKeywords                 = ""
	glob_searchComment                  = ""
)

type (
	// File system ArcpicsFS has to have at root special label file with name "arcpics-db-label"
	// and at least one character long arbitrary extension.
	// For example file "arcpics-db-label.a" has label value "a"
	// or "arcpics-db-label.my1TB_hard_drive" has label value "my1TB_hard_drive"
	//
	// ATTENTION!!
	// ArcpicsFS work fine with fs.WalkDir unless there are any file operations
	// Then use filepath.WalkDir(ArcpicsFS.Dir,...
	ArcpicsFS struct {
		Dir   string
		Label string
	}
)

type JinfoType = struct {
	Author   string `json:",omitempty"`
	Location string `json:",omitempty"`
	Keywords string `json:",omitempty"`
	Comment  string `json:",omitempty"`
}
type JfileType = struct {
	Name      string
	Size      string
	Time      string    //
	Info      JinfoType `json:",omitempty"`
	Thumbnail []byte    `json:",omitempty"`
	ThumbSrc  string    `json:",omitempty"`
}
type JdirType = struct {
	UpdateTime string      `json:",omitempty"`
	ByUser     bool        `json:",omitempty"`
	Info       JinfoType   `json:",omitempty"`
	Skip       []string    `json:",omitempty"`
	Files      []JfileType `json:",omitempty"`
	Dirs       []string    `json:",omitempty"`
}
type FrequencyCounterType map[string]int
type LabelMountsType map[string]string

type Node struct {
	Name  string
	Nodes []Node
}

const HelpTextFmt = `=== arcpics: manage archived of pictures not only at external hard drives ===
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
 -m picturesDirName       #mount original labeled dir for web
 -p port                  #web port definition
 -w                       #start web - default port 8080

Examples:
-c %sArc-Pics Vacation-2023 #creates label file inside of directory %sArc-Pics
-a1 %sArc-Pics               #populates  %sVacation-2023.db by data in subdirs at %sArc-Pics
-af %sArc-Pics               #updates arcpics.json dir files inside of directories at %sArc-Pics
-ab %sArc-Pics               #updates database %sVacation-2023.db
-m %sArc-Pics -w             #mounted one external directory and launch web
`
