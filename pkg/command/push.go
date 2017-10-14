package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

type PushParam struct {
	Timestamp string
	Endpoint  string
}

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

	sess := session.New(aws.NewConfig())
	st := storage.New(sess)

	rel, err := st.CreateRelease(param.Endpoint, name, bins)
	if err != nil {
		return err
	}

	log.Println("-->", "Uploading", binPaths, "to", param.Endpoint)

	if err := st.PushRelease(rel); err != nil {
		return err
	}

	log.Println("Uploaded", "to", rel.URL)

	return nil
}
