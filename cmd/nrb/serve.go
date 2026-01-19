package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/natrim/nrb/lib"
)

func serve() error {
	SetupWebServer()
	fileServer := lib.WrappedFileServer(outputDir)
	http.Handle("/", fileServer)
	var protocol string
	if isSecured {
		protocol = "https://"
	} else {
		protocol = "http://"
	}
	var err error
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	// get real port in case user used 0 for random port
	port = socket.Addr().(*net.TCPAddr).Port

	lib.PrintInfof("Listening on: %s%s:%d\n", protocol, host, port)

	if isSecured {
		return http.ServeTLS(socket, nil, certFile, keyFile)
	} else {
		err = http.Serve(socket, nil)
	}

	return err
}
