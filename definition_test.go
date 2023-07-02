package chioas

import (
	"bufio"
	"bytes"
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
		This:    api,
		Context: "svc",
		Methods: Methods{
			http.MethodGet: {
				MethodName: "GetRoot",
			},
		},
		Paths: Paths{
			"/subs": {
				Methods: Methods{
					http.MethodGet: {
						MethodName: "GetSubs",
					},
				},
				Paths: Paths{
					"/subsubs": {
						Methods: Methods{
							http.MethodGet: {
								MethodName: "GetSubSubs",
							},
						},
					},
				},
			},
		},
	}
	router := chi.NewRouter()
	d.SetupRoutes(router)

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
		Context:     "svc",
		Title:       "test title",
		Description: "test desc",
		Version:     "1.0.1",
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
					"/subsubs": {
						Methods: Methods{
							http.MethodGet: {
								Description: "get subs subs desc",
							},
						},
					},
				},
			},
		},
	}
	w := newYamlWriter(nil)
	err := d.writeYaml(w)
	require.NoError(t, err)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `openapi: "3.0.3"
info:
  title: "test title"
  description: "test desc"
  version: "1.0.1"
tags:
  - name: "Foo"
    description: "foo tag"
  - name: "Subs"
    description: "subs tag"
paths:
  "/svc/":
    get:
      description: "Root discovery"
  "/svc/subs":
    get:
      description: "get subs desc"
      tags:
        - "Subs"
  "/svc/subs/subsubs":
    get:
      description: "get subs subs desc"
      tags:
        - "Subs"
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
