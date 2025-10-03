package plugins

import (
	"strings"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/natrim/nrb/lib"
)

func AliasPlugin(aliases map[string]string) api.Plugin {
	if len(aliases) == 0 {
		return api.Plugin{
			Name: "alias-stub",
			Setup: func(build api.PluginBuild) {
			},
		}
	}
	var filter strings.Builder
	filter.WriteString("^(")
	count := len(aliases)
	for alias := range aliases {
		filter.WriteString(escapeRegExp(alias))
		if count > 1 {
			filter.WriteString("|")
		}
		count--
	}
	filter.WriteString(")$")
	return api.Plugin{
		Name: "alias",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(api.OnResolveOptions{Filter: filter.String()},
				func(args api.OnResolveArgs) (api.OnResolveResult, error) {
					return api.OnResolveResult{
						Path: lib.RealQuickPath(aliases[args.Path]),
					}, nil
				})
		},
	}
}
