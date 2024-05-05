package chioas

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"math"
)

// OasVersion is the default OAS version for docs
var OasVersion = "3.0.3"

// ApplyMiddlewares is a function that returns middlewares to be applied to a Path or the api root
//
// can be used on Path.ApplyMiddlewares and Definition.ApplyMiddlewares (for api root)
type ApplyMiddlewares func(thisApi any) chi.Middlewares

// DisablerFunc is a function that can be used by Path.Disabled
type DisablerFunc func() bool

const (
	root = "/"

	hdrAllow                = "Allow"
	tagNameAnyOf            = "anyOf"
	tagNameAllOf            = "allOf"
	tagNameApplicationJson  = contentTypeJson
	tagNameComponents       = "components"
	tagNameContact          = "contact"
	tagNameContent          = "content"
	tagNameDefault          = "default"
	tagNameDeprecated       = "deprecated"
	tagNameDescription      = "description"
	tagNameDiscriminator    = "discriminator"
	tagNameEmail            = "email"
	tagNameEnum             = "enum"
	tagNameExample          = "example"
	tagNameExamples         = "examples"
	tagNameExclusiveMaximum = "exclusiveMaximum"
	tagNameExclusiveMinimum = "exclusiveMinimum"
	tagNameExternalDocs     = "externalDocs"
	tagNameFormat           = "format"
	tagNameIn               = "in"
	tagNameInfo             = "info"
	tagNameItems            = "items"
	tagNameItemType         = "itemType"
	tagNameLicense          = "license"
	tagNameMapping          = "mapping"
	tagNameMaxItems         = "maxItems"
	tagNameMaxLength        = "maxLength"
	tagNameMaxProperties    = "maxProperties"
	tagNameMaximum          = "maximum"
	tagNameMinItems         = "minItems"
	tagNameMinLength        = "minLength"
	tagNameMinProperties    = "minProperties"
	tagNameMinimum          = "minimum"
	tagNameMultipleOf       = "multipleOf"
	tagNameName             = "name"
	tagNameNullable         = "nullable"
	tagNameOneOf            = "oneOf"
	tagNameOpenApi          = "openapi"
	tagNameOperationId      = "operationId"
	tagNameParameters       = "parameters"
	tagNamePaths            = "paths"
	tagNamePattern          = "pattern"
	tagNameProperties       = "properties"
	tagNamePropertyName     = "propertyName"
	tagNameRef              = "$ref"
	tagNameRequestBodies    = "requestBodies"
	tagNameRequestBody      = "requestBody"
	tagNameRequired         = "required"
	tagNameResponses        = "responses"
	tagNameSchema           = "schema"
	tagNameSchemas          = "schemas"
	tagNameScheme           = "scheme"
	tagNameSecurity         = "security"
	tagNameSecuritySchemes  = "securitySchemes"
	tagNameServers          = "servers"
	tagNameSummary          = "summary"
	tagNameTags             = "tags"
	tagNameTermsOfService   = "termsOfService"
	tagNameTitle            = "title"
	tagNameType             = "type"
	tagNameUniqueItems      = "uniqueItems"
	tagNameUrl              = "url"
	tagNameValue            = "value"
	tagNameVersion          = "version"

	tagValuePath        = "path"
	tagValueQuery       = "query"
	tagValueTypeObject  = "object"
	tagValueTypeArray   = "array"
	tagValueTypeString  = "string"
	tagValueTypeNull    = "null"
	tagValueTypeInteger = "integer"
	tagValueTypeNumber  = "number"
	tagValueTypeBoolean = "boolean"
)

func nilString(v string) (result any) {
	result = v
	if v == "" {
		result = nil
	}
	return
}

func nilBool(v bool) (result any) {
	result = v
	if !v {
		result = nil
	}
	return
}

func nilNumber(n json.Number) (result any) {
	if n != "" {
		if i, err := n.Int64(); err == nil {
			result = i
		} else if f, err := n.Float64(); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			result = n
		}
	}
	return
}

func nilUint(v uint) (result any) {
	result = v
	if v == 0 {
		result = nil
	}
	return
}
