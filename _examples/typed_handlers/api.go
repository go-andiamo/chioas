package main

import (
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/typed"
	"github.com/go-chi/chi/v5"
	"net/http"
)

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
func (a *api) GetPeople() ([]Person, error) {
	return dummyPeopleStore, nil
}

// AddPerson is the typed handler for POST /person - note the typed input and return args
func (a *api) AddPerson(person *Person) (Person, error) {
	person.Id = len(dummyPeopleStore)
	dummyPeopleStore = append(dummyPeopleStore, *person)
	return *person, nil
}

// GetPerson is the typed handler for GET /person/{personId}
//
// Note the arg type PersonId - which is extracted from the path params by the PersonIdArgBuilder
func (a *api) GetPerson(personId PersonId) (Person, error) {
	if personId >= 0 && int(personId) < len(dummyPeopleStore) {
		return dummyPeopleStore[personId], nil
	}
	return Person{}, typed.NewApiErrorf(http.StatusNotFound, "not found person id %d", personId)
}
