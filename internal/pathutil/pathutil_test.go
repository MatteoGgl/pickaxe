package pathutil_test

import (
	"os"
	"strings"
	"testing"

	"github.com/matteo/pickaxe/internal/pathutil"
)

func TestExpandHome_Tilde(t *testing.T) {
	got, err := pathutil.ExpandHome("~/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	if !strings.HasPrefix(got, home) {
		t.Errorf("got %q, want prefix %q", got, home)
	}
	if !strings.HasSuffix(got, "/foo/bar") {
		t.Errorf("got %q, want suffix /foo/bar", got)
	}
}

func TestExpandHome_NoTilde(t *testing.T) {
	got, err := pathutil.ExpandHome("/absolute/path")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/absolute/path" {
		t.Errorf("got %q, want %q", got, "/absolute/path")
	}
}

func TestExpandHome_TildeOnly(t *testing.T) {
	got, err := pathutil.ExpandHome("~")
	if err != nil {
		t.Fatal(err)
	}
	// bare ~ is not expanded
	if got != "~" {
		t.Errorf("got %q, want %q", got, "~")
	}
}

func TestExpandHome_Empty(t *testing.T) {
	got, err := pathutil.ExpandHome("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestExpandAndResolve_Tilde(t *testing.T) {
	got, err := pathutil.ExpandAndResolve("~/foo")
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	want := home + "/foo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
