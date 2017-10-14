package command

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/yuuki/binrep/pkg/storage"
)

type PullParam struct {
	BinName   string
	Timestamp string
	Endpoint  string
}

func Pull(param *PullParam, name, installPath string) error {
	sess := session.New()
	st := storage.New(sess)

	// FindLatestRelease
	rel, err := st.FindLatestRelease(param.Endpoint, name)
	if err != nil {
		return err
	}

	log.Println("-->", "Downloading", param.Endpoint, "to", installPath)

	if err := st.PullRelease(rel, installPath); err != nil {
		return err
	}

	return nil
}
