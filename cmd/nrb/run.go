package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func run(packageJson map[string]any, args []string) int {
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
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 1
			} else {
				fmt.Printf(OK+" Run \"%s\" done.\n", npmRun)
				return 0
			}
		} else {
			fmt.Println(ERR, "No script found in package.json")
			return 1
		}
	} else {
		fmt.Println(ERR, "No scripts found in package.json")
		return 1
	}
}
