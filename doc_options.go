package chioas

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas/rapidoc_ui"
	"github.com/go-andiamo/chioas/swagger_ui"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type UIStyle uint

const (
	Redoc   UIStyle = iota // style for Redoc UI
	Swagger                // style for swagger-ui
	Rapidoc                // style for rapidoc-ui
)

// DocOptions determines whether/how the api will serve interactive docs page
type DocOptions struct {
	// ServeDocs whether to serve api docs
	ServeDocs bool
	// Context is the optional path prefix for all paths in OAS spec
	Context string
	// NoCache if set to true, docs page and spec aren't cached (and built on each call)
	NoCache bool
	// AsJson if set to true, serves the spec as JSON
	AsJson bool
	// Path the path on which to serve api docs page and spec (defaults to "/docs")
	Path string
	// DocIndexPage the name of the docs index page (defaults to "index.html")
	DocIndexPage string
	// SupportFiles is an optional handler that is used for other files under "/docs" path
	//
	// see _examples/swagger_ui for example usage
	SupportFiles http.Handler
	// SupportFilesStripPrefix if set to true (and SupportFiles is specified) then
	// calls to SupportFiles have the "/docs" path prefix stripped from the http.Request
	SupportFilesStripPrefix bool
	// Title the title in the docs index page (defaults to "API Documentation")
	Title string
	// UIStyle is the style of the API docs UI
	//
	// use Redoc, Swagger or Rapidoc (defaults to Redoc)
	UIStyle UIStyle
	// AlternateUIDocs allow your docs UI to be served as different styles on different paths
	//
	// where the key for each is the docs path
	AlternateUIDocs AlternateUIDocs
	// RedocOptions redoc options to be used (see https://github.com/Redocly/redoc#redoc-options-object)
	//
	// use map[string]any or &RedocOptions or anything that implements ToMap (or anything that can be marshalled and then unmarshalled to map[string]any)
	//
	// Only used if DocOptions.UIStyle is Redoc
	RedocOptions any
	// SwaggerOptions swagger-ui options to be used (see https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/configuration.md)
	//
	// Only used if DocOptions.UIStyle is Swagger
	SwaggerOptions any
	// RapidocOptions rapidoc options to be used
	//
	// Only used if DocOptions.UIStyle is Rapidoc
	RapidocOptions any
	// SpecName the name of the OAS spec (defaults to "spec.yaml")
	SpecName string
	// DocTemplate template for the docs page (defaults to internal template if an empty string)
	DocTemplate string
	// StylesOverride css styling overrides (injected into docs index page)
	StylesOverride string
	// RedocJsUrl is the URL for the Redoc JS
	//
	// defaults to:
	// https://cdn.jsdelivr.net/npm/redoc@2.0.0-rc.77/bundles/redoc.standalone.min.js
	//
	// Only used if DocOptions.UIStyle is Redoc
	RedocJsUrl string
	// TryJsUrl is the URL for the Try Redoc JS
	//
	// defaults to:
	// https://cdn.jsdelivr.net/gh/wll8/redoc-try@1.4.7/dist/try.js
	//
	// Only used if DocOptions.UIStyle is Redoc
	TryJsUrl string
	// DefaultResponses is the default responses for methods that don't have any responses defined
	//
	// is a map of http status code and response
	DefaultResponses Responses
	// HideHeadMethods indicates that all HEAD methods should be hidden from docs
	HideHeadMethods bool
	// HideAutoOptionsMethods indicates that automatically added OPTIONS methods (see Definition.AutoOptionsMethods) are hidden from docs
	HideAutoOptionsMethods bool
	// OperationIdentifier is an optional function called by Method to generate `operationId` tag value
	OperationIdentifier OperationIdentifier
	// Middlewares is any chi.Middlewares for everything served under '/docs' path
	Middlewares chi.Middlewares
	// CheckRefs when set to true, all internal $ref's are checked
	CheckRefs bool
	// specData is used internally where api has been generated from spec (see FromJson and FromYaml)
	specData []byte
}

type SupportFiles interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// OperationIdentifier is a function that can be provided to DocOptions
type OperationIdentifier func(method Method, methodName string, path string, parentTag string) string

const (
	defaultDocsPath     = "/docs"
	defaultIndexName    = "index.html"
	defaultSpecName     = "spec.yaml"
	defaultSpecNameJson = "spec.json"
	defaultTitle        = "API Documentation"
	defaultRedocJsUrl   = "https://cdn.jsdelivr.net/npm/redoc@2.0.0-rc.77/bundles/redoc.standalone.min.js"
	defaultTryJsUrl     = "https://cdn.jsdelivr.net/gh/wll8/redoc-try@1.4.7/dist/try.js"
)

func (d *DocOptions) SetupRoutes(def *Definition, route chi.Router) error {
	if d.ServeDocs {
		tmp, err := getTemplate(d.UIStyle, d.DocTemplate)
		if err != nil {
			return err
		}
		path := defValue(d.Path, defaultDocsPath)
		indexPage := defValue(d.DocIndexPage, defaultIndexName)
		specName, data := d.getTemplateData()
		docsRoute := chi.NewRouter()
		docsRoute.Use(d.Middlewares...)
		redirectPath := path + root + indexPage
		docsRoute.Get(root, http.RedirectHandler(redirectPath, http.StatusMovedPermanently).ServeHTTP)
		if d.specData != nil || !d.NoCache {
			if err = setupCachedRoutes(def, d.AsJson, d.specData, docsRoute, tmp, data, indexPage, specName); err != nil {
				return err
			}
		} else {
			setupNoCachedRoutes(def, d.AsJson, docsRoute, tmp, data, indexPage, specName)
		}
		setupSupportFiles(defValue(d.Path, defaultDocsPath), d.UIStyle, d.SupportFiles, d.SupportFilesStripPrefix, docsRoute)
		route.Mount(path, docsRoute)
		for altPath, alt := range d.AlternateUIDocs {
			if !strings.HasPrefix(altPath, "/") {
				altPath = "/" + altPath
			}
			if altPath == "/" || altPath == path {
				return fmt.Errorf("invalid aletrnate docs path '%s'", altPath)
			}
			if err = alt.setupRoute(def, route, altPath, d.specData); err != nil {
				return err
			}
		}
	}
	return nil
}

func getTemplate(uiStyle UIStyle, docTemplate string) (*template.Template, error) {
	if docTemplate == "" {
		switch uiStyle {
		case Swagger:
			return template.New("index").Parse(defaultSwaggerTemplate)
		case Rapidoc:
			return template.New("index").Parse(defaultRapidocTemplate)
		default:
			return template.New("index").Parse(defaultRedocTemplate)
		}
	}
	return template.New("index").Parse(docTemplate)
}

func (d *DocOptions) getTemplateData() (specName string, data map[string]any) {
	if d.SpecName != "" {
		specName = d.SpecName
	} else if d.AsJson {
		specName = defaultSpecNameJson
	} else {
		specName = defaultSpecName
	}
	switch d.UIStyle {
	case Swagger:
		swaggerOpts, presets, plugins := getSwaggerOptions(d.SwaggerOptions, specName)
		data = map[string]any{
			htmlTagTitle:          defValue(d.Title, defaultTitle),
			htmlTagStylesOverride: template.CSS(d.StylesOverride),
			htmlTagSwaggerOpts:    swaggerOpts,
			htmlTagSwaggerPresets: presets,
			htmlTagSwaggerPlugins: plugins,
		}
	case Rapidoc:
		data = optionsToMap(d.RapidocOptions)
		data[htmlTagTitle] = defValue(d.Title, defaultTitle)
		data[htmlTagStylesOverride] = template.CSS(d.StylesOverride)
		data[htmlTagSpecName] = specName
	default:
		data = map[string]any{
			htmlTagTitle:          defValue(d.Title, defaultTitle),
			htmlTagStylesOverride: template.CSS(defValue(d.StylesOverride, defaultRedocStylesOverride)),
			htmlTagSpecName:       specName,
			htmlTagRedocOpts:      optionsToMap(d.RedocOptions),
			htmlTagRedocUrl:       defValue(d.RedocJsUrl, defaultRedocJsUrl),
			htmlTagTryUrl:         defValue(d.TryJsUrl, defaultTryJsUrl),
		}
	}
	return
}

func setupCachedRoutes(def *Definition, asJson bool, specData []byte, docsRoute *chi.Mux, tmp *template.Template, inData map[string]any, indexPage, specName string) (err error) {
	indexData, err := buildIndexData(tmp, inData)
	if err != nil {
		return err
	}
	docsRoute.Get(root+indexPage, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentTypeHtml)
		_, _ = writer.Write(indexData)
	})
	var data []byte
	contentType := contentTypeYaml
	if specData != nil {
		data = specData
	} else if asJson {
		contentType = contentTypeJson
		if data, err = def.AsJson(); err != nil {
			return err
		}
	} else {
		if data, err = def.AsYaml(); err != nil {
			return err
		}
	}
	docsRoute.Get(root+specName, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentType)
		_, _ = writer.Write(data)
	})
	return nil
}

func setupNoCachedRoutes(def *Definition, asJson bool, docsRoute *chi.Mux, tmp *template.Template, inData map[string]any, indexPage, specName string) {
	docsRoute.Get(root+indexPage, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentTypeHtml)
		_ = tmp.Execute(writer, inData)
	})
	if asJson {
		docsRoute.Get(root+specName, func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set(hdrContentType, contentTypeJson)
			_ = def.WriteJson(writer)
		})
	} else {
		docsRoute.Get(root+specName, func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set(hdrContentType, contentTypeYaml)
			_ = def.WriteYaml(writer)
		})
	}
}

func setupSupportFiles(path string, uiStyle UIStyle, supportFiles http.Handler, stripPrefix bool, docsRoute *chi.Mux) {
	sf := getSupportFiles(supportFiles, stripPrefix, path)
	switch uiStyle {
	case Swagger:
		docsRoute.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
			name := strings.TrimPrefix(request.URL.Path, path+"/")
			if data, err := swagger_ui.SwaggerUIStaticFiles.ReadFile(name); err == nil {
				if ctype := mime.TypeByExtension(filepath.Ext(name)); ctype != "" {
					writer.Header().Set(hdrContentType, ctype)
				}
				_, _ = writer.Write(data)
				return
			} else if sf != nil {
				sf.ServeHTTP(writer, request)
				return
			}
			writer.WriteHeader(http.StatusNotFound)
		})
	case Rapidoc:
		docsRoute.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
			name := strings.TrimPrefix(request.URL.Path, path+"/")
			if data, err := rapidoc_ui.RapidocUIStaticFiles.ReadFile(name); err == nil {
				if ctype := mime.TypeByExtension(filepath.Ext(name)); ctype != "" {
					writer.Header().Set(hdrContentType, ctype)
				}
				_, _ = writer.Write(data)
				return
			} else if sf != nil {
				sf.ServeHTTP(writer, request)
				return
			}
			writer.WriteHeader(http.StatusNotFound)
		})
	default:
		if sf != nil {
			docsRoute.Get("/*", func(writer http.ResponseWriter, request *http.Request) {
				sf.ServeHTTP(writer, request)
			})
		}
	}
}

func getSupportFiles(supportFiles http.Handler, stripPrefix bool, path string) http.Handler {
	if supportFiles != nil && stripPrefix {
		return http.StripPrefix(path+"/", supportFiles)
	}
	return supportFiles
}

func getSwaggerOptions(swaggerOptions any, specName string) (map[string]any, template.JS, template.JS) {
	m := optionsToMap(swaggerOptions)
	if _, ok := m["dom_id"]; !ok {
		m["dom_id"] = "#swagger-ui"
	}
	m["url"] = specName
	presets := template.JS("")
	plugins := template.JS("")
	switch so := swaggerOptions.(type) {
	case *SwaggerOptions:
		presets, plugins = so.jsPresets(), so.jsPlugins()
	default:
		presets = "cfg.presets = [SwaggerUIBundle.presets.apis,SwaggerUIStandalonePreset]"
	}
	return m, presets, plugins
}

func optionsToMap(opts any) map[string]any {
	m := map[string]any{}
	if opts != nil {
		if o, ok := opts.(MappableOptions); ok {
			return o.ToMap()
		} else {
			switch topt := opts.(type) {
			case map[string]any:
				return topt
			default:
				if data, err := json.Marshal(opts); err == nil {
					var res map[string]any
					if json.Unmarshal(data, &res) == nil {
						return res
					}
				}
			}
		}
	}
	return m
}

func buildIndexData(tmp *template.Template, inData map[string]any) (data []byte, err error) {
	var buffer bytes.Buffer
	w := bufio.NewWriter(&buffer)
	if err = tmp.Execute(w, inData); err == nil {
		if err = w.Flush(); err == nil {
			data = buffer.Bytes()
		}
	}
	return
}

const (
	contentTypeHtml       = "text/html; charset=utf-8"
	contentTypeJson       = "application/json"
	contentTypeYaml       = "application/yaml"
	hdrContentType        = "Content-Type"
	htmlTagTitle          = "title"
	htmlTagStylesOverride = "stylesOverride"
	htmlTagSpecName       = "specName"
	htmlTagRedocOpts      = "redocopts"
	htmlTagRedocUrl       = "redocurl"
	htmlTagTryUrl         = "tryurl"
	htmlTagSwaggerOpts    = "swaggeropts"
	htmlTagSwaggerPresets = "swaggerpresets"
	htmlTagSwaggerPlugins = "swaggerplugins"
)

const defaultRedocTemplate = `<html>
    <head>
        <title>{{.title}}</title>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
        <style>{{.stylesOverride}}</style>
    </head>
    <body>
        <div id="redoc-container"></div>
        <script src="{{.redocurl}}" crossorigin="anonymous"></script>
        <script src="{{.tryurl}}" crossorigin="anonymous"></script>
        <script>
            initTry({
                openApi: {{.specName}},
                redocOptions: {{.redocopts}}
            })
        </script>
    </body>
</html>`

const defaultSwaggerTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{.title}}</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
    <link rel="stylesheet" type="text/css" href="./index.css" />
    <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
    <style>{{.stylesOverride}}</style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="./swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="./swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
      window.onload = function() {
		let cfg = {{.swaggeropts}}
		{{.swaggerpresets}}
		{{.swaggerplugins}}
        const ui = SwaggerUIBundle(cfg)
        window.ui = ui
      }
    </script>
  </body>
</html>`

const defaultRapidocTemplate = `<!doctype html>
<html lang="en">
    <head>
        <title>{{.title}}</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1,user-scalable=yes">
        <link rel="stylesheet" href="default.min.css">
        <script src="highlight.min.js"></script>
        <script defer="defer" src="rapidoc-min.js"></script>
        <style>{{.stylesOverride}}</style>
    </head>
    <body>
        <rapi-doc id="thedoc" heading-text="{{.heading_text}}" spec-url="{{.specName}}" theme="{{.theme}}" render-style="{{.render_style}}" schema-style="{{.schema_style}}" show-method-in-nav-bar="{{.show_method_in_nav_bar}}" use-path-in-nav-bar="{{.use_path_in_nav_bar}}" show-components="{{.show_components}}" show-info="{{.show_info}}" show-header="{{.show_header}}" allow-search="{{.allow_search}}" allow-advanced-search="{{.allow_advanced_search}}" allow-spec-url-load="{{.allow_spec_url_load}}" allow-spec-file-load="{{.allow_spec_file_load}}" allow-try="{{.allow_try}}" allow-spec-file-download="{{.allow_spec_file_download}}" allow-server-selection="{{.allow_server_selection}}" allow-authentication="{{.allow_authentication}}" update-route="{{.update_route}}" match-type="{{.match_type}}" persist_auth="{{.persist_auth}}"></rapi-doc>
    </body>
</html>`

const defaultRedocStylesOverride = `body {
	margin: 0;
	padding: 0;
}
/* restyle Try button */
button.tryBtn {
	margin-right: 0.5em;
	padding: 4px 8px 4px 8px;
	border-radius: 4px;
	text-transform: capitalize;
}
/* restyle Try copy to clipboard button */
div.copy-to-clipboard {
	text-align: right;
	margin-top: -1em;
}
div.copy-to-clipboard button {
	min-height: 1em;
}
/* try responses area - make scrollable */
div.responses-inner > div > div > table.live-responses-table {
	table-layout: fixed;
}
div.responses-inner > div > div > table.live-responses-table td.response-col_description {
	width: 80%;
}
body #redoc-container {
	float: left;
	width: 100%;
}
`

func defValue(v, def string) string {
	if v != "" {
		return v
	}
	return def
}

func yaml2Json(yamlData []byte) (data []byte, err error) {
	r := &node{}
	if err = yaml.Unmarshal(yamlData, r); err == nil {
		data, err = json.Marshal(r.Value)
	}
	return
}

type node struct {
	Key   string
	Value any
}

func (n *node) UnmarshalYAML(value *yaml.Node) (err error) {
	n.Key = value.Tag
	switch value.Kind {
	case yaml.ScalarNode:
		n.Value = value.Value
	case yaml.MappingNode:
		childMap := make(map[string]any)
		for i := 0; i < len(value.Content) && err == nil; i += 2 {
			keyNode := value.Content[i]
			valNode := value.Content[i+1]
			child := &node{}
			err = child.UnmarshalYAML(valNode)
			childMap[keyNode.Value] = child.Value
		}
		n.Value = childMap
	case yaml.SequenceNode:
		childSlice := make([]any, len(value.Content))
		for i, childNode := range value.Content {
			child := &node{}
			err = child.UnmarshalYAML(childNode)
			childSlice[i] = child.Value
		}
		n.Value = childSlice
	}
	return
}
