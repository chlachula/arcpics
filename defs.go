package arcpics

/*
line:
 duplicate Windows — Shift + Alt + Up/Down
 select    Windows/Ubuntu — Ctrl + L
 delete    Windows/Ubuntu — Ctrl + Shift + K
 move Windows/Ubuntu — Alt + Up/Down arrow
*/

var defaultArcpicsDbLabel = "arcpics-db-label."
var defaultPicturesDirName = "Arc-Pics"
var defaultDatabaseDirName = "DB"

var SYSTEM_BUCKET = []byte("SYSTEM")
var INIT_LABEL_KEY = "ARC-PICS-LABEL-KEY"
var FILES_BUCKET = []byte("FILES")

var jsonFilePrefix = "sample.json"
var jsonUserData = "arcpics-user-data.json"
var timeStampJsonFormat = "2006-01-02_15:04:05.99"

type JfileType = struct {
	Name    string
	Size    string
	Time    string //
	Comment string `json:",omitempty"`
}
type JdirType = struct {
	Description string      `json:",omitempty"`
	Location    string      `json:",omitempty"`
	Files       []JfileType `json:",omitempty"`
}
