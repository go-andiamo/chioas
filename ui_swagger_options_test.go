package chioas

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestSwaggerOptions_ToMap_Empty(t *testing.T) {
	o := &SwaggerOptions{}
	m := o.ToMap()
	assert.Empty(t, m)
}

func TestSwaggerOptions_ToMap_NonEmpty(t *testing.T) {
	o := &SwaggerOptions{
		DomId:                    "foo",
		Layout:                   "foo",
		DeepLinking:              true,
		DisplayOperationId:       true,
		DefaultModelsExpandDepth: 1,
		DefaultModelExpandDepth:  1,
		DefaultModelRendering:    "foo",
		DisplayRequestDuration:   true,
		DocExpansion:             "foo",
		MaxDisplayedTags:         1,
		ShowExtensions:           true,
		ShowCommonExtensions:     true,
		UseUnsafeMarkdown:        true,
		SyntaxHighlightTheme:     "foo",
		TryItOutEnabled:          true,
		RequestSnippetsEnabled:   true,
		SupportedSubmitMethods:   []string{"post"},
		ValidatorUrl:             "foo",
		WithCredentials:          true,
		PersistAuthorization:     true,
	}
	m := o.ToMap()
	assert.Equal(t, 20, len(m))
	assert.Contains(t, m, "dom_id")
	assert.Contains(t, m, "layout")
	assert.Contains(t, m, "deepLinking")
	assert.Contains(t, m, "displayOperationId")
	assert.Contains(t, m, "defaultModelsExpandDepth")
	assert.Contains(t, m, "defaultModelExpandDepth")
	assert.Contains(t, m, "defaultModelRendering")
	assert.Contains(t, m, "displayRequestDuration")
	assert.Contains(t, m, "docExpansion")
	assert.Contains(t, m, "maxDisplayedTags")
	assert.Contains(t, m, "showExtensions")
	assert.Contains(t, m, "showCommonExtensions")
	assert.Contains(t, m, "useUnsafeMarkdown")
	assert.Contains(t, m, "syntaxHighlight.theme")
	assert.Contains(t, m, "tryItOutEnabled")
	assert.Contains(t, m, "requestSnippetsEnabled")
	assert.Contains(t, m, "supportedSubmitMethods")
	assert.Contains(t, m, "validatorUrl")
	assert.Contains(t, m, "withCredentials")
	assert.Contains(t, m, "persistAuthorization")
}

func TestSwaggerOptions_OverrideFavIcons(t *testing.T) {
	o := SwaggerOptions{
		FavIcons: FavIcons{
			64: "test.png",
		},
	}
	icons := o.GetFavIcons()
	assert.NotNil(t, icons)
	html := icons.toHtml()
	assert.Equal(t, template.HTML(`<link rel="icon" type="image/png" href="./test.png" sizes="64x64" />`), html)
}
