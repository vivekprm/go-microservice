package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/vivekprm/go-microservices/grpc-demo/productpb"
	"google.golang.org/grpc"
)

type Product struct {
	ID         int
	Name       string
	USDPerUnit float64
	Unit       string
}

func main() {
	go startGRPCServer()
	time.Sleep(1 * time.Second)
	callGRPCService()
	/*
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
	*/
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

type ProductService struct {
	productpb.UnimplementedProductServer
}

func (ps ProductService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.GetProductReply, error) {
	for _, p := range products {
		if p.ID == int(req.ProductId) {
			return &productpb.GetProductReply{
				Product: &productpb.Product{
					Id:         int32(p.ID),
					Name:       p.Name,
					UsdPerUnit: p.USDPerUnit,
					Unit:       p.Unit,
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("product not found with ID: %v", req.ProductId)
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", "localhost:4001")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	productpb.RegisterProductServer(grpcServer, &ProductService{})
	log.Fatal(grpcServer.Serve(lis))
}

func callGRPCService() {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("localhost:4001", opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := productpb.NewProductClient(conn)
	res, err := client.GetProduct(context.Background(), &productpb.GetProductRequest{ProductId: 3})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("%+v", res.Product)
}
