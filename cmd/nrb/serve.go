package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/natrim/nrb/lib"
)

func serve() int {
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
		if lib.IsErrorAddressAlreadyInUse(err) {
			socket, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, 0))
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return 1
			}
			_, _ = fmt.Fprintln(os.Stderr, ERR, "port", port, "is in use")
			port = socket.Addr().(*net.TCPAddr).Port
		} else {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}

	fmt.Printf(INFO+" Listening on: %s%s:%d\n", protocol, host, port)

	if isSecured {
		err = http.ServeTLS(socket, nil, certFile, keyFile)
	} else {
		err = http.Serve(socket, nil)
	}

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
