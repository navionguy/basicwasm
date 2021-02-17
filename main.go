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

	rt := setup()
	log.Fatal(http.ListenAndServe(*listen, rt))
	//log.Fatal(http.ListenAndServe(*listen, http.FileServer(fileserv.WrapFileOrg())))
}

func setup() *mux.Router {

	flag.Parse()
	log.Printf("listening on %q...", *listen)

	r := mux.NewRouter()

	// setup my routes

	r.HandleFunc("/assets/{type}/{file}", fileserv.FileServ)
	r.HandleFunc("/webmodules/{file}", fileserv.FileServ)
	r.HandleFunc("/", BasicWasm).Name(rootRt)

	return r
}

// BasicWasm serves up the main page
func BasicWasm(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "./assets/html/gwbasic.html")
}
