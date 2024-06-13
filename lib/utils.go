package lib

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CopyDir copies the content of src to dst. src should be a full path.
func CopyDir(dst, src string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))

		if info.IsDir() {
			_ = os.MkdirAll(outpath, info.Mode())
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(in *os.File) {
			_ = in.Close()
		}(in)

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer func(fh *os.File) {
			_ = fh.Close()
		}(fh)

		// make it the same
		_ = fh.Chmod(info.Mode())

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}

// FileExists does?
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

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

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
