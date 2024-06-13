package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/natrim/nrb/lib"
)

func makeCertificate() error {
	if !lib.CommandExists("mkcert") {
		return errors.New("you need to have \"mkcert\" binary installed (ie. \"brew install mkcert\" or \"choco install mkcert\" or \"npm install -g mkcert\")")
	}
	_ = os.RemoveAll(filepath.Join(baseDir, ".cert"))
	if err := os.Mkdir(filepath.Join(baseDir, ".cert"), 0755); err == nil {
		cmd := exec.Command("mkcert -key-file " + baseDir + "/.cert/key.pem -cert-file " + baseDir + "/.cert/cert.pem '" + host + "'")
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		return errors.New("cannot create \".cert\" dir")
	}

	return nil
}
