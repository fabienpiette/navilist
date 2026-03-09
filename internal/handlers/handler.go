package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/user/navilist/pkg/navidrome"
)

// Templates routes ExecuteTemplate calls to the correct per-page template set.
// Each page is parsed into its own cloned set so {{define "content"}} blocks
// don't overwrite each other across pages.
type Templates struct {
	sets map[string]*template.Template
}

// NewTemplates creates a Templates dispatcher from a map of name → template set.
func NewTemplates(sets map[string]*template.Template) *Templates {
	return &Templates{sets: sets}
}

// ExecuteTemplate finds the set registered under name and executes it.
func (ts *Templates) ExecuteTemplate(w io.Writer, name string, data any) error {
	t, ok := ts.sets[name]
	if !ok {
		return fmt.Errorf("no template registered for %q", name)
	}
	return t.ExecuteTemplate(w, name, data)
}

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	nd      *navidrome.Client
	tpl     *Templates
	version string
}

// New creates a Handler with the given Navidrome client and template dispatcher.
func New(nd *navidrome.Client, tpl *Templates, version string) *Handler {
	return &Handler{nd: nd, tpl: tpl, version: version}
}

// baseData returns a map pre-populated with fields common to every page.
func (h *Handler) baseData(activeTab string) map[string]any {
	return map[string]any{
		"ActiveTab": activeTab,
		"Version":   h.version,
	}
}

// renderError sends an error response. For HTMX requests it emits a toast
// trigger header instead of writing a full error page.
func (h *Handler) renderError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if code >= 400 {
		log.Printf("error %d %s %s: %s", code, r.Method, r.URL.Path, msg)
	}
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Reswap", "none")
		w.Header().Set("HX-Trigger", `{"showToast":"`+msg+`"}`)
		w.WriteHeader(code)
		return
	}
	http.Error(w, msg, code)
}
