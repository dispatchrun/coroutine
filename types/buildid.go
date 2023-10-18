package types

import (
	"bytes"
	"errors"
	"strconv"
)

// BuildID returns the build identifier for the binary
// this function was called from.
func BuildID() string {
	return buildid
}

func parseBuildID(data []byte) error {
	// From https://github.com/golang/go/blob/3803c858/src/cmd/internal/buildid/buildid.go#L300
	i := bytes.Index(data, buildIDPrefix)
	if i < 0 {
		return errBuildIDNotFound
	}
	j := bytes.Index(data[i+len(buildIDPrefix):], buildIDEnd)
	if j < 0 {
		return errBuildIDNotFound
	}
	quoted := data[i+len(buildIDPrefix)-1 : i+len(buildIDPrefix)+j+1]
	id, err := strconv.Unquote(string(quoted))
	if err != nil {
		return errBuildIDNotFound
	}
	buildid = id
	return nil
}

var (
	buildIDPrefix = []byte("\xff Go build ID: \"")
	buildIDEnd    = []byte("\"\n \xff")

	errBuildIDNotFound = errors.New("build ID not found")

	buildid string
)
