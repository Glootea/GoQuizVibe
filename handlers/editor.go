package handlers

import (
	"net/http"

	"github.com/goquizvibe/pages/editor"
)

type EditorHandler struct{}

func NewEditor() *EditorHandler {
	return &EditorHandler{}
}

func (h *EditorHandler) EditorPage(w http.ResponseWriter, r *http.Request) error {
	return editor.EditorPage().Render(r.Context(), w)
}