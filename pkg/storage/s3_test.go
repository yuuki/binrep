package storage

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kylelemons/godebug/pretty"
	"github.com/yuuki/binrep/pkg/release"
)

func newTestS3(svc s3API, uploader s3UploaderAPI) *_s3 {
	return &_s3{
		bucket:   "binrep-testing",
		svc:      svc,
		uploader: nil,
	}
}

func TestBuildReleaseURL(t *testing.T) {
	store := newTestS3(&fakeS3API{}, &fakeS3UploaderAPI{})

	u, err := store.buildReleaseURL("github.com/yuuki/droot", "20171016152508")
	if err != nil {
		panic(err)
	}

	if u.String() != "s3://binrep-testing/github.com/yuuki/droot/20171016152508" {
		t.Errorf("got %q, want %q", u.String(), "github.com/yuuki/droot/20171016152508")
	}
}

func TestS3LatestTimestamp(t *testing.T) {
	fakeS3 := &fakeS3API{
		FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
			if *input.Bucket != "binrep-testing" {
				t.Errorf("got %q, want %q", *input.Bucket, "binrep-testing")
			}
			if *input.Prefix != "github.com/yuuki/droot/" {
				t.Errorf("got %q, want %q", *input.Prefix, "github.com/yuuki/droot/")
			}
			return &s3.ListObjectsV2Output{
				CommonPrefixes: []*s3.CommonPrefix{
					{Prefix: aws.String("github.com/yuuki/droot/20171016152508")},
					{Prefix: aws.String("github.com/yuuki/droot/20171017152508")},
					{Prefix: aws.String("github.com/yuuki/droot/20171015152508")},
				},
			}, nil
		},
	}
	store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

	timestamp, err := store.latestTimestamp("github.com/yuuki/droot")

	if err != nil {
		t.Fatalf("should not raise error: %s", err)
	}

	if timestamp != "20171017152508" {
		t.Errorf("got: %q, want %q", timestamp, "20171017152508")
	}
}

func TestS3CreateMeta(t *testing.T) {
	fakeS3 := &fakeS3API{
		FakePutObject: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
			if *input.Bucket != "binrep-testing" {
				t.Errorf("got %q, want %q", *input.Bucket, "binrep-testing")
			}
			expectedKey := "/github.com/yuuki/droot/20171017152508/meta.yml"
			if *input.Key != expectedKey {
				t.Errorf("got %q, want %q", *input.Key, expectedKey)
			}
			body, err := ioutil.ReadAll(input.Body)
			if err != nil {
				panic(err)
			}
			expectedBody := strings.TrimPrefix(`
binaries:
- name: droot
  checksum: ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48
  mode: 493
- name: grabeni
  checksum: 3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259
  mode: 493
`, "\n")
			if diff := pretty.Compare(string(body), expectedBody); diff != "" {
				t.Errorf("diff: (-actual +expected)\n%s", diff)
			}
			return &s3.PutObjectOutput{}, nil
		},
	}
	store := newTestS3(fakeS3, &fakeS3UploaderAPI{})
	u, err := url.Parse("s3://binrep-testing/github.com/yuuki/droot/20171017152508")
	if err != nil {
		panic(err)
	}
	bins := []*release.Binary{
		{
			Name:     "droot",
			Checksum: "ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48",
			Mode:     0755,
			Body:     bytes.NewBufferString("droot-body"),
		},
		{
			Name:     "grabeni",
			Checksum: "3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259",
			Mode:     0755,
			Body:     bytes.NewBufferString("grabeni-body"),
		},
	}

	meta, err := store.createMeta(u, bins)

	if err != nil {
		t.Fatalf("should not raise error: %s", err)
	}

	if diff := pretty.Compare(meta.Binaries, bins); diff != "" {
		t.Errorf("diff: (-actual +expected)\n%s", diff)
	}
}
