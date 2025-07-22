package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		logRequestDetail(r)
		fmt.Fprint(w, "Handling incomming requests from orders\n")
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		logRequestDetail(r)
		fmt.Fprintf(w, "Handling incomming requests from users\n")
	})

	port := 3000

	// Load the TLS cert and key
	cert := "cert.pem"
	key := "key.pem"

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Create custom server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   nil,
		TLSConfig: tlsConfig,
	}

	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Couldn't start server", err)
	}

	// http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func logRequestDetail(r *http.Request) {
	httpVersion := r.Proto
	fmt.Println("The version of the current http request is: ", httpVersion)

	if r.TLS != nil {
		tlsVersion := getTlsVersion(r.TLS.Version)
		fmt.Println("received request with TLS version", tlsVersion)
	} else {
		fmt.Println("Received request without TLS")
	}
}

func getTlsVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown TLS Version"
	}
}
