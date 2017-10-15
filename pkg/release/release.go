package release

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"

	strftime "github.com/jehiah/go-strftime"
	"github.com/pkg/errors"
)

type Release struct {
	Meta *Meta
	URL  *url.URL
}

func New(meta *Meta, u *url.URL) *Release {
	return &Release{Meta: meta, URL: u}
}

func (rel *Release) Timestamp() string {
	return filepath.Base(rel.URL.Path)
}

// Inspect inspetcs the release information.
func (rel *Release) Inspect(w io.Writer) {
	fmt.Fprintf(w, "URL\tTIMESTAMP\t")
	for i := 1; i <= len(rel.Meta.Binaries); i++ {
		fmt.Fprintf(w, "BINNAME%d\tBINVERSION%d\tBINCHECKSUM%d\t", i, i, i)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "%s\t%s\t", rel.URL, rel.Timestamp())
	for _, b := range rel.Meta.Binaries {
		b.Inspect(w)
	}
	fmt.Fprintln(w)
}

func Now() string {
	t := time.Now()
	utc, _ := time.LoadLocation("UTC")
	t = t.In(utc)
	return strftime.Format("%Y%m%d%H%M%S", t)
}

// BuildURL builds the binary file url for S3.
func BuildURL(urlStr string, name, timestamp string) (*url.URL, error) {
	//TODO: validate version
	u, err := url.Parse(urlStr + "/" + filepath.Join(name, timestamp))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	return u, nil
}
