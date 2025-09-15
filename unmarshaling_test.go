package chioas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"math"
	"net/http"
	"strings"
	"testing"
)

func TestDefinition_UnmarshalJSON(t *testing.T) {
	data, err := fullDefJson()
	require.NoError(t, err)
	//fmt.Println(string(data))
	d := Definition{}
	err = json.Unmarshal(data, &d)
	require.NoError(t, err)
}

func TestDefinition_UnmarshalJSON_BadJson(t *testing.T) {
	d := Definition{}
	err := json.Unmarshal([]byte(`[]`), &d)
	require.Error(t, err)
}

func TestDefinition_UnmarshalYAML(t *testing.T) {
	data, err := fullDefYaml()
	require.NoError(t, err)
	//fmt.Println(string(data))
	d := Definition{}
	err = yaml.Unmarshal(data, &d)
	require.NoError(t, err)
}

func TestDefinition_unmarshalObj_Errors(t *testing.T) {
	t.Run("bad info", func(t *testing.T) {
		d := &Definition{}
		err := d.unmarshalObj(map[string]any{
			tags.Info: "not an object",
		})
		require.Error(t, err)
	})
	t.Run("bad externalDocs", func(t *testing.T) {
		d := &Definition{}
		err := d.unmarshalObj(map[string]any{
			tags.ExternalDocs: "not an object",
		})
		require.Error(t, err)
	})
}

func TestDefinition_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Info:         map[string]any{},
		tags.ExternalDocs: map[string]any{},
		tags.Tags: []any{
			map[string]any{
				tags.Name: "test tag",
			},
		},
		tags.Servers: []any{
			map[string]any{
				tags.Url: "test url",
			},
		},
		tags.Security: []any{
			map[string]any{
				"foo": []any{"all"},
			},
		},
		tags.Paths: map[string]any{
			"/root": map[string]any{},
		},
		tags.Components: map[string]any{},
		"x-foo":         "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Definition](m)
		require.NoError(t, err)
		assert.Len(t, r.Tags, 1)
		assert.Len(t, r.Servers, 1)
		assert.Len(t, r.Security, 1)
		assert.NotNil(t, r.Components)
		assert.Len(t, r.Paths, 1)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Method](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestDefinition_unmarshalPaths(t *testing.T) {
	t.Run("root methods", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				root: map[string]any{
					"get":     map[string]any{},
					"options": map[string]any{},
				},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.NoError(t, err)
		assert.Len(t, d.Methods, 2)
		_, ok := d.Methods[http.MethodGet]
		assert.True(t, ok)
		_, ok = d.Methods[http.MethodOptions]
		assert.True(t, ok)
	})
	t.Run("fill in ancestry", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				"/foo/bar/baz": map[string]any{
					"get": map[string]any{},
				},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.NoError(t, err)
		assert.Len(t, d.Paths, 1)
		p, ok := d.Paths["/foo"]
		require.True(t, ok)
		p, ok = p.Paths["/bar"]
		require.True(t, ok)
		p, ok = p.Paths["/baz"]
		require.True(t, ok)
		assert.Len(t, p.Methods, 1)
	})
	t.Run("root method not object", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				root: map[string]any{
					"get": "not an object",
				},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.Error(t, err)
	})
	t.Run("path method not object", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				"/foo/bar": map[string]any{
					"get": "not an object",
				},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.Error(t, err)
	})
	t.Run("path method unmarshal object fails", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				"/foo/bar": map[string]any{
					"get": map[string]any{
						tags.Description: false,
					},
				},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.Error(t, err)
	})
	t.Run("fails to split path", func(t *testing.T) {
		m := map[string]any{
			tags.Paths: map[string]any{
				"/api/{unbalanced}}": map[string]any{},
			},
		}
		d := &Definition{}
		err := d.unmarshalPaths(m)
		require.Error(t, err)
	})
}

func TestMethod_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Summary:     "test summary",
		tags.OperationId: "test operation id",
		tags.Deprecated:  true,
		tags.Tags:        []any{"test tag"},
		tags.RequestBody: map[string]any{},
		tags.Parameters: []any{
			map[string]any{},
		},
		tags.Responses: map[string]any{
			"200": map[string]any{},
		},
		tags.Security: []any{
			map[string]any{},
		},
		"x-foo": "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Method](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test summary", r.Summary)
		assert.Equal(t, "test operation id", r.OperationId)
		assert.True(t, r.Deprecated)
		assert.Equal(t, "test tag", r.Tag)
		assert.NotNil(t, r.Request)
		assert.Len(t, r.QueryParams, 1)
		assert.Len(t, r.Responses, 1)
		_, ok := r.Responses[200]
		assert.True(t, ok)
		assert.Len(t, r.Security, 0)
		assert.True(t, r.OptionalSecurity)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("success with security", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Security] = []any{
			map[string]any{
				"foo": nil,
			},
		}
		r, err := fromObj[Method](m2)
		require.NoError(t, err)
		assert.Len(t, r.Security, 1)
		assert.Equal(t, "foo", r.Security[0].Name)
		assert.False(t, r.OptionalSecurity)
	})
	t.Run("fails with invalid response code", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Responses] = map[string]any{
			"not a number": map[string]any{},
		}
		_, err := fromObj[Method](m2)
		require.Error(t, err)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Method](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestInfo_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Title:          "test title",
		tags.Description:    "test description",
		tags.Version:        "test version",
		tags.TermsOfService: "test terms of service",
		tags.Contact:        map[string]any{},
		tags.License:        map[string]any{},
		"x-foo":             "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Info](m)
		require.NoError(t, err)
		assert.Equal(t, "test title", r.Title)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test version", r.Version)
		assert.Equal(t, "test terms of service", r.TermsOfService)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Info](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestContact_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:  "test name",
		tags.Url:   "test url",
		tags.Email: "test email",
		"x-foo":    "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Contact](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test url", r.Url)
		assert.Equal(t, "test email", r.Email)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Contact](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestLicense_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name: "test name",
		tags.Url:  "test url",
		"x-foo":   "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[License](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test url", r.Url)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[License](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestExternalDocs_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Url:         "test url",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[ExternalDocs](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test url", r.Url)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[ExternalDocs](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestTag_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:         "test name",
		tags.Description:  "test description",
		tags.ExternalDocs: map[string]any{},
		"x-foo":           "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Tag](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test description", r.Description)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Tag](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestServer_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Server](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Server](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestSecurityScheme_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Type:        "test type",
		tags.Scheme:      "test scheme",
		tags.Name:        "test name",
		tags.In:          "test in",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[SecurityScheme](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test type", r.Type)
		assert.Equal(t, "test scheme", r.Scheme)
		assert.Equal(t, "test name", r.ParamName)
		assert.Equal(t, "test in", r.In)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[SecurityScheme](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestExample_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Summary:     "test summary",
		tags.Value:       "test value",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Example](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test summary", r.Summary)
		assert.Equal(t, "test value", r.Value)
		assert.Empty(t, r.ExampleRef)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Ref] = "test ref"
		r, err := fromObj[Example](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.ExampleRef)
		assert.Empty(t, r.Description)
		assert.Empty(t, r.Summary)
		assert.Empty(t, r.Value)
		assert.Empty(t, r.Extensions)

		m2[tags.Ref] = struct{}{}
		_, err = fromObj[Example](m2)
		require.Error(t, err)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if k != tags.Value && !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Example](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestCommonParameter_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:        "test name",
		tags.Description: "test description",
		tags.Required:    true,
		tags.In:          "test in",
		tags.Example:     "test example",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[CommonParameter](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test description", r.Description)
		assert.True(t, r.Required)
		assert.Equal(t, "test in", r.In)
		assert.Equal(t, "test example", r.Example)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if k != tags.Example && !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[CommonParameter](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestQueryParam_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:        "test name",
		tags.Description: "test description",
		tags.Required:    true,
		tags.In:          "test in",
		tags.Example:     "test example",
		"x-foo":          "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[QueryParam](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test description", r.Description)
		assert.True(t, r.Required)
		assert.Equal(t, "test in", r.In)
		assert.Equal(t, "test example", r.Example)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if k != tags.Example && !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[QueryParam](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestProperty_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:        "test name",
		tags.Description: "test description",
		tags.Type:        "test type",
		tags.Required:    true,
		tags.Format:      "test format",
		tags.Deprecated:  true,
		tags.Example:     "test example",
		tags.Enum:        []any{nil},
		tags.Properties: map[string]any{
			"foo": map[string]any{},
		},
		"x-foo": "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Property](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test type", r.Type)
		assert.Equal(t, "", r.ItemType)
		assert.True(t, r.Required)
		assert.Equal(t, "test format", r.Format)
		assert.True(t, r.Deprecated)
		assert.Equal(t, "test example", r.Example)
		assert.Len(t, r.Enum, 1)
		assert.Len(t, r.Extensions, 1)
		assert.Empty(t, r.SchemaRef)
	})
	t.Run("success ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Ref] = "test ref"
		r, err := fromObj[Property](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.SchemaRef)
		assert.Empty(t, r.Name)
		assert.Empty(t, r.Description)
		assert.Empty(t, r.Type)
		assert.Empty(t, r.ItemType)
		assert.False(t, r.Required)
		assert.Empty(t, r.Format)
		assert.False(t, r.Deprecated)
		assert.Empty(t, r.Example)
		assert.Empty(t, r.Enum)
		assert.Empty(t, r.Extensions)
	})
	t.Run("success items", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Type] = "object"
		m2[tags.Items] = map[string]any{
			tags.Type: "object",
			tags.Properties: map[string]any{
				"foo": map[string]any{
					tags.Type:     "string",
					tags.Required: true,
				},
			},
		}
		r, err := fromObj[Property](m2)
		require.NoError(t, err)
		assert.Len(t, r.Properties, 1)
		assert.Equal(t, "", r.Name)
		assert.Equal(t, "", r.Description)
		assert.Equal(t, "object", r.Type)
		assert.Equal(t, "object", r.ItemType)
		assert.False(t, r.Required)
		assert.Equal(t, "", r.Format)
		assert.False(t, r.Deprecated)
		assert.Nil(t, r.Example)
		assert.Empty(t, r.Enum)
		assert.Empty(t, r.Extensions)
		assert.Empty(t, r.SchemaRef)
	})
	t.Run("success items ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Type] = "object"
		m2[tags.Items] = map[string]any{
			tags.Ref: "test ref",
		}
		r, err := fromObj[Property](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.SchemaRef)
		assert.Empty(t, r.Name)
		assert.Empty(t, r.Description)
		assert.Equal(t, "object", r.Type)
		assert.Equal(t, "object", r.ItemType)
		assert.False(t, r.Required)
		assert.Empty(t, r.Format)
		assert.False(t, r.Deprecated)
		assert.Empty(t, r.Example)
		assert.Empty(t, r.Enum)
		assert.Empty(t, r.Extensions)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if k != tags.Example && !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Property](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestDiscriminator_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.PropertyName: "test property name",
		tags.Mapping: map[string]any{
			"foo": "bar",
		},
		"x-foo": "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Discriminator](m)
		require.NoError(t, err)
		assert.Equal(t, "test property name", r.PropertyName)
		assert.Len(t, r.Mapping, 1)
		assert.Len(t, r.Extensions, 1)
	})
	t.Run("bad mapping item", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Mapping] = map[string]any{
			"foo": false,
		}
		_, err := fromObj[Discriminator](m2)
		require.Error(t, err)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Discriminator](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestRequest_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Required:    true,
		/*
			tags.Content: map[string]any{
				contentTypeJson: map[string]any{
					tags.Schema: map[string]any{},
				},
			},
		*/
		tags.Examples: []any{
			map[string]any{},
		},
		"x-foo": "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Request](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.True(t, r.Required)
		assert.Len(t, r.Examples, 1)
		assert.Len(t, r.Extensions, 1)
		assert.Empty(t, r.Ref)
	})
	t.Run("success with content", func(t *testing.T) {
		m2 := map[string]any{
			tags.Description: "test description",
			tags.Required:    true,
			tags.Content: map[string]any{
				contentTypeJson: map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
			},
		}
		r, err := fromObj[Request](m2)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.True(t, r.Required)
		assert.Len(t, r.Extensions, 1)
		assert.Equal(t, contentTypeJson, r.ContentType)
		assert.Empty(t, r.AlternativeContentTypes)
		assert.Empty(t, r.Ref)
	})
	t.Run("success with multi content", func(t *testing.T) {
		m2 := map[string]any{
			tags.Description: "test description",
			tags.Required:    true,
			tags.Content: map[string]any{
				contentTypeJson: map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
				"text/csv": map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
			},
		}
		r, err := fromObj[Request](m2)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.True(t, r.Required)
		assert.Len(t, r.Extensions, 1)
		assert.Equal(t, contentTypeJson, r.ContentType)
		assert.Len(t, r.AlternativeContentTypes, 1)
		_, ok := r.AlternativeContentTypes["text/csv"]
		assert.True(t, ok)
		assert.Empty(t, r.Ref)
	})
	t.Run("success ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Ref] = "test ref"
		r, err := fromObj[Request](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.Ref)
		assert.Empty(t, r.Description)
		assert.False(t, r.Required)
		assert.Empty(t, r.Examples)
		assert.Empty(t, r.Extensions)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Request](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestResponse_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Description: "test description",
		tags.Examples: []any{
			map[string]any{},
		},
		"x-foo": "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Response](m)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Len(t, r.Examples, 1)
		assert.Len(t, r.Extensions, 1)
		assert.Empty(t, r.Ref)
	})
	t.Run("success with content", func(t *testing.T) {
		m2 := map[string]any{
			tags.Description: "test description",
			tags.Content: map[string]any{
				contentTypeJson: map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
			},
		}
		r, err := fromObj[Response](m2)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Len(t, r.Extensions, 1)
		assert.Equal(t, contentTypeJson, r.ContentType)
		assert.Empty(t, r.AlternativeContentTypes)
		assert.Empty(t, r.Ref)
	})
	t.Run("success with multi content", func(t *testing.T) {
		m2 := map[string]any{
			tags.Description: "test description",
			tags.Content: map[string]any{
				contentTypeJson: map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
				"text/csv": map[string]any{
					tags.Schema: map[string]any{},
					"x-foo":     "bar",
				},
			},
		}
		r, err := fromObj[Response](m2)
		require.NoError(t, err)
		assert.Equal(t, "test description", r.Description)
		assert.Len(t, r.Extensions, 1)
		assert.Equal(t, contentTypeJson, r.ContentType)
		assert.Len(t, r.AlternativeContentTypes, 1)
		_, ok := r.AlternativeContentTypes["text/csv"]
		assert.True(t, ok)
		assert.Empty(t, r.Ref)
	})
	t.Run("success ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Ref] = "test ref"
		r, err := fromObj[Response](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.Ref)
		assert.Empty(t, r.Description)
		assert.Empty(t, r.Examples)
		assert.Empty(t, r.Extensions)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Request](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestSchema_unmarshalObj(t *testing.T) {
	m := map[string]any{
		tags.Name:        "test name",
		tags.Description: "test description",
		tags.Type:        "test type",
		tags.Format:      "test format",
		tags.Required:    []any{"test required"},
		tags.Properties: map[string]any{
			"foo": map[string]any{},
		},
		tags.Discriminator: map[string]any{},
		tags.Default:       "test default",
		tags.Example:       "test example",
		"x-foo":            "bar",
	}
	t.Run("success", func(t *testing.T) {
		r, err := fromObj[Schema](m)
		require.NoError(t, err)
		assert.Equal(t, "test name", r.Name)
		assert.Equal(t, "test description", r.Description)
		assert.Equal(t, "test type", r.Type)
		assert.Equal(t, "test format", r.Format)
		assert.Len(t, r.RequiredProperties, 1)
		assert.Len(t, r.Properties, 1)
		assert.NotNil(t, r.Discriminator)
		assert.Equal(t, "test default", r.Default)
		assert.Equal(t, "test example", r.Example)
		assert.Len(t, r.Extensions, 1)
		assert.Empty(t, r.SchemaRef)
	})
	t.Run("success ref", func(t *testing.T) {
		m2 := maps.Clone(m)
		m2[tags.Ref] = "test ref"
		r, err := fromObj[Schema](m2)
		require.NoError(t, err)
		assert.Equal(t, "test ref", r.SchemaRef)
		assert.Empty(t, r.Name)
		assert.Empty(t, r.Description)
		assert.Empty(t, r.Type)
		assert.Empty(t, r.Format)
		assert.Empty(t, r.RequiredProperties)
		assert.Empty(t, r.Properties)
		assert.Nil(t, r.Discriminator)
		assert.Empty(t, r.Default)
		assert.Empty(t, r.Example)
		assert.Empty(t, r.Extensions)
	})
	t.Run("errors", func(t *testing.T) {
		type bad struct{}
		for k := range m {
			if k != tags.Example && k != tags.Default && !strings.HasPrefix(k, "x-") {
				m2 := maps.Clone(m)
				m2[k] = bad{}
				_, err := fromObj[Schema](m2)
				require.Error(t, err)
			}
		}
	})
}

func TestOfsFrom(t *testing.T) {
	t.Run("success oneOf ref", func(t *testing.T) {
		m := map[string]any{
			tags.OneOf: []any{
				map[string]any{
					tags.Ref: "test ref",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, OneOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.True(t, ofs.Of[0].IsRef())
		assert.Equal(t, "test ref", ofs.Of[0].Ref())
		assert.Nil(t, ofs.Of[0].Schema())
	})
	t.Run("success oneOf schema", func(t *testing.T) {
		m := map[string]any{
			tags.OneOf: []any{
				map[string]any{
					tags.Description: "test description",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, OneOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.False(t, ofs.Of[0].IsRef())
		assert.NotNil(t, ofs.Of[0].Schema())
	})
	t.Run("success anyOf ref", func(t *testing.T) {
		m := map[string]any{
			tags.AnyOf: []any{
				map[string]any{
					tags.Ref: "test ref",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, AnyOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.True(t, ofs.Of[0].IsRef())
		assert.Equal(t, "test ref", ofs.Of[0].Ref())
		assert.Nil(t, ofs.Of[0].Schema())
	})
	t.Run("success anyOf schema", func(t *testing.T) {
		m := map[string]any{
			tags.AnyOf: []any{
				map[string]any{
					tags.Description: "test description",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, AnyOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.False(t, ofs.Of[0].IsRef())
		assert.NotNil(t, ofs.Of[0].Schema())
	})
	t.Run("success allOf ref", func(t *testing.T) {
		m := map[string]any{
			tags.AllOf: []any{
				map[string]any{
					tags.Ref: "test ref",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, AllOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.True(t, ofs.Of[0].IsRef())
		assert.Equal(t, "test ref", ofs.Of[0].Ref())
		assert.Nil(t, ofs.Of[0].Schema())
	})
	t.Run("success allOf schema", func(t *testing.T) {
		m := map[string]any{
			tags.AllOf: []any{
				map[string]any{
					tags.Description: "test description",
				},
			},
		}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, ofs)
		assert.Equal(t, AllOf, ofs.OfType)
		assert.Len(t, ofs.Of, 1)
		assert.False(t, ofs.Of[0].IsRef())
		assert.NotNil(t, ofs.Of[0].Schema())
	})
	t.Run("failure ref", func(t *testing.T) {
		m := map[string]any{
			tags.AllOf: []any{
				map[string]any{
					tags.Ref: true,
				},
			},
		}
		_, err := ofsFrom(m)
		require.Error(t, err)
	})
	t.Run("failure schema", func(t *testing.T) {
		m := map[string]any{
			tags.AllOf: []any{
				map[string]any{
					tags.Description: true,
				},
			},
		}
		_, err := ofsFrom(m)
		require.Error(t, err)
	})
	t.Run("none", func(t *testing.T) {
		m := map[string]any{}
		ofs, err := ofsFrom(m)
		require.NoError(t, err)
		assert.Nil(t, ofs)
	})
	t.Run("not array", func(t *testing.T) {
		m := map[string]any{
			tags.OneOf: "not an array",
		}
		_, err := ofsFrom(m)
		require.Error(t, err)
	})
	t.Run("invalid element", func(t *testing.T) {
		m := map[string]any{
			tags.OneOf: []any{"some ref"},
		}
		_, err := ofsFrom(m)
		require.Error(t, err)
	})
}

func TestServersFrom(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := map[string]any{
			tags.Servers: []any{
				map[string]any{
					tags.Url: "test url",
				},
			},
		}
		severs, err := serversFrom(m)
		require.NoError(t, err)
		assert.Len(t, severs, 1)
		_, ok := severs["test url"]
		assert.True(t, ok)
	})
	t.Run("none", func(t *testing.T) {
		m := map[string]any{}
		servers, err := serversFrom(m)
		require.NoError(t, err)
		assert.Nil(t, servers)
	})
	t.Run("invalid url", func(t *testing.T) {
		m := map[string]any{
			tags.Servers: []any{
				map[string]any{
					tags.Url: false,
				},
			},
		}
		_, err := serversFrom(m)
		require.Error(t, err)
	})
	t.Run("invalid sever", func(t *testing.T) {
		m := map[string]any{
			tags.Servers: []any{
				map[string]any{
					tags.Url:         "test url",
					tags.Description: false,
				},
			},
		}
		_, err := serversFrom(m)
		require.Error(t, err)
	})
	t.Run("invalid element", func(t *testing.T) {
		m := map[string]any{
			tags.Servers: []any{
				"not an object",
			},
		}
		_, err := serversFrom(m)
		require.Error(t, err)
	})
	t.Run("not an array", func(t *testing.T) {
		m := map[string]any{
			tags.Servers: "not an array",
		}
		_, err := serversFrom(m)
		require.Error(t, err)
	})
}

func TestSecurityFrom(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := map[string]any{
			tags.Security: []any{
				map[string]any{
					"test": []any{"foo", "bar"},
				},
			},
		}
		secs, err := securityFrom(m)
		require.NoError(t, err)
		assert.Len(t, secs, 1)
		assert.Equal(t, "test", secs[0].Name)
		assert.Equal(t, []string{"foo", "bar"}, secs[0].Scopes)
	})
	t.Run("non-string scope", func(t *testing.T) {
		m := map[string]any{
			tags.Security: []any{
				map[string]any{
					"test": []any{false},
				},
			},
		}
		_, err := securityFrom(m)
		require.Error(t, err)
	})
	t.Run("not array", func(t *testing.T) {
		m := map[string]any{
			tags.Security: "not an array",
		}
		_, err := securityFrom(m)
		require.Error(t, err)
	})
	t.Run("invalid element 1", func(t *testing.T) {
		m := map[string]any{
			tags.Security: []any{
				"not an object",
			},
		}
		_, err := securityFrom(m)
		require.Error(t, err)
	})
	t.Run("invalid element 2", func(t *testing.T) {
		m := map[string]any{
			tags.Security: []any{
				map[string]any{
					"test": nil,
				},
			},
		}
		_, err := securityFrom(m)
		require.Error(t, err)
	})
	t.Run("none", func(t *testing.T) {
		m := map[string]any{}
		secs, err := securityFrom(m)
		require.NoError(t, err)
		assert.Nil(t, secs)
	})
}

func TestComponentsFrom(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.Schemas: map[string]any{
					"test": map[string]any{},
				},
				tags.SecuritySchemes: map[string]any{
					"test": map[string]any{},
				},
				tags.Examples: map[string]any{
					"test": map[string]any{},
				},
				tags.Parameters: map[string]any{
					"test": map[string]any{},
				},
				tags.RequestBodies: map[string]any{
					"test": map[string]any{},
				},
				tags.Responses: map[string]any{
					"test": map[string]any{},
				},
				"x-foo": "bar",
			},
		}
		c, err := componentsFrom(m)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.Len(t, c.Schemas, 1)
		assert.Len(t, c.SecuritySchemes, 1)
		assert.Len(t, c.Examples, 1)
		assert.Len(t, c.Parameters, 1)
		assert.Len(t, c.Requests, 1)
		assert.Len(t, c.Responses, 1)
		assert.Len(t, c.Extensions, 1)
	})
	t.Run("bad schema", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.Schemas: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("bad security scheme", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.SecuritySchemes: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("bad example", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.Examples: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("bad parameter", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.Parameters: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("bad request", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.RequestBodies: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("bad response", func(t *testing.T) {
		m := map[string]any{
			tags.Components: map[string]any{
				tags.Responses: map[string]any{
					"test": map[string]any{
						tags.Description: true,
					},
				},
			},
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
	t.Run("none", func(t *testing.T) {
		m := map[string]any{}
		c, err := componentsFrom(m)
		require.NoError(t, err)
		assert.Nil(t, c)
	})
	t.Run("not an object", func(t *testing.T) {
		m := map[string]any{
			tags.Components: false,
		}
		_, err := componentsFrom(m)
		require.Error(t, err)
	})
}

func TestContentType_unmarshalObj(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Description: "test description",
				tags.Type:        "object",
			},
			tags.Examples: map[string]any{
				"foo": map[string]any{},
			},
			"x-foo": "bar",
		}
		r, err := fromObj[contentType](m)
		require.NoError(t, err)
		assert.False(t, r.isArray)
		assert.NotNil(t, r.schema)
		assert.Equal(t, "test description", r.schema.Description)
		assert.Equal(t, "object", r.schema.Type)
		assert.Len(t, r.examples, 1)
		assert.Len(t, r.extensions, 1)
	})
	t.Run("success with ref", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Ref: "some ref",
			},
		}
		r, err := fromObj[contentType](m)
		require.NoError(t, err)
		assert.False(t, r.isArray)
		assert.Nil(t, r.schema)
		assert.Equal(t, "some ref", r.ref)
	})
	t.Run("success with items", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Type: "array",
				tags.Items: map[string]any{
					tags.Description: "test description",
				},
			},
		}
		r, err := fromObj[contentType](m)
		require.NoError(t, err)
		assert.True(t, r.isArray)
		assert.Equal(t, "array", r.xType)
		assert.Equal(t, "test description", r.schema.Description)
		assert.NotNil(t, r.schema)
	})
	t.Run("fails with items not array type", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Type: "object",
				tags.Items: map[string]any{
					tags.Description: "test description",
				},
			},
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("fails array type no items", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Type: "array",
			},
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("fails items not an object", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Type:  "array",
				tags.Items: "not an object",
			},
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("missing schema", func(t *testing.T) {
		m := map[string]any{}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("bad schema", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: "not an object",
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("bad schema type", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Type: false,
			},
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
	t.Run("bad example", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Description: "test description",
			},
			tags.Examples: map[string]any{
				"foo": map[string]any{
					tags.Description: true,
				},
			},
		}
		_, err := fromObj[contentType](m)
		require.Error(t, err)
	})
}

func Test_schemaFrom(t *testing.T) {
	t.Run("success ref", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Ref: "test ref",
			},
		}
		ref, schema, err := schemaFrom(m)
		require.NoError(t, err)
		assert.Equal(t, "test ref", ref)
		assert.Nil(t, schema)
	})
	t.Run("success schema", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{},
		}
		ref, schema, err := schemaFrom(m)
		require.NoError(t, err)
		assert.Empty(t, ref)
		assert.NotNil(t, schema)
	})
	t.Run("failure not object", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: "not an object",
		}
		_, _, err := schemaFrom(m)
		require.Error(t, err)
	})
	t.Run("failure ref", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Ref: struct{}{},
			},
		}
		_, _, err := schemaFrom(m)
		require.Error(t, err)
	})
	t.Run("failure schema", func(t *testing.T) {
		m := map[string]any{
			tags.Schema: map[string]any{
				tags.Name: struct{}{},
			},
		}
		_, _, err := schemaFrom(m)
		require.Error(t, err)
	})
}

func TestObjFromProperty(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r, err := objFromProperty[Request](map[string]any{"request": map[string]any{}}, "request")
		require.NoError(t, err)
		require.NotNil(t, r)
	})
	t.Run("nil", func(t *testing.T) {
		r, err := objFromProperty[Request](map[string]any{}, "request")
		require.NoError(t, err)
		require.Nil(t, r)
	})
	t.Run("bad", func(t *testing.T) {
		type bad struct{}
		_, err := objFromProperty[bad](map[string]any{"request": map[string]any{}}, "request")
		require.Error(t, err)
	})
	t.Run("not object", func(t *testing.T) {
		_, err := objFromProperty[Request](map[string]any{"request": "not an object"}, "request")
		require.Error(t, err)
	})
}

func TestFromObj(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r, err := fromObj[Request](map[string]any{})
		require.NoError(t, err)
		require.NotNil(t, r)
	})
	t.Run("bad", func(t *testing.T) {
		type bad struct{}
		_, err := fromObj[bad](map[string]any{})
		require.Error(t, err)
	})
}

func TestSliceFromProperty(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		r, err := sliceFromProperty[Request](map[string]any{}, "request")
		require.NoError(t, err)
		require.Nil(t, r)
	})
	t.Run("ok", func(t *testing.T) {
		r, err := sliceFromProperty[Request](map[string]any{"request": []any{map[string]any{}}}, "request")
		require.NoError(t, err)
		require.Len(t, r, 1)
	})
	t.Run("bad", func(t *testing.T) {
		type bad struct{}
		_, err := sliceFromProperty[bad](map[string]any{"request": []any{map[string]any{}}}, "request")
		require.Error(t, err)
	})
	t.Run("invalid element", func(t *testing.T) {
		_, err := sliceFromProperty[Request](map[string]any{"request": []any{"not an object"}}, "request")
		require.Error(t, err)
	})
	t.Run("not array", func(t *testing.T) {
		_, err := sliceFromProperty[Request](map[string]any{"request": nil}, "request")
		require.Error(t, err)
	})
}

func TestNamedSliceFromProperty(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		r, n, err := namedSliceFromProperty[Request](map[string]any{}, "request")
		require.NoError(t, err)
		require.Nil(t, r)
		require.Nil(t, n)
	})
	t.Run("ok map[string]any", func(t *testing.T) {
		r, n, err := namedSliceFromProperty[Request](map[string]any{"request": map[string]any{"name": map[string]any{}}}, "request")
		require.NoError(t, err)
		require.Len(t, r, 1)
		require.Len(t, n, 1)
		assert.Equal(t, "name", n[0])
	})
	t.Run("ok map[any]any", func(t *testing.T) {
		r, n, err := namedSliceFromProperty[Request](map[string]any{"request": map[any]any{400: map[string]any{}}}, "request")
		require.NoError(t, err)
		require.Len(t, r, 1)
		require.Len(t, n, 1)
		assert.Equal(t, "400", n[0])
	})
	t.Run("not a map", func(t *testing.T) {
		_, _, err := namedSliceFromProperty[Request](map[string]any{"request": "not an object"}, "request")
		require.Error(t, err)
	})
	t.Run("bad map[string]any", func(t *testing.T) {
		type bad struct{}
		_, _, err := namedSliceFromProperty[bad](map[string]any{"request": map[string]any{"name": map[string]any{}}}, "request")
		require.Error(t, err)
	})
	t.Run("bad map[any]any", func(t *testing.T) {
		type bad struct{}
		_, _, err := namedSliceFromProperty[bad](map[string]any{"request": map[any]any{400: map[string]any{}}}, "request")
		require.Error(t, err)
	})
	t.Run("invalid element map[string]any", func(t *testing.T) {
		_, _, err := namedSliceFromProperty[Request](map[string]any{"request": map[string]any{"name": "not an object"}}, "request")
		require.Error(t, err)
	})
	t.Run("invalid element map[any]any", func(t *testing.T) {
		_, _, err := namedSliceFromProperty[Request](map[string]any{"request": map[any]any{400: "not an object"}}, "request")
		require.Error(t, err)
	})
}

func TestHasRef(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ref, ok, err := hasRef(map[string]any{tags.Ref: "some ref"})
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "some ref", ref)
	})
	t.Run("no", func(t *testing.T) {
		_, ok, err := hasRef(map[string]any{})
		require.NoError(t, err)
		require.False(t, ok)
	})
	t.Run("not string", func(t *testing.T) {
		_, _, err := hasRef(map[string]any{tags.Ref: true})
		require.Error(t, err)
	})
	t.Run("UnmarshalStrictRef", func(t *testing.T) {
		m := map[string]any{
			tags.Ref:         "some ref",
			"other property": "other value",
		}
		_, ok, err := hasRef(m)
		require.NoError(t, err)
		assert.True(t, ok)
		UnmarshalStrictRef = true
		defer func() { UnmarshalStrictRef = false }()
		_, _, err = hasRef(m)
		require.Error(t, err)
	})
}

func TestStringFromProperty(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s, err := stringFromProperty(map[string]any{"pty": "some value"}, "pty")
		require.NoError(t, err)
		require.Equal(t, "some value", s)
	})
	t.Run("not there", func(t *testing.T) {
		s, err := stringFromProperty(map[string]any{}, "pty")
		require.NoError(t, err)
		require.Equal(t, "", s)
	})
	t.Run("not string", func(t *testing.T) {
		_, err := stringFromProperty(map[string]any{"pty": true}, "pty")
		require.Error(t, err)
	})
}

func TestJsonNumberFromProperty(t *testing.T) {
	testCases := []struct {
		value     any
		expectErr bool
		expect    string
	}{
		{
			value:     nil,
			expectErr: true,
		},
		{
			value:  json.Number("1"),
			expect: "1",
		},
		{
			value:  "1",
			expect: "1",
		},
		{
			value:  1,
			expect: "1",
		},
		{
			value:  float32(1.1),
			expect: "1.1",
		},
		{
			value:  1.1,
			expect: "1.1",
		},
		{
			value:     math.NaN(),
			expectErr: true,
		},
		{
			value:     math.Inf(-1),
			expectErr: true,
		},
		{
			value:     float32(math.NaN()),
			expectErr: true,
		},
		{
			value:     float32(math.Inf(-1)),
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			m := map[string]any{
				"value": tc.value,
			}
			jn, err := jsonNumberFromProperty(m, "value")
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, json.Number(tc.expect), jn)
			}
		})
	}
	t.Run("not there", func(t *testing.T) {
		jn, err := jsonNumberFromProperty(map[string]any{}, "value")
		require.NoError(t, err)
		require.Equal(t, json.Number(""), jn)
	})
}

func TestUintFromProperty(t *testing.T) {
	testCases := []struct {
		value     any
		expectErr bool
		expect    uint
	}{
		{
			value:     nil,
			expectErr: true,
		},
		{
			value:  uint(1),
			expect: 1,
		},
		{
			value:  uint8(1),
			expect: 1,
		},
		{
			value:  uint16(1),
			expect: 1,
		},
		{
			value:  uint32(1),
			expect: 1,
		},
		{
			value:  uint64(1),
			expect: 1,
		},
		{
			value:  int(1),
			expect: 1,
		},
		{
			value:  int8(1),
			expect: 1,
		},
		{
			value:  int16(1),
			expect: 1,
		},
		{
			value:  int32(1),
			expect: 1,
		},
		{
			value:  int64(1),
			expect: 1,
		},
		{
			value:     -1,
			expectErr: true,
		},
		{
			value:  json.Number("1"),
			expect: 1,
		},
		{
			value:     json.Number("-1"),
			expectErr: true,
		},
		{
			value:  "1",
			expect: 1,
		},
		{
			value:     "-1",
			expectErr: true,
		},
		{
			value:  1.1,
			expect: 1,
		},
		{
			value:     math.Inf(-1),
			expectErr: true,
		},
		{
			value:     math.NaN(),
			expectErr: true,
		},
		{
			value:  float32(1.1),
			expect: 1,
		},
		{
			value:     float32(math.Inf(-1)),
			expectErr: true,
		},
		{
			value:     float32(math.NaN()),
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			m := map[string]any{
				"value": tc.value,
			}
			n, err := uintFromProperty(m, "value")
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, n)
			}
		})
	}
	t.Run("not there", func(t *testing.T) {
		n, err := uintFromProperty(map[string]any{}, "value")
		require.NoError(t, err)
		require.Equal(t, uint(0), n)
	})
}

func TestStringsSliceFromProperty(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s, err := stringsSliceFromProperty(map[string]any{"value": []any{"test"}}, "value")
		require.NoError(t, err)
		require.Len(t, s, 1)
		require.Equal(t, "test", s[0])
	})
	t.Run("not there", func(t *testing.T) {
		s, err := stringsSliceFromProperty(map[string]any{}, "value")
		require.NoError(t, err)
		require.Nil(t, s)
	})
	t.Run("not array", func(t *testing.T) {
		_, err := stringsSliceFromProperty(map[string]any{"value": "not an array"}, "value")
		require.Error(t, err)
	})
	t.Run("invalid element", func(t *testing.T) {
		_, err := stringsSliceFromProperty(map[string]any{"value": []any{true}}, "value")
		require.Error(t, err)
	})
}

func TestAnySliceFromProperty(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s, err := anySliceFromProperty(map[string]any{"value": []any{"test"}}, "value")
		require.NoError(t, err)
		require.Len(t, s, 1)
		require.Equal(t, "test", s[0])
	})
	t.Run("not there", func(t *testing.T) {
		s, err := anySliceFromProperty(map[string]any{}, "value")
		require.NoError(t, err)
		require.Nil(t, s)
	})
	t.Run("not array", func(t *testing.T) {
		_, err := anySliceFromProperty(map[string]any{"value": "not an array"}, "value")
		require.Error(t, err)
	})
}

func TestBooleanFromProperty(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		b, err := booleanFromProperty(map[string]any{"value": true}, "value")
		require.NoError(t, err)
		require.True(t, b)
	})
	t.Run("ok (string)", func(t *testing.T) {
		b, err := booleanFromProperty(map[string]any{"value": "true"}, "value")
		require.NoError(t, err)
		require.True(t, b)
	})
	t.Run("not there", func(t *testing.T) {
		b, err := booleanFromProperty(map[string]any{}, "value")
		require.NoError(t, err)
		require.False(t, b)
	})
	t.Run("invalid value", func(t *testing.T) {
		_, err := booleanFromProperty(map[string]any{"value": 1}, "value")
		require.Error(t, err)
	})
}

func TestExtensionsFrom(t *testing.T) {
	ex := extensionsFrom(map[string]any{
		"foo": "bar",
		"x-1": 1,
		"x-2": "2",
	})
	require.Len(t, ex, 2)
	require.Equal(t, 1, ex["1"])
	require.Equal(t, "2", ex["2"])
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
	p, ok = p.Paths["/{id}"]
	require.True(t, ok)
	require.Len(t, p.PathParams, 1)
	m, ok := p.Methods["GET"]
	require.True(t, ok)
	require.Len(t, m.QueryParams, 0)
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
						Extensions: Extensions{
							"foo": "bar",
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
						Extensions: Extensions{
							"foo": "bar",
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
        - $ref: "#/components/parameters/id"
      WAS-parameters:
        - name: id
          description: "id path param"
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
          description: "id path param"
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
          description: "id path param"
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
          description: "id path param"
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
components:
  parameters:
    id:
      name: id
      description: common id param
      in: path
      required: true
      example: example
`
