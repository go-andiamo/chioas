package typed

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

type outValueHandler = func(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool

type outsBuilder struct {
	len              int
	errArg           int
	statusCodeArg    int
	marshableArg     int
	marshableHandler outValueHandler
}

func newOutsBuilder(mf reflect.Value) (*outsBuilder, error) {
	mft := mf.Type()
	l := mft.NumOut()
	if l > 3 {
		return nil, errors.New(errTooMany)
	}
	result := &outsBuilder{
		len:           l,
		errArg:        -1,
		statusCodeArg: -1,
		marshableArg:  -1,
	}
	if err := result.makeHandlers(mft); err != nil {
		return nil, err
	}
	return result, nil
}

var interfaceTypeError = reflect.TypeOf((*error)(nil)).Elem()
var interfaceTypeResponseMarshaler = reflect.TypeOf((*ResponseMarshaler)(nil)).Elem()

const (
	errTooMany        = "has too many return args"
	errMultiErrs      = "has multiple error return args"
	errMultiMarshable = "has multiple marshalable return args"
	errMultiStatus    = "has multiple status code return args"
)

func (ob *outsBuilder) makeHandlers(mft reflect.Type) error {
	hasAnyArg := false
	for i := 0; i < ob.len; i++ {
		arg := mft.Out(i)
		switch arg.String() {
		case "interface {}":
			if ob.marshableArg != -1 {
				return errors.New(errMultiMarshable)
			}
			ob.marshableArg = i
			ob.marshableHandler = anyHandler
			hasAnyArg = true
		case "error":
			if ob.errArg != -1 {
				return errors.New(errMultiErrs)
			}
			ob.errArg = i
		case "typed.JsonResponse":
			if ob.marshableArg != -1 {
				return errors.New(errMultiMarshable)
			}
			ob.marshableArg = i
			ob.marshableHandler = jsonResponseHandler
		case "*typed.JsonResponse":
			if ob.marshableArg != -1 {
				return errors.New(errMultiMarshable)
			}
			ob.marshableArg = i
			ob.marshableHandler = jsonResponsePtrHandler
		case "[]byte", "[]uint8":
			if ob.marshableArg != -1 {
				return errors.New(errMultiMarshable)
			}
			ob.marshableArg = i
			ob.marshableHandler = bytesResponseHandler
		case "int":
			if ob.statusCodeArg != -1 {
				return errors.New(errMultiStatus)
			}
			ob.statusCodeArg = i
		default:
			if isErr := arg.Implements(interfaceTypeError); isErr {
				if ob.errArg != -1 {
					return errors.New(errMultiErrs)
				}
				ob.errArg = i
			} else if isRm := arg.Implements(interfaceTypeResponseMarshaler); isRm {
				if ob.marshableArg != -1 {
					return errors.New(errMultiMarshable)
				}
				ob.marshableArg = i
				ob.marshableHandler = responseMarshalerHandler
			} else {
				if ob.marshableArg != -1 {
					return errors.New(errMultiMarshable)
				}
				ob.marshableArg = i
				if arg.Kind() == reflect.Pointer {
					ob.marshableHandler = marshalerPtrHandler
				} else {
					ob.marshableHandler = marshalerHandler
				}
			}
		}
	}
	if hasAnyArg && ob.errArg == -1 {
		ob.marshableHandler = anyOrErrorHandler
	}
	return nil
}

func anyHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	result := false
	if v.Interface() != nil {
		switch v.Interface().(type) {
		case []uint8:
			result = bytesResponseHandler(v, b, thisApi, statusCode, writer, request)
		case JsonResponse:
			result = jsonResponseHandler(v, b, thisApi, statusCode, writer, request)
		case *JsonResponse:
			result = jsonResponsePtrHandler(v, b, thisApi, statusCode, writer, request)
		case ResponseMarshaler:
			result = responseMarshalerHandler(v, b, thisApi, statusCode, writer, request)
		default:
			result = marshalerHandler(v, b, thisApi, statusCode, writer, request)
		}
	}
	return result
}

func anyOrErrorHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	result := false
	switch v.Interface().(type) {
	case error, ApiError:
		result = true
		b.getErrorHandler(thisApi).HandleError(writer, request, v.Interface().(error))
	default:
		result = anyHandler(v, b, thisApi, statusCode, writer, request)
	}
	return result
}

func jsonResponseHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	jr := v.Interface().(JsonResponse)
	if err := jr.write(writer); err != nil {
		b.getErrorHandler(thisApi).HandleError(writer, request, err)
	}
	return true
}

func jsonResponsePtrHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	if !v.IsNil() {
		jr := v.Interface().(*JsonResponse)
		if err := jr.write(writer); err != nil {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	} else {
		writer.WriteHeader(defaultStatusCode(statusCode, http.StatusNoContent))
	}
	return true
}

func bytesResponseHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	if data := v.Interface().([]byte); len(data) > 0 {
		writer.WriteHeader(defaultStatusCode(statusCode, http.StatusOK))
		_, _ = writer.Write(data)
	} else {
		writer.WriteHeader(defaultStatusCode(statusCode, http.StatusNoContent))
	}
	return true
}

func responseMarshalerHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	if v.IsValid() && !v.IsZero() && !v.IsNil() {
		rm := v.Interface().(ResponseMarshaler)
		if data, sc, hdrs, err := rm.Marshal(request); err == nil {
			if len(data) == 0 {
				sc = defaultStatusCode(sc, defaultStatusCode(statusCode, http.StatusNoContent))
			} else {
				sc = defaultStatusCode(sc, defaultStatusCode(statusCode, http.StatusOK))
			}
			for _, hd := range hdrs {
				writer.Header().Set(hd[0], hd[1])
			}
			writer.WriteHeader(sc)
			_, _ = writer.Write(data)
		} else {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
		return true
	}
	return false
}

func marshalerHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	if data, err := json.Marshal(v.Interface()); err == nil {
		writer.Header().Set(hdrContentType, contentTypeJson)
		writer.WriteHeader(defaultStatusCode(statusCode, http.StatusOK))
		_, _ = writer.Write(data)
	} else {
		b.getErrorHandler(thisApi).HandleError(writer, request, err)
	}
	return true
}

func marshalerPtrHandler(v reflect.Value, b *builder, thisApi any, statusCode int, writer http.ResponseWriter, request *http.Request) bool {
	if v.IsNil() {
		return false
	}
	return marshalerHandler(v, b, thisApi, statusCode, writer, request)
}

func (ob *outsBuilder) handleReturnArgs(retArgs []reflect.Value, b *builder, thisApi any, writer http.ResponseWriter, request *http.Request) {
	if ob.errArg != -1 {
		if errArg := retArgs[ob.errArg]; errArg.IsValid() && !errArg.IsNil() {
			if rh := b.getResponseHandler(thisApi); rh != nil {
				rh.WriteErrorResponse(writer, request, errArg.Interface().(error), thisApi)
			} else {
				b.getErrorHandler(thisApi).HandleError(writer, request, errArg.Interface().(error))
			}
			return
		}
	}
	statusCode := 0
	if ob.statusCodeArg != -1 {
		statusCode = retArgs[ob.statusCodeArg].Interface().(int)
	}
	handled := false
	if ob.marshableArg != -1 && retArgs[ob.marshableArg].IsValid() {
		if rh := b.getResponseHandler(thisApi); rh != nil {
			handled = true
			rh.WriteResponse(writer, request, retArgs[ob.marshableArg].Interface(), statusCode, thisApi)
		} else {
			handled = ob.marshableHandler(retArgs[ob.marshableArg], b, thisApi, statusCode, writer, request)
		}
	}
	if !handled {
		writer.WriteHeader(defaultStatusCode(statusCode, http.StatusOK))
	}
}
