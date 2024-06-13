package main

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func runNpmScript(packageJson PackageJson, args []string) error {
	if scripts, ok := packageJson["scripts"]; ok {
		if script, ok := scripts.(map[string]any)[npmRun]; ok {
			args = append([]string{"-c"}, strings.Join(append([]string{script.(string)}, args...), " "))
			cmd := exec.Command("bash", args...)
			if runtime.GOOS == "windows" {
				cmd = exec.Command("bash.exe", args...)
			}
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			} else {
				return nil
			}
		} else {
			return errors.New("no script found in package.json")
		}
	} else {
		return errors.New("no scripts found in package.json")
	}
}
