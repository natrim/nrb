package lib

import (
	"io"
	"io/fs"
	"os"
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
