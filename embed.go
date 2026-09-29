// Package pismotx exposes project-wide embedded assets.
// A root package is required because //go:embed can only reference
// files in the same directory as (or below) the Go file that declares it,
// and the CLAUDE.md layout keeps migrations at the repo root.
package pismotx

import "embed"

//go:embed migrations/*.sql
var Migrations embed.FS
