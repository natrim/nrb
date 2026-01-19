package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/natrim/nrb/lib"
)

// init in vars.go (flag parsing)

var config = &lib.Config{}

func init() {
	SetupFlags(config)
	if path, err := os.Getwd(); err == nil {
		// escape scripts dir
		if filepath.Base(path) == "scripts" {
			if sourceDir != "" {
				sourceDir = filepath.Join("..", sourceDir)
			}
			if outputDir != "" {
				outputDir = filepath.Join("..", outputDir)
			}
			if staticDir != "" {
				staticDir = filepath.Join("..", staticDir)
			}
			baseDir = ".."
		}
	} else {
		lib.PrintError(err)
		os.Exit(1)
	}
}

func main() {
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		// dont need print err, CommandLine does that
		os.Exit(1)
	}

	lib.UseColor(useColor)

	if flag.NArg() > 1 {
		lib.PrintError("use flags before", lib.Yellow("command"))
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		os.Exit(1)
	}

	if isVersion {
		lib.PrintInfo("NRB version is:", lib.Yellow(lib.Version))
		os.Exit(0)
	}

	if sourceDir == "" {
		sourceDir = "."
	}

	if outputDir == "" {
		lib.PrintError("failed to find build directory")
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "build":
		if err := build(config.PreloadPathsStartingWith); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
	case "watch":
		if err := watch(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
	case "serve":
		if err := serve(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
	default:
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		lib.PrintInfof("use %s with '%s' to build the app, '%s' for watch mode, '%s' to serve build folder and '%s' to show this help\n",
			lib.Yellow("command"), lib.Yellow("build"), lib.Yellow("watch"), lib.Yellow("serve"), lib.Yellow("help"),
		)
		lib.Printe("Flags:")
		flag.PrintDefaults()
	}
	os.Exit(0)
}
