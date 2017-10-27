package release

import (
	"bytes"
	"net/url"
	"testing"
)

func TestNew(t *testing.T) {
	meta := NewMeta([]*Binary{
		{
			Name:     "droot",
			Checksum: "ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48",
		},
		{
			Name:     "grabeni",
			Checksum: "3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259",
		},
	})

	u, err := url.Parse("s3://binreptestbucket/github.com/yuuki/tools/20171019204009")
	if err != nil {
		panic(err)
	}
	rel := New(meta, u)

	if rel.Name() != "github.com/yuuki/tools" {
		t.Errorf("release.Name() = %q; want %q", rel.Name(), "github.com/yuuki/tools")
	}
	if rel.Timestamp() != "20171019204009" {
		t.Errorf("release.Name() = %q; want %q", rel.Timestamp(), "20171019204009")
	}
	if rel.Prefix() != "github.com/yuuki/tools/20171019204009" {
		t.Errorf("release.Name() = %q; want %q", rel.Prefix(), "github.com/yuuki/tools/20171019204009")
	}
	if rel.MetaPath() != "github.com/yuuki/tools/20171019204009/meta.yml" {
		t.Errorf("release.Name() = %q; want %q", rel.Prefix(), "github.com/yuuki/tools/20171019204009/meta.yml")
	}
}

func TestReleaseInspect(t *testing.T) {
	meta := NewMeta([]*Binary{
		{
			Name:     "droot",
			Checksum: "ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48",
		},
		{
			Name:     "grabeni",
			Checksum: "3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259",
		},
	})

	u, err := url.Parse("s3://binreptestbucket/github.com/yuuki/tools/20171019204009")
	if err != nil {
		panic(err)
	}
	rel := New(meta, u)

	out := new(bytes.Buffer)

	rel.Inspect(out)

	expected := "NAME\tTIMESTAMP\tBINNARY1\tBINNARY2\t\ngithub.com/yuuki/tools\t20171019204009\tdroot/ec9efb6\tgrabeni/3e30f16\t\n"
	if out.String() != expected {
		t.Errorf("got: %q, want: %q", out.String(), expected)
	}
}

func TestParseName(t *testing.T) {
	tests := []struct {
		desc         string
		input        string
		expectedOk   bool
		expectedName string
	}{
		{
			desc:         "ok: 5 depth",
			input:        "github.com/yuuki/droot/20171017152508/droot",
			expectedOk:   true,
			expectedName: "github.com/yuuki/droot",
		},
		{
			desc:         "ok: 4 depth",
			input:        "github.com/yuuki/droot/20171017152508",
			expectedOk:   true,
			expectedName: "github.com/yuuki/droot",
		},
		{
			desc:         "ng: 3 depth",
			input:        "github.com/yuuki/droot",
			expectedOk:   false,
			expectedName: "",
		},
		{
			desc:         "ng: 2 depth",
			input:        "github.com/yuuki",
			expectedOk:   false,
			expectedName: "",
		},
		{
			desc:         "ng: 1 depth",
			input:        "github.com",
			expectedOk:   false,
			expectedName: "",
		},
		{
			desc:         "ng: empty",
			input:        "",
			expectedOk:   false,
			expectedName: "",
		},
		{
			desc:         "ng: invalid timestamp",
			input:        "github.com/yuuki/droot/2017/droot",
			expectedOk:   false,
			expectedName: "",
		},
	}
	for _, tt := range tests {
		ok, name := ParseName(tt.input)
		if ok != tt.expectedOk {
			t.Errorf("desc: %s, ParseName should be %v to %q", tt.desc, tt.expectedOk, tt.input)
		}
		if name != tt.expectedName {
			t.Errorf("desc: %s, got: %q, want: %q", tt.desc, name, tt.expectedName)
		}
	}
}
