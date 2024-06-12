package main

import (
	"fmt"
	"github.com/natrim/nrb/lib"
	"os"
	"os/exec"
	"path/filepath"
)

func mkcert() int {
	if !lib.CommandExists("mkcert") {
		fmt.Println(ERR, "you need to have \"mkcert\" binary installed (ie. \"brew install mkcert\" or \"choco install mkcert\" or \"npm install -g mkcert\")")
		return 1
	}
	_ = os.RemoveAll(filepath.Join(baseDir, ".cert"))
	if err := os.Mkdir(filepath.Join(baseDir, ".cert"), 0755); err == nil {
		cmd := exec.Command("mkcert -key-file " + baseDir + "/.cert/key.pem -cert-file " + baseDir + "/.cert/cert.pem '" + host + "'")
		if err := cmd.Run(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		fmt.Println(ERR, "cannot create \".cert\" dir")
		return 1
	}

	return 0
}
