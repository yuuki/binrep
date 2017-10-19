package release

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	strftime "github.com/jehiah/go-strftime"
)

// Release represents a `<host>/<user>/<project>/<timestamp>/` layout.
type Release struct {
	Meta *Meta
	URL  *url.URL
}

// New returns a Release object.
func New(meta *Meta, u *url.URL) *Release {
	return &Release{Meta: meta, URL: u}
}

// Name returns the `<host>/<user>/<project>`.
func (rel *Release) Name() string {
	return strings.TrimPrefix(filepath.Dir(rel.URL.Path), "/")
}

// Timestamp returns the timestamp.
func (rel *Release) Timestamp() string {
	return filepath.Base(rel.URL.Path)
}

// Prefix returns the `<host>/<user>/<project>/<timestamp>/`.
func (rel *Release) Prefix() string {
	return strings.Join([]string{rel.Name(), rel.Timestamp()}, "/")
}

// MetaPath returns the meta file path within the release.
// eg. [`<host>/<user>/<project>/<timestamp>/meta.yml`]
func (rel *Release) MetaPath() string {
	return filepath.Join(rel.Prefix(), MetaFileName)
}

// Inspect inspetcs the release information.
func (rel *Release) Inspect(w io.Writer) {
	fmt.Fprintf(w, "NAME\tTIMESTAMP\t")
	for i := 1; i <= len(rel.Meta.Binaries); i++ {
		fmt.Fprintf(w, "BINNARY%d\t", i)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\t%s\t", rel.Name(), rel.Timestamp())
	for _, b := range rel.Meta.Binaries {
		b.Inspect(w)
	}
	fmt.Fprintln(w)
}

// Now returns the current UTC timestamp.
func Now() string {
	t := time.Now()
	utc, _ := time.LoadLocation("UTC")
	t = t.In(utc)
	return strftime.Format("%Y%m%d%H%M%S", t)
}
