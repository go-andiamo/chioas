package chioas

import (
	"errors"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testHtml = `<html><head>
        <title>{{.title}}</title>
		<style>{{.stylesOverride}}</style>
	</head>
	<body>
        <div id="redoc-container"></div>
        <script>
            initTry({
                openApi: {{.specName}},
                redocOptions: {{.redocopts}}
            })
        </script>
	</body>
</html>`

type apiSpec struct {
	Definition
}

var specYaml = &apiSpec{
	Definition: apiDefYaml,
}

var specJson = &apiSpec{
	Definition: apiDefJson,
}

func (a *apiSpec) SetupRoutes(r chi.Router) {
	err := a.Definition.SetupRoutes(r, a)
	if err != nil {
		panic(err)
	}
}

var defPaths = Paths{
	"/foo": {
		Methods: Methods{
			http.MethodGet: {
				Handler: func(writer http.ResponseWriter, request *http.Request) {},
			},
		},
	},
}

var apiDefYaml = Definition{
	DocOptions: DocOptions{
		ServeDocs:   true,
		Title:       "My API Docs",
		DocTemplate: testHtml,
		SpecName:    "myspec.yaml",
		Path:        "/apidocs",
		RedocOptions: map[string]any{
			"scrollYOffset":            0,
			"showObjectSchemaExamples": true,
			"theme": map[string]any{
				"sidebar": map[string]any{
					"width": "220px",
				},
				"rightPanel": map[string]any{
					"width": "60%",
				},
			},
		},
	},
	Paths: defPaths,
}

var apiDefJson = Definition{
	DocOptions: DocOptions{
		ServeDocs:   true,
		Title:       "My API Docs",
		DocTemplate: testHtml,
		AsJson:      true,
		SpecName:    "myspec.json",
		Path:        "/apidocs",
		RedocOptions: map[string]any{
			"scrollYOffset":            0,
			"showObjectSchemaExamples": true,
			"theme": map[string]any{
				"sidebar": map[string]any{
					"width": "220px",
				},
				"rightPanel": map[string]any{
					"width": "60%",
				},
			},
		},
	},
	Paths: defPaths,
}

func TestDocOptions(t *testing.T) {
	router := chi.NewRouter()
	specYaml.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/apidocs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html><head>
        <title>My API Docs</title>
		<style>body {
`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.Contains(t, string(body), `openApi: "myspec.yaml",`)
	assert.Contains(t, string(body), `redocOptions: {"scrollYOffset":0,"showObjectSchemaExamples":true,"theme":{"rightPanel":{"width":"60%"},"sidebar":{"width":"220px"}}}`)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/myspec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err = io.ReadAll(res.Body)
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
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
`
	assert.Equal(t, expectYaml, string(body))
}

func TestDocOptions_Json(t *testing.T) {
	router := chi.NewRouter()
	specJson.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/apidocs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html><head>
        <title>My API Docs</title>
		<style>body {
`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.Contains(t, string(body), `openApi: "myspec.json",`)
	assert.Contains(t, string(body), `redocOptions: {"scrollYOffset":0,"showObjectSchemaExamples":true,"theme":{"rightPanel":{"width":"60%"},"sidebar":{"width":"220px"}}}`)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/myspec.json", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"paths":{"/foo":{"get":{`)
}

func TestDocOptions_NoCache(t *testing.T) {
	d := &DocOptions{
		ServeDocs:    true,
		NoCache:      true,
		DocIndexPage: "docs.htm",
	}
	router := chi.NewRouter()
	err := d.setupRoutes(&apiDefYaml, router)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/docs.htm", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html>`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.True(t, strings.Contains(string(body), `openApi: "spec.yaml",`))

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecName, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err = io.ReadAll(res.Body)
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
          description: "OK"
          content:
            application/json:
              schema:
                type: "object"
`
	assert.Equal(t, expectYaml, string(body))
}

func TestDocOptions_NoCache_Json(t *testing.T) {
	d := &DocOptions{
		ServeDocs:    true,
		NoCache:      true,
		AsJson:       true,
		DocIndexPage: "docs.htm",
	}
	router := chi.NewRouter()
	err := d.setupRoutes(&apiDefJson, router)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/docs.htm", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html>`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.True(t, strings.Contains(string(body), `openApi: "spec.json",`))

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecNameJson, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"paths":{"/foo":{"get":{`)
}

func TestDocOptions_ErrorsWithBadTemplate(t *testing.T) {
	d := DocOptions{
		ServeDocs:   true,
		DocTemplate: `<html>{{badFunc}}</html>`,
	}
	router := chi.NewRouter()
	err := d.setupRoutes(nil, router)
	assert.Error(t, err)
}

func TestDocOptions_ErrorsWithCache_BadTemplate(t *testing.T) {
	d := DocOptions{
		ServeDocs:   true,
		DocTemplate: `<input type="text"</input>`,
	}
	def := &Definition{}
	router := chi.NewRouter()
	err := d.setupRoutes(def, router)
	assert.Error(t, err)
}

func TestDocOptions_ErrorsWithCache_BadYaml(t *testing.T) {
	d := DocOptions{
		ServeDocs: true,
	}
	def := &Definition{
		Additional: &errorAdditional{},
	}
	router := chi.NewRouter()
	err := d.setupRoutes(def, router)
	assert.Error(t, err)
}

func TestDocOptions_ErrorsWithCache_BadJson(t *testing.T) {
	d := DocOptions{
		ServeDocs: true,
		AsJson:    true,
	}
	def := &Definition{
		Additional: &errorAdditional{},
	}
	router := chi.NewRouter()
	err := d.setupRoutes(def, router)
	assert.Error(t, err)
}

type errorAdditional struct {
}

func (e errorAdditional) Write(on any, w yaml.Writer) {
	w.SetError(errors.New("foo"))
}

func TestDocOptions_RedocOptions(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			DocTemplate: `<html>
<body>
  <script>
    initTry({
      openApi: {{.specName}},
      redocOptions: {{.redocopts}}
    })
  </script>
</body>
</html>`,
			RedocOptions: RedocOptions{
				ScrollYOffset: "200px",
				Theme: &RedocTheme{
					Typography: &RedocTypography{
						FontFamily: "times",
					},
				},
			},
		},
	}
	router := chi.NewRouter()
	err := def.SetupRoutes(router, nil)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/docs/index.html", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	data := res.Body.Bytes()
	const expect = `<html>
<body>
  <script>
    initTry({
      openApi: "spec.yaml",
      redocOptions: {"scrollYOffset":"200px","theme":{"typography":{"fontFamily":"times","optimizeSpeed":false}}}
    })
  </script>
</body>
</html>`
	assert.Equal(t, expect, string(data))
}
