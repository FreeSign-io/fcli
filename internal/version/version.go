// Package version exposes build-time information injected via -ldflags.
package version

// Build metadata, populated at link time by goreleaser:
//
//	go build -ldflags "-X github.com/FreeSign-io/fcli/internal/version.Version=v0.1.0 \
//	                   -X github.com/FreeSign-io/fcli/internal/version.Commit=abcdef \
//	                   -X github.com/FreeSign-io/fcli/internal/version.Date=2026-05-04T00:00:00Z"
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
