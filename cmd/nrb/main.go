package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
)

var config = func() *lib.Config {
	cfg := lib.DefaultConfig()
	return &cfg
}()

var envFiles = ""
var envLoaded = false
var baseDir = "."

var isSecured = false
var certFile, keyFile string

var buildOptions api.BuildOptions

var isHelp = false
var isVersion = false
var useColor = true
var packagePath = "package.json"

var versionData = "dev"
var definedReplacements lib.MapFlags

var currentConfigOverrides lib.ConfigOverrides

func main() {
	cliState, configOverrides, err := ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	isHelp = cliState.IsHelp
	isVersion = cliState.IsVersion
	useColor = cliState.UseColor
	envFiles = cliState.EnvFiles
	currentConfigOverrides = configOverrides

	if isVersion {
		lib.PrintInfo("NRB version is:", lib.Yellow(lib.Version))
		os.Exit(0)
	}

	if path, err := os.Getwd(); err == nil {
		if filepath.Base(path) == "scripts" {
			baseDir = ".."
		}
	} else {
		lib.PrintError(err)
		os.Exit(1)
	}

	command := flag.Arg(0)
	switch command {
	case "build", "watch":
		if err := refreshRuntimeConfig(true); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
		if command == "watch" {
			if err := watch(); err != nil {
				lib.PrintError(err)
				os.Exit(1)
			}
		} else {
			if err := build(); err != nil {
				lib.PrintError(err)
				os.Exit(1)
			}
		}
	case "serve":
		if err := refreshRuntimeConfig(false); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
		if err := serve(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
	default:
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		lib.PrintInfof(
			"use %s with '%s' to build the app, '%s' for watch mode, '%s' to serve build folder and '%s' to show this help\n",
			lib.Yellow("command"), lib.Yellow("build"), lib.Yellow("watch"), lib.Yellow("serve"), lib.Yellow("help"),
		)
		lib.Printe("Flags:")
		flag.PrintDefaults()
	}
}
