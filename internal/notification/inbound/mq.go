package inbound

import (
	"context"
	"slices"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
)

const (
	topicUserRegistration                = entity.UserRegistrationDestination
	consumerUserRegistrationNotification = "auth.user.registration.notification"
)

func RegisterMQConsumer(
	ctx context.Context,
	cfg config.Config,
	routine *goroutine.Manager,
	messenger messaging.Messaging,
	uc uc,
) {
	mqHanlder := &MQHandler{uc: uc}

	enableConsumerNames := cfg.GetArray("modules.notification.consumer_names")

	var consumers = []struct {
		name               string
		topic              string // destination where publisher sent message
		nsqConsumerName    string // for nsq
		natsConsumerName   string // for nats
		kafkaConsumerName  string // for kafka
		pubsubConsumerName string // for google pubusb
		handler            messaging.Handler
	}{
		{
			name:               consumerUserRegistrationNotification,
			topic:              topicUserRegistration,
			nsqConsumerName:    consumerUserRegistrationNotification,
			natsConsumerName:   consumerUserRegistrationNotification,
			kafkaConsumerName:  consumerUserRegistrationNotification,
			pubsubConsumerName: consumerUserRegistrationNotification,
			handler:            mqHanlder.UserRegistrationNotification,
		},
	}

	for _, consumer := range consumers {
		if len(enableConsumerNames) > 0 && slices.Contains(enableConsumerNames, consumer.name) {
			routine.Go(ctx, func(pCtx context.Context) error {
				return messenger.Consume(pCtx,
					consumer.topic,
					consumer.handler,
					messaging.WithChannel(consumer.nsqConsumerName),
					messaging.WithQueueGroup(consumer.natsConsumerName),
					messaging.WithGroup(consumer.kafkaConsumerName),
					messaging.WithSubscription(consumer.pubsubConsumerName),
					messaging.WithAutoAck(true),
					messaging.WithConcurrency(10),
					messaging.WithMaxInFlight(10),
				)
			})
		}
	}
}
