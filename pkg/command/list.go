package command

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// ListParam represents the option parameter of `list`.
type ListParam struct {
	Endpoint string
}

// List lists releases.
func List(param *ListParam) error {
	sess := session.New()
	st := storage.New(sess, param.Endpoint)

	err := st.WalkReleases(1, func(rel *release.Release) error {
		fmt.Println(rel.Prefix())
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
