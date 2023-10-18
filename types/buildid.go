package types

// BuildID returns the build identifier for the binary
// this function was called from.
func BuildID() string {
	return buildid
}

var buildid string
