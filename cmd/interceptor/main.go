package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/bigkevmcd/slack-webhook-interceptor/pkg/interception"
)

var (
	port = flag.Int("port", 8080, "port to listen on")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", interception.Handler)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
