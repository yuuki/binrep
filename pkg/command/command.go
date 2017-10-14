package command

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	strftime "github.com/jehiah/go-strftime"
	"github.com/pkg/errors"
)

func init() {
	log.SetFlags(0)
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
