package main

import (
	"errors"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	"github.com/natrim/nrb/lib"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var proxyPort uint16
var server api.ServeResult
var protocol string
var broker *lib.Broker

var reloadJS = "(()=>{if(window.esIn)return;window.esIn=true;function c(){var s=new EventSource(\"/esbuild\");s.onopen=()=>{s.onerror=()=>{s.close();setTimeout(c,10000)};s.onmessage=()=>{window.location.reload()}}c()})();"

func watch() {
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

	done := make(chan bool)

	go func() {
		buildOptions.Outdir = filepath.Join(staticDir, assetsDir)
		server, err := api.Serve(api.ServeOptions{
			Servedir: staticDir,
			Port:     proxyPort,
			Host:     host,
		}, buildOptions)

		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// sync values used by esbuild to real used ones
		proxyPort = server.Port
		// nope- host = server.Host

		// wait a bit cause
		time.Sleep(250 * time.Millisecond)

		// send to next step
		done <- true

		// wait for esbuild serve to die
		_ = server.Wait()

		// send next step
		done <- true
	}()

	<-done
	defer server.Stop()

	var err error

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	fmt.Println(INFO, "watching:", sourceDir)
	tspath := filepath.Join(baseDir, "tsconfig.json")
	if lib.FileExists(tspath) {
		if err := watcher.Add(tspath); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(INFO, "watching:", "tsconfig.json")
	}

	absWalkPath := lib.RealQuickPath(sourceDir)
	if err := filepath.WalkDir(absWalkPath, watchDir(watcher)); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	go func() {
		var (
			timer *time.Timer
			//			lastEvent fsnotify.Event
		)
		timer = time.NewTimer(time.Millisecond)
		<-timer.C

		defer func() {
			done <- true
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//fmt.Printf("EVENT! %s\n", event.String())
				// skip only chmod events
				if event.Op == fsnotify.Chmod {
					continue
				}
				//lastEvent = event
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Printf(DASH+" Change in %s%s\n", sourceDir, strings.TrimPrefix(event.Name, absWalkPath))
				}
				timer.Reset(time.Millisecond * 100)

				// add new directories to watcher
				if event.Op&fsnotify.Create == fsnotify.Create {
					stat, err := os.Stat(event.Name)
					if err == nil && stat.IsDir() {
						err = filepath.WalkDir(event.Name, watchDir(watcher))
						if err != nil {
							_, _ = fmt.Fprintln(os.Stderr, err)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				_, _ = fmt.Fprintln(os.Stderr, err)
			case <-timer.C:
				fmt.Printf(RELOAD + " Change detected, reloading...\n")
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
			if lib.IsErrorAddressAlreadyInUse(err) {
				socket, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, 0))
				if err != nil {
					_, _ = fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				_, _ = fmt.Fprintln(os.Stderr, ERR, "port", port, "is in use")
				port = socket.Addr().(*net.TCPAddr).Port
			} else {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		fmt.Printf(INFO+" Listening on: %s%s:%d\n", protocol, host, port)

		if isSecured {
			err = http.ServeTLS(socket, nil, certFile, keyFile)
		} else {
			err = http.Serve(socket, nil)
		}

		if !errors.Is(err, http.ErrServerClosed) {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	<-done
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(watcher *fsnotify.Watcher) fs.WalkDirFunc {
	return func(path string, fi os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return watcher.Add(path)
		}
		return nil
	}
}

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
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", host, proxyPort, uri))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		if !errors.Is(err, syscall.EPIPE) {
			error404(w, true)
		}
		return
	}

	if resp.StatusCode != 200 {
		// esbuild errors
		if isIndex && resp.StatusCode == http.StatusServiceUnavailable && strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(fmt.Sprintf("<!doctype html><head><meta charset=utf-8><title>error</title><script>%s</script></head><body><pre>", reloadJS)))
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				_, _ = w.Write([]byte("Error: cannot build app"))
			}
			_, _ = w.Write([]byte("</pre></body>"))
			_ = resp.Body.Close()
			return
		}
		_ = resp.Body.Close()
		error404(w, true)
		return
	}

	// copy esbuild response
	readBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		if !errors.Is(err, syscall.EPIPE) {
			error404(w, false)
		}
		return
	}
	_ = resp.Body.Close()

	// copy esbuild headers
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	//	content-length needs to be potentionaly customized
	w.Header().Set("Access-Control-Allow-Origin", resp.Header.Get("Access-Control-Allow-Origin"))
	w.Header().Set("Date", resp.Header.Get("Date"))

	// replace custom vars in index.html nad inject js/css scripts from esbuild
	if isIndex {
		index, _ := lib.InjectVarsIntoIndex(readBody, entryFileName, assetsDir, publicUrl)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(index)))
		_, _ = w.Write(index)
		return
	}

	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	_, _ = w.Write(readBody)
}

func error404(res http.ResponseWriter, writeHeader bool) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if writeHeader {
		res.WriteHeader(http.StatusNotFound)
	}
	_, _ = res.Write([]byte("404 - Not Found"))
}
