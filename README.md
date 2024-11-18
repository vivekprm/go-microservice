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

# Creating HTTP Servers
- Creating Servers
- Handling Requests
- Interacting with HTTP Protocol
  - Headers
  - Cookies
  - Status Codes

## Creating Default Server

## Creating Secure Server with TLS
Using below two methods weget access to default server: 
https://pkg.go.dev/net/http@go1.23.3#ListenAndServe
https://pkg.go.dev/net/http@go1.23.3#ListenAndServeTLS

### Using ListenAndServe
This is what we did above.
```go
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```
## Creating Custom Servers
Below is the signature for ListenAndServeTLS
```go
func ListenAndServeTLS(addr, certFile, keyFile string, handler Handler) error
```

### Creating self-signed certificate
We are going to run a small go program that comes with standard library.

```sh
git clone https://github.com/golang/go golang.org/go
go run $GOPATH/src/golang.org/go/src/crypto/tls/generate_cert.go --host localhost
```

Now create server as below:
```go
package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Customer service!\n")
	}

	http.HandleFunc("/", helloHandler)
	log.Fatal(http.ListenAndServeTLS(":3000", "./cert.pem", "./key.pem", nil))
}
```