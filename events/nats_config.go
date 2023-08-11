package events

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"go.infratographer.com/x/viperx"
)

var (
	// NATSDefaultConnectTimeout is the default connection timeout.
	NATSDefaultConnectTimeout = 5 * time.Second
	// NATSDefaultSubscriberFetchBatchSize is the default pull subscribe batch size.
	NATSDefaultSubscriberFetchBatchSize = 20
	// NATSDefaultSubscriberFetchTimeout is the max time a fetch will block before releasing.
	NATSDefaultSubscriberFetchTimeout = 5 * time.Second
	// NATSDefaultSubscriberFetchBackoff is the delay between a batch attempts.
	NATSDefaultSubscriberFetchBackoff = 5 * time.Second
	// NATSDefaultShutdownTimeout is the timeout for a shutdown to complete.
	NATSDefaultShutdownTimeout = 5 * time.Second
)

// NATSConfig defines the NATS connection configuration.
type NATSConfig struct {
	URL             string
	SubscribePrefix string
	PublishPrefix   string
	QueueGroup      string
	Token           string
	CredsFile       string
	Source          string

	ConnectTimeout           time.Duration
	ShutdownTimeout          time.Duration
	SubscriberFetchBatchSize int
	SubscriberFetchTimeout   time.Duration
	SubscriberFetchBackoff   time.Duration
	SubscriberNoAckExplicit  bool
	SubscriberNoManualAck    bool

	SubscriberDeliveryPolicy string
	SubscriberStartSequence  uint64
	SubscriberStartTime      time.Time

	logger           *zap.SugaredLogger
	connectOptions   []nats.Option
	jetStreamOptions []nats.JSOpt
	subscribeOptions []nats.SubOpt
}

// Configured checks whether the provider has been configured.
func (c NATSConfig) Configured() bool {
	return c.URL != "" || c.QueueGroup != ""
}

// Validate ensures the configuration is valid.
func (c NATSConfig) Validate() error {
	var err error

	if c.Token != "" && c.CredsFile != "" {
		err = multierr.Append(err, ErrNATSInvalidAuthConfiguration)
	}

	switch c.SubscriberDeliveryPolicy {
	case "", "all", "last", "last-per-subject", "new", "start-sequence", "start-time":
	default:
		err = multierr.Append(err, ErrNATSInvalidDeliveryPolicy)
	}

	return err
}

// WithDefaults sets default values for the field unset.
func (c NATSConfig) WithDefaults() NATSConfig {
	if c.logger == nil {
		c.logger = zap.NewNop().Sugar()
	}

	if c.SubscriberFetchBatchSize == 0 {
		c.SubscriberFetchBatchSize = NATSDefaultSubscriberFetchBatchSize
	}

	if c.SubscriberFetchTimeout == 0 {
		c.SubscriberFetchTimeout = NATSDefaultSubscriberFetchTimeout
	}

	if c.SubscriberFetchBackoff == 0 {
		c.SubscriberFetchBackoff = NATSDefaultSubscriberFetchBackoff
	}

	if !c.SubscriberNoAckExplicit {
		c.subscribeOptions = append(c.subscribeOptions, nats.AckExplicit())
	}

	if !c.SubscriberNoManualAck {
		c.subscribeOptions = append(c.subscribeOptions, nats.ManualAck())
	}

	switch c.SubscriberDeliveryPolicy {
	case "all":
		c.subscribeOptions = append(c.subscribeOptions, nats.DeliverAll())
	case "last":
		c.subscribeOptions = append(c.subscribeOptions, nats.DeliverLast())
	case "last-per-subject":
		c.subscribeOptions = append(c.subscribeOptions, nats.DeliverLastPerSubject())
	case "new":
		c.subscribeOptions = append(c.subscribeOptions, nats.DeliverNew())
	case "start-sequence":
		c.subscribeOptions = append(c.subscribeOptions, nats.StartSequence(c.SubscriberStartSequence))
	case "start-time":
		c.subscribeOptions = append(c.subscribeOptions, nats.StartTime(c.SubscriberStartTime))
	}

	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = NATSDefaultShutdownTimeout
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = NATSDefaultConnectTimeout
	}

	c.connectOptions = append(c.connectOptions, nats.Timeout(c.ConnectTimeout))

	if c.Token != "" {
		c.connectOptions = append(c.connectOptions, nats.Token(c.Token))
	}

	if c.CredsFile != "" {
		c.connectOptions = append(c.connectOptions, nats.UserCredentials(c.CredsFile))
	}

	if c.Source != "" {
		c.connectOptions = append(c.connectOptions, nats.Name(c.Source))
	}

	return c
}

// NATSOption defines a nats configuration option.
type NATSOption func(c *NATSConfig) error

// WithNATSLogger sets the logger for the nats connection.
func WithNATSLogger(logger *zap.SugaredLogger) NATSOption {
	return func(c *NATSConfig) error {
		c.logger = logger

		return nil
	}
}

// WithNATSConnectOptions configures the connection options for nats.
func WithNATSConnectOptions(options ...nats.Option) NATSOption {
	return func(c *NATSConfig) error {
		c.connectOptions = append(c.connectOptions, options...)

		return nil
	}
}

// WithNATSJetStreamOptions configures the jetstream connection options.
func WithNATSJetStreamOptions(options ...nats.JSOpt) NATSOption {
	return func(c *NATSConfig) error {
		c.jetStreamOptions = append(c.jetStreamOptions, options...)

		return nil
	}
}

// WithNATSSubscribeOptions configures the subscribe options for new subscriptions.
func WithNATSSubscribeOptions(options ...nats.SubOpt) NATSOption {
	return func(c *NATSConfig) error {
		c.subscribeOptions = append(c.subscribeOptions, options...)

		return nil
	}
}

// MustViperFlagsForNATS returns the cobra flags and viper config for a nats handler.
func MustViperFlagsForNATS(v *viper.Viper, flags *pflag.FlagSet, appName string) {
	flags.String("events-nats-url", "nats://nats:4222", "nats server connection url")
	viperx.MustBindFlag(v, "events.nats.url", flags.Lookup("events-nats-url"))

	v.MustBindEnv("events.nats.subscribePrefix")
	v.MustBindEnv("events.nats.publishPrefix")
	v.MustBindEnv("events.nats.queueGroup")
	v.MustBindEnv("events.nats.token")
	v.MustBindEnv("events.nats.credsFile")
	v.MustBindEnv("events.nats.source")
	v.MustBindEnv("events.nats.connectTimeout")
	v.MustBindEnv("events.nats.shutdownTimeout")
	v.MustBindEnv("events.nats.subscriberFetchBatchSize")
	v.MustBindEnv("events.nats.subscriberFetchTimeout")
	v.MustBindEnv("events.nats.subscriberFetchBackoff")
	v.MustBindEnv("events.nats.subscriberNoAckExplicit")
	v.MustBindEnv("events.nats.subscriberNoManualAck")
	v.MustBindEnv("events.nats.subscriberDeliveryPolicy")
	v.MustBindEnv("events.nats.subscriberStartSequence")
	v.MustBindEnv("events.nats.subscriberStartTime")

	v.SetDefault("events.nats.connectTimeout", defaultTimeout)
	v.SetDefault("events.nats.source", appName)
}
