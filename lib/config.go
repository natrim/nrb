package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type Config struct {
	EnvPrefix                string
	SourceDir                string
	EntryFileName            string
	OutputDir                string
	StaticDir                string
	AssetsDir                string
	Port                     int
	Host                     string
	PublicURL                string
	Target                   string
	AssetNames               string
	ChunkNames               string
	EntryNames               string
	JSXFactory               string
	JSXFragment              string
	JSXImportSource          string
	JSXSideEffects           bool
	JSX                      api.JSX
	LegalComments            api.LegalComments
	SourceMap                api.SourceMap
	Metafile                 bool
	TSConfigPath             string
	AliasPackages            MapFlags
	ResolveModules           MapFlags
	PreloadPathsStartingWith ArrayFlags
	Injects                  ArrayFlags
	InlineSize               int64
	InlineExtensions         []string
	Loaders                  LoaderFlags
	Splitting                bool
}

type OptionalBool struct {
	Value bool
	Set   bool
}

type OptionalInt struct {
	Value int
	Set   bool
}

type OptionalInt64 struct {
	Value int64
	Set   bool
}

type OptionalString struct {
	Value string
	Set   bool
}

type OptionalEnum[T any] struct {
	Value T
	Set   bool
}

type ConfigPatch struct {
	EnvPrefix       OptionalString
	SourceDir       OptionalString
	EntryFileName   OptionalString
	OutputDir       OptionalString
	StaticDir       OptionalString
	AssetsDir       OptionalString
	Port            OptionalInt
	Host            OptionalString
	PublicURL       OptionalString
	Target          OptionalString
	AssetNames      OptionalString
	ChunkNames      OptionalString
	EntryNames      OptionalString
	JSXFactory      OptionalString
	JSXFragment     OptionalString
	JSXImportSource OptionalString
	JSXSideEffects  OptionalBool
	JSX             OptionalEnum[api.JSX]
	LegalComments   OptionalEnum[api.LegalComments]
	SourceMap       OptionalEnum[api.SourceMap]
	Metafile        OptionalBool
	TSConfigPath    OptionalString
	Splitting       OptionalBool

	AliasPackages            MapFlags
	ResolveModules           MapFlags
	PreloadPathsStartingWith ArrayFlags
	Injects                  ArrayFlags
	InlineExtensions         ArrayFlags
	InlineSize               OptionalInt64
	Loaders                  LoaderFlags
}

type ConfigOverrides struct {
	EnvPrefix       OptionalString
	SourceDir       OptionalString
	EntryFileName   OptionalString
	OutputDir       OptionalString
	StaticDir       OptionalString
	AssetsDir       OptionalString
	Port            OptionalInt
	Host            OptionalString
	PublicURL       OptionalString
	Target          OptionalString
	AssetNames      OptionalString
	ChunkNames      OptionalString
	EntryNames      OptionalString
	JSXFactory      OptionalString
	JSXFragment     OptionalString
	JSXImportSource OptionalString
	JSXSideEffects  OptionalBool
	JSX             OptionalEnum[api.JSX]
	LegalComments   OptionalEnum[api.LegalComments]
	SourceMap       OptionalEnum[api.SourceMap]
	Metafile        OptionalBool
	TSConfigPath    OptionalString
	Splitting       OptionalBool

	AliasPackages            MapFlags
	ResolveModules           MapFlags
	PreloadPathsStartingWith ArrayFlags
	Injects                  ArrayFlags
	InlineExtensions         ArrayFlags
	InlineSize               OptionalInt64
	Loaders                  LoaderFlags
}

type PackageJson map[string]any

func DefaultConfig() Config {
	return Config{
		EnvPrefix:     "REACT_APP_",
		SourceDir:     "src",
		EntryFileName: "index.tsx",
		OutputDir:     "build",
		StaticDir:     "public",
		AssetsDir:     "assets",
		Port:          3000,
		Host:          "localhost",
		PublicURL:     "/",
		AssetNames:    "media/[name]-[hash]",
		ChunkNames:    "chunks/[name]-[hash]",
		EntryNames:    "[name]",
		LegalComments: api.LegalCommentsEndOfFile,
		JSX:           api.JSXAutomatic,
		SourceMap:     api.SourceMapLinked,
		TSConfigPath:  "tsconfig.json",
	}
}

func ParsePackageJson(packagePath string) (PackageJson, error) {
	if !FileExists(packagePath) {
		return nil, errors.New("no " + packagePath + " found")
	}

	jsonFile, err := os.ReadFile(packagePath)
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

func MergeConfig(base Config, overlay ConfigPatch) Config {
	mergeOptionalString(&base.EnvPrefix, overlay.EnvPrefix)
	mergeOptionalString(&base.SourceDir, overlay.SourceDir)
	mergeOptionalString(&base.EntryFileName, overlay.EntryFileName)
	mergeOptionalString(&base.OutputDir, overlay.OutputDir)
	mergeOptionalString(&base.StaticDir, overlay.StaticDir)
	mergeOptionalString(&base.AssetsDir, overlay.AssetsDir)
	mergeOptionalInt(&base.Port, overlay.Port)
	mergeOptionalString(&base.Host, overlay.Host)
	mergeOptionalString(&base.PublicURL, overlay.PublicURL)
	mergeOptionalString(&base.Target, overlay.Target)
	mergeOptionalString(&base.AssetNames, overlay.AssetNames)
	mergeOptionalString(&base.ChunkNames, overlay.ChunkNames)
	mergeOptionalString(&base.EntryNames, overlay.EntryNames)
	mergeOptionalString(&base.JSXFactory, overlay.JSXFactory)
	mergeOptionalString(&base.JSXFragment, overlay.JSXFragment)
	mergeOptionalString(&base.JSXImportSource, overlay.JSXImportSource)
	mergeOptionalBool(&base.JSXSideEffects, overlay.JSXSideEffects)
	mergeOptionalEnum(&base.JSX, overlay.JSX)
	mergeOptionalEnum(&base.LegalComments, overlay.LegalComments)
	mergeOptionalEnum(&base.SourceMap, overlay.SourceMap)
	mergeOptionalBool(&base.Metafile, overlay.Metafile)
	mergeOptionalString(&base.TSConfigPath, overlay.TSConfigPath)
	mergeOptionalBool(&base.Splitting, overlay.Splitting)

	if overlay.AliasPackages != nil {
		base.AliasPackages = overlay.AliasPackages
	}
	if overlay.ResolveModules != nil {
		base.ResolveModules = overlay.ResolveModules
	}
	if overlay.PreloadPathsStartingWith != nil {
		base.PreloadPathsStartingWith = overlay.PreloadPathsStartingWith
	}
	if overlay.Injects != nil {
		base.Injects = overlay.Injects
	}
	if overlay.InlineExtensions != nil {
		base.InlineExtensions = overlay.InlineExtensions
	}
	mergeOptionalInt64(&base.InlineSize, overlay.InlineSize)
	if overlay.Loaders != nil {
		base.Loaders = overlay.Loaders
	}

	return base
}

func ApplyOverrides(cfg *Config, overrides ConfigOverrides) {
	mergeOptionalString(&cfg.EnvPrefix, overrides.EnvPrefix)
	mergeOptionalString(&cfg.SourceDir, overrides.SourceDir)
	mergeOptionalString(&cfg.EntryFileName, overrides.EntryFileName)
	mergeOptionalString(&cfg.OutputDir, overrides.OutputDir)
	mergeOptionalString(&cfg.StaticDir, overrides.StaticDir)
	mergeOptionalString(&cfg.AssetsDir, overrides.AssetsDir)
	mergeOptionalInt(&cfg.Port, overrides.Port)
	mergeOptionalString(&cfg.Host, overrides.Host)
	mergeOptionalString(&cfg.PublicURL, overrides.PublicURL)
	mergeOptionalString(&cfg.Target, overrides.Target)
	mergeOptionalString(&cfg.AssetNames, overrides.AssetNames)
	mergeOptionalString(&cfg.ChunkNames, overrides.ChunkNames)
	mergeOptionalString(&cfg.EntryNames, overrides.EntryNames)
	mergeOptionalString(&cfg.JSXFactory, overrides.JSXFactory)
	mergeOptionalString(&cfg.JSXFragment, overrides.JSXFragment)
	mergeOptionalString(&cfg.JSXImportSource, overrides.JSXImportSource)
	mergeOptionalBool(&cfg.JSXSideEffects, overrides.JSXSideEffects)
	mergeOptionalEnum(&cfg.JSX, overrides.JSX)
	mergeOptionalEnum(&cfg.LegalComments, overrides.LegalComments)
	mergeOptionalEnum(&cfg.SourceMap, overrides.SourceMap)
	mergeOptionalBool(&cfg.Metafile, overrides.Metafile)
	mergeOptionalString(&cfg.TSConfigPath, overrides.TSConfigPath)
	mergeOptionalBool(&cfg.Splitting, overrides.Splitting)

	if overrides.AliasPackages != nil {
		cfg.AliasPackages = overrides.AliasPackages
	}
	if overrides.ResolveModules != nil {
		cfg.ResolveModules = overrides.ResolveModules
	}
	if overrides.PreloadPathsStartingWith != nil {
		cfg.PreloadPathsStartingWith = overrides.PreloadPathsStartingWith
	}
	if overrides.Injects != nil {
		cfg.Injects = overrides.Injects
	}
	if overrides.InlineExtensions != nil {
		cfg.InlineExtensions = overrides.InlineExtensions
	}
	mergeOptionalInt64(&cfg.InlineSize, overrides.InlineSize)
	if overrides.Loaders != nil {
		cfg.Loaders = overrides.Loaders
	}
}

func mergeOptionalString(dst *string, value OptionalString) {
	if value.Set {
		*dst = value.Value
	}
}

func mergeOptionalInt(dst *int, value OptionalInt) {
	if value.Set {
		*dst = value.Value
	}
}

func mergeOptionalInt64(dst *int64, value OptionalInt64) {
	if value.Set {
		*dst = value.Value
	}
}

func mergeOptionalEnum[T any](dst *T, value OptionalEnum[T]) {
	if value.Set {
		*dst = value.Value
	}
}

func mergeOptionalBool(dst *bool, value OptionalBool) {
	if value.Set {
		*dst = value.Value
	}
}

func ParseJsonConfig(packageJson PackageJson) (ConfigPatch, error) {
	config := ConfigPatch{}

	raw, ok := packageJson["nrb"]
	if !ok || raw == nil {
		return config, nil
	}

	options, ok := raw.(map[string]any)
	if !ok {
		return config, errors.New("wrong 'nrb' key in 'package.json', use object")
	}

	if err := parseOptionalString(options, "envPrefix", &config.EnvPrefix); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "sourceDir", &config.SourceDir); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "entryFileName", &config.EntryFileName); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "outputDir", &config.OutputDir); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "staticDir", &config.StaticDir); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "assetsDir", &config.AssetsDir); err != nil {
		return config, err
	}
	if err := parseOptionalInt(options, "port", &config.Port); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "host", &config.Host); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "publicUrl", &config.PublicURL); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "target", &config.Target); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "assetNames", &config.AssetNames); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "chunkNames", &config.ChunkNames); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "entryNames", &config.EntryNames); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "jsxFactory", &config.JSXFactory); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "jsxFragment", &config.JSXFragment); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "jsxImportSource", &config.JSXImportSource); err != nil {
		return config, err
	}
	if err := parseOptionalBool(options, "jsxSideEffects", &config.JSXSideEffects); err != nil {
		return config, err
	}
	if err := parseOptionalJSX(options, "jsx", &config.JSX); err != nil {
		return config, err
	}
	if err := parseOptionalLegalComments(options, "legalComments", &config.LegalComments); err != nil {
		return config, err
	}
	if err := parseOptionalSourceMap(options, "sourceMap", &config.SourceMap); err != nil {
		return config, err
	}
	if err := parseOptionalBool(options, "metafile", &config.Metafile); err != nil {
		return config, err
	}
	if err := parseOptionalString(options, "tsconfig", &config.TSConfigPath); err != nil {
		return config, err
	}
	if err := parseOptionalBool(options, "splitting", &config.Splitting); err != nil {
		return config, err
	}

	if err := parseStringMap(options, "alias", &config.AliasPackages); err != nil {
		return config, err
	}
	if err := parseStringMap(options, "resolve", &config.ResolveModules); err != nil {
		return config, err
	}
	if err := parseStringSlice(options, "preload", &config.PreloadPathsStartingWith); err != nil {
		return config, err
	}
	if err := parseStringSlice(options, "inject", &config.Injects); err != nil {
		return config, err
	}
	if err := parseLoaderMap(options, "loaders", &config.Loaders); err != nil {
		return config, err
	}
	if err := parseInline(options, &config); err != nil {
		return config, err
	}

	return config, nil
}

func ParseJSX(value string) (api.JSX, error) {
	switch value {
	case "transform":
		return api.JSXTransform, nil
	case "preserve":
		return api.JSXPreserve, nil
	case "automatic":
		return api.JSXAutomatic, nil
	default:
		return 0, fmt.Errorf("wrong 'jsx' value in 'package.json', use automatic|transform|preserve")
	}
}

func JSXString(value api.JSX) string {
	switch value {
	case api.JSXTransform:
		return "transform"
	case api.JSXPreserve:
		return "preserve"
	case api.JSXAutomatic:
		return "automatic"
	default:
		return "unknown"
	}
}

func ParseLegalComments(value string) (api.LegalComments, error) {
	switch value {
	case "default":
		return api.LegalCommentsDefault, nil
	case "none":
		return api.LegalCommentsNone, nil
	case "inline":
		return api.LegalCommentsInline, nil
	case "eof":
		return api.LegalCommentsEndOfFile, nil
	case "linked":
		return api.LegalCommentsLinked, nil
	case "external":
		return api.LegalCommentsExternal, nil
	default:
		return 0, fmt.Errorf("wrong 'legalComments' value in 'package.json', use none|inline|eof|linked|external")
	}
}

func LegalCommentsString(value api.LegalComments) string {
	switch value {
	case api.LegalCommentsDefault:
		return "default"
	case api.LegalCommentsNone:
		return "none"
	case api.LegalCommentsInline:
		return "inline"
	case api.LegalCommentsEndOfFile:
		return "eof"
	case api.LegalCommentsLinked:
		return "linked"
	case api.LegalCommentsExternal:
		return "external"
	default:
		return "unknown"
	}
}

func ParseSourceMap(value string) (api.SourceMap, error) {
	switch value {
	case "none":
		return api.SourceMapNone, nil
	case "inline":
		return api.SourceMapInline, nil
	case "linked":
		return api.SourceMapLinked, nil
	case "external":
		return api.SourceMapExternal, nil
	case "both":
		return api.SourceMapInlineAndExternal, nil
	default:
		return 0, fmt.Errorf("wrong 'sourceMap' value in 'package.json', use none|inline|linked|external|both")
	}
}

func SourceMapString(value api.SourceMap) string {
	switch value {
	case api.SourceMapNone:
		return "none"
	case api.SourceMapInline:
		return "inline"
	case api.SourceMapLinked:
		return "linked"
	case api.SourceMapExternal:
		return "external"
	case api.SourceMapInlineAndExternal:
		return "both"
	default:
		return "unknown"
	}
}

func parseOptionalString(options map[string]any, key string, target *OptionalString) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use string", key)
	}

	*target = OptionalString{Value: s, Set: true}
	return nil
}

func parseOptionalJSX(options map[string]any, key string, target *OptionalEnum[api.JSX]) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use string", key)
	}

	parsed, err := ParseJSX(s)
	if err != nil {
		return err
	}

	*target = OptionalEnum[api.JSX]{Value: parsed, Set: true}
	return nil
}

func parseOptionalLegalComments(options map[string]any, key string, target *OptionalEnum[api.LegalComments]) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use string", key)
	}

	parsed, err := ParseLegalComments(s)
	if err != nil {
		return err
	}

	*target = OptionalEnum[api.LegalComments]{Value: parsed, Set: true}
	return nil
}

func parseOptionalSourceMap(options map[string]any, key string, target *OptionalEnum[api.SourceMap]) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use string", key)
	}

	parsed, err := ParseSourceMap(s)
	if err != nil {
		return err
	}

	*target = OptionalEnum[api.SourceMap]{Value: parsed, Set: true}
	return nil
}

func parseOptionalBool(options map[string]any, key string, target *OptionalBool) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	b, ok := value.(bool)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use boolean: true|false", key)
	}

	*target = OptionalBool{Value: b, Set: true}
	return nil
}

func parseOptionalInt(options map[string]any, key string, target *OptionalInt) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	i, err := parseNumberValue(value)
	if err != nil {
		return fmt.Errorf("wrong '%s' key in 'package.json', use number", key)
	}

	*target = OptionalInt{Value: int(i), Set: true}
	return nil
}

func parseOptionalInt64(options map[string]any, key string, target *OptionalInt64) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	i, err := parseNumberValue(value)
	if err != nil {
		return fmt.Errorf("wrong '%s' key in 'package.json', use number", key)
	}

	*target = OptionalInt64{Value: i, Set: true}
	return nil
}

func parseStringMap(options map[string]any, key string, target *MapFlags) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	rawMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use object", key)
	}

	result := make(MapFlags, len(rawMap))
	for name, mappedValue := range rawMap {
		result[name] = fmt.Sprintf("%v", mappedValue)
	}
	*target = result
	return nil
}

func parseStringSlice(options map[string]any, key string, target *ArrayFlags) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	rawSlice, ok := value.([]any)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use array", key)
	}

	result := make(ArrayFlags, len(rawSlice))
	for i, mappedValue := range rawSlice {
		result[i] = fmt.Sprintf("%v", mappedValue)
	}
	*target = result
	return nil
}

func parseLoaderMap(options map[string]any, key string, target *LoaderFlags) error {
	value, ok := options[key]
	if !ok {
		return nil
	}

	rawMap, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("wrong '%s' key in 'package.json', use object", key)
	}

	result := make(LoaderFlags, len(rawMap))
	for ext, loaderValue := range rawMap {
		loaderString, ok := loaderValue.(string)
		if !ok {
			return fmt.Errorf("wrong '%s' value in 'package.json': %q = %v", key, ext, loaderValue)
		}

		loader, err := ParseLoader(loaderString)
		if err != nil {
			return fmt.Errorf("wrong '%s' value in 'package.json': %q = %q", key, ext, loaderString)
		}

		result["."+strings.TrimPrefix(ext, ".")] = loader
	}

	*target = result
	return nil
}

func parseInline(options map[string]any, config *ConfigPatch) error {
	value, ok := options["inline"]
	if !ok {
		return nil
	}

	rawMap, ok := value.(map[string]any)
	if !ok {
		return errors.New("wrong 'inline' key in 'package.json', use object")
	}

	if _, ok := rawMap["size"]; ok {
		if err := parseOptionalInt64(rawMap, "size", &config.InlineSize); err != nil {
			return err
		}
	}
	if extensions, ok := rawMap["extensions"]; ok {
		rawSlice, ok := extensions.([]any)
		if !ok {
			return errors.New("wrong 'inline.extensions' key in 'package.json', use array")
		}
		result := make(ArrayFlags, len(rawSlice))
		for i, mappedValue := range rawSlice {
			result[i] = fmt.Sprintf("%v", mappedValue)
		}
		config.InlineExtensions = result
	}

	return nil
}

func parseNumberValue(value any) (int64, error) {
	switch n := value.(type) {
	case float64:
		if n != float64(int64(n)) {
			return 0, fmt.Errorf("not an integer: %v", n)
		}
		return int64(n), nil
	case float32:
		if n != float32(int64(n)) {
			return 0, fmt.Errorf("not an integer: %v", n)
		}
		return int64(n), nil
	case int:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		return int64(n), nil
	case uint8:
		return int64(n), nil
	case uint16:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		return int64(n), nil
	case json.Number:
		n64, err := n.Int64()
		if err != nil {
			return 0, err
		}
		return n64, nil
	default:
		return 0, fmt.Errorf("unsupported number type %T", value)
	}
}
