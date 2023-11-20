package typed

import (
	"encoding/json"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/typed"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var router *chi.Mux

func init() {
	router = chi.NewRouter()
	const traditionalPath = "/traditional/{id}"
	router.Get(traditionalPath, traditionalGet)
	router.Put(traditionalPath, traditionalPut)

	const typedPath = "/typed/{id}"
	mb := typed.NewTypedMethodsHandlerBuilder()
	hf, _ := mb.BuildHandler(typedPath, http.MethodGet, chioas.Method{Handler: typedGet}, nil)
	router.Get(typedPath, hf)
	hf, _ = mb.BuildHandler(typedPath, http.MethodPut, chioas.Method{Handler: typedPut}, nil)
	router.Put(typedPath, hf)
}

const (
	hdrContentType  = "Content-Type"
	contentTypeJson = "application/json"
)

func TestRoutesWork(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/traditional/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, contentTypeJson, w.Result().Header.Get(hdrContentType))
	assert.Equal(t, `{"Id":"1","Name":"test","Age":16}`, w.Body.String())

	r, _ = http.NewRequest(http.MethodGet, "/traditional/0", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)

	r, _ = http.NewRequest(http.MethodPut, "/traditional/16", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, contentTypeJson, w.Result().Header.Get(hdrContentType))
	assert.Equal(t, `{"Id":"16","Name":"Bilbo","Age":99}`, w.Body.String())

	r, _ = http.NewRequest(http.MethodPut, "/traditional/0", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)

	r, _ = http.NewRequest(http.MethodGet, "/typed/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, contentTypeJson, w.Result().Header.Get(hdrContentType))
	assert.Equal(t, `{"Id":"1","Name":"test","Age":16}`, w.Body.String())

	r, _ = http.NewRequest(http.MethodGet, "/typed/0", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)

	r, _ = http.NewRequest(http.MethodPut, "/typed/16", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, contentTypeJson, w.Result().Header.Get(hdrContentType))
	assert.Equal(t, `{"Id":"16","Name":"Bilbo","Age":99}`, w.Body.String())

	r, _ = http.NewRequest(http.MethodPut, "/typed/0", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func BenchmarkTraditional_Get(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodGet, "/traditional/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
	}
}

func BenchmarkTyped_Get(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodGet, "/typed/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
	}
}

func BenchmarkTraditional_Put(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodPut, "/traditional/1", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
	}
}

func BenchmarkTyped_Put(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest(http.MethodPut, "/typed/1", strings.NewReader(`{"Name":"Bilbo","Age":99}`))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
	}
}

type testerResponse struct {
	Id   string
	Name string
	Age  int
}

type testerRequest struct {
	Name string
	Age  int
}

func traditionalGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "0" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	response := &testerResponse{Id: id, Name: "test", Age: 16}
	if data, err := json.Marshal(response); err == nil {
		w.Header().Set(hdrContentType, contentTypeJson)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func typedGet(id string) (*testerResponse, error) {
	if id == "0" {
		return nil, typed.NewApiError(http.StatusNotFound, "")
	}
	return &testerResponse{Id: id, Name: "test", Age: 16}, nil
}

func traditionalPut(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "0" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	request := &testerRequest{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := &testerResponse{Id: id, Name: request.Name, Age: request.Age}
	if data, err := json.Marshal(response); err == nil {
		w.Header().Set(hdrContentType, contentTypeJson)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func typedPut(id string, req *testerRequest) (*testerResponse, error) {
	if id == "0" {
		return nil, typed.NewApiError(http.StatusNotFound, "")
	}
	return &testerResponse{Id: id, Name: req.Name, Age: req.Age}, nil
}
