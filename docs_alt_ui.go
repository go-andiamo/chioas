package chioas

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
)

type AlternateUIDocs map[string]AlternateUIDoc

type AlternateUIDoc struct {
	// DocIndexPage the name of the docs index page (defaults to "index.html")
	DocIndexPage string
	// SupportFiles is an optional handler that is used for other files under "/docs" path
	//
	// see _examples/swagger_ui for example usage
	SupportFiles http.Handler
	// SupportFilesStripPrefix if set to true (and SupportFiles is specified) then
	// calls to SupportFiles have the "/docs" path prefix stripped from the http.Request
	SupportFilesStripPrefix bool
	// Title the title in the docs index page (defaults to "API Documentation")
	Title string
	// UIStyle is the style of the API docs UI
	//
	// use Redoc, Swagger or Rapidoc (defaults to Redoc)
	UIStyle UIStyle
	// RedocOptions redoc options to be used (see https://github.com/Redocly/redoc#redoc-options-object)
	//
	// use map[string]any or &RedocOptions or anything that implements ToMap (or anything that can be marshalled and then unmarshalled to map[string]any)
	//
	// Only used if DocOptions.UIStyle is Redoc
	RedocOptions any
	// SwaggerOptions swagger-ui options to be used (see https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/configuration.md)
	//
	// Only used if DocOptions.UIStyle is Swagger
	SwaggerOptions any
	// RapidocOptions rapidoc options to be used
	//
	// Only used if DocOptions.UIStyle is Rapidoc
	RapidocOptions any
	// SpecName the name of the OAS spec (defaults to "spec.yaml")
	SpecName string
	// DocTemplate template for the docs page (defaults to internal template if an empty string)
	DocTemplate string
	// StylesOverride css styling overrides (injected into docs index page)
	StylesOverride string
	// RedocJsUrl is the URL for the Redoc JS
	//
	// defaults to:
	// https://cdn.jsdelivr.net/npm/redoc@2.0.0-rc.77/bundles/redoc.standalone.min.js
	//
	// Only used if DocOptions.UIStyle is Redoc
	RedocJsUrl string
	// TryJsUrl is the URL for the Try Redoc JS
	//
	// defaults to:
	// https://cdn.jsdelivr.net/gh/wll8/redoc-try@1.4.7/dist/try.js
	//
	// Only used if DocOptions.UIStyle is Redoc
	TryJsUrl string
	// Middlewares is any chi.Middlewares for everything served under this docs path
	Middlewares chi.Middlewares
	// NoCache if set to true, docs page and spec aren't cached (and built on each call)
	NoCache bool
	// AsJson if set to true, serves the spec as JSON
	AsJson bool
}

func (d *AlternateUIDoc) setupRoute(def *Definition, route chi.Router, path string, specData []byte) error {
	tmp, err := getTemplate(d.UIStyle, d.DocTemplate)
	if err != nil {
		return err
	}
	indexPage := defValue(d.DocIndexPage, defaultIndexName)
	specName, data := d.getTemplateData()
	docsRoute := chi.NewRouter()
	docsRoute.Use(d.Middlewares...)
	redirectPath := path + root + indexPage
	docsRoute.Get(root, http.RedirectHandler(redirectPath, http.StatusMovedPermanently).ServeHTTP)
	if specData != nil || !d.NoCache {
		if err = setupCachedRoutes(def, d.AsJson, specData, docsRoute, tmp, data, indexPage, specName); err != nil {
			return err
		}
	} else {
		setupNoCachedRoutes(def, d.AsJson, docsRoute, tmp, data, indexPage, specName)
	}
	setupSupportFiles(defValue(path, defaultDocsPath), d.UIStyle, d.SupportFiles, d.SupportFilesStripPrefix, docsRoute)
	route.Mount(path, docsRoute)
	return nil
}

func (d *AlternateUIDoc) getTemplateData() (specName string, data map[string]any) {
	if d.SpecName != "" {
		specName = d.SpecName
	} else if d.AsJson {
		specName = defaultSpecNameJson
	} else {
		specName = defaultSpecName
	}
	switch d.UIStyle {
	case Swagger:
		swaggerOpts, presets, plugins := getSwaggerOptions(d.SwaggerOptions, specName)
		data = map[string]any{
			htmlTagTitle:          defValue(d.Title, defaultTitle),
			htmlTagStylesOverride: template.CSS(d.StylesOverride),
			htmlTagSwaggerOpts:    swaggerOpts,
			htmlTagSwaggerPresets: presets,
			htmlTagSwaggerPlugins: plugins,
			htmlTagFavIcons:       optionsFavIcons(d.SwaggerOptions),
		}
		addHeaderAndScripts(data, swaggerOpts)
	case Rapidoc:
		data = optionsToMap(d.RapidocOptions)
		data[htmlTagTitle] = defValue(d.Title, defaultTitle)
		data[htmlTagStylesOverride] = template.CSS(d.StylesOverride)
		data[htmlTagSpecName] = specName
		data[htmlTagFavIcons] = optionsFavIcons(d.RapidocOptions)
	default:
		redocOpts := optionsToMap(d.RedocOptions)
		data = map[string]any{
			htmlTagTitle:          defValue(d.Title, defaultTitle),
			htmlTagStylesOverride: template.CSS(defValue(d.StylesOverride, defaultRedocStylesOverride)),
			htmlTagSpecName:       specName,
			htmlTagRedocOpts:      redocOpts,
			htmlTagRedocUrl:       defValue(d.RedocJsUrl, defaultRedocJsUrl),
			htmlTagTryUrl:         defValue(d.TryJsUrl, defaultTryJsUrl),
			htmlTagFavIcons:       optionsFavIcons(d.RedocOptions),
		}
		addHeaderAndScripts(data, redocOpts)
	}
	return
}
