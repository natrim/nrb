package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func version(update bool) int {
	if update {
		fmt.Println(INFO, "Incrementing build number...")
		v, _ := strconv.Atoi(fmt.Sprintf("%v", metaData["version"]))
		metaData["version"] = v + 1

		j, err := json.Marshal(metaData)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		err = os.WriteFile(filepath.Join(staticDir, versionPath), j, 0644)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 1
		}

		fmt.Println(OK, "App version has been updated")
		fmt.Println(OK, "Current version number is:", metaData["version"])
		return 0
	} else {
		fmt.Println(OK, "Current version number is:", metaData["version"])
		return 0
	}
}
