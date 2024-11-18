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

### Creating Secure Server with TLS
Using below two methods weget access to default server: 
https://pkg.go.dev/net/http@go1.23.3#ListenAndServe
https://pkg.go.dev/net/http@go1.23.3#ListenAndServeTLS

#### Using ListenAndServe
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

#### Creating self-signed certificate
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

## Creating Custom Server
It's important to understand where it's all located. Look at below server type that we can create:

https://pkg.go.dev/net/http@go1.23.3#Server

The Default Server we have been using is preconfigured instance of s Server object that's exactly the same type we can create. 

There are some disadvantages of using that default server and that surrounds the methods that we have available on the Server object when we have access to that instance directly, e.g. (Close)[https://pkg.go.dev/net/http@go1.23.3#Server.Close], (Shutdown)[https://pkg.go.dev/net/http@go1.23.3#Server.Shutdown] & (RegisterOnShutdown)[https://pkg.go.dev/net/http@go1.23.3#Server.RegisterOnShutdown] methods, there are couple of others but these are the three that we really want to key in on.

**Close** method doesn't allow any inflight request to finish and we get 500 error for inflight request. However **Shutdown** method is more graceful and allows in-flight requests to complete.

Here in ListenAndServe we don't have to pass addr string as its a field in Server type.

Our handler is going to work exactly the same, because our handler is actually registering requests with that **DefaultServeMux**, which we will look at later. So as long as we pass nil in ListenAndServe, default server mux will be used and we will be fine.

http.ListenAndServer() method used earlier returns error but the reason we never get that error is both ListenAndServer and ListenAndServeTLS both block as long as server is running. So only time it stops when server actually closes out.

So since we want to do somthing else with this server object we will wrap it in a goroutine and create server object as below:

```go
s := http.Server{
		Addr: ":3000",
	}
	go func() {
		log.Fatal(s.ListenAndServeTLS("../cert.pem", "../key.pem"))
	}()
	fmt.Println("Server started, press <ENTER> to shutdown")
	fmt.Scanln()
	s.Shutdown(context.Background())
    fmt.Println("Server stopped")
```

Reson we need to pass context to shutdown is because if it takes too long and context gets cancelled, you can cancel the shutdown and immediately go to close. But we don't actually need that so we can have Background context here.