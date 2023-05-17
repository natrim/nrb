package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"github.com/natrim/nrb/lib/plugins"
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
var preloadPathsStartingWith arrayFlags
var resolveModules mapFlags
var buildOptions api.BuildOptions
var aliasPackages mapFlags
var injects arrayFlags
var metaData map[string]any
var legalComments = "eof"
var jsx = "automatic"
var jsxSideEffects = false
var jsxImportSource = ""
var jsxFactory = ""
var jsxFragment = ""

var isBuild = false
var isServe = false
var isMakeCert = false
var isVersion = false
var isVersionUpdate = false
var isWatch = false
var isHelp = false
var useColor = true
var generateMetafile = false
var tsConfigPath = "tsconfig.json"
var versionPath = "version.json"
var npmRun = ""
var customBrowserTarget = ""
var envFiles string

var isSecured = false
var certFile, keyFile string

var inlineSize int64
var inlineExtensions []string

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
var DASH = "–"

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type mapFlags map[string]string

func (i *mapFlags) String() string {
	val := ""
	for a, p := range *i {
		val = val + "," + a + ":" + p
	}
	return strings.TrimPrefix(val, ",")
}
func (i *mapFlags) Set(value string) error {
	alias := strings.SplitN(value, ":", 2)
	(*i)[alias[0]] = alias[1]
	return nil
}

func init() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = func() {
		// nothing, app will print it's stuff
	}
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

	flag.StringVar(&customBrowserTarget, "target", customBrowserTarget, "custom browser target, overrides the one form tsconfig")

	flag.StringVar(&assetNames, "assetNames", assetNames, "asset names schema for esbuild")
	flag.StringVar(&chunkNames, "chunkNames", chunkNames, "chunk names schema for esbuild")
	flag.StringVar(&entryNames, "entryNames", entryNames, "entry names schema for esbuild")

	flag.StringVar(&jsxFactory, "jsxFactory", jsxFactory, "What to use for JSX instead of \"React.createElement\"")
	flag.StringVar(&jsxFragment, "jsxFragment", jsxFragment, "What to use for JSX instead of \"React.Fragment\"")
	flag.StringVar(&jsxImportSource, "jsxImportSource", jsxImportSource, "Override the package name for the automatic runtime (default \"react\")")
	flag.BoolVar(&jsxSideEffects, "jsxSideEffects", jsxSideEffects, "Do not remove unused JSX expressions")
	flag.StringVar(&jsx, "jsx", jsx, "tells esbuild what to do about JSX syntax, available options: automatic|transform|preserve")
	flag.StringVar(&legalComments, "legalComments", legalComments, "what to do with legal comments, available options: none|inline|eof|linked|external")

	flag.Var(&preloadPathsStartingWith, "preload", "paths to module=preload on build, can have multiple flags, ie. --preload=src/index,node_modules/react")
	flag.Var(&resolveModules, "resolve", "resolve package import with 'package:path', can have multiple flags, ie. --resolve=react:packages/super-react/index.js,redux:node_modules/redax/lib/index.js")
	flag.Var(&aliasPackages, "alias", "alias package with another 'package:aliasedpackage', can have multiple flags, ie. --alias=react:preact-compat,react-dom:preact-compat")
	flag.Var(&injects, "inject", "allows you to automatically replace a global variable with an import from another file, can have multiple flags, ie. --inject=./process-shim.js,./react-shim.js")

	flag.BoolVar(&generateMetafile, "metafile", generateMetafile, "generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/")
	flag.StringVar(&tsConfigPath, "tsconfig", tsConfigPath, "path to tsconfig json, relative to current work directory")
	flag.StringVar(&versionPath, "versionfile", versionPath, "path to version.json, relative to current work directory")

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

	if useColor {
		ERR = ShRed + ERR + ShNc
		INFO = ShYellow + INFO + ShNc
		OK = ShGreen + OK + ShNc
		RELOAD = ShBlue + RELOAD + ShNc
		//ITEM = ShWhite+ITEM+ShNc
		DASH = ShBlue + DASH + ShNc
	}

	if flag.NArg() > 1 {
		fmt.Println(ERR, "use flags before", ShYellow+"command"+ShNc)
		fmt.Println(INFO, "Usage:", ShBlue+filepath.Base(os.Args[0])+ShNc, "[flags]", ShYellow+"command"+ShNc)
		os.Exit(1)
	}

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
	case "cert":
		isMakeCert = true
		break
	case "version":
		isVersion = true
		break
	case "version-update":
		isVersionUpdate = true
		break
	case "run":
		npmRun = flag.Arg(1)
		break
	case "help":
		isHelp = true
		break
	}

	isHelp = isHelp || (!isBuild && !isServe && !isMakeCert && !isVersion && !isVersionUpdate && !isWatch && npmRun == "")

	if isHelp {
		fmt.Println(INFO, "Usage:", ShBlue+filepath.Base(os.Args[0])+ShNc, "[flags]", ShYellow+"command"+ShNc)
		fmt.Println(INFO, "use "+ShYellow+"command"+ShNc+" with '"+ShYellow+"build"+ShNc+"' to build the app, '"+ShYellow+"watch"+ShNc+"' for watch mode, '"+ShYellow+"serve"+ShNc+"' to serve build folder, '"+ShYellow+"version-update"+ShNc+"' to update build number, '"+ShYellow+"version"+ShNc+"' for current build version, '"+ShYellow+"cert"+ShNc+"' to make https certificate for watch/serve, '"+ShYellow+"run"+ShNc+"' to run npm scripts and '"+ShYellow+"help"+ShNc+"' to show this help")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if isMakeCert {
		_ = os.RemoveAll(filepath.Join(baseDir, ".cert")) //nuke old dir
		if err := os.Mkdir(filepath.Join(baseDir, ".cert"), 0755); err == nil {
			cmd := exec.Command("mkcert -key-file " + baseDir + "/.cert/key.pem -cert-file " + baseDir + "/.cert/cert.pem '" + host + "'")
			if err := cmd.Run(); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			fmt.Println(ERR, "cannot create \".cert\" dir")
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !lib.FileExists(filepath.Join(baseDir, "package.json")) {
		fmt.Println(ERR, "no", filepath.Join(baseDir, "package.json"), "found")
		os.Exit(1)
	}

	jsonFile, err := os.ReadFile(filepath.Join(baseDir, "package.json"))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var packageJson map[string]any
	err = json.Unmarshal(jsonFile, &packageJson)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	jsonFile = nil

	if npmRun != "" {
		os.Exit(run(packageJson, os.Args[3:]))
	}

	if !lib.FileExists(filepath.Join(staticDir, versionPath)) {
		fmt.Println(ERR, "no", filepath.Join(staticDir, versionPath), "found")
		os.Exit(1)
	}

	jsonFile, err = os.ReadFile(filepath.Join(staticDir, versionPath))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = json.Unmarshal(jsonFile, &metaData)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	jsonFile = nil

	if isVersion {
		os.Exit(version(false))
	}

	if isVersionUpdate {
		os.Exit(version(true))
	}

	// check for alias/resolve/preload/inject options from package.json
	if packageJsonOptions, ok := packageJson["nrb"]; ok {
		options := packageJsonOptions.(map[string]any)

		if alias, ok := options["alias"]; ok {
			if _, ok = alias.(map[string]any); ok {
				aliasPackages = make(mapFlags)
				for name, aliasPath := range alias.(map[string]any) {
					aliasPackages[name] = fmt.Sprintf("%v", aliasPath)
				}
			} else {
				fmt.Println(ERR, "wrong 'alias' key in 'package.json', use object: {package:alias,another:alias}")
				os.Exit(1)
			}
		}
		if resolve, ok := options["resolve"]; ok {
			if _, ok = resolve.(map[string]any); ok {
				resolveModules = make(mapFlags)
				for name, resolvePath := range resolve.(map[string]any) {
					resolveModules[name] = fmt.Sprintf("%v", resolvePath)
				}
			} else {
				fmt.Println(ERR, "wrong 'resolve' key in 'package.json', use object: {package:path,maybenaother:morepath}")
				os.Exit(1)
			}
		}
		if preload, ok := options["preload"]; ok {
			if _, ok = preload.([]any); ok {
				preloadPathsStartingWith = make(arrayFlags, len(preload.([]any)))
				for i, pr := range preload.([]any) {
					preloadPathsStartingWith[i] = fmt.Sprintf("%v", pr)
				}
			} else {
				fmt.Println(ERR, "wrong 'preload' key in 'package.json', use array: [pathtopreload,maybeanotherpath]")
				os.Exit(1)
			}
		}
		if inject, ok := options["inject"]; ok {
			if _, ok = inject.([]any); ok {
				injects = make(arrayFlags, len(inject.([]any)))
				for i, p := range inject.([]any) {
					injects[i] = fmt.Sprintf("%v", p)
				}
			} else {
				fmt.Println(ERR, "wrong 'inject' key in 'package.json', use array: [pathtoinject,maybeanotherpath]")
				os.Exit(1)
			}
		}
		if inline, ok := options["inline"]; ok {
			if inlineSz, ok := inline.(map[string]any)["size"]; ok {
				t, _ := strconv.Atoi(fmt.Sprintf("%v", inlineSz))
				inlineSize = int64(t)
			}
			if inlineExts, ok := inline.(map[string]any)["extensions"]; ok {
				if _, ok = inlineExts.([]any); ok {
					inlineExtensions = make(arrayFlags, len(inlineExts.([]any)))
					for i, pr := range inlineExts.([]any) {
						inlineExtensions[i] = fmt.Sprintf("%v", pr)
					}
				} else {
					fmt.Println(ERR, "wrong 'inline.extensions' key in 'package.json', use array: [jpg,png,other_extension]")
					os.Exit(1)
				}
			}
		}
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
	_ = mime.AddExtensionType(".webmanifest", "applicaton/json")
	_ = mime.AddExtensionType(".webp", "image/webp")
	_ = mime.AddExtensionType(".md", "text/markdown")
	_ = mime.AddExtensionType(".svg", "image/svg+xml")
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".ico", "image/x-icon")

	if isServe {
		os.Exit(serve())
	}

	browserTarget := api.DefaultTarget

	if customBrowserTarget != "" {
		jsonFile, err = os.ReadFile(filepath.Join(baseDir, tsConfigPath))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		} else {
			var tsconfigJson map[string]any
			err = json.Unmarshal(jsonFile, &tsconfigJson)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			jsonFile = nil

			customBrowserTarget = tsconfigJson["compilerOptions"].(map[string]any)["target"].(string)
		}
	}

	if customBrowserTarget != "" {
		switch customBrowserTarget {
		case "ES2015", "es2015", "Es2015":
			browserTarget = api.ES2015
			break
		case "ES2016", "es2016", "Es2016":
			browserTarget = api.ES2016
			break
		case "ES2017", "es2017", "Es2017":
			browserTarget = api.ES2017
			break
		case "ES2018", "es2018", "Es2018":
			browserTarget = api.ES2018
			break
		case "ES2019", "es2019", "Es2019":
			browserTarget = api.ES2019
			break
		case "ES2020", "es2020", "Es2020":
			browserTarget = api.ES2020
			break
		case "ES2021", "es2021", "Es2021":
			browserTarget = api.ES2021
			break
		case "ES2022", "es2022", "Es2022":
			browserTarget = api.ES2022
			break
		case "ESNEXT", "esnext", "ESNext", "ESnext":
			browserTarget = api.ESNext
			break
		case "ES5", "es5", "Es5":
			browserTarget = api.ES5
			break
		case "ES6", "es6", "Es6":
			browserTarget = api.ES2015
			break
		default:
			fmt.Printf(ERR+"Unsupported \"%s\" target!\n", customBrowserTarget)
			os.Exit(1)
		}
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
		Splitting:   true,
		TreeShaking: api.TreeShakingDefault, // default shakes if bundle true, or format iife
		// moved lower to switch via flag
		// LegalComments:     api.LegalCommentsLinked,
		Metafile:          generateMetafile,
		MinifyIdentifiers: !isWatch,
		MinifySyntax:      !isWatch,
		MinifyWhitespace:  !isWatch,
		Write:             true,
		Alias:             aliasPackages,

		Define: makeEnv(),
		Inject: injects,

		Sourcemap: api.SourceMapLinked,
		Tsconfig:  filepath.Join(baseDir, tsConfigPath),

		Plugins: []api.Plugin{
			plugins.AliasPlugin(resolveModules),
			plugins.InlinePlugin(inlineSize, inlineExtensions),
		},

		// react stuff
		// mode is set under this-. JSX: api.JSXAutomatic,
		JSXDev:          isWatch,
		JSXFactory:      jsxFactory,
		JSXFragment:     jsxFragment,
		JSXImportSource: jsxImportSource,
		JSXSideEffects:  jsxSideEffects,
	}

	if jsx == "automatic" {
		buildOptions.JSX = api.JSXAutomatic
	} else if jsx == "transform" {
		buildOptions.JSX = api.JSXTransform
	} else if jsx == "preserve" {
		buildOptions.JSX = api.JSXPreserve
	} else {
		fmt.Println(ERR, "wrong \"--jsx\" mode! (allowed: automatic|tranform|preserve)")
		os.Exit(1)
	}

	if legalComments == "default" {
		buildOptions.LegalComments = api.LegalCommentsDefault
	} else if legalComments == "none" {
		buildOptions.LegalComments = api.LegalCommentsNone
	} else if legalComments == "inline" {
		buildOptions.LegalComments = api.LegalCommentsInline
	} else if legalComments == "eof" {
		buildOptions.LegalComments = api.LegalCommentsEndOfFile
	} else if legalComments == "linked" {
		buildOptions.LegalComments = api.LegalCommentsLinked
	} else if legalComments == "external" {
		buildOptions.LegalComments = api.LegalCommentsExternal
	} else {
		fmt.Println(ERR, "wrong \"--legalComments\" mode! (allowed: none|inline|eof|linked|external)")
		os.Exit(1)
	}

	if useColor {
		buildOptions.Color = api.ColorIfTerminal
	} else {
		buildOptions.Color = api.ColorNever
	}

	if isWatch {
		os.Exit(watch())
	}

	if isBuild {
		os.Exit(build())
	}
}
