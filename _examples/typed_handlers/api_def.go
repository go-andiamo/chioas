package main

import (
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/typed"
	"github.com/go-andiamo/urit"
	"net/http"
	"reflect"
	"strconv"
)

var def = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs: true,
	},
	MethodHandlerBuilder: typed.NewTypedMethodsHandlerBuilder(&PersonIdArgBuilder{}),
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
			Paths: chioas.Paths{
				"/{personId}": {
					Methods: chioas.Methods{
						http.MethodGet: {
							Handler: "GetPerson",
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

type PersonId int

// PersonIdArgBuilder is an implementation of typed.ArgBuilder that extracts an arg type of PersonId from path params
type PersonIdArgBuilder struct{}

var _ typed.ArgBuilder = &PersonIdArgBuilder{}

func (ab *PersonIdArgBuilder) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	return argType == reflect.TypeOf(PersonId(0)), false
}

func (ab *PersonIdArgBuilder) BuildValue(argType reflect.Type, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	v := PersonId(-1)
	var err error = typed.NewApiError(http.StatusInternalServerError, "path param 'personId' not present")
	for _, pv := range params {
		if pv.Name == "personId" {
			var i int64
			if i, err = strconv.ParseInt(pv.Value.(string), 10, 32); err == nil {
				v = PersonId(i)
			} else {
				err = typed.WrapApiErrorMsg(http.StatusBadRequest, err, "personId is not a integer")
			}
			break
		}
	}
	return reflect.ValueOf(v), err
}
