package bppubsub

import (
	"time"

	"github.com/google/uuid"
)

/*
PubSubEventType represents an event type that can be published or consumed within the pub-sub system.
Generally, the PubSubEventType is related to an event entity and the possible actions performed on it.
It is preferable to use the past participle to indicate that the event was generated as a result
of an application state change.
*/
type PubSubEventType string

/*
List of avaiable events can be published and consumed within the pub-sub system.
*/
const (
	UserCreatedEvent PubSubEventType = "user.created"
	UserUpdatedEvent PubSubEventType = "user.updated"
	UserDeletedEvent PubSubEventType = "user.deleted"
)

/*
PubSubEvent represents a generic struct for events. All the events must be structured in this way,
ensuring the payload of the event itself is stored inside the EventEntity.
*/
type PubSubEvent struct {
	EventID     uuid.UUID
	EventTime   time.Time
	EventType   PubSubEventType
	EventEntity interface{}
}
