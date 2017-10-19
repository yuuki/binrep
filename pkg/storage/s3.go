package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/yuuki/binrep/pkg/release"
)

// S3 defines the interface of the storage backend layer for S3.
type S3 interface {
	FindLatestRelease(name string) (*release.Release, error)
	FindReleaseByTimestamp(name, timestamp string) (*release.Release, error)
	CreateRelease(name string, bins []*release.Binary) (*release.Release, error)
	PushRelease(rel *release.Release) error
	PullRelease(rel *release.Release, installDir string) error
	DeleteRelease(name, timestamp string) error
	PruneReleases(name string, keep int) ([]string, error)
	WalkLatestReleases(releaseFn func(*release.Release) error) error
}

type _s3 struct {
	bucket     string
	svc        s3iface.S3API
	uploader   s3manageriface.UploaderAPI
	downloader s3manageriface.DownloaderAPI
}

// New creates a S3 client object.
func New(sess *session.Session, bucket string) S3 {
	return &_s3{
		bucket:     strings.TrimPrefix(bucket, "s3://"),
		svc:        s3.New(sess),
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

// BuildReleaseURL builds the binary file url for S3.
func (s *_s3) buildReleaseURL(name, timestamp string) (*url.URL, error) {
	urlStr := fmt.Sprintf("s3://%s/%s", s.bucket, filepath.Join(name, timestamp))
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %v", urlStr)
	}
	return u, nil
}

// FindLatestRelease finds the release including the latest timestamp.
func (s *_s3) FindLatestRelease(name string) (*release.Release, error) {
	latest, err := s.latestTimestamp(name)
	if err != nil {
		return nil, err
	}
	u, err := s.buildReleaseURL(name, latest)
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

// FindReleaseByTimestamp finds the release including the `timestamp`.
func (s *_s3) FindReleaseByTimestamp(name, timestamp string) (*release.Release, error) {
	u, err := s.buildReleaseURL(name, timestamp)
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

// CreateRelease creates the release on S3.
func (s *_s3) CreateRelease(name string, bins []*release.Binary) (*release.Release, error) {
	u, err := s.buildReleaseURL(name, release.Now())
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
func (s *_s3) latestTimestamp(name string) (string, error) {
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(name + "/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to list objects (bucket: %v, path: %v/)", s.bucket, name)
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

// CreateMeta creates the meta.yml on S3.
func (s *_s3) CreateMeta(u *url.URL, bins []*release.Binary) (*release.Meta, error) {
	m := release.NewMeta(bins)
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal yaml")
	}
	_, err = s.svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filepath.Join(u.Path, release.MetaFileName)),
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
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filepath.Join(u.Path, release.MetaFileName)),
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
	for _, b := range m.Binaries {
		b.Body, err = s.GetBinaryBody(u, b.Name)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

func (s *_s3) GetBinaryBody(relURL *url.URL, binName string) (io.Reader, error) {
	key := filepath.Join(relURL.Path, binName)
	resp, err := s.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, errors.Wrapf(err, "not found %v", key)
			default:
			}
		}
		return nil, errors.Wrapf(err, "failed to get object from s3 %s", relURL)
	}
	return resp.Body, nil
}

// PushRelease pushes the release into S3.
func (s *_s3) PushRelease(rel *release.Release) error {
	for _, bin := range rel.Meta.Binaries {
		_, err := s.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(filepath.Join(rel.URL.Path, bin.Name)),
			Body:   bin.Body,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to upload file to %s", rel.URL)
		}
	}
	return nil
}

// PullRelease pulls the binary files within the release on S3 to installDir.
func (s *_s3) PullRelease(rel *release.Release, installDir string) error {
	for _, bin := range rel.Meta.Binaries {
		path := filepath.Join(installDir, bin.Name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", path)
		}

		_, err = s.downloader.Download(file, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
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

func (s *_s3) ascTimestamps(name string) ([]string, error) {
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(name + "/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list objects (bucket: %v, path: %v/)", s.bucket, name)
	}
	if len(resp.CommonPrefixes) < 1 {
		return nil, errors.Errorf("no such projects %v", name)
	}
	timestamps := make([]string, 0, len(resp.CommonPrefixes))
	for _, cp := range resp.CommonPrefixes {
		timestamps = append(timestamps, filepath.Base(*cp.Prefix))
	}
	sort.Strings(timestamps)
	return timestamps, nil
}

// DeleteRelease deletes the release with the `timestamp`.
func (s *_s3) DeleteRelease(name, timestamp string) error {
	rel, err := s.FindReleaseByTimestamp(name, timestamp)
	if err != nil {
		return err
	}
	// recursively delete
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(rel.Prefix()),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list objects (bucket: %v, key: %v/)", s.bucket, rel.URL.Path)
	}
	if *resp.IsTruncated {
		//TODO: paging
		log.Printf("too many objects (bucket: %v, key: %v/)\n", s.bucket, rel.URL.Path)
	}
	for _, obj := range resp.Contents {
		_, err := s.svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    obj.Key,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete object (bucket: %v, key: %v)", s.bucket, obj.Key)
		}
	}
	return nil
}

// PruneReleases prunes the `keep` of old releases.
func (s *_s3) PruneReleases(name string, keep int) ([]string, error) {
	timestamps, err := s.ascTimestamps(name)
	if err != nil {
		return nil, err
	}
	var prunedTimestamps []string
	if len(timestamps) > keep {
		n := len(timestamps) - keep
		prunedTimestamps = timestamps[0:n]
		for _, t := range prunedTimestamps {
			if err := s.DeleteRelease(name, t); err != nil {
				return nil, err
			}
		}
	}
	return prunedTimestamps, nil
}

const (
	timestampFormat = "20060102150405"
)

func isTimestamp(str string) bool {
	if _, err := time.Parse(timestampFormat, str); err != nil {
		if _, ok := err.(*time.ParseError); ok {
			return false
		}
		return false
	}
	return true
}

func parseName(str string) (bool, string) {
	// str is expected to be 'github.com/yuuki/droot/20171017152508/droot'
	// or 'github.com/yuuki/droot/20171017152508'
	items := strings.Split(str, "/")
	if len(items) < 2 {
		return false, ""
	}
	tail, oneBeforeTail := items[len(items)-1], items[len(items)-2]
	if isTimestamp(tail) {
		return true, strings.Join(items[0:len(items)-1], "/")
	}
	if isTimestamp(oneBeforeTail) {
		return true, strings.Join(items[0:len(items)-2], "/")
	}
	return false, ""
}

func (s *_s3) walkNames(prefix string, walkfn func(name string) error) error {
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list objects (bucket: %v, key: %v/)", s.bucket, prefix)
	}
	if *resp.IsTruncated {
		//TODO: paging
		log.Printf("too many objects (bucket: %v, key: %v/)\n", s.bucket, prefix)
	}
	for _, obj := range resp.Contents {
		if ok, name := parseName(*obj.Key); ok {
			log.Println(name)
			if err := walkfn(name); err != nil {
				return err
			}
			return nil
		}
		if err := s.walkNames(filepath.Join(prefix, *obj.Key), walkfn); err != nil {
			return err
		}
	}
	return nil
}

// WalkLatestReleases walks the latest releases with the bucket.
func (s *_s3) WalkLatestReleases(releaseFn func(*release.Release) error) error {
	err := s.walkNames("", func(name string) error {
		rel, err := s.FindLatestRelease(name)
		if err != nil {
			return err
		}
		if err := releaseFn(rel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
