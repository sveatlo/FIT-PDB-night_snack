package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"time"

	"github.com/moderntv/cadre"
	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/registry"
	"github.com/moderntv/cadre/registry/file"
	"github.com/moderntv/cadre/status"
	"github.com/nats-io/nats.go"
	nats_protobuf "github.com/nats-io/nats.go/encoders/protobuf"
	grpc_zerolog "github.com/rkollar/go-grpc-middleware/logging/zerolog"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	mongo_options "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"gorm.io/gorm/logger"

	"github.com/sveatlo/night_snack/internal/database"
	"github.com/sveatlo/night_snack/internal/restaurant"
	"github.com/sveatlo/night_snack/internal/snacker"
	"github.com/sveatlo/night_snack/internal/snacker/config"
	restaurant_pb "github.com/sveatlo/night_snack/proto/restaurant"
	snacker_pb "github.com/sveatlo/night_snack/proto/snacker"
)

var (
	Version string
)

// @title Snacker HTTP API
// @version 1.0
// @description Snacker is the main API component of NightSnack

func main() {
	var err error

	appCtx, appCtxCancel := context.WithCancel(context.Background())
	defer appCtxCancel()

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	// flags
	var (
		configFilePath string
		printVersion   bool
	)

	// flags parsing
	flag.StringVar(&configFilePath, "config", "snacker.yaml", "path to config file")
	flag.BoolVar(&printVersion, "version", false, "print version and exit")
	flag.Parse()
	if printVersion {
		fmt.Println("Version: ", Version)
		return
	}

	// configuration
	var appConfig config.Config
	appConfig, err = config.NewConfig(configFilePath)
	if err != nil {
		stdlog.Printf("app configuration failed: %s", err)
		return
	}

	// logging
	var log zerolog.Logger
	{
		var level zerolog.Level
		level, err = zerolog.ParseLevel(appConfig.Loglevel)
		if err != nil {
			stdlog.Printf("parsing loglevel failed: %s", err)
			return
		}

		zerolog.SetGlobalLevel(level)
		zerolog.DurationFieldUnit = time.Second
		zerolog.DurationFieldInteger = false
		var writers = []io.Writer{
			zerolog.ConsoleWriter{
				Out:        os.Stderr,
				TimeFormat: time.RFC3339,
			},
		}

		log = zerolog.New(zerolog.MultiLevelWriter(writers...)).With().Timestamp().Logger()
	}

	// metrics
	var metricsRegistry *metrics.Registry
	metricsRegistry, err = metrics.NewRegistry("snacker", nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("cannot create prometheus registry")
		return
	}

	// application status reporting
	appStatus := status.NewStatus(Version)

	// create registry
	var svcRegistry registry.Registry
	switch appConfig.Registry.Type {
	case "file":
		svcRegistry, err = file.NewRegistry(appConfig.Registry.FilePath, file.WithWatch())
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to create registry")
			return
		}
	default:
		err = fmt.Errorf("unknown registry type")
		log.Error().
			Err(err).
			Str("type", appConfig.Registry.Type).
			Msg("failed to create registry resolver")
		return

	}
	// register registry
	resolver.Register(registry.NewResolverBuilder(svcRegistry))

	// database
	db, err := database.NewConnection(appConfig.Database.Host, appConfig.Database.Port, appConfig.Database.Username, metricsRegistry, appStatus, logger.Warn)
	if err != nil {
		err = fmt.Errorf("cannot create database connection: %w", err)
		return
	}

	// mongo
	mongoClient, err := mongo.NewClient(
		mongo_options.Client().ApplyURI(fmt.Sprintf(appConfig.Mongo.URI)),
		mongo_options.Client().SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
	)
	if err != nil {
		log.Error().Err(err).Msg("cannot create mongo client")
		return
	}
	mongoConnectCtx, mongoConnectCtxCancel := context.WithTimeout(appCtx, 10*time.Second)
	defer mongoConnectCtxCancel()
	err = mongoClient.Connect(mongoConnectCtx)
	if err != nil {
		log.Error().Err(err).Msg("mongo client cannot connect")
		return
	}
	defer mongoClient.Disconnect(context.Background())
	mongo := mongoClient.Database("night_snack")

	// nats
	var (
		nc                  *nats.Conn
		nec                 *nats.EncodedConn
		natsComponentStatus *status.ComponentStatus
	)
	{
		natsComponentStatus, err = appStatus.Register("nats")
		if err != nil {
			log.Error().Err(err).Msg("cannot register nats status")
			return
		}

		nc, err = nats.Connect(
			appConfig.NATS.Servers,
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(appConfig.NATS.MaxReconnects),
			nats.ReconnectWait(appConfig.NATS.ReconnectWait),
			nats.DisconnectErrHandler(func(lnc *nats.Conn, err error) {
				natsComponentStatus.SetStatus(status.WARN, fmt.Sprintf("disconnected: err=%v", err))
			}),
			nats.ReconnectHandler(func(lnc *nats.Conn) {
				natsComponentStatus.SetStatus(status.OK, "ok")
			}),
			nats.ClosedHandler(func(lnc *nats.Conn) {
				natsComponentStatus.SetStatus(status.ERROR, "connection closed")
			}),
		)
		if err != nil {
			err = fmt.Errorf("cannot create nats connection: %w", err)
			log.Error().Err(err).Str("servers", appConfig.NATS.Servers).Msg("nats connection failed")
			return
		}

		nec, err = nats.NewEncodedConn(nc, nats_protobuf.PROTOBUF_ENCODER)
		if err != nil {
			log.Error().Err(err).Msg("cannot create encoded connection")
			return
		}

		if nc.Status() != nats.CONNECTED {
			natsComponentStatus.SetStatus(status.WARN, "nats not immediately connected after start")
		} else {
			natsComponentStatus.SetStatus(status.OK, "ok")
		}

		defer nc.Close()
	}

	// services
	snackerService, err := snacker.New(db, metricsRegistry, appStatus, log)
	if err != nil {
		log.Error().Err(err).Msg("cannot create new snacker service")
		return
	}
	defer snackerService.Close()
	snackerRegistrator := func(s *grpc.Server) {
		snacker_pb.RegisterSnackerServer(s, snackerService)
	}

	restaurantCommandService, err := restaurant.NewCommandService(nec, db, mongo, metricsRegistry, appStatus, log)
	if err != nil {
		log.Error().Err(err).Msg("cannot create new restaurant service")
		return
	}
	defer restaurantCommandService.Close()
	restaurantCommandRegistrator := func(s *grpc.Server) {
		restaurant_pb.RegisterCommandServiceServer(s, restaurantCommandService)
	}

	restaurantQueryService, err := restaurant.NewQueryService(nec, mongo, metricsRegistry, appStatus, log)
	if err != nil {
		log.Error().Err(err).Msg("cannot create new restaurant service")
		return
	}
	defer restaurantQueryService.Close()
	restaurantQueryRegistrator := func(s *grpc.Server) {
		restaurant_pb.RegisterQueryServiceServer(s, restaurantQueryService)
	}

	// HTTP gateway
	gw, err := snacker.NewHTTP(snackerService, restaurantCommandService, restaurantQueryService, log)
	if err != nil {
		log.Error().Err(err).Msg("cannot create http gateway")
		return
	}

	// gRPC options
	var logOptions = []grpc_zerolog.Option{
		grpc_zerolog.WithLevels(func(code codes.Code) zerolog.Level {
			switch code {
			case codes.OK, codes.Canceled, codes.ResourceExhausted, codes.Aborted:
				return zerolog.TraceLevel
			case codes.InvalidArgument, codes.PermissionDenied, codes.AlreadyExists, codes.NotFound, codes.Unauthenticated, codes.FailedPrecondition, codes.OutOfRange, codes.DeadlineExceeded, codes.DataLoss:
				return zerolog.WarnLevel
			case codes.Unknown, codes.Unimplemented, codes.Internal, codes.Unavailable:
				return zerolog.ErrorLevel
			default:
				return zerolog.ErrorLevel
			}
		}),
	}
	grpcOptions := []cadre.GRPCOption{
		cadre.WithGRPCListeningAddress(appConfig.ListenAddressGRPC),
		cadre.WithService("snacker.snacker", snackerRegistrator),
		cadre.WithService("snacker.restaurant.command", restaurantCommandRegistrator),
		cadre.WithService("snacker.restaurant.query", restaurantQueryRegistrator),
		cadre.WithLoggingOptions(logOptions),
	}
	if appConfig.ListenAddressChannelz != "" {
		grpcOptions = append(grpcOptions, cadre.WithChannelz(appConfig.ListenAddressChannelz))
	}

	var (
		b *cadre.Builder
		c cadre.Cadre
	)
	b, err = cadre.NewBuilder(
		"snacker",
		cadre.WithContext(appCtx),
		cadre.WithLogger(log),
		cadre.WithStatus(appStatus),
		cadre.WithMetricsRegistry(metricsRegistry),
		cadre.WithMetricsListeningAddress(appConfig.ListenAddressPrometheus),
		cadre.WithGRPC(grpcOptions...),
		cadre.WithHTTP(
			"main_http",
			cadre.WithHTTPListeningAddress(appConfig.ListenAddressHTTP),
			cadre.WithRoutingGroup(gw.GetRoutes()),
		),
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create cadre builder")
		return
	}
	c, err = b.Build()
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to build cadre")
		return
	}

	fmt.Println(`
____ _  _ ____ ____ _  _ ____ ____
[__  |\ | |__| |    |_/  |___ |__/
___] | \| |  | |___ | \_ |___ |  \

Running...`)

	err = c.Start()
	log.Error().
		Err(err).
		Msg("cadre failed")
}
