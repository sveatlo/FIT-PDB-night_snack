package orders

import (
	"context"
	"fmt"

	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/sveatlo/night_snack/internal/restaurant"
	"github.com/sveatlo/night_snack/internal/stock"
	orders_pb "github.com/sveatlo/night_snack/proto/orders"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
	stock_pb "github.com/sveatlo/night_snack/proto/stock"
)

type Service struct {
	orders_pb.UnimplementedOrdersServiceServer

	log    zerolog.Logger
	status *status.ComponentStatus

	restaurantQueryService *restaurant.QueryService
	stockService           *stock.Service
	repo                   *Repository
}

func NewService(restaurantQueryService *restaurant.QueryService, stockService *stock.Service, nec *nats.EncodedConn, mongo *mongo.Database, metricsRegistry *metrics.Registry, appStatus *status.Status, log zerolog.Logger) (c *Service, err error) {
	cs, err := appStatus.Register("order/svc")
	if err != nil {
		return
	}

	repo, err := NewRepository(nec, mongo, log)
	if err != nil {
		err = fmt.Errorf("cannot create order repository: %w", err)
	}

	c = &Service{
		log:    log.With().Str("component", "order/svc").Logger(),
		status: cs,

		restaurantQueryService: restaurantQueryService,
		stockService:           stockService,
		repo:                   repo,
	}

	return
}

func (s *Service) Close() {}

func (s *Service) Create(ctx context.Context, cmd *orders_pb.CmdCreateOrder) (event *orders_pb.OrderCreated, err error) {
	// try inventory
	reservedItemsIDs := []string{}
	allAvailable := true
	for _, itemID := range cmd.ItemIds {
		_, err = s.stockService.DecreaseStock(ctx, &stock_pb.CmdDecreaseStock{
			ItemId: itemID,
			N:      1,
		})
		if err != nil {
			allAvailable = false
			break
		}

		reservedItemsIDs = append(reservedItemsIDs, itemID)
	}
	if !allAvailable {
		err = s.releaseReservedItems(ctx, reservedItemsIDs...)
		if err != nil {
			return
		}
		err = fmt.Errorf("cannot create order: some items are not in stock")
		return
	}

	var (
		r     = &restaurant.Restaurant{}
		items = []*restaurant.MenuItem{}
	)
	res, err := s.restaurantQueryService.Get(ctx, &restaurant_pb.GetRestaurant{
		Id: cmd.RestaurantId,
	})
	if err != nil {
		errRelease := s.releaseReservedItems(ctx, reservedItemsIDs...)
		if errRelease != nil {
			err = fmt.Errorf("cannot release items after query failed: %w", err)
		}
		return
	}

	r = restaurant.NewRestaurantFromProto(res)
	r.MenuCategories = nil

	for _, category := range res.Categories {
		for _, item := range category.Items {
			items = append(items, restaurant.NewMenuItemFromProto(item))
		}
	}

	e, err := s.repo.CreateOrder(ctx, r, items)
	if err != nil {
		errRelease := s.releaseReservedItems(ctx, reservedItemsIDs...)
		if errRelease != nil {
			err = fmt.Errorf("cannot release items after query failed: %w", err)
		}
		return
	}

	event = e.ToProto().(*orders_pb.OrderCreated)

	return
}

func (s *Service) UpdateStatus(ctx context.Context, cmd *orders_pb.CmdUpdateStatus) (res *orders_pb.StatusUpdated, err error) {
	event, err := s.repo.UpdateStatus(ctx, cmd.GetId(), cmd.GetStatus().Enum())
	if err != nil {
		err = fmt.Errorf("incrase failed: %w", err)
		return
	}

	res = event.ToProto().(*orders_pb.StatusUpdated)

	return
}

func (s *Service) releaseReservedItems(ctx context.Context, reservedItemsIDs ...string) (err error) {
	for _, itemID := range reservedItemsIDs {
		_, err = s.stockService.IncreaseStock(ctx, &stock_pb.CmdIncreaseStock{
			ItemId: itemID,
			N:      1,
		})
		if err != nil {
			err = fmt.Errorf("fatal error: %v", err)
			return
		}
	}

	return
}
