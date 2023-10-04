package chioas

import (
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteSchemaRef(t *testing.T) {
	testCases := []struct {
		ref     string
		isArray bool
		expect  string
	}{
		{
			ref: "foo",
			expect: `$ref: "#/components/schemas/foo"
`,
		},
		{
			ref:     "foo",
			isArray: true,
			expect: `type: "array"
items:
  $ref: "#/components/schemas/foo"
`,
		},
		{
			ref: "some/uri",
			expect: `$ref: "some/uri"
`,
		},
		{
			ref:     "some/uri",
			isArray: true,
			expect: `type: "array"
items:
  $ref: "some/uri"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			writeSchemaRef(tc.ref, tc.isArray, w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}

func TestNilString(t *testing.T) {
	v := nilString("")
	assert.Nil(t, v)
	v = nilString("foo")
	assert.NotNil(t, v)
}

func TestNilBool(t *testing.T) {
	v := nilBool(false)
	assert.Nil(t, v)
	v = nilBool(true)
	assert.NotNil(t, v)
}

func TestNilNumber(t *testing.T) {
	v := nilNumber("")
	assert.Nil(t, v)
	v = nilNumber("NaN")
	assert.Nil(t, v)
	v = nilNumber("Inf")
	assert.Nil(t, v)
	v = nilNumber("1")
	assert.NotNil(t, v)
	assert.Equal(t, int64(1), v)
	v = nilNumber("1.1")
	assert.NotNil(t, v)
	assert.Equal(t, json.Number("1.1"), v)
}

func TestNilUint(t *testing.T) {
	v := nilUint(0)
	assert.Nil(t, v)
	v = nilUint(1)
	assert.NotNil(t, v)
	assert.Equal(t, uint(1), v)
}
