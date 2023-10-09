package main

import (
	"embed"
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
	"net/http"
)

//go:embed petstore.yaml
var fs embed.FS

func main() {
	// open the embedded spec file...
	f, err := fs.Open("petstore.yaml")
	if err != nil {
		panic(err)
	}

	opts := &chioas.FromOptions{
		// serve interactive api docs...
		DocOptions: &chioas.DocOptions{ServeDocs: true},
		// the api implementation with handler methods - e.g. x-handler: ".GetPets"
		Api: &myApi{},
		// lookup for handler funcs - e.g. x-handler: "GetRoot"
		Handlers: chioas.Handlers{
			"GetRoot": getRoot,
		},
	}
	// read definition from spec file...
	// Note: look at the example petstore.yaml and notice the `x-handler` additional tags on each method
	def, err := chioas.FromYaml(f, opts)
	if err != nil {
		panic(err)
	}
	_ = f.Close()

	// setup routes and serve api...
	// run main and browse to http://localhost:8080/docs
	router := chi.NewRouter()
	err = def.SetupRoutes(router, nil)
	if err != nil {
		panic(err)
	}
	_ = http.ListenAndServe(":8080", router)
}
