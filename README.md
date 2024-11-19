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

# Adding Headers
One thing that is important to understand with headers is we generally need to pass headers before we pass the response body back.

```go
func (mh myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Powered-By", "energetic gophers")
	fmt.Fprintln(w, string(mh))
	fmt.Fprintln(w, r.Header)
}
```

# Handling Http Cookies
To work with cookie we can't work with ResponseWriter as it has no method to work with cookies. In http package there is a SetCookie function which takes ResponseWriter as first parameter. 

```go
func (mh myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session-id",
		Value:   "12345",
		Expires: time.Now().Add(24 * time.Hour * 365),
	})
	w.Header().Add("X-Powered-By", "energetic gophers")
	fmt.Fprintln(w, string(mh))
	fmt.Fprintln(w, r.Header)
}
```

# Setting Status Codes
Http status codes are defined here. https://pkg.go.dev/net/http@go1.23.3#pkg-constants

```go
func (mh myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "session-id",
		Value:   "12345",
		Expires: time.Now().Add(24 * time.Hour * 365),
	})
	w.Header().Add("X-Powered-By", "energetic gophers")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, string(mh))
	fmt.Fprintln(w, r.Header)
}
```

# Serving Static Content
There are below methods to server static content:
- Using ```fmt.Fprint```
- Using ```http.ServeFile```
- Using ```http.ServeContent```
- Using ```http.FileServer```

We are going to customer.csv file here, but instead of csv processing we are going to send this text file directly to the requester.

## Using ```fmt.Fprint```

```go
http.HandleFunc("/fprint", func(w http.ResponseWriter, r *http.Request) {
    customerFile, err := os.Open("./customers.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer customerFile.Close()

    data, err := io.ReadAll(customerFile)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Fprint(w, string(data))
})
```

What's important for us to remember at this point of time is, if we are going to use **Fprint**, we need something that meets the **io.Writer** interface. In this case w variable that we receive in our handler implements **io.Writer** interface.

Now another option that we have that doesn't require type conversion is ```io.Copy``` function, which allows us to copy from a **Reader** to **Writer**.

```go
http.HandleFunc("/fprint", func(w http.ResponseWriter, r *http.Request) {
    customerFile, err := os.Open("./customers.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer customerFile.Close()
    io.Copy(w, customerFile)
})
```

It's directly going to stream the data directly from file to response.

## Using ```http.ServeFile```
Next function that we are going to look at is: https://pkg.go.dev/net/http@go1.23.3#ServeFile

```go
func ServeFile(w ResponseWriter, r *Request, name string)
```

Here last parameter **name** is relative path to the file we want to serve.

```go
http.HandleFunc("/servefile", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./customers.csv")
})
```

You notice in this case that here we do get that file downloaded. So **ServeFile** is going to assume that when we are serving that file, it's going to add appropriate headers for a downloadable file. Now to get some more insight on what just happened, lets open our developer tools and check the headers that's being set. 
E.g. in this case Response Header **Content-Type** is being set as **application/vnd.ms-excel**. So ServeFile automatically detected the content type for us, since chrome doesn't have native reader for that it decided to download it.

## Using ```http.ServeContent```
[http.ServeContent](https://pkg.go.dev/net/http@go1.23.3#ServeContent) looks very similar to ServeFile method.

```go
func ServeContent(w ResponseWriter, req *Request, name string, modtime time.Time, content io.ReadSeeker)
```

**ServeFile** is optimized for serving a file off of your file system. **ServeContent** is more flexible, it's designed to serve anything that meets the [ReadSeeker](https://pkg.go.dev/io#ReadSeeker) interface. In short, ReadSeeker had ```Read``` mwthod on it and ```Seek``` method on it.

Now a common **ReadSeeker** is, in fact, a file but you can use different things. It's more flexible because it's not relying on a specific file sitting on your file system. In other words, it's work like **io.Copy** but is more optimized for HTTP transaction.

Using third parameter name go does content type detection.

```go
http.HandleFunc("/servecontent", func(w http.ResponseWriter, r *http.Request) {
    customerFile, err := os.Open("./customers.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer customerFile.Close()
    http.ServeContent(w, r, "customerdata.csv", time.Now(), customerFile)
})
```

## Using ```http.FileServer```
What are we talking about setting a file server? haven't we done that?
Not quite, **ServeFile** function is designed to serve a single resource. You notice that we actually provide the name of the file we want to serve. What happens if we have got multiple resources that we want to serve. That's what we want to look at next.

Now there are several functions that typically work together to commplish this goal. So lets go ahead and jump over to the HTTP documentation.

https://pkg.go.dev/net/http@go1.23.3#FileServer

First function that we want to look at is Constructor function what we mean by that is we are actually going to be constructing a **Handler**.

So **FileServer** are handler, they have complicated internal API that allows them to accomplish their mission.
```go
func FileServer(root FileSystem) Handler
```

It takes one parameter 'root' which is of [FileSystem](https://pkg.go.dev/net/http@go1.23.3#FileSystem) type. **FileSystem** type is an interface which has just one method **Open**

[Dir](https://pkg.go.dev/net/http@go1.23.3#Dir) is of type FileSystem as it implements Open method. It's extension of string.

```go
type FileSystem interface {
	Open(name string) (File, error)
}
```

Last function that we need is [StripPrefix] (https://pkg.go.dev/net/http@go1.23.3#StripPrefix) constructor function.

```go
func StripPrefix(prefix string, h Handler) Handler
```

It takes two parameters, it takes a Handler and returns a Handler. So it's a decorator. It's going to take a Handler and it's going to modify the behaviour of that handler in someway. In this case it's going to strip prefix off.

When in path pattern we have a trailing slash, we are implying that there is a collection of resources. So think about it as a single file in your file system versus a directory.

```go
http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("."))))
```

If we open up our developer tool and click on any file, it sends correct Content-Type set.
So **FileServer** type in Go is very powerful. It allows us directory browsing, it allows us to serve up resources and it provides automatic content detection.

So if you've got multiple static resources, so for example you are building a web application and you've got CSS files, JavaScript files and image files that you need to serve up. You can setup a static file server using one simple Handler and Go is going to take care of all the heavy lifting for you.

# JSON Messaging
We have two general categories of messaging that we can talk about:
- Language Specific Messaging
  - E.g. Golang based
- Platform neutral options

## Pros & Cons
- Language specific messaging are fast, platform neutral messaging is slower.
- Language specific messaging tends to be very efficient, platform neutral messaging tends to be less efficient.
  - Platform neutral messaging need more data to be send.
- Language specific messaging is Easy to implement as they are often built into standard libraries or frameworks that are optimized to make using these strategies very easy. Platform neutral formats can add complexity.
- Language specific options tend to encourage reusable code because client and server in these type of architecture are in the same language, we often have shared source code between those clients and servers. In platform neutral messaging we also have reusability but in messaging format. So format often is defined in single place 
- With Language Specific options we have risk of platform lock-in. In platform neutral messaging we have freedom. And this is the single biggest reason that platform neutral option is mostly used.

## Platform Neutral Messaging
- JSON Messaging
- gRPC Messaging

## Sending JSON Messages
```go
import "encoding/json"
import "bytes"

type Customer struct {
    ID int
    Firstname string
    Lastname string
    Address string
}

func convertToJSON(c Customer) ([]byte, error) {
    data, err := json.Marshal(c)
    return data, err
}
```

Marshal function coverts Go objects into JSON. You might notice that Marshal returns []byte. JSON is text based protocol but since it's often used to communicate between services those inter service communications are always handled at the binary level and the way that Go models is using byte slices.

In Go wherever we have public field, it starts with capital letter and so when we convert it into JSON representation, GO is just going to take those field names and it's going to use those as field identifiers in our JSON message. The problem comes in when we look at the JavaScript convention, which is what JSON follows for variable naming and that is it uses lowercase first letters almost universally. So, if were to just convert this Customer object into a JSON representation, the field names would look right to us as GO devlopers but it wouldn't look right to anybody else because where we would pass in a Capital ID, they would be expecting lowercase id.

The way that we describe that transformation between the go field names and the JavaScript convention is using what are called **Struct Field tags** as below:

```go
type Customer struct {
    ID int  `json="id"`
    Firstname string `json="firstName"`
    Lastname string `json="lastName"`
    Address string  `json="address"`
}
```

Marshal function is not the only way to convert from a Go object to JSON representation. We also have an object that we can construct from the JSON package and that is called an **Encoder**.

```go
func convertToJSON(c Customer) ([]byte, error) {
    var b bytes.Buffer
    enc := json.NewEncoder(b)
    err := enc.Encode(c)
    return b.Bytes(), err
}
```

Encoders are resuable however marshal function is one shot deal. If you have go one Go object that you want to convert to JSON, then Marshal function is going to be a great way to go.

However, if you are encoding multiple objects, which can sometimes happen in streaming scenarios if you want to encode multiple objects to the response stream then the Encoder allows that Encode method to be called multiple times so you can pass multiple objects in, and then the client can receive those however it needs to.

```go
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
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
		log.Fatal(s.ListenAndServeTLS("../cert.pem", "../key.pem"))
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
```

## Reading JSON Messages
```go
import "encoding/json"
import "bytes"

type Customer struct {
    ID        int    `json="id"`
	Firstname string `json="firstName"`
	Lastname  string `json="lastName"`
	Address   string `json="address"`
}

func convertFromJSON(data []byte) (Customer, error) {
    var c Customer
    err := json.UnMarshal(data, &c)
    return c, err
}
```

Here Go used reflection to understand what is the structure of the Object &c.
It can also populate a map instead of GO object.

What happens if there are multiple objects coming in? For that we have the **Decoder**. Decoder takes a datasource that we can read from consistently. We can pass request body but we don't have one. In this case we can pass data of type []byte but byte slice is not a reader. So we are going to create a buffer from bytes package.

```go
func convertFromJSON(data []byte) (Customer, error) {
    b := bytes.NewBuffer(data)  // must be an io.Reader
    dec := json.NewDecoder(b)
    var c Customer //could use map[string]any too
    err := dec.Decode(&c)
    return c, err
}
```

# Routing Requests
When we talk about HTTP routing, we generally have three types of routing concerns to keep in mind.
- The request for single resource and it uses URL patterns like: /customers
- Request for resource collection: /customers/
- Parametric routes: Request a resource by parameter: /customers/{id}

## Basic Routing
Basic routing really boils down to below part of the parameter list when we call the **Handle** function or the **HandleFunc** function from the http package.

```go
http.Handle(pattern string, handler http.Handler)
```

Go is going to select the most specific pattern that it can that meets the incoming URL pattern. Now there are two different forms of this pattern that we can register with the **Handle** and **HandleFunc** functions.

- request a single resource
```go
http.Handle("/customers", handler)
```
- request a resource collection
```go
http.Handle("/customers/", handler)
```

## Parametric Routing
A very common practice with HTTP services is encoding information into the URL itself that we rely on to help the microservice understand what we are asking it to do.

There are generally two techniques available to us within the standard library to extract those parameters from the URL:

### String Splitting
```go
func handlerFunc(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path // "/customers/123/address/city"

    parts := strings.Split(path, "/") // ["" "customers "123" "address" "city"]

    // validate the parameters & route requests.
}
```

So in this case we have two parameters 123 & city. In String splitting we simply split the string with '/'. We need to validate the parameters & route requests because it doesn't guarantee that parameters are properly formatted.

### Using Regular Expression
Let's assume in ```/customers/123/address/city``` customers and address are static part of the route and 123 & city are parameters, they can change but we want the same basic handling we're just going to apply slightly different business logic depending on the parameters that are coming into our handler.

So how do we break this down with regular expressions, and what are the advantages?
We are going to start with Regular Expression that looks like this pattern and that looks something like this:

```go
var pattern = `^\/customers\/(\d+?)\/address\/(\S+?)`
var exp regexp.Regexp = regexp.MustCompile(pattern)
```

```\S``` instructs to match non-whitespace characters.
```\d``` instructs to mathc numeric value.

With Go regular expressions are not just strings, they are objects and we construct those by using from the ```regexp``` package, we construct that regular expression. 

There are multiple constructor functions that are available, we are using ```MustCompile``` constructor function.

Once we have regular expression object there's rich API of matching methods that are included on those regular expressions. In this case we are using **FindStringSubmatch**. If you look at regular expression package there are 20-30 different methods that are available, each one tuned for specific usecase that we want.

In this case, we decompose **FindStringSubmatch**, we see that **Find** means that we're going to try and find a match and **String** means that we're matching against a string and the reason we have that is because Go allows us to match regular expressions against byte slices as well, so we need to be specific about what we're providing. And then **Submatch** indicates that there are elements of the match that we are interested in extracting later.

So we don't just want Go to say, yep, this string matches your pattern. We also want to provide key pieces of information so that we can work with those later and those key pieces of information, if we go back up to the pattern, notice that we have delimited the parameters in these parantheses. These parantheses indicate that this is interesting for me to work with later. In other words, these are our Submatches and that's why we see in the matches slice that we get as a result we see that we have 3 elements.

First element is always the full string that got matched, then after that there are any submatches or any what are called capture groups in the terminology of regular expressions.
So our first capture group, our first set of parantheses is around that string "123", and our second capture group is around the pattern that gets matched to the word "city".

```go
var pattern = `^\/customers\/(\d+?)\/address\/(\S+?)`
var exp regexp.Regexp = regexp.MustCompile(pattern)

func handlerFunc(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path // "/customers/123/address/city"
    matches := exp.FindStringSubmatch(path) // ["/customers/123/address/city", "123", "city"]
}
```

So while Regular expressions are more complicated to construct, they do often make the validation logic in our handler a little bit easier. Having said that though we often do have a little bit of validation that's required.

## Common Third Party Routers
Two popular thrid party routers are:
- **gorilla/mux**
```sh
go get github.com/gorilla/mux
```
We can work with mux router without dealing with complexity of regular expression but get the benifits of regular expression

```go
r := mux.NewRouter()
r.HandleFunc("/products/{id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idRaw := vars["id"]
    if len(idRaw) == 0 {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    id, err := strconv.Atoi(idRaw)
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

http.Handle("/", r)
```
- **go-chi/chi**
```sh
go get github.com/go-chi/chi/v5
```

# Middlewares
When we have concerns that are not directly related to business logic of the webservice, we keep that logic in middlewares.

## What is Middleware
So far, all of our discussions have been around this thing called a **handler**. That is sometimes a function and sometimes an object. But it's always intended to be the thing that's going to hold the business logic that actually handles the request coming in from a client.

So, the model for how these things work is, a request comes in, the handler gets invoked, executes it's business logic, and then a response goes back out. Middleware changes our model slightly by sitting in the middle of Request-Response pipeline.

So instead of request directly going to our request handler, now the request is going to go the middleware. The middleware is going to pass that request on to the request handler, or optionally we could have more middleware in that chain, and then the response is going to go back to the middleware as well.

We often have concerns in our microservices that are not directly related to our business logic e.g. 
- We might want to add a caching strategy into our service.
- We might want to put an authentication mechanism
- We can have session management.
- We can add logging and telemetry into our systems.
- We can do response compression, making sure that we are making efficient use of network resources by compressing the responses down and minimizing the data that's sent.

## Creating Middleware
Middleware sits in the middle of request-response pipeline. Middleware receives the request, passes it on to the request handler, but it also has to generate the response, which means it has the same responsibilities as any other handler.

Let's look at **ListenAndServe** function from **http** package.
```go
http.ListenAndServe(addr string, handler Handler) error
```

So far we have been passing nil for the second parameter because that tells Go to use something called the **DefaultServeMux**, which handles all of the internal routing logic of our application and makes a lot of the hard work for creating services, very simple for us. But this is a valid parameter that we can do other things with.

Middleware is just a **Handler**. 
```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

However, there is one trick or one pattern that we often introduce that makes a distinction between what is middleware and what is normal request handler.

So if we start out to build our middleware we have to answer the question, how do we get that request pass along the chain? This is how we often do it. **We build our middleware as a struct**.

```go
type MyMiddleware struct {
    Next http.Handler
}

func (m MyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // do something like authentication, logging etc. before next handler
    m.Next.ServeHTTP()  // pass request to next handler
    // do things after next handler
} 
```

## Types Of Middlewares
- Global
  - Applies to all requests
```go
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
```
- Route-Specific
  - Applies to specific handler
```go
type validationMiddleware struct {
	next http.Handler
}

func (vm *validationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if vm.next == nil {
		log.Panic("No next handler defined for validation middleware")
	}
	if r.Method != http.MethodPost {
		vm.next.ServeHTTP(w, r)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var c Cart
	err = json.Unmarshal(data, &c)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := http.Head(fmt.Sprintf("http://localhost:3000/customers/%v", c.CustomerID))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if res.StatusCode == http.StatusNotFound {
		log.Println("Invalid customer ID")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	b := bytes.NewBuffer(data)
	r.Body = io.NopCloser(b)
	vm.next.ServeHTTP(w, r)
}
```

**DefaultServeMux** makes sense for global middleware because that gets invoked before the routing in the application. With route-specific middleware we've already gone through the routing process, we know exactly what handler should be invoked and so there is no default that we can pick if that next handler has not been defined.

Request body can be read once. So we need to be careful, if we are reading it in middleware we can't read it anywhere else. So for example above in validationMiddleware we already read the request body so it will fail in running cartsHandler post method handling, where we are reading request body.
So how do we get request body back again?
So we create a new buffer from the bytes package and populate it with the data that we got when we read the request body out. So actually we are going to create a brand new request body, but we are going to create it using buffer instead of the default object that came in with the request. This works because the request body simply has to implement the **ReadCloser** interface. We don't actually have any details about what actual object Go is using here. 

Only thing that we have to keep in mind is that **ByteBuffer** doesn't have a **Close** method on it, so there is a decorator that we can use from the io package called **NopCloser**.

```go
b := bytes.NewBuffer(data)
r.Body = io.NopCloser(b)
```

Now how do we actually get the route specific middleware registered?
Global middleware we register at server level. Route specific middleware, we register at handler level. So just like Global middleware wraps DefaultServeMux or the serve mux that we are using, we are going to wrap our handler with our validationMiddleware.

```go
cartMux.Handle("/carts", &validationMiddleware{next: http.HandlerFunc(cartsHandler)})
```

# gRPC Messaging
## gRPC Structure
The way client and server are going to communicate with gRPC is using some generated code and that's going to be one of the key things that we are going to be talking about.
This generated code on the client talks to the generated code on the server using a **Transport Protocol** and that transport protocol is used for both sending the request to the server and getting those responses back.

The primary protocol that is used in gRPC is a technology called **ProtocolBuffers**.

## Protocol Buffers 101
```proto
syntax = "proto3";
package product;
option go_package = "demo/productpb";
message Product {
	int32 id = 1;
	string name = 2;
	double usdPerUnit = 3;
	string unit = 4;
}
```

Both JSON and XML have a shortcoming where they actually have to encode the schema or the format of the data along with the data itself, which made network messaging with JSON & XML inefficient.
So **ProtocolBuffer** attempts to solve that problem by using numerous techniques.

One of those techniques is we are going to define the schema before we send the messages. So clients and servers are both going to know the structure of the data, not because we send the data schema, but because they have that pre-generated for them.

So let's talk about defining messages. So let's consider the above example:
- First line is syntax line. There are multiple versions of ProtocolBuffers that have been developed over time. In this case we are usin version 3.
- Next line is package identifier. This doesn't have anything to do with the organization of our generated code. This is the package identifier we use when we have **ProtocolBuffer** definition files that are talking to one another.
- Next line is go_package option. There are quite a few optional parameters that can be added into a definition file that controls the code generation for specific languages. In this case the go_package option is going to specify the package identifier that we want our generated code to use.
- Next thing that we have is message, this defines some data structure that can be sent across the ProtocolBuffer. The number in the end of the field is critical because the way ProtocolBuffer works is that **when it actually transmits a message from a client to a server it doesn't send the field names, it sends these field identifiers.** So field is going to be interpreted by the client and the server as the field id because we've defined that in this message format.

Now to generate code, we download a compiler called protoc compiler.
https://protobuf.dev/downloads/

When we are working with Go, protoc compiler is not enough. So we are going to add on an additional tool that's going to be installed in our module and this is **protoc-gen-go** command that we don't use directly but is used by the **protoc** compiler when we are generating Go soure code.

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Finally we invoke the protoc compiler using below command:
```sh
protoc -I=. --go_out=. /product.proto
```

-I in include parameter that tells the protoc compiler where to find all the ProtocolBuffer messages -go_out parameter defines where we want the go source to be rooted at.

To learn more about protocol buffer: https://protobuf.dev

We can define the product as below:
```go
p := productpb.Product{
	Id:         int32(products[0].ID),
	Name:       products[0].Name,
	UsdPerUnit: products[0].USDPerUnit,
	Unit:       products[0].Unit,
}
```

Now, the next thing we have to do is we have to convert this Protocol Buffer message into a format that can be sent across the network. We are going to use a Marshal function from **proto** package. **proto** package is not part of standard library. To grab that:
```sh
go get google.golang.org/protobuf
```

```go
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
```

If we print it, it looks like garbage data. That's actually one of the things that we like about Protocol Buffers. While it's not necessarily very readable on the wire, this is much more efficient format for sending this data structure than when we compare it to something like JSON.

SO it's not only compressed format, but it's also transmitted as binary data too, which can be more efficient than the string based messaging that JSON uses.

How do we convert back to product we can work with?
