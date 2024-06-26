package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/natrim/nrb/lib"
)

func parsePackageJson() (PackageJson, error) {
	if !lib.FileExists(filepath.Join(baseDir, packagePath)) {
		return nil, errors.New("no " + filepath.Join(baseDir, packagePath) + " found")
	}

	jsonFile, err := os.ReadFile(filepath.Join(baseDir, packagePath))
	if err != nil {
		return nil, err
	}
	var packageJson PackageJson
	err = json.Unmarshal(jsonFile, &packageJson)
	if err != nil {
		return nil, err
	}

	return packageJson, nil
}

func parseJsonConfig(packageJson PackageJson) (*Config, error) {
	config := Config{}

	// check for alias/resolve/preload/inject options from package.json
	if packageJsonOptions, ok := packageJson["nrb"]; ok {
		options := packageJsonOptions.(map[string]any)

		if alias, ok := options["alias"]; ok {
			if _, ok = alias.(map[string]any); ok {
				config.AliasPackages = make(mapFlags)
				for name, aliasPath := range alias.(map[string]any) {
					config.AliasPackages[name] = fmt.Sprintf("%v", aliasPath)
				}
			} else {
				return &config, errors.New("wrong 'alias' key in 'package.json', use object: {package:alias,another:alias,...}")
			}
		}
		if loader, ok := options["loaders"]; ok {
			if _, ok = loader.(map[string]any); ok {
				config.Loaders = make(loaderFlags)
				for ext, loaderString := range loader.(map[string]any) {
					l, err := lib.ParseLoader(loaderString.(string))
					if err != nil {
						return &config, fmt.Errorf("wrong 'loaders' value in 'package.json': %q = %q", ext, loaderString)
					}
					config.Loaders["."+strings.TrimPrefix(ext, ".")] = l
				}
			} else {
				return &config, errors.New("wrong 'loaders' key in 'package.json', use object: {ext:loader,ext2:loadr,...}")
			}
		}
		if resolve, ok := options["resolve"]; ok {
			if _, ok = resolve.(map[string]any); ok {
				config.ResolveModules = make(mapFlags)
				for name, resolvePath := range resolve.(map[string]any) {
					config.ResolveModules[name] = fmt.Sprintf("%v", resolvePath)
				}
			} else {
				return &config, errors.New("wrong 'resolve' key in 'package.json', use object: {package:path,maybenaother:morepath,...}")
			}
		}
		if preload, ok := options["preload"]; ok {
			if _, ok = preload.([]any); ok {
				config.PreloadPathsStartingWith = make(arrayFlags, len(preload.([]any)))
				for i, pr := range preload.([]any) {
					config.PreloadPathsStartingWith[i] = fmt.Sprintf("%v", pr)
				}
			} else {
				return &config, errors.New("wrong 'preload' key in 'package.json', use array: [pathtopreload,maybeanotherpath,...]")
			}
		}
		if inject, ok := options["inject"]; ok {
			if _, ok = inject.([]any); ok {
				config.Injects = make(arrayFlags, len(inject.([]any)))
				for i, p := range inject.([]any) {
					config.Injects[i] = fmt.Sprintf("%v", p)
				}
			} else {
				return &config, errors.New("wrong 'inject' key in 'package.json', use array: [pathtoinject,maybeanotherpath,...]")
			}
		}
		if inline, ok := options["inline"]; ok {
			if inlineSz, ok := inline.(map[string]any)["size"]; ok {
				t, _ := strconv.Atoi(fmt.Sprintf("%v", inlineSz))
				config.InlineSize = int64(t)
			}
			if inlineExts, ok := inline.(map[string]any)["extensions"]; ok {
				if _, ok = inlineExts.([]any); ok {
					config.InlineExtensions = make(arrayFlags, len(inlineExts.([]any)))
					for i, pr := range inlineExts.([]any) {
						config.InlineExtensions[i] = fmt.Sprintf("%v", pr)
					}
				} else {
					return &config, errors.New("wrong 'inline.extensions' key in 'package.json', use array: [jpg,png,...]")
				}
			}
		}
		if splittingRaw, ok := options["splitting"]; ok {
			if splitting, ok := splittingRaw.(bool); ok {
				config.Splitting = splitting
			} else {
				return &config, errors.New("wrong 'splitting' key in 'package.json', use boolean: true|false")
			}
		}
	}

	return &config, nil
}
