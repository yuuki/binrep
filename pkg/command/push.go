package command

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// PushParam represents the option parameter of `push`.
type PushParam struct {
	Timestamp    string
	KeepReleases int
	Force        bool
	Endpoint     string
}

// Push pushes the binary files of binPaths as release of the name(<host>/<user>/<project>).
func Push(param *PushParam, name string, binPaths []string) error {
	// TODO: Validate the same file name
	bins := make([]*release.Binary, 0, len(binPaths))
	for _, binPath := range binPaths {
		file, err := os.Open(binPath)
		if err != nil {
			return errors.Wrapf(err, "failed to open %v", binPath)
		}

		bin, err := release.BuildBinary(filepath.Base(file.Name()), file)
		if err != nil {
			return err
		}
		bins = append(bins, bin)
	}

	sess := session.New()
	st := storage.New(sess, param.Endpoint)

	if !param.Force {
		ok, err := st.HaveSameChecksums(name, bins)
		if err != nil {
			return err
		}
		if ok {
			log.Println("Skip pushing the binaries because they have the same checksum with the latest binaries on the remote storage")
			return nil
		}
	}

	rel, err := st.CreateRelease(name, bins)
	if err != nil {
		return err
	}

	log.Println("-->", "Uploading", binPaths, "to", rel.URL)

	if err := st.PushRelease(rel); err != nil {
		return err
	}

	log.Println("Uploaded", "to", rel.URL)

	log.Println("--> Cleaning up the old releases")

	timestamps, err := st.PruneReleases(name, param.KeepReleases)
	if err != nil {
		return err
	}

	log.Println("Cleaned", "up", strings.Join(timestamps, ","))

	return nil
}
