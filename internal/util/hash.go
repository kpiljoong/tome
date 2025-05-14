package util

import (
	"math/rand"
	"os"
	"time"

	"github.com/oklog/ulid/v2"
)

func GenerateULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()
}

func ModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
