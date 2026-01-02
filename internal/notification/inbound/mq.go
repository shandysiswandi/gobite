package inbound

import (
	"context"
	"log/slog"
	"slices"

	"github.com/shandysiswandi/gobite/internal/pkg/config"
	"github.com/shandysiswandi/gobite/internal/pkg/goroutine"
	"github.com/shandysiswandi/gobite/internal/pkg/instrument"
	"github.com/shandysiswandi/gobite/internal/pkg/messaging"
	"github.com/shandysiswandi/gobite/internal/pkg/uid"
	"github.com/shandysiswandi/gobite/internal/shared/event"
)

func RegisterMQConsumer(
	ctx context.Context,
	cfg config.Config,
	routine *goroutine.Manager,
	messenger messaging.Messaging,
	uuid uid.StringID,
	uc uc,
	ins instrument.Instrumentation,
) {
	mqHanlder := &MQHandler{uc: uc, uuid: uuid, ins: ins}

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
			name:               event.UserRegistrationDestinationConsumerNotification,
			topic:              event.UserRegistrationDestination,
			nsqConsumerName:    event.UserRegistrationDestinationConsumerNotification,
			natsConsumerName:   event.UserRegistrationDestinationConsumerNotification,
			kafkaConsumerName:  event.UserRegistrationDestinationConsumerNotification,
			pubsubConsumerName: event.UserRegistrationDestinationConsumerNotification,
			handler:            mqHanlder.UserRegistrationNotification,
		},
		{
			name:               event.UserForgotPasswordConsumerNotification,
			topic:              event.UserForgotPasswordDestination,
			nsqConsumerName:    event.UserForgotPasswordConsumerNotification,
			natsConsumerName:   event.UserForgotPasswordConsumerNotification,
			kafkaConsumerName:  event.UserForgotPasswordConsumerNotification,
			pubsubConsumerName: event.UserForgotPasswordConsumerNotification,
			handler:            mqHanlder.UserForgotPasswordNotification,
		},
	}

	for _, consumer := range consumers {
		if len(enableConsumerNames) > 0 && slices.Contains(enableConsumerNames, consumer.name) {
			routine.Go(ctx, func(pCtx context.Context) error {
				slog.InfoContext(ctx, "Running job for handling consumer", "consumer", consumer.name)
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
