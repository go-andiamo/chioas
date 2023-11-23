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

func TestDefinition_SetupRoutes_AutoOptions(t *testing.T) {
	d := Definition{
		AutoHeadMethods:    true,
		AutoOptionsMethods: true,
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
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
					http.MethodPost: {
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

	req, err := http.NewRequest(http.MethodOptions, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, HEAD, OPTIONS", res.Result().Header.Get(hdrAllow))

	req, err = http.NewRequest(http.MethodOptions, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, HEAD, POST, OPTIONS", res.Result().Header.Get(hdrAllow))
}

func TestDefinition_SetupRoutes_AutoOptions_WithRootPayload(t *testing.T) {
	d := Definition{
		AutoOptionsMethods:          true,
		OptionsMethodPayloadBuilder: NewRootOptionsMethodPayloadBuilder(),
		DocOptions: DocOptions{
			HideAutoOptionsMethods: true,
		},
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
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
					http.MethodPost: {
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

	req, err := http.NewRequest(http.MethodOptions, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, OPTIONS", res.Result().Header.Get(hdrAllow))
	assert.Equal(t, contentTypeYaml, res.Result().Header.Get(hdrContentType))
	const expectYaml = `openapi: "3.0.3"
info:
  title: "API Documentation"
  version: "1.0.0"
paths:
  "/":
    get:
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
  "/subs":
    get:
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
    post:
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
`
	assert.Equal(t, expectYaml, res.Body.String())

	d.DocOptions.AsJson = true
	req, err = http.NewRequest(http.MethodOptions, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, OPTIONS", res.Result().Header.Get(hdrAllow))
	assert.Equal(t, contentTypeJson, res.Result().Header.Get(hdrContentType))
	assert.Contains(t, res.Body.String(), `"title":"API Documentation"`)

	d.DocOptions.specData = []byte("null")
	req, err = http.NewRequest(http.MethodOptions, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, OPTIONS", res.Result().Header.Get(hdrAllow))
	assert.Equal(t, contentTypeJson, res.Result().Header.Get(hdrContentType))
	assert.Equal(t, "null", res.Body.String())

	req, err = http.NewRequest(http.MethodOptions, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, POST, OPTIONS", res.Result().Header.Get(hdrAllow))
	assert.Equal(t, "", res.Body.String())
}

func TestDefinition_SetupRoutes_PathAutoOptions(t *testing.T) {
	d := Definition{
		AutoHeadMethods:       true,
		RootAutoOptionsMethod: true,
		Paths: Paths{
			"/subs": {
				AutoOptionsMethod: true,
				Methods: Methods{
					http.MethodGet: {
						Handler: func(writer http.ResponseWriter, request *http.Request) {
							writer.WriteHeader(http.StatusConflict)
						},
					},
					http.MethodPost: {
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

	req, err := http.NewRequest(http.MethodOptions, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "OPTIONS", res.Result().Header.Get(hdrAllow))

	req, err = http.NewRequest(http.MethodOptions, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "GET, HEAD, POST, OPTIONS", res.Result().Header.Get(hdrAllow))
}

func TestDefinition_SetupRoutes_AuthMethodNotAllowed(t *testing.T) {
	d := Definition{
		AutoHeadMethods:      true,
		AutoOptionsMethods:   true,
		AutoMethodNotAllowed: true,
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusNotFound)
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
					http.MethodPost: {
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

	req, err := http.NewRequest(http.MethodPut, "/", nil)
	require.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Equal(t, "GET, HEAD, OPTIONS", res.Result().Header.Get(hdrAllow))

	req, err = http.NewRequest(http.MethodPut, "/subs", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Equal(t, "GET, HEAD, POST, OPTIONS", res.Result().Header.Get(hdrAllow))

	hf := d.methodNotAllowedHandler(nil)
	req, err = http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	res = httptest.NewRecorder()
	hf.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMethodNotAllowed, res.Code)
	assert.Equal(t, "OPTIONS", res.Result().Header.Get(hdrAllow))
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

var testSec = SecuritySchemes{
	{
		Name:        "ApiKey",
		Description: "foo",
		Type:        "apiKey",
		In:          "header",
		ParamName:   "X-API-KEY",
	},
	{
		Name:   "MyOauth",
		Type:   "oauth2",
		Scopes: []string{"write:foo", "read:foo"},
	},
}

var testDef = Definition{
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
		SecuritySchemes: testSec,
	},
	Security: testSec,
	Comment:  "this is a test comment\nand so is this",
}

const expectYaml = `#this is a test comment
#and so is this
openapi: "3.0.3"
info:
  title: "test title"
  description: "test desc"
  version: "1.0.1"
servers:
  - url: "/api/v1"
    description: original
tags:
  - name: Foo
    description: "foo tag"
  - name: Subs
    description: "subs tag"
paths:
  "/svc/":
    get:
      description: "Root discovery"
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
  "/svc/subs":
    get:
      description: "get subs desc"
      tags:
        - Subs
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
  "/svc/subs/{subId}":
    get:
      description: "get specific sub"
      tags:
        - Subs
      parameters:
        - name: subId
          description: "id of sub"
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
  "/svc/subs/{subId}/subitems/{subitemId}":
    get:
      description: "get specific sub-item of sub"
      tags:
        - Subs
      parameters:
        - name: subId
          description: "id of sub"
          in: path
          required: true
          schema:
            type: string
        - name: subitemId
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
components:
  schemas:
    "fooReq":
      description: "foo desc"
      type: object
      required:
        - foo
      properties:
        "foo":
          type: string
  securitySchemes:
    ApiKey:
      description: foo
      type: apiKey
      in: header
      name: X-API-KEY
    MyOauth:
      type: oauth2
security:
  - ApiKey: []
  - MyOauth:
    - "write:foo"
    - "read:foo"
`

func TestDefinition_writeYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	err := testDef.writeYaml(w)
	assert.NoError(t, err)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, expectYaml, string(data))

	var buff bytes.Buffer
	bw := bufio.NewWriter(&buff)
	err = testDef.WriteYaml(bw)
	assert.NoError(t, err)
	err = bw.Flush()
	assert.NoError(t, err)
	data = buff.Bytes()
	assert.Equal(t, expectYaml, string(data))
}

func TestDefinition_WriteYaml(t *testing.T) {
	var buff bytes.Buffer
	bw := bufio.NewWriter(&buff)
	err := testDef.WriteYaml(bw)
	assert.NoError(t, err)
	err = bw.Flush()
	assert.NoError(t, err)
	data := buff.Bytes()
	assert.Equal(t, expectYaml, string(data))
}

func TestDefinition_AsYaml(t *testing.T) {
	data, err := testDef.AsYaml()
	assert.NoError(t, err)
	assert.Equal(t, expectYaml, string(data))
}

const expectJson = `{"components":{"schemas":{"fooReq":{"description":"foo desc","properties":{"foo":{"type":"string"}},"required":["foo"],"type":"object"}},"securitySchemes":{"ApiKey":{"description":"foo","in":"header","name":"X-API-KEY","type":"apiKey"},"MyOauth":{"type":"oauth2"}}},"info":{"description":"test desc","title":"test title","version":"1.0.1"},"openapi":"3.0.3","paths":{"/svc/":{"get":{"description":"Root discovery","responses":{"200":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"OK"}}}},"/svc/subs":{"get":{"description":"get subs desc","responses":{"200":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"OK"}},"tags":["Subs"]}},"/svc/subs/{subId}":{"get":{"description":"get specific sub","parameters":[{"description":"id of sub","in":"path","name":"subId","required":"true","schema":{"type":"string"}}],"responses":{"200":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"OK"}},"tags":["Subs"]}},"/svc/subs/{subId}/subitems/{subitemId}":{"get":{"description":"get specific sub-item of sub","parameters":[{"description":"id of sub","in":"path","name":"subId","required":"true","schema":{"type":"string"}},{"in":"path","name":"subitemId","required":"true","schema":{"type":"string"}}],"responses":{"200":{"content":{"application/json":{"schema":{"type":"object"}}},"description":"OK"}},"tags":["Subs"]}}},"security":[{"ApiKey":[]},{"MyOauth":["write:foo","read:foo"]}],"servers":[{"description":"original","url":"/api/v1"}],"tags":[{"description":"foo tag","name":"Foo"},{"description":"subs tag","name":"Subs"}]}`

func TestDefinition_AsJson(t *testing.T) {
	data, err := testDef.AsJson()
	assert.NoError(t, err)
	assert.Equal(t, expectJson, string(data))
}

func TestDefinition_WriteJson(t *testing.T) {
	var buff bytes.Buffer
	bw := bufio.NewWriter(&buff)
	err := testDef.WriteJson(bw)
	assert.NoError(t, err)
	err = bw.Flush()
	assert.NoError(t, err)
	data := buff.Bytes()
	assert.Equal(t, expectJson, string(data))
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
foo: bar
`
	assert.Equal(t, expect, string(data))
}

func TestDefinition_Middlewares(t *testing.T) {
	d := Definition{
		Middlewares: []func(http.Handler) http.Handler{testUnAuthPostMiddleware},
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {},
			},
			http.MethodPost: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {},
			},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestDefinition_ApplyMiddlewares(t *testing.T) {
	d := Definition{
		ApplyMiddlewares: func(thisApi any) chi.Middlewares {
			return chi.Middlewares{testUnAuthPostMiddleware}
		},
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {},
			},
			http.MethodPost: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {},
			},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodPost, "/", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestDefinition_Path_ApplyMiddlewares(t *testing.T) {
	d := Definition{
		Paths: Paths{
			"/foo": {
				ApplyMiddlewares: func(thisApi any) chi.Middlewares {
					return chi.Middlewares{testUnAuthPostMiddleware}
				},
				Methods: Methods{
					http.MethodGet: {
						Handler: func(writer http.ResponseWriter, request *http.Request) {},
					},
					http.MethodPost: {
						Handler: func(writer http.ResponseWriter, request *http.Request) {},
					},
				},
			},
		},
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(router, nil)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/foo", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req, err = http.NewRequest(http.MethodPost, "/foo", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func testUnAuthPostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusUnauthorized)
		}
		next.ServeHTTP(w, r)
	})
}
