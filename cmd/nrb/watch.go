package main

import (
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
	"time"
)

var proxyPort uint16 = 8032

var server api.ServeResult
var protocol string
var broker *lib.Broker

func startServer() {
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

	server.Wait()
}

func watch() {
	go startServer()
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

	if err := filepath.WalkDir(lib.RealQuickPath(sourceDir), watchDir(watcher)); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	done := make(chan bool)

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
				lastEvent = event
				timer.Reset(time.Millisecond * 100)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("ERROR", err)
			case <-timer.C:
				if lastEvent.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Printf("EVENT! %s\n", lastEvent.String())
					fmt.Printf("↻ Updated file %s\n", filepath.Base(lastEvent.Name))
					broker.Notifier <- []byte("update")
				} else if lastEvent.Op&fsnotify.Create == fsnotify.Create {
					//fmt.Printf("EVENT! %s\n", lastEvent.String())
					fmt.Printf("↻ Created file %s\n", filepath.Base(lastEvent.Name))
					broker.Notifier <- []byte("update")
				} else if lastEvent.Op&fsnotify.Remove == fsnotify.Remove {
					//fmt.Printf("EVENT! %s\n", lastEvent.String())
					fmt.Printf("↻ Removed file %s\n", filepath.Base(lastEvent.Name))
					broker.Notifier <- []byte("update")
				} else if lastEvent.Op&fsnotify.Rename == fsnotify.Rename {
					//fmt.Printf("EVENT! %s\n", lastEvent.String())
					fmt.Printf("↻ Renamed file %s\n", filepath.Base(lastEvent.Name))
					broker.Notifier <- []byte("update")
				}
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
		os.Exit(1)
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
		os.Exit(1)
	}

	// and close
	err = resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
