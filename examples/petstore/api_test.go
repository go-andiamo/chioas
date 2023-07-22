package main

import (
	"bufio"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestWriteYaml(t *testing.T) {
	fo, err := os.Create("petstore.yaml")
	require.NoError(t, err)
	w := bufio.NewWriter(fo)

	err = petStoreApi.WriteYaml(w)
	assert.NoError(t, err)
}

func TestDocs(t *testing.T) {
	router := chi.NewRouter()
	testApi := &api{
		Definition: apiDef,
	}
	err := testApi.SetupRoutes(router)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/docs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "/docs/index.htm", res.Header().Get("Location"))

	req, _ = http.NewRequest(http.MethodGet, "/docs/index.htm", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "text/html; charset=utf-8", res.Header().Get("Content-Type"))

	req, _ = http.NewRequest(http.MethodGet, "/docs/spec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "text/plain; charset=utf-8", res.Header().Get("Content-Type"))
}
