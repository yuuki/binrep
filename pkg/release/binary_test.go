package release

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestBuildBinary(t *testing.T) {
	body := new(bytes.Buffer)
	got, err := BuildBinary("github.com/yuuki/droot", 0755, body)
	if err != nil {
		t.Fatalf("should not raise error: %s", err)
	}

	expectedName := "github.com/yuuki/droot"
	if got.Name != expectedName {
		t.Errorf("Binary.Name = %q; want %q", got.Name, expectedName)
	}

	expectedChecksum := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got.Checksum != expectedChecksum {
		t.Errorf("Binary.Checksum = %q; want %q", got.Checksum, expectedChecksum)
	}
	var expectedMode os.FileMode = 0755
	if got.Mode != expectedMode {
		t.Errorf("Binary.Name = %q; want %q", got.Mode, expectedMode)
	}

	if got.Body != body {
		t.Errorf("Binary.Body = %v; want %v", got.Body, body)
	}
}

func TestBuildBinary_errorChecksumNil(t *testing.T) {
	var body io.Reader
	_, err := BuildBinary("github.com/yuuki/droot", 0755, body)
	if fmt.Sprintf("%s", err) != "try to read nil" {
		t.Fatalf("err = %q; want %q", err, "try to read nil")
	}
}

func TestBuildBinary_errorChecksumClosedFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "fake")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	_, err = BuildBinary("github.com/yuuki/droot", 0755, tmpfile)

	if fmt.Sprintf("%s", err) != "failed to read data for checksum" {
		t.Fatalf("err = %q; want %q", err, "failed to read data for checksum")
	}
}

func TestBinaryInspect(t *testing.T) {
	b, err := BuildBinary("github.com/yuuki/droot", 0755, bytes.NewBufferString("body"))
	if err != nil {
		panic(err)
	}
	w := new(bytes.Buffer)

	b.Inspect(w)

	expected := "github.com/yuuki/droot/-rwxr-xr-x/230d835\t"
	if w.String() != expected {
		t.Errorf("got: %v, want: %v", w.String(), expected)
	}
}
