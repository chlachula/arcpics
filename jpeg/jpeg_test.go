package jpeg

import (
	"testing"
)

func TestVoid(t *testing.T) {
	want := "void"
	got := "void"
	if got != want {
		t.Errorf("error - WANT: %s; GOT: %s", want, got)
	}
}
