package s3

import (
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	gos3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
)

const (
	BIN_NAME = "BINARY"
)

type S3 interface {
	LatestTimestamp(urlStr string, name string) (string, error)
	PushBinary(in io.Reader, url *url.URL) (string, error)
	PullBinary(w io.WriterAt, url *url.URL) error
}

type _s3 struct {
	svc        s3iface.S3API
	uploader   s3manageriface.UploaderAPI
	downloader s3manageriface.DownloaderAPI
}

// BuildURL builds the binary file url for S3.
func BuildURL(urlStr string, name, timestamp, binName string) (*url.URL, error) {
	//TODO: validate version
	u, err := url.Parse(urlStr + "/" + filepath.Join(name, timestamp, binName))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	return u, nil
}

// New creates a S3 client object.
func New(sess *session.Session) S3 {
	return &_s3{
		svc:        gos3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

// LatestTimestamp gets the latest timestamp.
func (s *_s3) LatestTimestamp(urlStr string, name string) (string, error) {
	u, err := url.Parse(urlStr + "/" + name)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	resp, err := s.svc.ListObjectsV2(&gos3.ListObjectsV2Input{
		Bucket:    aws.String(u.Host),
		Prefix:    aws.String(strings.TrimLeft(u.Path, "/") + "/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to list objects (bucket: %v, path: %v/)", u.Host, u.Path)
	}
	if len(resp.CommonPrefixes) < 1 {
		return "", errors.Errorf("no such projects %v", name)
	}
	timestamps := make([]string, 0, len(resp.CommonPrefixes))
	for _, cp := range resp.CommonPrefixes {
		timestamps = append(timestamps, filepath.Base(*cp.Prefix))
	}
	sort.Strings(timestamps)
	return timestamps[len(timestamps)-1], nil
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

// PullBinary pulls the binary file data from S3.
func (s *_s3) PullBinary(w io.WriterAt, url *url.URL) error {
	_, err := s.downloader.Download(w, &gos3.GetObjectInput{
		Bucket: aws.String(url.Host),
		Key:    aws.String(url.Path),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upload file to %v", url)
	}
	return nil
}
