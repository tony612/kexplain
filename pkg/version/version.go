package version

import "fmt"

var (
	gitCommit string
	version   string
)

func GitCommit() string {
	if gitCommit == "" {
		return "unknown"
	}
	return gitCommit
}

func Version() string {
	if version == "" {
		return "unknown"
	}
	return version
}

func FullVersion() string {
	return fmt.Sprintf("%s (%s)", Version(), GitCommit())
}
