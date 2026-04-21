package lib

import (
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

// Version will contain nrb build number on build
var Version = "dev"

func init() {
	Version = fixBuild(Version)
}

// ParseVersion will use git to get current revision to use as version string
func ParseVersion() string {
	versionDataCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	if versionDataB, err := versionDataCmd.Output(); err == nil {
		return strings.TrimSpace(string(versionDataB))
	}

	return time.Now().Format("v20060102150405")
}

// try to get itnernal build name
func fixBuild(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v
	}

	if strings.HasPrefix(info.Main.Version, "v") {
		return info.Main.Version
	}

	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			v = setting.Value
			return v
		}
	}

	// If we didn't find vcs.revision, try the old key for backward compatibility
	for _, setting := range info.Settings {
		if setting.Key == "gitrevision" {
			v = setting.Value
			return v
		}
	}

	return v
}
