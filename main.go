// neilpa.me: start of the new site
//
// https://stackoverflow.com/a/40494806/1999152
package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	mgr := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("neilpa.me"),
		Cache:      autocert.DirCache("certs"),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: mgr.GetCertificate,
		},
	}

	// Needed to make http available because of letsencrypt security issue
	// https://community.letsencrypt.org/t/important-what-you-need-to-know-about-tls-sni-validation-issues/50811
	// https://github.com/golang/go/issues/21890
	//
	// As a bonus of this fix we now have http-->https redirect.
	go http.ListenAndServe(":http", mgr.HTTPHandler(nil))

	// Key and cert are coming from Let's Encrypt
	log.Fatal(server.ListenAndServeTLS("", ""))
}
