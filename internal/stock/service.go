package stock

import (
	"context"
	"fmt"

	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	stock_pb "github.com/sveatlo/night_snack/proto/stock"
)

type Service struct {
	stock_pb.UnimplementedStockServiceServer

	log    zerolog.Logger
	status *status.ComponentStatus

	repo *Repository
}

func NewService(nec *nats.EncodedConn, mongo *mongo.Database, metricsRegistry *metrics.Registry, appStatus *status.Status, log zerolog.Logger) (c *Service, err error) {
	cs, err := appStatus.Register("stock/command_svc")
	if err != nil {
		return
	}

	repo, err := NewRepository(nec, mongo, log)
	if err != nil {
		err = fmt.Errorf("cannot create stock repository: %w", err)
	}

	c = &Service{
		log:    log.With().Str("component", "stock/svc").Logger(),
		status: cs,

		repo: repo,
	}

	return
}

func (s *Service) Close() {}

func (s *Service) IncreaseStock(ctx context.Context, cmd *stock_pb.CmdIncreaseStock) (res *stock_pb.StockIncreased, err error) {
	event, err := s.repo.IncreaseStock(ctx, cmd.GetItemId(), cmd.GetN())
	if err != nil {
		err = fmt.Errorf("incrase failed: %w", err)
		return
	}

	res = event.ToProto().(*stock_pb.StockIncreased)

	return
}

func (s *Service) DecreaseStock(ctx context.Context, cmd *stock_pb.CmdDecreaseStock) (res *stock_pb.StockDecreased, err error) {
	event, err := s.repo.DecreaseStock(ctx, cmd.GetItemId(), cmd.GetN())
	s.log.Debug().Interface("event", event).Err(err).Msg("check")
	if err != nil {
		err = fmt.Errorf("decrease failed: %w", err)
		return
	}

	res = event.ToProto().(*stock_pb.StockDecreased)

	return
}
