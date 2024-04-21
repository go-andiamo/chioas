package chioas

// RapidocOptions describes the rapidoc-ui options
//
// for documentation see https://github.com/rapi-doc/RapiDoc
type RapidocOptions struct {
	ShowHeader             bool   // show-header
	HeadingText            string // heading-text
	Theme                  string // theme="light"
	RenderStyle            string // render-style="view"
	SchemaStyle            string // schema-style="table"
	ShowMethodInNavBar     bool   // show-method-in-nav-bar="true"
	UsePathInNavBar        bool   // use-path-in-nav-bar="true"
	ShowComponents         bool   // show-components="true"
	HideInfo               bool   // !show-info="true"
	AllowSearch            bool   // allow-search="false"
	AllowAdvancedSearch    bool   // allow-advanced-search="false"
	AllowSpecUrlLoad       bool   // allow-spec-url-load="false"
	AllowSpecFileLoad      bool   // allow-spec-file-load="false"
	DisallowTry            bool   // !allow-try="true"
	DisallowSpecDownload   bool   // !allow-spec-file-download="true"
	AllowServerSelection   bool   // allow-server-selection="false"
	DisallowAuthentication bool   // !allow-authentication="true"
	NoPersistAuth          bool   // !persist-auth="true"
	UpdateRoute            bool   // update-route="true"
	MatchType              string // match-type="regex"
}

func (o *RapidocOptions) ToMap() map[string]any {
	m := map[string]any{}
	addMapAlways(m, o.ShowHeader, "show_header")
	addMapAlways(m, o.HeadingText, "heading_text")
	addMapOrDef(m, o.Theme, "theme", "light")
	addMapOrDef(m, o.RenderStyle, "render_style", "view")
	addMapOrDef(m, o.SchemaStyle, "schema_style", "table")
	addMapAlways(m, o.ShowMethodInNavBar, "show_method_in_nav_bar")
	addMapAlways(m, o.UsePathInNavBar, "use_path_in_nav_bar")
	addMapAlways(m, o.ShowComponents, "show_components")
	addMapAlways(m, !o.HideInfo, "show_info")
	addMapAlways(m, o.AllowSearch, "allow_search")
	addMapAlways(m, o.AllowAdvancedSearch, "allow_advanced_search")
	addMapAlways(m, o.AllowSpecUrlLoad, "allow_spec_url_load")
	addMapAlways(m, o.AllowSpecFileLoad, "allow_spec_file_load")
	addMapAlways(m, !o.DisallowTry, "allow_try")
	addMapAlways(m, !o.DisallowSpecDownload, "allow_spec_file_download")
	addMapAlways(m, o.AllowServerSelection, "allow_server_selection")
	addMapAlways(m, !o.DisallowAuthentication, "allow_authentication")
	addMapAlways(m, !o.NoPersistAuth, "persist_auth")
	addMapAlways(m, o.UpdateRoute, "update_route")
	addMapOrDef(m, o.MatchType, "match_type", "regex")
	return m
}
