package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/natrim/nrb/lib"
	"os"
	"path/filepath"
	"strings"
)

func makeEnv() map[string]string {
	envFiles = strings.Join(strings.Fields(strings.Trim(envFiles, ",")), "")
	if lib.FileExists(filepath.Join(baseDir, ".env")) {
		if envFiles != "" {
			envFiles = ".env," + envFiles
		} else {
			envFiles = ".env"
		}
	}
	if envFiles != "" {
		fmt.Printf(INFO+" loading .env file/s: %s\n", envFiles)
		err := godotenv.Overload(strings.Split(envFiles, ",")...)
		if err != nil {
			fmt.Println(ERR, "Error loading .env file/s:", err)
			os.Exit(1)
		}
	}

	var MODE = os.Getenv("NODE_ENV")
	if MODE == "" && isWatch {
		MODE = "development"
	} else if MODE == "" && !isWatch {
		MODE = "production"
	}

	fmt.Printf(INFO+" mode: \"%s\"\n", MODE)

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
		"process.env." + envPrefix + "VERSION": fmt.Sprintf("\"%v\"", metaData["version"]),
		"import.meta." + envPrefix + "VERSION": fmt.Sprintf("\"%v\"", metaData["version"]),
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

	return define
}
