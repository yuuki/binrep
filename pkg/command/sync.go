package command

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// SyncParam represents the option parameter of `show`.
type SyncParam struct {
	Endpoint string
}

// Sync syncs the latest releases to rootDir.
func Sync(param *SyncParam, rootDir string) error {
	sess := session.New()
	st := storage.New(sess, param.Endpoint)

	err := st.WalkLatestReleases(func(rel *release.Release) error {
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
			if _, err := io.Copy(dest, bin.Body); err != nil {
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
