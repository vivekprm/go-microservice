package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world!")
	})
	// By providing nil, we tell Go to use standard handler
	log.Fatal(http.ListenAndServe(":3000", nil))
}
