package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	"github.com/natrim/nrb/lib"
)

var proxyPort uint16
var protocol string
var broker *lib.Broker

var reloadJS = "(()=>{if(window.esIn)return;window.esIn=true;function c(){var s=new EventSource(\"/esbuild\");s.onerror=()=>{s.close();setTimeout(c,10000)};s.onmessage=()=>{window.location.reload()}}c()})();"

func watch() error {
	esbuildContext, err := startEsbuildServe()
	if err != nil {
		return err
	}

	// wait a bit cause stuff
	time.Sleep(250 * time.Millisecond)

	// schedule esbuild context cleanup
	defer func() {
		if esbuildContext != nil {
			esbuildContext.Dispose()
		}
	}()

	// start file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	//lib.PrintOk("Watcher start done")

	// schedule watcher cleanup
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	lib.PrintInfo("watching:", sourceDir)

	extraWatch := []string{filepath.Join(baseDir, tsConfigPath), filepath.Join(baseDir, packagePath), filepath.Join(baseDir, versionPath)}
	for _, vpath := range extraWatch {
		if lib.FileExists(vpath) {
			if err := watcher.Add(vpath); err != nil {
				return err
			}
			lib.PrintInfo("watching:", vpath)
		}
	}

	absWalkPath := lib.RealQuickPath(sourceDir)
	if err := filepath.WalkDir(absWalkPath, watchDir(watcher)); err != nil {
		return err
	}

	done := make(chan error)
	go func() {
		timer := time.NewTimer(time.Millisecond)
		<-timer.C

		// send done to main on goroutine end
		defer func() {
			done <- nil
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//fmt.Printf("EVENT! %s\n", event.String())
				// skip event that has only chmod operation
				if event.Op == fsnotify.Chmod {
					continue
				}

				if slices.Contains(extraWatch, event.Name) {
					if event.Has(fsnotify.Create | fsnotify.Write) {
						if esbuildContext != nil {
							esbuildContext.Dispose()
							esbuildContext = nil
						}
						buildEsbuildConfig()
						esbuildContext, err = startEsbuildServe()
						if err != nil {
							lib.PrintError(err)
							os.Exit(1)
						}
						lib.PrintReload("Config change detected, reloading esbuild...")
						time.Sleep(250 * time.Millisecond)
						broker.Notifier <- []byte("update")
					}
					continue
				}

				//lastEvent = event
				// event has write operation
				if event.Has(fsnotify.Write) {
					lib.PrintItemf("Change in %s/%s\n", sourceDir, strings.TrimLeft(strings.TrimPrefix(event.Name, absWalkPath), "/"))
				}
				timer.Reset(time.Millisecond * 100)

				// add new directories to watcher if event has create operation
				if event.Has(fsnotify.Create) {
					stat, err := os.Stat(event.Name)
					if err == nil && stat.IsDir() {
						err = filepath.WalkDir(event.Name, watchDir(watcher))
						if err != nil {
							lib.PrintError(err)
						}
					}
				}
				// remove old dirs
				//				if event.Has(fsnotify.Remove) {
				//					stat, err := os.Stat(event.Name)
				//					if err == nil && stat.IsDir() {
				//						err = filepath.WalkDir(event.Name, unwatchDir(watcher))
				//						if err != nil {
				//							lib.PrintError(err)
				//						}
				//					}
				//				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				lib.PrintError(err)
			case <-timer.C:
				lib.PrintReload("Change detected, reloading...")
				broker.Notifier <- []byte("update")
			}
		}
	}()

	go func() {
		if isSecured {
			protocol = "https://"
		} else {
			protocol = "http://"
		}

		//TODO: gzip response
		broker = lib.NewStreamServer()
		fileServer := lib.PipedFileServerWithMiddleware(staticDir, pipeRequestToEsbuild, func(next http.HandlerFunc) http.HandlerFunc {
			return func(writer http.ResponseWriter, request *http.Request) {
				//pipe index directly to esbuild to skip loading of index by staticServer
				if request.URL.Path == "/index.html" || filepath.Ext(request.URL.Path) == "" {
					pipeRequestToEsbuild(writer, request)
					return
				}

				next(writer, request)
			}
		})

		http.HandleFunc("/", fileServer)
		http.Handle("/esbuild", broker)

		socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			done <- err
			return
		}

		// get real port in case user uses 0 for random port
		port = socket.Addr().(*net.TCPAddr).Port

		lib.PrintInfof("Listening on: %s%s:%d\n", protocol, host, port)

		if isSecured {
			err = http.ServeTLS(socket, nil, certFile, keyFile)
		} else {
			err = http.Serve(socket, nil)
		}

		if !errors.Is(err, http.ErrServerClosed) {
			done <- err
			return
		}
	}()

	return <-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(watcher *fsnotify.Watcher) fs.WalkDirFunc {
	return func(path string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && fi.Name() != ".git" && fi.Name() != ".svn" && fi.Name() != ".hg" {
			return watcher.Add(path)
		}
		return nil
	}
}

// unwatchDir gets run as a walk func, searching for directories to remove from watcher
//func unwatchDir(watcher *fsnotify.Watcher) fs.WalkDirFunc {
//	return func(path string, fi os.DirEntry, err error) error {
//		if err != nil {
//			return err
//		}
//		if fi.IsDir() && fi.Name() != ".git" && fi.Name() != ".svn" && fi.Name() != ".hg" {
//			return watcher.Remove(path)
//		}
//		return nil
//	}
//}

func pipeRequestToEsbuild(w http.ResponseWriter, r *http.Request) {
	var uri string
	// if not file request then go to index
	if filepath.Ext(r.URL.Path) == "" {
		uri = "/index.html"
	} else {
		uri = r.URL.RequestURI()
	}

	var isIndex bool
	if uri == "/index.html" {
		isIndex = true
	}

	// normalize path prefix
	uri = strings.TrimPrefix(uri, "/")

	// get the file from esbuild
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // cancel after minute
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://%s:%d/%s", host, proxyPort, uri), nil)
	req.Header.Set("Host", r.Header.Get("Host"))
	setXForwardedFrom(req, r)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		lib.PrintError(err)
		if !errors.Is(err, syscall.EPIPE) {
			error404(w, true)
		}
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 && resp.StatusCode != 204 && resp.StatusCode != 206 {
		// esbuild errors
		if isIndex && resp.StatusCode == http.StatusServiceUnavailable && strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(fmt.Sprintf("<!doctype html><head><meta charset=utf-8><title>error</title><script>%s</script></head><body><pre>", reloadJS)))
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				_, _ = w.Write([]byte("Error: cannot build app"))
			}
			_, _ = w.Write([]byte("</pre></body>"))
			return
		}
		error404(w, true)
		return
	}

	// copy esbuild headers
	copyHeaders(w.Header(), resp.Header)

	// replace custom vars in index.html nad inject js/css scripts from esbuild
	if isIndex {
		// read esbuild response
		readBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lib.PrintError(err)
			if !errors.Is(err, syscall.EPIPE) {
				error404(w, false)
			}
			return
		}

		index, _ := lib.InjectVarsIntoIndex(readBody, entryFileName, assetsDir, publicUrl)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(index)))
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(index)
	} else {
		w.WriteHeader(resp.StatusCode)
		// copy esbuild response
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			lib.PrintError(err)
			if !errors.Is(err, syscall.EPIPE) {
				error404(w, false)
			}
			return
		}
	}
}

func setXForwardedFrom(req *http.Request, src *http.Request) {
	clientIP, _, err := net.SplitHostPort(src.RemoteAddr)
	if err == nil {
		prior := req.Header["X-Forwarded-For"]
		if len(prior) > 0 {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		req.Header.Set("X-Forwarded-For", clientIP)
	} else {
		req.Header.Del("X-Forwarded-For")
	}
	req.Header.Set("X-Forwarded-Host", src.Host)
	if src.TLS == nil {
		req.Header.Set("X-Forwarded-Proto", "http")
	} else {
		req.Header.Set("X-Forwarded-Proto", "https")
	}
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Set(k, v)
		}
	}
}

func error404(res http.ResponseWriter, writeHeader bool) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.Header().Del("Content-Length")
	res.Header().Del("Content-Range")
	if writeHeader {
		res.WriteHeader(http.StatusNotFound)
	}
	_, _ = res.Write([]byte("404 - Not Found"))
}

func startEsbuildServe() (api.BuildContext, error) {
	// inject hot reload watcher to js
	if buildOptions.Banner == nil {
		buildOptions.Banner = map[string]string{"js": reloadJS}
	} else {
		if _, ok := buildOptions.Banner["js"]; !ok {
			buildOptions.Banner["js"] = reloadJS
		} else {
			buildOptions.Banner["js"] = reloadJS + buildOptions.Banner["js"]
		}
	}

	// set outdir
	buildOptions.Outdir = filepath.Join(staticDir, assetsDir)

	// dont write files on watch
	buildOptions.Write = false

	// get esbuild context
	ctx, ctxerr := api.Context(buildOptions)
	if ctxerr != nil {
		return nil, ctxerr
	}

	// start esbuild server
	server, err := ctx.Serve(api.ServeOptions{
		Servedir: staticDir,
		Port:     proxyPort,
		Host:     host,
	})

	if err != nil {
		return nil, err
	}

	// sync values used by esbuild to real used ones
	proxyPort = server.Port
	// nope- host = server.Host

	//lib.PrintOk("Esbuild start done")

	return ctx, nil
}
