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
