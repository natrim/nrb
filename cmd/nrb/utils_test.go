package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
)

func TestParseFlagsReturnsExplicitOverrides(t *testing.T) {
	originalArgs := os.Args
	originalFlagSet := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalFlagSet
		resetCLIParsingState()
	})

	resetCLIParsingState()
	os.Args = []string{
		"nrb",
		"--env=.env.local,.env.dev",
		"--outputDir=dist",
		"--splitting=false",
		"build",
	}

	state, overrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	if state.IsHelp {
		t.Fatal("expected IsHelp to be false")
	}
	if state.IsVersion {
		t.Fatal("expected IsVersion to be false")
	}
	if !state.UseColor {
		t.Fatal("expected UseColor to default to true")
	}
	if state.EnvFiles != ".env.local,.env.dev" {
		t.Fatalf("expected env files to be preserved, got %q", state.EnvFiles)
	}

	if !overrides.OutputDir.Set {
		t.Fatal("expected OutputDir override to be marked as set")
	}
	if overrides.OutputDir.Value != "dist" {
		t.Fatalf("expected OutputDir override value %q, got %q", "dist", overrides.OutputDir.Value)
	}

	if !overrides.Splitting.Set {
		t.Fatal("expected Splitting override to be marked as set")
	}
	if overrides.Splitting.Value {
		t.Fatal("expected explicit false splitting override")
	}

	if overrides.SourceDir.Set {
		t.Fatal("did not expect SourceDir override when flag was not passed")
	}
}

func TestParseFlagsDoesNotLeakPassedFlagsAcrossCalls(t *testing.T) {
	originalArgs := os.Args
	originalFlagSet := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalFlagSet
		resetCLIParsingState()
	})

	resetCLIParsingState()
	os.Args = []string{
		"nrb",
		"--outputDir=dist",
		"--splitting=false",
		"build",
	}

	_, firstOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("first ParseFlags returned error: %v", err)
	}

	if !firstOverrides.OutputDir.Set {
		t.Fatal("expected first call to set OutputDir override")
	}
	if !firstOverrides.Splitting.Set {
		t.Fatal("expected first call to set Splitting override")
	}

	resetCLIParsingState()
	os.Args = []string{
		"nrb",
		"--env=.env.test",
		"build",
	}

	secondState, secondOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("second ParseFlags returned error: %v", err)
	}

	if secondState.EnvFiles != ".env.test" {
		t.Fatalf("expected second call env files %q, got %q", ".env.test", secondState.EnvFiles)
	}
	if secondOverrides.OutputDir.Set {
		t.Fatal("did not expect second call to inherit OutputDir override")
	}
	if secondOverrides.Splitting.Set {
		t.Fatal("did not expect second call to inherit Splitting override")
	}
}

func TestParseFlagsDoesNotAccumulateCollectionOverridesAcrossCalls(t *testing.T) {
	originalArgs := os.Args
	originalFlagSet := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalFlagSet
		resetCLIParsingState()
	})

	resetCLIParsingState()
	os.Args = []string{
		"nrb",
		"--preload=src/index,node_modules/react",
		"--alias=react:preact-compat",
		"build",
	}

	_, firstOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("first ParseFlags returned error: %v", err)
	}

	if len(firstOverrides.PreloadPathsStartingWith) != 2 {
		t.Fatalf("expected 2 preload entries on first call, got %d", len(firstOverrides.PreloadPathsStartingWith))
	}
	if len(firstOverrides.AliasPackages) != 1 {
		t.Fatalf("expected 1 alias entry on first call, got %d", len(firstOverrides.AliasPackages))
	}

	os.Args = []string{
		"nrb",
		"--preload=src/other",
		"--alias=redux:custom-redux",
		"build",
	}

	_, secondOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("second ParseFlags returned error: %v", err)
	}

	if got := len(secondOverrides.PreloadPathsStartingWith); got != 1 {
		t.Fatalf("expected second call to have 1 preload entry, got %d: %#v", got, secondOverrides.PreloadPathsStartingWith)
	}
	if secondOverrides.PreloadPathsStartingWith[0] != "src/other" {
		t.Fatalf("expected second call preload entry %q, got %q", "src/other", secondOverrides.PreloadPathsStartingWith[0])
	}

	if got := len(secondOverrides.AliasPackages); got != 1 {
		t.Fatalf("expected second call to have 1 alias entry, got %d: %#v", got, secondOverrides.AliasPackages)
	}
	if secondOverrides.AliasPackages["redux"] != "custom-redux" {
		t.Fatalf("expected second call alias redux=custom-redux, got %#v", secondOverrides.AliasPackages)
	}
	if _, exists := secondOverrides.AliasPackages["react"]; exists {
		t.Fatalf("did not expect second call to inherit react alias: %#v", secondOverrides.AliasPackages)
	}
}

func TestPrepareBuildPreloadPathsUsesMergedConfig(t *testing.T) {
	originalArgs := os.Args
	originalFlagSet := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalFlagSet
		resetCLIParsingState()
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()
	writePackageJSON(t, tempDir, `{"nrb":{"preload":["pkg/preload"]}}`)

	resetCLIParsingState()
	resetRuntimeBridgeState()
	baseDir = tempDir

	os.Args = []string{
		"nrb",
		"--preload=cli/preload",
		"build",
	}

	_, overrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}

	configOverrides = overrides
	buildEsbuildConfig(true)

	preloadPaths := config.PreloadPathsStartingWith
	if got := len(preloadPaths); got != 1 {
		t.Fatalf("expected 1 preload path, got %d: %#v", got, preloadPaths)
	}
	if preloadPaths[0] != "cli/preload" {
		t.Fatalf("expected build preload %q, got %#v", "cli/preload", preloadPaths)
	}
}

func TestBuildEsbuildConfigDoesNotUseStaleLegacyCollectionGlobals(t *testing.T) {
	originalArgs := os.Args
	originalFlagSet := flag.CommandLine

	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalFlagSet
		resetCLIParsingState()
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()
	writePackageJSON(t, tempDir, `{"nrb":{"preload":["pkg/preload"],"alias":{"react":"pkg-react"}}}`)

	resetCLIParsingState()
	resetRuntimeBridgeState()
	baseDir = tempDir

	os.Args = []string{
		"nrb",
		"--preload=cli/preload",
		"--alias=react:cli-react",
		"build",
	}

	_, firstOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("first ParseFlags returned error: %v", err)
	}
	configOverrides = firstOverrides

	resetRuntimeBridgeState()
	baseDir = tempDir

	os.Args = []string{
		"nrb",
		"build",
	}

	_, secondOverrides, err := ParseFlags()
	if err != nil {
		t.Fatalf("second ParseFlags returned error: %v", err)
	}
	configOverrides = secondOverrides
	buildEsbuildConfig(true)

	if got := len(config.PreloadPathsStartingWith); got != 1 {
		t.Fatalf("expected package preload only, got %d: %#v", got, config.PreloadPathsStartingWith)
	}
	if config.PreloadPathsStartingWith[0] != "pkg/preload" {
		t.Fatalf("expected package preload %q, got %#v", "pkg/preload", config.PreloadPathsStartingWith)
	}
	if got := config.AliasPackages["react"]; got != "pkg-react" {
		t.Fatalf("expected package alias react=pkg-react, got %#v", config.AliasPackages)
	}
}

func TestBuildRuntimeConfigMergesDefaultsPackageAndCLIOverrides(t *testing.T) {
	t.Cleanup(func() {
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()
	writePackageJSON(t, tempDir, `{"nrb":{"publicUrl":"/pkg/","host":"0.0.0.0","preload":["pkg/preload"]}}`)

	resetRuntimeBridgeState()
	baseDir = tempDir
	configOverrides = lib.ConfigOverrides{
		OutputDir:                lib.OptionalString{Value: "dist", Set: true},
		Port:                     lib.OptionalInt{Value: 4567, Set: true},
		PreloadPathsStartingWith: lib.ArrayFlags{"cli/preload"},
	}

	mergedConfig, err := buildRuntimeConfig(true)
	if err != nil {
		t.Fatalf("buildRuntimeConfig returned error: %v", err)
	}

	if mergedConfig.SourceDir != "src" {
		t.Fatalf("expected default source dir %q, got %q", "src", mergedConfig.SourceDir)
	}
	if mergedConfig.PublicURL != "/pkg/" {
		t.Fatalf("expected package public url %q, got %q", "/pkg/", mergedConfig.PublicURL)
	}
	if mergedConfig.Host != "0.0.0.0" {
		t.Fatalf("expected package host %q, got %q", "0.0.0.0", mergedConfig.Host)
	}
	if mergedConfig.OutputDir != "dist" {
		t.Fatalf("expected CLI output dir %q, got %q", "dist", mergedConfig.OutputDir)
	}
	if mergedConfig.Port != 4567 {
		t.Fatalf("expected CLI port %d, got %d", 4567, mergedConfig.Port)
	}
	if got := len(mergedConfig.PreloadPathsStartingWith); got != 1 {
		t.Fatalf("expected CLI preload override to replace package values, got %d: %#v", got, mergedConfig.PreloadPathsStartingWith)
	}
	if mergedConfig.PreloadPathsStartingWith[0] != "cli/preload" {
		t.Fatalf("expected CLI preload %q, got %#v", "cli/preload", mergedConfig.PreloadPathsStartingWith)
	}
}

func TestBuildEsbuildConfigUsesFinalMergedConfigForEnvParsing(t *testing.T) {
	originalAppGreeting := os.Getenv("APP_GREETING")

	t.Cleanup(func() {
		if originalAppGreeting == "" {
			os.Unsetenv("APP_GREETING")
		} else {
			os.Setenv("APP_GREETING", originalAppGreeting)
		}
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()
	writePackageJSON(t, tempDir, `{"nrb":{"publicUrl":"/pkg/","sourceDir":"client","entryFileName":"main.ts","outputDir":"dist","assetsDir":"static"}}`)

	resetRuntimeBridgeState()
	baseDir = tempDir
	configOverrides = lib.ConfigOverrides{
		EnvPrefix: lib.OptionalString{Value: "APP_", Set: true},
	}

	if err := os.Setenv("APP_GREETING", "hello"); err != nil {
		t.Fatalf("failed to set APP_GREETING: %v", err)
	}

	buildEsbuildConfig(false)

	lastTwoBase := func(s string) string {
		f := filepath.Base(s)
		dir := filepath.Base(filepath.Dir(s))
		return filepath.Join(dir, f)
	}

	if got := definedReplacements["process.env.PUBLIC_URL"]; got != "\"/pkg\"" {
		t.Fatalf("expected merged public url define %q, got %q", "\"/pkg\"", got)
	}
	if got := definedReplacements["import.meta.env.BASE_URL"]; got != "\"/pkg\"" {
		t.Fatalf("expected merged base url define %q, got %q", "\"/pkg\"", got)
	}
	if got := definedReplacements["process.env.APP_GREETING"]; got != "\"hello\"" {
		t.Fatalf("expected env define for CLI prefix, got %q", got)
	}
	if got := buildOptions.EntryPoints[0]; lastTwoBase(got) != filepath.Join("client", "main.ts") {
		t.Fatalf("expected entry point %q, got %q", filepath.Join("client", "main.ts"), got)
	}
	if got := buildOptions.Outdir; lastTwoBase(got) != filepath.Join("dist", "static") {
		t.Fatalf("expected outdir %q, got %q", filepath.Join("dist", "static"), got)
	}
}

func TestRefreshRuntimeConfigForServeDoesNotRequirePackageJSON(t *testing.T) {
	t.Cleanup(func() {
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()

	resetRuntimeBridgeState()
	baseDir = tempDir
	configOverrides = lib.ConfigOverrides{
		OutputDir: lib.OptionalString{Value: "dist", Set: true},
		Port:      lib.OptionalInt{Value: 4321, Set: true},
	}

	if err := refreshRuntimeConfig(false); err != nil {
		t.Fatalf("refreshRuntimeConfig(false) returned error without package.json: %v", err)
	}

	if got := config.OutputDir; filepath.Base(got) != "dist" {
		t.Fatalf("expected CLI output dir %q, got %q", "dist", got)
	}
	if got := config.Port; got != 4321 {
		t.Fatalf("expected CLI port %d, got %d", 4321, got)
	}
	if got := config.SourceDir; filepath.Base(got) != "src" {
		t.Fatalf("expected default source dir %q, got %q", "src", got)
	}
}

func TestBuildEsbuildConfigRegeneratesDefinesWhenMergedConfigChanges(t *testing.T) {
	originalAppGreeting := os.Getenv("APP_GREETING")
	originalWebGreeting := os.Getenv("WEB_GREETING")

	t.Cleanup(func() {
		if originalAppGreeting == "" {
			os.Unsetenv("APP_GREETING")
		} else {
			os.Setenv("APP_GREETING", originalAppGreeting)
		}
		if originalWebGreeting == "" {
			os.Unsetenv("WEB_GREETING")
		} else {
			os.Setenv("WEB_GREETING", originalWebGreeting)
		}
		resetRuntimeBridgeState()
	})

	tempDir := t.TempDir()

	if err := os.Setenv("APP_GREETING", "hello-app"); err != nil {
		t.Fatalf("failed to set APP_GREETING: %v", err)
	}
	if err := os.Setenv("WEB_GREETING", "hello-web"); err != nil {
		t.Fatalf("failed to set WEB_GREETING: %v", err)
	}

	resetRuntimeBridgeState()
	baseDir = tempDir

	writePackageJSON(t, tempDir, `{"nrb":{"publicUrl":"/app/"}}`)
	configOverrides = lib.ConfigOverrides{
		EnvPrefix: lib.OptionalString{Value: "APP_", Set: true},
	}

	buildEsbuildConfig(false)

	if got := definedReplacements["process.env.PUBLIC_URL"]; got != "\"/app\"" {
		t.Fatalf("expected first PUBLIC_URL define %q, got %q", "\"/app\"", got)
	}
	if got := definedReplacements["import.meta.env.BASE_URL"]; got != "\"/app\"" {
		t.Fatalf("expected first BASE_URL define %q, got %q", "\"/app\"", got)
	}
	if got := definedReplacements["process.env.APP_GREETING"]; got != "\"hello-app\"" {
		t.Fatalf("expected APP define %q, got %q", "\"hello-app\"", got)
	}

	writePackageJSON(t, tempDir, `{"nrb":{"publicUrl":"/web/"}}`)
	configOverrides = lib.ConfigOverrides{
		EnvPrefix: lib.OptionalString{Value: "WEB_", Set: true},
	}

	buildEsbuildConfig(false)

	if got := definedReplacements["process.env.PUBLIC_URL"]; got != "\"/web\"" {
		t.Fatalf("expected second PUBLIC_URL define %q, got %q", "\"/web\"", got)
	}
	if got := definedReplacements["import.meta.env.BASE_URL"]; got != "\"/web\"" {
		t.Fatalf("expected second BASE_URL define %q, got %q", "\"/web\"", got)
	}
	if got := definedReplacements["process.env.WEB_GREETING"]; got != "\"hello-web\"" {
		t.Fatalf("expected WEB define %q, got %q", "\"hello-web\"", got)
	}
	if _, exists := definedReplacements["process.env.APP_GREETING"]; exists {
		t.Fatalf("did not expect old APP define to remain after config change: %#v", definedReplacements)
	}
}

func resetCLIParsingState() {
	cfg := lib.DefaultConfig()
	config = &cfg
	cliState = CLIState{}
}

func resetRuntimeBridgeState() {
	cfg := lib.DefaultConfig()
	config = &cfg
	configOverrides = lib.ConfigOverrides{}
	envLoaded = false
	baseDir = "."
	packagePath = "package.json"
	versionData = "dev"
	definedReplacements = nil
	buildOptions = api.BuildOptions{}
}

func writePackageJSON(t *testing.T, dir string, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(contents), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}
}
