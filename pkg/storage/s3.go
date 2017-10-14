package storage

import (
	"bytes"
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
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
	"github.com/yuuki/binrep/pkg/release"
)

const (
	META_FILE_NAME = "meta.yml"
)

type S3 interface {
	FindLatestRelease(endpoint, name string) (*release.Release, error)
	CreateRelease(endpoint, name string, bins []*release.Binary) (*release.Release, error)
	PushRelease(rel *release.Release) error
	PullRelease(rel *release.Release, installDir string) error
}

type _s3 struct {
	svc        s3iface.S3API
	uploader   s3manageriface.UploaderAPI
	downloader s3manageriface.DownloaderAPI
}

// New creates a S3 client object.
func New(sess *session.Session) S3 {
	return &_s3{
		svc:        s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

func (s *_s3) FindLatestRelease(endpoint, name string) (*release.Release, error) {
	latest, err := s.latestTimestamp(endpoint, name)
	if err != nil {
		return nil, err
	}
	u, err := release.BuildURL(endpoint, name, latest)
	if err != nil {
		return nil, err
	}
	meta, err := s.FindMeta(u)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, errors.Errorf("meta.yml not found %s", u)
	}
	return release.New(meta, u), nil
}

func (s *_s3) CreateRelease(endpoint, name string, bins []*release.Binary) (*release.Release, error) {
	u, err := release.BuildURL(endpoint, name, release.Now())
	if err != nil {
		return nil, err
	}
	meta, err := s.CreateMeta(u, bins)
	if err != nil {
		return nil, err
	}
	return release.New(meta, u), nil
}

// latestTimestamp gets the latest timestamp.
func (s *_s3) latestTimestamp(urlStr string, name string) (string, error) {
	u, err := url.Parse(urlStr + "/" + name)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
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

func (s *_s3) CreateMeta(u *url.URL, bins []*release.Binary) (*release.Meta, error) {
	m := release.NewMeta(bins)
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal yaml")
	}
	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
		Body:   aws.ReadSeekCloser(bytes.NewReader(data)),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to put meta.yml into s3 (%s)", u)
	}
	return m, nil
}

// FindMeta finds metadata from S3, and returns nil if meta.yml is not found.
func (s *_s3) FindMeta(u *url.URL) (*release.Meta, error) {
	resp, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(filepath.Join(u.Path, META_FILE_NAME)),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, nil
			default:
			}
		}
		return nil, errors.Wrapf(err, "failed to get object from s3 %s", u)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read meta.yml on s3")
	}
	var m release.Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, errors.Wrapf(err, "failed to read meta.yml on s3")
	}
	return &m, nil
}

// PushRelease pushes the release into S3.
func (s *_s3) PushRelease(rel *release.Release) error {
	for _, bin := range rel.Meta.Binaries {
		_, err := s.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(rel.URL.Host),
			Key:    aws.String(filepath.Join(rel.URL.Path, bin.Name)),
			Body:   bin.Body,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upload file to %s", rel.URL)
		}
	}
	return nil
}

func (s *_s3) PullRelease(rel *release.Release, installDir string) error {
	for _, bin := range rel.Meta.Binaries {
		path := filepath.Join(installDir, bin.Name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", path)
		}

		_, err = s.downloader.Download(file, &s3.GetObjectInput{
			Bucket: aws.String(rel.URL.Host),
			Key:    aws.String(filepath.Join(rel.URL.Path, bin.Name)),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upload file to %v", rel.URL)
		}

		if err := bin.ValidateChecksum(file); err != nil {
			os.Remove(path)
			return err
		}
	}
	return nil
}
