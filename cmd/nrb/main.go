package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"github.com/natrim/nrb/lib/plugins"
)

// init in vars.go (flag parsing)

func init() {
	SetupFlags()
	if path, err := os.Getwd(); err == nil {
		// escape scripts dir
		if filepath.Base(path) == "scripts" {
			sourceDir = filepath.Join("..", sourceDir)
			outputDir = filepath.Join("..", outputDir)
			staticDir = filepath.Join("..", staticDir)
			baseDir = ".."
		}
	} else {
		lib.PrintError(err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	lib.UseColor(useColor)

	if flag.NArg() > 1 {
		lib.PrintError("use flags before", lib.Yellow("command"))
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "build":
		isBuild = true
	case "watch":
		isWatch = true
	case "serve":
		isServe = true
	case "cert":
		isMakeCert = true
	case "version":
		isVersionGet = true
	case "version-update":
		isVersionUpdate = true
	case "run":
		npmRun = flag.Arg(1)
	case "help":
		isHelp = true
	}

	isHelp = isHelp || (!isVersion && !isBuild && !isServe && !isMakeCert && !isVersionGet && !isVersionUpdate && !isWatch && npmRun == "")

	if isServe || isWatch {
		SetupMime()
	}

	if isHelp {
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		lib.PrintInfof("use %s with '%s' to build the app, '%s' for watch mode, '%s' to serve build folder, '%s' to update build number, '%s' for current build version, '%s' to make https certificate for watch/serve, '%s' to run npm scripts and '%s' to show this help\n",
			lib.Yellow("command"), lib.Yellow("build"), lib.Yellow("watch"), lib.Yellow("serve"), lib.Yellow("version-update"), lib.Yellow("version"), lib.Yellow("cert"), lib.Yellow("run"), lib.Yellow("help"),
		)
		lib.Printe("Flags:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if isVersion {
		lib.PrintInfo("NRB version is:", lib.Yellow(lib.Version))
		os.Exit(0)
	}

	if isMakeCert {
		if err := makeCertificate(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	packageJson, err := parsePackageJson()
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	if npmRun != "" {
		if err := runNpmScript(packageJson, os.Args[3:]); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		lib.PrintOkf(" Run \"%s\" done.\n", npmRun)
		os.Exit(0)
	}

	versionData, err := parseVersionData()
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	if isVersionGet {
		if err := version(versionData, false); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if isVersionUpdate {
		if err := version(versionData, true); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	config, err := parseJsonConfig(packageJson)
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
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

	if isServe {
		if err := serve(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	browserTarget := api.DefaultTarget

	if customBrowserTarget != "" {
		jsonFile, err := os.ReadFile(filepath.Join(baseDir, tsConfigPath))
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

	if customBrowserTarget != "" {
		switch customBrowserTarget {
		case "ES2015", "es2015", "Es2015":
			browserTarget = api.ES2015
		case "ES2016", "es2016", "Es2016":
			browserTarget = api.ES2016
		case "ES2017", "es2017", "Es2017":
			browserTarget = api.ES2017
		case "ES2018", "es2018", "Es2018":
			browserTarget = api.ES2018
		case "ES2019", "es2019", "Es2019":
			browserTarget = api.ES2019
		case "ES2020", "es2020", "Es2020":
			browserTarget = api.ES2020
		case "ES2021", "es2021", "Es2021":
			browserTarget = api.ES2021
		case "ES2022", "es2022", "Es2022":
			browserTarget = api.ES2022
		case "ES2023", "es2023", "Es2023":
			browserTarget = api.ES2023
		case "ESNEXT", "esnext", "ESNext", "ESnext":
			browserTarget = api.ESNext
		case "ES5", "es5", "Es5":
			browserTarget = api.ES5
		case "ES6", "es6", "Es6":
			browserTarget = api.ES2015
		default:
			lib.PrintError("unsupported target", customBrowserTarget)
			os.Exit(1)
		}
	}

	definedReplacements, err := makeEnv(versionData)
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
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
		Splitting:   splitting,
		TreeShaking: api.TreeShakingDefault, // default shakes if bundle true, or format iife
		// moved lower to switch via flag
		// LegalComments:     api.LegalCommentsLinked,
		Metafile:          generateMetafile,
		MinifyIdentifiers: !isWatch,
		MinifySyntax:      !isWatch,
		MinifyWhitespace:  !isWatch,
		Write:             true,
		Alias:             config.AliasPackages,

		Define: definedReplacements,
		Inject: config.Injects,

		Sourcemap: api.SourceMapLinked,
		Tsconfig:  filepath.Join(baseDir, tsConfigPath),

		Plugins: []api.Plugin{
			plugins.AliasPlugin(config.ResolveModules),
			plugins.InlinePlugin(config.InlineSize, config.InlineExtensions),
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
		lib.PrintError("wrong \"--jsx\" mode! (allowed: automatic|transform|preserve)")
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
		lib.PrintError("wrong \"--legalComments\" mode! (allowed: none|inline|eof|linked|external)")
		os.Exit(1)
	}

	if useColor {
		buildOptions.Color = api.ColorIfTerminal
	} else {
		buildOptions.Color = api.ColorNever
	}

	if isWatch {
		if err := watch(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if isBuild {
		if err := build(&config); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}
}
