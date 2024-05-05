package main

import (
	"embed"
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (a *api) SetupRoutes(r chi.Router) error {
	return a.Definition.SetupRoutes(r, a)
}

//go:embed *.png
var supportFilesFS embed.FS

var apiDef = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs:               true, // makes docs served as interactive UI on /docs/index.htm
		UIStyle:                 chioas.Rapidoc,
		SupportFiles:            http.FileServer(http.FS(supportFilesFS)),
		SupportFilesStripPrefix: true,
		RapidocOptions: &chioas.RapidocOptions{
			ShowHeader:  true,
			HeadingText: "Petstore",
			//AllowSpecUrlLoad:   true,
			Theme:              "dark",
			ShowMethodInNavBar: "as-colored-block",
			UsePathInNavBar:    true,
			UpdateRoute:        true,
			//LogoSrc:            "dog.png",
			InnerHtml: `<img id="logo" slot="logo" src="dog.png" width="30" height="30">
<div slot="nav-logo" style="display: flex; align-items: center; justify-content: center;"> 
    <img src="dog.png" style="width:40px; margin-right: 20px"> <span><b>PetStore API</b></span>
</div>`,
			AdditionalAttributes: map[string]string{
				"nav-active-item-marker": "left-bar",
				"text-color":             "#133863",
				"bg-color":               "#ffffff",
			},
		},
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
