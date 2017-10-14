package release

import (
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
