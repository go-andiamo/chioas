package main

import (
	"encoding/json"
	"github.com/go-andiamo/chioas"
	"net/http"
)

var petsPaths = chioas.Path{
	Tag: "pet",
	Methods: map[string]chioas.Method{
		http.MethodGet: {
			Comment:     chioas.SourceComment(commenter("GetPets")...),
			Handler:     "GetPets",
			Summary:     "List/query pets",
			Description: "List/query pets",
			OperationId: "getPets",
			QueryParams: []chioas.QueryParam{
				{
					Name:        "status",
					Description: "Status values that need to be considered for filter",
					SchemaRef:   "Status",
				},
			},
			Responses: map[int]chioas.Response{
				http.StatusOK: {
					Description: "Resultant pets",
					IsArray:     true,
					SchemaRef:   "Pet",
				},
			},
		},
		http.MethodPost: {
			Comment:     chioas.SourceComment(commenter("PostPets")...),
			Handler:     "PostPets",
			Summary:     "Add pet",
			Description: "Add pet",
			OperationId: "addPet",
			Request: &chioas.Request{
				Description: "Pet to be added to the store",
				SchemaRef:   "Pet",
			},
			Responses: map[int]chioas.Response{
				http.StatusOK: {
					Description: "Created pet",
					IsArray:     true,
					SchemaRef:   "Pet",
				},
			},
		},
	},
	Paths: map[string]chioas.Path{
		"/{petId}": {
			PathParams: map[string]chioas.PathParam{
				"petId": {
					Description: "id of the Pet",
				},
			},
			Methods: chioas.Methods{
				http.MethodGet: {
					Comment:     chioas.SourceComment(commenter("GetPet")...),
					Handler:     "GetPet",
					Summary:     "Get an existing pet",
					Description: "Get an existing pet by Id",
					OperationId: "getPet",
					Responses: chioas.Responses{
						http.StatusOK: {
							Description: "Successful operation",
							SchemaRef:   "Pet",
						},
					},
				},
				http.MethodPut: {
					Comment:     chioas.SourceComment(commenter("UpdatePet")...),
					Handler:     "UpdatePet",
					Summary:     "Update an existing pet",
					Description: "Update an existing pet by Id",
					OperationId: "updatePet",
					Request: &chioas.Request{
						Description: "Update an existent pet in the store",
						SchemaRef:   "Pet",
					},
					Responses: chioas.Responses{
						http.StatusOK: {
							Description: "Successful operation",
							SchemaRef:   "Pet",
						},
					},
				},
			},
		},
	},
}

func (d *api) GetPets(writer http.ResponseWriter, request *http.Request) {
	res := map[string]any{
		"hello": "you listed/queried pets",
	}
	enc := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	_ = enc.Encode(res)
}

func (d *api) PostPets(writer http.ResponseWriter, request *http.Request) {
	res := map[string]any{
		"hello": "you added a pet",
	}
	enc := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	_ = enc.Encode(res)
}

func (d *api) GetPet(writer http.ResponseWriter, request *http.Request) {
	res := map[string]any{
		"hello": "you retrieved a pet",
	}
	enc := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	_ = enc.Encode(res)
}

func (d *api) UpdatePet(writer http.ResponseWriter, request *http.Request) {
	res := map[string]any{
		"hello": "you updated a pet",
	}
	enc := json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")
	_ = enc.Encode(res)
}
