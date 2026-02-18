package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
)

var config = &lib.Config{}
var envFiles = ""
var envPrefix = "REACT_APP_"
var envLoaded = false
var sourceDir = "src"
var entryFileName = "index.tsx"
var outputDir = "build"
var staticDir = "public"
var assetsDir = "assets"
var publicUrl = "/"
var baseDir = "."
var port = 3000
var host = "localhost"
var assetNames = "media/[name]-[hash]"
var chunkNames = "chunks/[name]-[hash]"
var entryNames = "[name]"
var legalComments = "eof"
var jsx = "automatic"
var jsxSideEffects = false
var jsxImportSource = ""
var jsxFactory = ""
var jsxFragment = ""
var sourceMap = "linked"
var customBrowserTarget = ""

var isSecured = false
var certFile, keyFile string

var buildOptions api.BuildOptions

var isHelp = false
var isVersion = false
var useColor = true
var generateMetafile = false
var packagePath = "package.json"
var tsConfigPath = "tsconfig.json"

var versionData = "dev"
var definedReplacements lib.MapFlags

var cliPreloadPathsStartingWith lib.ArrayFlags
var cliInjects lib.ArrayFlags
var cliResolveModules lib.MapFlags
var cliAliasPackages lib.MapFlags
var cliLoaders lib.LoaderFlags
var cliSplitting bool
var cliInlineExtensions lib.ArrayFlags
var cliInlineSize int64 // 100kb

func init() {
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
	// parse flags
	err := ParseFlags(config)
	if err != nil {
		// dont need to print err, CommandLine does that by itself, just exit with error code
		os.Exit(1)
	}

	// show version and exit quickly, no need to do any other checks or work
	if isVersion {
		lib.PrintInfo("NRB version is:", lib.Yellow(lib.Version))
		os.Exit(0)
	}

	// if no output dir provided, fail, we need output dir to build the app, and also to serve in serve mode, so its required
	if outputDir == "" {
		lib.PrintError("failed to find build directory")
		os.Exit(1)
	}

	// if no source dir provided, use current dir
	if sourceDir == "" {
		sourceDir = "."
	}
	if staticDir == "" {
		staticDir = "."
	}

	// parse and run the command
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
