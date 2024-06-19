## NRB

- just simple builder for react app's

### Installation

```shell
go install github.com/natrim/nrb/cmd/nrb@latest
```

### Project structure (by default)

- root/
    - package.json (config and deps)
    - tsconfig.json (do you really write without TS?)
    - src/
        - index.tsx (app entry point)
    - public
        - index.html (static page to show before js kicks in, js/css gets injected to head)
        - version.json (optional, for "version-update" command)

#### Package.json config example

```json
{
    "nrb": {
        "alias": {
            "react": "preact"
        },
        "resolve": {
            "@material-ui/pickers": "node_modules/@material-ui/pickers/dist/material-ui-pickers.js",
            "@material-ui/core": "node_modules/@material-ui/core/index.js"
        },
        "preload": [
            "node_modules/preact@",
            "node_modules/preact/",
            "src/index"
        ],
        "inject": [
            "src/inject.js"
        ],
        "inline": {
            "size": 10000,
            "extensions": [
                "svg",
                "png",
                "jpg"
            ]
        }
    }
}
```

### Usage

```
> Usage: nrb [flags] command
> use command with 'build' to build the app, 'watch' for watch mode, 'serve' to serve build folder, 'version-update' to update build number, 'version' for current build version, 'cert' to make https certificate for watch/serve, 'run' to run npm scripts and 'help' to show this help
Flags:
  -assetNames string
    	asset names schema for esbuild (default "media/[name]-[hash]")
  -assetsDir string
    	assets dir name in output (default "assets")
  -chunkNames string
    	chunk names schema for esbuild (default "chunks/[name]-[hash]")
  -color
    	colorize output (default true)
  -entryFileName string
    	entry file name in 'sourceDir' (default "index.tsx")
  -entryNames string
    	entry names schema for esbuild (default "[name]")
  -env string
    	env files to load from (always loads .env first)
  -envPrefix string
    	env variables prefix (default "REACT_APP_")
  -gzip
    	gzip responses (default true)
  -h	alias of -help
  -help
    	this help
  -host string
    	host (default "localhost")
  -jsx string
    	tells esbuild what to do about JSX syntax, available options: automatic|transform|preserve (default "automatic")
  -jsxFactory string
    	What to use for JSX instead of "React.createElement"
  -jsxFragment string
    	What to use for JSX instead of "React.Fragment"
  -jsxImportSource string
    	Override the package name for the automatic runtime (default "react")
  -jsxSideEffects
    	Do not remove unused JSX expressions
  -legalComments string
    	what to do with legal comments, available options: none|inline|eof|linked|external (default "eof")
  -metafile
    	generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/
  -outputDir string
    	output dir name (default "build")
  -port int
    	port (default 3000)
  -publicUrl string
    	public url (default "/")
  -sourceDir string
    	source directory name (default "src")
  -split
    	alias of -splitting
  -splitting
    	enable code splitting
  -staticDir string
    	static dir name (default "public")
  -target string
    	custom browser target, defaults to tsconfig target if possible, else esnext
  -tsconfig string
    	path to tsconfig json, relative to current work directory (default "tsconfig.json")
  -v	alias of -version
  -version
    	nrb version number
  -versionfile string
    	path to version.json, relative to current work directory (default "public/version.json")
```

### TODO

- better config
- documentation
- more thing to read from config (custom config file instead of just reading package.json?)
- logging
- prebuilt binaries (npm?)
- custom plugins?
