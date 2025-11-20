-- name: GetFolder :one
SELECT id, user_id, name, description, parent_folder_id, created_at
FROM folder
WHERE id = $1;

-- name: GetNote :one
SELECT id, folder_id, title, body, created_at, updated_at
FROM note
WHERE id = $1;

-- name: GetNotesByFolder :many
SELECT id, folder_id, title, body, created_at, updated_at
FROM note
WHERE folder_id = $1;

-- name: ListFolders :many
SELECT id, user_id, name, description, parent_folder_id, created_at
FROM folder
ORDER BY name;

-- name: ListFoldersByUser :many
SELECT id, user_id, name, description, parent_folder_id, created_at
FROM folder
WHERE user_id = $1
ORDER BY name;

-- name: ListNotes :many
SELECT id, folder_id, title, body, created_at, updated_at
FROM note
ORDER BY title;

-- name: CreateFolder :one
INSERT INTO folder (user_id, name, description, parent_folder_id)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, description, parent_folder_id, created_at;

-- name: CreateNote :one
INSERT INTO note (title, body, folder_id)
VALUES ($1, $2, $3)
RETURNING id, title, body, folder_id, created_at;

-- name: UpdateFolder :exec
UPDATE folder
SET name = $2, description = $3, parent_folder_id = $4, user_id = $5
WHERE id = $1;

-- name: UpdateNote :exec
UPDATE note
SET title = $2, body = $3, folder_id = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteFolder :exec
DELETE FROM folder
WHERE id = $1;

-- name: DeleteNote :exec
DELETE FROM note
WHERE id = $1;

-- name: GetUser :one
SELECT id, username, email, password, created_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password, created_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password, created_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, username, email, created_at
FROM users
ORDER BY username;

-- name: CreateUser :one
INSERT INTO users (username, email, password)
VALUES ($1, $2, $3)
RETURNING id, username, email, created_at;

-- name: UpdateUser :exec
UPDATE users
SET username = $2, email = $3, password = $4
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;