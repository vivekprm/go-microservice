package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Product struct {
	ID         int
	Name       string
	USDPerUnit float64
	Unit       string
}

func main() {
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(products)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	})
	http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		idRaw := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for _, p := range products {
			if p.ID == id {
				data, err := json.Marshal(p)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Add("Content-Type", "application/json")
				w.Write(data)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	s := http.Server{
		Addr: ":4000",
	}

	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	fmt.Println("Server started, press <ENTER> to shutdown")
	fmt.Scanln()
	s.Shutdown(context.Background())
	fmt.Println("Server stopped")
}
