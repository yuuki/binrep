package command

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	humanize "github.com/dustin/go-humanize"
	"github.com/fujiwara/shapeio"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// SyncParam represents the option parameter of `show`.
type SyncParam struct {
	Endpoint     string
	Concurrency  int
	MaxBandWidth string
}

// Sync syncs the latest releases to rootDir.
func Sync(param *SyncParam, rootDir string) error {
	var maxBandWidth uint64
	if param.MaxBandWidth != "" {
		totalMaxBandWidth, err := humanize.ParseBytes(param.MaxBandWidth)
		if err != nil {
			return errors.Errorf("failed to parse --max-bandwidth %v", humanize.Bytes(maxBandWidth))
		}
		maxBandWidth = totalMaxBandWidth / uint64(param.Concurrency)
		log.Printf("Set max bandwidth total: %s/sec, per-release: %s/sec\n",
			humanize.Bytes(uint64(totalMaxBandWidth)), humanize.Bytes(uint64(maxBandWidth)))
	}

	sess := session.New()
	st := storage.New(sess, param.Endpoint)

	err := st.WalkLatestReleases(param.Concurrency, func(rel *release.Release) error {
		relPrefix := filepath.Join(rootDir, rel.Prefix())
		// Skip download if the release already exists on the local filesystem.
		if _, err := os.Stat(relPrefix); err == nil {
			log.Printf("Skipped the download of %v\n", relPrefix)
			return nil
		}

		log.Printf("--> Downloading to %v\n", relPrefix)

		if err := os.MkdirAll(relPrefix, 0755); err != nil {
			return errors.Wrapf(err, "failed to create directory %v", relPrefix)
		}
		for _, bin := range rel.Meta.Binaries {
			binPath := filepath.Join(relPrefix, bin.Name)
			dest, err := os.OpenFile(binPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return errors.Wrapf(err, "failed to open %v", binPath)
			}
			lsrc := shapeio.NewReader(bin.Body)
			if maxBandWidth != 0 {
				lsrc.SetRateLimit(float64(maxBandWidth))
			}
			if _, err := io.Copy(dest, lsrc); err != nil {
				return errors.Wrapf(err, "failed to write file %v", binPath)
			}
		}
		metaData, err := yaml.Marshal(rel.Meta)
		if err != nil {
			return errors.Wrap(err, "failed to marshal yaml")
		}
		metaPath := filepath.Join(rootDir, rel.MetaPath())
		if err := ioutil.WriteFile(metaPath, metaData, 0644); err != nil {
			return errors.Wrapf(err, "failed to write file %v", metaPath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
