package lib

import (
	"os/exec"
	"strings"
	"time"
)

// Version will contain nrb build number on build
var Version = "dev"

// ParseVersion will use git to get current revision to use as version string
func ParseVersion() string {
	versionDataCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if versionDataB, err := versionDataCmd.Output(); err == nil {
		return strings.TrimSpace(string(versionDataB))
	}

	return time.Now().Format("v20060102150405")
}
