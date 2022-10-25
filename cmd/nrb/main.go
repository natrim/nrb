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
var port = 3000
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
var useColor = true

var npmRun = ""
var envFiles string

var isSecured = false
var certFile, keyFile string

const (
	ShRed    = "\033[0;31m"
	ShGreen  = "\033[0;32m"
	ShYellow = "\033[0;33m"
	ShBlue   = "\033[0;34m"
	ShNc     = "\033[0m"
)

var ERR = "×"
var INFO = ">"
var OK = "✓"
var RELOAD = "↻"
var ITEM = "-"

func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = func() {
		// nothing, app will print it's stuff
	}
	//	flag.BoolVar(&isWatch, "w", isWatch, "watch mode")
	//	flag.BoolVar(&isWatch, "watch", isWatch, "alias of -w")
	//	flag.BoolVar(&isBuild, "b", isBuild, "build mode")
	//	flag.BoolVar(&isBuild, "build", isBuild, "alias of -b")
	//	flag.BoolVar(&isServe, "s", isServe, "serve mode")
	//	flag.BoolVar(&isServe, "serve", isServe, "alias of -s")
	//	flag.BoolVar(&isTest, "t", isTest, "test mode")
	//	flag.BoolVar(&isTest, "test", isTest, "test mode")
	//	flag.BoolVar(&isMakeCert, "c", isMakeCert, "make cert")
	flag.BoolVar(&isHelp, "h", isVersion, "alias of -help")
	flag.BoolVar(&isHelp, "help", isVersion, "this help")
	//	flag.BoolVar(&isVersion, "v", isHelp, "app version")
	//	flag.BoolVar(&isVersion, "version", isHelp, "alias of -v")
	//	flag.BoolVar(&isVersionUpdate, "u", isVersionUpdate, "app version update")
	//	flag.StringVar(&npmRun, "r", npmRun, "npm run but faster")
	//	flag.StringVar(&npmRun, "run", npmRun, "alias of -r")
	flag.StringVar(&envFiles, "env", envFiles, "env files to load from (always loads .env for fallback, no overriding)")

	flag.BoolVar(&useColor, "color", useColor, "colorize output")

	flag.StringVar(&envPrefix, "envPrefix", envPrefix, "env variables prefix")
	flag.StringVar(&sourceDir, "sourceDir", sourceDir, "source directory name")
	flag.StringVar(&entryFileName, "entryFileName", entryFileName, "entry file name in 'sourceDir'")
	flag.StringVar(&outputDir, "outputDir", outputDir, "output dir name")
	flag.StringVar(&staticDir, "staticDir", staticDir, "static dir name")
	flag.StringVar(&assetsDir, "assetsDir", assetsDir, "assets dir name in output")
	flag.IntVar(&port, "port", port, "port")
	flag.StringVar(&host, "host", host, "host")
	flag.StringVar(&publicUrl, "publicUrl", publicUrl, "public url")

	if envFiles == "" && lib.FileExists(filepath.Join(baseDir, ".env.local")) {
		envFiles = ".env.local"
	}

	if path, err := os.Getwd(); err == nil {
		// escape scripts dir
		if filepath.Base(path) == "scripts" {
			sourceDir = filepath.Join("..", sourceDir)
			outputDir = filepath.Join("..", outputDir)
			staticDir = filepath.Join("..", staticDir)
			baseDir = ".."
		}
	} else {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "build":
		isBuild = true
		break
	case "watch":
		isWatch = true
		break
	case "serve":
		isServe = true
		break
	case "test":
		isTest = true
		break
	case "cert":
		isMakeCert = true
		break
	case "version":
		isVersion = true
		break
	case "update":
		isVersionUpdate = true
		break
	case "run":
		npmRun = flag.Arg(1)
		break
	case "help":
		isHelp = true
		break
	}

	isHelp = isHelp || (!isBuild && !isServe && !isTest && !isMakeCert && !isVersion && !isVersionUpdate && !isWatch && npmRun == "")

	if useColor {
		ERR = ShRed + ERR + ShNc
		INFO = ShYellow + INFO + ShNc
		OK = ShGreen + OK + ShNc
		RELOAD = ShBlue + RELOAD + ShNc
		//ITEM = ShWhite+ITEM+ShNc
	}

	if isHelp {
		fmt.Println(ERR, "No help defined")
		fmt.Println(INFO, "just kidding, run this command with '"+ShYellow+"build"+ShNc+"' to build the app, '"+ShYellow+"watch"+ShNc+"' for watch mode, '"+ShYellow+"serve"+ShNc+"' to serve build folder, '"+ShYellow+"update"+ShNc+"' to update build number, '"+ShYellow+"test"+ShNc+"' for test's, '"+ShYellow+"version"+ShNc+"' for current build version, '"+ShYellow+"cert"+ShNc+"' to make https certificate for dev, '"+ShYellow+"run"+ShNc+"' to run npm scripts")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !lib.FileExists(filepath.Join(baseDir, "package.json")) {
		fmt.Println(ERR, "no", filepath.Join(staticDir, "version.json"), "found")
		os.Exit(1)
	}

	if !lib.FileExists(filepath.Join(staticDir, "version.json")) {
		fmt.Println(ERR, "no", filepath.Join(staticDir, "version.json"), "found")
		os.Exit(1)
	}

	if isTest {
		//TODO: try to run npm test
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
