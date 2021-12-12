package snacker

import (
	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	snacker_pb "github.com/sveatlo/night_snack/proto/snacker"
)

type SnackerSvc struct {
	snacker_pb.UnimplementedSnackerServer

	log    zerolog.Logger
	status *status.ComponentStatus
}

func New(db *gorm.DB, metricsRegistry *metrics.Registry, appStatus *status.Status, log zerolog.Logger) (c *SnackerSvc, err error) {
	cs, err := appStatus.Register("snacker")
	if err != nil {
		return
	}

	c = &SnackerSvc{
		log:    log.With().Str("component", "snacker").Logger(),
		status: cs,
	}

	return
}

func (s *SnackerSvc) Close() {
}
