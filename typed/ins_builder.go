package typed

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/urit"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type insBuilder struct {
	len           int
	valueBuilders []inValueBuilder
	argTypes      []reflect.Type
	path          string
	method        string
	pathTemplate  urit.Template
	isVaradic     bool
	parentBuilder *builder
}

func newInsBuilder(mf reflect.Value, path string, method string, parentBuilder *builder) (*insBuilder, error) {
	pathTemplate, err := urit.NewTemplate(path)
	if err != nil {
		return nil, err
	}
	mft := mf.Type()
	l := mft.NumIn()
	result := &insBuilder{
		len:           l,
		valueBuilders: make([]inValueBuilder, l),
		argTypes:      make([]reflect.Type, l),
		path:          path,
		method:        method,
		pathTemplate:  pathTemplate,
		isVaradic:     mft.IsVariadic(),
		parentBuilder: parentBuilder,
	}
	if err := result.makeBuilders(mft); err != nil {
		return nil, err
	}
	return result, nil
}

func (inb *insBuilder) makeBuilders(mft reflect.Type) error {
	pathParam := 0
	seenBody := 0
	for i := 0; i < inb.len; i++ {
		arg := mft.In(i)
		inb.argTypes[i] = arg
		ok := false
		var countBody int
		argTypeStr := arg.String()
		if ok, countBody = inb.makeBuilderCommon(argTypeStr, i); ok {
			seenBody += countBody
		} else if ok, countBody = inb.makeBuilderFromArgBuilders(arg, i); ok {
			seenBody += countBody
		} else {
			switch arg.Kind() {
			case reflect.String:
				inb.valueBuilders[i] = newPathParamBuilder(pathParam)
				pathParam++
				ok = true
			case reflect.Slice:
				switch arg.Elem().Kind() {
				case reflect.String:
					inb.valueBuilders[i] = newPathParamsBuilder(pathParam)
					ok = true
				case reflect.Struct:
					if ok = !isExcPackage(argTypeStr); ok {
						inb.valueBuilders[i] = newSliceParamBuilder(arg, inb.parentBuilder.unmarshaler)
						seenBody++
					}
				}
			case reflect.Struct:
				if ok = !isExcPackage(argTypeStr); ok {
					inb.valueBuilders[i] = newStructParamBuilder(arg, inb.parentBuilder.unmarshaler)
					seenBody++
				}
			case reflect.Pointer:
				if ok = arg.Elem().Kind() == reflect.Struct && !isExcPackage(argTypeStr); ok {
					inb.valueBuilders[i] = newStructPtrParamBuilder(arg, inb.parentBuilder.unmarshaler)
					seenBody++
				}
			}
		}
		if !ok {
			return fmt.Errorf("cannot determine arg %d", i)
		} else if seenBody > 1 {
			return errors.New("multiple args could be from request.Body")
		}
	}
	return nil
}

func isExcPackage(argType string) bool {
	if strings.HasPrefix(argType, "[]") {
		argType = argType[2:]
	}
	return strings.HasPrefix(argType, "http.") || strings.HasPrefix(argType, "*http.") ||
		strings.HasPrefix(argType, "multipart.") || strings.HasPrefix(argType, "*multipart.") ||
		strings.HasPrefix(argType, "url.") || strings.HasPrefix(argType, "*url.")
}

func (inb *insBuilder) makeBuilderCommon(argType string, i int) (ok bool, countBody int) {
	ok = true
	switch argType {
	case "http.ResponseWriter":
		inb.valueBuilders[i] = commonBuilderHttpResponseWriter
	case "*http.Request":
		inb.valueBuilders[i] = commonBuilderHttpRequest
	case "context.Context":
		inb.valueBuilders[i] = commonBuilderContext
	case "*chi.Context":
		inb.valueBuilders[i] = commonBuilderChiContextPtr
	case "chi.Context":
		inb.valueBuilders[i] = commonBuilderChiContext
	case "http.Header":
		inb.valueBuilders[i] = commonBuilderHttpHeader
	case "[]*http.Cookie":
		inb.valueBuilders[i] = commonBuilderCookies
	case "*url.URL":
		inb.valueBuilders[i] = commonBuilderUrl
	case "typed.Headers":
		inb.valueBuilders[i] = commonBuilderTypedHeaders
	case "map[string][]string":
		inb.valueBuilders[i] = commonBuilderMapHeaders
	case "typed.PathParams":
		inb.valueBuilders[i] = commonBuilderTypedPathParams
	case "typed.QueryParams":
		inb.valueBuilders[i] = commonBuilderQueryParams
	case "typed.RawQuery":
		inb.valueBuilders[i] = commonBuilderRawQuery
	case "json.RawMessage":
		inb.valueBuilders[i] = commonBuilderJsonRawMessage
		countBody = 1
	case "typed.PostForm":
		if inb.method == http.MethodPost || inb.method == http.MethodPut || inb.method == http.MethodPatch {
			inb.valueBuilders[i] = commonBuilderPostForm
			countBody = 1
		} else {
			inb.valueBuilders[i] = commonBuilderPostFormEmpty
		}
	case "typed.BasicAuth":
		inb.valueBuilders[i] = commonBuilderBasicAuth
	case "*typed.BasicAuth":
		inb.valueBuilders[i] = commonBuilderBasicAuthPtr
	case "[]uint8", "[]byte":
		inb.valueBuilders[i] = commonBuilderByteBody
		countBody = 1
	default:
		ok = false
	}
	return
}

func (inb *insBuilder) makeBuilderFromArgBuilders(arg reflect.Type, i int) (ok bool, countBody int) {
	if !inb.isVaradic || (inb.isVaradic && i < inb.len-1) {
		isBody := false
		for _, a := range inb.parentBuilder.argBuilders {
			if ok, isBody = a.IsApplicable(arg, inb.method, inb.path); ok {
				b := a
				inb.valueBuilders[i] = func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
					if rv, err := b.BuildValue(argType, request, params); err == nil {
						return rv, nil
					} else {
						return reflect.Value{}, err
					}
				}
				if isBody {
					countBody = 1
				}
				break
			}
		}
	}
	return
}

func commonBuilderHttpResponseWriter(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(writer), nil
}
func commonBuilderHttpRequest(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(request), nil
}
func commonBuilderContext(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(request.Context()), nil
}
func commonBuilderChiContextPtr(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	ctx := chi.RouteContext(request.Context())
	return reflect.ValueOf(ctx), nil
}
func commonBuilderChiContext(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if ctx := chi.RouteContext(request.Context()); ctx != nil {
		return reflect.ValueOf(*ctx), nil
	}
	return reflect.ValueOf(chi.Context{}), nil
}
func commonBuilderHttpHeader(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(request.Header), nil
}
func commonBuilderCookies(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(request.Cookies()), nil
}
func commonBuilderUrl(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(request.URL), nil
}
func commonBuilderTypedHeaders(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	hdrs := Headers(request.Header)
	return reflect.ValueOf(hdrs), nil
}
func commonBuilderMapHeaders(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	hdrs := map[string][]string(request.Header)
	return reflect.ValueOf(hdrs), nil
}
func commonBuilderTypedPathParams(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	result := PathParams{}
	for _, pp := range params {
		result[pp.Name] = append(result[pp.Name], pp.Value.(string))
	}
	return reflect.ValueOf(result), nil
}
func commonBuilderQueryParams(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if request.URL == nil {
		return reflect.ValueOf(QueryParams{}), nil
	}
	result := QueryParams(request.URL.Query())
	return reflect.ValueOf(result), nil
}
func commonBuilderRawQuery(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if request.URL == nil {
		return reflect.ValueOf(RawQuery("")), nil
	}
	result := RawQuery(request.URL.RawQuery)
	return reflect.ValueOf(result), nil
}
func commonBuilderJsonRawMessage(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if request.Body == nil {
		return reflect.ValueOf(json.RawMessage([]byte{})), nil
	} else if data, err := io.ReadAll(request.Body); err == nil {
		return reflect.ValueOf(json.RawMessage(data)), nil
	} else {
		return reflect.Value{}, err
	}
}
func commonBuilderPostForm(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if err := request.ParseForm(); err == nil {
		return reflect.ValueOf(PostForm(request.PostForm)), nil
	} else {
		return reflect.Value{}, err
	}
}
func commonBuilderPostFormEmpty(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	return reflect.ValueOf(PostForm{}), nil
}
func commonBuilderBasicAuth(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	result := BasicAuth{}
	result.Username, result.Password, result.Ok = request.BasicAuth()
	return reflect.ValueOf(result), nil
}
func commonBuilderBasicAuthPtr(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	var result *BasicAuth
	if auth := request.Header.Get("Authorization"); auth != "" && strings.HasPrefix(auth, "Basic ") {
		result = &BasicAuth{}
		result.Username, result.Password, result.Ok = request.BasicAuth()
	}
	return reflect.ValueOf(result), nil
}
func commonBuilderByteBody(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	if request.Body == nil {
		return reflect.ValueOf([]byte{}), nil
	} else if data, err := io.ReadAll(request.Body); err == nil {
		return reflect.ValueOf(data), nil
	} else {
		return reflect.Value{}, err
	}
}

func newPathParamBuilder(index int) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		if index <= len(params)-1 {
			return reflect.ValueOf(params[index].Value), nil
		}
		return reflect.ValueOf(""), nil
	}
}

func newPathParamsBuilder(from int) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		remaining := make([]string, 0, len(params))
		for i := from; i < len(params); i++ {
			remaining = append(remaining, params[i].Value.(string))
		}
		return reflect.ValueOf(remaining), nil
	}
}

func newStructParamBuilder(argT reflect.Type, um Unmarshaler) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		v := reflect.New(argT)
		av := v.Interface()
		if request.Body != nil {
			if err := um.Unmarshal(request, av); err != nil {
				return v, err
			}
		}
		return reflect.Indirect(reflect.ValueOf(av)), nil
	}
}

func newStructPtrParamBuilder(argT reflect.Type, um Unmarshaler) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		if request.Body != nil {
			v := reflect.New(argT.Elem())
			av := v.Interface()
			if err := um.Unmarshal(request, av); err != nil {
				return v, err
			}
			return reflect.ValueOf(av), nil
		} else {
			v := reflect.New(argT)
			return reflect.Indirect(v), nil
		}
	}
}

func newSliceParamBuilder(argT reflect.Type, um Unmarshaler) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		if request.Body != nil {
			vs := reflect.New(argT)
			av := vs.Interface()
			if err := um.Unmarshal(request, &av); err != nil {
				return vs, err
			}
			if av != nil {
				return reflect.Indirect(reflect.ValueOf(av)), nil
			}
		}
		return reflect.Indirect(reflect.New(argT)), nil
	}
}

func (inb *insBuilder) build(writer http.ResponseWriter, request *http.Request) ([]reflect.Value, error) {
	var params []urit.PathVar
	if ctx := chi.RouteContext(request.Context()); ctx != nil {
		params = make([]urit.PathVar, 0, len(ctx.URLParams.Keys))
		pos := 0
		npos := map[string]int{}
		for i, k := range ctx.URLParams.Keys {
			if k != "*" {
				params = append(params, urit.PathVar{
					Position:      pos,
					NamedPosition: npos[k],
					Name:          k,
					Value:         ctx.URLParams.Values[i],
				})
				pos++
				npos[k] = npos[k] + 1
			}
		}
	} else {
		// fallback path params - if context is not Chi (usually during direct testing)
		matchParams, ok := inb.pathTemplate.MatchesRequest(request)
		if !ok {
			return nil, errors.New("unable to extract path params")
		}
		params = matchParams.GetAll()
	}
	result := make([]reflect.Value, 0, inb.len)
	for i, valueBuilder := range inb.valueBuilders {
		if inb.isVaradic && i == inb.len-1 {
			if v, err := valueBuilder(inb.argTypes[i], writer, request, params); err == nil {
				l := v.Len()
				for j := 0; j < l; j++ {
					result = append(result, v.Index(j))
				}
			} else {
				return nil, err
			}
		} else if v, err := valueBuilder(inb.argTypes[i], writer, request, params); err == nil {
			result = append(result, v)
		} else {
			return nil, err
		}
	}
	return result, nil
}

type inValueBuilder = func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error)
