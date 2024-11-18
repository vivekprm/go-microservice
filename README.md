# Building Microservices in Go
One package that we are going to use that contains 80% functionality that we need for building microservices.
https://pkg.go.dev/std
https://pkg.go.dev/net@go1.23.3
https://pkg.go.dev/net/http@go1.23.3

Two steps:
- Register our handlers.
  - There are two ways to register handlers.
    - Simple: Using HandleFunc
    ```go
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world!")
	})
    ```
- Start the server
    ```go
    // By providing nil, we tell Go to use standard handler
	log.Fatal(http.ListenAndServe(":3000", nil))
    ```