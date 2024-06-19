package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/natrim/nrb/lib"
)

func version(versionData VersionData, update bool) error {
	versionFilePath := filepath.Join(staticDir, versionPath)
	if versionData == nil {
		var err error
		if versionData, err = parseVersionData(); versionData == nil {
			errr := errors.New("you need to have " + versionFilePath)
			if err != nil {
				return errors.Join(errr, err)
			}
			return errr
		}
	}

	if update {
		lib.PrintInfo("Incrementing build number...")
		v, _ := strconv.Atoi(fmt.Sprintf("%v", versionData["version"]))
		versionData["version"] = v + 1

		j, err := json.Marshal(versionData)
		if err != nil {
			return err
		}

		err = os.WriteFile(versionFilePath, j, 0644)
		if err != nil {
			return err
		}

		lib.PrintOk("App version has been updated")
	}

	lib.PrintOk("Current version number is:", versionData["version"])
	return nil
}
