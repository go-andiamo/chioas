package typed

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJsonResponse_Write(t *testing.T) {
	testCases := []struct {
		response          JsonResponse
		expectErr         string
		expectStatus      int
		expect            string
		expectContentType string
	}{
		{
			expectStatus: http.StatusNoContent,
		},
		{
			response:     JsonResponse{},
			expectStatus: http.StatusNoContent,
		},
		{
			response: JsonResponse{
				Headers: [][2]string{{hdrContentType, "text/plain"}},
			},
			expectStatus: http.StatusNoContent,
		},
		{
			response: JsonResponse{
				Error: errors.New("fooey"),
			},
			expectErr: `fooey`,
		},
		{
			response: JsonResponse{
				Body: json.RawMessage([]byte{}),
			},
			expectStatus: http.StatusNoContent,
			expect:       ``,
		},
		{
			response: JsonResponse{
				StatusCode: http.StatusPaymentRequired,
				Body:       json.RawMessage([]byte{}),
			},
			expectStatus: http.StatusPaymentRequired,
			expect:       ``,
		},
		{
			response: JsonResponse{
				Body: json.RawMessage([]byte{'{', '}'}),
			},
			expectStatus: http.StatusOK,
			expect:       `{}`,
		},
		{
			response: JsonResponse{
				Body: map[string]any{"foo": "bar"},
			},
			expectStatus: http.StatusOK,
			expect:       `{"foo":"bar"}`,
		},
		{
			response: JsonResponse{
				Body: badStruct{},
			},
			expectErr: "json: error calling MarshalJSON for type typed.badStruct: fooey",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			res := httptest.NewRecorder()
			err := tc.response.write(res)
			if tc.expectErr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectStatus, res.Result().StatusCode)
				assert.Equal(t, tc.expect, res.Body.String())
				overrideCt := false
				for _, hdr := range tc.response.Headers {
					assert.Equal(t, hdr[1], res.Result().Header.Get(hdr[0]))
					overrideCt = overrideCt || hdr[0] == hdrContentType
				}
				if !overrideCt {
					assert.Equal(t, contentTypeJson, res.Result().Header.Get(hdrContentType))
				}
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			}
		})
	}
}

func TestMultiResponseMarshaler(t *testing.T) {
	testCases := []struct {
		marshaler    ResponseMarshaler
		accept       string
		expectStatus int
		expectErr    bool
		expectApiErr bool
		expectHdrs   [][2]string
		expectData   string
	}{
		{
			marshaler:    &MultiResponseMarshaler{},
			expectErr:    true,
			expectApiErr: true,
			expectStatus: http.StatusNotAcceptable,
		},
		{
			marshaler:    &MultiResponseMarshaler{},
			accept:       contentTypeJson,
			expectStatus: http.StatusNoContent,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: goodStruct{Foo: "bar"},
			},
			accept:       contentTypeJson,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body:    goodStruct{Foo: "bar"},
				Headers: [][2]string{{"foo", "bar"}},
			},
			accept:       contentTypeJson,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{"foo", "bar"}, {hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body:    goodStruct{Foo: "bar"},
				Headers: [][2]string{{"foo", "bar"}, {hdrContentType, "contentTypeJson"}},
			},
			accept:       contentTypeJson,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{"foo", "bar"}, {hdrContentType, "contentTypeJson"}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: goodStruct{Foo: "bar"},
			},
			accept:       contentTypeJson + "+ext; charset=utf-8",
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body:                goodStruct{Foo: "bar"},
				FallbackContentType: contentTypeJson,
			},
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: badStruct{},
			},
			accept:    contentTypeJson,
			expectErr: true,
		},
		{
			marshaler: &MultiResponseMarshaler{
				ExcludeJson: true,
			},
			accept:       contentTypeJson,
			expectStatus: http.StatusNotAcceptable,
			expectErr:    true,
			expectApiErr: true,
		},
		{
			marshaler:    &MultiResponseMarshaler{},
			accept:       contentTypeYaml,
			expectStatus: http.StatusNoContent,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeYaml}},
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: goodStruct{Foo: "bar"},
			},
			accept:       contentTypeYaml,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeYaml}},
			expectData: `foo: bar
`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body:                goodStruct{Foo: "bar"},
				FallbackContentType: contentTypeYaml,
			},
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeYaml}},
			expectData: `foo: bar
`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: badStruct{},
			},
			accept:    contentTypeYaml,
			expectErr: true,
		},
		{
			marshaler: &MultiResponseMarshaler{
				ExcludeYaml: true,
			},
			accept:       contentTypeYaml,
			expectErr:    true,
			expectApiErr: true,
			expectStatus: http.StatusNotAcceptable,
		},
		{
			marshaler:    &MultiResponseMarshaler{},
			accept:       contentTypeXml,
			expectStatus: http.StatusNoContent,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeXml}},
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: struct {
					XMLName xml.Name `xml:"item"`
					Foo     string   `xml:"foo"`
				}{
					Foo: "bar",
				},
			},
			accept:       contentTypeXml,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeXml}},
			expectData:   `<item><foo>bar</foo></item>`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: struct {
					XMLName xml.Name `xml:"item"`
					Foo     string   `xml:"foo"`
				}{
					Foo: "bar",
				},
				FallbackContentType: contentTypeXml,
			},
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeXml}},
			expectData:   `<item><foo>bar</foo></item>`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body: badStruct{},
			},
			accept:    contentTypeXml,
			expectErr: true,
		},
		{
			marshaler: &MultiResponseMarshaler{
				ExcludeXml: true,
			},
			accept:       contentTypeXml,
			expectErr:    true,
			expectApiErr: true,
			expectStatus: http.StatusNotAcceptable,
		},
		{
			marshaler:    &MultiResponseMarshaler{},
			accept:       `text/plain`,
			expectErr:    true,
			expectApiErr: true,
			expectStatus: http.StatusNotAcceptable,
		},
		{
			marshaler: &MultiResponseMarshaler{
				Body:                goodStruct{Foo: "bar"},
				FallbackContentType: contentTypeJson,
			},
			accept:       `text/plain`,
			expectStatus: http.StatusOK,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				StatusCode:          http.StatusPaymentRequired,
				Body:                goodStruct{Foo: "bar"},
				FallbackContentType: contentTypeJson,
			},
			accept:       `text/plain`,
			expectStatus: http.StatusPaymentRequired,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
			expectData:   `{"foo":"bar"}`,
		},
		{
			marshaler: &MultiResponseMarshaler{
				StatusCode:          http.StatusPaymentRequired,
				FallbackContentType: contentTypeJson,
			},
			accept:       `text/plain`,
			expectStatus: http.StatusPaymentRequired,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
		},
		{
			marshaler: &MultiResponseMarshaler{
				FallbackContentType: contentTypeJson,
			},
			accept:       `text/plain`,
			expectStatus: http.StatusNoContent,
			expectHdrs:   [][2]string{{hdrContentType, contentTypeJson}},
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(hdrAccept, tc.accept)
			data, statusCode, hdrs, err := tc.marshaler.Marshal(req)
			if tc.expectErr {
				assert.Error(t, err)
				if tc.expectApiErr {
					apiErr, ok := err.(ApiError)
					assert.True(t, ok)
					assert.Equal(t, tc.expectStatus, apiErr.StatusCode())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectStatus, statusCode)
				assert.Equal(t, tc.expectHdrs, hdrs)
				assert.Equal(t, tc.expectData, string(data))
			}
		})
	}
}

type goodStruct struct {
	XMLName xml.Name `xml:"item" json:"-" yaml:"-"`
	Foo     string   `json:"foo" yaml:"foo" xml:"foo"`
}

type badStruct struct {
}

func (b badStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("fooey")
}

func (b badStruct) MarshalYAML() (any, error) {
	return nil, errors.New("fooey")
}

func (b badStruct) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return errors.New("fooey")
}
