package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/sbrepo/pkg/meta"
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

	sess := session.New(aws.NewConfig().WithLogLevel(aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors))
	s3Client := s3.New(sess)

	timestamp := now()
	url, err := s3.BuildURL(param.Endpoint, name, timestamp)
	if err != nil {
		return errors.Wrap(err, "failed to parse url")
	}
	var binName string
	if param.BinName == "" {
		binName = filepath.Base(file.Name())
	}
	sum, err := checksum(file)
	if err != nil {
		return err
	}
	err = s3Client.CreateOrUpdateMeta(url, &meta.Binary{
		Name:      binName,
		Checksum:  sum,
		Timestamp: timestamp,
	})
	if err != nil {
		return err
	}

	log.Println("-->", "Uploading", binPath, "to", param.Endpoint)

	location, err := s3Client.PushBinary(file, url, binName)
	if err != nil {
		return err
	}
	log.Println("Uploaded", "to", location)

	return nil
}
