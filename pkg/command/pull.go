package command

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/s3"
)

type PullParam struct {
	Name     string
	Endpoint string
	Version  string
}

func Pull(param *PullParam, installPath string) error {
	file, err := os.OpenFile(installPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open %v", installPath)
	}

	sess := session.New()
	s3Client := s3.New(sess)
	log.Println("-->", "Downloading", param.Endpoint, "to", installPath)

	url, err := s3.ParseURL(param.Endpoint, param.Name, param.Version)
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	if err := s3Client.PullBinary(file, url); err != nil {
		return err
	}

	return nil
}
