package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "microservices-grpc-go-python/catalog/ecommerce"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func getDiscountConnection(host string) (*grpc.ClientConn, error) {
	wd, _ := os.Getwd()
	parentDir := filepath.Dir(wd)
	certFile := filepath.Join(parentDir, "keys", "cert.pem")
	creds, _ := credentials.NewClientTLSFromFile(certFile, "")
	return grpc.Dial(host, grpc.WithTransportCredentials(creds))
}

func getFakeProducts() []*pb.Product {
	p1 := pb.Product{Id: 1, Slug: "iphone-x", Description: "64GB, black and iOS 12", PriceInCents: 99999}
	p2 := pb.Product{Id: 2, Slug: "notebook-avell-g1511", Description: "Notebook Gamer Intel Core i7", PriceInCents: 150000}
	p3 := pb.Product{Id: 3, Slug: "playstation-4-slim", Description: "1TB Console", PriceInCents: 32999}
	return []*pb.Product{&p1, &p2, &p3}
}

func getCustomerByID(id int) (pb.Customer, error) {
	c1 := pb.Customer{Id: 1, FirstName: "Joe", LastName: "Doe"}
	c2 := pb.Customer{Id: 2, FirstName: "Miha", LastName: "Sedej"}

	customers := map[int]pb.Customer{
		1: c1,
		2: c2,
	}

	customer, ok := customers[id]

	if ok {
		return customer, nil
	}

	return customer, errors.New("customer not found")
}

func handleError(err error) map[string][]string {
	return map[string][]string{"errors": []string{err.Error()}}
}

func getProductsWithDiscountApplied(customer pb.Customer, products []*pb.Product) []*pb.Product {
	host := os.Getenv("DISCOUNT_SERVICE_HOST")
	if len(host) == 0 {
		host = "localhost:443"
	}

	conn, err := getDiscountConnection(host)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewDiscountClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	productsWithDiscountApplied := make([]*pb.Product, 0)
	for _, product := range products {
		r, err := c.ApplyDiscount(ctx, &pb.DiscountRequest{Customer: &customer, Product: product})
		if err == nil {
			productsWithDiscountApplied = append(productsWithDiscountApplied, r.GetProduct())
		} else {
			log.Println("failed to apply discount", err)
		}
	}

	if len(productsWithDiscountApplied) > 0 {
		return productsWithDiscountApplied
	}

	return products
}

func handleGetProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Get customer
	customerIDStr := params["customerID"]
	customerID, _ := strconv.Atoi(customerIDStr)
	customer, err := getCustomerByID(customerID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(handleError(err))
		return
	}

	// Get products
	products := getFakeProducts()

	productsWithDiscountApplied := getProductsWithDiscountApplied(customer, products)
	json.NewEncoder(w).Encode(productsWithDiscountApplied)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	port := "11080"

	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	r := mux.NewRouter()
	r.HandleFunc("/products/{customerID:[0-9]+}", handleGetProducts).Methods("GET")
	http.Handle("/", r)

	fmt.Println("Server running on", port)
	http.ListenAndServe(":"+port, nil)
}
