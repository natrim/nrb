package plugins

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
	"strings"
)

func AliasPlugin(aliases map[string]string) api.Plugin {
	if len(aliases) == 0 {
		return api.Plugin{
			Name: "alias-stub",
			Setup: func(build api.PluginBuild) {
			},
		}
	}
	var filter = "^("
	for alias := range aliases {
		filter += escapeRegExp(alias) + "|"
	}
	filter = strings.TrimSuffix(filter, "|")
	filter += ")$"
	return api.Plugin{
		Name: "alias",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(api.OnResolveOptions{Filter: filter},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{
						Path: lib.RealQuickPath(aliases[args.Path]),
					}, nil
				})
		},
	}
}
