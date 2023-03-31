package arcpics

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
)

func TestDbLabel0(t *testing.T) {
	want := "there is no file arcpics-db-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-0") + " - should be e.g. arcpics-db-label.001"
	picturesDirName := filepath.Join("example", "Arc-Pics-wrong-0")
	_, err := DbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}
func TestDbLabel2(t *testing.T) {
	want := "unexpected number files arcpics-db-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-2")
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
	wantDbDir := defaultDatabaseDirName
	args := make([]string, 1)
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
	wantDbDir := defaultDatabaseDirName
	args := make([]string, 2)
	args[1] = wantPicDir
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
	args := make([]string, 3)
	args[1] = wantPicDir
	args[2] = wantDbDir
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
	args := make([]string, 3)
	args[0] = "program-name-will-be-removed-anyway"
	args[1] = wantPicDir
	args[2] = wantDbDir
	var gotDb *bolt.DB
	var gotPicDir string
	var gotErr error

	gotPicDir, gotDb, gotErr = AssignPicturesDirectoryWithDatabase(args)
	if wantPicDir != gotPicDir {
		t.Errorf("error - wantPicDir: %s; gotPicDir: %s", wantPicDir, gotPicDir)
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