package codegen

import (
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

const (
	defaultPackage = "api"
	defaultVarName = "definition"
	chioasPkg      = `"github.com/go-andiamo/chioas"`
	defaultAlias   = "chioas."
)

// Options is the options for GenerateCode
type Options struct {
	Package         string // e.g. "api" (default "api")
	VarName         string // e.g. "Spec" (default "definition" - or "Definition" when PublicVars is true)
	ImportAlias     string // e.g. "ch" (default none - i.e. "chioas")
	SkipPrologue    bool   // don't write package & imports
	OmitZeroValues  bool   // skip zero-valued fields
	HoistPaths      bool   // make top-level vars for root paths
	HoistComponents bool   // make top-level vars for named components (Schemas/Parameters/Responses/Requests/Examples)
	PublicVars      bool   // top-level vars are public (for paths & components)
	UseHttpConsts   bool   // if set, "GET" becomes http.MethodGet, 400 becomes http.StatusBadRequest etc.
	InlineHandlers  bool   // if set, generates stub inline handler funcs within the definition code
	// Format if set, formats output in canonical gofmt style (and checks syntax)
	//
	// Note: using this option means the output will be buffered before writing to the final writer
	Format  bool
	UseCRLF bool // true to use \r\n as the line terminator
}

func (o Options) topVarName() string {
	if o.VarName != "" {
		return o.VarName
	}
	return o.varName("", defaultVarName)
}

func (o Options) varName(kind string, name string) (result string) {
	if name == "" {
		name = "x"
	}
	if kind == "" {
		result = toPascal(name)
	} else {
		result = toPascal(kind + " " + name)
	}
	if !o.PublicVars {
		rs := []rune(result)
		rs[0] = unicode.ToLower(rs[0])
		result = string(rs)
	}
	if !unicode.IsLetter(rune(result[0])) {
		if o.PublicVars {
			result = "X" + result
		} else {
			result = "x" + result
		}
	}
	return result
}

func toPascal(s string) string {
	var b strings.Builder
	word := func(w string) {
		if w == "" {
			return
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	// keep only letters/digits, treat others as separators
	var cur strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
		} else {
			word(cur.String())
			cur.Reset()
		}
	}
	word(cur.String())
	return b.String()
}

func (o Options) alias() string {
	switch o.ImportAlias {
	case "":
		return defaultAlias
	case ".":
		return ""
	default:
		return o.ImportAlias + "."
	}
}

var (
	httpMethods = map[string]string{
		http.MethodGet:     "http.MethodGet",
		http.MethodHead:    "http.MethodHead",
		http.MethodPost:    "http.MethodPost",
		http.MethodPut:     "http.MethodPut",
		http.MethodPatch:   "http.MethodPatch",
		http.MethodDelete:  "http.MethodDelete",
		http.MethodOptions: "http.MethodOptions",
		http.MethodConnect: "http.MethodConnect",
		http.MethodTrace:   "http.MethodTrace",
	}
	httpStatusCodes = map[int]string{
		http.StatusContinue:                      "http.StatusContinue",
		http.StatusSwitchingProtocols:            "http.StatusSwitchingProtocols",
		http.StatusProcessing:                    "http.StatusProcessing",
		http.StatusEarlyHints:                    "http.StatusEarlyHints",
		http.StatusOK:                            "http.StatusOK",
		http.StatusCreated:                       "http.StatusCreated",
		http.StatusAccepted:                      "http.StatusAccepted",
		http.StatusNonAuthoritativeInfo:          "http.StatusNonAuthoritativeInfo",
		http.StatusNoContent:                     "http.StatusNoContent",
		http.StatusResetContent:                  "http.StatusResetContent",
		http.StatusPartialContent:                "http.StatusPartialContent",
		http.StatusMultiStatus:                   "http.StatusMultiStatus",
		http.StatusAlreadyReported:               "http.StatusAlreadyReported",
		http.StatusIMUsed:                        "http.StatusIMUsed",
		http.StatusMultipleChoices:               "http.StatusMultipleChoices",
		http.StatusMovedPermanently:              "http.StatusMovedPermanently",
		http.StatusFound:                         "http.StatusFound",
		http.StatusSeeOther:                      "http.StatusSeeOther",
		http.StatusNotModified:                   "http.StatusNotModified",
		http.StatusUseProxy:                      "http.StatusUseProxy",
		http.StatusTemporaryRedirect:             "http.StatusTemporaryRedirect",
		http.StatusPermanentRedirect:             "http.StatusPermanentRedirect",
		http.StatusBadRequest:                    "http.StatusBadRequest",
		http.StatusUnauthorized:                  "http.StatusUnauthorized",
		http.StatusPaymentRequired:               "http.StatusPaymentRequired",
		http.StatusForbidden:                     "http.StatusForbidden",
		http.StatusNotFound:                      "http.StatusNotFound",
		http.StatusMethodNotAllowed:              "http.StatusMethodNotAllowed",
		http.StatusNotAcceptable:                 "http.StatusNotAcceptable",
		http.StatusProxyAuthRequired:             "http.StatusProxyAuthRequired",
		http.StatusRequestTimeout:                "http.StatusRequestTimeout",
		http.StatusConflict:                      "http.StatusConflict",
		http.StatusGone:                          "http.StatusGone",
		http.StatusLengthRequired:                "http.StatusLengthRequired",
		http.StatusPreconditionFailed:            "http.StatusPreconditionFailed",
		http.StatusRequestEntityTooLarge:         "http.StatusRequestEntityTooLarge",
		http.StatusRequestURITooLong:             "http.StatusRequestURITooLong",
		http.StatusUnsupportedMediaType:          "http.StatusUnsupportedMediaType",
		http.StatusRequestedRangeNotSatisfiable:  "http.StatusRequestedRangeNotSatisfiable",
		http.StatusExpectationFailed:             "http.StatusExpectationFailed",
		http.StatusTeapot:                        "http.StatusTeapot",
		http.StatusMisdirectedRequest:            "http.StatusMisdirectedRequest",
		http.StatusUnprocessableEntity:           "http.StatusUnprocessableEntity",
		http.StatusLocked:                        "http.StatusLocked",
		http.StatusFailedDependency:              "http.StatusFailedDependency",
		http.StatusTooEarly:                      "http.StatusTooEarly",
		http.StatusUpgradeRequired:               "http.StatusUpgradeRequired",
		http.StatusPreconditionRequired:          "http.StatusPreconditionRequired",
		http.StatusTooManyRequests:               "http.StatusTooManyRequests",
		http.StatusRequestHeaderFieldsTooLarge:   "http.StatusRequestHeaderFieldsTooLarge",
		http.StatusUnavailableForLegalReasons:    "http.StatusUnavailableForLegalReasons",
		http.StatusInternalServerError:           "http.StatusInternalServerError",
		http.StatusNotImplemented:                "http.StatusNotImplemented",
		http.StatusBadGateway:                    "http.StatusBadGateway",
		http.StatusServiceUnavailable:            "http.StatusServiceUnavailable",
		http.StatusGatewayTimeout:                "http.StatusGatewayTimeout",
		http.StatusHTTPVersionNotSupported:       "http.StatusHTTPVersionNotSupported",
		http.StatusVariantAlsoNegotiates:         "http.StatusVariantAlsoNegotiates",
		http.StatusInsufficientStorage:           "http.StatusInsufficientStorage",
		http.StatusLoopDetected:                  "http.StatusLoopDetected",
		http.StatusNotExtended:                   "http.StatusNotExtended",
		http.StatusNetworkAuthenticationRequired: "http.StatusNetworkAuthenticationRequired",
	}
)

func (o Options) translateMethod(method string) string {
	if o.UseHttpConsts {
		if v, ok := httpMethods[method]; ok {
			return v
		}
	}
	return strconv.Quote(strings.ToUpper(method))
}

func (o Options) translateStatus(status int) string {
	if o.UseHttpConsts {
		if v, ok := httpStatusCodes[status]; ok {
			return v
		}
	}
	return strconv.Itoa(status)
}
