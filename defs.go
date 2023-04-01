package arcpics

/*
line:
 duplicate Windows — Shift + Alt + Up/Down
 select    Windows/Ubuntu — Ctrl + L
 delete    Windows/Ubuntu — Ctrl + Shift + K
 move Windows/Ubuntu — Alt + Up/Down arrow
*/

type JfileType = struct {
	Name    string
	Time    string
	Comment string
}
type JdirType = struct {
	Description string
	Location    string
	Files       []JfileType
}

var SYSTEM_BUCKET = []byte("SYSTEM")
var FILES_BUCKET = []byte("FILES")
var INIT_LABEL_KEY = "ARC-PICS-LABEL-KEY"
var defaultPicturesDirName = "Arc-Pics"
var defaultDatabaseDirName = "DB"
var jsonFilePrefix = "sample.json"
var jsonUserData = "arcpics-user-data.json"
