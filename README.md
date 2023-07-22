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
    if err := myApi.SetupRoutes(router); err != nil {
        panic(err)
    }
    _ = http.ListenAndServe(":8080", router)
}

var myApi = chioas.Definition{
    DocOptions: chioas.DocOptions{
        ServeDocs: true,
    },
    Paths: map[string]chioas.Path{
        "/foos": {
            Methods: map[string]chioas.Method{
                http.MethodGet: {
                    Handler: getFoos,
                },
                http.MethodPost: {
                    Handler: postFoos,
                },
            },
            Paths: map[string]chioas.Path{
                "/{fooId}": {
                    Methods: map[string]chioas.Method{
                        http.MethodGet: {
                            Handler: getFoo,
                        },
                        http.MethodDelete: {
                            Handler: deleteFoo,
                        },
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
