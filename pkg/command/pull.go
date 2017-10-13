package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/s3"
)

type PullParam struct {
	BinName   string
	Timestamp string
	Endpoint  string
}

func Pull(param *PullParam, name, installPath string) error {
	file, err := os.OpenFile(installPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open %v", installPath)
	}

	sess := session.New()
	s3Client := s3.New(sess)
	log.Println("-->", "Downloading", param.Endpoint, "to", installPath)

	latest, err := s3Client.LatestTimestamp(param.Endpoint, name)
	if err != nil {
		return err
	}

	binName := filepath.Base(name)
	url, err := s3.BuildURL(param.Endpoint, name, latest, binName)
	if err != nil {
		return err
	}
	if err := s3Client.PullBinary(file, url); err != nil {
		return err
	}

	return nil
}
