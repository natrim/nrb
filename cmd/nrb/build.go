package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
)

func build() error {
	start := time.Now()

	// prepare esbuild build options
	buildEsbuildConfig(true)

	lib.PrintOk("Init done")
	lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

	// remove output directory
	err := os.RemoveAll(config.OutputDir)
	if err != nil {
		return errors.Join(errors.New("failed to clean build directory"), err)
	}

	lib.PrintOk("Cleaned output dir")
	lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

	// copy static directory to build directory
	if config.StaticDir != "" {
		err = lib.CopyDir(config.OutputDir, config.StaticDir)
		if err != nil {
			return errors.Join(errors.New("failed to copy static directory"), err)
		}

		lib.PrintOk("Copied static files to output dir")
		lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())
	} else {
		os.MkdirAll(config.OutputDir, 0755)
	}

	lib.PrintItem("Building..")
	lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

	// use metafile
	buildOptions.Metafile = true

	// make sure to write files on build
	buildOptions.Write = true

	// esbuild app
	result := api.Build(buildOptions)

	if len(result.Errors) > 0 {
		lib.PrintError("failed to build")
		lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

		errs := make([]error, len(result.Errors))
		for i, err := range result.Errors {
			errs[i] = errors.New("-*- " + err.Text)
		}
		return errors.Join(errs...)
	}

	lib.PrintOk("Esbuild done")
	lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

	if config.Metafile {
		if err = os.WriteFile(filepath.Join(config.OutputDir, "build-meta.json"), []byte(result.Metafile), 0644); err != nil {
			lib.PrintError("failed to save metafile", err)
		} else {
			lib.PrintOk("Metafile saved to 'build-meta.json'")
			lib.PrintInfof("use e.g. https://esbuild.github.io/analyze/ to analyze the bundle\n")
			lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())
		}
	}

	err = os.WriteFile(filepath.Join(config.OutputDir, "version.json"), fmt.Appendf(nil, "{\"hash\":\"%s\",\"time\":%d}", versionData, start.Unix()), 0644)
	if err != nil {
		lib.PrintError("failed to save version.json", err)
	}

	lib.PrintItem("Building index.html file...")
	err = makeIndex(config.PreloadPathsStartingWith, &result)
	if err != nil {
		return err
	}
	lib.PrintOk("Build done")
	lib.PrintInfof("Time: %dms\n", time.Since(start).Milliseconds())

	lib.PrintOk("All work done 🎂")

	return nil
}

func makeIndex(preloadPathsStartingWith lib.ArrayFlags, result *api.BuildResult) error {
	var metafile Metadata
	err := json.Unmarshal([]byte(result.Metafile), &metafile)
	if err != nil {
		return errors.Join(errors.New("failed to parse build metadata"), err)
	}

	indexFile, err := os.ReadFile(filepath.Join(config.OutputDir, "index.html"))
	if err != nil {
		indexFile, err = os.ReadFile(filepath.Join(baseDir, "index.html"))

		if err != nil {
			return errors.Join(errors.New("failed to read build index.html"), err)
		}
	}

	//inject main js/css if not already in index.html
	indexFile, saveIndexFile := lib.InjectVarsIntoIndex(indexFile, config.EntryFileName, config.AssetsDir, config.PublicURL)

	// find chunks to preload
	if len(preloadPathsStartingWith) > 0 {
		var chunksToPreload = make(map[string]bool)
		for chunk, m := range metafile.Outputs {
			for i := range m.Inputs {
				if exists := chunksToPreload[chunk]; exists {
					continue
				}
				for _, p := range preloadPathsStartingWith {
					if p != "" && strings.HasPrefix(i, p) {
						chunksToPreload[chunk] = true
					}
				}
			}
		}

		if len(chunksToPreload) > 0 {
			publicURL := strings.TrimSuffix(config.PublicURL, "/")
			indexFileName := strings.TrimSuffix(filepath.Base(config.EntryFileName), filepath.Ext(config.EntryFileName))
			findP := regexp.MustCompile(fmt.Sprintf("<link rel=([\"']?)modulepreload([\"']?) href=([\"']?)%s/%s/%s\\.js([\"']?)( ?/?)>(\n?)", publicURL, config.AssetsDir, indexFileName))
			saveIndexFile = true
			replace := strings.Builder{}
			for chunk := range chunksToPreload {
				fmt.Fprintf(&replace, "<link rel=${1}modulepreload${2} href=${3}%s/%s${4}${5}>${6}", publicURL, strings.ReplaceAll(chunk, filepath.Join(config.OutputDir, config.AssetsDir), config.AssetsDir))
			}
			// replace modulepreload index.js with modulepreload index.js and others
			indexFile = findP.ReplaceAll(indexFile, []byte(replace.String()))
		}
	}

	if saveIndexFile {
		err = os.WriteFile(filepath.Join(config.OutputDir, "index.html"), indexFile, 0644)
		if err != nil {
			return errors.Join(errors.New("failed to write built index.html"), err)
		}
	} else {
		lib.PrintItem("No changes to index.html")
	}

	return nil
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
