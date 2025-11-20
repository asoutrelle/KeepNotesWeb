package handlers

import (
	"context"
	"database/sql"
	sqlc "keepnotesweb/db/sqlc"
	"keepnotesweb/views"
	"net/http"
)

type UserHandler struct {
	queries *sqlc.Queries
}

func NewUserHandler(q *sqlc.Queries) *UserHandler {
	return &UserHandler{queries: q}
}

func (h *UserHandler) LayoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var notes, err = h.queries.ListNotes(ctx)
	if err != nil {
		http.Error(w, "Error al listar Notes", http.StatusInternalServerError)
		return
	}
	var folders, erro = h.queries.ListFolders(ctx)
	if erro != nil {
		http.Error(w, "Error al listar Folders", http.StatusInternalServerError)
		return
	}

	views.Layout("NotesKeep", views.NotesPage(notes), views.FolderPage(folders)).Render(r.Context(), w)
}

func (h *UserHandler) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	// Guardar en SQLC
	_, err := h.queries.CreateNote(r.Context(), sqlc.CreateNoteParams{
		Title: title,
		Body: sql.NullString{
			String: body,
			Valid:  body != "",
		},
		// Si tenés folder:
		// FolderID: folderID,
	})

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// PRG: Redirigir al listado
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *UserHandler) CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")

	// Guardar en SQLC
	_, err := h.queries.CreateNote(r.Context(), sqlc.CreateNoteParams{
		Title: title,
		Body: sql.NullString{
			String: body,
			Valid:  body != "",
		},
		// Si tenés folder:
		// FolderID: folderID,
	})

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// PRG: Redirigir al listado
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
