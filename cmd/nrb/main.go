package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"github.com/natrim/nrb/lib/plugins"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var envPrefix = "REACT_APP_"
var sourceDir = "src"
var entryFileName = "index.tsx"
var outputDir = "build"
var staticDir = "public"
var assetsDir = "assets"
var publicUrl = "/"
var baseDir = "."
var wwwPort = 3000
var host = "localhost"
var preloadPathsStartingWith = []string{"node_modules/.pnpm/react@", "node_modules/react/", "src/core/index", "src/index"}

var buildOptions api.BuildOptions
var metaData map[string]any

var isBuild = false
var isServe = false
var isTest = false
var isMakeCert = false
var isVersion = false
var isVersionUpdate = false
var isWatch = false
var isHelp = false

var npmRun = ""
var envFiles string

var isSecured = false
var certFile, keyFile string

var ShRed = "\033[0;31m"
var ShGreen = "\033[0;32m"
var ShYellow = "\033[0;33m"
var ShBlue = "\033[0;34m"
var ShNc = "\033[0m"

var ERR = ShRed + "×" + ShNc
var INFO = ShYellow + ">" + ShNc
var OK = ShGreen + "✓" + ShNc
var RELOAD = ShBlue + "↻" + ShNc
var ITEM = "-"

func init() {
	flag.BoolVar(&isWatch, "w", isWatch, "watch mode")
	flag.BoolVar(&isBuild, "b", isBuild, "build mode")
	flag.BoolVar(&isServe, "s", isServe, "serve mode")
	flag.BoolVar(&isTest, "t", isTest, "test mode")
	flag.BoolVar(&isMakeCert, "c", isMakeCert, "make cert")
	flag.BoolVar(&isHelp, "h", isVersion, "help")
	flag.BoolVar(&isVersion, "v", isHelp, "app version")
	flag.BoolVar(&isVersionUpdate, "u", isVersionUpdate, "app version update")
	flag.StringVar(&npmRun, "r", npmRun, "npm run but faster")
	flag.StringVar(&envFiles, "env", envFiles, "env files to load from (always loads .env for fallback, no overriding)")

	if envFiles == "" && lib.FileExists(filepath.Join(baseDir, ".env.local")) {
		envFiles = ".env.local"
	}

	if path, err := os.Getwd(); err == nil {
		if filepath.Base(path) == "scripts" {
			sourceDir = "../" + sourceDir
			outputDir = "../" + outputDir
			staticDir = "../" + staticDir
			baseDir = ".."
		}
	} else {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if !lib.FileExists(filepath.Join(baseDir, "package.json")) {
		fmt.Println(ERR, "no", filepath.Join(staticDir, "version.json"), "found")
		os.Exit(1)
	}

	if !lib.FileExists(filepath.Join(staticDir, "version.json")) {
		fmt.Println(ERR, "no", filepath.Join(staticDir, "version.json"), "found")
		os.Exit(1)
	}

	flag.Parse()
	isHelp = isHelp || (!isBuild && !isServe && !isTest && !isMakeCert && !isVersion && !isVersionUpdate && !isWatch && npmRun == "")

	if isHelp {
		fmt.Println(ERR, "No help defined")
		fmt.Println(INFO, "just kidding, run this script with -b to build the app, -w for watch mode, -s to serve build folder, -u to update build number, -t for test's, -v for current build version, -c to make https certificate for dev, -env to set custom .env files, -r to run npm scripts, and -h for this help")
		os.Exit(0)
	}

	if isTest {
		fmt.Println(ERR, "No test's defined")
		os.Exit(1)
	}

	if npmRun != "" {
		run()
		os.Exit(0)
	}

	if isMakeCert {
		os.RemoveAll(filepath.Join(baseDir, ".cert"))
		os.Mkdir(filepath.Join(baseDir, ".cert"), 0755)
		cmd := exec.Command("mkcert -key-file " + baseDir + "/.cert/key.pem -cert-file " + baseDir + "/.cert/cert.pem '" + host + "'")
		if err := cmd.Run(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	jsonFile, err := os.ReadFile(filepath.Join(staticDir, "version.json"))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = json.Unmarshal(jsonFile, &metaData)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if isVersion {
		fmt.Println(OK, "Current version number is:", metaData["version"])
		os.Exit(0)
	}

	if isVersionUpdate {
		fmt.Println(INFO, "Incrementing build number...")
		v, _ := strconv.Atoi(fmt.Sprintf("%v", metaData["version"]))
		metaData["version"] = v + 1

		j, err := json.Marshal(metaData)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		err = os.WriteFile(filepath.Join(staticDir, "version.json"), j, 0644)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println(OK, "App version has been updated")
		fmt.Println(OK, "Current version number is:", metaData["version"])
		os.Exit(0)
	}

	if !isSecured && os.Getenv("DEV_SERVER_CERT") != "" {
		if lib.FileExists(filepath.Join(baseDir, os.Getenv("DEV_SERVER_CERT"))) {
			certFile = filepath.Join(baseDir, os.Getenv("DEV_SERVER_CERT"))
			keyFile = filepath.Join(baseDir, os.Getenv("DEV_SERVER_KEY"))
			isSecured = true
		}
	}

	if !isSecured {
		if lib.FileExists(filepath.Join(baseDir, ".cert/cert.pem")) {
			certFile = filepath.Join(baseDir, ".cert/cert.pem")
			keyFile = filepath.Join(baseDir, ".cert/key.pem")
			isSecured = true
		}
	}

	// mime fallbacks
	mime.AddExtensionType(".webmanifest", "applicaton/json")
	mime.AddExtensionType(".webp", "image/webp")
	mime.AddExtensionType(".md", "text/markdown")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".ico", "image/x-icon")

	if isServe {
		serve()
		os.Exit(0)
	}

	buildOptions = api.BuildOptions{
		EntryPoints:       []string{filepath.Join(sourceDir, entryFileName)},
		Outdir:            filepath.Join(outputDir, assetsDir),
		PublicPath:        fmt.Sprintf("/%s/", assetsDir), // change in index.html too, needs to be same as above
		AssetNames:        "media/[name]-[hash]",
		ChunkNames:        "chunks/[name]-[hash]",
		EntryNames:        "[name]", // change in index.html too, js and css
		Bundle:            true,
		Format:            api.FormatESModule,
		Splitting:         true,
		TreeShaking:       api.TreeShakingDefault, // default shakes if bundle true, or format iife
		Sourcemap:         api.SourceMapLinked,
		LegalComments:     api.LegalCommentsLinked,
		MinifyIdentifiers: !isWatch,
		MinifySyntax:      !isWatch,
		MinifyWhitespace:  !isWatch,
		Write:             true,

		Define: makeEnv(),

		Plugins: []api.Plugin{
			// fix old mui
			plugins.AliasPlugin(map[string]string{
				"@material-ui/pickers": lib.RealQuickPath(filepath.Join(baseDir, "node_modules/@material-ui/pickers/dist/material-ui-pickers.js")),
				"@material-ui/core":    lib.RealQuickPath(filepath.Join(baseDir, "node_modules/@material-ui/core/index.js")),
			}),
			plugins.InlinePluginDefault(),
		},

		// react stuff
		JSXMode: api.JSXModeAutomatic,
		JSXDev:  isWatch,
	}

	if isWatch {
		watch()
		os.Exit(0)
	}

	if isBuild {
		build()
		os.Exit(0)
	}
}
