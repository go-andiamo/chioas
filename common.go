package chioas

var OasVersion = "3.0.3"

const (
	root = "/"

	tagNameApplicationJson = contentTypeJson
	tagNameComponents      = "components"
	tagNameContact         = "contact"
	tagNameContent         = "content"
	tagNameDefault         = "default"
	tagNameDeprecated      = "deprecated"
	tagNameDescription     = "description"
	tagNameEmail           = "email"
	tagNameEnum            = "enum"
	tagNameExample         = "example"
	tagNameExternalDocs    = "externalDocs"
	tagNameFormat          = "format"
	tagNameIn              = "in"
	tagNameInfo            = "info"
	tagNameItems           = "items"
	tagNameLicense         = "license"
	tagNameName            = "name"
	tagNameOpenApi         = "openapi"
	tagNameOperationId     = "operationId"
	tagNameParameters      = "parameters"
	tagNamePaths           = "paths"
	tagNameProperties      = "properties"
	tagNameRef             = "$ref"
	tagNameRequestBodies   = "requestBodies"
	tagNameRequestBody     = "requestBody"
	tagNameRequired        = "required"
	tagNameResponses       = "responses"
	tagNameSecurity        = "security"
	tagNameSchema          = "schema"
	tagNameSchemas         = "schemas"
	tagNameScheme          = "scheme"
	tagNameSecuritySchemes = "securitySchemes"
	tagNameServers         = "servers"
	tagNameSummary         = "summary"
	tagNameTags            = "tags"
	tagNameTermsOfService  = "termsOfService"
	tagNameTitle           = "title"
	tagNameType            = "type"
	tagNameUrl             = "url"
	tagNameVersion         = "version"

	refPathSchemas    = "#/" + tagNameComponents + "/" + tagNameSchemas + "/"
	refPathRequests   = "#/" + tagNameComponents + "/" + tagNameRequestBodies + "/"
	refPathResponses  = "#/" + tagNameComponents + "/" + tagNameResponses + "/"
	refPathParameters = "#/" + tagNameComponents + "/" + tagNameParameters + "/"

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

func nilString(v any) any {
	result := v
	switch vt := v.(type) {
	case string:
		if vt == "" {
			result = nil
		}
	}
	return result
}
