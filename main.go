package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/navionguy/basicwasm/fileserv"
)

var (
	listen = flag.String("listen", ":8080", "listen address")
)

func main() {

	flag.Parse()
	log.Printf("listening on %q...", *listen)
	//log.Fatal(http.ListenAndServe(*listen, http.FileServer(http.Dir(*dir))))
	log.Fatal(http.ListenAndServe(*listen, http.FileServer(fileserv.WrapFileOrg())))
}
