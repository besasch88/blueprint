package user

import (
	"github.com/besasch88/blueprint/internal/pkg/bppubsub"
	"github.com/besasch88/blueprint/internal/pkg/bputils"
	"go.uber.org/zap"
)

type userConsumerInterface interface {
	subscribe()
}

type userConsumer struct {
	pubSub  *bppubsub.PubSubAgent
	service userServiceInterface
}

func newUserConsumer(pubSub *bppubsub.PubSubAgent, service userServiceInterface) userConsumer {
	consumer := userConsumer{
		pubSub:  pubSub,
		service: service,
	}
	return consumer
}

func (r userConsumer) subscribe() {
	go func() {
		messageChannel := r.pubSub.Subscribe(bppubsub.TopicUserV1)
		isChannelOpen := true
		for isChannelOpen {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.L().Error("Panic occured in handling a new message", zap.String("service", "user-consumer"))
					}
				}()
				msg, channelOpen := <-messageChannel
				if !channelOpen {
					isChannelOpen = false
					zap.L().Info(
						"Channel closed. No more events to listen... quit!",
						zap.String("service", "user-consumer"),
					)
					return
				}
				zap.L().Info(
					"Received Event Message",
					zap.String("service", "user-consumer"),
					zap.String("event-id", msg.Message.EventID.String()),
					zap.String("event-type", string(msg.Message.EventType)),
				)
				if msg.Message.EventType != bppubsub.UserCreatedEvent {
					return
				}

				event := msg.Message.EventEntity.(bppubsub.UserEventEntity)
				userID := event.ID
				input := createUserInputDto{
					ID:        bputils.GetStringFromUUID(event.ID),
					Firstname: event.Firstname,
					Lastname:  event.Lastname,
					Email:     event.Email,
				}
				_, err := r.service.createUser(msg.Context, userID, input)
				if err != nil {
					zap.L().Error("Impossible to create a new user", zap.String("service", "user-consumer"))
					return
				}
			}()
		}
	}()
}
