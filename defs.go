package arcpics

/*
line:
 duplicate Windows — Shift + Alt + Up/Down
 select    Windows/Ubuntu — Ctrl + L
 delete    Windows/Ubuntu — Ctrl + Shift + K
 move Windows/Ubuntu — Alt + Up/Down arrow
*/

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
	Name    string
	Size    string
	Time    string //
	Comment string `json:",omitempty"`
}
type JdirType = struct {
	Description string      `json:",omitempty"`
	MostComment string      `json:",omitempty"`
	Location    string      `json:",omitempty"`
	Skip        []string    `json:",omitempty"`
	Files       []JfileType `json:",omitempty"`
	Dirs        []string    `json:",omitempty"`
}
