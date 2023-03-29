package main

import (
	"testing"
)

func TestDbLabel0(t *testing.T) {
	want := "there is no file arcpics-db-label.* at directory Arc-Pics-wrong-0 - should be e.g. arcpics-db-label.001"
	picturesDirName := "Arc-Pics-wrong-0"
	_, err := dbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}
func TestDbLabel2(t *testing.T) {
	want := "unexpected number files arcpics-db-label.* at directory Arc-Pics-wrong-2"
	picturesDirName := "Arc-Pics-wrong-2"
	_, err := dbLabel(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}

func TestDbLabel1(t *testing.T) {
	want := "a"
	picturesDirName := "Arc-Pics-good-1"
	label, _ := dbLabel(picturesDirName)
	got := label
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}
