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

package eventtools

import (
	"errors"
	"fmt"
	"os"
	"sync"
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

	// ErrNack is returned if a nack is received instead of an ack
	ErrNack = errors.New("nack received")
	// ErrNoAck is returned when no ack is received and the timeout was hit
	ErrNoAck = errors.New("no ack received")
)

// TestNats maintains the nats environment
type TestNats struct {
	Server    *server.Server
	Conn      *nats.Conn
	JetStream nats.JetStreamContext
	Config    events.Config
}

// Close closes the connection
func (s *TestNats) Close() {
	s.Conn.Close() //nolint:errcheck
}

// Deprecated: Use CaptureMsgAck instead.
//
//nolint:revive
func (s *TestNats) SetConsumerSampleFrequency(consumer, frequency string) error {
	return s.setConsumerSampleFrequency(consumer, frequency)
}

func (s *TestNats) setConsumerSampleFrequency(consumer, frequency string) error {
	info, err := s.JetStream.ConsumerInfo("events-tests", consumer)
	if err != nil {
		return err
	}

	cfg := info.Config
	cfg.SampleFrequency = frequency

	_, err = s.JetStream.UpdateConsumer("events-tests", &cfg)
	if err != nil {
		return err
	}

	return nil
}

// Deprecated: Use CaptureMsgAck instead.
//
//nolint:revive
func (s *TestNats) WaitForAck(consumer string, timeout time.Duration) error { //nolint:revive
	return s.waitForAck(consumer, timeout)
}

func (s *TestNats) waitForAck(consumer string, timeout time.Duration) error {
	// We should only ever receive one Ack, so we close the channel directly if we get one.
	ackCh := make(chan struct{})
	ackSub, err := s.Conn.Subscribe("$JS.EVENT.METRIC.CONSUMER.ACK.*."+consumer, func(_ *nats.Msg) {
		close(ackCh)
	})

	if err != nil {
		return err
	}

	defer ackSub.Unsubscribe() //nolint:errcheck

	// We may receive many Naks in a single test, so we use a sync.Once to close the channel.
	nakCh := make(chan struct{})

	var nakOnce sync.Once

	nakSub, err := s.Conn.Subscribe("$JS.EVENT.ADVISORY.CONSUMER.MSG_NAKED.*."+consumer, func(_ *nats.Msg) {
		nakOnce.Do(func() {
			close(nakCh)
		})
	})

	if err != nil {
		return err
	}

	defer nakSub.Unsubscribe() //nolint:errcheck

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ackCh:
		return nil
	case <-nakCh:
		return ErrNack
	case <-timer.C:
		return ErrNoAck
	}
}

// CaptureMsgAck waits for an ack message to be received, returns error if Nack or timeout is hit.
// To ensure Acks are received, ensure you have set ManualAck, AckExplicit and Durable subscriber options.
// This call requires a function that reads the message from the queue that you are attempting to capture the ack from.
func (s *TestNats) CaptureMsgAck(consumer string, timeout time.Duration, msgFunc func()) error {
	err := s.setConsumerSampleFrequency(consumer, "100")
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.waitForAck(consumer, timeout)
	}()

	// without this there is still a small failure rate where we miss the ack message from nats
	time.Sleep(200 * time.Microsecond) //nolint:mnd
	msgFunc()

	return <-errCh
}

// NewNatsServer returns a simple NATs server that starts and stores it's data in a tmp dir
func NewNatsServer() (*TestNats, error) {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "test-nats")
	if err != nil {
		return nil, fmt.Errorf("failed making tmp dir for nats storage: %w", err)
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
		return nil, fmt.Errorf("building nats server: %w", err)
	}

	// uncomment to enable nats server logging
	// s.ConfigureLogger()

	if err = server.Run(s); err != nil {
		return nil, err
	}

	if !s.ReadyForConnections(natsTimeout) {
		return nil, errors.New("starting nats server: timeout") //nolint:err113
	}

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "events-tests",
		Subjects: Subjects,
	})
	if err != nil {
		return nil, err
	}

	return &TestNats{
		Server:    s,
		Conn:      nc,
		JetStream: js,
		Config: events.Config{
			NATS: events.NATSConfig{
				URL:             s.ClientURL(),
				SubscribePrefix: Prefix,
				PublishPrefix:   Prefix,
			},
		},
	}, nil
}
