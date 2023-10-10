package main

import (
	"embed"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func (a *api) SetupRoutes(r chi.Router) error {
	return a.Definition.SetupRoutes(r, a)
}

//go:embed files/*
var swaggerFs embed.FS

var apiDef = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs:    true, // makes docs served as interactive UI on /docs/index.htm
		SupportFiles: &swaggerStatics{fs: swaggerFs},
		DocTemplate:  swaggerTemplate,
	},
	Info: chioas.Info{
		Title: "Swagger Petstore - OpenAPI 3.0",
		Description: `This is a sample Pet Store Server based on the OpenAPI 3.0 specification.


You can find out more about Swagger at [http://swagger.io](http://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!


You can now help us improve the API whether it's by making changes to the definition itself or to the code.
That way, with time, we can improve the API in general, and expose some of the new features in OAS3.


Some useful links:

- [The Pet Store
repository](https://github.com/swagger-api/swagger-petstore)

- [The source API definition for the Pet
Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)`,
		Version: "1.0.17",
	},
	Tags: chioas.Tags{
		{
			Name:        "pet",
			Description: "Everything about your Pets",
			ExternalDocs: &chioas.ExternalDocs{
				Description: "Find out more",
				Url:         "http://swagger.io",
			},
		},
		{
			Name:        "store",
			Description: "Access to Petstore orders",
			ExternalDocs: &chioas.ExternalDocs{
				Description: "Find out more about our store",
				Url:         "http://swagger.io",
			},
		},
		{
			Name:        "user",
			Description: "Operations about user",
		},
	},
	Paths: chioas.Paths{
		"/pets": petsPaths,
	},
	Components: &components,
}

var components = chioas.Components{
	Schemas: chioas.Schemas{
		{
			Name:               "Pet",
			Type:               "object",
			RequiredProperties: []string{"name", "photoUrls"},
			Properties: []chioas.Property{
				{
					Name:    "id",
					Type:    "integer",
					Example: 10,
				},
				{
					Name:    "name",
					Type:    "string",
					Example: "doggie",
				},
				{
					Name:      "category",
					SchemaRef: "Category",
				},
				{
					Name:     "photoUrls",
					Type:     "array",
					ItemType: "string",
					Example:  "https://example.com/dog-picture.jpg",
				},
			},
		},
		{
			Name: "Category",
			Type: "object",
			Properties: []chioas.Property{
				{
					Name:    "id",
					Type:    "integer",
					Example: 1,
				},
				{
					Name:    "name",
					Type:    "string",
					Example: "Dogs",
				},
			},
		},
		{
			Name:    "Status",
			Type:    "string",
			Default: "available",
			Enum:    []any{"available", "pending", "sold"},
		},
	},
}

type api struct {
	chioas.Definition
}

var petStoreApi = &api{
	Definition: apiDef,
}

func commenter(handlerMethod string, comments ...string) []string {
	if handlerMethod != "" {
		if mbn, ok := reflect.TypeOf(&api{}).MethodByName(handlerMethod); ok {
			fn := mbn.Func.Pointer()
			fi := runtime.FuncForPC(fn)
			ff, fl := fi.FileLine(fi.Entry())
			n := ""
			if pts := strings.Split(ff, "/"); len(pts) > 0 {
				n = pts[len(pts)-1]
			}
			comments = append(comments, fmt.Sprintf("handler: %s()  %s:%d", handlerMethod, n, fl))
		}
	}
	return comments
}

type swaggerStatics struct {
	fs embed.FS
}

func (m *swaggerStatics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fn := strings.TrimPrefix(r.URL.String(), "/docs/")
	if data, err := m.fs.ReadFile("files/" + fn); err == nil {
		if ctype := mime.TypeByExtension(filepath.Ext(fn)); ctype != "" {
			w.Header().Set("Content-Type", ctype)
		}
		_, _ = w.Write(data)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

const swaggerTemplate = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
        box-sizing: inherit;
    }

    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>
  </defs>
</svg>

<div id="swagger-ui"></div>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
window.onload = function() {
  const ui = SwaggerUIBundle({
    url: "{{.specName}}",
    dom_id: "#swagger-ui",
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
  })
  window.ui = ui
}
</script>
</body>
</html>
`
