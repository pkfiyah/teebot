package api

import (
	"fmt"
	"log"
	"net/http"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Health Status: OK")
	fmt.Println("[Endpoint Hit]: health")
}

func HandleRequests() {
	http.HandleFunc("/health", health)
	log.Fatal(http.ListenAndServe(":9001", nil))
}
