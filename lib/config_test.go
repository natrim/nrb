package lib

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestDefaultConfigIncludesCurrentRuntimeDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.EnvPrefix != "REACT_APP_" {
		t.Fatalf("EnvPrefix = %q, want %q", cfg.EnvPrefix, "REACT_APP_")
	}
	if cfg.SourceDir != "src" {
		t.Fatalf("SourceDir = %q, want %q", cfg.SourceDir, "src")
	}
	if cfg.EntryFileName != "index.tsx" {
		t.Fatalf("EntryFileName = %q, want %q", cfg.EntryFileName, "index.tsx")
	}
	if cfg.OutputDir != "build" {
		t.Fatalf("OutputDir = %q, want %q", cfg.OutputDir, "build")
	}
	if cfg.StaticDir != "public" {
		t.Fatalf("StaticDir = %q, want %q", cfg.StaticDir, "public")
	}
	if cfg.AssetsDir != "assets" {
		t.Fatalf("AssetsDir = %q, want %q", cfg.AssetsDir, "assets")
	}
	if cfg.Port != 3000 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 3000)
	}
	if cfg.Host != "localhost" {
		t.Fatalf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.PublicURL != "/" {
		t.Fatalf("PublicURL = %q, want %q", cfg.PublicURL, "/")
	}
	if cfg.AssetNames != "media/[name]-[hash]" {
		t.Fatalf("AssetNames = %q, want %q", cfg.AssetNames, "media/[name]-[hash]")
	}
	if cfg.ChunkNames != "chunks/[name]-[hash]" {
		t.Fatalf("ChunkNames = %q, want %q", cfg.ChunkNames, "chunks/[name]-[hash]")
	}
	if cfg.EntryNames != "[name]" {
		t.Fatalf("EntryNames = %q, want %q", cfg.EntryNames, "[name]")
	}
	if cfg.LegalComments != api.LegalCommentsEndOfFile {
		t.Fatalf("LegalComments = %v, want %v", cfg.LegalComments, api.LegalCommentsEndOfFile)
	}
	if cfg.JSX != api.JSXAutomatic {
		t.Fatalf("JSX = %v, want %v", cfg.JSX, api.JSXAutomatic)
	}
	if cfg.SourceMap != api.SourceMapLinked {
		t.Fatalf("SourceMap = %v, want %v", cfg.SourceMap, api.SourceMapLinked)
	}
	if cfg.TSConfigPath != "tsconfig.json" {
		t.Fatalf("TSConfigPath = %q, want %q", cfg.TSConfigPath, "tsconfig.json")
	}
}

func TestParseJsonConfigReadsScalarAndBooleanKeys(t *testing.T) {
	pkg := PackageJson{
		"nrb": map[string]any{
			"envPrefix":       "APP_",
			"sourceDir":       "client",
			"entryFileName":   "main.tsx",
			"outputDir":       "dist",
			"staticDir":       "public-assets",
			"assetsDir":       "static",
			"port":            4173.0,
			"host":            "0.0.0.0",
			"publicUrl":       "/app",
			"target":          "es2022",
			"assetNames":      "assets/[name]",
			"chunkNames":      "chunks/[name]",
			"entryNames":      "entry/[name]",
			"jsxFactory":      "h",
			"jsxFragment":     "Fragment",
			"jsxImportSource": "preact",
			"jsxSideEffects":  true,
			"jsx":             "preserve",
			"legalComments":   "linked",
			"sourceMap":       "external",
			"metafile":        true,
			"tsconfig":        "tsconfig.app.json",
			"splitting":       true,
			"alias": map[string]any{
				"react": "preact",
			},
			"resolve": map[string]any{
				"@app": "src/app",
			},
			"preload": []any{"src/index"},
			"inject":  []any{"src/inject.js"},
			"loaders": map[string]any{
				".txt": "copy",
			},
			"inline": map[string]any{
				"size":       10000.0,
				"extensions": []any{"svg", "png"},
			},
		},
	}

	patch, err := ParseJsonConfig(pkg)
	if err != nil {
		t.Fatalf("ParseJsonConfig returned error: %v", err)
	}

	if !patch.EnvPrefix.Set || patch.EnvPrefix.Value != "APP_" {
		t.Fatalf("unexpected EnvPrefix patch: %#v", patch.EnvPrefix)
	}
	if !patch.SourceDir.Set || patch.SourceDir.Value != "client" {
		t.Fatalf("unexpected SourceDir patch: %#v", patch.SourceDir)
	}
	if !patch.EntryFileName.Set || patch.EntryFileName.Value != "main.tsx" {
		t.Fatalf("unexpected EntryFileName patch: %#v", patch.EntryFileName)
	}
	if !patch.OutputDir.Set || patch.OutputDir.Value != "dist" {
		t.Fatalf("unexpected OutputDir patch: %#v", patch.OutputDir)
	}
	if !patch.Port.Set || patch.Port.Value != 4173 {
		t.Fatalf("Port patch = %#v, want %d", patch.Port, 4173)
	}
	if !patch.JSXSideEffects.Set || !patch.JSXSideEffects.Value {
		t.Fatalf("expected JSXSideEffects package value to be preserved: %#v", patch.JSXSideEffects)
	}
	if !patch.JSX.Set || patch.JSX.Value != api.JSXPreserve {
		t.Fatalf("unexpected JSX patch: %#v", patch.JSX)
	}
	if !patch.LegalComments.Set || patch.LegalComments.Value != api.LegalCommentsLinked {
		t.Fatalf("unexpected LegalComments patch: %#v", patch.LegalComments)
	}
	if !patch.SourceMap.Set || patch.SourceMap.Value != api.SourceMapExternal {
		t.Fatalf("unexpected SourceMap patch: %#v", patch.SourceMap)
	}
	if !patch.Metafile.Set || !patch.Metafile.Value {
		t.Fatalf("expected Metafile package value to be preserved: %#v", patch.Metafile)
	}
	if !patch.Splitting.Set || !patch.Splitting.Value {
		t.Fatalf("expected Splitting package value to be preserved: %#v", patch.Splitting)
	}
	if !patch.InlineSize.Set || patch.InlineSize.Value != 10000 {
		t.Fatalf("unexpected InlineSize patch: %#v", patch.InlineSize)
	}
	if got := patch.AliasPackages["react"]; got != "preact" {
		t.Fatalf("AliasPackages[react] = %q, want %q", got, "preact")
	}
	if got, ok := patch.Loaders[".txt"]; !ok {
		t.Fatalf("expected loader override to be preserved: %#v", patch.Loaders)
	} else if gotString, err := StringifyLoader(got); err != nil || gotString != "copy" {
		t.Fatalf("Loaders[.txt] = %q, err=%v, want %q", gotString, err, "copy")
	}
	if len(patch.InlineExtensions) != 2 || patch.InlineExtensions[0] != "svg" || patch.InlineExtensions[1] != "png" {
		t.Fatalf("unexpected InlineExtensions patch: %#v", patch.InlineExtensions)
	}
}

func TestParseJsonConfigPreservesExplicitZeroFalseAndEmptyValues(t *testing.T) {
	pkg := PackageJson{
		"nrb": map[string]any{
			"port":      0.0,
			"metafile":  false,
			"splitting": false,
			"envPrefix": "",
			"sourceDir": "",
			"publicUrl": "",
			"inline": map[string]any{
				"size": 0.0,
			},
		},
	}

	patch, err := ParseJsonConfig(pkg)
	if err != nil {
		t.Fatalf("ParseJsonConfig returned error: %v", err)
	}

	if !patch.Port.Set || patch.Port.Value != 0 {
		t.Fatalf("Port patch = %#v, want explicit zero", patch.Port)
	}
	if !patch.Metafile.Set || patch.Metafile.Value {
		t.Fatalf("Metafile patch = %#v, want explicit false", patch.Metafile)
	}
	if !patch.Splitting.Set || patch.Splitting.Value {
		t.Fatalf("Splitting patch = %#v, want explicit false", patch.Splitting)
	}
	if !patch.EnvPrefix.Set || patch.EnvPrefix.Value != "" {
		t.Fatalf("EnvPrefix patch = %#v, want explicit empty string", patch.EnvPrefix)
	}
	if !patch.SourceDir.Set || patch.SourceDir.Value != "" {
		t.Fatalf("SourceDir patch = %#v, want explicit empty string", patch.SourceDir)
	}
	if !patch.PublicURL.Set || patch.PublicURL.Value != "" {
		t.Fatalf("PublicURL patch = %#v, want explicit empty string", patch.PublicURL)
	}
	if !patch.InlineSize.Set || patch.InlineSize.Value != 0 {
		t.Fatalf("InlineSize patch = %#v, want explicit zero", patch.InlineSize)
	}
}

func TestParseJsonConfigRejectsInvalidEnumValues(t *testing.T) {
	tests := []struct {
		name string
		pkg  PackageJson
	}{
		{
			name: "jsx must be valid option",
			pkg: PackageJson{
				"nrb": map[string]any{"jsx": "invalid"},
			},
		},
		{
			name: "legalComments must be valid option",
			pkg: PackageJson{
				"nrb": map[string]any{"legalComments": "bad"},
			},
		},
		{
			name: "sourceMap must be valid option",
			pkg: PackageJson{
				"nrb": map[string]any{"sourceMap": "bad"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseJsonConfig(tt.pkg); err == nil {
				t.Fatal("expected ParseJsonConfig to fail")
			}
		})
	}
}

func TestMergeConfigAppliesPackageThenCliOverrides(t *testing.T) {
	base := DefaultConfig()
	pkg := ConfigPatch{
		SourceDir: OptionalString{Value: "client", Set: true},
		Port:      OptionalInt{Value: 4173, Set: true},
		Splitting: OptionalBool{Value: true, Set: true},
	}
	overrides := ConfigOverrides{
		Port:      OptionalInt{Value: 8080, Set: true},
		Splitting: OptionalBool{Value: false, Set: true},
		SourceDir: OptionalString{Value: "web", Set: true},
	}

	cfg := MergeConfig(base, pkg)
	ApplyOverrides(&cfg, overrides)

	if cfg.SourceDir != "web" {
		t.Fatalf("SourceDir = %q, want %q", cfg.SourceDir, "web")
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Splitting {
		t.Fatalf("Splitting = true, want false after explicit CLI override")
	}
}

func TestMergeConfigAndApplyOverridesPreserveExplicitZeroFalseAndEmptyValues(t *testing.T) {
	base := DefaultConfig()
	base.EnvPrefix = "BASE_"
	base.SourceDir = "base"
	base.PublicURL = "/base"
	base.Port = 1234
	base.Metafile = true
	base.Splitting = true
	base.InlineSize = 99

	cfg := MergeConfig(base, ConfigPatch{
		EnvPrefix:  OptionalString{Value: "", Set: true},
		SourceDir:  OptionalString{Value: "", Set: true},
		PublicURL:  OptionalString{Value: "", Set: true},
		Port:       OptionalInt{Value: 0, Set: true},
		Metafile:   OptionalBool{Value: false, Set: true},
		Splitting:  OptionalBool{Value: false, Set: true},
		InlineSize: OptionalInt64{Value: 0, Set: true},
	})

	if cfg.EnvPrefix != "" {
		t.Fatalf("EnvPrefix = %q, want explicit empty string", cfg.EnvPrefix)
	}
	if cfg.SourceDir != "" {
		t.Fatalf("SourceDir = %q, want explicit empty string", cfg.SourceDir)
	}
	if cfg.PublicURL != "" {
		t.Fatalf("PublicURL = %q, want explicit empty string", cfg.PublicURL)
	}
	if cfg.Port != 0 {
		t.Fatalf("Port = %d, want explicit zero", cfg.Port)
	}
	if cfg.Metafile {
		t.Fatalf("Metafile = true, want explicit false")
	}
	if cfg.Splitting {
		t.Fatalf("Splitting = true, want explicit false")
	}
	if cfg.InlineSize != 0 {
		t.Fatalf("InlineSize = %d, want explicit zero", cfg.InlineSize)
	}

	ApplyOverrides(&cfg, ConfigOverrides{
		EnvPrefix:  OptionalString{Value: "", Set: true},
		SourceDir:  OptionalString{Value: "", Set: true},
		PublicURL:  OptionalString{Value: "", Set: true},
		Port:       OptionalInt{Value: 0, Set: true},
		Metafile:   OptionalBool{Value: false, Set: true},
		Splitting:  OptionalBool{Value: false, Set: true},
		InlineSize: OptionalInt64{Value: 0, Set: true},
	})

	if cfg.EnvPrefix != "" {
		t.Fatalf("EnvPrefix = %q, want explicit empty string after overrides", cfg.EnvPrefix)
	}
	if cfg.SourceDir != "" {
		t.Fatalf("SourceDir = %q, want explicit empty string after overrides", cfg.SourceDir)
	}
	if cfg.PublicURL != "" {
		t.Fatalf("PublicURL = %q, want explicit empty string after overrides", cfg.PublicURL)
	}
	if cfg.Port != 0 {
		t.Fatalf("Port = %d, want explicit zero after overrides", cfg.Port)
	}
	if cfg.Metafile {
		t.Fatalf("Metafile = true, want explicit false after overrides")
	}
	if cfg.Splitting {
		t.Fatalf("Splitting = true, want explicit false after overrides")
	}
	if cfg.InlineSize != 0 {
		t.Fatalf("InlineSize = %d, want explicit zero after overrides", cfg.InlineSize)
	}
}

func TestParseJsonConfigRejectsInvalidScalarTypes(t *testing.T) {
	tests := []struct {
		name string
		pkg  PackageJson
	}{
		{
			name: "port must be number",
			pkg: PackageJson{
				"nrb": map[string]any{"port": "3000"},
			},
		},
		{
			name: "jsxSideEffects must be boolean",
			pkg: PackageJson{
				"nrb": map[string]any{"jsxSideEffects": "true"},
			},
		},
		{
			name: "tsconfig must be string",
			pkg: PackageJson{
				"nrb": map[string]any{"tsconfig": 123.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseJsonConfig(tt.pkg); err == nil {
				t.Fatal("expected ParseJsonConfig to fail")
			}
		})
	}
}

func TestParseJsonConfigPreservesStructuredOptions(t *testing.T) {
	patch, err := ParseJsonConfig(PackageJson{
		"nrb": map[string]any{
			"alias":   map[string]any{"react": "preact"},
			"resolve": map[string]any{"foo": "./vendor/foo.js"},
			"preload": []any{"src/index"},
			"inject":  []any{"src/shim.js"},
			"loaders": map[string]any{".txt": "copy"},
			"inline": map[string]any{
				"size":       10000.0,
				"extensions": []any{"svg", "png"},
			},
			"splitting": true,
		},
	})
	if err != nil {
		t.Fatalf("ParseJsonConfig returned error: %v", err)
	}

	if got := patch.AliasPackages["react"]; got != "preact" {
		t.Fatalf("AliasPackages[react] = %q, want %q", got, "preact")
	}
	if got := patch.ResolveModules["foo"]; got != "./vendor/foo.js" {
		t.Fatalf("ResolveModules[foo] = %q, want %q", got, "./vendor/foo.js")
	}
	if len(patch.PreloadPathsStartingWith) != 1 || patch.PreloadPathsStartingWith[0] != "src/index" {
		t.Fatalf("unexpected PreloadPathsStartingWith: %#v", patch.PreloadPathsStartingWith)
	}
	if len(patch.Injects) != 1 || patch.Injects[0] != "src/shim.js" {
		t.Fatalf("unexpected Injects: %#v", patch.Injects)
	}
	if got, ok := patch.Loaders[".txt"]; !ok {
		t.Fatalf("expected loader override to be preserved: %#v", patch.Loaders)
	} else if gotString, err := StringifyLoader(got); err != nil || gotString != "copy" {
		t.Fatalf("Loaders[.txt] = %q, err=%v, want %q", gotString, err, "copy")
	}
	if !patch.InlineSize.Set || patch.InlineSize.Value != 10000 {
		t.Fatalf("unexpected InlineSize: %#v", patch.InlineSize)
	}
	if len(patch.InlineExtensions) != 2 || patch.InlineExtensions[0] != "svg" || patch.InlineExtensions[1] != "png" {
		t.Fatalf("unexpected InlineExtensions: %#v", patch.InlineExtensions)
	}
	if !patch.Splitting.Set || !patch.Splitting.Value {
		t.Fatalf("unexpected Splitting patch: %#v", patch.Splitting)
	}
}
