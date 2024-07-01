package main

import (
	"encoding/json"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/typed"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type api struct {
	chioas.Definition
}

var _ typed.ResponseHandler = &api{}

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

var dummyPeopleStore = []Person{
	{
		Id:   0,
		Name: "Me",
	},
	{
		Id:   1,
		Name: "You",
	},
}

// GetPeople is the typed handler for GET /people - note the typed return args
func (a *api) GetPeople() ([]Person, int, error) {
	return dummyPeopleStore, http.StatusOK, nil
}

// AddPerson is the typed handler for POST /person - note the typed input and return args
func (a *api) AddPerson(person *Person) (Person, int, error) {
	person.Id = len(dummyPeopleStore)
	dummyPeopleStore = append(dummyPeopleStore, *person)
	return *person, http.StatusCreated, nil
}

// GetPerson is the typed handler for GET /person/{personId}
//
// Note the arg type PersonId - which is extracted from the path params by the PersonIdArgBuilder
func (a *api) GetPerson(personId PersonId) (Person, int, error) {
	if personId >= 0 && int(personId) < len(dummyPeopleStore) {
		return dummyPeopleStore[personId], http.StatusOK, nil
	}
	return Person{}, 0, typed.NewApiErrorf(http.StatusNotFound, "not found person id %d", personId)
}

// WriteResponse implements typed.ResponseHandler to write responses
func (a *api) WriteResponse(writer http.ResponseWriter, request *http.Request, value any, statusCode int, thisApi any) {
	if statusCode >= http.StatusContinue {
		writer.WriteHeader(statusCode)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	enc := json.NewEncoder(writer)
	_ = enc.Encode(value)
}

// WriteErrorResponse implements typed.ResponseHandler to write error responses
func (a *api) WriteErrorResponse(writer http.ResponseWriter, request *http.Request, err error, thisApi any) {
	if apiErr, ok := err.(typed.ApiError); ok {
		writer.WriteHeader(apiErr.StatusCode())
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}
	enc := json.NewEncoder(writer)
	_ = enc.Encode(map[string]any{
		"$error": err.Error(),
	})
}
