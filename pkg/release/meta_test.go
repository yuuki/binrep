package release

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestNewMeta(t *testing.T) {
	bins := []*Binary{
		{
			Name:     "github.com/yuuki/droot",
			Checksum: "ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48",
		},
		{
			Name:     "github.com/yuuki/grabeni",
			Checksum: "3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259",
		},
	}

	meta := NewMeta(bins)

	if diff := pretty.Compare(meta.Binaries, bins); diff != "" {
		t.Errorf("diff: (-actual +expected)\n%s", diff)
	}
}
