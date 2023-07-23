package chioas

// RedocOptions for use in DocOptions.RedocOptions (from https://github.com/Redocly/redoc#redoc-options-object)
type RedocOptions struct {
	DisableSearch                   bool        `json:"disableSearch,omitempty"`                   // disable search indexing and search box.
	MinCharacterLengthToInitSearch  int         `json:"minCharacterLengthToInitSearch,omitempty"`  // set minimal characters length to init search, default 3, minimal 1.
	ExpandDefaultServerVariables    bool        `json:"expandDefaultServerVariables,omitempty"`    // enable expanding default server variables, default false.
	ExpandResponses                 string      `json:"expandResponses,omitempty"`                 // specify which responses to expand by default by response codes. Values should be passed as comma-separated list without spaces e.g. expandResponses="200,201". Special value "all" expands all responses by default. Be careful: this option can slow-down documentation rendering time.
	GeneratedPayloadSamplesMaxDepth int         `json:"generatedPayloadSamplesMaxDepth,omitempty"` // set the maximum render depth for JSON payload samples (responses and request body). The default value is 10.
	MaxDisplayedEnumValues          int         `json:"maxDisplayedEnumValues,omitempty"`          // display only specified number of enum values. hide rest values under spoiler.
	HideDownloadButton              bool        `json:"hideDownloadButton,omitempty"`              // do not show "Download" spec button. THIS DOESN'T MAKE YOUR SPEC PRIVATE, it just hides the button.
	DownloadFileName                string      `json:"downloadFileName,omitempty"`                // set a custom file name for the downloaded API definition file.
	DownloadDefinitionUrl           string      `json:"downloadDefinitionUrl,omitempty"`           // If the 'Download' button is visible in the API reference documentation (hideDownloadButton=false), the URL configured here opens when that button is selected. Provide it as an absolute URL with the full URI scheme.
	HideHostname                    bool        `json:"hideHostname,omitempty"`                    // if set, the protocol and hostname is not shown in the operation definition.
	HideLoading                     bool        `json:"hideLoading,omitempty"`                     // do not show loading animation. Useful for small docs.
	HideFab                         bool        `json:"hideFab,omitempty"`                         // do not show FAB in mobile view. Useful for implementing a custom floating action button.
	HideSchemaPattern               bool        `json:"hideSchemaPattern,omitempty"`               // if set, the pattern is not shown in the schema.
	HideSingleRequestSampleTab      bool        `json:"hideSingleRequestSampleTab,omitempty"`      // do not show the request sample tab for requests with only one sample.
	ShowObjectSchemaExamples        bool        `json:"showObjectSchemaExamples,omitempty"`        // show object schema example in the properties, default false.
	ExpandSingleSchemaField         bool        `json:"expandSingleSchemaField,omitempty"`         // automatically expand single field in a schema
	SchemaExpansionLevel            any         `json:"schemaExpansionLevel,omitempty"`            // specifies whether to automatically expand schemas. Special value "all" expands all levels. The default value is 0.
	JsonSampleExpandLevel           any         `json:"jsonSampleExpandLevel,omitempty"`           // set the default expand level for JSON payload samples (responses and request body). Special value "all" expands all levels. The default value is 2.
	HideSchemaTitles                bool        `json:"hideSchemaTitles,omitempty"`                // do not display schema title next to the type
	SimpleOneOfTypeLabel            bool        `json:"simpleOneOfTypeLabel,omitempty"`            // show only unique oneOf types in the label without titles
	SortEnumValuesAlphabetically    bool        `json:"sortEnumValuesAlphabetically,omitempty"`    // set to true, sorts all enum values in all schemas alphabetically
	SortOperationsAlphabetically    bool        `json:"sortOperationsAlphabetically,omitempty"`    // set to true, sorts operations in the navigation sidebar and in the middle panel alphabetically
	SortTagsAlphabetically          bool        `json:"sortTagsAlphabetically,omitempty"`          // set to true, sorts tags in the navigation sidebar and in the middle panel alphabetically
	MenuToggle                      bool        `json:"menuToggle,omitempty"`                      // if true, clicking second time on expanded menu item collapses it, default true.
	NativeScrollbars                bool        `json:"nativeScrollbars,omitempty"`                // use native scrollbar for sidemenu instead of perfect-scroll (scrolling performance optimization for big specs).
	OnlyRequiredInSamples           bool        `json:"onlyRequiredInSamples,omitempty"`           // shows only required fields in request samples.
	PathInMiddlePanel               bool        `json:"pathInMiddlePanel,omitempty"`               // show path link and HTTP verb in the middle panel instead of the right one.
	RequiredPropsFirst              bool        `json:"requiredPropsFirst,omitempty"`              // show required properties first ordered in the same order as in required array.
	ScrollYOffset                   any         `json:"scrollYOffset,omitempty"`                   // If set, specifies a vertical scroll-offset. This is often useful when there are fixed positioned elements at the top of the page, such as navbars, headers etc; scrollYOffset can be specified in various ways:
	ShowExtensions                  bool        `json:"showExtensions,omitempty"`                  // show vendor extensions ("x-" fields). Extensions used by Redoc are ignored. Can be boolean or an array of string with names of extensions to display.
	SortPropsAlphabetically         bool        `json:"sortPropsAlphabetically,omitempty"`         // sort properties alphabetically.
	PayloadSampleIndex              int         `json:"payloadSampleIdx,omitempty"`                // if set, payload sample is inserted at this index or last. Indexes start from 0.
	Theme                           *RedocTheme `json:"theme,omitempty"`                           // Redoc theme
	UntrustedSpec                   bool        `json:"untrustedSpec,omitempty"`                   // if set, the spec is considered untrusted and all HTML/markdown is sanitized to prevent XSS. Disabled by default for performance reasons. Enable this option if you work with untrusted user data!
	Nonce                           any         `json:"nonce,omitempty"`                           // if set, the provided value is injected in every injected HTML element in the nonce attribute. Useful when using CSP, see https://webpack.js.org/guides/csp/.
	SideNavStyle                    string      `json:"sideNavStyle,omitempty"`                    // can be specified in various ways: "summary-only" (default), "path-only" or id-only"
	ShowWebhookVerb                 bool        `json:"showWebhookVerb,omitempty"`                 // when set to true, shows the HTTP request method for webhooks in operations and in the sidebar.
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

type RedocSpacing struct {
	Unit              int `json:"unit,omitempty"`              // 5 # main spacing unit used in autocomputed theme values later
	SectionHorizontal int `json:"sectionHorizontal,omitempty"` // 40 # Horizontal section padding. COMPUTED: spacing.unit * 8
	SectionVertical   int `json:"sectionVertical,omitempty"`   // 40 # Horizontal section padding. COMPUTED: spacing.unit * 8
}

// RedocBreakpoints breakpoints for switching three/two and mobile view layouts
type RedocBreakpoints struct {
	Small  string `json:"small,omitempty"`  // '50rem'
	Medium string `json:"medium,omitempty"` // '85rem'
	Large  string `json:"large"`            // '105rem'
}

type RedocColors struct {
	TonalOffset float32 `json:"tonalOffset,omitempty"` // 0.3 # default tonal offset used in computations
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

type RedocHeadings struct {
	FontFamily string `json:"fontFamily,omitempty"` // 'Montserrat, sans-serif'
	FontWeight string `json:"fontWeight,omitempty"` // '400'
	LineHeight string `json:"lineHeight,omitempty"` // '1.6em'
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

type RedocLinks struct {
	Color               string `json:"color,omitempty"`          // COMPUTED: colors.primary.main
	Visited             string `json:"visited,omitempty"`        // COMPUTED: typography.links.color
	Hover               string `json:"hover,omitempty"`          // COMPUTED: lighten(0.2 typography.links.color)
	TextDecoration      string `json:"textDecoration,omitempty"` // 'auto'
	HoverTextDecoration string `json:"hoverTextDecoration"`      // 'auto'
}

type RedocSidebar struct {
	Width           string            `json:"width,omitempty"`           // '260px'
	BackgroundColor string            `json:"backgroundColor,omitempty"` // '#fafafa'
	TextColor       string            `json:"textColor,omitempty"`       // '#333333'
	ActiveTextColor string            `json:"activeTextColor"`           // COMPUTED: theme.sidebar.textColor (if set by user) or theme.colors.primary.main
	GroupItems      *RedocGroupItems  `json:"groupItems,omitempty"`      // Group headings
	Level1Items     *RedocLevel1Items `json:"level1Items,omitempty"`     //  Level 1 items like tags or section 1st level items
	Arrow           *RedocArrow       `json:"arrow,omitempty"`           // sidebar arrow
}

type RedocGroupItems struct {
	ActiveBackgroundColor string `json:"activeBackgroundColor,omitempty"` // COMPUTED: theme.sidebar.backgroundColor
	ActiveTextColor       string `json:"activeTextColor"`                 // COMPUTED: theme.sidebar.activeTextColor
	TextTransform         string `json:"textTransform"`                   // 'uppercase'
}

type RedocLevel1Items struct {
	ActiveBackgroundColor string `json:"activeBackgroundColor,omitempty"` // COMPUTED: theme.sidebar.backgroundColor
	ActiveTextColor       string `json:"activeTextColor,omitempty"`       // COMPUTED: theme.sidebar.activeTextColor
	TextTransform         string `json:"textTransform"`                   // 'none'
}

type RedocArrow struct {
	Size  string `json:"size,omitempty"` // '1.5em'
	Color string `json:"color"`          // COMPUTED: theme.sidebar.textColor
}

type RedocLogo struct {
	MaxHeight int    `json:"maxHeight,omitempty"` // COMPUTED: sidebar.width
	MaxWidth  int    `json:"maxWidth,omitempty"`  // COMPUTED: sidebar.width
	Gutter    string `json:"gutter"`              // '2px' # logo image padding
}

type RedocRightPanel struct {
}

type RedocFab struct {
	BackgroundColor string `json:"backgroundColor,omitempty"` // '#263238'
	Color           string `json:"color,omitempty"`           // '#ffffff'
}
