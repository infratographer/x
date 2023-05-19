package eventtools

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"

	"go.infratographer.com/x/events"
)

const (
	natsTimeout    = 2 * time.Second
	maxControlLine = 2048
)

var (
	// Prefix to use when creating the nats server jetstream subjects
	Prefix = "com.infratographer.testing"
	// Subjects to create in jetstream
	Subjects = []string{Prefix + ".events.>", Prefix + ".changes.>"}
)

// NewNatsServer returns a simple NATs server that starts and stores it's data in a tmp dir
func NewNatsServer() (pCfg events.PublisherConfig, sCfg events.SubscriberConfig, err error) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "tenant-nats")
	if err != nil {
		err = fmt.Errorf("failed making tmp dir for nats storage: %w", err)
		return
	}

	s, err := server.NewServer(&server.Options{
		Host:           "127.0.0.1",
		Debug:          false,
		Trace:          false,
		TraceVerbose:   false,
		Port:           server.RANDOM_PORT,
		NoLog:          false,
		NoSigs:         true,
		MaxControlLine: maxControlLine,
		JetStream:      true,
		StoreDir:       tmpdir,
	})
	if err != nil {
		err = fmt.Errorf("building nats server: %w", err)
		return
	}

	// uncomment to enable nats server logging
	// s.ConfigureLogger()

	if err = server.Run(s); err != nil {
		return
	}

	if !s.ReadyForConnections(natsTimeout) {
		err = errors.New("starting nats server: timeout") //nolint:goerr113
		return
	}

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		return
	}

	js, err := nc.JetStream()
	if err != nil {
		return
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "events-tests",
		Subjects: Subjects,
	})
	if err != nil {
		return
	}

	pCfg.URL = s.ClientURL()
	pCfg.Prefix = Prefix

	sCfg.URL = s.ClientURL()
	sCfg.Prefix = Prefix

	return
}
