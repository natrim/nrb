package main

import (
	"flag"
	"mime"
	"os"

	"github.com/evanw/esbuild/pkg/api"
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
var assetNames = "media/[name]-[hash]"
var chunkNames = "chunks/[name]-[hash]"
var entryNames = "[name]"
var legalComments = "eof"
var jsx = "automatic"
var jsxSideEffects = false
var jsxImportSource = ""
var jsxFactory = ""
var jsxFragment = ""
var splitting = false
var customBrowserTarget = ""
var envFiles string
var isSecured = false
var certFile, keyFile string

var buildOptions api.BuildOptions

var isBuild = false
var isServe = false
var isMakeCert = false
var isVersionGet = false
var isVersionUpdate = false
var isWatch = false
var isHelp = false
var isVersion = false
var useColor = true
var generateMetafile = false
var packagePath = "package.json"
var tsConfigPath = "tsconfig.json"
var versionPath = "version.json"
var npmRun = ""

func SetupFlags() {
	// now start settings flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = func() {
		// nothing, app will print it's stuff
	}
	flag.BoolVar(&isVersion, "version", isVersion, "nrb version number")
	flag.BoolVar(&isVersion, "v", isVersion, "alias of -version")
	flag.BoolVar(&isHelp, "h", isHelp, "alias of -help")
	flag.BoolVar(&isHelp, "help", isHelp, "this help")
	flag.StringVar(&envFiles, "env", envFiles, "env files to load from (always loads .env first)")

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

	flag.StringVar(&customBrowserTarget, "target", customBrowserTarget, "custom browser target, defaults to tsconfig target if possible, else esnext")

	flag.StringVar(&assetNames, "assetNames", assetNames, "asset names schema for esbuild")
	flag.StringVar(&chunkNames, "chunkNames", chunkNames, "chunk names schema for esbuild")
	flag.StringVar(&entryNames, "entryNames", entryNames, "entry names schema for esbuild")

	flag.StringVar(&jsxFactory, "jsxFactory", jsxFactory, "What to use for JSX instead of \"React.createElement\"")
	flag.StringVar(&jsxFragment, "jsxFragment", jsxFragment, "What to use for JSX instead of \"React.Fragment\"")
	flag.StringVar(&jsxImportSource, "jsxImportSource", jsxImportSource, "Override the package name for the automatic runtime (default \"react\")")
	flag.BoolVar(&jsxSideEffects, "jsxSideEffects", jsxSideEffects, "Do not remove unused JSX expressions")
	flag.StringVar(&jsx, "jsx", jsx, "tells esbuild what to do about JSX syntax, available options: automatic|transform|preserve")
	flag.StringVar(&legalComments, "legalComments", legalComments, "what to do with legal comments, available options: none|inline|eof|linked|external")
	flag.BoolVar(&splitting, "splitting", splitting, "enable code splitting")
	flag.BoolVar(&splitting, "split", splitting, "alias of -splitting")

	//FIXME: redo after global config stuff is done
	//	flag.Var(&config.PreloadPathsStartingWith, "preload", "paths to module=preload on build, can have multiple flags, ie. --preload=src/index,node_modules/react")
	//	flag.Var(&config.ResolveModules, "resolve", "resolve package import with 'package:path', can have multiple flags, ie. --resolve=react:packages/super-react/index.js,redux:node_modules/redax/lib/index.js")
	//	flag.Var(&config.AliasPackages, "alias", "alias package with another 'package:aliasedpackage', can have multiple flags, ie. --alias=react:preact-compat,react-dom:preact-compat")
	//	flag.Var(&config.Injects, "inject", "allows you to automatically replace a global variable with an import from another file, can have multiple flags, ie. --inject=./process-shim.js,./react-shim.js")

	flag.BoolVar(&generateMetafile, "metafile", generateMetafile, "generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/")
	flag.StringVar(&tsConfigPath, "tsconfig", tsConfigPath, "path to tsconfig json, relative to current work directory")
	flag.StringVar(&versionPath, "versionfile", versionPath, "path to version.json, relative to current work directory")
}

func SetupMime() {
	// register some mime fallbacks
	_ = mime.AddExtensionType(".webmanifest", "applicaton/json")
	_ = mime.AddExtensionType(".webp", "image/webp")
	_ = mime.AddExtensionType(".md", "text/markdown")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".ico", "image/x-icon")
}
