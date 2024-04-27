package chioas

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAlternateUIDocs(t *testing.T) {
	router := chi.NewRouter()
	multiYaml.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/docs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/docs/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/docs/spec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectYaml = `openapi: "3.0.3"
info:
  title: "API Documentation"
  version: "1.0.0"
paths:
  "/foo":
    get:
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
`
	assert.Equal(t, expectYaml, string(body))

	// alt redoc path...
	req, _ = http.NewRequest(http.MethodGet, "/redoc", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/redoc/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/redoc/spec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, expectYaml, string(body))

	// alt swagger path...
	req, _ = http.NewRequest(http.MethodGet, "/swagger", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/swagger/spec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, expectYaml, string(body))

	// alt rapidoc path...
	req, _ = http.NewRequest(http.MethodGet, "/rapidoc", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/rapidoc/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))

	req, _ = http.NewRequest(http.MethodGet, "/rapidoc/spec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, expectYaml, string(body))
}

func TestAlternateUIDocs_NamedSpec(t *testing.T) {
	router := chi.NewRouter()
	alt := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"swagger": {
					UIStyle:  Swagger,
					SpecName: "myspec.yaml",
				},
			},
		},
	}
	err := alt.DocOptions.SetupRoutes(alt, router)
	require.NoError(t, err)
	req, _ := http.NewRequest(http.MethodGet, "/swagger/myspec.yaml", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func TestAlternateUIDocs_AsJson(t *testing.T) {
	router := chi.NewRouter()
	alt := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"swagger": {
					UIStyle: Swagger,
					AsJson:  true,
				},
			},
		},
	}
	err := alt.DocOptions.SetupRoutes(alt, router)
	require.NoError(t, err)
	req, _ := http.NewRequest(http.MethodGet, "/swagger/spec.json", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func TestAlternateUIDocs_ErrorsWithDuplicatePath(t *testing.T) {
	router := chi.NewRouter()
	bad := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"docs": {},
			},
		},
	}
	err := bad.DocOptions.SetupRoutes(bad, router)
	assert.Error(t, err)
}

func TestAlternateUIDocs_NoCache(t *testing.T) {
	router := chi.NewRouter()
	bad := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"alt": {
					NoCache: true,
				},
			},
		},
	}
	err := bad.DocOptions.SetupRoutes(bad, router)
	assert.NoError(t, err)
}

func TestAlternateUIDocs_ErrorsWithBadTemplate(t *testing.T) {
	router := chi.NewRouter()
	bad := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"errors": {
					DocTemplate: `<html>{{badFunc}}</html>`,
				},
			},
		},
	}
	err := bad.DocOptions.SetupRoutes(bad, router)
	assert.Error(t, err)
}

func TestAlternateUIDocs_ErrorsWithCache(t *testing.T) {
	router := chi.NewRouter()
	bad := &Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			AlternateUIDocs: AlternateUIDocs{
				"errors": {
					DocTemplate: `<input type="text"</input>`,
				},
			},
		},
	}
	err := bad.DocOptions.SetupRoutes(bad, router)
	assert.Error(t, err)
}

var multiYaml = &apiSpec{
	Definition: apiMultiDefYaml,
}

var apiMultiDefYaml = Definition{
	DocOptions: DocOptions{
		ServeDocs:   true,
		Title:       "My API Docs",
		DocTemplate: testHtml,
		UIStyle:     Redoc,
		AlternateUIDocs: AlternateUIDocs{
			"/redoc": {
				UIStyle: Redoc,
			},
			"/swagger": {
				UIStyle: Swagger,
			},
			"/rapidoc": {
				UIStyle: Rapidoc,
			},
		},
	},
	Paths: defPaths,
}
