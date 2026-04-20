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

type CLIState struct {
	IsHelp    bool
	IsVersion bool
	UseColor  bool
	EnvFiles  string
}

func ParseFlags() (CLIState, lib.ConfigOverrides, error) {
	state := CLIState{}
	overrides := lib.ConfigOverrides{}
	defaults := *config

	// start settings flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = func() {
		// nothing, app will print it's stuff
	}

	isVersionFlag := false
	isHelpFlag := false
	useColorFlag := true
	envFilesFlag := ""

	envPrefixFlag := defaults.EnvPrefix
	sourceDirFlag := defaults.SourceDir
	entryFileNameFlag := defaults.EntryFileName
	outputDirFlag := defaults.OutputDir
	staticDirFlag := defaults.StaticDir
	assetsDirFlag := defaults.AssetsDir
	portFlag := defaults.Port
	hostFlag := defaults.Host
	publicURLFlag := defaults.PublicURL
	customBrowserTargetFlag := defaults.Target
	assetNamesFlag := defaults.AssetNames
	chunkNamesFlag := defaults.ChunkNames
	entryNamesFlag := defaults.EntryNames
	jsxFactoryFlag := defaults.JSXFactory
	jsxFragmentFlag := defaults.JSXFragment
	jsxImportSourceFlag := defaults.JSXImportSource
	jsxSideEffectsFlag := defaults.JSXSideEffects
	jsxFlag := lib.JSXString(defaults.JSX)
	legalCommentsFlag := lib.LegalCommentsString(defaults.LegalComments)
	sourceMapFlag := lib.SourceMapString(defaults.SourceMap)
	splittingFlag := defaults.Splitting
	generateMetafileFlag := defaults.Metafile
	tsConfigPathFlag := defaults.TSConfigPath
	var preloadFlag lib.ArrayFlags
	var resolveFlag lib.MapFlags
	var aliasFlag lib.MapFlags
	var injectFlag lib.ArrayFlags
	var inlineFlag lib.ArrayFlags
	inlineSizeFlag := defaults.InlineSize
	var loadersFlag lib.LoaderFlags

	flag.BoolVar(&isVersionFlag, "version", isVersionFlag, "nrb version number")
	flag.BoolVar(&isVersionFlag, "v", isVersionFlag, "alias of -version")
	flag.BoolVar(&isHelpFlag, "h", isHelpFlag, "alias of -help")
	flag.BoolVar(&isHelpFlag, "help", isHelpFlag, "this help")
	flag.StringVar(&envFilesFlag, "env", envFilesFlag, "env files to load from (always loads .env first)")

	flag.BoolVar(&useColorFlag, "color", useColorFlag, "colorize output")

	flag.StringVar(&envPrefixFlag, "envPrefix", envPrefixFlag, "env variables prefix")
	flag.StringVar(&sourceDirFlag, "sourceDir", sourceDirFlag, "source directory name")
	flag.StringVar(&entryFileNameFlag, "entryFileName", entryFileNameFlag, "entry file name in 'sourceDir'")
	flag.StringVar(&outputDirFlag, "outputDir", outputDirFlag, "output dir name")
	flag.StringVar(&staticDirFlag, "staticDir", staticDirFlag, "static dir name")
	flag.StringVar(&assetsDirFlag, "assetsDir", assetsDirFlag, "assets dir name in output")
	flag.IntVar(&portFlag, "port", portFlag, "port")
	flag.StringVar(&hostFlag, "host", hostFlag, "host")
	flag.StringVar(&publicURLFlag, "publicUrl", publicURLFlag, "public url")

	flag.StringVar(&customBrowserTargetFlag, "target", customBrowserTargetFlag, "custom browser target, defaults to tsconfig target if possible, else esnext")

	flag.StringVar(&assetNamesFlag, "assetNames", assetNamesFlag, "asset names schema for esbuild")
	flag.StringVar(&chunkNamesFlag, "chunkNames", chunkNamesFlag, "chunk names schema for esbuild")
	flag.StringVar(&entryNamesFlag, "entryNames", entryNamesFlag, "entry names schema for esbuild")

	flag.StringVar(&jsxFactoryFlag, "jsxFactory", jsxFactoryFlag, "What to use for JSX instead of \"React.createElement\"")
	flag.StringVar(&jsxFragmentFlag, "jsxFragment", jsxFragmentFlag, "What to use for JSX instead of \"React.Fragment\"")
	flag.StringVar(&jsxImportSourceFlag, "jsxImportSource", jsxImportSourceFlag, "Override the package name for the automatic runtime (default \"react\")")
	flag.BoolVar(&jsxSideEffectsFlag, "jsxSideEffects", jsxSideEffectsFlag, "Do not remove unused JSX expressions")
	flag.StringVar(&jsxFlag, "jsx", jsxFlag, "tells esbuild what to do about JSX syntax, available options: automatic|transform|preserve")
	flag.StringVar(&legalCommentsFlag, "legalComments", legalCommentsFlag, "what to do with legal comments, available options: none|inline|eof|linked|external")
	flag.StringVar(&sourceMapFlag, "sourceMap", sourceMapFlag, "what sourcemap to use, available options: none|inline|linked|external|both")
	flag.BoolVar(&splittingFlag, "splitting", splittingFlag, "enable code splitting")
	flag.BoolVar(&splittingFlag, "split", splittingFlag, "alias of -splitting")

	flag.Var(&preloadFlag, "preload", "paths to module=preload on build, overrides values from package.json, can have multiple flags, ie. --preload=src/index,node_modules/react")
	flag.Var(&resolveFlag, "resolve", "resolve package import with 'package:path', overrides values from package.json, can have multiple flags, ie. --resolve=react:packages/super-react/index.js,redux:node_modules/redax/lib/index.js")
	flag.Var(&aliasFlag, "alias", "alias package with another 'package:aliasedpackage', overrides values from package.json, can have multiple flags, ie. --alias=react:preact-compat,react-dom:preact-compat")
	flag.Var(&injectFlag, "inject", "allows you to automatically replace a global variable with an import from another file, overrides values from package.json, can have multiple flags, ie. --inject=./process-shim.js,./react-shim.js")

	flag.Var(&inlineFlag, "inline", "file extensions to inline as base64 dataurls, overrides values from package.json, ie. --inline=png,jpg,svg")
	flag.Int64Var(&inlineSizeFlag, "inlineSize", inlineSizeFlag, "set max file size to inline as base64 dataurls as int in bytes, default is 0 which inlines ALL, overrides values from package.json, ie. for 10kb set --inlineSize=10000")

	flag.BoolVar(&generateMetafileFlag, "metafile", generateMetafileFlag, "generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/")
	flag.StringVar(&tsConfigPathFlag, "tsconfig", tsConfigPathFlag, "path to tsconfig json, relative to current work directory")

	flag.Var(&loadersFlag, "loaders", "esbuild file loaders, overrides values from package.json, ie. --loaders=png:dataurl,.txt:copy,data:json")

	// parse flags
	err := flag.CommandLine.Parse(os.Args[1:])

	if err != nil {
		return state, overrides, err
	}

	state = CLIState{
		IsHelp:    isHelpFlag,
		IsVersion: isVersionFlag,
		UseColor:  useColorFlag,
		EnvFiles:  envFilesFlag,
	}

	// set color output before any output
	lib.UseColor(state.UseColor)

	// handle too many arguments, only flags and one command allowed
	if flag.NArg() > 1 {
		lib.PrintError("use flags before", lib.Yellow("command"))
		lib.PrintInfo("Usage:", lib.Blue(filepath.Base(os.Args[0])), "[flags]", lib.Yellow("command"))
		return state, overrides, errors.New("too many arguments, only one command allowed")
	}

	passedFlags := collectPassedFlags(flag.CommandLine)

	if passedFlags["envPrefix"] {
		overrides.EnvPrefix = lib.OptionalString{Value: envPrefixFlag, Set: true}
	}
	if passedFlags["sourceDir"] {
		overrides.SourceDir = lib.OptionalString{Value: sourceDirFlag, Set: true}
	}
	if passedFlags["entryFileName"] {
		overrides.EntryFileName = lib.OptionalString{Value: entryFileNameFlag, Set: true}
	}
	if passedFlags["outputDir"] {
		overrides.OutputDir = lib.OptionalString{Value: outputDirFlag, Set: true}
	}
	if passedFlags["staticDir"] {
		overrides.StaticDir = lib.OptionalString{Value: staticDirFlag, Set: true}
	}
	if passedFlags["assetsDir"] {
		overrides.AssetsDir = lib.OptionalString{Value: assetsDirFlag, Set: true}
	}
	if passedFlags["port"] {
		overrides.Port = lib.OptionalInt{Value: portFlag, Set: true}
	}
	if passedFlags["host"] {
		overrides.Host = lib.OptionalString{Value: hostFlag, Set: true}
	}
	if passedFlags["publicUrl"] {
		overrides.PublicURL = lib.OptionalString{Value: publicURLFlag, Set: true}
	}
	if passedFlags["target"] {
		overrides.Target = lib.OptionalString{Value: customBrowserTargetFlag, Set: true}
	}
	if passedFlags["assetNames"] {
		overrides.AssetNames = lib.OptionalString{Value: assetNamesFlag, Set: true}
	}
	if passedFlags["chunkNames"] {
		overrides.ChunkNames = lib.OptionalString{Value: chunkNamesFlag, Set: true}
	}
	if passedFlags["entryNames"] {
		overrides.EntryNames = lib.OptionalString{Value: entryNamesFlag, Set: true}
	}
	if passedFlags["jsxFactory"] {
		overrides.JSXFactory = lib.OptionalString{Value: jsxFactoryFlag, Set: true}
	}
	if passedFlags["jsxFragment"] {
		overrides.JSXFragment = lib.OptionalString{Value: jsxFragmentFlag, Set: true}
	}
	if passedFlags["jsxImportSource"] {
		overrides.JSXImportSource = lib.OptionalString{Value: jsxImportSourceFlag, Set: true}
	}
	if passedFlags["jsxSideEffects"] {
		overrides.JSXSideEffects = lib.OptionalBool{Value: jsxSideEffectsFlag, Set: true}
	}
	if passedFlags["jsx"] {
		jsxMode, err := lib.ParseJSX(jsxFlag)
		if err != nil {
			lib.Printe(err)
			os.Exit(1)
		}
		overrides.JSX = lib.OptionalEnum[api.JSX]{Value: jsxMode, Set: true}
	}
	if passedFlags["legalComments"] {
		legalCommentsMode, err := lib.ParseLegalComments(legalCommentsFlag)
		if err != nil {
			lib.Printe(err)
			os.Exit(1)
		}
		overrides.LegalComments = lib.OptionalEnum[api.LegalComments]{Value: legalCommentsMode, Set: true}
	}
	if passedFlags["sourceMap"] {
		sourceMapMode, err := lib.ParseSourceMap(sourceMapFlag)
		if err != nil {
			lib.Printe(err)
			os.Exit(1)
		}
		overrides.SourceMap = lib.OptionalEnum[api.SourceMap]{Value: sourceMapMode, Set: true}
	}
	if passedFlags["metafile"] {
		overrides.Metafile = lib.OptionalBool{Value: generateMetafileFlag, Set: true}
	}
	if passedFlags["tsconfig"] {
		overrides.TSConfigPath = lib.OptionalString{Value: tsConfigPathFlag, Set: true}
	}
	if passedFlags["splitting"] || passedFlags["split"] {
		overrides.Splitting = lib.OptionalBool{Value: splittingFlag, Set: true}
	}
	if passedFlags["alias"] {
		overrides.AliasPackages = aliasFlag
	}
	if passedFlags["resolve"] {
		overrides.ResolveModules = resolveFlag
	}
	if passedFlags["preload"] {
		overrides.PreloadPathsStartingWith = preloadFlag
	}
	if passedFlags["inject"] {
		overrides.Injects = injectFlag
	}
	if passedFlags["inline"] {
		overrides.InlineExtensions = inlineFlag
	}
	if passedFlags["inlineSize"] {
		overrides.InlineSize = lib.OptionalInt64{Value: inlineSizeFlag, Set: true}
	}
	if passedFlags["loaders"] {
		overrides.Loaders = loadersFlag
	}

	return state, overrides, nil
}

func collectPassedFlags(flagSet *flag.FlagSet) map[string]bool {
	passedFlags := make(map[string]bool)
	flagSet.Visit(func(f *flag.Flag) {
		passedFlags[f.Name] = true
	})
	return passedFlags
}

func buildRuntimeConfig(requirePackageJSON bool) (lib.Config, error) {
	mergedConfig := lib.DefaultConfig()
	packageFilePath := filepath.Join(baseDir, packagePath)

	if !lib.FileExists(packageFilePath) {
		if requirePackageJSON {
			return lib.Config{}, errors.New("no " + packageFilePath + " found")
		}
	} else {
		packageJson, err := lib.ParsePackageJson(packageFilePath)
		if err != nil {
			return lib.Config{}, err
		}
		configPatch, err := lib.ParseJsonConfig(packageJson)
		if err != nil {
			return lib.Config{}, err
		}
		mergedConfig = lib.MergeConfig(mergedConfig, configPatch)
	}

	lib.ApplyOverrides(&mergedConfig, configOverrides)

	return mergedConfig, nil
}

func normalizeRuntimeConfig(cfg lib.Config) lib.Config {
	if cfg.SourceDir != "" {
		cfg.SourceDir = filepath.Join(baseDir, cfg.SourceDir)
	}
	if cfg.OutputDir != "" {
		cfg.OutputDir = filepath.Join(baseDir, cfg.OutputDir)
	}
	if cfg.StaticDir != "" {
		cfg.StaticDir = filepath.Join(baseDir, cfg.StaticDir)
	}

	if cfg.SourceDir == "" {
		cfg.SourceDir = "."
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "."
	}

	return cfg
}

func refreshRuntimeConfig(requirePackageJSON bool) error {
	cfg, err := buildRuntimeConfig(requirePackageJSON)
	if err != nil {
		return err
	}

	cfg = normalizeRuntimeConfig(cfg)
	config = &cfg

	if config.OutputDir == "" {
		return errors.New("failed to find build directory")
	}

	return nil
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

func resolveEnvFiles() string {
	envPaths := strings.Join(strings.Fields(strings.Trim(cliState.EnvFiles, ",")), "")
	if lib.FileExists(filepath.Join(baseDir, ".env")) {
		if envPaths != "" {
			return ".env," + envPaths
		}
		return ".env"
	}
	return envPaths
}

func loadEnvFiles() (string, error) {
	envPaths := resolveEnvFiles()
	if envPaths == "" {
		return "", nil
	}

	if err := godotenv.Overload(strings.Split(envPaths, ",")...); err != nil {
		return "", errors.Join(errors.New("cannot load .env file/s"), err)
	}

	return envPaths, nil
}

func buildDefinedReplacements(cfg lib.Config, isBuildMode bool) string {
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
		"process.env.PUBLIC_URL":   fmt.Sprintf("\"%s\"", strings.TrimSuffix(cfg.PublicURL, "/")),

		// import.meta stuff
		"import.meta.env.MODE":     fmt.Sprintf("\"%s\"", MODE),
		"import.meta.env.BASE_URL": fmt.Sprintf("\"%s\"", strings.TrimSuffix(cfg.PublicURL, "/")),
		"import.meta.env.PROD":     isProduction,
		"import.meta.env.DEV":      isDevelopment,

		// metaData version
		"process.env." + cfg.EnvPrefix + "VERSION": fmt.Sprintf("\"%v\"", "\"dev\""),
		"import.meta." + cfg.EnvPrefix + "VERSION": fmt.Sprintf("\"%v\"", "\"dev\""),
	}

	envAll := os.Environ()
	for _, v := range envAll {
		env := strings.SplitN(v, "=", 2)
		if strings.HasPrefix(env[0], cfg.EnvPrefix) {
			define[fmt.Sprintf("process.env.%s", env[0])] = fmt.Sprintf("\"%s\"", env[1])
			define[fmt.Sprintf("import.meta.%s", env[0])] = fmt.Sprintf("\"%s\"", env[1])
		}
	}

	// fallback missing
	define["process.env"] = "{}"
	define["import.meta"] = "{}"

	definedReplacements = define

	return MODE
}

var envLoaded bool

func buildEsbuildConfig(isBuildMode bool) {
	if err := refreshRuntimeConfig(true); err != nil {
		lib.PrintError(err)
		os.Exit(1)
	}

	if !envLoaded {
		envLoaded = true

		env, err := loadEnvFiles()
		if err != nil {
			lib.PrintError(err)
			os.Exit(1)
		}

		if env != "" {
			lib.PrintInfof("env files: %s\n", env)
		}
	}

	mode := buildDefinedReplacements(*config, isBuildMode)
	if mode != "" {
		lib.PrintInfof("node mode: \"%s\"\n", mode)
	}

	browserTarget := api.DefaultTarget
	target := config.Target

	if target == "" {
		tspath := filepath.Join(baseDir, config.TSConfigPath)
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

				target = tsconfigJson["compilerOptions"].(map[string]any)["target"].(string)
			}
		}
	}

	if target != "" {
		var err error
		browserTarget, err = lib.ParseBrowserTarget(target)
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
		definedReplacements["process.env."+config.EnvPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData)
		definedReplacements["import.meta."+config.EnvPrefix+"VERSION"] = fmt.Sprintf("\"%v\"", versionData)
	}

	apiColor := api.ColorIfTerminal
	if !cliState.UseColor {
		apiColor = api.ColorNever
	}

	buildOptions = api.BuildOptions{
		Color:             apiColor,
		Target:            browserTarget,
		EntryPoints:       []string{filepath.Join(config.SourceDir, config.EntryFileName)},
		Outdir:            filepath.Join(config.OutputDir, config.AssetsDir),
		PublicPath:        fmt.Sprintf("/%s/", config.AssetsDir), // change in index.html too, needs to be same as above
		AssetNames:        config.AssetNames,
		ChunkNames:        config.ChunkNames,
		EntryNames:        config.EntryNames,
		Bundle:            true,
		Format:            api.FormatESModule,
		Splitting:         config.Splitting,
		TreeShaking:       api.TreeShakingDefault, // default shakes if bundle true, or format iife
		LegalComments:     config.LegalComments,
		Metafile:          config.Metafile,
		MinifyIdentifiers: isBuildMode,
		MinifySyntax:      isBuildMode,
		MinifyWhitespace:  isBuildMode,
		Write:             true,
		Alias:             config.AliasPackages,

		Define:    definedReplacements,
		Inject:    config.Injects,
		Loader:    config.Loaders,
		Sourcemap: config.SourceMap,

		Tsconfig: filepath.Join(baseDir, config.TSConfigPath),

		Plugins: []api.Plugin{
			plugins.AliasPlugin(config.ResolveModules),
			plugins.InlinePlugin(config.InlineSize, config.InlineExtensions),
		},

		// react stuff
		JSX:             config.JSX,
		JSXDev:          !isBuildMode,
		JSXFactory:      config.JSXFactory,
		JSXFragment:     config.JSXFragment,
		JSXImportSource: config.JSXImportSource,
		JSXSideEffects:  config.JSXSideEffects,
	}
}
