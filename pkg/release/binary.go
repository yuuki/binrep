package release

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

const (
	shortCheckSumLen int = 7
)

// Binary represents the binary file within release.
type Binary struct {
	Name     string    `yaml:"name"`
	Checksum string    `yaml:"checksum"`
	Version  string    `yaml:"version,omitempty"`
	Body     io.Reader `yaml:"-"`
}

// BuildBinary builds a Binary object. Return error if it is failed
// to calculate checksum of the body.
func BuildBinary(name string, body io.Reader) (*Binary, error) {
	sum, err := checksum(body)
	if err != nil {
		return nil, err
	}
	return &Binary{
		Name:     name,
		Checksum: sum,
		Body:     body,
	}, nil
}

func checksum(r io.Reader) (string, error) {
	if r == nil {
		return "", errors.New("try to read nil")
	}
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return "", errors.New("failed to read data for checksum")
	}
	return fmt.Sprintf("%x", sha256.Sum256(body)), nil
}

// InvalidChecksumError represents an error of the checksum.
type InvalidChecksumError struct {
	got  string
	want string
}

// Error returns the error message for InvalidChecksumError.
func (e *InvalidChecksumError) Error() string {
	return fmt.Sprintf("got: %s, want: %s", e.got, e.want)
}

// IsChecksumError returns that the type of err matches InvalidChecksumError type or not.
func IsChecksumError(err error) bool {
	_, ok := errors.Cause(err).(*InvalidChecksumError)
	return ok
}

// CopyAndValidateChecksum copies src to dst and calculate checksum of src, then check it.
func (b *Binary) CopyAndValidateChecksum(dst io.Writer, src io.Reader) (int64, error) {
	h := sha256.New()
	w := io.MultiWriter(h, dst)

	written, err := io.Copy(w, src)
	if err != nil {
		return written, err
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	if b.Checksum != sum {
		return written, errors.WithStack(&InvalidChecksumError{got: sum, want: b.Checksum})
	}

	return written, nil
}

func (b *Binary) shortChecksum() string {
	return b.Checksum[0:shortCheckSumLen]
}

// Inspect prints the binary information.
func (b *Binary) Inspect(w io.Writer) {
	fmt.Fprintf(w, "%s/%s/%s\t", b.Name, b.Version, b.shortChecksum())
}
