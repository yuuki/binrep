package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/s3"
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

	sess := session.New()
	s3Client := s3.New(sess)
	log.Println("-->", "Uploading", binPath, "to", param.Endpoint)

	var binName string
	if param.BinName == "" {
		binName = filepath.Base(binPath)
	}
	url, err := s3.BuildURL(param.Endpoint, name, now(), binName)
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	location, err := s3Client.PushBinary(file, url)
	if err != nil {
		return err
	}
	log.Println("Uploaded", "to", location)

	return nil
}
