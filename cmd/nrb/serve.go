package main

import (
	"fmt"
	"net"
	"net/http"

	"github.com/natrim/nrb/lib"
)

func serve() error {
	SetupWebServer()

	fileServer := lib.WrappedFileServer(config.OutputDir)
	http.Handle("/", fileServer)

	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
	if err != nil {
		return err
	}

	// get real port in case user used 0 for random port
	config.Port = socket.Addr().(*net.TCPAddr).Port

	protocol := "http://"
	if isSecured {
		protocol = "https://"
	}
	lib.PrintInfof("Listening on: %s%s:%d\n", protocol, config.Host, config.Port)

	if isSecured {
		return http.ServeTLS(socket, nil, certFile, keyFile)
	}

	return http.Serve(socket, nil)
}
