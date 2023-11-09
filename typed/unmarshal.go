package typed

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"
)

// Unmarshaler is an interface that can be passed as an option to NewTypedMethodsHandlerBuilder, allowing support
// for unmarshalling different content types (e.g. overriding default JSON unmarshalling and/or varying the unmarshalling according to the request `Content-Type` header)
type Unmarshaler interface {
	Unmarshal(request *http.Request, v any) error
}

var defaultUnmarshaler Unmarshaler = &jsonUnmarshaler{}

type jsonUnmarshaler struct {
}

func (ju *jsonUnmarshaler) Unmarshal(request *http.Request, v any) error {
	return json.NewDecoder(request.Body).Decode(v)
}

var (
	_MultiUnmarshaler = &multiUnmarshaler{}
)
var (
	MultiUnmarshaler Unmarshaler = _MultiUnmarshaler // MultiUnmarshaler is an Unmarshaler that supports unmarshalling json, yaml and xml - and can be used as an option passed to NewTypedMethodsHandlerBuilder
)

type multiUnmarshaler struct {
}

const (
	contentTypeYaml    = "application/yaml"
	contentTypeYamlX   = "application/x-yaml"
	contentTypeYamlTxt = "text/yaml"
	contentTypeXml     = "application/xml"
	contentTypeXmlTxt  = "text/xml"
)

func (mu *multiUnmarshaler) Unmarshal(request *http.Request, v any) error {
	ct := request.Header.Get(hdrContentType)
	if cAt := strings.IndexByte(ct, ';'); cAt != -1 {
		ct = ct[:cAt]
	}
	if pAt := strings.IndexByte(ct, '+'); pAt != -1 {
		ct = ct[:pAt]
	}
	switch ct {
	case contentTypeJson:
		return WrapApiError(http.StatusBadRequest, json.NewDecoder(request.Body).Decode(v))
	case contentTypeYaml, contentTypeYamlX, contentTypeYamlTxt:
		return WrapApiError(http.StatusBadRequest, yaml.NewDecoder(request.Body).Decode(v))
	case contentTypeXml, contentTypeXmlTxt:
		return WrapApiError(http.StatusBadRequest, xml.NewDecoder(request.Body).Decode(v))
	default:
		return NewApiErrorf(http.StatusUnsupportedMediaType, "")
	}
}
