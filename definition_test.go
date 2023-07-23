package chioas

import (
	"bufio"
	"bytes"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefinition_SetupRoutes(t *testing.T) {
	api := &dummyApi{
		calls: map[string]int{},
	}
	d := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			Context:   "svc",
		},
		Methods: Methods{
			http.MethodGet: {
				Handler: "GetRoot",
			},
		},
		Paths: Paths{
			"/subs": {
				Methods: Methods{
					http.MethodGet: {
						Handler: "GetSubs",
					},
				},
				Paths: Paths{
					"/subsubs": {
						Methods: Methods{
							http.MethodGet: {
								Handler: "GetSubSubs",
							},
						},
					},
				},
			},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, api)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodGet, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodGet, "/subs/subsubs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	assert.Equal(t, 1, api.calls["/"])
	assert.Equal(t, 1, api.calls["/subs"])
	assert.Equal(t, 1, api.calls["/subs/subsubs"])

	req, err = http.NewRequest(http.MethodGet, "/docs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	req, err = http.NewRequest(http.MethodGet, "/docs/index.html", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodGet, "/docs/spec.yaml", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestDefinition_SetupRoutes_AutoHeads(t *testing.T) {
	d := Definition{
		AutoHeadMethods: true,
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
				},
			},
			http.MethodHead: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusConflict)
				},
			},
		},
		Paths: Paths{
			"/subs": {
				Methods: Methods{
					http.MethodGet: {
						Handler: func(writer http.ResponseWriter, request *http.Request) {
							writer.WriteHeader(http.StatusConflict)
						},
					},
				},
			},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)
	req, err = http.NewRequest(http.MethodHead, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusConflict, res.Code)

	req, err = http.NewRequest(http.MethodGet, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusConflict, res.Code)
	req, err = http.NewRequest(http.MethodHead, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusConflict, res.Code)
}

func TestDefinition_SetupRoutes_ErrorsWithBadHandlers(t *testing.T) {
	d := Definition{
		Methods: Methods{
			http.MethodGet: {},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	assert.Error(t, err)

	d = Definition{
		Paths: Paths{
			"/sub": {
				Paths: Paths{
					"/subsub": {
						Methods: Methods{
							http.MethodGet: {},
						},
					},
				},
			},
		},
	}
	router = chi.NewRouter()
	err = d.SetupRoutes(router, nil)
	assert.Error(t, err)
}

func TestDefinition_SetupRoutes_ErrorsWithBadDocTemplate(t *testing.T) {
	d := Definition{
		DocOptions: DocOptions{
			ServeDocs:   true,
			DocTemplate: `{{`,
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	assert.Error(t, err)
}

type dummyApi struct {
	calls map[string]int
}

func (d *dummyApi) GetRoot(writer http.ResponseWriter, request *http.Request) {
	d.calls[request.URL.Path] = d.calls[request.URL.Path] + 1
}

func (d *dummyApi) GetSubs(writer http.ResponseWriter, request *http.Request) {
	d.calls[request.URL.Path] = d.calls[request.URL.Path] + 1
}

func (d *dummyApi) GetSubSubs(writer http.ResponseWriter, request *http.Request) {
	d.calls[request.URL.Path] = d.calls[request.URL.Path] + 1
}

func TestDefinition_writeYaml(t *testing.T) {
	d := Definition{
		DocOptions: DocOptions{
			Context: "svc",
		},
		Servers: Servers{
			"/api/v1": {
				Description: "original",
			},
		},
		Info: Info{
			Title:       "test title",
			Description: "test desc",
			Version:     "1.0.1",
		},
		Tags: Tags{
			{
				Name:        "Foo",
				Description: "foo tag",
			},
			{
				Name:        "Subs",
				Description: "subs tag",
			},
		},
		Methods: Methods{
			http.MethodGet: {
				Description: "Root discovery",
			},
		},
		Paths: Paths{
			"/subs": {
				Tag: "Subs",
				Methods: Methods{
					http.MethodGet: {
						Description: "get subs desc",
					},
				},
				Paths: Paths{
					"/{subId: [a-z]*}": {
						PathParams: PathParams{
							"subId": {
								Description: "id of sub",
							},
						},
						Methods: Methods{
							http.MethodGet: {
								Description: "get specific sub",
							},
						},
						Paths: Paths{
							"/subitems": {
								Paths: Paths{
									"/{subitemId}": {
										Methods: Methods{
											http.MethodGet: {
												Description: "get specific sub-item of sub",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Components: &Components{
			Schemas: Schemas{
				{
					Name:               "fooReq",
					Description:        "foo desc",
					RequiredProperties: []string{"foo"},
					Properties: []Property{
						{
							Name: "foo",
						},
					},
				},
			},
		},
	}
	w := yaml.NewWriter(nil)
	err := d.writeYaml(w)
	assert.NoError(t, err)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `openapi: "3.0.3"
info:
  title: "test title"
  description: "test desc"
  version: "1.0.1"
servers:
  - url: "/api/v1"
    description: "original"
tags:
  - name: "Foo"
    description: "foo tag"
  - name: "Subs"
    description: "subs tag"
paths:
  "/svc/":
    get:
      description: "Root discovery"
      responses:
        200:
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
  "/svc/subs":
    get:
      description: "get subs desc"
      tags:
        - "Subs"
      responses:
        200:
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
  "/svc/subs/{subId}":
    get:
      description: "get specific sub"
      tags:
        - "Subs"
      parameters:
        - name: "subId"
          description: "id of sub"
          in: "path"
          required: true
      responses:
        200:
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
  "/svc/subs/{subId}/subitems/{subitemId}":
    get:
      description: "get specific sub-item of sub"
      tags:
        - "Subs"
      parameters:
        - name: "subId"
          description: "id of sub"
          in: "path"
          required: true
        - name: "subitemId"
          in: "path"
          required: true
      responses:
        200:
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
components:
  schemas:
    "fooReq":
      description: "foo desc"
      type: "object"
      required:
        - "foo"
      properties:
        "foo":
          type: "string"
`
	assert.Equal(t, expect, string(data))

	data, err = d.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expect, string(data))

	var buff bytes.Buffer
	bw := bufio.NewWriter(&buff)
	err = d.WriteYaml(bw)
	assert.NoError(t, err)
	err = bw.Flush()
	assert.NoError(t, err)
	data = buff.Bytes()
	assert.Equal(t, expect, string(data))
}

type testAdditional struct {
}

func (t *testAdditional) Write(on any, w yaml.Writer) {
	w.WriteTagValue("foo", "bar")
}

func TestDefinition_writeYaml_WithAdditional(t *testing.T) {
	d := Definition{
		Additional: &testAdditional{},
	}
	w := yaml.NewWriter(nil)
	err := d.writeYaml(w)
	assert.NoError(t, err)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `openapi: "3.0.3"
info:
  title: "API Documentation"
  version: "1.0.0"
paths:
foo: "bar"
`
	assert.Equal(t, expect, string(data))
}
