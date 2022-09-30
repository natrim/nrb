package plugins

import (
	"github.com/evanw/esbuild/pkg/api"
	"os"
	"strings"
)

// InlinePluginDefault is a plugin to inline your images with default settings
func InlinePluginDefault() api.Plugin {
	return InlinePlugin(0, nil)
}

// InlinePlugin is a plugin to inline your images
func InlinePlugin(customLimit int64, customExtensions []string) api.Plugin {
	extensions := customExtensions
	if len(customExtensions) == 0 {
		extensions = []string{"svg", "png", "jpeg", "jpg", "gif", "webp", "avif"}
	}
	limit := customLimit
	if limit <= 0 {
		limit = 100000
	}
	var filter = "\\.("
	for _, ext := range extensions {
		if ext != "" {
			filter += escapeRegExp(strings.TrimPrefix(ext, ".")) + "|"
		}
	}
	filter = strings.TrimSuffix(filter, "|")
	filter += ")$"
	return api.Plugin{
		Name: "inline",
		Setup: func(build api.PluginBuild) {
			if build.InitialOptions.Loader == nil {
				build.InitialOptions.Loader = map[string]api.Loader{}
			}
			for _, ext := range extensions {
				if ext != "" {
					build.InitialOptions.Loader["."+strings.TrimPrefix(ext, ".")] = api.LoaderFile
				}
			}

			build.OnLoad(api.OnLoadOptions{Filter: filter},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					stat, err := os.Stat(args.Path)
					if err != nil {
						return api.OnLoadResult{}, err
					}
					if stat.Size() < limit {
						bytes, err := os.ReadFile(args.Path)
						if err != nil {
							return api.OnLoadResult{}, err
						}
						contents := string(bytes)
						return api.OnLoadResult{
							Contents: &contents,
							Loader:   api.LoaderDataURL,
						}, nil

					}
					return api.OnLoadResult{
						Loader: api.LoaderFile,
					}, nil
				})
		},
	}
}
