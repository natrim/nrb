package plugins

import (
	"github.com/evanw/esbuild/pkg/api"
	"strings"
)

func AliasPlugin(aliases map[string]string) api.Plugin {
	var filter = "^("
	for alias, _ := range aliases {
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
						Path: aliases[args.Path],
					}, nil
				})
		},
	}
}
