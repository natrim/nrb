package main

import "strings"

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type mapFlags map[string]string

func (i *mapFlags) String() string {
	val := ""
	for a, p := range *i {
		val = val + "," + a + ":" + p
	}
	return strings.TrimPrefix(val, ",")
}
func (i *mapFlags) Set(value string) error {
	alias := strings.SplitN(value, ":", 2)
	(*i)[alias[0]] = alias[1]
	return nil
}

type Config struct {
	AliasPackages            mapFlags
	ResolveModules           mapFlags
	PreloadPathsStartingWith arrayFlags
	Injects                  arrayFlags
	InlineSize               int64
	InlineExtensions         []string
}

type PackageJson map[string]any

type VersionData map[string]any
