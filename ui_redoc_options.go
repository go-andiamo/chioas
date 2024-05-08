package chioas

import "html/template"

const defaultRedocTemplate = `<html>
    <head>
        <title>{{.title}}</title>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
        {{.favIcons}}
        <style>{{.stylesOverride}}</style>
        {{.headScript}}
    </head>
    <body>
        {{.headerHtml}}
        <div id="redoc-container"></div>
        <script src="{{.redocurl}}" crossorigin="anonymous"></script>
        <script src="{{.tryurl}}" crossorigin="anonymous"></script>
        <script>
            initTry({
                openApi: {{.specName}},
                redocOptions: {{.redocopts}}
            })
        </script>
        {{.bodyScript}}
    </body>
</html>`

const defaultRedocStylesOverride = `body {
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

// RedocOptions for use in DocOptions.RedocOptions (from https://github.com/Redocly/redoc#redoc-options-object)
type RedocOptions struct {
	DisableSearch                   bool          `json:"disableSearch,omitempty"`                   // disable search indexing and search box.
	MinCharacterLengthToInitSearch  int           `json:"minCharacterLengthToInitSearch,omitempty"`  // set minimal characters length to init search, default 3, minimal 1.
	ExpandDefaultServerVariables    bool          `json:"expandDefaultServerVariables,omitempty"`    // enable expanding default server variables, default false.
	ExpandResponses                 string        `json:"expandResponses,omitempty"`                 // specify which responses to expand by default by response codes. Values should be passed as comma-separated list without spaces e.g. expandResponses="200,201". Special value "all" expands all responses by default. Be careful: this option can slow-down documentation rendering time.
	GeneratedPayloadSamplesMaxDepth int           `json:"generatedPayloadSamplesMaxDepth,omitempty"` // set the maximum render depth for JSON payload samples (responses and request body). The default value is 10.
	MaxDisplayedEnumValues          int           `json:"maxDisplayedEnumValues,omitempty"`          // display only specified number of enum values. hide rest values under spoiler.
	HideDownloadButton              bool          `json:"hideDownloadButton,omitempty"`              // do not show "Download" spec button. THIS DOESN'T MAKE YOUR SPEC PRIVATE, it just hides the button.
	DownloadFileName                string        `json:"downloadFileName,omitempty"`                // set a custom file name for the downloaded API definition file.
	DownloadDefinitionUrl           string        `json:"downloadDefinitionUrl,omitempty"`           // If the 'Download' button is visible in the API reference documentation (hideDownloadButton=false), the URL configured here opens when that button is selected. Provide it as an absolute URL with the full URI scheme.
	HideHostname                    bool          `json:"hideHostname,omitempty"`                    // if set, the protocol and hostname is not shown in the operation definition.
	HideLoading                     bool          `json:"hideLoading,omitempty"`                     // do not show loading animation. Useful for small docs.
	HideFab                         bool          `json:"hideFab,omitempty"`                         // do not show FAB in mobile view. Useful for implementing a custom floating action button.
	HideSchemaPattern               bool          `json:"hideSchemaPattern,omitempty"`               // if set, the pattern is not shown in the schema.
	HideSingleRequestSampleTab      bool          `json:"hideSingleRequestSampleTab,omitempty"`      // do not show the request sample tab for requests with only one sample.
	ShowObjectSchemaExamples        bool          `json:"showObjectSchemaExamples,omitempty"`        // show object schema example in the properties, default false.
	ExpandSingleSchemaField         bool          `json:"expandSingleSchemaField,omitempty"`         // automatically expand single field in a schema
	SchemaExpansionLevel            any           `json:"schemaExpansionLevel,omitempty"`            // specifies whether to automatically expand schemas. Special value "all" expands all levels. The default value is 0.
	JsonSampleExpandLevel           any           `json:"jsonSampleExpandLevel,omitempty"`           // set the default expand level for JSON payload samples (responses and request body). Special value "all" expands all levels. The default value is 2.
	HideSchemaTitles                bool          `json:"hideSchemaTitles,omitempty"`                // do not display schema title next to the type
	SimpleOneOfTypeLabel            bool          `json:"simpleOneOfTypeLabel,omitempty"`            // show only unique oneOf types in the label without titles
	SortEnumValuesAlphabetically    bool          `json:"sortEnumValuesAlphabetically,omitempty"`    // set to true, sorts all enum values in all schemas alphabetically
	SortOperationsAlphabetically    bool          `json:"sortOperationsAlphabetically,omitempty"`    // set to true, sorts operations in the navigation sidebar and in the middle panel alphabetically
	SortTagsAlphabetically          bool          `json:"sortTagsAlphabetically,omitempty"`          // set to true, sorts tags in the navigation sidebar and in the middle panel alphabetically
	MenuToggle                      bool          `json:"menuToggle,omitempty"`                      // if true, clicking second time on expanded menu item collapses it, default true.
	NativeScrollbars                bool          `json:"nativeScrollbars,omitempty"`                // use native scrollbar for sidemenu instead of perfect-scroll (scrolling performance optimization for big specs).
	OnlyRequiredInSamples           bool          `json:"onlyRequiredInSamples,omitempty"`           // shows only required fields in request samples.
	PathInMiddlePanel               bool          `json:"pathInMiddlePanel,omitempty"`               // show path link and HTTP verb in the middle panel instead of the right one.
	RequiredPropsFirst              bool          `json:"requiredPropsFirst,omitempty"`              // show required properties first ordered in the same order as in required array.
	ScrollYOffset                   any           `json:"scrollYOffset,omitempty"`                   // If set, specifies a vertical scroll-offset. This is often useful when there are fixed positioned elements at the top of the page, such as navbars, headers etc; scrollYOffset can be specified in various ways:
	ShowExtensions                  bool          `json:"showExtensions,omitempty"`                  // show vendor extensions ("x-" fields). Extensions used by Redoc are ignored. Can be boolean or an array of string with names of extensions to display.
	SortPropsAlphabetically         bool          `json:"sortPropsAlphabetically,omitempty"`         // sort properties alphabetically.
	PayloadSampleIndex              int           `json:"payloadSampleIdx,omitempty"`                // if set, payload sample is inserted at this index or last. Indexes start from 0.
	Theme                           *RedocTheme   `json:"theme,omitempty"`                           // Redoc theme
	UntrustedSpec                   bool          `json:"untrustedSpec,omitempty"`                   // if set, the spec is considered untrusted and all HTML/markdown is sanitized to prevent XSS. Disabled by default for performance reasons. Enable this option if you work with untrusted user data!
	Nonce                           any           `json:"nonce,omitempty"`                           // if set, the provided value is injected in every injected HTML element in the nonce attribute. Useful when using CSP, see https://webpack.js.org/guides/csp/.
	SideNavStyle                    string        `json:"sideNavStyle,omitempty"`                    // can be specified in various ways: "summary-only" (default), "path-only" or id-only"
	ShowWebhookVerb                 bool          `json:"showWebhookVerb,omitempty"`                 // when set to true, shows the HTTP request method for webhooks in operations and in the sidebar.
	FavIcons                        FavIcons      `json:"-"`
	HeaderHtml                      template.HTML `json:"-"`
	HeadScript                      template.JS   `json:"-"`
	BodyScript                      template.JS   `json:"-"`
}

func (o RedocOptions) GetFavIcons() FavIcons {
	return o.FavIcons
}

func (o RedocOptions) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.DisableSearch, "disableSearch")
	addMap(m, o.MinCharacterLengthToInitSearch, "minCharacterLengthToInitSearch")
	addMap(m, o.ExpandDefaultServerVariables, "expandDefaultServerVariables")
	addMap(m, o.ExpandResponses, "expandResponses")
	addMap(m, o.GeneratedPayloadSamplesMaxDepth, "generatedPayloadSamplesMaxDepth")
	addMap(m, o.MaxDisplayedEnumValues, "maxDisplayedEnumValues")
	addMap(m, o.HideDownloadButton, "hideDownloadButton")
	addMap(m, o.DownloadFileName, "downloadFileName")
	addMap(m, o.DownloadDefinitionUrl, "downloadDefinitionUrl")
	addMap(m, o.HideHostname, "hideHostname")
	addMap(m, o.HideLoading, "hideLoading")
	addMap(m, o.HideFab, "hideFab")
	addMap(m, o.HideSchemaPattern, "hideSchemaPattern")
	addMap(m, o.HideSingleRequestSampleTab, "hideSingleRequestSampleTab")
	addMap(m, o.ShowObjectSchemaExamples, "showObjectSchemaExamples")
	addMap(m, o.ExpandSingleSchemaField, "expandSingleSchemaField")
	addMapNonNil(m, o.SchemaExpansionLevel, "schemaExpansionLevel")
	addMapNonNil(m, o.JsonSampleExpandLevel, "jsonSampleExpandLevel")
	addMap(m, o.HideSchemaTitles, "hideSchemaTitles")
	addMap(m, o.SimpleOneOfTypeLabel, "simpleOneOfTypeLabel")
	addMap(m, o.SortEnumValuesAlphabetically, "sortEnumValuesAlphabetically")
	addMap(m, o.SortOperationsAlphabetically, "sortOperationsAlphabetically")
	addMap(m, o.SortTagsAlphabetically, "sortTagsAlphabetically")
	addMap(m, o.MenuToggle, "menuToggle")
	addMap(m, o.NativeScrollbars, "nativeScrollbars")
	addMap(m, o.OnlyRequiredInSamples, "onlyRequiredInSamples")
	addMap(m, o.PathInMiddlePanel, "pathInMiddlePanel")
	addMap(m, o.RequiredPropsFirst, "requiredPropsFirst")
	addMapNonNil(m, o.ScrollYOffset, "scrollYOffset")
	addMap(m, o.ShowExtensions, "showExtensions")
	addMap(m, o.SortPropsAlphabetically, "sortPropsAlphabetically")
	addMap(m, o.PayloadSampleIndex, "payloadSampleIdx")
	addMapMappable(m, o.Theme, "theme")
	addMap(m, o.UntrustedSpec, "untrustedSpec")
	addMapNonNil(m, o.Nonce, "nonce")
	addMap(m, o.SideNavStyle, "sideNavStyle")
	addMap(m, o.ShowWebhookVerb, "showWebhookVerb")
	if o.HeaderHtml != "" {
		m[htmlTagHeaderHtml] = o.HeaderHtml
	}
	if o.HeadScript != "" {
		m[htmlTagHeadScript] = template.HTML(`<script>` + o.HeadScript + `</script>`)
	}
	if o.BodyScript != "" {
		m[htmlTagBodyScript] = template.HTML(`<script>` + o.BodyScript + `</script>`)
	}
	return m
}

type RedocTheme struct {
	Spacing     *RedocSpacing     `json:"spacing,omitempty"`
	Breakpoints *RedocBreakpoints `json:"breakpoints,omitempty"`
	Colors      *RedocColors      `json:"colors,omitempty"`
	Typography  *RedocTypography  `json:"typography,omitempty"`
	Sidebar     *RedocSidebar     `json:"sidebar,omitempty"`
	Logo        *RedocLogo        `json:"logo,omitempty"`
	RightPanel  *RedocRightPanel  `json:"rightPanel,omitempty"`
	Fab         *RedocFab         `json:"fab,omitempty"`
}

func (o *RedocTheme) ToMap() map[string]any {
	m := map[string]any{}
	addMapMappable(m, o.Spacing, "spacing")
	addMapMappable(m, o.Breakpoints, "breakpoints")
	addMapMappable(m, o.Colors, "colors")
	addMapMappable(m, o.Typography, "typography")
	addMapMappable(m, o.Sidebar, "sidebar")
	addMapMappable(m, o.Logo, "logo")
	addMapMappable(m, o.RightPanel, "rightPanel")
	addMapMappable(m, o.Fab, "fab")
	return m
}

type RedocSpacing struct {
	Unit              int `json:"unit,omitempty"`              // 5 # main spacing unit used in autocomputed theme values later
	SectionHorizontal int `json:"sectionHorizontal,omitempty"` // 40 # Horizontal section padding. COMPUTED: spacing.unit * 8
	SectionVertical   int `json:"sectionVertical,omitempty"`   // 40 # Horizontal section padding. COMPUTED: spacing.unit * 8
}

func (o *RedocSpacing) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.Unit, "unit")
	addMap(m, o.SectionHorizontal, "sectionHorizontal")
	addMap(m, o.SectionVertical, "sectionVertical")
	return m
}

// RedocBreakpoints breakpoints for switching three/two and mobile view layouts
type RedocBreakpoints struct {
	Small  string `json:"small,omitempty"`  // '50rem'
	Medium string `json:"medium,omitempty"` // '85rem'
	Large  string `json:"large,omitempty"`  // '105rem'
}

func (o *RedocBreakpoints) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.Small, "small")
	addMap(m, o.Medium, "medium")
	addMap(m, o.Large, "large")
	return m
}

type RedocColors struct {
	TonalOffset float32 `json:"tonalOffset,omitempty"` // 0.3 # default tonal offset used in computations
}

func (o *RedocColors) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.TonalOffset, "tonalOffset")
	return m
}

type RedocTypography struct {
	FontSize          string         `json:"fontSize,omitempty"`          // '14px'
	LineHeight        string         `json:"lineHeight,omitempty"`        // '1.5em'
	FontWeightRegular string         `json:"fontWeightRegular,omitempty"` // '400'
	FontWeightBold    string         `json:"fontWeightBold,omitempty"`    // '600'
	FontWeightLight   string         `json:"fontWeightLight,omitempty"`   // '300'
	FontFamily        string         `json:"fontFamily,omitempty"`        // 'Roboto, sans-serif'
	Smoothing         string         `json:"smoothing,omitempty"`         // 'antialiased'
	OptimizeSpeed     bool           `json:"optimizeSpeed"`               // true
	Headings          *RedocHeadings `json:"headings,omitempty"`
	Code              *RedocCode     `json:"code,omitempty"`
	Links             *RedocLinks    `json:"links,omitempty"`
}

func (o RedocTypography) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.FontSize, "fontSize")
	addMap(m, o.LineHeight, "lineHeight")
	addMap(m, o.FontWeightRegular, "fontWeightRegular")
	addMap(m, o.FontWeightBold, "fontWeightBold")
	addMap(m, o.FontWeightLight, "fontWeightLight")
	addMap(m, o.FontFamily, "fontFamily")
	addMap(m, o.Smoothing, "smoothing")
	addMapDef(m, o.OptimizeSpeed, "optimizeSpeed", true)
	addMapMappable(m, o.Headings, "headings")
	addMapMappable(m, o.Code, "code")
	addMapMappable(m, o.Links, "links")
	return m
}

type RedocHeadings struct {
	FontFamily string `json:"fontFamily,omitempty"` // 'Montserrat, sans-serif'
	FontWeight string `json:"fontWeight,omitempty"` // '400'
	LineHeight string `json:"lineHeight,omitempty"` // '1.6em'
}

func (o *RedocHeadings) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.FontFamily, "fontFamily")
	addMap(m, o.FontWeight, "fontWeight")
	addMap(m, o.LineHeight, "lineHeight")
	return m
}

type RedocCode struct {
	FontSize        string `json:"fontSize,omitempty"`        // '13px'
	FontFamily      string `json:"fontFamily,omitempty"`      // 'Courier, monospace'
	LineHeight      string `json:"lineHeight,omitempty"`      // COMPUTED: typography.lineHeight
	FontWeight      string `json:"fontWeight,omitempty"`      // COMPUTED: typography.fontWeightRegular
	Color           string `json:"color,omitempty"`           // '#e53935'
	BackgroundColor string `json:"backgroundColor,omitempty"` // 'rgba(38, 50, 56, 0.05)'
	Wrap            bool   `json:"wrap,omitempty"`            // whether to break word for inline blocks (otherwise they can overflow)
}

func (o *RedocCode) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.FontFamily, "fontFamily")
	addMap(m, o.FontWeight, "fontWeight")
	addMap(m, o.LineHeight, "lineHeight")
	addMap(m, o.FontWeight, "fontWeight")
	addMap(m, o.Color, "color")
	addMap(m, o.BackgroundColor, "backgroundColor")
	addMap(m, o.Wrap, "wrap")
	return m
}

type RedocLinks struct {
	Color               string `json:"color,omitempty"`               // COMPUTED: colors.primary.main
	Visited             string `json:"visited,omitempty"`             // COMPUTED: typography.links.color
	Hover               string `json:"hover,omitempty"`               // COMPUTED: lighten(0.2 typography.links.color)
	TextDecoration      string `json:"textDecoration,omitempty"`      // 'auto'
	HoverTextDecoration string `json:"hoverTextDecoration,omitempty"` // 'auto'
}

func (o *RedocLinks) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.Color, "color")
	addMap(m, o.Visited, "visited")
	addMap(m, o.Hover, "hover")
	addMap(m, o.TextDecoration, "textDecoration")
	addMap(m, o.HoverTextDecoration, "hoverTextDecoration")
	return m
}

type RedocSidebar struct {
	Width           string            `json:"width,omitempty"`           // '260px'
	BackgroundColor string            `json:"backgroundColor,omitempty"` // '#fafafa'
	TextColor       string            `json:"textColor,omitempty"`       // '#333333'
	ActiveTextColor string            `json:"activeTextColor,omitempty"` // COMPUTED: theme.sidebar.textColor (if set by user) or theme.colors.primary.main
	GroupItems      *RedocGroupItems  `json:"groupItems,omitempty"`      // Group headings
	Level1Items     *RedocLevel1Items `json:"level1Items,omitempty"`     //  Level 1 items like tags or section 1st level items
	Arrow           *RedocArrow       `json:"arrow,omitempty"`           // sidebar arrow
}

func (o *RedocSidebar) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.Width, "width")
	addMap(m, o.BackgroundColor, "backgroundColor")
	addMap(m, o.TextColor, "textColor")
	addMap(m, o.ActiveTextColor, "activeTextColor")
	addMapMappable(m, o.GroupItems, "groupItems")
	addMapMappable(m, o.Level1Items, "level1Items")
	addMapMappable(m, o.Arrow, "arrow")
	return m
}

type RedocGroupItems struct {
	ActiveBackgroundColor string `json:"activeBackgroundColor,omitempty"` // COMPUTED: theme.sidebar.backgroundColor
	ActiveTextColor       string `json:"activeTextColor,omitempty"`       // COMPUTED: theme.sidebar.activeTextColor
	TextTransform         string `json:"textTransform,omitempty"`         // 'uppercase'
}

func (o *RedocGroupItems) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.ActiveBackgroundColor, "activeBackgroundColor")
	addMap(m, o.ActiveTextColor, "activeTextColor")
	addMap(m, o.TextTransform, "textTransform")
	return m
}

type RedocLevel1Items struct {
	ActiveBackgroundColor string `json:"activeBackgroundColor,omitempty"` // COMPUTED: theme.sidebar.backgroundColor
	ActiveTextColor       string `json:"activeTextColor,omitempty"`       // COMPUTED: theme.sidebar.activeTextColor
	TextTransform         string `json:"textTransform,omitempty"`         // 'none'
}

func (o *RedocLevel1Items) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.ActiveBackgroundColor, "activeBackgroundColor")
	addMap(m, o.ActiveTextColor, "activeTextColor")
	addMap(m, o.TextTransform, "textTransform")
	return m
}

type RedocArrow struct {
	Size  string `json:"size,omitempty"`  // '1.5em'
	Color string `json:"color,omitempty"` // COMPUTED: theme.sidebar.textColor
}

func (o *RedocArrow) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.Size, "size")
	addMap(m, o.Color, "color")
	return m
}

type RedocLogo struct {
	MaxHeight int    `json:"maxHeight,omitempty"` // COMPUTED: sidebar.width
	MaxWidth  int    `json:"maxWidth,omitempty"`  // COMPUTED: sidebar.width
	Gutter    string `json:"gutter,omitempty"`    // '2px' # logo image padding
}

func (o *RedocLogo) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.MaxHeight, "maxHeight")
	addMap(m, o.MaxWidth, "maxWidth")
	addMap(m, o.Gutter, "gutter")
	return m
}

type RedocRightPanel struct {
	BackgroundColor string                  `json:"backgroundColor,omitempty"` // '#263238'
	Width           string                  `json:"width,omitempty"`           // '40%'
	TextColor       string                  `json:"textColor,omitempty"`       // '#ffffff'
	Servers         *RedocRightPanelServers `json:"servers,omitempty"`
}

func (o *RedocRightPanel) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.BackgroundColor, "backgroundColor")
	addMap(m, o.Width, "width")
	addMap(m, o.TextColor, "textColor")
	addMapMappable(m, o.Servers, "servers")
	return m
}

type RedocRightPanelServers struct {
	Overlay *RedocRightPanelServersOverlay `json:"overlay,omitempty"`
	Url     *RedocRightPanelServersUrl     `json:"url,omitempty"`
}

func (o *RedocRightPanelServers) ToMap() map[string]any {
	m := map[string]any{}
	addMapMappable(m, o.Overlay, "overlay")
	addMapMappable(m, o.Url, "url")
	return m
}

type RedocRightPanelServersOverlay struct {
	BackgroundColor string `json:"backgroundColor,omitempty"` // '#fafafa'
	TextColor       string `json:"textColor,omitempty"`       // '#263238'
}

func (o *RedocRightPanelServersOverlay) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.BackgroundColor, "backgroundColor")
	addMap(m, o.TextColor, "textColor")
	return m
}

type RedocRightPanelServersUrl struct {
	BackgroundColor string `json:"backgroundColor,omitempty"` // '#fff'
}

func (o *RedocRightPanelServersUrl) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.BackgroundColor, "backgroundColor")
	return m
}

type RedocFab struct {
	BackgroundColor string `json:"backgroundColor,omitempty"` // '#263238'
	Color           string `json:"color,omitempty"`           // '#ffffff'
}

func (o *RedocFab) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.BackgroundColor, "backgroundColor")
	addMap(m, o.Color, "color")
	return m
}
