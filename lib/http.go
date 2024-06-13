package lib

import (
	"net/http"
	"path/filepath"
)

// Middleware is a definition of  what a middleware is,
// take in one handlerfunc and wrap it within another handlerfunc
type Middleware func(http.HandlerFunc) http.HandlerFunc

// BuildChain builds the middlware chain recursively, functions are first class
func BuildChain(f http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	// if our chain is done, use the original handlerfunc
	if len(m) == 0 {
		return f
	}
	// otherwise nest the handlerfuncs
	return m[0](BuildChain(f, m[1:cap(m)]...))
}

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
func PipedFileServerWithMiddleware(baseDir string, pipe http.HandlerFunc, middleware func(next http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	return middleware(PipedFileServer(baseDir, pipe))
}
