package storage

import (
	"io"
	"net/url"

	"github.com/yuuki/binrep/pkg/release"
)

type TestStorageAPI interface {
	API
	latestTimestamp(name string) (string, error)
	createMeta(u *url.URL, bins []*release.Binary) (*release.Meta, error)
	getBinaryBody(relURL *url.URL, binName string) (io.Reader, error)
}

type fakeStorage struct {
	*_s3
	TestStorageAPI
	FakeLatestTimestamp func(name string) (string, error)
	FakeCreateMeta      func(u *url.URL, bins []*release.Binary) (*release.Meta, error)
	FakeGetBinaryBody   func(relURL *url.URL, binName string) (io.Reader, error)
}

func (s *fakeStorage) latestTimestamp(name string) (string, error) {
	if s.FakeLatestTimestamp == nil {
		return s._s3.latestTimestamp(name)
	}
	return s.FakeLatestTimestamp(name)
}

func (s *fakeStorage) createMeta(u *url.URL, bins []*release.Binary) (*release.Meta, error) {
	if s.FakeCreateMeta == nil {
		return s._s3.createMeta(u, bins)
	}
	return s.FakeCreateMeta(u, bins)
}

func (s *fakeStorage) getBinaryBody(relURL *url.URL, binName string) (io.Reader, error) {
	return s.FakeGetBinaryBody(relURL, binName)
}
