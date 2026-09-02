#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
go test ./...
go vet ./...
mkdir -p dist
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
  -trimpath \
  -buildvcs=false \
  -ldflags='-s -w -buildid= -H=windowsgui' \
	-o dist/BomberRush.exe \
	./cmd/bomberrush
printf 'Gotowe: %s/dist/BomberRush.exe\n' "$PWD"
