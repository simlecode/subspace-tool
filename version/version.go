package version

var (
	CurrentCommit string

	BuildVersion = "0.0.0"

	Version = BuildVersion + CurrentCommit
)
