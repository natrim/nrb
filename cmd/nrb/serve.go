package main

import (
	"fmt"
	"github.com/natrim/nrb/lib"
	"net/http"
	"os"
)

func serve() {
	fileServer := lib.WrappedFileServer(outputDir)
	http.Handle("/", fileServer)
	var protocol string
	if isSecured {
		protocol = "https://"
	} else {
		protocol = "http://"
	}
	fmt.Printf(INFO+" Listening on: %s%s:%d\n", protocol, host, wwwPort)
	var err error
	if isSecured {
		err = http.ListenAndServeTLS(fmt.Sprintf("%s:%d", host, wwwPort), certFile, keyFile, nil)
	} else {
		err = http.ListenAndServe(fmt.Sprintf("%s:%d", host, wwwPort), nil)
	}

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
