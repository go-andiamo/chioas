# CHIOAS
[![GoDoc](https://godoc.org/github.com/go-andiamo/chioas?status.svg)](https://pkg.go.dev/github.com/go-andiamo/chioas)
[![Latest Version](https://img.shields.io/github/v/tag/go-andiamo/chioas.svg?sort=semver&style=flat&label=version&color=blue)](https://github.com/go-andiamo/chioas/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-andiamo/chioas)](https://goreportcard.com/report/github.com/go-andiamo/chioas)

*Chioas* is an add-on for the popular [Chi router](https://github.com/go-chi/chi)

> pronounce *Chioas* however you like...
> 
>  but I can't help feeling that it looks a little like "chaos" - the chaos of undocumented APIs that it tries to solve!

### What it does

*Chioas* does three things:
* Defines your API as a single (or modular) struct and then builds the Chi routes accordingly
* Produces an OpenApi spec (OAS) of that definition
* Optionally serves an interactive API docs web-page (as part of your service)

### Advantages
With your actual API and OpenAPI spec (OAS) being specified in _**one**_ place there's no chance of them becoming out of sync! (eliminating having to manually update spec yaml/json to match API or vice versa)

You can even keep your request/response OAS schemas in sync with your request/response structs!<br>
See [petstore example](https://github.com/go-andiamo/chioas/tree/main/_examples/petstore)

### Already have an OAS spec?
No problem, use `chioas.FromJson()` or `chioas.FromYaml()` to read the spec definition.
All you'll need to do is add `x-handler` tags to each method in the spec.

See [From example](https://github.com/go-andiamo/chioas/tree/main/_examples/from)

### Choice of UI styles
*Chioas* supports two different styles of API docs UI:
1. Redoc (see [Redocly.redoc](https://github.com/Redocly/redoc) and [Redoc-try](https://github.com/wll8/redoc-try))
2. Swagger (see [swagger-ui](https://github.com/swagger-api/swagger-ui))

Set the desired style using `DocOptions.UIStyle` (the default is `Redoc`).
Notes:
* statics for `Redoc` are served from CDNs
* statics for `Swagger` are all served directly from *Chioas*

## Installation
To install chioas, use go get:

    go get github.com/go-andiamo/chioas

To update chioas to the latest version, run:

    go get -u github.com/go-andiamo/chioas

### Basic Example
```go
package main

import (
    "github.com/go-andiamo/chioas"
    "github.com/go-chi/chi/v5"
    "net/http"
)

func main() {
    router := chi.NewRouter()
    if err := myApi.SetupRoutes(router, myApi); err != nil {
        panic(err)
    }
    _ = http.ListenAndServe(":8080", router)
}

var myApi = chioas.Definition{
    AutoHeadMethods: true,
    DocOptions: chioas.DocOptions{
        ServeDocs:       true,
        HideHeadMethods: true,
    },
    Paths: chioas.Paths{
        "/foos": {
            Methods: chioas.Methods{
                http.MethodGet: {
                    Handler: getFoos,
                    Responses: chioas.Responses{
                        http.StatusOK: {
                            Description: "List of foos",
                            IsArray:     true,
                            SchemaRef:   "foo",
                        },
                    },
                },
                http.MethodPost: {
                    Handler: postFoos,
                    Request: &chioas.Request{
                        Description: "Foo to create",
                        SchemaRef:   "foo",
                    },
                    Responses: chioas.Responses{
                        http.StatusCreated: {
                            Description: "New foo",
                            SchemaRef:   "foo",
                        },
                    },
                },
                http.MethodHead: {
                    Handler: getFoos,
                },
            },
            Paths: chioas.Paths{
                "/{fooId}": {
                    Methods: chioas.Methods{
                        http.MethodGet: {
                            Handler: getFoo,
                            Responses: chioas.Responses{
                                http.StatusOK: {
                                    Description: "The foo",
                                    SchemaRef:   "foo",
                                },
                            },
                        },
                        http.MethodDelete: {
                            Handler: deleteFoo,
                        },
                    },
                },
            },
        },
    },
    Components: &chioas.Components{
        Schemas: chioas.Schemas{
            {
                Name:               "foo",
                RequiredProperties: []string{"name", "address"},
                Properties:         chioas.Properties{
                    {
                        Name: "name",
                        Type: "string",
                    },
                    {
                        Name: "address",
                        Type: "string",
                    },
                },
            },
        },
    },
}

func getFoos(writer http.ResponseWriter, request *http.Request) {
}

func postFoos(writer http.ResponseWriter, request *http.Request) {
}

func getFoo(writer http.ResponseWriter, request *http.Request) {
}

func deleteFoo(writer http.ResponseWriter, request *http.Request) {
}
```
Run and then check out http://localhost:8080/docs !

Or simply generate the OpenAPI spec...
```go
package main

import (
    "net/http"

    "github.com/go-andiamo/chioas"
)

func main() {
    data, _ := myApi.AsYaml()
    println(string(data))
}

var myApi = chioas.Definition{
    AutoHeadMethods: true,
    DocOptions: chioas.DocOptions{
        ServeDocs:       true,
        HideHeadMethods: true,
    },
    Paths: chioas.Paths{
        "/foos": {
            Methods: chioas.Methods{
                http.MethodGet: {
                    Handler: getFoos,
                    Responses: chioas.Responses{
                        http.StatusOK: {
                            Description: "List of foos",
                            IsArray:     true,
                            SchemaRef:   "foo",
                        },
                    },
                },
                http.MethodPost: {
                    Handler: postFoos,
                    Request: &chioas.Request{
                        Description: "Foo to create",
                        SchemaRef:   "foo",
                    },
                    Responses: chioas.Responses{
                        http.StatusCreated: {
                            Description: "New foo",
                            SchemaRef:   "foo",
                        },
                    },
                },
                http.MethodHead: {
                    Handler: getFoos,
                },
            },
            Paths: chioas.Paths{
                "/{fooId}": {
                    Methods: chioas.Methods{
                        http.MethodGet: {
                            Handler: getFoo,
                            Responses: chioas.Responses{
                                http.StatusOK: {
                                    Description: "The foo",
                                    SchemaRef:   "foo",
                                },
                            },
                        },
                        http.MethodDelete: {
                            Handler: deleteFoo,
                        },
                    },
                },
            },
        },
    },
    Components: &chioas.Components{
        Schemas: chioas.Schemas{
            {
                Name:               "foo",
                RequiredProperties: []string{"name", "address"},
                Properties: chioas.Properties{
                    {
                        Name: "name",
                        Type: "string",
                    },
                    {
                        Name: "address",
                        Type: "string",
                    },
                },
            },
        },
    },
}

func getFoos(writer http.ResponseWriter, request *http.Request) {
}

func postFoos(writer http.ResponseWriter, request *http.Request) {
}

func getFoo(writer http.ResponseWriter, request *http.Request) {
}

func deleteFoo(writer http.ResponseWriter, request *http.Request) {
}
```
[try on go-playground](https://go.dev/play/p/0zaWsmsw2FD)

## Typed Handlers
see [typed README](https://github.com/go-andiamo/chioas/blob/main/typed/README.md)