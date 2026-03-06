package handlers

import "net/http"

// Stub handlers — implemented in later tasks.

func (h *Handler) NewPlaylist(w http.ResponseWriter, r *http.Request)  { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request) { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) EditPlaylist(w http.ResponseWriter, r *http.Request)  { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) NewSmart(w http.ResponseWriter, r *http.Request)      { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) CreateSmart(w http.ResponseWriter, r *http.Request)   { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) EditSmart(w http.ResponseWriter, r *http.Request)     { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) UpdateSmart(w http.ResponseWriter, r *http.Request)   { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) Export(w http.ResponseWriter, r *http.Request)        { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) BatchDelete(w http.ResponseWriter, r *http.Request)   { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) BatchExport(w http.ResponseWriter, r *http.Request)   { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) ImportForm(w http.ResponseWriter, r *http.Request)    { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) Import(w http.ResponseWriter, r *http.Request)        { http.Error(w, "not implemented", http.StatusNotImplemented) }
func (h *Handler) ImportConfirm(w http.ResponseWriter, r *http.Request) { http.Error(w, "not implemented", http.StatusNotImplemented) }
