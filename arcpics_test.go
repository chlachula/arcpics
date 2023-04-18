package arcpics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

func TestDbLabel0(t *testing.T) {
	want := "there is no file arcpics-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-0") + " - should be e.g. arcpics-label.001"
	picturesDirName := filepath.Join("example", "Arc-Pics-wrong-0")
	_, err := DbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}
func TestDbLabel2(t *testing.T) {
	want := "unexpected number files arcpics-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-2")
	picturesDirName := filepath.Join("example", "Arc-Pics-wrong-2")
	_, err := DbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}

func TestDbLabel1(t *testing.T) {
	want := "a"
	picturesDirName := filepath.Join("example", "Arc-Pics-good-1")
	label, _ := DbLabel(picturesDirName)
	got := label
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}

func TestPicturesAndDatabaseDirectories0(t *testing.T) {
	wantPicDir := defaultPicturesDirName
	wantDbDir := GetDatabaseDirName()
	args := make([]string, 0)
	gotPicDir, gotDbDir := picturesAndDatabaseDirectories(args)
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantPicDir: %s; gotPicDir: %s", wantPicDir, gotPicDir)
	}
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantDbDir: %s; gotDbDir: %s", wantDbDir, gotDbDir)
	}
}
func TestPicturesAndDatabaseDirectories1(t *testing.T) {
	wantPicDir := "ABCD"
	wantDbDir := GetDatabaseDirName()
	args := make([]string, 1)
	args[0] = wantPicDir
	gotPicDir, gotDbDir := picturesAndDatabaseDirectories(args)
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantPicDir: %s; gotPicDir: %s", wantPicDir, gotPicDir)
	}
	if wantDbDir != gotDbDir {
		t.Errorf("error - wantDbDir: %s; gotDbDir: %s", wantDbDir, gotDbDir)
	}
}
func TestPicturesAndDatabaseDirectories2(t *testing.T) {
	wantPicDir := "ABCD"
	wantDbDir := "XYZ"
	args := make([]string, 2)
	args[0] = wantPicDir
	args[1] = wantDbDir
	gotPicDir, gotDbDir := picturesAndDatabaseDirectories(args)
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantPicDir: %s; gotPicDir: %s", wantPicDir, gotPicDir)
	}
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantDbDir: %s; gotDbDir: %s", wantDbDir, gotDbDir)
	}
}

func TestAssignPicturesDirectoryWithDatabase(t *testing.T) {
	wantPicDir := filepath.Join("example", defaultPicturesDirName)
	wantDbDir, err := os.MkdirTemp("", "test-arcpics-bolt-*")
	defer os.RemoveAll(wantDbDir)
	if err != nil {
		t.Errorf("error creating temp dir: " + err.Error())
	}
	args := make([]string, 2)
	args[0] = wantPicDir
	args[1] = wantDbDir
	var gotDb *bolt.DB
	var gotArcFS ArcpicsFS
	var gotErr error

	gotArcFS, gotDb, gotErr = AssignPicturesDirectoryWithDatabase(args[0], args[1])
	if wantPicDir != gotArcFS.Dir {
		t.Errorf("error - wantPicDir: %s; gotPicDir: %s", wantPicDir, gotArcFS.Dir)
	}
	if gotDb == nil {
		t.Errorf("error - gotDb *bolt.DB is nil")
	}
	if gotErr != nil {
		t.Errorf(gotErr.Error())
	}

	wantLabel, err := DbLabel(wantPicDir)
	if err != nil {
		t.Errorf("error getting label: " + err.Error())
	}

	gotLabel := GetDbValue(gotDb, SYSTEM_BUCKET, INIT_LABEL_KEY)
	if wantLabel != gotLabel {
		t.Errorf("error - wantLabel: %s; gotLabel: %s", wantLabel, gotLabel)
	}

	if err := gotDb.Close(); err != nil {
		t.Errorf("error - closing bolt DB: " + err.Error())
	}
}

func TestFilesCount(t *testing.T) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	fs, err := OpenArcpicsFS(picDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	wantCount := 19 // find example/Arc-Pics | wc - 19, including directories
	gotCount := FilesCount(fs)
	if wantCount != gotCount {
		t.Errorf("error - wantCount: %d; gotCount: %d", wantCount, gotCount)
	}
}
func TestDirFilesCount(t *testing.T) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	fs, err := OpenArcpicsFS(picDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	wantDirCount := 7    // find example/Arc-Pics -type d | wc .. 7 directories
	wantFilesCount := 12 // find example/Arc-Pics -type f | wc .. 12 files
	gotDirCount, gotFilesCount := DirFilesCount(fs)
	if wantDirCount != gotDirCount {
		t.Errorf("error - wantDirCount: %d; gotDirCount: %d", wantDirCount, gotDirCount)
	}
	if wantFilesCount != gotFilesCount {
		t.Errorf("error - wantFilesCount: %d; gotFilesCount: %d", wantFilesCount, gotFilesCount)
	}
}
func TestDirCount(t *testing.T) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	fs, err := OpenArcpicsFS(picDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	wantDirCount := 7 // find example/Arc-Pics -type d | wc .. 7 directories
	gotDirCount, totalPathLength := DirCount(fs)
	t.Log("Root of Dirs: ", picDir)
	t.Log("Dirs count:", gotDirCount, "- total path lenght:", totalPathLength)
	if wantDirCount != gotDirCount {
		t.Errorf("error - wantDirCount: %d; gotDirCount: %d totalPathLenght:%d", wantDirCount, gotDirCount, totalPathLength)
	}
}

func TestDirPaths(t *testing.T) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	afs, err := OpenArcpicsFS(picDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	wantLenPaths := 7 // find example/Arc-Pics -type d | wc .. 7 directories
	gotPaths, err := afs.DirPaths()
	if err != nil {
		t.Errorf("error - DirPaths: %s", err.Error())
	}
	gotLenPaths := len(gotPaths)
	t.Log("Root of Dirs: ", picDir)
	if wantLenPaths != gotLenPaths {
		t.Errorf("error - wantLenPaths: %d; gotLenPaths: %d", wantLenPaths, gotLenPaths)
	}
}

func TestDirPathsUpdate(t *testing.T) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	afs, err := OpenArcpicsFS(picDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	err = afs.DirPathsUpdate()
	if err != nil {
		t.Errorf("error - DirPathsUpdate: %s", err.Error())
	}
}

// go test -run TestArcpicsFilesUpdate
func TestArcpicsFilesUpdate(t *testing.T) {
	//arcDir := filepath.Join("example", defaultPicturesDirName)
	arcDir, err := makeAndPopulateTempPicDir()
	if err != nil {
		t.Errorf("error - makeAndPopulateTempPicDir: " + err.Error())
	}

	defer os.RemoveAll(arcDir)

	fs, err := OpenArcpicsFS(arcDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	err = ArcpicsFilesUpdate(fs.Dir)
	if err != nil {
		t.Errorf("error - ArcpicsFilesUpdate: " + err.Error())
	}
}

// go test -run TestPurgeJson__
func TestPurgeJson__(t *testing.T) {
	arcDir := filepath.Join("example", defaultPicturesDirName)
	arcFS, err := OpenArcpicsFS(arcDir)
	if err != nil {
		t.Errorf("error - ArcpicsFS: " + err.Error())
	}
	err = PurgeJson__(arcFS.Dir)
	if err != nil {
		t.Errorf("error - Purge: " + err.Error())
	}
}

func makeAndPopulateTempPicDir() (string, error) {
	picDir := filepath.Join("example", defaultPicturesDirName)
	wantPicDir, err := os.MkdirTemp("", "test-arcpics-dir-*")
	if err != nil {
		return wantPicDir, err
	}
	err = CopyDirFromTo(picDir, wantPicDir)
	if err != nil {
		return wantPicDir, err
	}
	return wantPicDir, nil
}

func TestAbsRootPath(t *testing.T) {
	want := "/tmp"
	got := absRootPath("/tmp")
	if want != got {
		t.Errorf("error #1 - wantAbsDir: %s; gotAbsDir: %s", want, got)
	}
	got = absRootPath("/tmp/")
	if want != got {
		t.Errorf("error #2 - wantAbsDir: %s; gotAbsDir: %s", want, got)
	}
}
func TestRelPath(t *testing.T) {
	want := "./"
	got := relPath("/tmp", "/tmp")
	if want != got {
		t.Errorf("error relPath #1a - want: %s; got: %s", want, got)
	}
	got = relPath("/tmp", "/tmp/")
	if want != got {
		t.Errorf("error relPath #1b - want: %s; got: %s", want, got)
	}
	want = "abc"
	got = relPath("/tmp", "/tmp/abc")
	if want != got {
		t.Errorf("error relPath #2a - want: %s; got: %s", want, got)
	}
	got = relPath("/tmp", "/tmp/abc/")
	if want != got {
		t.Errorf("error relPath #2b - want: %s; got: %s", want, got)
	}
	got = relPath("/tmp", "/tmp/./abc")
	if want != got {
		t.Errorf("error relPath #2c - want: %s; got: %s", want, got)
	}
	got = relPath("/tmp", "/tmp/./abc/")
	if want != got {
		t.Errorf("error relPath #2d - want: %s; got: %s", want, got)
	}
}
