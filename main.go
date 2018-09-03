// neilpa.me: start of the new site
//
// Started from a minimal Let's Encrypt server on StackOverflow.
// <https://stackoverflow.com/a/40494806/1999152>
package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// host is the root name of the site.
const host = "neilpa.me"

// secure requires certificates.
const secure bool = true

// local for non-production and dev builds.
var local = false

// version identifies the running instance.
var version string = "?"

// main is the entry point to the app.
func main() {
	log.SetPrefix(host + ": ")

	flag.BoolVar(&local, "local", false, "run local dev")
	flag.Parse()

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
		log.Printf("access: %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)

		body, code := handleRequest(r)
		code, size := renderRequest(w, body, code)

		log.Printf("result: %s %s %s %d %d", r.RemoteAddr, r.Method, r.URL.Path, code, size)
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
	log.Printf("run: listening at %s:80 (http redirect)\n", host)
	go http.ListenAndServe(":http", api.certs.HTTPHandler(nil))

	// Key and cert are coming from Let's Encrypt
	log.Printf("run: listening at %s:443\n", host)
	log.Fatal(api.server.ListenAndServeTLS("", ""))
}

// handleRequest uses the method and path to figure out what the client wants.
func handleRequest(r *http.Request) (body string, code int) {
	code = 200

	if r.Method != "GET" {
		code = 405 // Method not allowed
	}

	switch r.URL.Path {
	case "/":
		body = "Hello world!"
	case "/health":
		body = "200 OK"
	case "/search":
		body = "TODO"
	case "/status":
		body = strings.Join([]string{
			"version: " + version,
			"...",
		}, "\n")
	case "/version":
		body = version
		// todo: site versioning vs content versioning
	default:
		body, code = handleStatic(r)
	}

	return
}

// handleStatic looks up files, optionally without an extension via globbing.
//
// TODO Do smarter mapping of data
func handleStatic(r *http.Request) (body string, code int) {
	code = 500

	pattern := filepath.Join("static", r.URL.Path) + "*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("glob: %s: %s\n", pattern, err)
		body = "unexpected error"
		return
	} else if len(matches) == 0 {
		body, code = "not found", 404
		return
	}

	f, err := os.Open(matches[0])
	if err != nil {
		log.Println("open: ", err)
		body = "unexpected error"
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("read: %s: %s\n", f.Name(), err)
		body = "unexpected error"
		return
	}

	body, code = string(b), 200
	return
}

// renderRequest writes the body and response code in a best effort approach to
// what the the client accepts. Under certian conditions (e.g. no content) the
// code may be tweaked. The actual response code and size of the payload are
// returned.
func renderRequest(w http.ResponseWriter, body string, code int) (int, int) {
	if code >= 400 {
		if body == "" {
			body = http.StatusText(code)
		}
		http.Error(w, body, code)
	} else {
		if code == 200 && body == "" {
			code = 204
		}
		w.WriteHeader(code)
		w.Write([]byte(body))
		// TODO w.Write([]byte("\n"))
	}
	return code, len(body)
}
