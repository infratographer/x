package pubsubx

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Publisher struct {
	topic  *pubsub.Topic
	logger *zap.SugaredLogger
}

// func NewPublisher(ctx context.Context, psURL string, logger *zap.SugaredLogger, opts ...SubscriptionOption) (*Publisher, error) {
// 	u, err := url.Parse(psURL)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch u.Scheme {
// 	case "nats":
// 		fmt.Println("Start a NATs Sub")
// 	default:
// 		return nil, errors.New("currently only NATs is supported for pubsub")
// 	}

// 	natsConn, err := nats.Connect(psURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// defer natsConn.Drain()

// 	// natsConn.Close()

// 	sub, err := natspubsub.OpenSubscription(
// 		natsConn,
// 		"com.infratographer.events.*.*",
// 		&natspubsub.SubscriptionOptions{Queue: "permissionsapi"})
// 	if err != nil {
// 		return nil, err
// 	}

// 	logger = logger.Named("worker")

// 	return &Subscription{sub: sub, logger: logger}, nil
// }

func HackySendMsg(ctx context.Context, t string, msg *Message) error {
	natsConn, err := nats.Connect("nats://localhost")
	if err != nil {
		return err
	}
	defer natsConn.Close()

	topic, err := natspubsub.OpenTopic(natsConn, t, nil)
	if err != nil {
		return err
	}
	defer topic.Shutdown(ctx)

	return SendMsg(topic, msg)
}

func SendMsg(topic *pubsub.Topic, msg *Message) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return topic.Send(context.Background(), &pubsub.Message{
		Body: msgBytes,
	})
}
