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

var config = &Config{}

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

	if sourceDir == "" {
		sourceDir = "."
	}

	if outputDir == "" {
		lib.PrintError("failed to find build directory")
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

	if npmRun != "" {
		packageJson, err := parsePackageJson()
		if err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
		if err := runNpmScript(packageJson, os.Args[3:]); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		lib.PrintOkf(" Run \"%s\" done.\n", npmRun)
		os.Exit(0)
	}

	if isVersionGet {
		if err := version(nil, false); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if isVersionUpdate {
		if err := version(nil, true); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

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

	if isServe {
		if err := serve(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	mode, env, err := makeEnv()

	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	buildEsbuildConfig()

	if env != "" {
		lib.PrintInfof("env: %s\n", env)
	}
	if mode != "" {
		lib.PrintInfof("mode: \"%s\"\n", mode)
	}

	if isWatch {
		if err := watch(); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if isBuild {
		if err := build(config.PreloadPathsStartingWith); err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func buildEsbuildConfig() {
	packageJson, err := parsePackageJson()
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}
	config, err = parseJsonConfig(packageJson)
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	// override by values from cli
	if len(cliPreloadPathsStartingWith) > 0 {
		config.PreloadPathsStartingWith = cliPreloadPathsStartingWith
	}
	if len(cliResolveModules) > 0 {
		config.ResolveModules = cliResolveModules
	}
	if len(cliAliasPackages) > 0 {
		config.AliasPackages = cliAliasPackages
	}
	if len(cliInjects) > 0 {
		config.Injects = cliInjects
	}
	if len(cliLoaders) > 0 {
		config.Loaders = cliLoaders
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

	versionData, err := parseVersionData()
	if err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	if versionData == nil {
		lib.PrintItem("No versionData available")
	} else {
		definedReplacements["process.env."+envPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData["version"])
		definedReplacements["import.meta."+envPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData["version"])
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
		lib.Printe("wrong \"--jsx\" mode! (allowed: automatic|transform|preserve)")
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
		lib.Printe("wrong \"--legalComments\" mode! (allowed: none|inline|eof|linked|external)")
		os.Exit(1)
	}

	if sourceMap == "none" {
		buildOptions.Sourcemap = api.SourceMapNone
	} else if sourceMap == "inline" {
		buildOptions.Sourcemap = api.SourceMapInline
	} else if sourceMap == "linked" {
		buildOptions.Sourcemap = api.SourceMapLinked
	} else if sourceMap == "external" {
		buildOptions.Sourcemap = api.SourceMapExternal
	} else if sourceMap == "both" {
		buildOptions.Sourcemap = api.SourceMapInlineAndExternal
	} else {
		lib.Printe("wrong \"--sourceMap\" value! (allowed: none|inline|linked|external|both)")
		os.Exit(1)
	}

	if useColor {
		buildOptions.Color = api.ColorIfTerminal
	} else {
		buildOptions.Color = api.ColorNever
	}
}
