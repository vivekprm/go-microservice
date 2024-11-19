package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Product struct {
	ID         int
	Name       string
	USDPerUnit float64
	Unit       string
}

func main() {
	r := chi.NewRouter()
	r.Get("/products", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(products)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	})
	// /products/3
	r.Get("/products/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		idRaw := chi.URLParam(r, "id")
		if len(idRaw) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
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

	http.Handle("/", r)

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
