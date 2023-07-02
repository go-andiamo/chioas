package chioas

import (
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

var spec = &apiSpec{
	Definition: apiDef,
}

func (a *apiSpec) SetupRoutes(r chi.Router) {
	a.Definition.This = a
	a.Definition.SetupRoutes(r)
}

var apiDef = Definition{
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
	Paths: Paths{
		"/foo": {
			Methods: Methods{
				http.MethodGet: {
					Handler: func(writer http.ResponseWriter, request *http.Request) {},
				},
			},
		},
	},
}

func TestDocOptions(t *testing.T) {
	router := chi.NewRouter()
	spec.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/apidocs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/index.htm", nil)
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
	assert.True(t, strings.Contains(string(body), `openApi: "myspec.yaml",`))
	assert.True(t, strings.Contains(string(body), `redocOptions: {"scrollYOffset":0,"showObjectSchemaExamples":true,"theme":{"rightPanel":{"width":"60%"},"sidebar":{"width":"220px"}}}`))

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/myspec.yaml", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectYaml = `openapi: "3.0.3"
info:
paths:
  "/foo":
    get:
`
	assert.Equal(t, expectYaml, string(body))
}

func TestDocOptions_PanicsWithBadTemplate(t *testing.T) {
	d := DocOptions{
		ServeDocs:   true,
		DocTemplate: `<html>{{badFunc}}</html>`,
	}
	router := chi.NewRouter()
	assert.Panics(t, func() {
		d.setupRoutes(nil, router)
	})
}
