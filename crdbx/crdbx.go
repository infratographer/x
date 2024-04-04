// Copyright 2023 The Infratographer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crdbx

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	_ "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5" // crdb retries and postgres interface
	_ "github.com/jackc/pgx/v5/stdlib"                        // Register pgx driver.
	_ "github.com/lib/pq"                                     // Register the Postgres driver.
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
