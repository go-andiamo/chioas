package yaml

import (
	"bufio"
	"bytes"
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
	testCases := []struct {
		value      any
		allowEmpty bool
		expect     string
	}{
		{
			value:  "foo",
			expect: `"foo"`,
		},
		{
			value:  "",
			expect: ``,
		},
		{
			value:      "",
			allowEmpty: true,
			expect:     `""`,
		},
		{
			value:  1,
			expect: `1`,
		},
		{
			value:  1.1,
			expect: `1.100000`,
		},
		{
			value:  true,
			expect: `true`,
		},
		{
			value:  &pstr,
			expect: `"foo"`,
		},
		{
			value:  nilstr,
			expect: ``,
		},
		{
			value:      nilstr,
			allowEmpty: true,
			expect:     `""`,
		},
		{
			value:  &pbool,
			expect: `true`,
		},
		{
			value:  nilbool,
			expect: ``,
		},
		{
			value:      nilbool,
			allowEmpty: true,
			expect:     `false`,
		},
		{
			value:  uint8(1),
			expect: `1`,
		},
		{
			value:  &pint,
			expect: `16`,
		},
		{
			value:  &pint8,
			expect: `16`,
		},
		{
			value:  &pint16,
			expect: `16`,
		},
		{
			value:  &pint32,
			expect: `16`,
		},
		{
			value:  &pint64,
			expect: `16`,
		},
		{
			value:  &puint,
			expect: `16`,
		},
		{
			value:  &puint8,
			expect: `16`,
		},
		{
			value:  &puint16,
			expect: `16`,
		},
		{
			value:  &puint32,
			expect: `16`,
		},
		{
			value:  &puint64,
			expect: `16`,
		},
		{
			value:  &pf32,
			expect: `16.160000`,
		},
		{
			value:  &pf64,
			expect: `16.160000`,
		},
		{
			value: `aaa
bbb
ccc`,
			expect: `>-
  aaa
  bbb
  ccc`,
		},
		{
			value: `aaa

bbb

ccc

`,
			expect: `>-
  aaa
  
  bbb
  
  ccc
  
  `,
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

func TestWriter_WriteTagValue(t *testing.T) {
	testCases := []struct {
		tag    string
		value  any
		expect string
	}{
		{
			tag:    "foo",
			value:  "bar",
			expect: "foo: \"bar\"\n",
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
	assert.Equal(t, "- \"foo\"\n", string(data))

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
}

func TestWriter_WriteItemStart(t *testing.T) {
	w := newWriter(nil)
	w.WriteItemStart("foo", "bar")
	data, err := w.Bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo: \"bar\"\n", string(data))
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
