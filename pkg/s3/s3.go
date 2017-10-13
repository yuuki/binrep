package s3

import (
	"io"
	"net/url"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
)

type S3 interface {
	PushBinary(in io.Reader, url *url.URL) (string, error)
}

type _s3 struct {
	uploader   s3manageriface.UploaderAPI
	downloader s3manageriface.DownloaderAPI
}

func ParseURL(urlStr string, name, version string) (*url.URL, error) {
	//TODO: validate version
	url, err := url.Parse(filepath.Join(urlStr, name, version, name))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	return url, nil
}

// New creates a S3 client object.
func New(sess *session.Session) S3 {
	return &_s3{
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

// PushBinary pushes the binary file data into S3.
func (s *_s3) PushBinary(in io.Reader, url *url.URL) (string, error) {
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(url.Host),
		Key:    aws.String(url.Path),
		Body:   in,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to upload file to %v", url)
	}
	return result.Location, nil
}
