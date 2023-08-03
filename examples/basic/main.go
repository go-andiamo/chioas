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
				},
				http.MethodPost: {
					Handler: postFoos,
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
