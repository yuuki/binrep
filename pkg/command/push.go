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
	BinName   string
	Timestamp string
	Endpoint  string
}

func Push(param *PushParam, name string, binPath string) error {
	file, err := os.Open(binPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open %v", binPath)
	}

	sess := session.New(aws.NewConfig())
	st := storage.New(sess)

	var binName string
	if param.BinName == "" {
		binName = filepath.Base(file.Name())
	}
	// TODO: multiple Binary
	bin, err := release.BuildBinary(binName, file)
	if err != nil {
		return err
	}
	rel, err := st.CreateRelease(param.Endpoint, name, []*release.Binary{bin})
	if err != nil {
		return err
	}

	log.Println("-->", "Uploading", binPath, "to", param.Endpoint)

	if err := st.PushRelease(rel); err != nil {
		return err
	}
	log.Println("Uploaded", "to", rel.URL)

	return nil
}
