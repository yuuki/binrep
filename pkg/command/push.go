package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/yuuki/binrep/pkg/meta"
	"github.com/yuuki/binrep/pkg/s3"
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
	s3Client := s3.New(sess)

	var binName string
	if param.BinName == "" {
		binName = filepath.Base(file.Name())
	}
	bin, err := meta.BuildBinary(file, binName)
	if err != nil {
		return err
	}
	url, err := s3.BuildURL(param.Endpoint, bin.Name, bin.Timestamp)
	if err != nil {
		return err
	}
	if err = s3Client.CreateOrUpdateMeta(url, []*meta.Binary{bin}); err != nil {
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
