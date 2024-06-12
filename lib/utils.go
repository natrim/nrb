package lib

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()

	if err != nil {
		return nil, err
	}

	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

// NotFoundRedirectRespWr response for neutered fs
type NotFoundRedirectRespWr struct {
	http.ResponseWriter // We embed http.ResponseWriter
	status              int
}

func (w *NotFoundRedirectRespWr) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *NotFoundRedirectRespWr) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

// WrappedFileServer wraps your http.FileServer with neutered fs to remove dir browsing and redirects 404 to index.html
func WrappedFileServer(baseDir string) http.HandlerFunc {
	h := http.FileServer(neuteredFileSystem{http.Dir(baseDir)})

	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &NotFoundRedirectRespWr{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == 404 {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, filepath.Join(baseDir, "index.html"))
		}
	}
}

// PipedFileServer wraps your http.FileServer with neutered fs to remove dir browsing and pipes 404 to next handler
func PipedFileServer(baseDir string, pipe http.HandlerFunc) http.HandlerFunc {
	h := http.FileServer(neuteredFileSystem{http.Dir(baseDir)})

	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &NotFoundRedirectRespWr{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == 404 {
			pipe(w, r)
		}
	}
}

// PipedFileServerWithMiddleware wraps PipedFileServer with middleware
func PipedFileServerWithMiddleware(baseDir string, pipe http.HandlerFunc, middleware func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	return middleware(PipedFileServer(baseDir, pipe))
}

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
		in, _ := os.Open(path)
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

// IsErrorAddressAlreadyInUse checks if the error is bind: address already in use OR alternative
func IsErrorAddressAlreadyInUse(err error) bool {
	var eOsSyscall *os.SyscallError
	if !errors.As(err, &eOsSyscall) {
		return false
	}
	var errErrno syscall.Errno // doesn't need a "*" (ptr) because it's already a ptr (uintptr)
	if !errors.As(eOsSyscall, &errErrno) {
		return false
	}
	if errors.Is(errErrno, syscall.EADDRINUSE) {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}

func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
