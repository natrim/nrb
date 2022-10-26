package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func run(args []string) {
	jsonFile, err := os.ReadFile(filepath.Join(baseDir, "package.json"))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var packageJson map[string]any
	err = json.Unmarshal(jsonFile, &packageJson)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

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
				os.Exit(1)
			} else {
				fmt.Printf(OK+" Run \"%s\" done.\n", npmRun)
				os.Exit(0)
			}
		} else {
			fmt.Println(ERR, "No script found in package.json")
			os.Exit(1)
		}
	} else {
		fmt.Println(ERR, "No scripts found in package.json")
		os.Exit(1)
	}
}
