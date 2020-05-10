package main

import (
	"io"
	"log"
	"net/http"
)

const (
	listenAddress = "0.0.0.0:4778"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "index")
	})
	log.Printf("Listening for connections at %s", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
