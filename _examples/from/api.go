package main

import "net/http"

func getRoot(w http.ResponseWriter, r *http.Request) {
	println("getRoot called")
}

type myApi struct {
}

func (m *myApi) GetPets(w http.ResponseWriter, r *http.Request) {
	println("GetPets called")
}

func (m *myApi) PostPets(w http.ResponseWriter, r *http.Request) {
	println("PostPets called")
}

func (m *myApi) GetPet(w http.ResponseWriter, r *http.Request) {
	println("GetPet called")
}

func (m *myApi) PutPet(w http.ResponseWriter, r *http.Request) {
	println("PutPet called")
}
