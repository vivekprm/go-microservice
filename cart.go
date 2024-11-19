package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type Cart struct {
	ID         int   `json:"id,omitempty"`
	CustomerID int   `json:"customerId, omitempty`
	ProductIDs []int `json:productIds,omitempty`
}

var nextID int = 1
var carts = make([]Cart, 0)
var cartMux = http.NewServeMux()

func createShoppingCartService() *http.Server {
	cartMux.HandleFunc("/carts", cartsHandler)

	s := http.Server{
		Addr:    ":4040",
		Handler: &loggingMiddleware{next: cartMux},
	}
	return &s
}

type loggingMiddleware struct {
	next http.Handler
}

func (lm *loggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if lm.next == nil {
		lm.next = cartMux
	}
	slog.Info(fmt.Sprintf("Received %v request on route: %v", r.Method, r.URL.Path))
	now := time.Now()
	lm.next.ServeHTTP(w, r)
	slog.Info(fmt.Sprintf("Response generated for %v on route %v. Duration: %v", r.Method, r.URL.Path, time.Since(now)))
}

func cartsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		data, err := json.Marshal(carts)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	case http.MethodPost:
		var c Cart
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&c)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c.ID = nextID
		nextID++
		carts = append(carts, c)
		w.WriteHeader(http.StatusCreated)
		data, err := json.Marshal(c)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(data)
	}
}
