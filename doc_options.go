package chioas

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"net/http"
)

// DocOptions determines whether/how the api will serve interactive docs page
type DocOptions struct {
	// ServeDocs whether to serve api docs
	ServeDocs bool
	// Path the path on which to serve api docs page and spec (defaults to "/docs")
	Path string
	// DocIndexPage the name of the docs index page (defaults to "index.htm")
	DocIndexPage string
	// Title the title in the docs index page (defaults to "API Documentation")
	Title string
	// RedocOptions redoc options to be used
	RedocOptions map[string]any
	// SpecName the name of the OAS spec (defaults to "spec.yaml")
	SpecName string
	// DocTemplate template for the docs page (uses internal template if nil)
	DocTemplate string
	// StylesOverride css styling overrides (injected into docs index page)
	StylesOverride string
}

func (d *DocOptions) setupRoutes(def *Definition, route chi.Router) {
	if d.ServeDocs {
		tmpStr := defValue(d.DocTemplate, defaultTemplate)
		tmp := template.Must(template.New("index").Parse(tmpStr))
		path := defValue(d.Path, "/docs")
		indexPage := defValue(d.DocIndexPage, "index.htm")
		specName := defValue(d.SpecName, "spec.yaml")
		data := map[string]any{
			"title":          defValue(d.Title, "API Documentation"),
			"stylesOverride": template.CSS(defValue(d.StylesOverride, defaultStylesOverride)),
			"specName":       specName,
			"redocopts":      d.RedocOptions,
		}
		redirectPath := path + "/" + indexPage
		route.Route(path, func(r chi.Router) {
			r.Get("/", http.RedirectHandler(redirectPath, http.StatusMovedPermanently).ServeHTTP)
			r.Get("/"+indexPage, func(writer http.ResponseWriter, request *http.Request) {
				_ = tmp.Execute(writer, data)
			})
			r.Get("/"+specName, func(writer http.ResponseWriter, request *http.Request) {
				_ = def.WriteYaml(writer)
			})
		})
	}
}

const defaultTemplate = `<html lang="en">
    <head>
        <title>{{.title}}</title>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
        <style>{{.stylesOverride}}</style>
    </head>
    <body>
        <div id="redoc-container"></div>
        <script src="https://cdn.jsdelivr.net/npm/redoc@2.0.0-rc.55/bundles/redoc.standalone.min.js" integrity="sha256-JMQl1CS8zo2BXvovJJTJ0arp2sSpWYHT1ex/x9V2MVc= sha384-MGNhtF9OCmViLJZUa8VTjfh9rQntvAEtDw/8mWiqdTDRl/or6C8/UaCMTw6RVpai sha512-a0TxgNGlrfEEIsYkUi4bPES53uI+t/hnGkBDRKEdPT9p5R8tKTnY5nuMLLkC7FYn78Ha/40H3Pf8NQ9qtT8vUg==" crossorigin="anonymous"></script>
        <script src="https://cdn.jsdelivr.net/gh/wll8/redoc-try@1.4.0/dist/try.js" integrity="sha256-HfA6qJrDb3S0Vl7Jmf7vyKtuwpZlYt9k5nnP9dEHqok= sha384-tPALRNTmr7jwIcQsUqoLIqTxTiBM08yLWuEzqZT5QW0O3okj1exfVcqeFHIXcWjJ sha512-NW9V3GQ+LBvpT6ygUD/YtjdKWbafrCyYysVcbFHW6BusHw2Wwgr460z8XZiR1wmfLEJ42kFXiZ4o1jC0HHWdjQ==" crossorigin="anonymous"></script>
        <script>
            initTry({
                openApi: {{.specName}},
                redocOptions: {{.redocopts}}
            })
        </script>
    </body>
</html>`
const defaultStylesOverride = `body {
	margin: 0;
	padding: 0;
}
/* restyle Try button */
button.tryBtn {
	margin-right: 0.5em;
	padding: 4px 8px 4px 8px;
	border-radius: 4px;
	text-transform: capitalize;
}
/* restyle Try copy to clipboard button */
div.copy-to-clipboard {
	text-align: right;
	margin-top: -1em;
}
div.copy-to-clipboard button {
	min-height: 1em;
}
/* try responses area - make scrollable */
div.responses-inner > div > div > table.live-responses-table {
	table-layout: fixed;
}
div.responses-inner > div > div > table.live-responses-table td.response-col_description {
	width: 80%;
}
body #redoc-container {
	float: left;
	width: 100%;
}
`

func defValue(v, def string) string {
	if v != "" {
		return v
	}
	return def
}
