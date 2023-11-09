// Package typed provides for the capability of having typed handlers (methods/functions) for api endpoints
//
// # To utilise typed handlers, just set the MethodHandlerBuilder in chioas.Definition
//
// For full documentation - see https://github.com/go-andiamo/chioas/blob/main/typed/README.md
//
// Example:
//
//	package main
//
//	import (
//		"github.com/go-andiamo/chioas"
//		"github.com/go-andiamo/chioas/typed"
//		"github.com/go-chi/chi/v5"
//		"net/http"
//	)
//
//	var def = chioas.Definition{
//		DocOptions: chioas.DocOptions{
//			ServeDocs: true,
//		},
//		MethodHandlerBuilder: typed.NewTypedMethodsHandlerBuilder(nil),
//		Methods: map[string]chioas.Method{
//			http.MethodGet: {
//				Handler: func() (map[string]any, error) {
//					return map[string]any{
//						"root": "root discovery",
//					}, nil
//				},
//			},
//		},
//		Paths: map[string]chioas.Path{
//			"/people": {
//				Methods: map[string]chioas.Method{
//					http.MethodGet: {
//						Handler: "GetPeople",
//						Responses: chioas.Responses{
//							http.StatusOK: {
//								IsArray:   true,
//								SchemaRef: "Person",
//							},
//						},
//					},
//					http.MethodPost: {
//						Handler: "AddPerson",
//						Request: &chioas.Request{
//							Schema: personSchema,
//						},
//						Responses: chioas.Responses{
//							http.StatusOK: {
//								SchemaRef: "Person",
//							},
//						},
//					},
//				},
//			},
//		},
//		Components: &chioas.Components{
//			Schemas: chioas.Schemas{
//				personSchema,
//			},
//		},
//	}
//
//	type api struct {
//		chioas.Definition
//	}
//
//	var myApi = &api{
//		Definition: def,
//	}
//
//	func (a *api) SetupRoutes(r chi.Router) error {
//		return a.Definition.SetupRoutes(r, a)
//	}
//
//	type Person struct {
//		Id   int    `json:"id" oas:"description:The id of the person,example"`
//		Name string `json:"name" oas:"description:The name of the person,example"`
//	}
//
//	var personSchema = (&chioas.Schema{
//		Name: "Person",
//	}).Must(Person{
//		Id:   1,
//		Name: "Bilbo",
//	})
//
//	// GetPeople is the typed handler for GET /people - note the typed return args
//	func (a *api) GetPeople() ([]Person, error) {
//		return []Person{
//			{
//				Id:   0,
//				Name: "Me",
//			},
//			{
//				Id:   1,
//				Name: "You",
//			},
//		}, nil
//	}
//
//	// AddPerson is the typed handler for POST /person - note the typed input and return args
//	func (a *api) AddPerson(person *Person) (Person, error) {
//		return *person, nil
//	}
//
//	func main() {
//		router := chi.NewRouter()
//		if err := myApi.SetupRoutes(router); err != nil {
//			panic(err)
//		}
//		_ = http.ListenAndServe(":8080", router)
//	}
package typed
