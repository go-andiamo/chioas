package typed

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/urit"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
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
	var err error
	for i := 0; i < inb.len && err == nil; i++ {
		arg := mft.In(i)
		inb.argTypes[i] = arg
		ok := false
		var countBody int
		argTypeStr := arg.String()
		if ok, countBody = inb.makeBuilderCommon(arg, i); ok {
			seenBody += countBody
		} else if ok, countBody = inb.makeBuilderFromArgBuilders(arg, i); ok {
			seenBody += countBody
		} else if ok, err = inb.makeTypedBuilder(arg, i); !ok && err == nil {
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
		if !ok && err == nil {
			err = fmt.Errorf("cannot determine arg %d", i)
		} else if seenBody > 1 {
			err = errors.New("multiple args could be from request.Body")
		}
	}
	return err
}

var (
	typeNamedQueryParam = reflect.TypeOf((*NamedQueryParam)(nil)).Elem()
	typeNamedPathParam  = reflect.TypeOf((*NamedPathParam)(nil)).Elem()
	typeNamedHeader     = reflect.TypeOf((*NamedHeader)(nil)).Elem()
	typeUnmarshalerText = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeNamedCookie     = reflect.TypeOf((*NamedCookie)(nil)).Elem()
	typeHttpCookie      = reflect.TypeOf(http.Cookie{})
)

func (inb *insBuilder) makeTypedBuilder(arg reflect.Type, i int) (ok bool, err error) {
	var isPtr bool
	var isSlice bool
	t := arg
	switch t.Kind() {
	case reflect.Ptr:
		t = t.Elem()
		isPtr = true
	case reflect.Slice:
		t = t.Elem()
		isSlice = true
	}
	pt := reflect.PointerTo(t)
	if ok = pt.Implements(typeNamedQueryParam); ok {
		umt := pt.Implements(typeUnmarshalerText)
		if t.Kind() != reflect.String && !umt {
			return false, errors.New("named query params must be of underlying type string or implement encoding.TextUnmarshaler")
		}
		name := reflect.New(t).Interface().(NamedQueryParam).QueryParamName()
		switch {
		case isPtr:
			inb.valueBuilders[i] = inb.makeNamedQueryParamPtr(name, umt)
		case isSlice:
			inb.valueBuilders[i] = inb.makeNamedQueryParamSlice(name, umt)
		default:
			inb.valueBuilders[i] = inb.makeNamedQueryParam(name, umt)
		}
	} else if ok = pt.Implements(typeNamedPathParam); ok {
		umt := pt.Implements(typeUnmarshalerText)
		if t.Kind() != reflect.String && !umt {
			return false, errors.New("named path params must be of underlying type string or implement encoding.TextUnmarshaler")
		}
		name := reflect.New(t).Interface().(NamedPathParam).PathParamName()
		switch {
		case isPtr:
			inb.valueBuilders[i] = inb.makeNamedPathParamPtr(name, umt)
		case isSlice:
			inb.valueBuilders[i] = inb.makeNamedPathParamSlice(name, umt)
		default:
			inb.valueBuilders[i] = inb.makeNamedPathParam(name, umt)
		}
	} else if ok = pt.Implements(typeNamedHeader); ok {
		if t.Kind() != reflect.String {
			return false, errors.New("named header must be of underlying type string")
		}
		name := reflect.New(t).Interface().(NamedHeader).HeaderName()
		switch {
		case isPtr:
			inb.valueBuilders[i] = inb.makeNamedHeaderPtr(name)
		case isSlice:
			inb.valueBuilders[i] = inb.makeNamedHeaderSlice(name)
		default:
			inb.valueBuilders[i] = inb.makeNamedHeader(name)
		}
	} else if ok = pt.Implements(typeNamedCookie); ok {
		if !isPtr || !t.ConvertibleTo(typeHttpCookie) {
			return false, errors.New("named cookie must be of underlying type *http.Cookie")
		}
		name := reflect.New(t).Interface().(NamedCookie).CookieName()
		inb.valueBuilders[i] = func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
			for _, c := range request.Cookies() {
				if c.Name == name {
					return reflect.ValueOf(c).Convert(argType), nil
				}
			}
			return reflect.New(argType).Elem(), nil
		}
	}
	return ok, err
}

type ParamError struct {
	Msg  string
	Name string
	Err  error
}

func (e *ParamError) Error() string {
	return fmt.Sprintf(e.Msg+": %q", e.Name)
}
func (e *ParamError) Unwrap() error {
	return e.Err
}
func wrapParamErr(err error, name string, msg string) error {
	if err == nil {
		return nil
	}
	return &ParamError{
		Msg:  msg,
		Name: name,
		Err:  err,
	}
}

func (inb *insBuilder) makeNamedQueryParam(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.New(argType)
			err = wrapParamErr(rv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(request.URL.Query().Get(name))), name, "invalid query param value")
			return rv.Elem(), err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		value := request.URL.Query().Get(name)
		return reflect.ValueOf(value).Convert(argType), nil
	}
}

func (inb *insBuilder) makeNamedQueryParamPtr(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.New(argType).Elem()
			if values, ok := request.URL.Query()[name]; ok && len(values) > 0 {
				value := values[0]
				av := reflect.New(argType.Elem())
				if err = wrapParamErr(av.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)), name, "invalid query param value"); err == nil {
					rv.Set(av)
				}
			}
			return rv, err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		if values, ok := request.URL.Query()[name]; ok && len(values) > 0 {
			value := values[0]
			return reflect.ValueOf(&value).Convert(argType), nil
		}
		return reflect.New(argType).Elem(), nil
	}
}

func (inb *insBuilder) makeNamedQueryParamSlice(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.MakeSlice(argType, 0, 0)
			if values, ok := request.URL.Query()[name]; ok {
				et := argType.Elem()
				for _, value := range values {
					av := reflect.New(et)
					if err = wrapParamErr(av.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)), name, "invalid query param value"); err == nil {
						rv = reflect.Append(rv, av.Elem())
					} else {
						break
					}
				}
			}
			return rv, err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		sl := reflect.MakeSlice(argType, 0, 0)
		if values, ok := request.URL.Query()[name]; ok {
			et := argType.Elem()
			for _, value := range values {
				sl = reflect.Append(sl, reflect.ValueOf(value).Convert(et))
			}
		}
		return sl, nil
	}
}

func (inb *insBuilder) makeNamedPathParam(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.New(argType)
			for _, pp := range params {
				if pp.Name == name {
					value := pp.Value.(string)
					err = wrapParamErr(rv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)), name, "invalid path param value")
					break
				}
			}
			return rv.Elem(), err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		result := reflect.New(argType).Elem()
		for _, pp := range params {
			if pp.Name == name {
				value := pp.Value.(string)
				result = reflect.ValueOf(value).Convert(argType)
				break
			}
		}
		return result, nil
	}
}

func (inb *insBuilder) makeNamedPathParamPtr(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.New(argType).Elem()
			for _, pp := range params {
				if pp.Name == name {
					value := pp.Value.(string)
					av := reflect.New(argType.Elem())
					if err = wrapParamErr(av.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)), name, "invalid path param value"); err == nil {
						rv.Set(av)
					}
					break
				}
			}
			return rv, err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		result := reflect.New(argType).Elem()
		for _, pp := range params {
			if pp.Name == name {
				value := pp.Value.(string)
				result = reflect.ValueOf(&value).Convert(argType)
				break
			}
		}
		return result, nil
	}
}

func (inb *insBuilder) makeNamedPathParamSlice(name string, unmarshalText bool) inValueBuilder {
	if unmarshalText {
		return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (rv reflect.Value, err error) {
			rv = reflect.MakeSlice(argType, 0, 0)
			et := argType.Elem()
			for _, pp := range params {
				if pp.Name == name {
					value := pp.Value.(string)
					av := reflect.New(et)
					if err = wrapParamErr(av.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value)), name, "invalid path param value"); err == nil {
						rv = reflect.Append(rv, av.Elem())
					} else {
						break
					}
				}
			}
			return rv, err
		}
	}
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		sl := reflect.MakeSlice(argType, 0, len(params))
		et := argType.Elem()
		for _, pp := range params {
			if pp.Name == name {
				value := pp.Value.(string)
				sl = reflect.Append(sl, reflect.ValueOf(value).Convert(et))
			}
		}
		return sl, nil
	}
}

func (inb *insBuilder) makeNamedHeader(name string) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		value := request.Header.Get(name)
		return reflect.ValueOf(value).Convert(argType), nil
	}
}

func (inb *insBuilder) makeNamedHeaderPtr(name string) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		if values := request.Header.Values(name); len(values) > 0 {
			value := values[0]
			return reflect.ValueOf(&value).Convert(argType), nil
		}
		return reflect.New(argType).Elem(), nil
	}
}

func (inb *insBuilder) makeNamedHeaderSlice(name string) inValueBuilder {
	return func(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
		sl := reflect.MakeSlice(argType, 0, 0)
		if values, ok := request.Header[name]; ok {
			et := argType.Elem()
			for _, value := range values {
				sl = reflect.Append(sl, reflect.ValueOf(value).Convert(et))
			}
		}
		return sl, nil
	}
}

func isExcPackage(argType string) bool {
	if strings.HasPrefix(argType, "[]") {
		argType = argType[2:]
	}
	return strings.HasPrefix(argType, "http.") || strings.HasPrefix(argType, "*http.") ||
		strings.HasPrefix(argType, "multipart.") || strings.HasPrefix(argType, "*multipart.") ||
		strings.HasPrefix(argType, "url.") || strings.HasPrefix(argType, "*url.")
}

var (
	typeResponseWriter = reflect.TypeFor[http.ResponseWriter]()
	typeRequest        = reflect.TypeFor[*http.Request]()
	typeContext        = reflect.TypeFor[context.Context]()
	typeChiContextPtr  = reflect.TypeFor[*chi.Context]()
	typeChiContext     = reflect.TypeFor[chi.Context]()
	typeHeader         = reflect.TypeFor[http.Header]()
	typeCookies        = reflect.TypeFor[[]*http.Cookie]()
	typeUrl            = reflect.TypeFor[*url.URL]()
	typeHeaders        = reflect.TypeFor[Headers]()
	typeMapHeaders     = reflect.TypeFor[map[string][]string]()
	typePathParams     = reflect.TypeFor[PathParams]()
	typeQueryParams    = reflect.TypeFor[QueryParams]()
	typeRawQuery       = reflect.TypeFor[RawQuery]()
	typeRawMessage     = reflect.TypeFor[json.RawMessage]()
	typePostForm       = reflect.TypeFor[PostForm]()
	typeBasicAuth      = reflect.TypeFor[BasicAuth]()
	typeBasicAuthPtr   = reflect.TypeFor[*BasicAuth]()
	typeBodySliceUint8 = reflect.TypeFor[[]uint8]()
	typeBodySliceByte  = reflect.TypeFor[[]byte]()
)

func (inb *insBuilder) makeBuilderCommon(arg reflect.Type, i int) (ok bool, countBody int) {
	ok = true
	switch arg {
	case typeResponseWriter:
		inb.valueBuilders[i] = commonBuilderHttpResponseWriter
	case typeRequest:
		inb.valueBuilders[i] = commonBuilderHttpRequest
	case typeContext:
		inb.valueBuilders[i] = commonBuilderContext
	case typeChiContextPtr:
		inb.valueBuilders[i] = commonBuilderChiContextPtr
	case typeChiContext:
		inb.valueBuilders[i] = commonBuilderChiContext
	case typeHeader:
		inb.valueBuilders[i] = commonBuilderHttpHeader
	case typeCookies:
		inb.valueBuilders[i] = commonBuilderCookies
	case typeUrl:
		inb.valueBuilders[i] = commonBuilderUrl
	case typeHeaders:
		inb.valueBuilders[i] = commonBuilderTypedHeaders
	case typeMapHeaders:
		inb.valueBuilders[i] = commonBuilderMapHeaders
	case typePathParams:
		inb.valueBuilders[i] = commonBuilderTypedPathParams
	case typeQueryParams:
		inb.valueBuilders[i] = commonBuilderQueryParams
	case typeRawQuery:
		inb.valueBuilders[i] = commonBuilderRawQuery
	case typeRawMessage:
		inb.valueBuilders[i] = commonBuilderJsonRawMessage
		countBody = 1
	case typePostForm:
		if inb.method == http.MethodPost || inb.method == http.MethodPut || inb.method == http.MethodPatch {
			inb.valueBuilders[i] = commonBuilderPostForm
			countBody = 1
		} else {
			inb.valueBuilders[i] = commonBuilderPostFormEmpty
		}
	case typeBasicAuth:
		inb.valueBuilders[i] = commonBuilderBasicAuth
	case typeBasicAuthPtr:
		inb.valueBuilders[i] = commonBuilderBasicAuthPtr
	case typeBodySliceUint8, typeBodySliceByte:
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
		if matchParams, ok := inb.pathTemplate.MatchesRequest(request); ok {
			params = matchParams.GetAll()
		}
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
