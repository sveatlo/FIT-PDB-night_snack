package database

import (
	"fmt"
	"time"

	"github.com/moderntv/cadre/metrics"
	"github.com/moderntv/cadre/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	checkQuery           = "SELECT 1;"
	statusCheckTimeLimit = 3 * time.Second
	statusCheckInterval  = 3 * time.Second
)

// NewConnection function establishes connection to a mysql database.
func NewConnection(
	host string,
	port int,
	user string,
	registry *metrics.Registry,
	appStatus *status.Status,
	loglevel logger.LogLevel,
) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("postgresql://%s@%s:%d?sslmode=disable", user, host, port)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(loglevel),
	})
	if err != nil {
		return
	}

	if appStatus != nil {
		dbStatus, err := appStatus.Register("db")
		if err != nil {
			return nil, err
		}
		dbStatus.SetStatus(status.OK, "hopefully OK - not checking actively")
	}

	// dbName := "main"
	//
	// gormPrometheus := prometheus.New(prometheus.Config{
	//     DBName:          dbName,
	//     RefreshInterval: 15,
	//     MetricsCollector: []prometheus.MetricsCollector{
	//         &prometheus.MySQL{
	//             VariableNames: []string{"Threads_running"},
	//         },
	//     },
	// })
	// err = db.Use(gormPrometheus)
	// if err != nil {
	//     return
	// }
	//
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_max_open", dbName), gormPrometheus.MaxOpenConnections)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_open", dbName), gormPrometheus.OpenConnections)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_in_use", dbName), gormPrometheus.InUse)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_idle", dbName), gormPrometheus.Idle)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_waited_for", dbName), gormPrometheus.WaitCount)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_wait_duration", dbName), gormPrometheus.WaitDuration)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_max_idle_closed", dbName), gormPrometheus.MaxIdleClosed)
	// if err != nil {
	//     return
	// }
	// err = registry.Register(fmt.Sprintf("gorm_%s_connections_max_lifetime_closed", dbName), gormPrometheus.MaxLifetimeClosed)
	// if err != nil {
	//     return
	// }
	//
	return
}
