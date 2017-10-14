package command

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/yuuki/binrep/pkg/s3"
)

type PullParam struct {
	BinName   string
	Timestamp string
	Endpoint  string
}

func Pull(param *PullParam, name, installPath string) error {
	sess := session.New()
	s3Client := s3.New(sess)
	log.Println("-->", "Downloading", param.Endpoint, "to", installPath)

	latest, err := s3Client.LatestTimestamp(param.Endpoint, name)
	if err != nil {
		return err
	}

	url, err := s3.BuildURL(param.Endpoint, name, latest)
	if err != nil {
		return err
	}
	if err := s3Client.PullBinaries(url, installPath); err != nil {
		return err
	}

	return nil
}
