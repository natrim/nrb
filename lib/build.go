package lib

import (
	"bytes"
	"path/filepath"
	"strings"
)

// InjectVarsIntoIndex injects js/css import to index.html content, returns bool if injected into content
func InjectVarsIntoIndex(indexFile []byte, entryFileName, assetsDir, publicUrl string) ([]byte, bool) {
	indexFileName := strings.TrimSuffix(filepath.Base(entryFileName), filepath.Ext(entryFileName))
	publicUrl = strings.TrimSuffix(publicUrl, "/")
	changed := false

	//inject main js/css if not already in index.html
	if !bytes.Contains(indexFile, []byte("/"+assetsDir+"/"+indexFileName+".css")) {
		changed = true
		indexFile = bytes.Replace(indexFile, []byte("</head>"), []byte("<link rel=\"preload\" href=\""+publicUrl+"/"+assetsDir+"/"+indexFileName+".css\" as=\"style\">\n<link rel=\"stylesheet\" href=\""+publicUrl+"/"+assetsDir+"/"+indexFileName+".css\">\n</head>"), 1)
	}
	if !bytes.Contains(indexFile, []byte("/"+assetsDir+"/"+indexFileName+".js")) {
		changed = true
		indexFile = bytes.Replace(indexFile, []byte("</body>"), []byte("<script type=\"module\" src=\""+publicUrl+"/"+assetsDir+"/"+indexFileName+".js\"></script>\n</body>"), 1)
		indexFile = bytes.Replace(indexFile, []byte("</head>"), []byte("<link rel=\"modulepreload\" href=\""+publicUrl+"/"+assetsDir+"/"+indexFileName+".js\">\n</head>"), 1)
	}

	// replace %PUBLIC_URL%
	if bytes.Contains(indexFile, []byte("%PUBLIC_URL%")) {
		changed = true
		indexFile = bytes.ReplaceAll(indexFile, []byte("%PUBLIC_URL%"), []byte(publicUrl))
	}

	return indexFile, changed
}
