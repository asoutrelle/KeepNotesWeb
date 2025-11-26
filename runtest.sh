#!/bin/bash
set -e

echo "Ejecutando sqlc generate..."
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate

echo "Ejecutando templ generate..."
go run github.com/a-h/templ/cmd/templ@latest generate

echo "Levantando Docker..."
docker compose up --build

echo "Contendedor terminado. Prueba finalizada."
