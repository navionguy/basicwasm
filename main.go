package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/navionguy/basicwasm/fileserv"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
)

const (
	rootRt = "root"
)

func main() {

	rt := startup()
	log.Fatal(http.ListenAndServe(*listen, rt))
}

func startup() *mux.Router {

	flag.Parse()
	log.Printf("listening on %q...", *listen)

	r := mux.NewRouter()

	// setup my routes

	fileserv.WrapFileSources(r)
	r.HandleFunc("/", gwbasicHTML).Name("main page")

	return r
}

// gwbasicHTML serves up the main page
func gwbasicHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./assets/html/gwbasic.html")
}
