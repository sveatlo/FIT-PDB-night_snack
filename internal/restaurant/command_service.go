package restaurant

import (
	"context"
	"fmt"

	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type CommandService struct {
	restaurant_pb.UnimplementedCommandServiceServer

	log    zerolog.Logger
	status *status.ComponentStatus

	repo *WriteRepository
}

func NewCommandService(nec *nats.EncodedConn, db *gorm.DB, mongo *mongo.Database, metricsRegistry *metrics.Registry, appStatus *status.Status, log zerolog.Logger) (c *CommandService, err error) {
	cs, err := appStatus.Register("restaurant/command_svc")
	if err != nil {
		return
	}

	repo, err := NewWriteRepository(nec, db, mongo, log)
	if err != nil {
		err = fmt.Errorf("cannot create restaurant repository: %w", err)
	}

	c = &CommandService{
		log:    log.With().Str("component", "restaurant/svc").Logger(),
		status: cs,

		repo: repo,
	}

	return
}

func (s *CommandService) Close() {}

func (s *CommandService) Create(ctx context.Context, cmd *restaurant_pb.CmdRestaurantCreate) (res *restaurant_pb.RestaurantCreated, err error) {
	event, err := s.repo.Create(ctx, cmd.GetName())
	if err != nil {
		err = fmt.Errorf("creation failed: %w", err)
		return
	}

	res = event.ToProto().(*restaurant_pb.RestaurantCreated)

	return
}

func (s *CommandService) Update(ctx context.Context, cmd *restaurant_pb.CmdRestaurantUpdate) (res *restaurant_pb.RestaurantUpdated, err error) {
	event, err := s.repo.Update(ctx, cmd.GetId(), cmd.GetName())
	if err != nil {
		err = fmt.Errorf("update failed: %w", err)
		return
	}

	res = event.ToProto().(*restaurant_pb.RestaurantUpdated)

	return
}

func (s *CommandService) Delete(ctx context.Context, cmd *restaurant_pb.CmdRestaurantDelete) (res *restaurant_pb.RestaurantDeleted, err error) {
	event, err := s.repo.Delete(ctx, cmd.GetId())
	if err != nil {
		err = fmt.Errorf("deletion failed: %w", err)
		return
	}

	res = event.ToProto().(*restaurant_pb.RestaurantDeleted)

	return
}
