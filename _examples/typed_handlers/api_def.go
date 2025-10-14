package main

import (
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/typed"
	"net/http"
	"strconv"
)

var def = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs: true,
		UIStyle:   chioas.Rapidoc,
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
					Handler: (*api).GetPeople,
					Responses: chioas.Responses{
						http.StatusOK: {
							IsArray:   true,
							SchemaRef: "Person",
						},
					},
					QueryParams: chioas.QueryParams{
						{
							Name:        SearchParam("").QueryParamName(),
							Description: "Search people by name",
						},
					},
				},
				http.MethodPost: {
					Handler: (*api).AddPerson,
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
			Paths: chioas.Paths{
				"/{" + PersonId(0).PathParamName() + "}": {
					PathParams: chioas.PathParams{
						PersonId(0).PathParamName(): {
							Description: "ID of person",
							Schema: &chioas.Schema{
								Type: "integer",
							},
						},
					},
					Methods: chioas.Methods{
						http.MethodGet: {
							Handler: (*api).GetPerson,
							Responses: chioas.Responses{
								http.StatusOK: {
									SchemaRef: "Person",
								},
							},
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

var personSchema = (&chioas.Schema{
	Name: "Person",
}).Must(Person{
	Id:   1,
	Name: "Bilbo",
})

type SearchParam string

func (SearchParam) QueryParamName() string {
	return "search"
}

type PersonId int

func (PersonId) PathParamName() string {
	return "personId"
}

// UnmarshalText by implementing this method, the named path param can be a non-string (and self validating)
func (id *PersonId) UnmarshalText(text []byte) error {
	if i, err := strconv.Atoi(string(text)); err == nil {
		*id = PersonId(i)
		return nil
	} else {
		return err
	}
}
