package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

func main() {
	cs := createCustomerService()
	ps := createProductService()
	scs := createShoppingCartService()

	go func() {
		cs.ListenAndServe()
	}()
	go func() {
		ps.ListenAndServe()
	}()
	go func() {
		scs.ListenAndServe()
	}()

	time.Sleep(1 * time.Second)

	http.Post("http://localhost:4040/carts", "application/json", bytes.NewBufferString(`
		{
			"id": 1,
			"customerId": 999,
			"productIds": [1, 3]
		}
	`))

	res, err := http.Get("http://localhost:4040/carts")
	if err != nil {
		log.Println(err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println("Response: ", string(data))

	fmt.Println("Services started, press <ENTER> to shutdown")
	fmt.Scanln()
	cs.Shutdown(context.Background())
	ps.Shutdown(context.Background())
	scs.Shutdown(context.Background())
	fmt.Println("Services stopped")
}

func createCustomerService() *http.Server {
	f, err := os.Open("customers.csv")
	if err != nil {
		log.Fatal(err)
	}
	customers, err := readCustomers(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(customers)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return // important to exit out of handler
		}
		w.Header().Add("content-type", "application/json")
		w.Write(data)
	})
	pattern := regexp.MustCompile(`^\/customers\/(\d+?)$`)
	mux.HandleFunc("/customers/", func(w http.ResponseWriter, r *http.Request) {
		matches := pattern.FindStringSubmatch(r.URL.Path)
		if len(matches) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(matches[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, c := range customers {
			if id == c.ID {
				data, err := json.Marshal(c)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if r.Method == http.MethodGet {
					w.Header().Add("Content-Type", "application/json")
					w.Write(data)
				}
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})
	s := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	return &s
}

func createProductService() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
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
	pattern := regexp.MustCompile(`^\/products\/(\d+?)$`)
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		matches := pattern.FindStringSubmatch(r.URL.Path) // ["/products/3", "3"]
		if len(matches) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id, err := strconv.Atoi(matches[1])
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
	s := http.Server{
		Addr:    ":4000",
		Handler: mux,
	}
	return &s
}

func readCustomers(f *os.File) ([]Customer, error) {
	customers := make([]Customer, 0)
	csvReader := csv.NewReader(f)
	csvReader.Read() // throw away header
	for {
		fields, err := csvReader.Read()
		if err == io.EOF {
			return customers, nil
		}
		if err != nil {
			return nil, err
		}
		id, err := strconv.ParseInt(fields[0], 10, 8)
		if err != nil {
			log.Fatal(err)
		}
		customers = append(customers, Customer{
			ID:        int(id),
			Firstname: fields[1],
			Lastname:  fields[2],
			Address:   fields[3],
		})
	}
}

type Customer struct {
	ID        int    `json="id"`
	Firstname string `json="firstName"`
	Lastname  string `json="lastName"`
	Address   string `json="address"`
}

type Product struct {
	ID         int
	Name       string
	USDPerUnit float64
	Unit       string
}
