package codegen

import (
	"bytes"
	"fmt"
	"github.com/go-andiamo/chioas"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCodeWriter_writePrologue(t *testing.T) {
	t.Run("no alias", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{})
		w.writePrologue()
		require.NoError(t, w.err)
		require.Equal(t, `package api

import (
	"github.com/go-andiamo/chioas"
)

`, string(buf.Bytes()))
	})
	t.Run("no alias - specified package", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{Package: "my_pkg"})
		w.writePrologue()
		require.NoError(t, w.err)
		require.Equal(t, `package my_pkg

import (
	"github.com/go-andiamo/chioas"
)

`, buf.String())
	})
	t.Run("aliased", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{ImportAlias: "alias"})
		w.writePrologue()
		require.NoError(t, w.err)
		require.Equal(t, `package api

import (
	alias "github.com/go-andiamo/chioas"
)

`, buf.String())
	})
	t.Run("UseHttpConsts", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{UseHttpConsts: true})
		w.writePrologue()
		require.NoError(t, w.err)
		require.Equal(t, `package api

import (
	"net/http"

	"github.com/go-andiamo/chioas"
)

`, buf.String())
	})
}

func TestCodeWriter_writeVarStart(t *testing.T) {
	t.Run("no alias", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{})
		w.writeVarStart("foo", typeDefinition, false)
		require.NoError(t, w.err)
		require.Equal(t, `var foo = chioas.Definition{
`, buf.String())
	})
	t.Run("ptr no alias", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{})
		w.writeVarStart("foo", typeDefinition, true)
		require.NoError(t, w.err)
		require.Equal(t, `var foo = &chioas.Definition{
`, buf.String())
	})
	t.Run("aliased", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{ImportAlias: "alias"})
		w.writeVarStart("foo", typeDefinition, false)
		require.NoError(t, w.err)
		require.Equal(t, `var foo = alias.Definition{
`, buf.String())
	})
	t.Run("dot alias", func(t *testing.T) {
		var buf bytes.Buffer
		w := newCodeWriter(&buf, Options{ImportAlias: "."})
		w.writeVarStart("foo", typeDefinition, false)
		require.NoError(t, w.err)
		require.Equal(t, `var foo = Definition{
`, buf.String())
	})
}

func TestCodeWriter_writeValue(t *testing.T) {
	testCases := []struct {
		value  any
		expect string
	}{
		{
			value:  nil,
			expect: "nil,\n",
		},
		{
			value:  "str",
			expect: "\"str\",\n",
		},
		{
			value:  true,
			expect: "true,\n",
		},
		{
			value:  1,
			expect: "1,\n",
		},
		{
			value:  int8(1),
			expect: "1,\n",
		},
		{
			value:  int16(1),
			expect: "1,\n",
		},
		{
			value:  int32(1),
			expect: "1,\n",
		},
		{
			value:  int64(1),
			expect: "1,\n",
		},
		{
			value:  uint(1),
			expect: "1,\n",
		},
		{
			value:  uint8(1),
			expect: "1,\n",
		},
		{
			value:  uint16(1),
			expect: "1,\n",
		},
		{
			value:  uint32(1),
			expect: "1,\n",
		},
		{
			value:  uint64(1),
			expect: "1,\n",
		},
		{
			value:  float32(1.1),
			expect: "1.1,\n",
		},
		{
			value:  float64(1.1),
			expect: "1.1,\n",
		},
		{
			value: map[string]any{"a": "b"},
			expect: `map[string]any{
	"a": "b",
},
`,
		},
		{
			value: []any{"a", 1, true},
			expect: `[]any{
	"a",
	1,
	true,
},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, Options{})
			w.writeValue(0, tc.value)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func TestCodeWriter_writeExtensions(t *testing.T) {
	testCases := []struct {
		extensions chioas.Extensions
		options    Options
		indent     int
		expect     string
	}{
		{},
		{
			extensions: chioas.Extensions{
				"foo": "bar",
				"bar": 2,
				"baz": true,
			},
			expect: `Extensions: chioas.Extensions{
	"bar": 2,
	"baz": true,
	"foo": "bar",
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": nil,
			},
			expect: `Extensions: chioas.Extensions{
	"foo": nil,
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": "bar",
				"bar": 2,
				"baz": true,
			},
			options: Options{ImportAlias: "."},
			expect: `Extensions: Extensions{
	"bar": 2,
	"baz": true,
	"foo": "bar",
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			indent: 0,
			expect: `Extensions: chioas.Extensions{
	"foo": map[string]any{
		"foo": "bar",
	},
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": map[int]any{
					1: "bar",
				},
			},
			indent: 0,
			expect: `Extensions: chioas.Extensions{
	"foo": map[int]any{
		1: "bar",
	},
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": map[string]any{},
			},
			indent: 1,
			expect: `	Extensions: chioas.Extensions{
		"foo": map[string]any{},
	},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			indent: 1,
			expect: `	Extensions: chioas.Extensions{
		"foo": map[string]any{
			"foo": "bar",
		},
	},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo": "bar",
					},
				},
			},
			indent: 1,
			expect: `	Extensions: chioas.Extensions{
		"foo": map[string]any{
			"foo": map[string]any{
				"foo": "bar",
			},
		},
	},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": []string{},
			},
			indent: 0,
			expect: `Extensions: chioas.Extensions{
	"foo": []string{},
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": []string{"bar", "baz"},
			},
			indent: 1,
			expect: `	Extensions: chioas.Extensions{
		"foo": []string{
			"bar",
			"baz",
		},
	},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": []any{
					map[string]any{
						"foo": "bar",
					},
					"bar",
					2,
					nil,
				},
			},
			indent: 0,
			expect: `Extensions: chioas.Extensions{
	"foo": []any{
		map[string]any{
			"foo": "bar",
		},
		"bar",
		2,
		nil,
	},
},
`,
		},
		{
			extensions: chioas.Extensions{
				"foo": struct{}{},
			},
			indent: 0,
			expect: `Extensions: chioas.Extensions{
	"foo": "Unknown value type: struct {}",
},
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			w.writeExtensions(tc.indent, tc.extensions)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}

func Test_writeZeroField(t *testing.T) {
	testCases := []struct {
		options Options
		indent  int
		name    string
		value   any
		expect  string
	}{
		{
			name:   "test",
			value:  "",
			expect: "test: \"\",\n",
		},
		{
			name:   "test",
			value:  "",
			indent: 1,
			expect: "\ttest: \"\",\n",
		},
		{
			name:   "test",
			value:  0,
			expect: "test: 0,\n",
		},
		{
			name:   "test",
			value:  false,
			expect: "test: false,\n",
		},
		{
			name:   "test",
			value:  nil,
			expect: "test: nil,\n",
		},
		{
			options: Options{OmitZeroValues: true},
			name:    "test",
			value:   "",
			expect:  "",
		},
		{
			options: Options{OmitZeroValues: true},
			name:    "test",
			value:   0,
			expect:  "",
		},
		{
			options: Options{OmitZeroValues: true},
			name:    "test",
			value:   false,
			expect:  "",
		},
		{
			options: Options{OmitZeroValues: true},
			name:    "test",
			value:   nil,
			expect:  "",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			var buf bytes.Buffer
			w := newCodeWriter(&buf, tc.options)
			writeZeroField(w, tc.indent, tc.name, tc.value)
			require.NoError(t, w.err)
			require.Equal(t, tc.expect, buf.String())
		})
	}
}
