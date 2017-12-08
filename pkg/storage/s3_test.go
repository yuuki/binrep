package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kylelemons/godebug/pretty"
	"github.com/yuuki/binrep/pkg/release"
)

func newTestS3(svc s3API, uploader s3UploaderAPI) *_s3 {
	return &_s3{
		bucket:   "binrep-testing",
		svc:      svc,
		uploader: uploader,
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

func TestS3ExistRelease(t *testing.T) {
	t.Run("found", func(t *testing.T) {
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

		ok, err := store.ExistRelease("github.com/yuuki/droot")

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		if ok != true {
			t.Error("github.com/yuuki/droot should be found")
		}
	})

}

func TestS3HaveSameChecksums(t *testing.T) {
	fakeListObjects := func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
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
	}

	t.Run("checksum correct", func(t *testing.T) {
		getObjectCallCnt := 0
		fakeS3 := &fakeS3API{
			FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
				return fakeListObjects(input)
			},
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				getObjectCallCnt++
				switch getObjectCallCnt {
				case 1:
					expectedKey := "/github.com/yuuki/droot/20171017152508/meta.yml"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString(strings.TrimPrefix(`
binaries:
- name: droot
  checksum: ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48
  mode: 493
- name: grabeni
  checksum: 3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259
  mode: 493
`, "\n"))),
					}
					return resp, nil
				case 2:
					expectedKey := "/github.com/yuuki/droot/20171017152508/droot"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("droot-body")),
					}
					return resp, nil
				case 3:
					expectedKey := "/github.com/yuuki/droot/20171017152508/grabeni"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("grabeni-body")),
					}
					return resp, nil
				}
				return nil, nil
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

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

		ok, err := store.HaveSameChecksums("github.com/yuuki/droot", bins)

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		if ok != true {
			t.Error("github.com/yuuki/droot checksum should be correct")
		}
	})

	t.Run("checksum error", func(t *testing.T) {
		getObjectCallCnt := 0
		fakeS3 := &fakeS3API{
			FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
				return fakeListObjects(input)
			},
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				getObjectCallCnt++
				switch getObjectCallCnt {
				case 1:
					expectedKey := "/github.com/yuuki/droot/20171017152508/meta.yml"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString(strings.TrimPrefix(`
binaries:
- name: droot
  checksum: 0000000009e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2000000
  mode: 493
- name: grabeni
  checksum: 3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259
  mode: 493
`, "\n"))),
					}
					return resp, nil
				case 2:
					expectedKey := "/github.com/yuuki/droot/20171017152508/droot"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("droot-body")),
					}
					return resp, nil
				case 3:
					expectedKey := "/github.com/yuuki/droot/20171017152508/grabeni"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("grabeni-body")),
					}
					return resp, nil
				}
				return nil, nil
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

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

		ok, err := store.HaveSameChecksums("github.com/yuuki/droot", bins)

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		if ok != false {
			t.Error("github.com/yuuki/droot checksum should be error")
		}
	})
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

func TestS3FindMeta(t *testing.T) {
	callCnt := 0
	fakeS3 := &fakeS3API{
		FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			callCnt++
			switch callCnt {
			case 1:
				expectedKey := "/github.com/yuuki/droot/20171017152508/meta.yml"
				if *input.Key != expectedKey {
					t.Errorf("got %q, want %q", *input.Key, expectedKey)
				}
				resp := &s3.GetObjectOutput{
					Body: ioutil.NopCloser(bytes.NewBufferString(strings.TrimPrefix(`
binaries:
- name: droot
  checksum: ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48
  mode: 493
- name: grabeni
  checksum: 3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259
  mode: 493
`, "\n"))),
				}
				return resp, nil
			case 2:
				expectedKey := "/github.com/yuuki/droot/20171017152508/droot"
				if *input.Key != expectedKey {
					t.Errorf("got %q, want %q", *input.Key, expectedKey)
				}
				resp := &s3.GetObjectOutput{
					Body: ioutil.NopCloser(bytes.NewBufferString("droot-body")),
				}
				return resp, nil
			case 3:
				expectedKey := "/github.com/yuuki/droot/20171017152508/grabeni"
				if *input.Key != expectedKey {
					t.Errorf("got %q, want %q", *input.Key, expectedKey)
				}
				resp := &s3.GetObjectOutput{
					Body: ioutil.NopCloser(bytes.NewBufferString("grabeni-body")),
				}
				return resp, nil
			}
			return nil, nil
		},
	}
	store := newTestS3(fakeS3, &fakeS3UploaderAPI{})
	u, err := url.Parse("s3://binrep-testing/github.com/yuuki/droot/20171017152508")
	if err != nil {
		panic(err)
	}

	meta, err := store.FindMeta(u)

	if err != nil {
		t.Fatalf("should not raise error: %s", err)
	}

	expected := []*release.Binary{
		{
			Name:     "droot",
			Checksum: "ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48",
			Mode:     0755,
			Body:     ioutil.NopCloser(bytes.NewBufferString("droot-body")),
		},
		{
			Name:     "grabeni",
			Checksum: "3e30f16f0ec41ab92ceca57a527efff18b6bacabd12a842afda07b8329e32259",
			Mode:     0755,
			Body:     ioutil.NopCloser(bytes.NewBufferString("grabeni-body")),
		},
	}
	if diff := pretty.Compare(meta.Binaries, expected); diff != "" {
		t.Errorf("diff: (-actual +expected)\n%s", diff)
	}
}

func TestS3FindLatestRelease(t *testing.T) {
	fakeListObjects := func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
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
	}
	t.Run("normal", func(t *testing.T) {
		getObjectCallCnt := 0
		fakeS3 := &fakeS3API{
			FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
				return fakeListObjects(input)
			},
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				getObjectCallCnt++
				switch getObjectCallCnt {
				case 1:
					expectedKey := "/github.com/yuuki/droot/20171017152508/meta.yml"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString(strings.TrimPrefix(`
binaries:
- name: droot
  checksum: ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48
  mode: 493
`, "\n"))),
					}
					return resp, nil
				case 2:
					expectedKey := "/github.com/yuuki/droot/20171017152508/droot"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("droot-body")),
					}
					return resp, nil
				}
				return nil, nil
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

		rel, err := store.FindLatestRelease("github.com/yuuki/droot")

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		if rel.Name() != "github.com/yuuki/droot" {
			t.Errorf("got: %q, want %q", rel.Name(), "github.com/yuuki/droot")
		}
		if rel.Timestamp() != "20171017152508" {
			t.Errorf("got: %q, want %q", rel.Name(), "20171017152508")
		}
	})

	t.Run("meta.yml not found", func(t *testing.T) {
		fakeS3 := &fakeS3API{
			FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
				return fakeListObjects(input)
			},
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				return nil, awserr.New("NoSuchKey", "", nil)
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

		_, err := store.FindLatestRelease("github.com/yuuki/droot")

		if err == nil {
			t.Fatal("should raise error")
		}
		if !strings.Contains(fmt.Sprintf("%s", err), "meta.yml not found") {
			t.Errorf("error got: %q, want: %q", err, "meta.yml not found")
		}
	})
}

func TestS3FindReleaseByTimestamp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		getObjectCallCnt := 0
		fakeS3 := &fakeS3API{
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				getObjectCallCnt++
				switch getObjectCallCnt {
				case 1:
					expectedKey := "/github.com/yuuki/droot/20171016152508/meta.yml"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString(strings.TrimPrefix(`
binaries:
- name: droot
  checksum: ec9efb6249e0e4797bde75afbfe962e0db81c530b5bb1cfd2cbe0e2fc2c8cf48
  mode: 493
`, "\n"))),
					}
					return resp, nil
				case 2:
					expectedKey := "/github.com/yuuki/droot/20171016152508/droot"
					if *input.Key != expectedKey {
						t.Errorf("got %q, want %q", *input.Key, expectedKey)
					}
					resp := &s3.GetObjectOutput{
						Body: ioutil.NopCloser(bytes.NewBufferString("droot-body")),
					}
					return resp, nil
				}
				return nil, nil
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

		rel, err := store.FindReleaseByTimestamp("github.com/yuuki/droot", "20171016152508")

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		if rel.Name() != "github.com/yuuki/droot" {
			t.Errorf("got: %q, want %q", rel.Name(), "github.com/yuuki/droot")
		}
		if rel.Timestamp() != "20171016152508" {
			t.Errorf("got: %q, want %q", rel.Name(), "20171016152508")
		}
	})

	t.Run("meta.yml not found", func(t *testing.T) {
		fakeS3 := &fakeS3API{
			FakeGetObject: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				return nil, awserr.New("NoSuchKey", "", nil)
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

		_, err := store.FindReleaseByTimestamp("github.com/yuuki/droot", "20001016152508")

		if err == nil {
			t.Fatal("should raise error")
		}
		if !strings.Contains(fmt.Sprintf("%s", err), "meta.yml not found") {
			t.Errorf("error got: %q, want: %q", err, "meta.yml not found")
		}
	})
}

func TestS3CreateRelease(t *testing.T) {
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
	callCnt := 0
	fakeS3Uploader := &fakeS3UploaderAPI{
		FakeUpload: func(input *s3manager.UploadInput, fn ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			callCnt++
			switch callCnt {
			case 1:
				if *input.Bucket != "binrep-testing" {
					t.Errorf("got %q, want %q", *input.Bucket, "binrep-testing")
				}
				expectedKey := "/github.com/yuuki/droot/20171017152508/droot"
				if *input.Key != expectedKey {
					t.Errorf("got %q, want %q", *input.Key, expectedKey)
				}
				return &s3manager.UploadOutput{}, nil
			case 2:
				if *input.Bucket != "binrep-testing" {
					t.Errorf("got %q, want %q", *input.Bucket, "binrep-testing")
				}
				expectedKey := "/github.com/yuuki/droot/20171017152508/grabeni"
				if *input.Key != expectedKey {
					t.Errorf("got %q, want %q", *input.Key, expectedKey)
				}
				return &s3manager.UploadOutput{}, nil
			}
			return nil, nil
		},
	}
	store := newTestS3(fakeS3, fakeS3Uploader)
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

	rel, err := store.CreateRelease("github.com/yuuki/droot", "20171017152508", bins)

	if err != nil {
		t.Fatalf("should not raise error: %s", err)
	}

	if rel.Name() != "github.com/yuuki/droot" {
		t.Errorf("got: %q, want: %q", rel.Timestamp(), "20171017152508")
	}
	if rel.Timestamp() != "20171017152508" {
		t.Errorf("got: %q, want: %q", rel.Timestamp(), "20171017152508")
	}
	if diff := pretty.Compare(rel.Meta.Binaries, bins); diff != "" {
		t.Errorf("diff: (-actual +expected)\n%s", diff)
	}
}

func TestS3ascTimestamps(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
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

		timestamps, err := store.ascTimestamps("github.com/yuuki/droot")

		if err != nil {
			t.Fatalf("should not raise error: %s", err)
		}

		expected := []string{
			"20171015152508",
			"20171016152508",
			"20171017152508",
		}
		if diff := pretty.Compare(timestamps, expected); diff != "" {
			t.Errorf("diff: (-actual +expected)\n%s", diff)
		}
	})
	t.Run("zero length error", func(t *testing.T) {
		fakeS3 := &fakeS3API{
			FakeListObjectsV2: func(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
				if *input.Bucket != "binrep-testing" {
					t.Errorf("got %q, want %q", *input.Bucket, "binrep-testing")
				}
				if *input.Prefix != "github.com/yuuki/droot/" {
					t.Errorf("got %q, want %q", *input.Prefix, "github.com/yuuki/droot/")
				}
				return &s3.ListObjectsV2Output{CommonPrefixes: []*s3.CommonPrefix{}}, nil
			},
		}
		store := newTestS3(fakeS3, &fakeS3UploaderAPI{})

		_, err := store.ascTimestamps("github.com/yuuki/droot")

		if err == nil {
			t.Fatalf("should raise error: %s", err)
		}
		if !strings.Contains(err.Error(), "no such projects") {
			t.Errorf("got: %q, want: %q", err.Error(), "no such projects")
		}
	})
}
