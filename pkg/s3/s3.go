package s3

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	gos3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/meta"
)

const (
	BIN_NAME       = "BINARY"
	META_FILE_NAME = "meta.yml"
)

type S3 interface {
	LatestTimestamp(urlStr string, name string) (string, error)
	CreateOrUpdateMeta(u *url.URL, param *meta.Binary) error
	PushBinary(in io.Reader, url *url.URL, binName string) (string, error)
	PullBinary(w io.WriterAt, url *url.URL, binName string) error
	PullBinaries(u *url.URL, installDir string) error
}

type _s3 struct {
	svc        s3iface.S3API
	uploader   s3manageriface.UploaderAPI
	downloader s3manageriface.DownloaderAPI
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

func (s *_s3) CreateMeta(u *url.URL, param *meta.Binary) error {
	m := meta.New(param)
	data, err := yaml.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "failed to marshal yaml")
	}
	_, err = s.svc.PutObject(&gos3.PutObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
		Body:   aws.ReadSeekCloser(bytes.NewReader(data)),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to put meta.yml into s3 (%v)", param.Name)
	}
	return nil
}

func (s *_s3) CreateOrUpdateMeta(u *url.URL, param *meta.Binary) error {
	resp, err := s.svc.GetObject(&gos3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case gos3.ErrCodeNoSuchKey:
				if err := s.CreateMeta(u, param); err != nil {
					return err
				}
				return nil
			default:
			}
		}
		return errors.Wrapf(err, "failed to update meta.yml %s", u)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read meta.yml on s3")
	}

	var m meta.Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return errors.Wrapf(err, "failed to read meta.yml on s3")
	}
	m.AppendBinary(param)
	_, err = s.svc.PutObject(&gos3.PutObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
		Body:   aws.ReadSeekCloser(bytes.NewBuffer(data)),
	})
	if err != nil {
		return errors.Wrap(err, "failed to put meta.yml into s3")
	}

	return nil
}

// PushBinary pushes the binary file data into S3.
func (s *_s3) PushBinary(in io.Reader, url *url.URL, binName string) (string, error) {
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(url.Host),
		Key:    aws.String(filepath.Join(url.Path, binName)),
		Body:   in,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to upload file to %s", url)
	}
	return result.Location, nil
}

// PullBinary pulls the binary file data from S3.
func (s *_s3) PullBinary(w io.WriterAt, u *url.URL, binName string) error {
	_, err := s.downloader.Download(w, &gos3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, binName)),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upload file to %v", u)
	}
	return nil
}

func (s *_s3) PullBinaries(u *url.URL, installDir string) error {
	resp, err := s.svc.GetObject(&gos3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to get object from s3 %s", u)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read meta.yml on s3")
	}
	var m meta.Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return errors.Wrapf(err, "failed to read meta.yml on s3")
	}
	for _, bin := range m.Binaries {
		path := filepath.Join(installDir, bin.Name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", path)
		}
		if err := s.PullBinary(file, u, bin.Name); err != nil {
			return err
		}
		if err := bin.ValidateChecksum(file); err != nil {
			os.Remove(path)
			return err
		}
	}
	return nil
}
