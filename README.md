# Pathways - an opinionated RESTful web service framework for Go

The goal of Pathways is to make building RESTful web services simple.

*Opinionated*, in the context of Pathways, simply means the API prefers API request context to be in the request body, not in query parameters.

Pathways centers around the concept of services which define exactly how every endpoint in a web service is represented and handled. This provides several benefits, including the ability to construct client requests, and the ability to autogenerate API documentation.

It includes the following features:

## A simple yet flexible URL routing engine

For example, the following service definition might specify routes for a key/value store:

```go
s := pathways.NewService("/kv/")
s.Path("/").Name("List").Get()
s.Path("/{key}").Name("Create").Post()
s.Path("/{key}").Name("Get").Get()
s.Path("/{key}").Name("Delete").Delete()
```

## Automatic serialization/deserialization of requests/responses

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

## RESTful client using the service definition

The following will issue a `GET` request to `/kv/key` with the request body from `CreateRequest`. The response will be returned as a `CreateResponse` structure:

```go
c := pathways.NewClient(s, "application/x-msgpack")
response := &CreateResponse{}
err := c.Call("Create", pathways.Args{"key": "somekey"}, &CreateRequest{...}, response)
```

## Transparent support for JSON, MsgPack and BSON serialized requests/responses

The content-types for these formats are `application/json`, `application/x-msgpack` and `application/bson`. Setting the request `Content-Type` and/or `Accept` headers to one of these will set the desired serialization format. The default is JSON.

The Pathways server will detect the correct serialization format from request headers (falling back on JSON). The Pathways client will use the serialization format specified in the constructor.

Why not support only JSON? Primarily because JSON has [limitations on the numeric values](http://cdivilly.wordpress.com/2012/04/11/json-javascript-large-64-bit-integers/) that can be represented.
