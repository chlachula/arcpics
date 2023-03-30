package arcpics

import (
	"path/filepath"
	"testing"
)

func TestDbLabel0(t *testing.T) {
	want := "xthere is no file arcpics-db-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-0") + " - should be e.g. arcpics-db-label.001"
	picturesDirName := filepath.Join("example", "Arc-Pics-wrong-0")
	_, err := DbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}
func TestDbLabel2(t *testing.T) {
	want := "xunexpected number files arcpics-db-label.* at directory " + filepath.Join("example", "Arc-Pics-wrong-2")
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
