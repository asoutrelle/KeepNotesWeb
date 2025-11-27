package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	handlerDB "keepnotesweb/db/handlers"
	sqlc "keepnotesweb/db/sqlc"
	"keepnotesweb/handlers"
)

func main() {
	staticDir := "./static"
	fileServer := http.FileServer(http.Dir(staticDir))
	port := ":8080"
	conn, err := handlerDB.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	queries := sqlc.New(conn)
	userHandler := handlers.NewUserHandler(queries)

	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/", userHandler.LayoutHandler)
	http.HandleFunc("/notes", userHandler.CreateNoteHandler)
	http.HandleFunc("/notes/", userHandler.DeleteNoteHandler)
	http.HandleFunc("/folders", userHandler.CreateFolderHandler)
	http.HandleFunc("/folders/", func(w http.ResponseWriter, r *http.Request) {
		// Detectar si es ruta /folders/{id}/notes para HTMX
		if strings.HasSuffix(r.URL.Path, "/notes") {
			userHandler.GetNotesByFolderHandler(w, r)
			return
		}
		// Ruta normal /folders/{id}
		userHandler.ListNotesByFolderID(w, r)
	})

	fmt.Printf("Servidor ESTÁTICO escuchando en http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
