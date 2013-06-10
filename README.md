# Pathways - a RESTful web service framework for Go

The goal of Pathways is to make building RESTful web services simple.

Pathways centers around the concept of services. Services define exactly how every endpoint in a web service is accessed and handled. This provides several benefits, including the ability to construct client requests, and the ability to autogenerate API documentation.

## Example

Here's an example of an in-memory key-value service and client with Pathways:

```go
package main

import (
    "github.com/alecthomas/pathways"
    "net/http"
)

type KeyValueService struct {
    service *pathways.Service
    kv      map[string]string
}

func KeyValueServiceMap(root string) *pathways.Service {
    s := pathways.NewService(root)
    s.Path("/").Name("List").Get().ApiResponseType(map[string]string{})
    str := ""
    s.Path("/{key}").Name("Get").Get().ApiResponseType(&str)
    s.Path("/{key}").Name("Create").Post().ApiRequestType(&str)
    s.Path("/{key}").Name("Delete").Delete()
    return s
}

func NewClient(url string) *pathways.Client {
    s := KeyValueServiceMap(url)
    return pathways.NewClient(s, "application/json")
}

func NewKeyValueService(root string) *KeyValueService {
    k := &KeyValueService{
        service: KeyValueServiceMap(root),
        kv:      make(map[string]string),
    }
    k.service.Find("List").ApiFunction(k.List)
    k.service.Find("Get").ApiFunction(k.Get)
    k.service.Find("Create").ApiFunction(k.Create)
    k.service.Find("Delete").ApiFunction(k.Delete)
    return k
}

func (k *KeyValueService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    k.service.ServeHTTP(w, r)
}

func (k *KeyValueService) List(cx *pathways.Context) pathways.Response {
    return cx.ApiResponse(http.StatusOK, k.kv)
}

func (k *KeyValueService) Get(cx *pathways.Context) pathways.Response {
    return cx.ApiResponse(http.StatusOK, k.kv[cx.PathVars["key"]])
}

func (k *KeyValueService) Create(cx *pathways.Context, value *string) pathways.Response {
    k.kv[cx.PathVars["key"]] = *value
    return cx.ApiResponse(http.StatusCreated, "ok")
}

func (k *KeyValueService) Delete(cx *pathways.Context) pathways.Response {
    delete(k.kv, cx.PathVars["key"])
    return cx.ApiResponse(http.StatusOK, &struct{}{})
}

func main() {
    s := NewKeyValueService("/api/")
    http.ListenAndServe(":8080", s)
}
```

Test with:

```bash
$ curl http://localhost:8080/api/
{}
$ curl --data-binary '"hello world"' http://localhost:8080/api/foo
"ok"
$ curl http://localhost:8080/api/
{"foo":"hello world"}
$ curl http://localhost:8080/api/foo
"hello world"
```
## Features

### A simple yet flexible URL routing engine

For example, the following service definition might specify routes for a key/value store:

```go
s := pathways.NewService("/kv/")
s.Path("/").Name("List").Get()
s.Path("/{key}").Name("Create").Post()
s.Path("/{key}").Name("Get").Get()
s.Path("/{key}").Name("Delete").Delete()
```

### Automatic serialization/deserialization of requests/responses

Pathways routes can define the request and response structures expected, and route directly to functions and methods, passing the deserialized request as an argument:

```go
type KeyValueService struct {
}

func (k *KeyValueService) Create(cx *pathways.Context, req *CreateRequest) pathways.Response {
    // ... do something with deserialized request
    return cx.ApiResponse(http.StatusOK, &CreateResponse{})
}

kvs := &KeyValueService{}

s.Path("/{key}").Name("Create").Post().ApiRequestType(&CreateRequest{}).ApiResponseType(&CreateResponse{}).ApiFunction(kvs.Create)
```

### RESTful client using the service definition

The following will issue a `GET` request to `/kv/key` with the request body from `CreateRequest`. The response will be returned as a `CreateResponse` structure:

```go
c := pathways.NewClient(s, "application/x-msgpack")
response := &CreateResponse{}
err := c.Call("Create", pathways.Args{"key": "somekey"}, &CreateRequest{...}, response)
```

### Transparent support for JSON, MsgPack and BSON serialized requests/responses

The content-types for these formats are `application/json`, `application/x-msgpack` and `application/bson`. Setting the request `Content-Type` and/or `Accept` headers to one of these will set the desired serialization format. The default is JSON.

The Pathways server will detect the correct serialization format from request headers (falling back on JSON). The Pathways client will use the serialization format specified in the constructor.

Why not support only JSON? Primarily because JSON has [limitations on the numeric values](http://cdivilly.wordpress.com/2012/04/11/json-javascript-large-64-bit-integers/) that can be represented.
