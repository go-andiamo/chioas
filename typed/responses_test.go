package typed

import (
	"encoding/json"
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
				assert.Equal(t, tc.expectStatus, res.Code)
				assert.Equal(t, tc.expect, res.Body.String())
				overrideCt := false
				for _, hdr := range tc.response.Headers {
					assert.Equal(t, hdr[1], res.Header().Get(hdr[0]))
					overrideCt = overrideCt || hdr[0] == hdrContentType
				}
				if !overrideCt {
					assert.Equal(t, contentTypeJson, res.Header().Get(hdrContentType))
				}
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			}
		})
	}
}

type badStruct struct {
}

func (b badStruct) MarshalJSON() ([]byte, error) {
	return nil, errors.New("fooey")
}
