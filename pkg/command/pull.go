package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	humanize "github.com/dustin/go-humanize"
	"github.com/fujiwara/shapeio"
	"github.com/pkg/errors"

	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// PullParam represents the option parameter of `pull`.
type PullParam struct {
	Timestamp    string
	MaxBandWidth string
}

// Pull pulls the latest release of the name(<host>/<user>/<project>) to installPath.
func Pull(param *PullParam, name, installPath string) error {
	sess := session.New()
	st := storage.New(sess)

	fi, err := os.Stat(installPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open %q", installPath)
	}
	if !fi.IsDir() {
		return errors.Errorf("%q not directory", installPath)
	}

	rel, err := st.FindLatestRelease(name)
	if err != nil {
		return err
	}

	var maxBandWidth uint64
	if param.MaxBandWidth != "" {
		maxBandWidth, err = humanize.ParseBytes(param.MaxBandWidth)
		if err != nil {
			return errors.Errorf("failed to parse --max-bandwidth %v", humanize.Bytes(maxBandWidth))
		}
	}

	log.Println("-->", "Downloading", rel.URL, "to", installPath)

	return pullRelease(rel, installPath, maxBandWidth)
}

func pullRelease(rel *release.Release, installPath string, maxBandWidth uint64) error {
	for _, bin := range rel.Meta.Binaries {
		path := filepath.Join(installPath, bin.Name)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, bin.Mode)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", path)
		}
		lsrc := shapeio.NewReader(bin.Body)
		if maxBandWidth != 0 {
			log.Printf("Set max bandwidth total: %s/sec\n", humanize.Bytes(uint64(maxBandWidth)))
			lsrc.SetRateLimit(float64(maxBandWidth))
		}
		_, err = bin.CopyAndValidateChecksum(file, lsrc)
		if err != nil {
			if release.IsChecksumError(err) {
				os.Remove(path)
			}
			return err
		}
	}
	return nil
}
