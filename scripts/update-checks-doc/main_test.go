package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Disable trace logs while running tests. Commenting out this function is useful when debugging test cases.
func init() {
	log.SetOutput(io.Discard)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func testErr(t *testing.T, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("error did not occur")
	}
	msg := err.Error()
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("error message %q does not cotnain expected text %q", msg, w)
		}
	}
}

func TestMainGenerateOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("update-checks-doc doesn't support Windows")
	}
	root := t.TempDir()

	in := must(os.Open(filepath.FromSlash("testdata/ok/minimal.in")))
	path := filepath.FromSlash(root + "/minimal.in")
	tmp := must(os.Create(path))
	must(io.Copy(tmp, in))
	in.Close()
	tmp.Close()

	if err := Main([]string{"exe", path}); err != nil {
		t.Fatal(err)
	}

	want := must(os.ReadFile(filepath.FromSlash("testdata/ok/minimal.out")))
	have := must(os.ReadFile(path))
	if !bytes.Equal(want, have) {
		t.Fatal(cmp.Diff(want, have))
	}
}

func TestMainCheckOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("update-checks-doc doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.out")
	if err := Main([]string{"exe", "-check", path}); err != nil {
		t.Fatal(err)
	}
}

func TestMainCheckError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("update-checks-doc doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.in")
	testErr(t, Main([]string{"exe", "-check", path}), "checks document has some update")
}

func TestMainFileNotFound(t *testing.T) {
	testErr(t, Main([]string{"exe", "-check", "this-file-does-not-exist.md"}), "could not read the document file")
}

func TestMainTooManyArgs(t *testing.T) {
	testErr(t, Main([]string{"exe", "a", "b", "c"}), "usage: update-checks-doc [-check] FILE")
}

func TestMainInvalidCheckFlag(t *testing.T) {
	testErr(t, Main([]string{"exe", "-c", "foo.md"}), "usage: update-checks-doc [-check] FILE")
}

func TestMainNoUpdate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("update-checks-doc doesn't support Windows")
	}
	path := filepath.FromSlash("testdata/ok/minimal.out")
	if err := Main([]string{"exe", path}); err != nil {
		t.Fatal(err)
	}
}

func TestMainUpdateError(t *testing.T) {
	path := filepath.FromSlash("testdata/err/no_playground_link.md")
	if err := Main([]string{"exe", path}); err == nil {
		t.Fatal("no error occurred")
	}
}

func TestUpdateOK(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("update-checks-doc doesn't support Windows")
	}

	dir := filepath.FromSlash("testdata/ok")

	tests := []string{}
	for _, e := range must(os.ReadDir(dir)) {
		n := e.Name()
		if !strings.HasSuffix(n, ".in") {
			continue
		}
		tests = append(tests, strings.TrimSuffix(n, filepath.Ext(n)))
	}

	for _, tc := range tests {
		in := filepath.Join(dir, tc+".in")
		out := filepath.Join(dir, tc+".out")
		t.Run(tc, func(t *testing.T) {
			have, err := Update(must(os.ReadFile(in)))
			if err != nil {
				t.Fatal(err)
			}
			want := must(os.ReadFile(out))
			if !bytes.Equal(want, have) {
				t.Fatal(cmp.Diff(want, have))
			}
		})
	}
}

func TestUpdateError(t *testing.T) {
	dir := filepath.FromSlash("testdata/err")

	tests := []string{}
	for _, e := range must(os.ReadDir(dir)) {
		n := e.Name()
		if !strings.HasSuffix(n, ".md") {
			continue
		}
		tests = append(tests, strings.TrimSuffix(n, filepath.Ext(n)))
	}

	for _, tc := range tests {
		path := filepath.Join(dir, tc+".md")
		t.Run(tc, func(t *testing.T) {
			if _, err := Update(must(os.ReadFile(path))); err == nil {
				t.Fatal("no error occurred")
			}
		})
	}
}
