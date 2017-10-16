package command

import (
	"os"
	"text/tabwriter"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/yuuki/binrep/pkg/release"
	"github.com/yuuki/binrep/pkg/storage"
)

// ShowParam represents the option parameter of `show`.
type ShowParam struct {
	Timestamp string
	Endpoint  string
}

// Show shows the latest release of the name(<host>/<user>/<project>).
func Show(param *ShowParam, name string) error {
	sess := session.New()
	st := storage.New(sess)

	var (
		rel *release.Release
		err error
	)
	if param.Timestamp == "" {
		rel, err = st.FindLatestRelease(param.Endpoint, name)
		if err != nil {
			return err
		}
	} else {
		rel, err = st.FindReleaseByTimestamp(param.Endpoint, name, param.Timestamp)
		if err != nil {
			return err
		}
	}

	// Format in tab-separated columns with a tab stop of 8.
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	rel.Inspect(tw)
	tw.Flush()

	return nil
}
