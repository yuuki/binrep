package command

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/yuuki/binrep/pkg/storage"
)

// PullParam represents the option parameter of `pull`.
type PullParam struct {
	BinName   string
	Timestamp string
	Endpoint  string
}

// Pull pulls the latest release of the name(<host>/<user>/<project>) to installPath.
func Pull(param *PullParam, name, installPath string) error {
	sess := session.New()
	st := storage.New(sess)

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
