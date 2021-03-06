package core

import (
	"context"
	"time"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/repository"
)

type Configuration map[string]string

type BridgeImpl interface {
	// Target return the target of the bridge (e.g.: "github")
	Target() string

	// Configure handle the user interaction and return a key/value configuration
	// for future use
	Configure(repo repository.RepoCommon, params BridgeParams) (Configuration, error)

	// ValidateConfig check the configuration for error
	ValidateConfig(conf Configuration) error

	// NewImporter return an Importer implementation if the import is supported
	NewImporter() Importer

	// NewExporter return an Exporter implementation if the export is supported
	NewExporter() Exporter
}

type Importer interface {
	Init(conf Configuration) error
	ImportAll(ctx context.Context, repo *cache.RepoCache, since time.Time) (<-chan ImportResult, error)
}

type Exporter interface {
	Init(conf Configuration) error
	ExportAll(ctx context.Context, repo *cache.RepoCache, since time.Time) (<-chan ExportResult, error)
}
