package main

import (
	"embed"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// part of our spec is a static file! (though it doesn't have to be - just for demo purposes)
// we're also using it to serve up a logo image - which is used in the docs template html
//
//go:embed status_schema.yaml petstore-logo.png
var supportFilesFS embed.FS

var apiDef = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs: true, // makes docs served as interactive UI on /docs/index.htm
		UIStyle:   chioas.Swagger,
		StylesOverride: `.logo img {
	padding: inherit;
	margin: auto;
    width: 200px;
    display: block;
}`,
		SwaggerOptions: chioas.SwaggerOptions{
			HeaderHtml: `<div class="logo" style="align-content: center">
    <img src="./petstore-logo.png" alt="pet store logo">
</div>`,
		},
		SupportFiles:            http.FileServer(http.FS(supportFilesFS)),
		SupportFilesStripPrefix: true,
		CheckRefs:               true, // make sure that any $ref's are valid!
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
	},
	Paths: chioas.Paths{
		"/pets": petsPaths,
	},
	Components: &components,
}

var components = chioas.Components{
	Schemas: chioas.Schemas{
		// generate schema from the PetRequestResponse struct...
		(&chioas.Schema{
			Name:        "Pet",
			Description: "Pet request/response",
			Comment:     chioas.SourceComment(),
		}).Must(PetRequestResponse{
			Id:        1,
			Name:      "doggie",
			PhotoUrls: []string{"https://example.com/dog-picture.jpg"},
		}),
		// generate schema from the PetCategory struct...
		(&chioas.Schema{
			Name:        "Category",
			Description: "Pet category",
			Comment:     chioas.SourceComment(),
		}).Must(PetCategory{
			Id:   1,
			Name: "Dogs",
		}),
		{
			Name:      "Status",
			SchemaRef: "./status_schema.yaml", // we could have defined the schema here - but for demo purposes, we're serving a static file (see supportFiles)
			//			Type:    "string",
			//			Default: "available",
			//			Enum:    []any{"available", "pending", "sold"},
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

func (a *api) SetupRoutes(r chi.Router) error {
	return a.Definition.SetupRoutes(r, a)
}
