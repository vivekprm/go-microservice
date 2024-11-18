# Building Microservices in Go
One package that we are going to use that contains 80% functionality that we need for building microservices.
- https://pkg.go.dev/std
- https://pkg.go.dev/net@go1.23.3
- https://pkg.go.dev/net/http@go1.23.3

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
Using below two methods we get access to default server: 
- https://pkg.go.dev/net/http@go1.23.3#ListenAndServe
- https://pkg.go.dev/net/http@go1.23.3#ListenAndServeTLS

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

There are some disadvantages of using that default server and that surrounds the methods that we have available on the Server object when we have access to that instance directly, e.g. [Close](https://pkg.go.dev/net/http@go1.23.3#Server.Close), [Shutdown](https://pkg.go.dev/net/http@go1.23.3#Server.Shutdown) & [RegisterOnShutdown](https://pkg.go.dev/net/http@go1.23.3#Server.RegisterOnShutdown) methods, there are couple of others but these are the three that we really want to key in on.

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

# Handling Requests
Go provides two mechanisms for that:

## Using Simple Function
Register a simple function to handle the request itself and we do that using **HandleFunc** function from http package.
```go
http.HandleFunc(pattern string, func(w http.ResponseWriter, req *http.Request){})
```
ResponseWriter is an interface with three methods on it:
```go
type ResponseWriter interface {
    Header() Header                 // used to interact with response header 
    Write([]byte) (int, error)      // implements io.Writer interface
    WriteHeader(statusCode int)      
}
```

Request object is a pointer to an actual struct, so we are going to receive an actual pointer to an object here.
https://pkg.go.dev/net/http@go1.23.3#Request

So anything you want to find about the client, it's buried somewhere in the Request object.

```go
type Request struct {
    // a lot of stuff here! We'll see this later.
}
```

**HandleFunc** takes a function ```func(w http.ResponseWriter, req *http.Request){}``` that actually have a special type in Go, called a [HandlerFunc](https://pkg.go.dev/net/http@go1.23.3#HandlerFunc)

```go
type HandlerFunc func(ResponseWriter, *Request)
```

So it's a type of a function that takes **ResponseWriter** and pointer to **Request** object. So you can pass a function and Go will implicitly convert it to a **HandlerFunc**, it knows how to do that.

But if you are creating standalone functions, it's important to recognize Go can't just take a standard function, it does need it to be typed as **HandlerFunc**. 

So let's create a custom function that just prints the URI that request came from:

```go
var handlerFunc http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, r.URL.String())
}
http.HandleFunc("/url/", handlerFunc)
```

If we hit ```http://localhost:3000/url``` or ```http://localhost:3000/url/foo/bar/baz``` same request handler gets called. The reason for that is Go's pattern matching. First parameter is pattern. The pattern that we are matching is a best fit wins.

## Using Custom Types
The second mechanism that Go offers us to register Request Handlers is using a a [Handle](https://pkg.go.dev/net/http@go1.23.3#Handle) function. 

Notice that first parameter is same for both **Handle** and **HandleFunc**, but the second parameter is significantly different. With **HandleFunc** we pass a function, with **Handle** we actually have to pass in an instance of a **[Handler](https://pkg.go.dev/net/http@go1.23.3#Handler)**. So what's a handler?  

We see that it's an interface with one method defined on it that takes ResponseWriter & Request as parameter.
```go
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

So Handle function allows us to create any type we want as long as we have ServeHTTP method on it. So we can setup much more complicated object and can get much more granular over how we handle a request relative to the simple **HandleFunc**, which just accepts a single function.

```go
func main() {
    http.Handle("/", myHandler("Customer service"))
    ....
}
type myHandler string

func (mh myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, string(mh))
}
```

It's simple yet powerful concept, with custom handler in hand, we can bind as many method as we want to it and if we use a more complicated type such as a struct, I can store information in the handler, so I can store state information or whatever I need, in order to support much more complicated workflows than would normally be handled by a simple **HanderFunc**.