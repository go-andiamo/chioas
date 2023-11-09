package typed

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestDefaultUnmarshaler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"foo":"bar"}`))
	m := map[string]any{}
	err := defaultUnmarshaler.Unmarshal(req, &m)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"foo": "bar"}, m)
}

func TestMultiUnmarshaler(t *testing.T) {
	testCases := []struct {
		body         string
		contentType  string
		expectErr    bool
		expectStatus int
	}{
		{
			expectErr:    true,
			expectStatus: http.StatusUnsupportedMediaType,
		},
		{
			contentType:  contentTypeJson,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
		{
			contentType:  contentTypeYaml,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
		{
			contentType:  contentTypeXml,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
		{
			body:        `{"name":"foo","age":16}`,
			contentType: contentTypeJson,
		},
		{
			body:        `{"name":"foo","age":16}`,
			contentType: contentTypeJson + "+ext",
		},
		{
			body:        `{"name":"foo","age":16}`,
			contentType: contentTypeJson + "; charset=utf-8",
		},
		{
			body: `name: foo
age: 16`,
			contentType: contentTypeYaml,
		},
		{
			body: `name: foo
age: 16`,
			contentType: contentTypeYaml + "; charset=utf-8",
		},
		{
			body: `<root>
<name>foo</name>
<age>16</age>
</root>`,
			contentType: contentTypeXml,
		},
		{
			body: `<root>
<name>foo</name>
<age>16</age>
</root>`,
			contentType: contentTypeXml + "; charset=utf-8",
		},
		{
			body:        `{"name":"foo","age":16}`,
			contentType: contentTypeYaml, // but json is provided - is still ok!
		},
		// bad mixes...
		{
			body:         `{"name":"foo","age":16}`,
			contentType:  contentTypeXml,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
		{
			body: `name: foo
age: 16`,
			contentType:  contentTypeJson,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
		{
			body: `name: foo
age: 16`,
			contentType:  contentTypeXml,
			expectErr:    true,
			expectStatus: http.StatusBadRequest,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set(hdrContentType, tc.contentType)
			}
			obj := &testMultiReq{}
			err := MultiUnmarshaler.Unmarshal(req, obj)
			if tc.expectErr {
				assert.Error(t, err)
				apiErr, ok := err.(ApiError)
				assert.True(t, ok)
				assert.Equal(t, tc.expectStatus, apiErr.StatusCode())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testMultiReq{Name: "foo", Age: 16}, *obj)
			}
		})
	}
}

type testMultiReq struct {
	Name string `json:"name" yaml:"name" xml:"name"`
	Age  int    `json:"age" yaml:"age" xml:"age"`
}
