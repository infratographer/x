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

package goosex_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.infratographer.com/x/goosex"
)

var (
	migrations = fstest.MapFS{
		"migrations/0001_create-testtable.sql": &fstest.MapFile{
			Data: []byte(`
-- +goose Up
-- create "testtable" table
CREATE TABLE "testtable" (
	"id" character varying NOT NULL,
	"name" character varying(64) NOT NULL,
	PRIMARY KEY ("id")
);
-- +goose Down
-- reverse: create "testtable" table
DROP TABLE "testtable";
			`),
		},
	}
)

func TestMigrateUpContext(t *testing.T) {
	logger, err := zap.NewDevelopmentConfig().Build()
	require.NoError(t, err)

	server, err := testserver.NewTestServer()
	require.NoError(t, err)

	defer server.Stop()

	ctx := context.Background()

	goosex.SetLogger(logger.Sugar())
	goosex.MigrateUpContext(ctx, server.PGURL().String(), migrations)

	db, err := goose.OpenDBWithDriver("postgres", server.PGURL().String())
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'")
	require.NoError(t, err)

	var (
		testTableFound     bool
		versionsTableFound bool
	)

	for rows.Next() {
		var tableName string

		if err := rows.Scan(&tableName); err != nil {
			require.NoError(t, err)
		}

		switch tableName {
		case "testtable":
			testTableFound = true
		case "goose_db_version":
			versionsTableFound = true
		}
	}

	assert.True(t, testTableFound, "expected test table to exist")
	assert.True(t, versionsTableFound, "expected goose versions table to exist")
}
