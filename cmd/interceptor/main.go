package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/bigkevmcd/slack-webhook-interceptor/pkg/interception"
)

var (
	port = flag.Int("port", 8080, "port to listen on")
)

func main() {
	flag.Parse()

	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in cluster config: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		log.Fatalf("Failed to get the Kubernetes client set: %v", err)
	}

	http.Handle("/", interception.NewHandler(kubeClient))
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
