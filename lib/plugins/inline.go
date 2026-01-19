package plugins

import (
	"os"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// InlinePluginDefault is a plugin to inline your images with default settings
func InlinePluginDefault() api.Plugin {
	return InlinePlugin(100000, []string{"svg", "png", "jpeg", "jpg", "gif", "webp", "avif"})
}

// InlinePlugin is a plugin to inline your images
func InlinePlugin(limit int64, extensions []string) api.Plugin {
	if len(extensions) == 0 {
		return api.Plugin{
			Name: "inline-stub",
			Setup: func(build api.PluginBuild) {
			},
		}
	}
	return api.Plugin{
		Name: "inline",
		Setup: func(build api.PluginBuild) {
			if build.InitialOptions.Loader == nil {
				build.InitialOptions.Loader = map[string]api.Loader{}
			}

			filter := strings.Builder{}
			filter.WriteString("\\.(")
			count := 0
			for _, ext := range extensions {
				if ext != "" {
					build.InitialOptions.Loader["."+strings.TrimPrefix(ext, ".")] = api.LoaderFile

					if count > 0 {
						filter.WriteString("|")
					}
					filter.WriteString(escapeRegExp(strings.TrimPrefix(ext, ".")))
					count++
				}
			}
			filter.WriteString(")$")

			build.OnLoad(api.OnLoadOptions{Filter: filter.String()},
				func(args api.OnLoadArgs) (api.OnLoadResult, error) {
					var inline bool
					if limit <= 0 {
						inline = true
					} else {
						stat, err := os.Stat(args.Path)
						if err != nil {
							return api.OnLoadResult{}, err
						}
						inline = stat.Size() < limit
					}
					if inline {
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
