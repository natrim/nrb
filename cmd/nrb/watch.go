package main

import (
	"errors"
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
	"github.com/natrim/nrb/lib"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var proxyPort uint16 = 8032

var server api.ServeResult
var protocol string
var broker *lib.Broker

func startServer(done chan bool) {
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

	done <- true
	server.Wait()
}

func watch() {
	done := make(chan bool)

	go startServer(done)

	<-done

	var err error

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer watcher.Close()

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

		http.HandleFunc("/", pipeRequest)
		http.Handle("/esbuild", broker)
		fmt.Printf("> Listening on: %s%s:%d\n", protocol, host, wwwPort)

		if isSecured {
			err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, wwwPort), certFile, keyFile, nil)
		} else {
			err = http.ListenAndServe(fmt.Sprintf("%s:%d", host, wwwPort), nil)
		}

		if err != nil {
			fmt.Println(err)
			watcher.Close()
			server.Stop()
			os.Exit(1)
		}
	}()

	<-done
	server.Stop()
	watcher.Close()

	//TODO: add serve proxy for hot reloading
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

func pipeRequest(w http.ResponseWriter, r *http.Request) {
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
	err = resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		// error404(w, false)
	}
}

func error404(res http.ResponseWriter, writeHeader bool) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if writeHeader {
		res.WriteHeader(http.StatusNotFound)
	}
	res.Write([]byte("404 - Not Found"))
}
