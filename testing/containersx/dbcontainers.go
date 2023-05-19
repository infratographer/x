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
	imageTag := "latest-v22.2"

	if strings.Contains(img, ":") {
		p := strings.SplitN(img, ":", 2) //nolint:gomnd
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
		p := strings.SplitN(img, ":", 2) //nolint:gomnd
		imageTag = p[1]
	}

	uriFunc := func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://postgres:postgres@%s:%s/postgres?sslmode=disable", host, port.Port())
	}

	opts = append(opts,
		testcontainers.WithImage(fmt.Sprintf("%s:%s", image, imageTag)),
		testcontainers.WithWaitStrategy(wait.ForSQL(nat.Port("5432"), "postgres", uriFunc)),
		postgres.WithPassword("postgres"),
	)

	container, err := postgres.RunContainer(ctx, opts...)
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
