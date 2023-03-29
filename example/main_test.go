package main

import (
	"testing"
)

func TestDbPostFix0(t *testing.T) {
	want := "there is no file arcpics-db-postfix.* at directory Arc-Pics-wrong-0 - should be e.g. arcpics-db-postfix.001"
	picturesDirName := "Arc-Pics-wrong-0"
	_, err := dbPostfix(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}
func TestDbPostFix2(t *testing.T) {
	want := "unexpected number files arcpics-db-postfix.* at directory Arc-Pics-wrong-2"
	picturesDirName := "Arc-Pics-wrong-2"
	_, err := dbPostfix(picturesDirName)
	got := err.Error()
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}

func TestDbPostFix1(t *testing.T) {
	want := "a"
	picturesDirName := "Arc-Pics-good-1"
	postfix, _ := dbPostfix(picturesDirName)
	got := postfix
	if got != want {
		t.Errorf("Unexpected error - WANT: %s; GOT: %s", want, got)
	}
}
