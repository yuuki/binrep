package command

import (
	"log"
	"time"

	strftime "github.com/jehiah/go-strftime"
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
