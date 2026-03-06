package handlers

import (
	"html/template"
	"net/http"

	"github.com/user/navidrome-playlists/pkg/navidrome"
)

// Handler holds shared dependencies for all HTTP handlers.
type Handler struct {
	nd  *navidrome.Client
	tpl *template.Template
}

// New creates a Handler with the given Navidrome client and template set.
func New(nd *navidrome.Client, tpl *template.Template) *Handler {
	return &Handler{nd: nd, tpl: tpl}
}

// renderError sends an error response. For HTMX requests it emits a toast
// trigger header instead of writing a full error page.
func (h *Handler) renderError(w http.ResponseWriter, r *http.Request, msg string, code int) {
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Reswap", "none")
		w.Header().Set("HX-Trigger", `{"showToast":"`+msg+`"}`)
		w.WriteHeader(code)
		return
	}
	http.Error(w, msg, code)
}
