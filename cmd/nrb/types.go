package main

import (
	"errors"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"strings"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}
func (i *arrayFlags) Set(value string) error {
	f := strings.Split(value, ",")
	for _, v := range f {
		*i = append(*i, v)
	}
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
	f := strings.Split(value, ",")

	for _, v := range f {
		alias := strings.SplitN(v, ":", 2)

		if len(alias) < 2 {
			alias = strings.SplitN(v, "=", 2)
		}

		if len(alias) < 2 {
			return errors.New("invalid seperator, use key:val,key1:val1,... ")
		}

		if *i == nil {
			*i = make(mapFlags)
		}

		(*i)[alias[0]] = alias[1]
	}

	return nil
}

type loaderFlags map[string]api.Loader

func (i *loaderFlags) String() string {
	val := ""
	for a, p := range *i {
		l, e := lib.StringifyLoader(p)

		if e != nil {
			lib.PrintError(e)
		} else {
			val = val + "," + a + ":" + l
		}
	}
	return strings.TrimPrefix(val, ",")
}

func (i *loaderFlags) Set(value string) error {
	f := strings.Split(value, ",")

	for _, v := range f {
		alias := strings.SplitN(v, ":", 2)

		if len(alias) < 2 {
			alias = strings.SplitN(v, "=", 2)
		}

		if len(alias) < 2 {
			return errors.New("invalid seperator, use key:val,key1:val1,... ")
		}

		l, e := lib.ParseLoader(alias[1])

		if e != nil {
			return e
		}

		if *i == nil {
			*i = make(loaderFlags)
		}

		(*i)["."+strings.TrimPrefix(alias[0], ".")] = l
	}
	return nil
}

type Config struct {
	AliasPackages            mapFlags
	ResolveModules           mapFlags
	PreloadPathsStartingWith arrayFlags
	Injects                  arrayFlags
	InlineSize               int64
	InlineExtensions         []string
	Loaders                  loaderFlags
	Splitting                bool
}

type PackageJson map[string]any

type VersionData map[string]any
