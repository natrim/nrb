package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/natrim/nrb/lib"
)

var definedReplacements mapFlags

func makeEnv() (string, string, error) {
	envFiles := strings.Join(strings.Fields(strings.Trim(envFiles, ",")), "")
	if lib.FileExists(filepath.Join(baseDir, ".env")) {
		if envFiles != "" {
			envFiles = ".env," + envFiles
		} else {
			envFiles = ".env"
		}
	}
	if envFiles != "" {
		err := godotenv.Overload(strings.Split(envFiles, ",")...)
		if err != nil {
			return "", "", errors.Join(errors.New("cannot load .env file/s"), err)
		}
	}

	var MODE = os.Getenv("NODE_ENV")
	if MODE == "" && isWatch {
		MODE = "development"
	} else if MODE == "" && !isWatch {
		MODE = "production"
	}

	isDevelopment := "false"
	isProduction := "false"
	if MODE == "development" {
		isDevelopment = "true"
		isProduction = "false"
	} else {
		isDevelopment = "false"
		isProduction = "true"
	}

	define := map[string]string{
		// libs fallback
		"process.env.NODE_ENV": fmt.Sprintf("\"%s\"", MODE),

		// cra fallback
		"process.env.FAST_REFRESH": "false",
		"process.env.PUBLIC_URL":   fmt.Sprintf("\"%s\"", strings.TrimSuffix(publicUrl, "/")),

		// import.meta stuff
		"import.meta.env.MODE":     fmt.Sprintf("\"%s\"", MODE),
		"import.meta.env.BASE_URL": fmt.Sprintf("\"%s\"", strings.TrimSuffix(publicUrl, "/")),
		"import.meta.env.PROD":     isProduction,
		"import.meta.env.DEV":      isDevelopment,

		// metaData version
		"process.env." + envPrefix + "VERSION": fmt.Sprintf("\"%v\"", "\"dev\""),
		"import.meta." + envPrefix + "VERSION": fmt.Sprintf("\"%v\"", "\"dev\""),
	}

	envAll := os.Environ()
	for _, v := range envAll {
		env := strings.SplitN(v, "=", 2)
		if strings.HasPrefix(env[0], envPrefix) {
			define[fmt.Sprintf("process.env.%s", env[0])] = fmt.Sprintf("\"%s\"", env[1])
			define[fmt.Sprintf("import.meta.%s", env[0])] = fmt.Sprintf("\"%s\"", env[1])
		}
	}

	// fallback missing
	define["process.env"] = "{}"
	define["import.meta"] = "{}"

	definedReplacements = define

	return MODE, envFiles, nil
}
