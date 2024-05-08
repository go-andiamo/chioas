package main

import (
	"github.com/go-andiamo/chioas"
	"github.com/go-chi/chi/v5"
)

func (a *api) SetupRoutes(r chi.Router) error {
	return a.Definition.SetupRoutes(r, a)
}

var apiDef = chioas.Definition{
	DocOptions: chioas.DocOptions{
		ServeDocs:      true, // makes docs served as interactive UI on /docs/index.htm
		UIStyle:        chioas.Redoc,
		StylesOverride: styling,
		RedocOptions: chioas.RedocOptions{
			HeaderHtml: `<div style="display:flex; margin:10px; justify-content:center;flex-wrap: wrap;">
    <a class="chioasbtn" href="../swagger/index.html">Swagger</a>
    <a class="chioasbtn" href="../rapidoc/index.html">Rapidoc</a>
</div>`,
		},
		AlternateUIDocs: chioas.AlternateUIDocs{
			"/swagger": {
				UIStyle:        chioas.Swagger,
				StylesOverride: styling,
				SwaggerOptions: chioas.SwaggerOptions{
					DeepLinking: true,
					HeaderHtml: `<div style="display:flex; margin:10px; justify-content:center;flex-wrap: wrap;">
    <a class="chioasbtn" href="../docs/index.html">Redoc</a>
    <a class="chioasbtn" href="../rapidoc/index.html">Rapidoc</a>
</div>`,
				},
			},
			"/rapidoc": {
				UIStyle:        chioas.Rapidoc,
				StylesOverride: styling,
				RapidocOptions: &chioas.RapidocOptions{
					ShowHeader:         true,
					HeadingText:        "Petstore",
					Theme:              "dark",
					ShowMethodInNavBar: "as-colored-block",
					UsePathInNavBar:    true,
					SchemaStyle:        "table",
					HeadScript: `
function getRapiDoc(){
    return document.getElementById("thedoc");
}
function toggleView() {
	let currRender = getRapiDoc().getAttribute('render-style');
	let newRender = currRender === "read" ? "view" : "read";
	getRapiDoc().setAttribute('render-style', newRender );
}
function toggleTheme(){
	if (getRapiDoc().getAttribute('theme') === 'dark'){
	    getRapiDoc().setAttribute('theme',"light");
	}
	else{
	    getRapiDoc().setAttribute('theme',"dark");
	}
}`,
					InnerHtml: `<div style="display:flex; margin:10px; justify-content:center;flex-wrap: wrap;">
    <button class="chioasbtn" onclick="toggleView()">Change View</button>
    <button class="chioasbtn" onclick="toggleTheme()" style="margin-right:30px">Change Theme</button>
    <a class="chioasbtn" href="../docs/index.html">Redoc</a>
    <a class="chioasbtn" href="../swagger/index.html">Swagger</a>
</div>`,
				},
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

const styling = `.chioasbtn{
	min-width: 100px;
	background-color: #47AFE8;
	color: #fff;
	font-size: 12px;
	display: block;
	border: none;
	margin: 2px;
	border-radius: 2px;
	cursor: pointer;
	outline: none;
    text-align: center;
    padding: 4px;
    text-decoration: none;
    font-family: sans-serif;
}
.chioasbtn:visited{
    color: #fff;
}`

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
