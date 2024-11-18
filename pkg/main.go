package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// http://localhost:3000/files/customers.csv
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("."))))

	http.HandleFunc("/servecontent", func(w http.ResponseWriter, r *http.Request) {
		customerFile, err := os.Open("./customers.csv")
		if err != nil {
			log.Fatal(err)
		}
		defer customerFile.Close()
		http.ServeContent(w, r, "customerdata.csv", time.Now(), customerFile)
	})
	http.HandleFunc("/servefile", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./customers.csv")
	})
	http.HandleFunc("/fprint", func(w http.ResponseWriter, r *http.Request) {
		customerFile, err := os.Open("./customers.csv")
		if err != nil {
			log.Fatal(err)
		}
		defer customerFile.Close()

		// data, err := io.ReadAll(customerFile)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Fprint(w, string(data))
		io.Copy(w, customerFile)
	})
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
