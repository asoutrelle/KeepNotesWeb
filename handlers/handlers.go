package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"keepnotesweb/views"
	"log"
	"net/http"
	"strconv"

	sqlc "keepnotesweb/db/sqlc"

	"github.com/a-h/templ"
)

type UserHandler struct {
	queries *sqlc.Queries
}

func NewUserHandler(q *sqlc.Queries) *UserHandler {
	return &UserHandler{queries: q}
}

func (h *UserHandler) NotesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getNotes(w, r)
	case "POST":
		h.createNote(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}
func (h *UserHandler) NoteHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.getNoteByID(w, r)
	case "PUT":
		h.updateNote(w, r)
	case "DELETE":
		h.deleteNote(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) getNotes(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	notes, err := h.queries.ListNotes(ctx)
	if err != nil {
		http.Error(w, "Error al listar notas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *UserHandler) createNote(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	var note struct {
		Title    string  `json:"title"`
		Body     *string `json:"body"`      // puntero para distinguir NULL
		FolderID *int32  `json:"folder_id"` // puntero para distinguir NULL
	}

	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if note.Title == "" {
		http.Error(w, "El título es obligatorio", http.StatusBadRequest)
		return
	}

	// Preparar parámetros para sqlc
	params := sqlc.CreateNoteParams{
		Title: note.Title,
	}

	if note.Body != nil {
		params.Body = sql.NullString{String: *note.Body, Valid: true}
	} else {
		params.Body = sql.NullString{Valid: false}
	}

	if note.FolderID != nil {
		params.FolderID = sql.NullInt32{Int32: *note.FolderID, Valid: true}
	} else {
		params.FolderID = sql.NullInt32{Valid: false}
	}

	// Crear la nota en la base de datos
	createdNote, err := h.queries.CreateNote(ctx, params)
	if err != nil {
		http.Error(w, "Error al crear nota: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdNote)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) getNoteByID(w http.ResponseWriter, r *http.Request) {
	// Obtener ID desde URL
	idStr := r.URL.Path[len("/api/notes/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	// Buscar en la base de datos
	note, err := h.queries.GetNote(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(note)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}

}

func (h *UserHandler) updateNote(w http.ResponseWriter, r *http.Request) {
	// Obtener ID desde URL
	idStr := r.URL.Path[len("/api/notes/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	// Buscar en la base de datos
	note, err := h.queries.GetNote(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	// codigo para updatear nota
	var input struct {
		Title    string  `json:"title"`
		Body     *string `json:"body"`
		FolderID *int32  `json:"folder_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	params := sqlc.UpdateNoteParams{
		ID:    int32(id),
		Title: input.Title,
	}

	if input.Body != nil {
		params.Body = sql.NullString{String: *input.Body, Valid: true}
	} else {
		params.Body = sql.NullString{Valid: false}
	}

	if input.FolderID != nil {
		params.FolderID = sql.NullInt32{Int32: *input.FolderID, Valid: true}
	} else {
		params.FolderID = sql.NullInt32{Valid: false}
	}

	err = h.queries.UpdateNote(r.Context(), params)
	if err != nil {
		http.Error(w, "Error al actualizar la nota", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(note)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}

}

func (h *UserHandler) deleteNote(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/notes/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	err = h.queries.DeleteNote(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Error al borrar la nota", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) FoldersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📌 NoteHandler llamado con método:", r.Method)
	switch r.Method {
	case "GET":
		h.getFolders(w, r)
	case "POST":
		h.createFolder(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) FolderHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📌 NoteHandler llamado con método:", r.Method)
	switch r.Method {
	case "GET":
		h.getFolderByID(w, r)
	case "PUT":
		h.updateFolder(w, r)
	case "POST":
		h.createFolder(w, r)
	case "DELETE":
		h.deleteFolder(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) getFolders(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	folders, err := h.queries.ListFolders(ctx)
	if err != nil {
		http.Error(w, "Error al listar carpetas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}

func (h *UserHandler) createFolder(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var folder struct {
		UserID         *int32  `json:"user_id"`
		Name           string  `json:"name"`
		Description    *string `json:"description"`
		ParentFolderID *int32  `json:"parent_folder_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if folder.Name == "" {
		http.Error(w, "El nombre es obligatorio", http.StatusBadRequest)
		return
	}

	params := sqlc.CreateFolderParams{
		Name: folder.Name,
	}

	if folder.UserID != nil {
		params.UserID = sql.NullInt32{Int32: *folder.UserID, Valid: true}
	} else {
		params.UserID = sql.NullInt32{Valid: false}
	}

	if folder.Description != nil {
		params.Description = sql.NullString{String: *folder.Description, Valid: true}
	} else {
		params.Description = sql.NullString{Valid: false}
	}

	if folder.ParentFolderID != nil {
		params.ParentFolderID = sql.NullInt32{Int32: *folder.ParentFolderID, Valid: true}
	} else {
		params.ParentFolderID = sql.NullInt32{Valid: false}
	}

	createdFolder, err := h.queries.CreateFolder(ctx, params)
	if err != nil {
		http.Error(w, "Error al crear carpeta: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(createdFolder)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) getFolderByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/folders/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	// Buscar en la base de datos
	folder, err := h.queries.GetFolder(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(folder)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) updateFolder(w http.ResponseWriter, r *http.Request) {
	// Obtener ID desde URL
	idStr := r.URL.Path[len("/api/folders/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	// Buscar en la base de datos
	folder, err := h.queries.GetFolder(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	// codigo para updatear nota
	var input struct {
		Name           string  `json:"name"`
		Description    *string `json:"description"`
		ParentFolderID *int32  `json:"parent_folder_id"`
		UserID         *int32  `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	params := sqlc.UpdateFolderParams{
		ID:   int32(id),
		Name: input.Name,
	}

	if input.Description != nil {
		params.Description = sql.NullString{String: *input.Description, Valid: true}
	} else {
		params.Description = sql.NullString{Valid: false}
	}

	if input.ParentFolderID != nil {
		params.ParentFolderID = sql.NullInt32{Int32: *input.ParentFolderID, Valid: true}
	} else {
		params.ParentFolderID = sql.NullInt32{Valid: false}
	}

	if input.UserID != nil {
		params.UserID = sql.NullInt32{Int32: *input.UserID, Valid: true}
	} else {
		params.UserID = sql.NullInt32{Valid: false}
	}

	err = h.queries.UpdateFolder(r.Context(), params)
	if err != nil {
		http.Error(w, "Error al actualizar la Carpeta", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(folder)
	if err != nil {
		http.Error(w, "Error al codificar JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) deleteFolder(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/folders/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	err = h.queries.DeleteFolder(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Error al borrar la carpeta", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// ============= USERS HANDLERS =============

func (h *UserHandler) UsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📌 UsersHandler llamado con método:", r.Method)
	switch r.Method {
	case "GET":
		h.getUsers(w, r)
	case "POST":
		h.createUser(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) SingleUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("📌 SingleUserHandler llamado con método:", r.Method)
	switch r.Method {
	case "GET":
		h.getUserByID(w, r)
	case "PUT":
		h.updateUser(w, r)
	case "DELETE":
		h.deleteUser(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	users, err := h.queries.ListUsers(ctx)
	if err != nil {
		http.Error(w, "Error al listar usuarios: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Username, email y password son obligatorios", http.StatusBadRequest)
		return
	}

	// TODO: Aquí deberías hashear la contraseña con bcrypt
	// Por ahora la guardamos en texto plano (NO SEGURO para producción)
	params := sqlc.CreateUserParams{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}

	createdUser, err := h.queries.CreateUser(ctx, params)
	if err != nil {
		http.Error(w, "Error al crear usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUser(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Usuario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	// No devolver la contraseña en la respuesta
	response := struct {
		ID        int32  `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at,omitempty"`
	}{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Verificar que el usuario existe
	_, err = h.queries.GetUser(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Usuario no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "Username, email y password son obligatorios", http.StatusBadRequest)
		return
	}

	// TODO: Hashear la contraseña con bcrypt
	params := sqlc.UpdateUserParams{
		ID:       int32(id),
		Username: input.Username,
		Email:    input.Email,
		Password: input.Password,
	}

	err = h.queries.UpdateUser(r.Context(), params)
	if err != nil {
		http.Error(w, "Error al actualizar usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario actualizado correctamente"})
}

func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/users/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	err = h.queries.DeleteUser(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Error al borrar usuario: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// Login handler
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "Username y password son obligatorios", http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByUsername(ctx, credentials.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	// TODO: Comparar con bcrypt.CompareHashAndPassword
	// Por ahora comparación en texto plano (NO SEGURO)
	if user.Password != credentials.Password {
		http.Error(w, "Credenciales inválidas", http.StatusUnauthorized)
		return
	}

	// Login exitoso - devolver datos del usuario (sin password)
	response := struct {
		ID       int32  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Message  string `json:"message"`
	}{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Message:  "Login exitoso",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) ListFoldersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var folders, err = h.queries.ListFolders(ctx)
	if err != nil {
		http.Error(w, "Error al listar Folders", http.StatusInternalServerError)
		return
	}
	templ.Handler(views.FolderPage(folders)).ServeHTTP(w, r)
}

func (h *UserHandler) ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var notes, err = h.queries.ListNotes(ctx)
	if err != nil {
		http.Error(w, "Error al listar Notes", http.StatusInternalServerError)
		return
	}
	templ.Handler(views.NotesPage(notes)).ServeHTTP(w, r)
}

func (h *UserHandler) LayoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var notes, err = h.queries.ListNotes(ctx)
	if err != nil {
		http.Error(w, "Error al listar Notes", http.StatusInternalServerError)
		return
	}
	templ.Handler(views.Layout("NotesKeep", views.NotesPage(notes))).ServeHTTP(w, r)
}
