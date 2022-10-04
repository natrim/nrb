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
	"runtime"
	"strconv"
)

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

func init() {
	flag.BoolVar(&isWatch, "w", isWatch, "watch mode")
	flag.BoolVar(&isBuild, "b", isBuild, "build mode")
	flag.BoolVar(&isServe, "s", isServe, "serve mode")
	flag.BoolVar(&isTest, "t", isTest, "test mode")
	flag.BoolVar(&isMakeCert, "c", isMakeCert, "make cert")
	flag.BoolVar(&isHelp, "h", isVersion, "help")
	flag.BoolVar(&isVersion, "v", isHelp, "version")
	flag.BoolVar(&isVersionUpdate, "u", isVersionUpdate, "version update")
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
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	isHelp = isHelp || (!isBuild && !isServe && !isTest && !isMakeCert && !isVersion && !isVersionUpdate && !isWatch && npmRun == "")

	if isHelp {
		fmt.Println("× No help defined")
		fmt.Println("> just kidding, run this script with -b to build the app, -w for watch mode, -s to serve build folder, -u to update build number, -t for test's, -v for current build version, -c to make https certificate for dev, -env to set custom .env files, -r to run npm scripts, and -h for this help")
		os.Exit(0)
	}

	if isTest {
		fmt.Println("× No test's defined")
		os.Exit(1)
	}

	if npmRun != "" {
		jsonFile, err := os.ReadFile(filepath.Join(baseDir, "package.json"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		var packageJson map[string]any
		err = json.Unmarshal(jsonFile, &packageJson)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if scripts, ok := packageJson["scripts"]; ok {
			if script, ok := scripts.(map[string]any)[npmRun]; ok {
				cmd := exec.Command("bash", "-c", script.(string))
				if runtime.GOOS == "windows" {
					cmd = exec.Command("bash.exe", "-c", script.(string))
				}
				cmd.Env = os.Environ()
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Println(err)
					os.Exit(1)
				} else {
					fmt.Printf("✓ Run \"%s\" done.\n", npmRun)
					os.Exit(0)
				}
			} else {
				fmt.Println("× No script found in package.json")
				os.Exit(1)
			}
		} else {
			fmt.Println("× No scripts found in package.json")
			os.Exit(1)
		}
	}

	if isMakeCert {
		os.RemoveAll(filepath.Join(baseDir, ".cert"))
		os.Mkdir(filepath.Join(baseDir, ".cert"), 0755)
		cmd := exec.Command("mkcert -key-file " + baseDir + "/.cert/key.pem -cert-file " + baseDir + "/.cert/cert.pem '" + host + "'")
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	jsonFile, err := os.ReadFile(filepath.Join(staticDir, "version.json"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = json.Unmarshal(jsonFile, &metaData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if isVersion {
		fmt.Println("✓ Current version number is:", metaData["version"])
		os.Exit(0)
	}

	if isVersionUpdate {
		fmt.Println("> Incrementing build number...")
		v, _ := strconv.Atoi(fmt.Sprintf("%v", metaData["version"]))
		metaData["version"] = v + 1

		j, err := json.Marshal(metaData)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = os.WriteFile(filepath.Join(staticDir, "version.json"), j, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("✓ App version has been updated")
		fmt.Println("✓ Current version number is:", metaData["version"])
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
		// hot reload watcher
		reload := "(()=>{if(window.esIn)return;window.esIn=true;var s=new EventSource(\"/esbuild\"),r=0;s.onerror=()=>{r++;if(r>30){s.close();console.error('hot reload failed to init')}};s.onmessage=()=>location.reload()})();"
		if buildOptions.Banner == nil {
			buildOptions.Banner = map[string]string{"js": reload}
		} else {
			if _, ok := buildOptions.Banner["js"]; !ok {
				buildOptions.Banner["js"] = reload
			} else {
				buildOptions.Banner["js"] = reload + buildOptions.Banner["js"]
			}
		}

		watch()
		os.Exit(0)
	}

	if isBuild {
		build()
		os.Exit(0)
	}
}
