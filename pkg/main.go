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
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/customers/add", func(w http.ResponseWriter, r *http.Request) {
		var c Customer
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&c)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			return // important to exit out of handler
		}
		log.Println(c)
	})

	// Sending request
	go func() {
		time.Sleep(2 * time.Second)
		_, err := http.Post("http://localhost:3000/customers/add", "application/json", bytes.NewBuffer([]byte(`
			{
				"id": 999,
				"firstName": "Arthur",
				"lastName": "Dent",
				"address": "155 Country Lane, Cottington, England"
			}
		`)))
		if err != nil {
			log.Fatal(err)
		}
	}()

	http.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		customers, err := readCustomers()
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return // important to exit out of handler
		}
		data, err := json.Marshal(customers)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return // important to exit out of handler
		}
		w.Header().Add("content-type", "application/json")
		w.Write(data)
	})
	s := http.Server{
		Addr: ":3000",
	}
	go func() {
		log.Fatal(s.ListenAndServe())
	}()
	fmt.Println("Server started, press <ENTER> to shutdown")
	fmt.Scanln()
	s.Shutdown(context.Background())
	fmt.Println("Server stopped")
}

type Customer struct {
	ID        int    `json="id"`
	Firstname string `json="firstName"`
	Lastname  string `json="lastName"`
	Address   string `json="address"`
}

func readCustomers() ([]Customer, error) {
	f, err := os.Open("customers.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

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
