package inbound

import (
	"context"

	"github.com/shandysiswandi/gobite/internal/notification/entity"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgmessaging"
	"github.com/shandysiswandi/gobite/internal/pkg/pkgroutine"
)

const (
	topicUserRegistration                = entity.UserRegistrationDestination
	consumerUserRegistrationNotification = "auth.user.registration.notification"
)

func RegisterMQConsumer(ctx context.Context, routine *pkgroutine.Manager, messaging pkgmessaging.Messaging, uc uc) {
	mqHanlder := &MQHandler{uc: uc}

	var consumers = []struct {
		topic              string // destination where publisher sent message
		nsqConsumerName    string // for nsq
		natsConsumerName   string // for nats
		kafkaConsumerName  string // for kafka
		pubsubConsumerName string // for google pubusb
		handler            pkgmessaging.Handler
	}{
		{
			topic:              topicUserRegistration,
			nsqConsumerName:    consumerUserRegistrationNotification,
			natsConsumerName:   consumerUserRegistrationNotification,
			kafkaConsumerName:  consumerUserRegistrationNotification,
			pubsubConsumerName: consumerUserRegistrationNotification,
			handler:            mqHanlder.UserRegistrationNotification,
		},
	}

	for _, consumer := range consumers {
		routine.Go(ctx, func(pCtx context.Context) error {
			return messaging.Consume(pCtx,
				consumer.topic,
				consumer.handler,
				pkgmessaging.WithChannel(consumer.nsqConsumerName),
				pkgmessaging.WithQueueGroup(consumer.natsConsumerName),
				pkgmessaging.WithGroup(consumer.kafkaConsumerName),
				pkgmessaging.WithSubscription(consumer.pubsubConsumerName),
				pkgmessaging.WithAutoAck(true),
				pkgmessaging.WithConcurrency(10),
				pkgmessaging.WithMaxInFlight(10),
			)
		})
	}
}
