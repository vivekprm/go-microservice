package main

import (
	"fmt"
	"log"

	"github.com/vivekprm/go-microservices/grpc-demo/productpb"
	"google.golang.org/protobuf/proto"
)

type Product struct {
	ID         int
	Name       string
	USDPerUnit float64
	Unit       string
}

func main() {
	p := productpb.Product{
		Id:         int32(products[0].ID),
		Name:       products[0].Name,
		UsdPerUnit: products[0].USDPerUnit,
		Unit:       products[0].Unit,
	}
	data, err := proto.Marshal(&p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	var p2 productpb.Product
	err = proto.Unmarshal(data, &p2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", p2)
	// http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
	// 	data, err := json.Marshal(products)
	// 	if err != nil {
	// 		log.Println(err)
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}
	// 	w.Header().Add("Content-Type", "application/json")
	// 	w.Write(data)
	// })
	// // /products/3
	// pattern := regexp.MustCompile(`^\/products\/(\d+?)$`)
	// http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
	// 	matches := pattern.FindStringSubmatch(r.URL.Path) // ["/products/3", "3"]
	// 	if len(matches) == 0 {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	id, err := strconv.Atoi(matches[1])
	// 	if err != nil {
	// 		log.Println(err)
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}

	// 	for _, p := range products {
	// 		if p.ID == id {
	// 			data, err := json.Marshal(p)
	// 			if err != nil {
	// 				log.Println(err)
	// 				w.WriteHeader(http.StatusInternalServerError)
	// 				return
	// 			}
	// 			w.Header().Add("Content-Type", "application/json")
	// 			w.Write(data)
	// 			return
	// 		}
	// 	}
	// 	w.WriteHeader(http.StatusNotFound)
	// })
	// s := http.Server{
	// 	Addr: ":4000",
	// }
	// go func() {
	// 	log.Fatal(s.ListenAndServe())
	// }()
	// fmt.Println("Server started. Press <ENTER> to shutdown.")
	// fmt.Scanln()
	// s.Shutdown(context.Background())
	// fmt.Println("Server stopped")
}
