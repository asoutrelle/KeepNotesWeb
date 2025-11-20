package handlers

import (
	"context"
	"database/sql"
	sqlc "keepnotesweb/db/sqlc"
	"keepnotesweb/views"
	"net/http"
	"strconv"
	"strings"
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

	views.Layout("Home", views.NotesPage(notes, ""), views.FolderPage(folders)).Render(r.Context(), w)
}

func (h *UserHandler) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form values
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	folderIDStr := r.FormValue("folder_id") // <- se lee del form

	folderID, err := strconv.Atoi(folderIDStr)

	var note sqlc.CreateNoteParams

	if err == nil {
		// folderID es válido
		note = sqlc.CreateNoteParams{
			FolderID: sql.NullInt32{
				Int32: int32(folderID),
				Valid: true,
			},
			Title: title,
			Body: sql.NullString{
				String: body,
				Valid:  body != "",
			},
		}
	} else {
		// folderID invalido → no asignar folder
		note = sqlc.CreateNoteParams{
			Title: title,
			Body: sql.NullString{
				String: body,
				Valid:  body != "",
			},
		}
	}

	// Guardar nota usando SQLC
	_, err = h.queries.CreateNote(r.Context(), note)

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if folderIDStr != "" {
		http.Redirect(w, r, "/folders/"+folderIDStr, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
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

	name := r.FormValue("name")
	description := r.FormValue("description")

	folder, err := h.queries.CreateFolder(r.Context(), sqlc.CreateFolderParams{
		Name: name,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
	})

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/folders/"+strconv.Itoa(int(folder.ID)), http.StatusSeeOther)

}

func (h *UserHandler) ListNotesByFolderID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	idStr := parts[len(parts)-1] // "7"

	// Convertir string → int
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	n := sql.NullInt32{
		Int32: int32(idInt),
		Valid: true,
	}

	// Obtener todas las carpetas para el sidebar
	folders, err := h.queries.ListFolders(r.Context())
	if err != nil {
		http.Error(w, "error cargando folders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	notes, err := h.queries.GetNotesByFolder(r.Context(), n)
	if err != nil {
		http.Error(w, "Error fetching notes", http.StatusInternalServerError)
		return
	}

	sidebar := views.FolderPage(folders)
	body := views.NotesPage(notes, idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	folder, err := h.queries.GetFolder(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Folder not found", http.StatusNotFound)
		return
	}

	views.Layout(folder.Name, body, sidebar).Render(r.Context(), w)
}
