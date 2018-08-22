// neilpa.me: start of the new site
//
// Started from a minimal Let's Encrypt server on StackOverflow.
// <https://stackoverflow.com/a/40494806/1999152>
package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// host is the root name of the site.
const host = "neilpa.me"

// secure requires certificates.
var secure bool

// local for non-production and dev builds.
const local = false

// version identifies the running instance.
var version string = "?"

// main is the entry point to the app.
func main() {
	log.SetPrefix(host + ": ")

	// TODO Local experiment
	if local || !secure {
		NewAPI(nil).run()
	}

	certs := NewCerts("neilpa.me")
	NewAPI(certs).run()
}

// Certs wraps the certificate manager.
type Certs struct {
	autocert.Manager
	// host is the name the certs validate.
	host string
}

// NewCerts creates the default manager
func NewCerts(host string) *Certs {
	return &Certs{autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(host),
		Cache:      autocert.DirCache("certs"),
	}, host}
}

// API binds the router to the server.
type API struct {
	// certs is the certificate manager.
	certs *Certs
	// server the API is bound to.
	server *http.Server
}

// NewAPI creates the default API backed by the provided certs.
func NewAPI(certs *Certs) *API {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access: %s %s", r.Method, r.URL.Path)
		var code = 200
		var body string

		if r.Method != "GET" {
			code = 405 // Method not allowed
		}
		switch r.URL.Path {
		case "/":
			body = "Hello world!"
		case "/health":
			body = "200 OK"
		case "/status":
			body = strings.Join([]string{
				"version: " + version,
				"...",
			}, "\n")
		case "/version":
			body = version
		default:
			code = 404
		}

		if code >= 400 {
			if body == "" {
				body = http.StatusText(code)
			}
			http.Error(w, body, code)
		} else {
			if body == "" {
				code = 204
			}
			w.WriteHeader(code)
			w.Write([]byte(body))
			w.Write([]byte("\n"))
		}

		log.Printf("result: %s %s %d %d", r.Method, r.URL.Path, code, len(body))
	})

	if certs == nil {
		log.Println("warn: self-signed certs")
		return &API{nil, nil}
	}
	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certs.GetCertificate,
		},
	}
	return &API{certs, server}
}

// run starts the api.
func (api *API) run() {
	if api.certs == nil {
		log.Println("run: insecure: missing certs and/or server")
		log.Println("run: listening at *:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}

	if api.server == nil {
		log.Fatal("run: server required for https")
	}

	// Needed to make http available because of letsencrypt security issue
	// https://community.letsencrypt.org/t/important-what-you-need-to-know-about-tls-sni-validation-issues/50811
	// https://github.com/golang/go/issues/21890
	//
	// As a bonus of this fix we now have http-->https redirect.
	go http.ListenAndServe(":http", api.certs.HTTPHandler(nil))

	// Key and cert are coming from Let's Encrypt
	log.Fatal(api.server.ListenAndServeTLS("", ""))
}
