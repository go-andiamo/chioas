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
