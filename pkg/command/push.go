package command

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/s3"
)

type PushParam struct {
	Name     string
	Endpoint string
	Version  string
	FilePath string
}

func Push(param *PushParam, binPath string) error {
	file, err := os.Open(binPath)
	if err != nil {
		return errors.Wrapf(err, "failed to open %v", binPath)
	}

	sess := session.New()
	s3Client := s3.New(sess)
	log.Println("-->", "Uploading", param.Name, "to", param.Endpoint)

	url, err := s3.ParseURL(param.Endpoint, param.Name, param.Version)
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
