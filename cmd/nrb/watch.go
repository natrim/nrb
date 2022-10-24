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

var reloadJS = "(()=>{if(window.esIn)return;window.esIn=true;function c(){var s=new EventSource(\"/esbuild\");s.onopen=()=>{console.log('hot reload connected')};s.onerror=()=>{s.close();console.error('hot reload failed to init');setTimeout(c,10000)};s.onmessage=()=>{console.log('hot reload received');location.reload()}}c()})();"

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
			fmt.Println(err)
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
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	fmt.Println("> watching:", sourceDir)
	tspath := filepath.Join(baseDir, "tsconfig.json")
	if lib.FileExists(tspath) {
		if err := watcher.Add(tspath); err != nil {
			fmt.Println("ERROR", err)
			os.Exit(1)
		}
		fmt.Println("> watching:", "tsconfig.json")
	}

	absWalkPath := lib.RealQuickPath(sourceDir)
	if err := filepath.WalkDir(absWalkPath, watchDir(watcher)); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	go func() {
		var (
			timer     *time.Timer
			lastEvent fsnotify.Event
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
				lastEvent = event
				timer.Reset(time.Millisecond * 100)

				// add new directories to watcher
				if event.Op&fsnotify.Create == fsnotify.Create {
					stat, err := os.Stat(event.Name)
					if err == nil && stat.IsDir() {
						err = filepath.WalkDir(event.Name, watchDir(watcher))
						if err != nil {
							fmt.Println(err)
						}
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("ERROR", err)
			case <-timer.C:
				fmt.Printf("↻ Change in %s%s\n", sourceDir, strings.TrimPrefix(lastEvent.Name, absWalkPath))
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
				// load the main index file and check if it contains js/css import
				if request.URL.Path == "/index.html" {
					if index, err := os.ReadFile(filepath.Join(staticDir, "index.html")); err == nil {
						writer.Header().Set("Content-Type", "text/html; charset=utf-8")
						writer.WriteHeader(200)
						index, _ := lib.InjectJSCSSToIndex(index, entryFileName, assetsDir)
						_, _ = writer.Write(index)
						return
					}
				}
				next(writer, request)
			}
		})

		http.HandleFunc("/", fileServer)
		http.Handle("/esbuild", broker)

		socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, wwwPort))
		if err != nil {
			if lib.IsErrorAddressAlreadyInUse(err) {
				socket, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, 0))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				wwwPort = socket.Addr().(*net.TCPAddr).Port
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Printf("> Listening on: %s%s:%d\n", protocol, host, wwwPort)

		if isSecured {
			err = http.ServeTLS(socket, nil, certFile, keyFile)
		} else {
			err = http.Serve(socket, nil)
		}

		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println(err)
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

	// normalize path prefix
	uri = strings.TrimPrefix(uri, "/")

	// get the file from esbuild
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", host, proxyPort, uri))
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, syscall.EPIPE) {
			error404(w, true)
		}
		return
	}

	if resp.StatusCode != 200 {
		// esbuild errors
		if resp.StatusCode == http.StatusServiceUnavailable && strings.HasPrefix(resp.Header.Get("Content-Type"), "text/plain") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(fmt.Sprintf("<!doctype html><meta charset=utf-8><title>error</title><script>%s</script><body><pre>", reloadJS)))
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				_, _ = w.Write([]byte("Error: cannot build app"))
			}
			_ = resp.Body.Close()
			return
		}
		error404(w, true)
		return
	}

	// copy esbuild headers
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
	w.Header().Set("Access-Control-Allow-Origin", resp.Header.Get("Access-Control-Allow-Origin"))
	w.Header().Set("Date", resp.Header.Get("Date"))

	// copy esbuild response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, syscall.EPIPE) {
			error404(w, false)
		}
		return
	}

	// and close
	_ = resp.Body.Close()
}

func error404(res http.ResponseWriter, writeHeader bool) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if writeHeader {
		res.WriteHeader(http.StatusNotFound)
	}
	_, _ = res.Write([]byte("404 - Not Found"))
}
