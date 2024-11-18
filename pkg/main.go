package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	var handlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, r.URL.String())
	}
	http.HandleFunc("/url/", handlerFunc)
	
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Customer service!\n")
	}

	http.HandleFunc("/", helloHandler)
	s := http.Server{
		Addr: ":3000",
	}
	go func() {
		log.Fatal(s.ListenAndServeTLS("../cert.pem", "../key.pem"))
	}()
	fmt.Println("Server started, press <ENTER> to shutdown")
	fmt.Scanln()
	s.Shutdown(context.Background())
	fmt.Println("Server stopped")
}
