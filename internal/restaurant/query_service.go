package restaurant

import (
	"context"
	"fmt"

	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
)

type QueryService struct {
	restaurant_pb.UnimplementedQueryServiceServer

	log    zerolog.Logger
	status *status.ComponentStatus

	repo *ReadRepository
}

func NewQueryService(nec *nats.EncodedConn, mongoDB *mongo.Database, metricsRegistry *metrics.Registry, appStatus *status.Status, log zerolog.Logger) (c *QueryService, err error) {
	cs, err := appStatus.Register("restaurant/query_svc")
	if err != nil {
		return
	}

	repo, err := NewReadRepository(nec, mongoDB, log)
	if err != nil {
		err = fmt.Errorf("cannot create restaurant repository: %w", err)
	}

	c = &QueryService{
		log:    log.With().Str("component", "restaurant/svc").Logger(),
		status: cs,

		repo: repo,
	}

	return
}

func (s *QueryService) Close() {}

func (s *QueryService) Get(ctx context.Context, cmd *restaurant_pb.GetRestaurant) (res *restaurant_pb.Restaurant, err error) {
	restaurant, err := s.repo.Get(ctx, cmd.GetId())
	if err != nil {
		err = fmt.Errorf("restaurants query failed: %w", err)
		return
	}

	res = restaurant.ToProto()

	return
}

func (s *QueryService) GetAll(ctx context.Context, cmd *restaurant_pb.GetRestaurants) (res *restaurant_pb.Restaurants, err error) {
	restaurants, err := s.repo.GetAll(ctx)
	if err != nil {
		err = fmt.Errorf("restaurants query failed: %w", err)
		return
	}

	res = &restaurant_pb.Restaurants{
		Restaurants: []*restaurant_pb.Restaurant{},
	}
	for _, restaurant := range restaurants {
		res.Restaurants = append(res.Restaurants, restaurant.ToProto())
	}

	return
}
