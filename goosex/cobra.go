// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package goosex

import (
	"io/fs"

	_ "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx" // crdb retries and postgres interface
	_ "github.com/lib/pq"                                   // Register the Postgres driver.
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"go.infratographer.com/x/zapx"
)

var (
	dbURI  string
	logger *zap.SugaredLogger
)

// RegisterCobraCommand will add a migrate command to the cobra command provided
// that provides a wrapper for running goose commands.
func RegisterCobraCommand(cmd *cobra.Command, setupFunc func()) {
	cmd.AddCommand(&cobra.Command{
		Use:   "migrate <command> [args]",
		Short: "Manage database schema migrations",
		Long: `Migrate provides a wrapper around the "goose" migration tool.

Commands:
up                   Migrate the DB to the most recent version available
up-by-one            Migrate the DB up by 1
up-to VERSION        Migrate the DB to a specific VERSION
down                 Roll back the version by 1
down-to VERSION      Roll back to a specific VERSION
redo                 Re-run the latest migration
reset                Roll back all migrations
status               Dump the migration status for the current DB
version              Print the current version of the database
create NAME [sql|go] Creates new migration file with the current timestamp
fix                  Apply sequential ordering to migrations
	`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			setupFunc()
			migrate(args[0], args[1:])
		},
	})
}

func migrate(command string, args []string) {
	db, err := goose.OpenDBWithDriver("postgres", dbURI)
	if err != nil {
		logger.Fatalw("failed to open DB", "error", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatalw("failed to close DB", "error", err)
		}
	}()

	if err := goose.Run(command, db, "migrations", args...); err != nil {
		logger.Fatalw("migrate command failed", "command", command, "error", err)
	}
}

// MigrateUp will run migrations and is provided as an easy way to ensure migrations are ran in test suites
func MigrateUp(uri string, fsys fs.FS) {
	dbURI = uri

	goose.SetBaseFS(fsys)
	migrate("up", nil)
}

// SetBaseFS accepts an embedded golang filesystem and sets that as the location
// for goose migration files.
func SetBaseFS(fsys fs.FS) {
	goose.SetBaseFS(fsys)
}

// SetLogger accepts a zap logger and sets it as the logger for goose output
func SetLogger(l *zap.SugaredLogger) {
	logger = l.Named("goose")
	goose.SetLogger(zapx.NewGooseLogger(logger))
}

// SetDBURI accepts a URI  and saves it for use by goose during migrations
func SetDBURI(uri string) {
	dbURI = uri
}
