# Chioas Typed Handlers

Having [http.HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc) like...
```go
func postPet(w http.ResponseWriter, r *http.Request)
```
Is, in many ways, very convenient... it's a common de-facto practice, it's flexible (can unmarshal requests and marshal responses directly)<br>
However, there are some drawbacks...
* you have to read the code to understand what the request and response should be
* not great for APIs that support multiple content types (e.g. the ability to POST in either `application/json` or `text/xml`)
* errors have to translated into http status codes (and messages unmarshalled into response body) 

So _**Chioas Typed Handlers**_ solves this by allowing, for example...
```go
func postPet(req AddPetRequest) (Pet, error)
```

## How To Use
To optionally use _Chioas Typed Handlers_ just replace the default handler builder...
```go
chioas.Definition{
    MethodHandlerBuilder: typed.NewTypedMethodsHandlerBuilder()
}
```

## Handler Input Args
_Chioas Typed Handlers_ looks at the types of each handler arg to determine what needs to be passed.  This is based on the following rules...

| Signature                            | Description                                                                                                                                                                |
|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `func eg(w http.ResponseWriter)`     | `w` will be the original [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)                                                                                 |
| `func eg(r *http.Request)`           | `r` will be the original [*http.Request](https://pkg.go.dev/net/http#Request)                                                                                              |
| `func eg(ctx context.Context)`       | `ctx` will be the context from original [*http.Request](https://pkg.go.dev/net/http#Request)                                                                               |
| `func eg(ctx *chi.Context)`          | `ctx` will be the Chi context extracted from original [*http.Request](https://pkg.go.dev/net/http#Request)                                                                 |
| `func eg(ctx chi.Context)`           | `ctx` will be the Chi context extracted from original [*http.Request](https://pkg.go.dev/net/http#Request)                                                                 |
| `func eg(hdrs http.Header)`          | `hdrs` will be the request headers from original [*http.Request](https://pkg.go.dev/net/http#Request.Header)                                                               |
| `func eg(hdrs typed.Headers)`        | `hdrs` will be the request headers from original [*http.Request](https://pkg.go.dev/net/http#Request.Header)                                                               |
| `func eg(hdrs map[string][]string)`  | `hdrs` will be the request headers from original [*http.Request](https://pkg.go.dev/net/http#Request.Header)                                                               |
| `func eg(pps typed.PathParams)`      | `pps` will be the path params extracted from [*http.Request](https://pkg.go.dev/net/http#Request.URL)                                                                      |
| `func eg(cookies []*http.Cookie)`    | `cookies` will be the cookies extracted from [*http.Request](https://pkg.go.dev/net/http#Request.Cookies)                                                                  |
| `func eg(url *url.URL)`              | `url` will be the URL from original [*http.Request](https://pkg.go.dev/net/http#Request.URL)                                                                               |
| `func eg(pps ...string)`             | `pps` will be the path param values                                                                                                                                        |
| `func eg(pp1 string, pp2 string)`    | `pp1` will be the first path param value, `pp2` will be the second path param value etc.                                                                                   |
| `func eg(pp1 string, pps ...string)` | `pp1` will be the first path param value, `pps` will be the remaining path param values                                                                                    |
| `func eg(qps typed.QueryParams)`     | `qps` will be the request query params from [*http.Request](https://pkg.go.dev/net/http#Request.URL)                                                                       |
| `func eg(qps typed.RawQuery)`        | `qps` will be the raw request query params string from [*http.Request](https://pkg.go.dev/net/http#Request.URL)                                                            |
| `func eg(frm typed.PostForm)`        | `frm` will be the post form extracted from the [*http.Request](https://pkg.go.dev/net/http#Request.PostForm)                                                               |
| `func eg(frm typed.PostForm)`        | `frm` will be the post form extracted from the [*http.Request](https://pkg.go.dev/net/http#Request.PostForm)                                                               |
| `func eg(auth typed.BasicAuth)`      | `auth` will be the basic auth extracted from [*http.Request](https://pkg.go.dev/net/http#Request.BasicAuth)                                                                |
| `func eg(auth *typed.BasicAuth)`     | `auth` will be the basic auth extracted from [*http.Request](https://pkg.go.dev/net/http#Request.BasicAuth) or `nil` if no `Authorization` header present                  |
| `func eg(req []byte)`                | `req` will be the request body read from [*http.Request](https://pkg.go.dev/net/http#Request.Body) _(see also note 2 below)_                                               |
| `func eg(req MyStruct)`              | `req` will be the request body read from [*http.Request](https://pkg.go.dev/net/http#Request.Body) and unmarshalled into a `MyStruct` _(see also note 2 below)_            |
| `func eg(req *MyStruct)`             | `req` will be the request body read from [*http.Request](https://pkg.go.dev/net/http#Request.Body) and unmarshalled into a `*MyStruct` _(see also note 2 below)_           |
| `func eg(req []MyStruct)`            | `req` will be the request body read from [*http.Request](https://pkg.go.dev/net/http#Request.Body) and unmarshalled into a slice of `[]MyStruct` _(see also note 2 below)_ |
| `func eg(b bool)`                    | will cause an `error` when setting up routes _(see note 4 below)_                                                                                                          |
| `func eg(i int)`                     | will cause an `error` when setting up routes _(see note 4 below)_                                                                                                          |
| _etc._                               | will cause an `error` when setting up routes _(see note 4 below)_                                                                                                          |
#### Notes
1. Multiple input args can be specified - the same rules apply
2. If there are multiple arg types that involve reading the request body - this is reported as an error when setting up routes
3. To support for other arg types - provide an implementation of `typedArgBuilder` passed as an option to `typed.NewTypedMethodsHandlerBuilder(options ...any)`
4. Any other arg types will cause an error when setting up routes (unless supported by note 3)  

## Handler Return Args
Having called the handler, _Chioas Typed Handlers_ looks at the return arg types to determine what needs to be written to the [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter).
(_Note: if there are no return args - then nothing is written to [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter)_)

If any of the return args is `error` (and non-nil) then that error is written to the response writer (see below on how errors are written) - otherwise...

The following rules are applied to decide what is written to [http.ResponseWriter](https://pkg.go.dev/net/http#ResponseWriter):

| Return type               | Description                                                                                                                                                   |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `typed.ResponseMarshaler` | the `ResponseMarshaler.Marshal` method is called to determine what body, headers and status code should be written                                            |
| `typed.JsonResponse`      | the fields of `typed.JsonResponse` are used to determine what body, headers and status code should be written (unless `JsonResponse.Error` is set)            |
| `*typed.JsonResponse`     | same as `typed.JsonResponse` - unless `nil`, in which case no body is written and status code is set to **204 No Content**                                    |
| `[]byte`                  | the raw byte data is written to the body and status code is set to **200 OK** (or **204 No Content** if slice is empty)                                       |
| *anything else*   | the value is marshalled to JSON, `Content-Type` header is set to `application/json` and status code is set to **200 OK** (unless an error occurs marshalling) |

### Error Handling
By default, any `error` is handled by setting the response status to **500 Internal Server Error** and nothing is written to the response body - unless...

If the error implements `typed.ApiError` - in which case the status code is set from `ApiError.StatusCode()`

If the error implements `json.Marshaler` - then the response body is the JSON marshalled error.

All of this can be overridden by providing a `typed.ErrorHandler` as an option to `typed.NewTypedMethodsHandlerBuilder(options ...any)` (or if your api instance implements the `typed.ErrorHandler` interface) 

## How Does It Work
Yes, _Chioas Typed Handlers_ uses `reflect` to make the call to your typed handler (unless the handler is already a [http.HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc))

As with anything that uses `reflect`, there is a performance price to pay.
Although _Chioas Typed Handlers_ attempts to minimize this by gathering all the type information for the type handler up-front - so handler args are only interrogated once.

But if you're concerned with ultimate performance or want to stick to convention - _Chioas Typed Handlers_ is optional and entirely your own prerogative to use it.  Or if only some endpoints in your API are performance sensitive but other endpoints would benefit from readability (or flexible content handling; or improved error handling) then you can mix-and-match. 

## Working example
The following is a short working example of typed handlers...

```go
package main

import (
    "github.com/go-andiamo/chioas"
    "github.com/go-andiamo/chioas/typed"
    "github.com/go-chi/chi/v5"
    "net/http"
)

var def = chioas.Definition{
    DocOptions: chioas.DocOptions{
        ServeDocs: true,
    },
    MethodHandlerBuilder: typed.NewTypedMethodsHandlerBuilder(),
    Methods: map[string]chioas.Method{
        http.MethodGet: {
            Handler: func() (map[string]any, error) {
                return map[string]any{
                    "root": "root discovery",
                }, nil
            },
        },
    },
    Paths: map[string]chioas.Path{
        "/people": {
            Methods: map[string]chioas.Method{
                http.MethodGet: {
                    Handler: "GetPeople",
                    Responses: chioas.Responses{
                        http.StatusOK: {
                            IsArray:   true,
                            SchemaRef: "Person",
                        },
                    },
                },
                http.MethodPost: {
                    Handler: "AddPerson",
                    Request: &chioas.Request{
                        Schema: personSchema,
                    },
                    Responses: chioas.Responses{
                        http.StatusOK: {
                            SchemaRef: "Person",
                        },
                    },
                },
            },
        },
    },
    Components: &chioas.Components{
        Schemas: chioas.Schemas{
            personSchema,
        },
    },
}

type api struct {
    chioas.Definition
}

var myApi = &api{
    Definition: def,
}

func (a *api) SetupRoutes(r chi.Router) error {
    return a.Definition.SetupRoutes(r, a)
}

type Person struct {
    Id   int    `json:"id" oas:"description:The id of the person,example"`
    Name string `json:"name" oas:"description:The name of the person,example"`
}

var personSchema = (&chioas.Schema{
    Name: "Person",
}).Must(Person{
    Id:   1,
    Name: "Bilbo",
})

// GetPeople is the typed handler for GET /people - note the typed return args
func (a *api) GetPeople() ([]Person, error) {
    return []Person{
        {
            Id:   0,
            Name: "Me",
        },
        {
            Id:   1,
            Name: "You",
        },
    }, nil
}

// AddPerson is the typed handler for POST /person - note the typed input and return args
func (a *api) AddPerson(person *Person) (Person, error) {
    return *person, nil
}

func main() {
    router := chi.NewRouter()
    if err := myApi.SetupRoutes(router); err != nil {
        panic(err)
    }
    _ = http.ListenAndServe(":8080", router)
}
```