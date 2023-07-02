package chioas

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewYamlWriter(t *testing.T) {
	w := newYamlWriter(nil)
	assert.NotNil(t, w)
	assert.NotNil(t, w.w)

	var buf bytes.Buffer
	dw := bufio.NewWriter(&buf)
	w = newYamlWriter(dw)
	assert.NotNil(t, w)
	assert.Equal(t, dw, w.w)
}

func TestYamlWriter_IncIndent(t *testing.T) {
	w := newYamlWriter(nil)
	assert.Equal(t, 0, len(w.indent))
	w.incIndent()
	assert.Equal(t, 2, len(w.indent))
}

func TestYamlWriter_DecIndent(t *testing.T) {
	w := newYamlWriter(nil)
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

func TestYamlWriter_WriteIndent(t *testing.T) {
	w := newYamlWriter(nil)
	assert.Equal(t, 0, len(w.indent))
	assert.True(t, w.writeIndent())
	data, err := w.bytes()
	require.NoError(t, err)
	assert.Equal(t, 0, len(data))
	w.incIndent()
	assert.True(t, w.writeIndent())
	data, err = w.bytes()
	require.NoError(t, err)
	assert.Equal(t, 2, len(data))

	w.err = errors.New("foo")
	assert.False(t, w.writeIndent())
}

func TestYamlValue(t *testing.T) {
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
			expect: `1.1`,
		},
		{
			value:  true,
			expect: `true`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			v := yamlValue(tc.value, tc.allowEmpty)
			assert.Equal(t, tc.expect, v)
		})
	}
}

func TestYamlWriter_WriteTagValue(t *testing.T) {
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
			w := newYamlWriter(nil)
			w.writeTagValue(tc.tag, tc.value)
			data, err := w.bytes()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}

func TestYamlWriter_WriteTagStart(t *testing.T) {
	w := newYamlWriter(nil)
	w.writeTagStart("foo")
	data, err := w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "foo:\n", string(data))
}

func TestYamlWriter_WriteTagEnd(t *testing.T) {
	w := newYamlWriter(nil)
	w.incIndent()
	assert.Equal(t, 2, len(w.indent))
	w.writeTagEnd()
	assert.Equal(t, 0, len(w.indent))
	assert.NoError(t, w.err)
	w.writeTagEnd()
	assert.Error(t, w.err)
}

func TestYamlWriter_WritePathStart(t *testing.T) {
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
			w := newYamlWriter(nil)
			w.writePathStart(tc.context, tc.path)
			data, err := w.bytes()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
			assert.Equal(t, 2, len(w.indent))
		})
	}
}

func TestYamlWriter_WriteItem(t *testing.T) {
	w := newYamlWriter(nil)
	w.writeItem("foo")
	data, err := w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "- \"foo\"\n", string(data))

	w = newYamlWriter(nil)
	w.writeItem(1)
	data, err = w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "- 1\n", string(data))

	w = newYamlWriter(nil)
	w.writeItem(nil)
	data, err = w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "", string(data))

	w = newYamlWriter(nil)
	w.writeItem("")
	data, err = w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "", string(data))
}

func TestYamlWriter_WriteItemStart(t *testing.T) {
	w := newYamlWriter(nil)
	w.writeItemStart("foo", "bar")
	data, err := w.bytes()
	require.NoError(t, err)
	assert.Equal(t, "- foo: \"bar\"\n", string(data))
	assert.Equal(t, 2, len(w.indent))
}
