// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.

package crdbx

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgx" // crdb retries and postgres interface
	_ "github.com/lib/pq"                                   // Register the Postgres driver.
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewDB will open a connection to the database based on the config provided
func NewDB(cfg Config, tracing bool) (*sql.DB, error) {
	dbDriverName := "postgres"

	var err error

	if tracing {
		// Register an OTel SQL driver
		dbDriverName, err = otelsql.Register(dbDriverName,
			otelsql.WithAttributes(semconv.DBSystemCockroachdb))
		if err != nil {
			return nil, fmt.Errorf("failed creating sql tracer: %w", err)
		}
	}

	db, err := sql.Open(dbDriverName, cfg.GetURI())
	if err != nil {
		return nil, fmt.Errorf("failed connecting to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed verifying database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.Connections.MaxOpen)
	db.SetMaxIdleConns(cfg.Connections.MaxIdle)
	db.SetConnMaxIdleTime(cfg.Connections.MaxLifetime)

	return db, nil
}
