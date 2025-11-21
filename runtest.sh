#!/usr/bin/env bash
set -euo pipefail

BLUE='\033[1;34m'
RED='\033[1;31m'
NC='\033[0m'

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
err()  { echo -e "${RED}[ERROR]${NC} $1"; }

COMPOSE_FILE="docker-compose.yml"

info "Levantando Docker Compose..."
docker compose -f "$COMPOSE_FILE" up -d

info "Ejecutando sqlc generate..."
sqlc generate

info "Ejecutando templ generate..."
templ generate

info "Compilando servidor..."
go build -o bin/server ./cmd/server

info "Iniciando servidor local..."
./bin/server
