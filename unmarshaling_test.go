package chioas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"net/http"
	"testing"
)

func TestDefinition_UnmarshalJSON(t *testing.T) {
	data, err := fullDefJson()
	require.NoError(t, err)

	fmt.Println(string(data))

	d := Definition{}
	err = json.Unmarshal(data, &d)
	require.NoError(t, err)
}

func TestDefinition_UnmarshalYAML(t *testing.T) {
	data, err := fullDefYaml()
	require.NoError(t, err)

	fmt.Println(string(data))

	d := Definition{}
	err = yaml.Unmarshal(data, &d)
	require.NoError(t, err)
}

func TestDefinition_UnmarshalJSON_PetstoreYaml(t *testing.T) {
	data := []byte(petstoreYaml)
	d := Definition{}
	err := yaml.Unmarshal(data, &d)
	require.NoError(t, err)
	require.Len(t, d.Paths, 1)
	api, ok := d.Paths["/api"]
	require.True(t, ok)
	require.Len(t, api.Methods, 1)
	_, ok = api.Methods["GET"]
	require.True(t, ok)
	require.Len(t, api.Paths, 2)
	p, ok := api.Paths["/pets"]
	require.True(t, ok)
	require.Len(t, p.Methods, 2)
	_, ok = p.Methods["GET"]
	require.True(t, ok)
	_, ok = p.Methods["POST"]
	require.True(t, ok)
	require.Len(t, p.Paths, 1)
	_, ok = p.Paths["/{id}"]
	require.True(t, ok)

	p, ok = api.Paths["/categories"]
	require.True(t, ok)
	require.Len(t, p.Methods, 1)
	_, ok = p.Methods["GET"]
	require.True(t, ok)
	require.Len(t, p.Paths, 1)
	_, ok = p.Paths["/{id}"]
	require.True(t, ok)
}

func fullDefJson() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := testFullDef.WriteJson(buf)
	return buf.Bytes(), err
}

func fullDefYaml() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := testFullDef.WriteYaml(buf)
	return buf.Bytes(), err
}

var testFullDef = Definition{
	Servers: Servers{
		"/api/v1": {
			Description: "original",
			Extensions: Extensions{
				"foo": "bar",
			},
		},
	},
	Info: Info{
		Title:          "test title",
		Description:    "test desc",
		Version:        "1.0.1",
		TermsOfService: "http://example.com",
		Contact: &Contact{
			Name:  "John Doe",
			Url:   "http://example.com",
			Email: "test@example.com",
			Extensions: Extensions{
				"foo": "bar",
			},
		},
		License: &License{
			Name: "MIT",
			Url:  "https://opensource.org/licenses/MIT",
			Extensions: Extensions{
				"foo": "bar",
			},
		},
		ExternalDocs: &ExternalDocs{
			Description: "external documentation",
			Url:         "http://example.com",
			Extensions: Extensions{
				"foo": "bar",
			},
		},
		Extensions: Extensions{
			"foo": "bar",
		},
	},
	Tags: Tags{
		{
			Name:        "Foo",
			Description: "foo tag",
			ExternalDocs: &ExternalDocs{
				Description: "external documentation",
			},
			Extensions: Extensions{
				"foo": "bar",
			},
		},
		{
			Name:        "Subs",
			Description: "subs tag",
		},
	},
	Methods: Methods{
		http.MethodGet: {
			Description: "Root discovery",
			Tag:         "root",
			Extensions: Extensions{
				"foo": "bar",
			},
		},
	},
	Paths: Paths{
		"/subs": {
			Tag: "Subs",
			Methods: Methods{
				http.MethodGet: {
					Description: "get subs desc",
					QueryParams: QueryParams{
						{
							Name:        "search",
							Description: "search desc",
							Required:    true,
						},
					},
					Security: SecuritySchemes{
						{
							Name: "ApiKey",
						},
						{
							Name: "MyOauth",
						},
					},
				},
				http.MethodPost: {
					Description: "Post subs desc",
					Request: &Request{
						Required:    true,
						Description: "Get subs desc",
						Schema: &Schema{
							Type:               "object",
							RequiredProperties: []string{"foo", "bar"},
							Properties: Properties{
								{
									Name: "foo",
									Type: "string",
								},
								{
									Name: "bar",
									Type: "string",
								},
							},
						},
						Extensions: Extensions{
							"foo": "bar",
						},
					},
				},
			},
			Paths: Paths{
				"/{subId: [a-z]*}": {
					PathParams: PathParams{
						"subId": {
							Description: "id of sub",
						},
					},
					Methods: Methods{
						http.MethodGet: {
							Description: "get specific sub",
						},
					},
					Paths: Paths{
						"/subitems": {
							Paths: Paths{
								"/{subitemId}": {
									Methods: Methods{
										http.MethodGet: {
											Description: "get specific sub-item of sub",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"/foos": {
			Paths: Paths{
				"/bars": {
					Paths: Paths{
						"/bazs": {
							Methods: Methods{
								http.MethodGet: {
									Description:      "baz method",
									OperationId:      "bazId",
									OptionalSecurity: true,
								},
							},
						},
					},
				},
			},
		},
	},
	Components: &Components{
		Schemas: Schemas{
			{
				Name:               "foo",
				Type:               "object",
				RequiredProperties: []string{"bar"},
				Properties: Properties{
					{
						Name:      "bar",
						SchemaRef: "fobar",
						Constraints: Constraints{
							Pattern:          "[0-9]{3}",
							Maximum:          json.Number("3"),
							Minimum:          json.Number("1"),
							ExclusiveMaximum: true,
							ExclusiveMinimum: true,
							Nullable:         true,
							MultipleOf:       3,
							MaxLength:        3,
							MinLength:        3,
							MaxItems:         3,
							MinItems:         1,
							UniqueItems:      true,
							MaxProperties:    3,
						},
					},
				},
				Discriminator: &Discriminator{
					PropertyName: "bar",
					Mapping: map[string]string{
						"foo": "bar",
					},
					Extensions: Extensions{
						"foo": "bar",
					},
				},
			},
			{
				Name:               "foo2",
				Type:               "object",
				RequiredProperties: []string{"bar"},
				Properties: Properties{
					{
						Name: "bar",
						Type: "number",
						Constraints: Constraints{
							Pattern:          "[0-9]{3}",
							Maximum:          json.Number("3"),
							Minimum:          json.Number("1"),
							ExclusiveMaximum: true,
							ExclusiveMinimum: true,
							Nullable:         true,
							MultipleOf:       3,
							MaxLength:        3,
							MinLength:        3,
							MaxItems:         3,
							MinItems:         1,
							UniqueItems:      true,
							MaxProperties:    3,
						},
					},
				},
			},
			{
				Name:               "Person",
				Type:               "object",
				Description:        "desc",
				RequiredProperties: []string{"id", "name"},
				Properties: []Property{
					{
						Name: "id",
					},
					{
						Name: "name",
					},
				},
			},
			{
				Name:               "Address",
				Type:               "object",
				Description:        "desc",
				RequiredProperties: []string{"street", "city"},
				Properties: []Property{
					{
						Name: "street",
					},
					{
						Name: "city",
					},
				},
			},
			{
				Name:        "User",
				Description: "desc",
				Ofs: &Ofs{
					OfType: AllOf,
					Of: []OfSchema{
						&Of{SchemaRef: "Person"},
						&Of{SchemaRef: "Address"},
						&Schema{
							Type:               "object",
							RequiredProperties: []string{"email"},
							Properties: []Property{
								{
									Name:   "email",
									Type:   "string",
									Format: "email",
								},
							},
						},
					},
				},
			},
		},
		SecuritySchemes: SecuritySchemes{
			{
				Name:        "ApiKey",
				Description: "foo",
				Type:        "apiKey",
				In:          "header",
				ParamName:   "X-API-KEY",
			},
			{
				Name:   "MyOauth",
				Type:   "oauth2",
				Scopes: []string{"write:foo", "read:foo"},
			},
		},
		Examples: Examples{
			{
				Name:        "example1",
				Description: "example1",
				Summary:     "example1",
				Value:       "some value",
				//ExampleRef:  "example1",
			},
			{
				Name:       "example2",
				ExampleRef: "example1",
				Extensions: Extensions{
					"foo": "bar",
				},
			},
		},
		Parameters: CommonParameters{
			"foo": {
				Name:        "xfoo",
				Description: "xfoo",
				Required:    true,
				Example:     "example",
				Schema: &Schema{
					Description: "schema description",
					Type:        "number",
					Enum:        []any{1, 2, 3},
				},
				Extensions: Extensions{
					"foo": "bar",
				},
			},
			"bar": {
				Name:        "xbar",
				Description: "xbar",
				Required:    true,
				In:          "path",
				Example:     "example",
			},
			"baz": {
				Name:        "xbaz",
				Description: "xbaz",
				Required:    true,
				In:          "header",
				Example:     "example",
				SchemaRef:   "baz",
			},
		},
		Requests: CommonRequests{
			"req1": {
				Description: "req1",
				Required:    true,
				IsArray:     true,
				Examples: Examples{
					{
						Name:        "example1",
						Description: "example1",
						Value:       "some value",
					},
					{
						Name:        "example2",
						Description: "example2",
						ExampleRef:  "example2",
					},
				},
				Extensions: Extensions{
					"foo": "bar",
				},
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						Examples: Examples{
							{
								Name:        "example1",
								Description: "example1",
								Value:       "some value",
							},
							{
								Name:        "example2",
								Description: "example2",
								ExampleRef:  "example2",
							},
						},
					},
				},
			},
		},
		Responses: CommonResponses{
			"res1": {
				Description: "res1",
				IsArray:     true,
				Examples: Examples{
					{
						Name:        "example1",
						Description: "example1",
						Value:       "some value",
					},
					{
						Name:        "example2",
						Description: "example2",
						ExampleRef:  "example2",
					},
				},
				Extensions: Extensions{
					"foo": "bar",
				},
				AlternativeContentTypes: ContentTypes{
					"text/csv": {
						Examples: Examples{
							{
								Name:        "example1",
								Description: "example1",
								Value:       "some value",
							},
							{
								Name:        "example2",
								Description: "example2",
								ExampleRef:  "example2",
							},
						},
					},
				},
			},
		},
		Extensions: Extensions{
			"foo": "bar",
		},
	},
	Security: SecuritySchemes{
		{
			Name:        "ApiKey",
			Description: "foo",
			Type:        "apiKey",
			In:          "header",
			ParamName:   "X-API-KEY",
		},
		{
			Name:   "MyOauth",
			Type:   "oauth2",
			Scopes: []string{"write:foo", "read:foo"},
		},
	},
	Comment: "this is a test comment\nand so is this",
}

const petstoreYaml = `openapi: "3.0.3"
info:
  title: "Pet Store API"
  version: "1.0.0"
paths:
  "/api":
    get:
      description: "Root discovery"
      tags:
        - Root
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
  "/api/categories":
    get:
      description: "List categories"
      tags:
        - Categories
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: array
                items:
                  type: object
                  properties:
                    "$uri":
                      description: "URI of the category"
                      type: string
                      format: uri
                    "id":
                      description: "Unique identifier of the category"
                      type: string
                      format: uuid
                    "name":
                      description: "Name of the category"
                      type: string
  "/api/categories/{id}":
    get:
      description: "Get category"
      tags:
        - Categories
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
                properties:
                  "$uri":
                    description: "URI of the category"
                    type: string
                    format: uri
                  "id":
                    description: "Unique identifier of the category"
                    type: string
                    format: uuid
                  "name":
                    description: "Name of the category"
                    type: string
  "/api/pets":
    get:
      description: "List/search pets"
      tags:
        - Pets
      parameters:
        - name: category
          description: "Filter by category"
          in: query
          required: false
          schema:
            type: string
        - name: name
          description: "Search/filter by name"
          in: query
          required: false
          schema:
            type: string
        - name: dob
          description: "Filter by dob (date of birth)"
          in: query
          required: false
          schema:
            type: string
            format: date
        - name: order
          description: "Order result by property"
          in: query
          required: false
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: array
                items:
                  type: object
                  properties:
                    "$uri":
                      description: "URI of the pet"
                      type: string
                      format: uri
                    "id":
                      description: "Unique identifier of the pet"
                      type: string
                      format: uuid
                    "name":
                      description: "Name of the pet"
                      type: string
                    "dob":
                      description: "Date of birth"
                      type: string
                      format: date
                    "category":
                      description: "Category of the pet"
                      type: object
                      properties:
                        "$uri":
                          description: "URI of the category"
                          type: string
                          format: uri
                        "id":
                          description: "Unique identifier of the category"
                          type: string
                          format: uuid
                        "name":
                          description: "Name of the category"
                          type: string
    post:
      description: "Add pet"
      tags:
        - Pets
      requestBody:
        description: "Pet to add"
        required: true
        content:
          "application/json":
            schema:
              type: object
              required:
                - name
                - dob
                - category
              properties:
                "name":
                  description: "Name of the pet"
                  type: string
                "dob":
                  description: "Date of birth"
                  type: string
                  format: date
                "category":
                  description: "Category of the pet"
                  type: object
                  properties:
                    "id":
                      description: "ID of the category"
                      type: string
                      format: uuid
                    "name":
                      description: "Name of the category"
                      type: string
      responses:
        201:
          description: Created
          content:
            "application/json":
              schema:
                type: object
                properties:
                  "$uri":
                    description: "URI of the pet"
                    type: string
                    format: uri
                  "id":
                    description: "Unique identifier of the pet"
                    type: string
                    format: uuid
                  "name":
                    description: "Name of the pet"
                    type: string
                  "dob":
                    description: "Date of birth"
                    type: string
                    format: date
                  "category":
                    description: "Category of the pet"
                    type: object
                    properties:
                      "$uri":
                        description: "URI of the category"
                        type: string
                        format: uri
                      "id":
                        description: "Unique identifier of the category"
                        type: string
                        format: uuid
                      "name":
                        description: "Name of the category"
                        type: string
  "/api/pets/{id}":
    get:
      description: "Get pet"
      tags:
        - Pets
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
                properties:
                  "$uri":
                    description: "URI of the pet"
                    type: string
                    format: uri
                  "id":
                    description: "Unique identifier of the pet"
                    type: string
                    format: uuid
                  "name":
                    description: "Name of the pet"
                    type: string
                  "dob":
                    description: "Date of birth"
                    type: string
                    format: date
                  "category":
                    description: "Category of the pet"
                    type: object
                    properties:
                      "$uri":
                        description: "URI of the category"
                        type: string
                        format: uri
                      "id":
                        description: "Unique identifier of the category"
                        type: string
                        format: uuid
                      "name":
                        description: "Name of the category"
                        type: string
    put:
      description: "Update pet"
      tags:
        - Pets
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        description: "Pet update"
        required: true
        content:
          "application/json":
            schema:
              type: object
              required:
                - name
                - dob
              properties:
                "name":
                  description: "Name of the pet"
                  type: string
                "dob":
                  description: "Date of birth"
                  type: string
                  format: date
                "category":
                  description: "Category of the pet"
                  type: object
                  properties:
                    "id":
                      description: "ID of the category"
                      type: string
                      format: uuid
                    "name":
                      description: "Name of the category"
                      type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
                properties:
                  "$uri":
                    description: "URI of the pet"
                    type: string
                    format: uri
                  "id":
                    description: "Unique identifier of the pet"
                    type: string
                    format: uuid
                  "name":
                    description: "Name of the pet"
                    type: string
                  "dob":
                    description: "Date of birth"
                    type: string
                    format: date
                  "category":
                    description: "Category of the pet"
                    type: object
                    properties:
                      "$uri":
                        description: "URI of the category"
                        type: string
                        format: uri
                      "id":
                        description: "Unique identifier of the category"
                        type: string
                        format: uuid
                      "name":
                        description: "Name of the category"
                        type: string
    delete:
      description: "Delete pet"
      tags:
        - Pets
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        200:
          description: OK
          content:
            "application/json":
              schema:
                type: object
`
