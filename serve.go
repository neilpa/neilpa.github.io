package main

import (
    "time"
    "os"
    "log"
    "net/http"

	"golang.org/x/crypto/acme/autocert"
)

const host = "neilpa.me"

func main() {
	log.SetPrefix(host + ": ")

    prod := len(os.Args) == 2 && os.Args[1] == "prod"
    h := &handler{http.FileServer(http.Dir("."))}

    if prod {
        mgr := autocert.Manager {
            Prompt: autocert.AcceptTOS,
            HostPolicy: autocert.HostWhitelist(host),
            Cache: autocert.DirCache("certs"),
        }

        // Needed to make http available because of letsencrypt security issue
        // https://community.letsencrypt.org/t/important-what-you-need-to-know-about-tls-sni-validation-issues/50811
        // https://github.com/golang/go/issues/21890
        //
        // As a bonus of this fix we now have http-->https redirect.
        log.Printf("listening at %s:80 (http redirect)\n", host)
        go http.ListenAndServe(":http", mgr.HTTPHandler(nil))

        log.Printf("listening at %s:443\n", host)
        log.Fatal(http.Serve(mgr.Listener(), h))
    } else {
        log.Printf("listening at %s:8080\n", host)
        log.Fatal(http.ListenAndServe(":8080", h))
    }
}

type handler struct {
    fs http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    log.Printf("access: %s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
    rec := recorder{w, 200, 0}
    start := time.Now()
    h.fs.ServeHTTP(&rec, r)
    elapsed := time.Now().Sub(start)
	log.Printf("result: %s %s %s %d %d %d", r.RemoteAddr, r.Method, r.URL.Path, rec.status, rec.size, elapsed)
}

type recorder struct {
    http.ResponseWriter
    status int
    size int
}

func (r *recorder) WriteHeader(code int) {
    r.status = code
    r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(buf []byte) (int, error) {
    r.size += len(buf)
    return r.ResponseWriter.Write(buf)
}

