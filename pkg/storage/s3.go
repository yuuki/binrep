package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ivpusic/grpool"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/yuuki/binrep/pkg/config"
	"github.com/yuuki/binrep/pkg/release"
)

const (
	jobQueueLen = 100
)

type s3API interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	ListObjectsV2(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	DeleteObject(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

type s3UploaderAPI interface {
	Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type _s3 struct {
	bucket   string
	svc      s3API
	uploader s3UploaderAPI
}

// New creates a StorageAPI client object.
func New(sess *session.Session) API {
	return &_s3{
		bucket:   strings.TrimPrefix(config.Config.BackendEndpoint, "s3://"),
		svc:      s3.New(sess),
		uploader: s3manager.NewUploader(sess),
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

// ExistRelease returns whether the name exists or not.
func (s *_s3) ExistRelease(name string) (bool, error) {
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(name + "/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to list objects (bucket: %v, path: %v/)", s.bucket, name)
	}
	if len(resp.CommonPrefixes) < 1 {
		// not found name
		return false, nil
	}
	return true, nil
}

// HaveSameChecksums returns whether each checksum of given binaries is
// the same or not with each checksum of binaries on the S3.
func (s *_s3) HaveSameChecksums(name string, bins []*release.Binary) (bool, error) {
	latestRel, err := s.FindLatestRelease(name)
	if err != nil {
		return false, err
	}
	for _, lbin := range latestRel.Meta.Binaries {
		for _, bin := range bins {
			if bin.Name == lbin.Name {
				if bin.Checksum != lbin.Checksum {
					return false, nil
				}
			}
		}
	}
	return true, nil
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
func (s *_s3) CreateRelease(name string, timestamp string, bins []*release.Binary) (*release.Release, error) {
	u, err := s.buildReleaseURL(name, timestamp)
	if err != nil {
		return nil, err
	}
	meta, err := s.createMeta(u, bins)
	if err != nil {
		return nil, err
	}
	for _, bin := range meta.Binaries {
		_, err := s.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(filepath.Join(u.Path, bin.Name)),
			Body:   bin.Body,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to upload file to %s", u)
		}
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

// createMeta creates the meta.yml on S3.
func (s *_s3) createMeta(u *url.URL, bins []*release.Binary) (*release.Meta, error) {
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
		b.Body, err = s.getBinaryBody(u, b.Name)
		if err != nil {
			return nil, err
		}
	}
	return &m, nil
}

// getBinaryBody returns the binary body reader.
func (s *_s3) getBinaryBody(relURL *url.URL, binName string) (io.Reader, error) {
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

func (s *_s3) walkReleases(pool *grpool.Pool, prefix string, walkfn func(*release.Release) error) error {
	resp, err := s.svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list objects (bucket: %v, key: %v/)", s.bucket, prefix)
	}
	if *resp.IsTruncated {
		//TODO: paging
		log.Printf("too many objects (bucket: %v, key: %v/)\n", s.bucket, prefix)
	}
	var foundErr error // just use nonzeo exit
	for _, obj := range resp.CommonPrefixes {
		releasePath := *obj.Prefix
		if ok, name := release.ParseName(releasePath); ok {
			name := name
			pool.WaitCount(1)
			pool.JobQueue <- func() {
				defer pool.JobDone()

				rel, err := s.FindReleaseByTimestamp(name, filepath.Base(releasePath))
				if err != nil {
					log.Printf("failed to find release %s: %s\n", releasePath, err)
					// just put error log, not to exit
					foundErr = err
					return
				}
				if err := walkfn(rel); err != nil {
					log.Printf("failed to walk %s: %s\n", releasePath, err)
					// just put error log, not to exit
					foundErr = err
					return
				}
			}
		}
		if err := s.walkReleases(pool, releasePath, walkfn); err != nil {
			return err
		}
	}
	pool.WaitAll()
	if foundErr != nil {
		return foundErr
	}
	return nil
}

// WalkReleases walks releases.
func (s *_s3) WalkReleases(concurrency int, releaseFn func(*release.Release) error) error {
	pool := grpool.NewPool(concurrency, jobQueueLen)
	defer pool.Release()

	err := s.walkReleases(pool, "", func(rel *release.Release) error {
		return releaseFn(rel)
	})
	if err != nil {
		return err
	}
	return nil
}
