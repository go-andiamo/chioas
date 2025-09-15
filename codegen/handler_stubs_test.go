package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func TestGenerateHandlerStubs_Definition(t *testing.T) {
	testCases := []struct {
		def     chioas.Definition
		options HandlerStubOptions
		expect  string
	}{
		{
			def: chioas.Definition{},
			expect: `package api

import (
	"net/http"
)

`,
		},
		{
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet:     {},
					http.MethodOptions: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

func optionsRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			def: chioas.Definition{
				Paths: chioas.Paths{
					"api/": {
						Paths: chioas.Paths{
							"/foos": {
								Methods: chioas.Methods{
									http.MethodGet: {},
								},
								Paths: chioas.Paths{
									"/{id:*}": {
										Methods: chioas.Methods{
											http.MethodGet: {},
										},
									},
								},
							},
						},
					},
				},
			},
			expect: `package api

import (
	"net/http"
)

func getFoos(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

func getFoosId(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			options: HandlerStubOptions{PublicFuncs: true},
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet:     {},
					http.MethodOptions: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

func OptionsRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			options: HandlerStubOptions{PathParams: true, PublicFuncs: true},
			def: chioas.Definition{
				Paths: chioas.Paths{
					"/api": {
						Paths: chioas.Paths{
							"/foos": {
								Methods: chioas.Methods{},
								Paths: chioas.Paths{
									"/{id:*}": {
										Methods: chioas.Methods{
											http.MethodGet: {},
										},
									},
								},
							},
						},
					},
				},
			},
			expect: `package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func GetFoosId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = id
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			options: HandlerStubOptions{PathParams: true, PublicFuncs: true, GoDoc: true},
			def: chioas.Definition{
				Paths: chioas.Paths{
					"/api": {
						Paths: chioas.Paths{
							"/foos": {
								Methods: chioas.Methods{},
								Paths: chioas.Paths{
									"/{id:*}": {
										Methods: chioas.Methods{
											http.MethodGet: {},
										},
									},
								},
							},
						},
					},
				},
			},
			expect: `package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetFoosId GET /api/foos/{id:*}
func GetFoosId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_ = id
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			options: HandlerStubOptions{StubNaming: &emptyStubNames{}},
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`,
		},
		{
			options: HandlerStubOptions{Format: true},
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}
`,
		},
		{
			options: HandlerStubOptions{Format: true, UseCRLF: true},
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}
`,
		},
		{
			options: HandlerStubOptions{Receiver: "(a *MyApi)"},
			def: chioas.Definition{
				Methods: chioas.Methods{
					http.MethodGet: {},
				},
			},
			expect: `package api

import (
	"net/http"
)

func (a *MyApi) getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			err := GenerateHandlerStubs(tc.def, &buf, tc.options)
			require.NoError(t, err)
			if tc.options.UseCRLF {
				require.Equal(t, strings.ReplaceAll(tc.expect, "\n", "\r\n"), buf.String())
			} else {
				require.Equal(t, tc.expect, buf.String())
			}
		})
	}
}

func TestGenerateHandlerStubs_differentTypes(t *testing.T) {
	t.Run("ptr definitions", func(t *testing.T) {
		def := &chioas.Definition{
			Methods: chioas.Methods{
				http.MethodGet: {},
			},
		}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("paths", func(t *testing.T) {
		def := chioas.Paths{
			"/api": {
				Methods: chioas.Methods{
					http.MethodGet: {},
				},
			},
		}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func getApi(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("path", func(t *testing.T) {
		def := chioas.Path{
			Methods: chioas.Methods{
				http.MethodGet: {},
			},
		}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func get(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("ptr path", func(t *testing.T) {
		def := &chioas.Path{
			Methods: chioas.Methods{
				http.MethodGet: {},
			},
		}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func get(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("method", func(t *testing.T) {
		def := chioas.Method{}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
	t.Run("ptr method", func(t *testing.T) {
		def := &chioas.Method{}
		var buf bytes.Buffer
		err := GenerateHandlerStubs(def, &buf, HandlerStubOptions{})
		require.NoError(t, err)
		expect := `package api

import (
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// TODO implement me
	panic("implement me!")
}

`
		require.Equal(t, expect, buf.String())
	})
}

type emptyStubNames struct{}

func (e *emptyStubNames) Name(path string, method string, def chioas.Method) string {
	return ""
}
