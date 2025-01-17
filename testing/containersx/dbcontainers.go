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

package containersx

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// DBContainer represents a testcontainer running a database with a URI to connect to it
type DBContainer struct {
	Container testcontainers.Container
	URI       string
}

// NewCockroachDB will return a new DBContainer backed by cockroachdb
func NewCockroachDB(ctx context.Context, img string) (*DBContainer, error) {
	image := "docker.io/cockroachdb/cockroach"
	imageTag := "latest"

	if strings.Contains(img, ":") {
		p := strings.SplitN(img, ":", 2) //nolint:mnd
		imageTag = p[1]
	}

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("%s:%s", image, imageTag),
		ExposedPorts: []string{"26257/tcp", "8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
		Cmd:          []string{"start-single-node", "--insecure"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "26257")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://root@%s:%s/defaultdb?sslmode=disable", hostIP, mappedPort.Port())

	return &DBContainer{Container: container, URI: uri}, nil
}

// NewPostgresDB will return a new DBContainer backed by postgres
func NewPostgresDB(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DBContainer, error) {
	image := "docker.io/postgres"
	imageTag := "alpine"

	if strings.Contains(img, ":") {
		p := strings.SplitN(img, ":", 2) //nolint:mnd
		imageTag = p[1]
	}

	uriFunc := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())
	}

	opts = append(opts,
		testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port("5432"), "postgres", uriFunc)),
		postgres.WithPassword("postgres"),
	)

	container, err := postgres.Run(ctx, fmt.Sprintf("%s:%s", image, imageTag), opts...)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", hostIP, mappedPort.Port())

	return &DBContainer{Container: container, URI: uri}, nil
}
