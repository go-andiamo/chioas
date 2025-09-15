package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/format"
	"net/http"
	"strings"
	"testing"
)

func TestGenerateCode_Definition(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		def := chioas.Definition{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("with methods", func(t *testing.T) {
		def := chioas.Definition{
			Methods: chioas.Methods{
				http.MethodPost: {},
				http.MethodGet:  {},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true, UseHttpConsts: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"net/http"

	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Methods: chioas.Methods{
		http.MethodGet: {
		},
		http.MethodPost: {
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with paths (non hoisted)", func(t *testing.T) {
		def := chioas.Definition{
			Paths: chioas.Paths{
				"/foo": {
					Paths: chioas.Paths{
						"/bar": {
							Methods: chioas.Methods{
								http.MethodGet: {},
							},
						},
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true, UseHttpConsts: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"net/http"

	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Paths: chioas.Paths{
		"/foo": {
			Paths: chioas.Paths{
				"/bar": {
					Methods: chioas.Methods{
						http.MethodGet: {
						},
					},
				},
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with paths hoisted", func(t *testing.T) {
		def := chioas.Definition{
			Paths: chioas.Paths{
				"/": {
					Paths: chioas.Paths{
						"/foos": {
							Methods: chioas.Methods{
								http.MethodGet: {},
							},
						},
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true, UseHttpConsts: true, HoistPaths: true, PublicVars: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"net/http"

	"github.com/go-andiamo/chioas"
)

var Definition = chioas.Definition{
	Paths: chioas.Paths{
		"/": PathRoot,
	},
}

var (
	PathRoot = chioas.Path{
		Paths: chioas.Paths{
			"/foos": {
				Methods: chioas.Methods{
					http.MethodGet: {
					},
				},
			},
		},
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with multiple duplicate paths hoisted", func(t *testing.T) {
		def := chioas.Definition{
			Paths: chioas.Paths{
				"/foo": {
					Tag: "1st",
				},
				"/Foo": {
					Tag: "2nd",
				},
				"/foo2": {
					Tag: "3rd",
				},
				"Foo2": {
					Tag: "4th",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true, UseHttpConsts: true, HoistPaths: true, PublicVars: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"net/http"

	"github.com/go-andiamo/chioas"
)

var Definition = chioas.Definition{
	Paths: chioas.Paths{
		"/Foo": PathFoo,
		"/foo": PathFoo2,
		"/foo2": PathFoo2_2,
		"Foo2": PathFoo2_3,
	},
}

var (
	PathFoo = chioas.Path{
		Tag: "2nd",
	}
	PathFoo2 = chioas.Path{
		Tag: "1st",
	}
	PathFoo2_2 = chioas.Path{
		Tag: "3rd",
	}
	PathFoo2_3 = chioas.Path{
		Tag: "4th",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with zero values", func(t *testing.T) {
		def := chioas.Definition{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Comment: "",
	AutoHeadMethods: false,
	AutoOptionsMethods: false,
	RootAutoOptionsMethod: false,
	AutoMethodNotAllowed: false,
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with info", func(t *testing.T) {
		def := chioas.Definition{
			Info: chioas.Info{
				Title:          "title",
				Description:    "description",
				Version:        "semver",
				TermsOfService: "terms",
				Contact: &chioas.Contact{
					Name: "contact\nname",
				},
				License: &chioas.License{
					Name: "MIT",
				},
				Extensions: chioas.Extensions{
					"foo": "bar",
				},
				ExternalDocs: &chioas.ExternalDocs{
					Description: "ext\ndocs\ndescription",
				},
				Comment: "comment",
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
		Info: chioas.Info{
			Title: "title",
			Description: "description",
			Version: "semver",
			TermsOfService: "terms",
			Contact: &chioas.Contact{
				Name: "contact\nname",
			},
			License: &chioas.License{
				Name: "MIT",
			},
			Extensions: chioas.Extensions{
				"foo": "bar",
			},
			ExternalDocs: &chioas.ExternalDocs{
				Description: "ext\ndocs\ndescription",
			},
			Comment: "comment",
		},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with servers", func(t *testing.T) {
		def := chioas.Definition{
			Servers: chioas.Servers{
				"example.com": {
					Description: "example.com description",
					Extensions: chioas.Extensions{
						"foo": "bar",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Servers: chioas.Servers{
		"example.com": {
			Description: "example.com description",
			Extensions: chioas.Extensions{
				"foo": "bar",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with tags", func(t *testing.T) {
		def := chioas.Definition{
			Tags: chioas.Tags{
				{
					Name:        "tag1",
					Description: "description 1",
					Extensions:  chioas.Extensions{"foo": "bar"},
				},
				{
					Name:        "tag2",
					Description: "description 2",
					ExternalDocs: &chioas.ExternalDocs{
						Description: "ext\ndocs\ndescription",
						Extensions:  chioas.Extensions{"foo": "bar"},
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Tags: chioas.Tags{
		{
			Name: "tag1",
			Description: "description 1",
			Extensions: chioas.Extensions{
				"foo": "bar",
			},
		},
		{
			Name: "tag2",
			Description: "description 2",
			ExternalDocs: &chioas.ExternalDocs{
				Description: "ext\ndocs\ndescription",
				Extensions: chioas.Extensions{
					"foo": "bar",
				},
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with security", func(t *testing.T) {
		def := chioas.Definition{
			Security: chioas.SecuritySchemes{
				{
					Name:        "name",
					Description: "desc",
					Type:        "apiKey",
					Scheme:      "scheme",
					ParamName:   "X-API-KEY",
					In:          "header",
					Scopes:      []string{"a", "b"},
					Extensions:  chioas.Extensions{"foo": "bar"},
					Comment:     "comment",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Security: chioas.SecuritySchemes{
		{
			Name: "name",
			Description: "desc",
			Type: "apiKey",
			Scheme: "scheme",
			ParamName: "X-API-KEY",
			In: "header",
			Scopes: []string{
				"a",
				"b",
			},
			Extensions: chioas.Extensions{
				"foo": "bar",
			},
			Comment: "comment",
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components schemas", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Schemas: chioas.Schemas{
					{
						Name: "schema2",
					},
					{
						Name: "schema1",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		Schemas: chioas.Schemas{
			{
				Name: "schema1",
			},
			{
				Name: "schema2",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components requests", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Requests: chioas.CommonRequests{
					"reqB": {
						Description: "desc B",
					},
					"reqA": {
						Description: "desc A",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		Requests: chioas.CommonRequests{
			"reqA": {
				Description: "desc A",
			},
			"reqB": {
				Description: "desc B",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components responses", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Responses: chioas.CommonResponses{
					"reqB": {
						Description: "desc B",
					},
					"reqA": {
						Description: "desc A",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		Responses: chioas.CommonResponses{
			"reqA": {
				Description: "desc A",
			},
			"reqB": {
				Description: "desc B",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components examples", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Examples: chioas.Examples{
					{
						Name: "example2",
					},
					{
						Name: "example1",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		Examples: chioas.Examples{
			{
				Name: "example1",
			},
			{
				Name: "example2",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components parameters", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Parameters: chioas.CommonParameters{
					"paramB": {
						Example: "exampleB",
						Schema:  &chioas.Schema{},
					},
					"paramA": {
						SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		Parameters: chioas.CommonParameters{
			"paramA": {
				SchemaRef: "bar",
			},
			"paramB": {
				Example: "exampleB",
				Schema: &chioas.Schema{
				},
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components security schemes", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				SecuritySchemes: chioas.SecuritySchemes{
					{
						Name: "secB",
					},
					{
						Name: "secA",
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: &chioas.Components{
		SecuritySchemes: chioas.SecuritySchemes{
			{
					Name: "secA",
			},
			{
					Name: "secB",
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with components hoisted empty", func(t *testing.T) {
		def := chioas.Definition{
			Components: &chioas.Components{
				Extensions: chioas.Extensions{"foo": "bar"},
				Comment:    "comment",
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true, HoistComponents: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Definition{
	Components: components,
}

var (
	components = &chioas.Components{
		Extensions: chioas.Extensions{
			"foo": "bar",
		},
		Comment: "comment",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
}

func Test_generateComponentsVars(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		def := &chioas.Components{}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	components = &chioas.Components{
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with basics", func(t *testing.T) {
		def := &chioas.Components{
			Extensions: chioas.Extensions{"foo": "bar"},
			Comment:    "comment",
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Extensions: chioas.Extensions{
			"foo": "bar",
		},
		Comment: "comment",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with schemas", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "testB",
				},
				{
					Name: "testA",
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Schemas: chioas.Schemas{
			SchemaTestA,
			SchemaTestB,
		},
	}
	SchemaTestA = chioas.Schema{
		Name: "testA",
	}
	SchemaTestB = chioas.Schema{
		Name: "testB",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with requests", func(t *testing.T) {
		def := &chioas.Components{
			Requests: chioas.CommonRequests{
				"testB": {},
				"testA": {},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Requests: chioas.CommonRequests{
			"testA": RequestTestA,
			"testB": RequestTestB,
		},
	}
	RequestTestA = chioas.Request{
	}
	RequestTestB = chioas.Request{
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with responses", func(t *testing.T) {
		def := &chioas.Components{
			Responses: chioas.CommonResponses{
				"testB": {},
				"testA": {},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Responses: chioas.CommonResponses{
			"testA": ResponseTestA,
			"testB": ResponseTestB,
		},
	}
	ResponseTestA = chioas.Response{
	}
	ResponseTestB = chioas.Response{
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with examples", func(t *testing.T) {
		def := &chioas.Components{
			Examples: chioas.Examples{
				{
					Name: "testB",
				},
				{
					Name: "testA",
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Examples: chioas.Examples{
			ExampleTestA,
			ExampleTestB,
		},
	}
	ExampleTestA = chioas.Example{
		Name: "testA",
	}
	ExampleTestB = chioas.Example{
		Name: "testB",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with parameters", func(t *testing.T) {
		def := &chioas.Components{
			Parameters: chioas.CommonParameters{
				"testB": {},
				"testA": {},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Parameters: chioas.CommonParameters{
			"testA": ParameterTestA,
			"testB": ParameterTestB,
		},
	}
	ParameterTestA = chioas.CommonParameter{
	}
	ParameterTestB = chioas.CommonParameter{
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with security schemes", func(t *testing.T) {
		def := &chioas.Components{
			SecuritySchemes: chioas.SecuritySchemes{
				{
					Name: "testB",
				},
				{
					Name: "testA",
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		SecuritySchemes: chioas.SecuritySchemes{
			SecuritySchemeTestA,
			SecuritySchemeTestB,
		},
	}
	SecuritySchemeTestA = chioas.SecurityScheme{
		Name: "testA",
	}
	SecuritySchemeTestB = chioas.SecurityScheme{
		Name: "testB",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("with duplicate names", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "test2",
				},
				{
					Name: "test2",
				},
				{
					Name: "test",
				},
				{
					Name: "test",
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{PublicVars: true, OmitZeroValues: true})
		generateComponentsVars(def, w, false, false)
		require.NoError(t, w.err)
		const expect = `var (
	Components = &chioas.Components{
		Schemas: chioas.Schemas{
			SchemaTest,
			SchemaTest2,
			SchemaTest2_2,
			SchemaTest2_3,
		},
	}
	SchemaTest = chioas.Schema{
		Name: "test",
	}
	SchemaTest2 = chioas.Schema{
		Name: "test",
	}
	SchemaTest2_2 = chioas.Schema{
		Name: "test2",
	}
	SchemaTest2_3 = chioas.Schema{
		Name: "test2",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
}

func TestGenerateCode_Path(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		def := chioas.Path{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Path{
	Tag: "",
	HideDocs: false,
	Comment: "",
	AutoOptionsMethod: false,
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("empty with zero values", func(t *testing.T) {
		def := chioas.Path{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Path{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
}

func TestGenerateCode_OtherTypes(t *testing.T) {
	t.Run("definition ptr", func(t *testing.T) {
		def := &chioas.Definition{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = &chioas.Definition{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("path ptr", func(t *testing.T) {
		def := &chioas.Path{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = &chioas.Path{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("method", func(t *testing.T) {
		def := chioas.Method{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Method{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("method ptr", func(t *testing.T) {
		def := &chioas.Method{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = &chioas.Method{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("schema", func(t *testing.T) {
		def := chioas.Schema{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Schema{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("schema ptr", func(t *testing.T) {
		def := &chioas.Schema{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = &chioas.Schema{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("components", func(t *testing.T) {
		def := chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "test",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Components{
	Schemas: chioas.Schemas{
		{
			Name: "test",
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("components ptr", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "test",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = &chioas.Components{
	Schemas: chioas.Schemas{
		{
			Name: "test",
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("components hoisted", func(t *testing.T) {
		def := chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "test",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{HoistComponents: true, OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var (
	definition = chioas.Components{
		Schemas: chioas.Schemas{
			schemaTest,
		},
	}
	schemaTest = chioas.Schema{
		Name: "test",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("components ptr hoisted", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name: "test",
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{HoistComponents: true, OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var (
	definition = &chioas.Components{
		Schemas: chioas.Schemas{
			schemaTest,
		},
	}
	schemaTest = chioas.Schema{
		Name: "test",
	}
)

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("paths", func(t *testing.T) {
		def := chioas.Paths{
			"/api": {
				Methods: chioas.Methods{
					http.MethodGet: chioas.Method{},
				},
				Paths: chioas.Paths{
					"/foos": {
						Methods: chioas.Methods{
							http.MethodGet: chioas.Method{},
						},
					},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Paths{
	"/api": {
		Methods: chioas.Methods{
			"GET": {
			},
		},
		Paths: chioas.Paths{
			"/foos": {
				Methods: chioas.Methods{
					"GET": {
					},
				},
			},
		},
	},
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
	t.Run("paths empty", func(t *testing.T) {
		def := chioas.Paths{}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var definition = chioas.Paths{
}

`
		require.Equal(t, expect, buf.String())
		goFmtTest(t, buf.Bytes())
	})
}

func TestGenerateCode_Formatted(t *testing.T) {
	t.Run("components formatted", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name:        "test",
					Description: "test description",
					Type:        "object",
					Example:     map[string]any{"foo": "bar"},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{Format: true, HoistComponents: true, OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var (
	definition = &chioas.Components{
		Schemas: chioas.Schemas{
			schemaTest,
		},
	}
	schemaTest = chioas.Schema{
		Name:        "test",
		Description: "test description",
		Type:        "object",
		Example: map[string]any{
			"foo": "bar",
		},
	}
)
`
		require.Equal(t, expect, buf.String())
	})
	t.Run("components formatted CRLF", func(t *testing.T) {
		def := &chioas.Components{
			Schemas: chioas.Schemas{
				{
					Name:        "test",
					Description: "test description",
					Type:        "object",
					Example:     map[string]any{"foo": "bar"},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateCode(def, &buf, Options{Format: true, HoistComponents: true, UseCRLF: true, OmitZeroValues: true})
		require.NoError(t, err)
		const expect = `package api

import (
	"github.com/go-andiamo/chioas"
)

var (
	definition = &chioas.Components{
		Schemas: chioas.Schemas{
			schemaTest,
		},
	}
	schemaTest = chioas.Schema{
		Name:        "test",
		Description: "test description",
		Type:        "object",
		Example: map[string]any{
			"foo": "bar",
		},
	}
)
`
		require.Equal(t, strings.ReplaceAll(expect, "\n", "\r\n"), buf.String())
	})
}

func Test_generatePath(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		path := chioas.Path{}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{})
		generatePath(0, path, w)
		require.NoError(t, w.err)
		const expect = `	Tag: "",
	HideDocs: false,
	Comment: "",
	AutoOptionsMethod: false,
`
		require.Equal(t, expect, buf.String())
	})
	t.Run("empty omit zero values", func(t *testing.T) {
		path := chioas.Path{}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true})
		generatePath(0, path, w)
		require.NoError(t, w.err)
		const expect = ``
		require.Equal(t, expect, buf.String())
	})
	t.Run("with methods", func(t *testing.T) {
		path := chioas.Path{
			Methods: chioas.Methods{
				http.MethodPost: {},
				http.MethodGet:  {},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true, UseHttpConsts: true})
		generatePath(0, path, w)
		require.NoError(t, w.err)
		const expect = `	Methods: chioas.Methods{
		http.MethodGet: {
		},
		http.MethodPost: {
		},
	},
`
		require.Equal(t, expect, buf.String())
	})
	t.Run("with paths", func(t *testing.T) {
		path := chioas.Path{
			Paths: chioas.Paths{
				"/foo": {
					Paths: chioas.Paths{
						"/bar": {
							Methods: chioas.Methods{
								http.MethodGet: {},
							},
						},
					},
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true, UseHttpConsts: true})
		generatePath(0, path, w)
		require.NoError(t, w.err)
		const expect = `	Paths: chioas.Paths{
		"/foo": {
			Paths: chioas.Paths{
				"/bar": {
					Methods: chioas.Methods{
						http.MethodGet: {
						},
					},
				},
			},
		},
	},
`
		require.Equal(t, expect, buf.String())
	})
	t.Run("with path params", func(t *testing.T) {
		path := chioas.Path{
			PathParams: chioas.PathParams{
				"id2": {
					Description: "second",
					Example:     "x",
					SchemaRef:   refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
				"id1": {
					Description: "first",
					Schema: &chioas.Schema{
						Type: "string",
					},
				},
				"id3": {
					Description: "desc not seen",
					Ref:         refs.ComponentsPrefix + tags.Parameters + "/bar",
				},
			},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true})
		generatePath(0, path, w)
		require.NoError(t, w.err)
		const expect = `	PathParams: chioas.PathParams{
		"id1": {
			Description: "first",
			Schema: &chioas.Schema{
				Type: "string",
			},
		},
		"id2": {
			Description: "second",
			Example: "x",
			SchemaRef: "bar",
		},
		"id3": {
			Ref: "bar",
		},
	},
`
		require.Equal(t, expect, buf.String())
	})
}

func Test_generateMethods(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		methods := chioas.Methods{
			"PUT": {},
			"GET": {},
			"foo": {},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{OmitZeroValues: true})
		generateMethods(1, methods, w)
		require.NoError(t, w.err)
		const expect = `	Methods: chioas.Methods{
		"GET": {
		},
		"PUT": {
		},
		"FOO": {
		},
	},
`
		require.Equal(t, expect, buf.String())
	})
	t.Run("use http consts", func(t *testing.T) {
		methods := chioas.Methods{
			"PUT": {},
			"GET": {},
			"foo": {},
		}
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{UseHttpConsts: true, OmitZeroValues: true})
		generateMethods(1, methods, w)
		require.NoError(t, w.err)
		const expect = `	Methods: chioas.Methods{
		http.MethodGet: {
		},
		http.MethodPut: {
		},
		"FOO": {
		},
	},
`
		require.Equal(t, expect, buf.String())
	})
}

func Test_generateMethod(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		method  string
		def     chioas.Method
		expect  string
	}{
		{
			method: "GET",
			def:    chioas.Method{},
			expect: `"GET": {
	Description: "",
	Summary: "",
	OperationId: "",
	Tag: "",
	Deprecated: false,
	OptionalSecurity: false,
	Comment: "",
	HideDocs: false,
},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			indent:  1,
			method:  "GET",
			def:     chioas.Method{},
			expect: `	"GET": {
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			method:  "GET",
			def: chioas.Method{
				Security: chioas.SecuritySchemes{
					{
						Name:        "foo",
						Description: "foo desc not used",
					},
					{
						Name: "bar",
					},
				},
			},
			expect: `"GET": {
	Security: chioas.SecuritySchemes{
		{
			Name: "foo",
		},
		{
			Name: "bar",
		},
	},
},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			method:  "GET",
			def: chioas.Method{
				QueryParams: chioas.QueryParams{
					{
						Ref: "foo",
					},
					{
						Ref: refs.ComponentsPrefix + tags.Parameters + "/bar",
					},
					{
						Name: "baz",
						Example: map[string]any{
							"foo": "bar",
						},
					},
				},
			},
			expect: `"GET": {
	QueryParams: chioas.QueryParams{
		{
			Ref: "foo",
		},
		{
			Ref: "bar",
		},
		{
			Name: "baz",
			Example: map[string]any{
				"foo": "bar",
			},
		},
	},
},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			indent:  1,
			method:  "GET",
			def: chioas.Method{
				Request: &chioas.Request{
					ContentType: "application/json",
				},
			},
			expect: `	"GET": {
		Request: &chioas.Request{
			ContentType: "application/json",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true, UseHttpConsts: true},
			indent:  1,
			method:  "GET",
			def: chioas.Method{
				Responses: chioas.Responses{
					200: {
						Description: "okey dokey",
					},
					400: {
						Description: "bad request",
					},
				},
			},
			expect: `	http.MethodGet: {
		Responses: chioas.Responses{
			http.StatusOK: {
				Description: "okey dokey",
			},
			http.StatusBadRequest: {
				Description: "bad request",
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true, UseHttpConsts: true},
			indent:  1,
			method:  "GET",
			def: chioas.Method{
				Handler: "foo",
			},
			expect: `	http.MethodGet: {
		Handler: "foo",
	},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateMethod(tc.indent, tc.method, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_generateRequest(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		def     *chioas.Request
		expect  string
	}{
		{
			def: &chioas.Request{},
			expect: `	Description: "",
	Required: false,
	ContentType: "",
	IsArray: false,
	Comment: "",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def:     &chioas.Request{},
			expect:  ``,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Ref:         refs.ComponentsPrefix + tags.RequestBodies + "/bar",
				Description: "not shown",
			},
			expect: `	Ref: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
			},
			expect: `	SchemaRef: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Examples: chioas.Examples{
					{
						Name: "foo",
						Value: map[string]any{
							"foo": "bar",
						},
					},
					{
						Name:       "bar",
						ExampleRef: refs.ComponentsPrefix + tags.Examples + "/bar",
					},
				},
			},
			expect: `	Examples: chioas.Examples{
		{
			Name: "foo",
			Value: map[string]any{
				"foo": "bar",
			},
		},
		{
			ExampleRef: "bar",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				ContentType: "application/json",
				AlternativeContentTypes: chioas.ContentTypes{
					"text/csv": {
						IsArray:   true,
						SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
					},
				},
			},
			expect: `	ContentType: "application/json",
	AlternativeContentTypes: chioas.ContentTypes{
		"text/csv": {
			SchemaRef: "bar",
			IsArray: true,
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Description: "request description",
				Schema: chioas.Schema{
					SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
			},
			expect: `	Description: "request description",
	Schema: chioas.Schema{
		SchemaRef: "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Description: "request description",
				Schema: &chioas.Schema{
					SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
			},
			expect: `	Description: "request description",
	Schema: &chioas.Schema{
		SchemaRef: "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				ContentType: "application/json",
				Description: "request description",
				Schema: &chioas.Schema{
					SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
				AlternativeContentTypes: chioas.ContentTypes{
					"text/csv": {
						Schema: chioas.Schema{
							Type: "string",
						},
					},
				},
			},
			expect: `	Description: "request description",
	ContentType: "application/json",
	AlternativeContentTypes: chioas.ContentTypes{
		"text/csv": {
			Schema: chioas.Schema{
				Type: "string",
			},
		},
	},
	Schema: &chioas.Schema{
		SchemaRef: "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				ContentType: "application/json",
				Description: "request description",
				Schema: &chioas.Schema{
					SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
				AlternativeContentTypes: chioas.ContentTypes{
					"text/csv": {
						Schema: chioas.Schema{
							Type: "string",
						},
					},
				},
			},
			expect: `	Description: "request description",
	ContentType: "application/json",
	AlternativeContentTypes: chioas.ContentTypes{
		"text/csv": {
			Schema: chioas.Schema{
				Type: "string",
			},
		},
	},
	Schema: &chioas.Schema{
		SchemaRef: "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Description: "request description",
				Schema: &testSchemaConverter{
					schema: &chioas.Schema{
						Name: "bar",
					},
				},
			},
			expect: `	Description: "request description",
	Schema: &chioas.Schema{
		Name: "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Description: "request description",
				Schema: struct {
					Fld1 string `json:"fld1" oas:"description:fld \"desc\""`
				}{},
			},
			expect: `	Description: "request description",
	Schema: &chioas.Schema{
		Type: "object",
		Properties: chioas.Properties{
			{
				Name: "fld1",
				Description: "fld \"desc\"",
				Type: "string",
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Request{
				Description: "request description",
				Schema: struct {
					Fld1 string `oas:"bad token causes error!"`
				}{},
			},
			expect: `	Description: "request description",
	Schema: struct { Fld1 string "oas:\"bad token causes error!\"" },
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateRequest(tc.indent, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_generateResponse(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		def     chioas.Response
		expect  string
	}{
		{
			def: chioas.Response{},
			expect: `	Description: "",
	NoContent: false,
	ContentType: "",
	IsArray: false,
	Comment: "",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def:     chioas.Response{},
			expect:  ``,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Response{
				Ref:         refs.ComponentsPrefix + tags.Responses + "/bar",
				Description: "not shown",
			},
			expect: `	Ref: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Response{
				SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
			},
			expect: `	SchemaRef: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Response{
				Examples: chioas.Examples{
					{
						Name: "foo",
						Value: map[string]any{
							"foo": "bar",
						},
					},
					{
						Name:       "bar",
						ExampleRef: refs.ComponentsPrefix + tags.Examples + "/bar",
					},
				},
			},
			expect: `	Examples: chioas.Examples{
		{
			Name: "foo",
			Value: map[string]any{
				"foo": "bar",
			},
		},
		{
			ExampleRef: "bar",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Response{
				ContentType: "application/json",
				AlternativeContentTypes: chioas.ContentTypes{
					"text/csv": {
						IsArray:   true,
						SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
					},
				},
			},
			expect: `	ContentType: "application/json",
	AlternativeContentTypes: chioas.ContentTypes{
		"text/csv": {
			SchemaRef: "bar",
			IsArray: true,
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Response{
				Description: "response description",
				Schema: chioas.Schema{
					SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
				},
			},
			expect: `	Description: "response description",
	Schema: chioas.Schema{
		SchemaRef: "bar",
	},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateResponse(tc.indent, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_generateQueryParam(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		def     chioas.QueryParam
		expect  string
	}{
		{
			def: chioas.QueryParam{},
			expect: `	Name: "",
	Description: "",
	Required: false,
	In: "",
	Comment: "",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			indent:  1,
			def:     chioas.QueryParam{},
			expect:  ``,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.QueryParam{
				Description: "not used",
				Ref:         refs.ComponentsPrefix + tags.Parameters + "/bar",
			},
			expect: `	Ref: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.QueryParam{
				Example: []string{"foo", "bar"},
			},
			expect: `	Example: []string{
		"foo",
		"bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.QueryParam{
				Extensions: chioas.Extensions{
					"foo": "bar",
				},
			},
			expect: `	Extensions: chioas.Extensions{
		"foo": "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.QueryParam{
				Schema: &chioas.Schema{
					Description: "schema desc",
				},
			},
			expect: `	Schema: &chioas.Schema{
		Description: "schema desc",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.QueryParam{
				SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
			},
			expect: `	SchemaRef: "bar",
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateQueryParam(tc.indent, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_generateSchema(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		def     *chioas.Schema
		expect  string
	}{
		{
			def: &chioas.Schema{},
			expect: `	Name: "",
	Description: "",
	Type: "",
	Format: "",
	Comment: "",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def:     &chioas.Schema{},
			expect:  ``,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Description: "not used",
				SchemaRef:   refs.ComponentsPrefix + tags.Schemas + "/bar",
			},
			expect: `	SchemaRef: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				RequiredProperties: []string{"foo", "bar"},
			},
			expect: `	RequiredProperties: []string{
		"foo",
		"bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Properties: chioas.Properties{
					{
						Name: "foo",
						Type: "string",
					},
				},
			},
			expect: `	Properties: chioas.Properties{
		{
			Name: "foo",
			Type: "string",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Type:    "boolean",
				Default: true,
				Enum:    []any{true, false},
				Example: false,
			},
			expect: `	Type: "boolean",
	Default: true,
	Example: false,
	Enum: []any{
		true,
		false,
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Discriminator: &chioas.Discriminator{
					PropertyName: "foo",
					Mapping: map[string]string{
						"v1": refs.ComponentsPrefix + tags.Schemas + "/bar",
					},
				},
			},
			expect: `	Discriminator: &chioas.Discriminator{
		PropertyName: "foo",
		Mapping: map[string]string{
			"v1": "bar",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Ofs: &chioas.Ofs{
					OfType: chioas.AnyOf,
					Of: []chioas.OfSchema{
						&chioas.Of{
							SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
						},
						&chioas.Of{
							SchemaDef: &chioas.Schema{
								Type:               "object",
								RequiredProperties: []string{"foo", "bar"},
								Properties: chioas.Properties{
									{
										Name: "foo",
										Type: "string",
									},
									{
										Name: "bar",
										Type: "boolean",
									},
								},
							},
						},
					},
				},
			},
			expect: `	Ofs: &chioas.Ofs{
		OfType: chioas.AnyOf,
		Of: []chioas.OfSchema{
			&chioas.Of{
				SchemaRef: "bar",
			},
			&chioas.Of{
				SchemaDef: &chioas.Schema{
					Type: "object",
					RequiredProperties: []string{
						"foo",
						"bar",
					},
					Properties: chioas.Properties{
						{
							Name: "foo",
							Type: "string",
						},
						{
							Name: "bar",
							Type: "boolean",
						},
					},
				},
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Ofs: &chioas.Ofs{
					OfType: chioas.AllOf,
					Of: []chioas.OfSchema{
						&chioas.Of{
							SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
						},
						&chioas.Of{
							SchemaDef: nil,
						},
					},
				},
			},
			expect: `	Ofs: &chioas.Ofs{
		OfType: chioas.AllOf,
		Of: []chioas.OfSchema{
			&chioas.Of{
				SchemaRef: "bar",
			},
			&chioas.Of{
				SchemaDef: &chioas.Schema{
				},
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Ofs: &chioas.Ofs{
					OfType: chioas.OneOf,
					Of: []chioas.OfSchema{
						&chioas.Of{
							SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
						},
					},
				},
			},
			expect: `	Ofs: &chioas.Ofs{
		OfType: chioas.OneOf,
		Of: []chioas.OfSchema{
			&chioas.Of{
				SchemaRef: "bar",
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: &chioas.Schema{
				Ofs: &chioas.Ofs{
					OfType: 99,
					Of: []chioas.OfSchema{
						&chioas.Of{
							SchemaRef: refs.ComponentsPrefix + tags.Schemas + "/bar",
						},
					},
				},
			},
			expect: `	Ofs: &chioas.Ofs{
		OfType: 99,
		Of: []chioas.OfSchema{
			&chioas.Of{
				SchemaRef: "bar",
			},
		},
	},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateSchema(tc.indent, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_generateProperty(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		def     chioas.Property
		expect  string
	}{
		{
			def: chioas.Property{},
			expect: `	Name: "",
	Description: "",
	Type: "",
	ItemType: "",
	Required: false,
	Format: "",
	Deprecated: false,
	Comment: "",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			indent:  1,
			def:     chioas.Property{},
			expect:  ``,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Example: []string{"foo", "bar"},
			},
			expect: `	Example: []string{
		"foo",
		"bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Description: "not used",
				SchemaRef:   refs.ComponentsPrefix + tags.Schemas + "/bar",
			},
			expect: `	SchemaRef: "bar",
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Extensions: chioas.Extensions{
					"foo": "bar",
				},
			},
			expect: `	Extensions: chioas.Extensions{
		"foo": "bar",
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Constraints: chioas.Constraints{
					Pattern: "[a-z0-9]{3}",
					Maximum: "0",
				},
			},
			expect: `	Constraints: chioas.Constraints{
		Pattern: "[a-z0-9]{3}",
		Maximum: 0,
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Constraints: chioas.Constraints{
					Additional: map[string]any{
						"foo": []string{"bar", "baz"},
					},
				},
			},
			expect: `	Constraints: chioas.Constraints{
		Additional: map[string]any{
			"foo": []string{
				"bar",
				"baz",
			},
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Properties: chioas.Properties{
					{
						Name:     "foo",
						Type:     "string",
						Required: true,
					},
					{
						Name: "bar",
						Type: "number",
					},
				},
			},
			expect: `	Properties: chioas.Properties{
		{
			Name: "foo",
			Type: "string",
			Required: true,
		},
		{
			Name: "bar",
			Type: "number",
		},
	},
`,
		},
		{
			options: Options{OmitZeroValues: true},
			def: chioas.Property{
				Type: "boolean",
				Enum: []any{true, false},
			},
			expect: `	Type: "boolean",
	Enum: []any{
		true,
		false,
	},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			generateProperty(tc.indent, tc.def, w)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_compareMethods(t *testing.T) {
	require.True(t, compareMethods("a", "b"))
	require.False(t, compareMethods("b", "a"))
	require.True(t, compareMethods(http.MethodGet, "a"))
	require.False(t, compareMethods("a", http.MethodGet))
	require.True(t, compareMethods(http.MethodGet, http.MethodPost))
	require.False(t, compareMethods(http.MethodPost, http.MethodGet))
}

type testSchemaConverter struct {
	schema *chioas.Schema
}

func (t *testSchemaConverter) ToSchema() *chioas.Schema {
	return t.schema
}

func goFmtTest(t *testing.T, src []byte) {
	_, err := format.Source(src)
	assert.NoError(t, err)
}
