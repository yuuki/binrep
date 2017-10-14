package meta

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	strftime "github.com/jehiah/go-strftime"
	"github.com/pkg/errors"
)

type Binary struct {
	Name      string `yaml:"name"`
	Checksum  string `yaml:"checksum"`
	Timestamp string `yaml:"timestamp"`
	Version   string `yaml:"version,omitempty"`
}

type Meta struct {
	Binaries []*Binary `yaml:"binaries"`
}

func New(b *Binary) *Meta {
	return &Meta{Binaries: []*Binary{b}}
}

func (m *Meta) AppendBinary(b *Binary) {
	m.Binaries = append(m.Binaries, b)
}

func BuildBinary(f *os.File, name string) (*Binary, error) {
	sum, err := checksum(f)
	if err != nil {
		return nil, err
	}
	return &Binary{
		Name:      name,
		Checksum:  sum,
		Timestamp: now(),
	}, nil
}

func now() string {
	t := time.Now()
	utc, _ := time.LoadLocation("UTC")
	t = t.In(utc)
	return strftime.Format("%Y%m%d%H%M%S", t)
}

func checksum(f *os.File) (string, error) {
	body, err := ioutil.ReadAll(f)
	if err != nil {
		errors.Errorf("failed to read %v", f.Name())
	}
	return fmt.Sprintf("%x", sha1.Sum(body)), nil
}
