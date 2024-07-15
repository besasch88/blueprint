package user

import (
	"time"

	"github.com/besasch88/blueprint/internal/pkg/bperr"
	"github.com/besasch88/blueprint/internal/pkg/bppubsub"
	"github.com/besasch88/blueprint/internal/pkg/bputils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userServiceInterface interface {
	getUserByID(ctx *gin.Context, input getUserInputDto) (userEntity, error)
	createUser(ctx *gin.Context, requesterID uuid.UUID, input createUserInputDto) (userEntity, error)
}

type userService struct {
	storage     *gorm.DB
	pubSubAgent *bppubsub.PubSubAgent
	repository  userRepositoryInterface
}

func newUserService(storage *gorm.DB, pubSubAgent *bppubsub.PubSubAgent, repository userRepositoryInterface) userService {
	return userService{
		storage:     storage,
		pubSubAgent: pubSubAgent,
		repository:  repository,
	}
}

func (s userService) getUserByID(ctx *gin.Context, input getUserInputDto) (userEntity, error) {
	userID := uuid.MustParse(input.id)
	item, err := s.repository.getUserByID(s.storage, userID, false)
	if err != nil {
		return userEntity{}, bperr.ErrGeneric
	}
	if bputils.IsEmpty(item) {
		return userEntity{}, errUserNotFound
	}
	return item, nil
}

func (s userService) createUser(ctx *gin.Context, requesterID uuid.UUID, input createUserInputDto) (userEntity, error) {
	now := time.Now()
	user := userEntity{
		id:        uuid.MustParse(input.ID),
		firstname: input.Firstname,
		lastname:  input.Lastname,
		email:     input.Email,
		createdAt: now,
		updatedAt: now,
		deletedAt: nil,
		createdBy: requesterID,
		updatedBy: requesterID,
		deletedBy: nil,
	}
	errTransaction := s.storage.Transaction(func(tx *gorm.DB) error {
		_, err := s.repository.saveUser(tx, user)
		if err != nil {
			return bperr.ErrGeneric
		}
		return nil
	})
	if errTransaction != nil {
		return userEntity{}, errTransaction
	}
	go s.pubSubAgent.Publish(bppubsub.TopicUserV1, bppubsub.PubSubMessage{
		Context: ctx.Copy(),
		Message: bppubsub.PubSubEvent{
			EventID:   uuid.New(),
			EventTime: time.Now(),
			EventType: bppubsub.UserCreatedEvent,
			EventEntity: bppubsub.UserEventEntity{
				ID:        user.id,
				Firstname: user.firstname,
				Lastname:  user.lastname,
				Email:     user.email,
				CreatedAt: user.createdAt,
				UpdatedAt: user.updatedAt,
				DeletedAt: user.deletedAt,
				CreatedBy: user.createdBy,
				UpdatedBy: user.updatedBy,
				DeletedBy: user.deletedBy,
			},
		},
	})
	return user, nil
}
