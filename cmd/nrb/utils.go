package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/joho/godotenv"
	"github.com/natrim/nrb/lib"
	"github.com/natrim/nrb/lib/plugins"
)

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

func SetupFlags(config *lib.Config) {
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
	flag.StringVar(&sourceMap, "sourceMap", sourceMap, "what sourcemap to use, available options: none|inline|linked|external|both")
	flag.BoolVar(&cliSplitting, "splitting", cliSplitting, "enable code splitting")
	flag.BoolVar(&cliSplitting, "split", cliSplitting, "alias of -splitting")

	flag.Var(&cliPreloadPathsStartingWith, "preload", "paths to module=preload on build, overrides values from package.json, can have multiple flags, ie. --preload=src/index,node_modules/react")
	flag.Var(&cliResolveModules, "resolve", "resolve package import with 'package:path', overrides values from package.json, can have multiple flags, ie. --resolve=react:packages/super-react/index.js,redux:node_modules/redax/lib/index.js")
	flag.Var(&cliAliasPackages, "alias", "alias package with another 'package:aliasedpackage', overrides values from package.json, can have multiple flags, ie. --alias=react:preact-compat,react-dom:preact-compat")
	flag.Var(&cliInjects, "inject", "allows you to automatically replace a global variable with an import from another file, overrides values from package.json, can have multiple flags, ie. --inject=./process-shim.js,./react-shim.js")

	flag.Var(&cliInlineExtensions, "inline", "file extensions to inline as base64 dataurls, overrides values from package.json, ie. --inline=png,jpg,svg")
	flag.Int64Var(&cliInlineSize, "inlineSize", cliInlineSize, "set max file size to inline as base64 dataurls as int in bytes, default is 0 which inlines ALL, overrides values from package.json, ie. for 10kb set --inlineSize=10000")

	flag.BoolVar(&generateMetafile, "metafile", generateMetafile, "generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/")
	flag.StringVar(&tsConfigPath, "tsconfig", tsConfigPath, "path to tsconfig json, relative to current work directory")

	flag.Var(&cliLoaders, "loaders", "esbuild file loaders, overrides values from package.json, ie. --loaders=png:dataurl,.txt:copy,data:json")
}

func SetupWebServer() {
	// register some mime fallbacks
	_ = mime.AddExtensionType(".webmanifest", "application/json")
	_ = mime.AddExtensionType(".webp", "image/webp")
	_ = mime.AddExtensionType(".md", "text/markdown")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".ico", "image/x-icon")

	// check for certs
	if !isSecured && os.Getenv("DEV_SERVER_CERT") != "" {
		if lib.FileExists(filepath.Join(baseDir, os.Getenv("DEV_SERVER_CERT"))) {
			certFile = filepath.Join(baseDir, os.Getenv("DEV_SERVER_CERT"))
			keyFile = filepath.Join(baseDir, os.Getenv("DEV_SERVER_KEY"))
			isSecured = true
		}
	}

	// check for certs in .cert folder
	if !isSecured {
		if lib.FileExists(filepath.Join(baseDir, ".cert/cert.pem")) {
			certFile = filepath.Join(baseDir, ".cert/cert.pem")
			keyFile = filepath.Join(baseDir, ".cert/key.pem")
			isSecured = true
		}
	}
}

func parseEnvVars(isBuildMode bool) (string, string, error) {
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
	if MODE == "" && !isBuildMode {
		MODE = "development"
	} else if MODE == "" && isBuildMode {
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

func buildEsbuildConfig(isBuildMode bool) {
	if !envLoaded {
		envLoaded = true

		mode, env, err := parseEnvVars(isBuildMode)

		if err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		if env != "" {
			lib.PrintInfof("env files: %s\n", env)
		}
		if mode != "" {
			lib.PrintInfof("node mode: \"%s\"\n", mode)
		}
	}

	packageJson, err := lib.ParsePackageJson(filepath.Join(baseDir, packagePath))
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}
	config, err = lib.ParseJsonConfig(packageJson)
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	// override json values by values from cli
	if lib.IsFlagPassed("preload") {
		config.PreloadPathsStartingWith = cliPreloadPathsStartingWith
	}
	if lib.IsFlagPassed("resolve") {
		config.ResolveModules = cliResolveModules
	}
	if lib.IsFlagPassed("alias") {
		config.AliasPackages = cliAliasPackages
	}
	if lib.IsFlagPassed("inject") {
		config.Injects = cliInjects
	}
	if lib.IsFlagPassed("loaders") {
		config.Loaders = cliLoaders
	}
	if lib.IsFlagPassed("inline") {
		config.InlineExtensions = cliInlineExtensions
	}
	if lib.IsFlagPassed("inlineSize") {
		config.InlineSize = cliInlineSize
	}
	if lib.IsFlagPassed("splitting") || lib.IsFlagPassed("split") {
		config.Splitting = cliSplitting
	}

	browserTarget := api.DefaultTarget

	if customBrowserTarget == "" {
		tspath := filepath.Join(baseDir, tsConfigPath)
		if lib.FileExists(tspath) {
			jsonFile, err := os.ReadFile(tspath)
			if err != nil {
				lib.PrintError(err)
			} else {
				var tsconfigJson map[string]any
				err = json.Unmarshal(jsonFile, &tsconfigJson)
				if err != nil {
					lib.PrintError(err)
					os.Exit(1)
				}
				jsonFile = nil

				customBrowserTarget = tsconfigJson["compilerOptions"].(map[string]any)["target"].(string)
			}
		}
	}

	if customBrowserTarget != "" {
		browserTarget, err = lib.ParseBrowserTarget(customBrowserTarget)
		if err != nil {
			lib.Printe(err)
			os.Exit(1)
		}
	}

	if isBuildMode {
		versionData = lib.ParseVersion()
		lib.PrintInfo("app version:", versionData)
	}

	if versionData != "" {
		definedReplacements["process.env."+envPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData)
		definedReplacements["import.meta."+envPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData)
	}

	buildOptions = api.BuildOptions{
		Target:      browserTarget,
		EntryPoints: []string{filepath.Join(sourceDir, entryFileName)},
		Outdir:      filepath.Join(outputDir, assetsDir),
		PublicPath:  fmt.Sprintf("/%s/", assetsDir), // change in index.html too, needs to be same as above
		AssetNames:  assetNames,
		ChunkNames:  chunkNames,
		EntryNames:  entryNames,
		Bundle:      true,
		Format:      api.FormatESModule,
		Splitting:   config.Splitting,
		TreeShaking: api.TreeShakingDefault, // default shakes if bundle true, or format iife
		// moved lower to switch via flag
		// LegalComments:     api.LegalCommentsLinked,
		Metafile:          generateMetafile,
		MinifyIdentifiers: isBuildMode,
		MinifySyntax:      isBuildMode,
		MinifyWhitespace:  isBuildMode,
		Write:             true,
		Alias:             config.AliasPackages,

		Define: definedReplacements,
		Inject: config.Injects,
		Loader: config.Loaders,

		// moved lower to flag
		//Sourcemap: api.SourceMapLinked,

		Tsconfig: filepath.Join(baseDir, tsConfigPath),

		Plugins: []api.Plugin{
			plugins.AliasPlugin(config.ResolveModules),
			plugins.InlinePlugin(config.InlineSize, config.InlineExtensions),
		},

		// react stuff
		// mode is set under this-. JSX: api.JSXAutomatic,
		JSXDev:          !isBuildMode,
		JSXFactory:      jsxFactory,
		JSXFragment:     jsxFragment,
		JSXImportSource: jsxImportSource,
		JSXSideEffects:  jsxSideEffects,
	}

	switch jsx {
	case "automatic":
		buildOptions.JSX = api.JSXAutomatic
	case "transform":
		buildOptions.JSX = api.JSXTransform
	case "preserve":
		buildOptions.JSX = api.JSXPreserve
	default:
		lib.Printe("wrong \"--jsx\" mode! (allowed: automatic|transform|preserve)")
		os.Exit(1)
	}

	switch legalComments {
	case "default":
		buildOptions.LegalComments = api.LegalCommentsDefault
	case "none":
		buildOptions.LegalComments = api.LegalCommentsNone
	case "inline":
		buildOptions.LegalComments = api.LegalCommentsInline
	case "eof":
		buildOptions.LegalComments = api.LegalCommentsEndOfFile
	case "linked":
		buildOptions.LegalComments = api.LegalCommentsLinked
	case "external":
		buildOptions.LegalComments = api.LegalCommentsExternal
	default:
		lib.Printe("wrong \"--legalComments\" mode! (allowed: none|inline|eof|linked|external)")
		os.Exit(1)
	}

	switch sourceMap {
	case "none":
		buildOptions.Sourcemap = api.SourceMapNone
	case "inline":
		buildOptions.Sourcemap = api.SourceMapInline
	case "linked":
		buildOptions.Sourcemap = api.SourceMapLinked
	case "external":
		buildOptions.Sourcemap = api.SourceMapExternal
	case "both":
		buildOptions.Sourcemap = api.SourceMapInlineAndExternal
	default:
		lib.Printe("wrong \"--sourceMap\" value! (allowed: none|inline|linked|external|both)")
		os.Exit(1)
	}

	if useColor {
		buildOptions.Color = api.ColorIfTerminal
	} else {
		buildOptions.Color = api.ColorNever
	}
}
