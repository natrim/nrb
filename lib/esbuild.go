package lib

import (
	"errors"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
)

func ParseLoader(text string) (api.Loader, error) {
	switch text {
	case "base64":
		return api.LoaderBase64, nil
	case "binary":
		return api.LoaderBinary, nil
	case "copy":
		return api.LoaderCopy, nil
	case "css":
		return api.LoaderCSS, nil
	case "dataurl":
		return api.LoaderDataURL, nil
	case "default":
		return api.LoaderDefault, nil
	case "empty":
		return api.LoaderEmpty, nil
	case "file":
		return api.LoaderFile, nil
	case "global-css":
		return api.LoaderGlobalCSS, nil
	case "js":
		return api.LoaderJS, nil
	case "json":
		return api.LoaderJSON, nil
	case "jsx":
		return api.LoaderJSX, nil
	case "local-css":
		return api.LoaderLocalCSS, nil
	case "text":
		return api.LoaderText, nil
	case "ts":
		return api.LoaderTS, nil
	case "tsx":
		return api.LoaderTSX, nil
	default:
		return api.LoaderNone, fmt.Errorf(
			"invalid loader value: %q, valid values are \"base64\", \"binary\", \"copy\", \"css\", \"dataurl\", \"empty\", \"file\", \"global-css\", \"js\", \"json\", \"jsx\", \"local-css\", \"text\", \"ts\", or \"tsx\"", text,
		)
	}
}

func StringifyLoader(loader api.Loader) (string, error) {
	switch loader {
	case api.LoaderBase64:
		return "base64", nil
	case api.LoaderBinary:
		return "binary", nil
	case api.LoaderCopy:
		return "copy", nil
	case api.LoaderCSS:
		return "css", nil
	case api.LoaderDataURL:
		return "dataurl", nil
	case api.LoaderDefault:
		return "default", nil
	case api.LoaderEmpty:
		return "empty", nil
	case api.LoaderFile:
		return "file", nil
	case api.LoaderGlobalCSS:
		return "global-css", nil
	case api.LoaderJS:
		return "js", nil
	case api.LoaderJSON:
		return "json", nil
	case api.LoaderJSX:
		return "jsx", nil
	case api.LoaderLocalCSS:
		return "local-css", nil
	case api.LoaderText:
		return "text", nil
	case api.LoaderTS:
		return "ts", nil
	case api.LoaderTSX:
		return "tsx", nil
	default:
		return "", errors.New("invalid loader")
	}
}

func ParseBrowserTarget(customBrowserTarget string) (api.Target, error) {
	switch customBrowserTarget {
	case "ES2015", "es2015", "Es2015":
		return api.ES2015, nil
	case "ES2016", "es2016", "Es2016":
		return api.ES2016, nil
	case "ES2017", "es2017", "Es2017":
		return api.ES2017, nil
	case "ES2018", "es2018", "Es2018":
		return api.ES2018, nil
	case "ES2019", "es2019", "Es2019":
		return api.ES2019, nil
	case "ES2020", "es2020", "Es2020":
		return api.ES2020, nil
	case "ES2021", "es2021", "Es2021":
		return api.ES2021, nil
	case "ES2022", "es2022", "Es2022":
		return api.ES2022, nil
	case "ES2023", "es2023", "Es2023":
		return api.ES2023, nil
	case "ES2024", "es2024", "Es2024":
		return api.ES2024, nil
	case "ESNEXT", "esnext", "ESNext", "ESnext":
		return api.ESNext, nil
	case "ES5", "es5", "Es5":
		return api.ES5, nil
	case "ES6", "es6", "Es6":
		return api.ES2015, nil
	case "default", "Default", "none", " ", "":
		return api.DefaultTarget, nil
	default:
		return api.DefaultTarget, fmt.Errorf("unsupported target: %q, valid targets are \"es5\", \"es6\", \"es2015\", \"es2016\", \"es2017\", \"es2018\", \"es2019\", \"es2020\", \"es2021\", \"es2022\", \"es2023\", \"es2024\", \"esnext\", \"default\"", customBrowserTarget)
	}
}
