#!/usr/bin/env bash
set -euo pipefail

unformatted="$(gofmt -l cmd src)"
if [[ -n "$unformatted" ]]; then
  printf '%s\n' "$unformatted"
  exit 1
fi

go test ./...
go vet ./...
npm run check
npm test
npm run format
npm run loc

