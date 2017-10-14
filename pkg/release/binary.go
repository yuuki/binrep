package release

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

type Binary struct {
	Name     string `yaml:"name"`
	Checksum string `yaml:"checksum"`
	Version  string `yaml:"version,omitempty"`
	Body     io.Reader
}

func BuildBinary(name string, body io.Reader) (*Binary, error) {
	sum, err := checksum(body)
	if err != nil {
		return nil, err
	}
	return &Binary{
		Name:     name,
		Checksum: sum,
	}, nil
}

func checksum(r io.Reader) (string, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		errors.Errorf("failed to read data for checksum")
	}
	return fmt.Sprintf("%x", sha1.Sum(body)), nil
}

func (b *Binary) ValidateChecksum(r io.Reader) error {
	sum, err := checksum(r)
	if err != nil {
		return err
	}
	if b.Checksum != sum {
		return errors.Errorf("invalid checksum, got %v, want %v", sum, b.Checksum)
	}
	return nil
}
