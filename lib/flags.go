package lib

import (
	"errors"
	"flag"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

var flagsSet map[string]bool

func IsFlagPassed(name string) bool {
	if flagsSet == nil {
		flagsSet = make(map[string]bool)
		flag.Visit(func(f *flag.Flag) {
			flagsSet[f.Name] = true
		})
	}

	_, found := flagsSet[name]

	return found
}

type ArrayFlags []string

func (i *ArrayFlags) String() string {
	return strings.Join(*i, ",")
}
func (i *ArrayFlags) Set(value string) error {
	f := strings.SplitSeq(value, ",")
	for v := range f {
		*i = append(*i, v)
	}
	return nil
}

type MapFlags map[string]string

func (i *MapFlags) String() string {
	val := strings.Builder{}
	for a, p := range *i {
		val.WriteString(",")
		val.WriteString(a)
		val.WriteString(":")
		val.WriteString(p)
	}
	return strings.TrimPrefix(val.String(), ",")
}
func (i *MapFlags) Set(value string) error {
	f := strings.SplitSeq(value, ",")

	for v := range f {
		alias := strings.SplitN(v, ":", 2)

		if len(alias) < 2 {
			alias = strings.SplitN(v, "=", 2)
		}

		if len(alias) < 2 {
			return errors.New("invalid seperator, use key:val,key1:val1,... ")
		}

		if *i == nil {
			*i = make(MapFlags)
		}

		(*i)[alias[0]] = alias[1]
	}

	return nil
}

type LoaderFlags map[string]api.Loader

func (i *LoaderFlags) String() string {
	val := strings.Builder{}
	for a, p := range *i {
		l, e := StringifyLoader(p)

		if e != nil {
			PrintError(e)
		} else {
			val.WriteString(",")
			val.WriteString(a)
			val.WriteString(":")
			val.WriteString(l)
		}
	}
	return strings.TrimPrefix(val.String(), ",")
}

func (i *LoaderFlags) Set(value string) error {
	f := strings.SplitSeq(value, ",")

	for v := range f {
		alias := strings.SplitN(v, ":", 2)

		if len(alias) < 2 {
			alias = strings.SplitN(v, "=", 2)
		}

		if len(alias) < 2 {
			return errors.New("invalid seperator, use key:val,key1:val1,... ")
		}

		l, e := ParseLoader(alias[1])

		if e != nil {
			return e
		}

		if *i == nil {
			*i = make(LoaderFlags)
		}

		(*i)["."+strings.TrimPrefix(alias[0], ".")] = l
	}
	return nil
}
