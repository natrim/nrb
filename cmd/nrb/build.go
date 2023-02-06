package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func build() {
	start := time.Now()

	// remove output directory
	err := os.RemoveAll(outputDir)
	if err != nil {
		fmt.Println(ERR, "Failed to clean build directory:", err)
		return
	}
	// dont remake as we make copy of static dir:
	// os.MkdirAll(outputDir, 0755)

	fmt.Println(OK, "Cleaned output dir")
	fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())

	// copy static directory to build directory
	err = lib.CopyDir(outputDir, staticDir)
	if err != nil {
		fmt.Println(ERR, "Failed to copy static directory:", err)
		return
	}

	fmt.Println(OK, "Copied static files to output dir")
	fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())

	fmt.Println(ITEM, "Building..")
	fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())

	// use metafile
	buildOptions.Metafile = true

	// make sure to write files on build
	buildOptions.Write = true

	// esbuild app
	result := api.Build(buildOptions)

	if len(result.Errors) > 0 {
		fmt.Println(ERR, "Failed to build")
		fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())
		for _, err := range result.Errors {
			fmt.Println("-*-", err.Text)
		}
		os.Exit(1)
	}

	fmt.Println(OK, "Esbuild done")
	fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())

	if generateMetafile {
		if err = os.WriteFile(filepath.Join(outputDir, "build-meta.json"), []byte(result.Metafile), 0644); err != nil {
			fmt.Println(ERR, "Failed to save metafile:", err)
		}
		fmt.Println(OK, "Metafile saved to 'build-meta.json'")
		fmt.Printf(INFO + " use e.g. https://esbuild.github.io/analyze/ to analyze the bundle\n")
		fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())
	}

	fmt.Println(ITEM, "Building index.html file...")
	makeIndex(&result)

	fmt.Println(OK, "Build done")
	fmt.Printf(INFO+" Time: %dms\n", time.Since(start).Milliseconds())

	fmt.Println(OK, " All work done 🎂")
}

func makeIndex(result *api.BuildResult) {
	var metafile Metadata
	err := json.Unmarshal([]byte(result.Metafile), &metafile)
	if err != nil {
		fmt.Println(ERR, "Failed to parse build metadata:", err)
		os.Exit(1)
	}

	indexFile, err := os.ReadFile(filepath.Join(outputDir, "index.html"))
	if err != nil {
		fmt.Println(ERR, "Failed to read build index.html:", err)
		os.Exit(1)
	}

	//inject main js/css if not already in index.html
	indexFile, saveIndexFile := lib.InjectVarsIntoIndex(indexFile, entryFileName, assetsDir, publicUrl)

	// find chunks to preload
	if len(preloadPathsStartingWith) > 0 {
		var chunksToPreload []string
		for chunk, m := range metafile.Outputs {
			for i := range m.Inputs {
				for _, p := range preloadPathsStartingWith {
					if p != "" && strings.HasPrefix(i, p) {
						chunksToPreload = append(chunksToPreload, chunk)
					}
				}
			}
		}

		if len(chunksToPreload) > 0 {
			publicUrl := strings.TrimSuffix(publicUrl, "/")
			indexFileName := strings.TrimSuffix(filepath.Base(entryFileName), filepath.Ext(entryFileName))
			findP := regexp.MustCompile(fmt.Sprintf(`<link rel=(["']?)modulepreload(["']?) href=(["']?)%s/%s/%s\.js(["']?)( ?/?)>`, publicUrl, assetsDir, indexFileName))
			saveIndexFile = true
			var replace [][]byte
			for _, chunk := range chunksToPreload {
				replace = append(replace, []byte(fmt.Sprintf(`<link rel=${1}modulepreload${2} href=${3}%s/%s${4}${5}>`, publicUrl, strings.ReplaceAll(chunk, filepath.Join(outputDir, assetsDir), assetsDir))))
			}
			indexFile = findP.ReplaceAll(indexFile, bytes.Join(replace, []byte("\n")))
		}
	}

	if saveIndexFile {
		err = os.WriteFile(filepath.Join(outputDir, "index.html"), indexFile, 0644)
		if err != nil {
			fmt.Println(ERR, "Failed to write build index.html:", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(ITEM, "No changes to index.html")
	}
}

// Metadata is json equivalent of this esbuild metadata interface
//
//		interface Metadata {
//		 inputs: {
//		   [path: string]: {
//		     bytes: number
//		     imports: {
//		       path: string
//		       kind: string
//	        external?: boolean
//	        original?: string
//		     }[]
//	        format?: 'cjs' | 'esm'
//		   }
//		 }
//		 outputs: {
//		   [path: string]: {
//		     bytes: number
//		     inputs: {
//		       [path: string]: {
//		         bytesInOutput: number
//		       }
//		     }
//		     imports: {
//		       path: string
//		       kind: string
//		     }[]
//		     exports: string[]
//		     entryPoint?: string
//		     cssBundle?: string
//		   }
//		 }
//		}
type Metadata struct {
	Inputs map[string]struct {
		Bytes   float64 `json:"bytes"`
		Imports []struct {
			Path     string `json:"path"`
			Kind     string `json:"kind"`
			External bool   `json:"external"`
			Original string `json:"original"`
		} `json:"imports"`
		Format string `json:"format"`
	} `json:"inputs"`
	Outputs map[string]struct {
		Bytes  float64 `json:"bytes"`
		Inputs map[string]struct {
			BytesInOutput float64 `json:"bytesInOutput"`
		} `json:"inputs"`
		Imports []struct {
			Path string `json:"path"`
			Kind string `json:"kind"`
		} `json:"imports"`
		Exports    []string `json:"exports"`
		EntryPoint string   `json:"entryPoint"`
		CssBundle  string   `json:"cssBundle"`
	} `json:"outputs"`
}
