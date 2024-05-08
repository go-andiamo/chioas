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
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
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
          description: OK
          content:
            "application/json":
              schema:
                type: object
`
	assert.Equal(t, expectYaml, string(body))
}

func TestDocOptions_Json(t *testing.T) {
	router := chi.NewRouter()
	specJson.SetupRoutes(router)

	req, _ := http.NewRequest(http.MethodGet, "/apidocs", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/apidocs/index.html", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
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
	err := d.SetupRoutes(&apiDefYaml, router)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/docs.htm", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html>`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.True(t, strings.Contains(string(body), `openApi: "spec.yaml",`))

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecName, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
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
          description: OK
          content:
            "application/json":
              schema:
                type: object
`
	assert.Equal(t, expectYaml, string(body))
}

type testSupportFiles struct {
	requests []string
}

func (sf *testSupportFiles) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sf.requests = append(sf.requests, r.URL.Path)
	w.WriteHeader(http.StatusPaymentRequired)
}

func TestDocOptions_SupportFiles(t *testing.T) {
	sf := &testSupportFiles{}
	d := &DocOptions{
		ServeDocs:    true,
		SupportFiles: sf,
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(&apiDefYaml, router)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultIndexName, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html>`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.True(t, strings.Contains(string(body), `openApi: "spec.yaml",`))

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecName, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)

	assert.Equal(t, 0, len(sf.requests))
	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/some_styling.css", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)
	assert.Equal(t, 1, len(sf.requests))
	assert.Equal(t, defaultDocsPath+"/some_styling.css", sf.requests[0])
}

func TestDocOptions_SupportFilesStripPrefix(t *testing.T) {
	sf := &testSupportFiles{}
	d := &DocOptions{
		ServeDocs:               true,
		SupportFiles:            sf,
		SupportFilesStripPrefix: true,
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(&apiDefYaml, router)
	require.NoError(t, err)

	assert.Equal(t, 0, len(sf.requests))
	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath+"/some_styling.css", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)
	assert.Equal(t, 1, len(sf.requests))
	assert.Equal(t, "some_styling.css", sf.requests[0])
}

func TestDocOptions_NoCache_Json(t *testing.T) {
	d := &DocOptions{
		ServeDocs:    true,
		NoCache:      true,
		AsJson:       true,
		DocIndexPage: "docs.htm",
	}
	router := chi.NewRouter()
	err := d.SetupRoutes(&apiDefJson, router)
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusMovedPermanently, res.Result().StatusCode)

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/docs.htm", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	const expectHtmlStarts = `<html>`
	assert.True(t, strings.HasPrefix(string(body), expectHtmlStarts))
	assert.True(t, strings.Contains(string(body), `openApi: "spec.json",`))

	req, _ = http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecNameJson, nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
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
	err := d.SetupRoutes(nil, router)
	assert.Error(t, err)
}

func TestDocOptions_ErrorsWithCache_BadTemplate(t *testing.T) {
	d := DocOptions{
		ServeDocs:   true,
		DocTemplate: `<input type="text"</input>`,
	}
	def := &Definition{}
	router := chi.NewRouter()
	err := d.SetupRoutes(def, router)
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
	err := d.SetupRoutes(def, router)
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
	err := d.SetupRoutes(def, router)
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
			RedocOptions: &RedocOptions{
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
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

func TestDocOptions_RedocOptions_Custom(t *testing.T) {
	type CustomRedocOptions struct {
		ScrollYOffset string `json:"scrollYOffset"`
	}
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			UIStyle:   Redoc,
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
			RedocOptions: &CustomRedocOptions{
				ScrollYOffset: "200px",
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	data := res.Body.Bytes()
	const expect = `<html>
<body>
  <script>
    initTry({
      openApi: "spec.yaml",
      redocOptions: {"scrollYOffset":"200px"}
    })
  </script>
</body>
</html>`
	assert.Equal(t, expect, string(data))
}

func TestDocOptions_SwaggerOptions(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			UIStyle:   Swagger,
			SwaggerOptions: &SwaggerOptions{
				DeepLinking: true,
				Plugins:     []SwaggerPlugin{"MyPlugin1", "MyPlugin2"},
				Presets:     []SwaggerPreset{"MyPreset1", "MyPreset2"},
				HeaderHtml:  `<div>HEADER</div>`,
				HeadScript:  `head();`,
				BodyScript:  `body();`,
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	data := string(res.Body.Bytes())
	assert.Contains(t, data, `"deepLinking":true`)
	assert.Contains(t, data, `"dom_id":"#swagger-ui"`)
	assert.Contains(t, data, `"url":"spec.yaml"`)
	assert.Contains(t, data, `cfg.presets = [MyPreset1,MyPreset2]`)
	assert.Contains(t, data, `cfg.plugins = [MyPlugin1,MyPlugin2]`)
	assert.Contains(t, data, `<div>HEADER</div>`)
	assert.Contains(t, data, `<script>head();</script>`)
	assert.Contains(t, data, `<script>body();</script>`)

	expectedSupportFiles := map[string]string{
		"favicon-16x16.png":               "image/png",
		"favicon-32x32.png":               "image/png",
		"index.css":                       "text/css; charset=utf-8",
		"oauth2-redirect.html":            "text/html; charset=utf-8",
		"swagger-initializer.js":          "application/javascript",
		"swagger-ui.css":                  "text/css; charset=utf-8",
		"swagger-ui.js":                   "application/javascript",
		"swagger-ui-bundle.js":            "application/javascript",
		"swagger-ui-es-bundle.js":         "application/javascript",
		"swagger-ui-es-bundle-core.js":    "application/javascript",
		"swagger-ui-standalone-preset.js": "application/javascript",
	}
	for filename, contentType := range expectedSupportFiles {
		t.Run(filename, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/docs/"+filename, nil)
			assert.NoError(t, err)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			assert.Equal(t, http.StatusOK, res.Result().StatusCode)
			assert.Equal(t, contentType, res.Result().Header.Get(hdrContentType))
		})
	}

	req, err = http.NewRequest(http.MethodGet, "/docs/foo.txt", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}

func TestDocOptions_SwaggerOptions_DefaultPresets(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			UIStyle:   Swagger,
			SwaggerOptions: &SwaggerOptions{
				DeepLinking: true,
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	data := string(res.Body.Bytes())
	assert.Contains(t, data, `"deepLinking":true`)
	assert.Contains(t, data, `"dom_id":"#swagger-ui"`)
	assert.Contains(t, data, `"url":"spec.yaml"`)
	assert.Contains(t, data, `cfg.presets = [SwaggerUIBundle.presets.apis,SwaggerUIStandalonePreset]`)
}

func TestDocOptions_Swagger_WithSupportFiles(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs:    true,
			UIStyle:      Swagger,
			SupportFiles: &testSupportFiles{},
		},
	}
	router := chi.NewRouter()
	err := def.SetupRoutes(router, nil)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/docs/foo.txt", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)
}

func TestDocOptions_GetSwaggerOptions_Custom(t *testing.T) {
	m, presets, plugins := getSwaggerOptions(map[string]any{"foo": "bar"}, "test.yaml")
	assert.Equal(t, 3, len(m))
	assert.Equal(t, "bar", m["foo"])
	assert.Equal(t, "test.yaml", m["url"])
	assert.Equal(t, "#swagger-ui", m["dom_id"])
	assert.Equal(t, "cfg.presets = [SwaggerUIBundle.presets.apis,SwaggerUIStandalonePreset]", string(presets))
	assert.Equal(t, "", string(plugins))

	type CustomSwaggerOptions struct {
		Url   string `json:"url"`
		DomId string `json:"dom_id"`
		Foo   string `json:"foo"`
	}
	o := &CustomSwaggerOptions{
		Url:   "gets overridden",
		DomId: "#my-id",
		Foo:   "bar",
	}
	m, presets, plugins = getSwaggerOptions(o, "test.yaml")
	assert.Equal(t, 3, len(m))
	assert.Equal(t, "bar", m["foo"])
	assert.Equal(t, "test.yaml", m["url"])
	assert.Equal(t, "#my-id", m["dom_id"])
	assert.Equal(t, "cfg.presets = [SwaggerUIBundle.presets.apis,SwaggerUIStandalonePreset]", string(presets))
	assert.Equal(t, "", string(plugins))
}

func TestDocOptions_RapidocOptions(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs: true,
			UIStyle:   Rapidoc,
			RapidocOptions: &RapidocOptions{
				ShowHeader:  true,
				HeadingText: "this-test",
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
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	assert.Equal(t, contentTypeHtml, res.Result().Header.Get(hdrContentType))
	data := string(res.Body.Bytes())
	assert.Contains(t, data, `heading-text="this-test"`)
	assert.Contains(t, data, `show-header="true"`)

	expectedSupportFiles := map[string]string{
		"default.min.css":  "text/css; charset=utf-8",
		"highlight.min.js": "application/javascript",
		"rapidoc-min.js":   "application/javascript",
	}
	for filename, contentType := range expectedSupportFiles {
		t.Run(filename, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/docs/"+filename, nil)
			assert.NoError(t, err)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			assert.Equal(t, http.StatusOK, res.Result().StatusCode)
			assert.Equal(t, contentType, res.Result().Header.Get(hdrContentType))
		})
	}

	req, err = http.NewRequest(http.MethodGet, "/docs/foo.txt", nil)
	assert.NoError(t, err)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}

func TestDocOptions_Rapidoc_WithSupportFiles(t *testing.T) {
	def := Definition{
		DocOptions: DocOptions{
			ServeDocs:    true,
			UIStyle:      Rapidoc,
			SupportFiles: &testSupportFiles{},
		},
	}
	router := chi.NewRouter()
	err := def.SetupRoutes(router, nil)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/docs/foo.txt", nil)
	assert.NoError(t, err)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusPaymentRequired, res.Result().StatusCode)
}
