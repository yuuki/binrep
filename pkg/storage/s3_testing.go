package storage

import (
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type fakeS3API struct {
	s3API
	FakeGetObject     func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	FakeListObjectsV2 func(*s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error)
	FakePutObject     func(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	FakeDeleteObject  func(*s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
}

type fakeS3UploaderAPI struct {
	FakeUpload func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type fakeS3DownloaderAPI struct {
	FakeDownload func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
}

// GetObject fakes S3 GetObject.
func (s *fakeS3API) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.FakeGetObject(input)
}

// ListObjectsV2 fakes S3 ListObjectsV2.
func (s *fakeS3API) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return s.FakeListObjectsV2(input)
}

// PutObject fakes S3 PutObject.
func (s *fakeS3API) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return s.FakePutObject(input)
}

// DeleteObject fakes S3 DeleteObject.
func (s *fakeS3API) DeleteObject(input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return s.FakeDeleteObject(input)
}

// Upload fakes S3 Upload.
func (u *fakeS3UploaderAPI) Upload(input *s3manager.UploadInput, fn ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return u.FakeUpload(input, fn...)
}

// Download fakes S3 Download.
func (d *fakeS3DownloaderAPI) Download(w io.WriterAt, input *s3.GetObjectInput, fn ...func(*s3manager.Downloader)) (int64, error) {
	return d.FakeDownload(w, input, fn...)
}
