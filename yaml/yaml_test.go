package yaml

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrivateNewYamlWriter(t *testing.T) {
	w := newWriter(nil)
	assert.NotNil(t, w)
	assert.NotNil(t, w.w)

	var buf bytes.Buffer
	dw := bufio.NewWriter(&buf)
	w = newWriter(dw)
	assert.NotNil(t, w)
	assert.Equal(t, dw, w.w)
}

func TestNewYamlWriter(t *testing.T) {
	w := NewWriter(nil)
	assert.NotNil(t, w)
}

func TestWriter_Flush(t *testing.T) {
	w := NewWriter(nil)
	err := w.Flush()
	assert.NoError(t, err)

	w = NewWriter(nil)
	w.SetError(errors.New("foo"))
	err = w.Flush()
	assert.Error(t, err)
}

func TestWriter_Errored(t *testing.T) {
	w := NewWriter(nil)
	err := w.Errored()
	assert.NoError(t, err)

	w.SetError(errors.New("foo"))
	err = w.Errored()
	assert.Error(t, err)

	w.SetError(nil)
	err = w.Errored()
	assert.Error(t, err)
}

func TestWriter_IncIndent(t *testing.T) {
	w := newWriter(nil)
	assert.Equal(t, 0, len(w.indent))
	w.incIndent()
	assert.Equal(t, 2, len(w.indent))
}

func TestWriter_DecIndent(t *testing.T) {
	w := newWriter(nil)
	assert.Equal(t, 0, len(w.indent))
	w.incIndent()
	assert.Equal(t, 2, len(w.indent))
	w.decIndent()
	assert.Equal(t, 0, len(w.indent))
	assert.NoError(t, w.err)
	w.decIndent()
	assert.Equal(t, 0, len(w.indent))
	assert.Error(t, w.err)
}

func TestWriter_WriteIndent(t *testing.T) {
	w := newWriter(nil)
	assert.Equal(t, 0, len(w.indent))
	assert.True(t, w.writeIndent())
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, 0, len(data))
	w.incIndent()
	assert.True(t, w.writeIndent())
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, 2, len(data))

	w.err = errors.New("foo")
	assert.False(t, w.writeIndent())
}

func TestWriter_WriteLines(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	w.WriteLines("bar: 1", "", "baz: true", "")
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo:
  bar: 1
  
  baz: true
  
`
	assert.Equal(t, expect, string(data))
}

func TestWriter_WriteLines_Crlf(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	w.WriteLines("bar: 1\n\nbaz: true\n")
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo:
  bar: 1
  
  baz: true
  
`
	assert.Equal(t, expect, string(data))
}

func TestWriter_Write(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	w.Write([]byte("this\nthen this\nand this\n"))
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo:
this
then this
and this
`
	assert.Equal(t, expect, string(data))
}

func TestWriter_WriteComments(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	w.WriteComments("first", "", "second", "")
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo:
  #first
  #
  #second
  #
`
	assert.Equal(t, expect, string(data))
}

func TestWriter_WriteComments_Empty(t *testing.T) {
	w := newWriter(nil)
	w.WriteComments("")
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "", string(data))

	w = newWriter(nil)
	w.WriteComments("", "")
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, "#\n#\n", string(data))
}

func TestWriter_WriteComments_Crlf(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	w.WriteComments("first\n\nsecond\n")
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo:
  #first
  #
  #second
  #
`
	assert.Equal(t, expect, string(data))
}

func TestWriter_CurrentIndent(t *testing.T) {
	w := newWriter(nil)
	assert.Equal(t, "", w.CurrentIndent())
	w.WriteTagStart("foo")
	assert.Equal(t, "  ", w.CurrentIndent())
	w.WriteTagStart("bar")
	assert.Equal(t, "    ", w.CurrentIndent())
	w.WriteTagEnd()
	assert.Equal(t, "  ", w.CurrentIndent())
	w.WriteTagEnd()
	assert.Equal(t, "", w.CurrentIndent())
}

func TestYamlValue(t *testing.T) {
	pstr := "foo"
	pbool := true
	pint := int(16)
	pint8 := int8(16)
	pint16 := int16(16)
	pint32 := int32(16)
	pint64 := int64(16)
	puint := uint(16)
	puint8 := uint8(16)
	puint16 := uint16(16)
	puint32 := uint32(16)
	puint64 := uint64(16)
	pf32 := float32(16.16)
	pf64 := float64(16.16)
	var nilstr *string
	var nilbool *bool
	var nilint *int
	testCases := []struct {
		value      any
		allowEmpty bool
		expect     []string
	}{
		{
			value:  "foo",
			expect: []string{`foo`},
		},
		{
			value:  "foo bar",
			expect: []string{`"foo bar"`},
		},
		{
			value:  "",
			expect: []string{},
		},
		{
			value:      "",
			allowEmpty: true,
			expect:     []string{`""`},
		},
		{
			value:  1,
			expect: []string{`1`},
		},
		{
			value:  1.1,
			expect: []string{`1.1`},
		},
		{
			value:  true,
			expect: []string{`true`},
		},
		{
			value:  &pstr,
			expect: []string{`foo`},
		},
		{
			value:  nilstr,
			expect: []string{},
		},
		{
			value:      nilstr,
			allowEmpty: true,
			expect:     []string{`""`},
		},
		{
			value:  &pbool,
			expect: []string{`true`},
		},
		{
			value:  nilbool,
			expect: []string{},
		},
		{
			value:      nilbool,
			allowEmpty: true,
			expect:     []string{`false`},
		},
		{
			value:  uint8(1),
			expect: []string{`1`},
		},
		{
			value:  nilint,
			expect: []string{},
		},
		{
			value:  &pint,
			expect: []string{`16`},
		},
		{
			value:  &pint8,
			expect: []string{`16`},
		},
		{
			value:  &pint16,
			expect: []string{`16`},
		},
		{
			value:  &pint32,
			expect: []string{`16`},
		},
		{
			value:  &pint64,
			expect: []string{`16`},
		},
		{
			value:  &puint,
			expect: []string{`16`},
		},
		{
			value:  &puint8,
			expect: []string{`16`},
		},
		{
			value:  &puint16,
			expect: []string{`16`},
		},
		{
			value:  &puint32,
			expect: []string{`16`},
		},
		{
			value:  &puint64,
			expect: []string{`16`},
		},
		{
			value:  &pf32,
			expect: []string{`16.16`},
		},
		{
			value:  &pf64,
			expect: []string{`16.16`},
		},
		{
			value: `aaa
bbb
ccc`,
			expect: []string{`|-`, `aaa`, `bbb`, `ccc`},
		},
		{
			value: `aaa

bbb

ccc

`,
			expect: []string{`|+`, `aaa`, ``, `bbb`, ``, `ccc`, ``},
		},
		{
			value:  "aaa\rbbb\rccc",
			expect: []string{"\"aaa\\rbbb\\rccc\""},
		},
		{
			value:  "\r\r\r",
			expect: []string{"\"\\r\\r\\r\""},
		},
		{
			value:  "\n\t\n\t\n",
			expect: []string{"|2", "\t", "\t"},
		},
		{
			value: LiteralValue{
				Value: "[]",
			},
			expect: []string{`[]`},
		},
		{
			value: &LiteralValue{
				Value: "[]",
			},
			expect: []string{`[]`},
		},
		{
			value: LiteralValue{
				Value:            "",
				SafeEncodeString: true,
			},
			expect: []string{"\"\""},
		},
		{
			value: LiteralValue{
				Value:            string(rune(119070)),
				SafeEncodeString: true,
			},
			expect: []string{"\"\\U0001D11E\""},
		},
		{
			value: LiteralValue{
				Value:            "a\nb\nc",
				SafeEncodeString: true,
			},
			expect: []string{"|-", "a", "b", "c"},
		},
		{
			value: LiteralValue{
				Value:            "\n\n\n",
				SafeEncodeString: true,
			},
			expect: []string{"|2+", "", ""},
		},
		{
			value: &LiteralValue{
				Value:            "",
				SafeEncodeString: true,
			},
			expect: []string{"\"\""},
		},
		{
			value: map[string]any{
				"foo": map[string]any{
					"bar": map[string]any{
						"baz": map[string]any{
							"buzz": 1,
						},
					},
				},
			},
			expect: []string{"", `foo:`, `  bar:`, `    baz:`, `      buzz: 1`},
		},
		{
			value: &exampleStruct{
				Foo: "bar\nbaz",
			},
			expect: []string{"", "Foo: |-", "  bar", "  baz"},
		},
		{
			value:  []any{"a", 1, true},
			expect: []string{"", "- a", "- 1", "- true"},
		},
		{
			value:  json.Number(""),
			expect: []string{""},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := newWriter(nil)
			v := w.yamlValue(tc.value, tc.allowEmpty)
			assert.Equal(t, tc.expect, v)
		})
	}
}

type exampleStruct struct {
	Foo string `yaml:"Foo"`
}

func TestWriter_WriteTagValue(t *testing.T) {
	testCases := []struct {
		tag    string
		value  any
		expect string
	}{
		{
			tag:    "foo",
			value:  "bar",
			expect: "foo: bar\n",
		},
		{
			tag:    "foo",
			value:  "bar baz",
			expect: "foo: \"bar baz\"\n",
		},
		{
			tag:    "foo",
			value:  1,
			expect: "foo: 1\n",
		},
		{
			tag:    "foo",
			value:  true,
			expect: "foo: true\n",
		},
		{
			tag:    "foo",
			value:  nil,
			expect: "",
		},
		{
			tag:   "foo",
			value: "aaa\nbbb\nccc",
			expect: `foo: |-
  aaa
  bbb
  ccc
`,
		},
		{
			tag:    "foo",
			value:  "",
			expect: ``,
		},
		{
			tag:   "foo bar",
			value: "baz",
			expect: `"foo bar": baz
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := newWriter(nil)
			w.WriteTagValue(tc.tag, tc.value)
			data, err := w.Bytes()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}

func TestWriter_WriteTagStart(t *testing.T) {
	w := newWriter(nil)
	w.WriteTagStart("foo")
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "foo:\n", string(data))

	w = newWriter(nil)
	w.WriteTagStart("foo bar")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "\"foo bar\":\n", string(data))
}

func TestWriter_WriteTagEnd(t *testing.T) {
	w := newWriter(nil)
	w.incIndent()
	assert.Equal(t, 2, len(w.indent))
	w.WriteTagEnd()
	assert.Equal(t, 0, len(w.indent))
	assert.NoError(t, w.err)
	w.WriteTagEnd()
	assert.Error(t, w.err)
}

func TestWriter_WritePathStart(t *testing.T) {
	testCases := []struct {
		context string
		path    string
		expect  string
	}{
		{
			path:   "/",
			expect: "\"/\":\n",
		},
		{
			path:   "/foo",
			expect: "\"/foo\":\n",
		},
		{
			context: "bar",
			path:    "/foo",
			expect:  "\"/bar/foo\":\n",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := newWriter(nil)
			w.WritePathStart(tc.context, tc.path)
			data, err := w.Bytes()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
			assert.Equal(t, 2, len(w.indent))
		})
	}
}

func TestWriter_WriteItem(t *testing.T) {
	w := newWriter(nil)
	w.WriteItem("foo")
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo\n", string(data))

	w = newWriter(nil)
	w.WriteItem(1)
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- 1\n", string(data))

	w = newWriter(nil)
	w.WriteItem(nil)
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "", string(data))

	w = newWriter(nil)
	w.WriteItem("")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "", string(data))

	w = newWriter(nil)
	w.WriteItem("aaa\nbbb\nccc")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, `- |-
  aaa
  bbb
  ccc
`, string(data))
}

func TestWriter_WriteItemValue(t *testing.T) {
	w := newWriter(nil)
	w.WriteItemValue("foo", "bar")
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo: bar\n", string(data))
	assert.Equal(t, 0, len(w.indent))

	w = newWriter(nil)
	w.WriteItemValue("foo bar", "baz")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- \"foo bar\": baz\n", string(data))
	assert.Equal(t, 0, len(w.indent))

	w = newWriter(nil)
	w.WriteItemValue("foo", nil)
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo:\n", string(data))
	assert.Equal(t, 0, len(w.indent))

	w = newWriter(nil)
	w.WriteItemValue("foo", "aaa\nbbb\nccc")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, `- foo: |-
  aaa
  bbb
  ccc
`, string(data))
	assert.Equal(t, 0, len(w.indent))
}

func TestWriter_WriteItemStart(t *testing.T) {
	w := newWriter(nil)
	w.WriteItemStart("foo", "bar")
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo: bar\n", string(data))
	assert.Equal(t, 2, len(w.indent))

	w = newWriter(nil)
	w.WriteItemStart("foo bar", "baz")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- \"foo bar\": baz\n", string(data))
	assert.Equal(t, 2, len(w.indent))

	w = newWriter(nil)
	w.WriteItemStart("foo", nil)
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo:\n", string(data))
	assert.Equal(t, 2, len(w.indent))

	w = newWriter(nil)
	w.WriteItemStart("foo", "aaa\nbbb\nccc")
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, `- foo: |-
  aaa
  bbb
  ccc
`, string(data))
	assert.Equal(t, 2, len(w.indent))

	w = newWriter(nil)
	w.WriteItemStart("foo", map[string]any{
		"bar": map[string]any{
			"baz:": true,
		},
	})
	data, err = w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, `- foo:
  bar:
    'baz:': true
`, string(data))
	assert.Equal(t, 2, len(w.indent))
}

func TestWriter_Bytes_ReturnsErr(t *testing.T) {
	w := newWriter(nil)
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Empty(t, data)
	w.err = errors.New("foo")
	_, err = w.Bytes()
	require.Error(t, err)
}

func TestWriter_RefChecker(t *testing.T) {
	w := newWriter(nil)
	assert.Nil(t, w.refChecker)
	rc := w.RefChecker(nil)
	assert.NotNil(t, rc)
	assert.Equal(t, nullRefChecker, rc)
	err := rc.RefCheck("any", "any")
	assert.NoError(t, err)

	w.RefChecker(&testRefChecker{})
	rc = w.RefChecker(nil)
	assert.NotEqual(t, nullRefChecker, rc)
	assert.Error(t, w.RefChecker(nil).RefCheck("any", "any"))
}

type testRefChecker struct {
}

func (r *testRefChecker) RefCheck(area, ref string) error {
	return errors.New("fooey")
}

func TestFormattedString(t *testing.T) {
	testCases := []struct {
		s         string
		expect    []string
		expectErr bool
	}{
		{
			s:      "",
			expect: []string{`""`},
		},
		{
			s:      "a\nb\nc",
			expect: []string{"|-", "a", "b", "c"},
		},
		{
			s:      "\n",
			expect: []string{"|2+"},
		},
		{
			s:      "\n\n\n",
			expect: []string{"|2+", "", ""},
		},
		{
			s:      "aaa\n\n\nbbb",
			expect: []string{"|-", "aaa", "", "", "bbb"},
		},
		{
			s:      "aaa\n\nand \"this\" is quoted\n\bbb",
			expect: []string{"\"aaa\\n\\nand \\\"this\\\" is quoted\\n\\bbb\""},
		},
		{
			s:      string(rune(119070)) + "\n",
			expect: []string{"\"\\U0001D11E\\n\""},
		},
		{
			s:      string(rune(119070)) + "\n\n\n",
			expect: []string{"\"\\U0001D11E\\n\\n\\n\""},
		},
		{
			s:      "a\n" + string(rune(119070)) + "\n" + string(rune(119070)),
			expect: []string{"\"a\\n\\U0001D11E\\n\\U0001D11E\""},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := newWriter(nil)
			actual := w.formattedString(tc.s)
			if !tc.expectErr {
				assert.NoError(t, w.err)
				assert.Equal(t, tc.expect, actual)
			} else {
				assert.Error(t, w.err)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	testCases := map[string]string{
		``:                       `""`,
		`"`:                      `"\""`,
		`:`:                      `":"`,
		"\uABCD":                 `"\uABCD"`,
		string(rune(119070)):     `"\U0001D11E"`,
		string([]byte{0x1b}):     `"\e"`,
		string([]byte{0xf}):      `"\u000F"`,
		`foo`:                    `foo`,
		`foo bar`:                `"foo bar"`,
		`foo:bar`:                `"foo:bar"`,
		`123`:                    `123`,
		`application/json`:       `"application/json"`,
		"\u0000\a\b\f\n\r\t\v\"": `"\0\a\b\f\n\r\t\v\""`,
		`"foo"`:                  `"\"foo\""`,
	}
	for s, expect := range testCases {
		t.Run(s, func(t *testing.T) {
			s := safeString(s)
			assert.Equal(t, expect, s)
		})
	}
}

func TestSafeStringName(t *testing.T) {
	testCases := map[string]string{
		`"`:                      `"\"":`,
		`:`:                      `":":`,
		"\uABCD":                 `"\uABCD":`,
		string(rune(119070)):     `"\U0001D11E":`,
		string([]byte{0x1b}):     `"\e":`,
		string([]byte{0xf}):      `"\u000F":`,
		`foo`:                    `foo:`,
		`foo bar`:                `"foo bar":`,
		`foo:bar`:                `"foo:bar":`,
		`123`:                    `123:`,
		`application/json`:       `"application/json":`,
		"\u0000\a\b\f\n\r\t\v\"": `"\0\a\b\f\n\r\t\v\"":`,
		`"foo"`:                  `"foo":`,
	}
	for name, expect := range testCases {
		t.Run(name, func(t *testing.T) {
			s := safeStringName(name)
			assert.Equal(t, expect, s)
		})
	}
}
