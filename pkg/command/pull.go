package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// PullParam represents the option parameter of `pull`.
type PullParam struct {
	BinName   string
	Timestamp string
	Endpoint  string
}

// Pull pulls the latest release of the name(<host>/<user>/<project>) to installPath.
func Pull(param *PullParam, name, installPath string) error {
	sess := session.New()
	st := storage.New(sess, param.Endpoint)

	rel, err := st.FindLatestRelease(name)
	if err != nil {
		return err
	}

	log.Println("-->", "Downloading", rel.URL, "to", installPath)

	if err := pullRelease(rel, installPath); err != nil {
		return err
	}

	return nil
}

func pullRelease(rel *release.Release, installPath string) error {
	for _, bin := range rel.Meta.Binaries {
		path := filepath.Join(installPath, bin.Name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", path)
		}
		_, err = bin.CopyAndValidateChecksum(file, bin.Body)
		if err != nil {
			if release.IsChecksumError(err) {
				os.Remove(path)
			}
			return err
		}
	}
	return nil
}
