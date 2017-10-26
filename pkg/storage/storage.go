package storage

import (
	"github.com/yuuki/binrep/pkg/release"
)

// API defines the interface of the storage backend layer for S3.
type API interface {
	FindLatestRelease(name string) (*release.Release, error)
	FindReleaseByTimestamp(name, timestamp string) (*release.Release, error)
	CreateRelease(name string, bins []*release.Binary) (*release.Release, error)
	PushRelease(rel *release.Release) error
	DeleteRelease(name, timestamp string) error
	PruneReleases(name string, keep int) ([]string, error)
	WalkReleases(concurrency int, walkfn func(*release.Release) error) error
	WalkLatestReleases(concurrency int, releaseFn func(*release.Release) error) error
}
