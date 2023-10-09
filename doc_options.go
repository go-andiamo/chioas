package chioas

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
	"html/template"
	"net/http"
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
	// Title the title in the docs index page (defaults to "API Documentation")
	Title string
	// RedocOptions redoc options to be used (see https://github.com/Redocly/redoc#redoc-options-object)
	//
	// use map[string]any or &RedocOptions (or anything that can be marshalled and then unmarshalled to map[string]any)
	RedocOptions any
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
	RedocJsUrl string
	// TryJsUrl is the URL for the Try Redoc JS
	//
	// defaults to:
	// https://cdn.jsdelivr.net/gh/wll8/redoc-try@1.4.7/dist/try.js
	TryJsUrl string
	// DefaultResponses is the default responses for methods that don't have any responses defined
	//
	// is a map of http status code and response
	DefaultResponses Responses
	// HideHeadMethods indicates that all HEAD methods should be hidden from docs
	HideHeadMethods bool
	// OperationIdentifier is an optional function called by Method to generate `operationId` tag value
	OperationIdentifier OperationIdentifier
	// specData is used internally where api has been generated from spec (see FromJson and FromYaml)
	specData []byte
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

func (d *DocOptions) setupRoutes(def *Definition, route chi.Router) error {
	if d.ServeDocs {
		tmpStr := defValue(d.DocTemplate, defaultTemplate)
		tmp, err := template.New("index").Parse(tmpStr)
		if err != nil {
			return err
		}
		path := defValue(d.Path, defaultDocsPath)
		indexPage := defValue(d.DocIndexPage, defaultIndexName)
		specName := ""
		if d.AsJson {
			specName = defValue(d.SpecName, defaultSpecNameJson)
		} else {
			specName = defValue(d.SpecName, defaultSpecName)
		}
		data := map[string]any{
			htmlTagTitle:          defValue(d.Title, defaultTitle),
			htmlTagStylesOverride: template.CSS(defValue(d.StylesOverride, defaultStylesOverride)),
			htmlTagSpecName:       specName,
			htmlTagRedocOpts:      d.getRedocOptions(),
			htmlTagRedocUrl:       defValue(d.RedocJsUrl, defaultRedocJsUrl),
			htmlTagTryUrl:         defValue(d.TryJsUrl, defaultTryJsUrl),
		}
		redirectPath := path + root + indexPage
		docsRoute := chi.NewRouter()
		docsRoute.Get(root, http.RedirectHandler(redirectPath, http.StatusMovedPermanently).ServeHTTP)
		if d.specData != nil || !d.NoCache {
			if err := d.setupCachedRoutes(def, docsRoute, tmp, data, indexPage, specName); err != nil {
				return err
			}
		} else {
			d.setupNoCachedRoutes(def, docsRoute, tmp, data, indexPage, specName)
		}
		route.Mount(path, docsRoute)
	}
	return nil
}

func (d *DocOptions) setupCachedRoutes(def *Definition, docsRoute *chi.Mux, tmp *template.Template, inData map[string]any, indexPage, specName string) (err error) {
	indexData, err := d.buildIndexData(tmp, inData)
	if err != nil {
		return err
	}
	docsRoute.Get(root+indexPage, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentTypeHtml)
		_, _ = writer.Write(indexData)
	})
	var specData []byte
	contentType := contentTypeYaml
	if d.specData != nil {
		specData = d.specData
	} else if d.AsJson {
		contentType = contentTypeJson
		if specData, err = def.AsJson(); err != nil {
			return err
		}
	} else {
		if specData, err = def.AsYaml(); err != nil {
			return err
		}
	}
	docsRoute.Get(root+specName, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentType)
		_, _ = writer.Write(specData)
	})
	return nil
}

func (d *DocOptions) setupNoCachedRoutes(def *Definition, docsRoute *chi.Mux, tmp *template.Template, inData map[string]any, indexPage, specName string) {
	docsRoute.Get(root+indexPage, func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrContentType, contentTypeHtml)
		_ = tmp.Execute(writer, inData)
	})
	if d.AsJson {
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

func (d *DocOptions) getRedocOptions() map[string]any {
	if d.RedocOptions != nil {
		switch opt := d.RedocOptions.(type) {
		case map[string]any:
			return opt
		default:
			if data, err := json.Marshal(d.RedocOptions); err == nil {
				var res map[string]any
				if json.Unmarshal(data, &res) == nil {
					return res
				}
			}
		}
	}
	return map[string]any{}
}

func (d *DocOptions) buildIndexData(tmp *template.Template, inData map[string]any) (data []byte, err error) {
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
	contentTypeHtml       = "text/html"
	contentTypeJson       = "application/json"
	contentTypeYaml       = "application/yaml"
	hdrContentType        = "Content-Type"
	htmlTagTitle          = "title"
	htmlTagStylesOverride = "stylesOverride"
	htmlTagSpecName       = "specName"
	htmlTagRedocOpts      = "redocopts"
	htmlTagRedocUrl       = "redocurl"
	htmlTagTryUrl         = "tryurl"
)

const defaultTemplate = `<html>
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
const defaultStylesOverride = `body {
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
