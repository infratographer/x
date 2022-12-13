package pubsubx

var (
	tracer = otel.Tracer("go.infratographer.com/x/pubsubx")
)

type Subscription struct {
	sub    *pubsub.Subscription
	logger *zap.SugaredLogger
}

type SubscriptionOption func(*SubscriptionOptions) error

// Options can be used to create a customized connection.
type SubscriptionOptions struct {
	Queue string
}

func NewSubscription(ctx context.Context, psURL string, logger *zap.SugaredLogger, opts ...SubscriptionOption) (*Subscription, error) {
	u, err := url.Parse(psURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "nats":
		fmt.Println("Start a NATs Sub")
	default:
		return nil, errors.New("currently only NATs is supported for pubsub")
	}

	natsConn, err := nats.Connect(psURL)
	if err != nil {
		return nil, err
	}
	// defer natsConn.Drain()

	// natsConn.Close()

	sub, err := natspubsub.OpenSubscription(
		natsConn,
		"com.infratographer.events.*.*",
		&natspubsub.SubscriptionOptions{Queue: "permissionsapi"})
	if err != nil {
		return nil, err
	}

	logger = logger.Named("worker")

	return &Subscription{sub: sub, logger: logger}, nil
}

func (s *Subscription) Receive(ctx context.Context, st *query.Stores) error {
	ctx, span := tracer.Start(ctx, "HandleMessage")
	defer span.End()

	msg, err := s.sub.Receive(ctx)
	if err != nil {
		// Errors from Receive indicate that Receive will no longer succeed.
		log.Printf("Receiving message: %v", err)
		return err
	}
	// Do work based on the message, for example:
	// fmt.Printf("Got message: %q\n", msg.Body)

	var em *Message

	err = json.Unmarshal(msg.Body, &em)
	if err != nil {
		return err
	}

	s.ProcessMessage(ctx, st, em)

	// Messages must always be acknowledged with Ack.
	msg.Ack()

	return nil
}

func (s *Subscription) ProcessMessage(ctx context.Context, db *query.Stores, msg *Message) error {
	switch {
	case strings.HasSuffix(msg.EventType, ".added"):
		resource, err := query.NewResourceFromURN(msg.SubjectURN)
		if err != nil {
			fmt.Println("Error getting resource from URN for subject")
			return err
		}

		resource.Fields = msg.SubjectFields

		actor, err := query.NewResourceFromURN(msg.ActorURN)
		if err != nil {
			fmt.Println("Error getting resource from URN for actor")
			return err
		}

		_, err = query.CreateSpiceDBRelationships(ctx, db.SpiceDB, resource, actor)
		if err != nil {
			fmt.Println("Failed to create resource...oh well")
			fmt.Printf("Error: %+v", err)
		} else {
			// fmt.Println("Resource created...")

			// poor sampling. Don't print every single created message, that is so many messages. Instead aim for like 1 out of every 200
			if rand.Intn(10) == 9 {
				s.logger.Infow("created resource",
					"type", resource.ResourceType.Name,
					"id", resource.Fields["id"],
					"name", resource.Fields["name"],
				)

			}
		}
	default:
		fmt.Printf("I don't care about %s\n", msg.EventType)
	}

	return nil
}