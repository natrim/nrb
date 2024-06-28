## NRB

- just simple builder for react app's
- it is used mostly for my projects and work stuff
- but maybe it will be usefull for someone else

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

### Usage

```
> Usage: nrb [flags] command
> use command with 'build' to build the app, 'watch' for watch mode, 'serve' to serve build folder and 'help' to show this help
Flags:
  -alias value
    	alias package with another 'package:aliasedpackage', overrides values from package.json, can have multiple flags, ie. --alias=react:preact-compat,react-dom:preact-compat
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
  -h	alias of -help
  -help
    	this help
  -host string
    	host (default "localhost")
  -inject value
    	allows you to automatically replace a global variable with an import from another file, overrides values from package.json, can have multiple flags, ie. --inject=./process-shim.js,./react-shim.js
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
  -loaders value
    	esbuild file loaders, overrides values from package.json, ie. --loaders=png:dataurl,.txt:copy,data:json
  -metafile
    	generate metafile for bundle analysis, ie. on https://esbuild.github.io/analyze/
  -outputDir string
    	output dir name (default "build")
  -port int
    	port (default 3000)
  -preload value
    	paths to module=preload on build, overrides values from package.json, can have multiple flags, ie. --preload=src/index,node_modules/react
  -publicUrl string
    	public url (default "/")
  -resolve value
    	resolve package import with 'package:path', overrides values from package.json, can have multiple flags, ie. --resolve=react:packages/super-react/index.js,redux:node_modules/redax/lib/index.js
  -sourceDir string
    	source directory name (default "src")
  -sourceMap string
    	what sourcemap to use, available options: none|inline|linked|external|both (default "linked")
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
    	nrb version numberg
```

#### HTTPS for watch/serve

for example use ["mkcert"](https://mkcert.dev)
( ie. `rm -rf ./.cert && mkdir ./.cert && mkcert -key-file ./.cert/key.pem -cert-file ./.cert/cert.pem 'localhost'` )

`nrb` searches `.cert/cert.pem` and `.cert/key.pem` for certificate by default

or set `ENV` variables `DEV_SERVER_CERT` and `DEV_SERVER_KEY` with paths to cert files

#### Package.json nrb config example

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
        "loaders": {
            ".data": "json",
            "txt": "copy"
        },
        "inline": {
            "size": 10000,
            "extensions": [
                "svg",
                "png",
                "jpg"
            ]
        },
        "splitting": true
    }
}
```

### TODO

- better config
- documentation
- more thing to read from config (custom config file instead of just reading package.json?)
- logging
- prebuilt binaries (npm?)
- custom plugins?
