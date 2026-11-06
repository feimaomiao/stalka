package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/feimaomiao/stalka/client"
	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	_ "embed"
)

//go:embed static/schema.sql
var schema string

const (
	LivesPollInterval = 5 * time.Minute
	// requestTimeout bounds a whole request. Without it a server that accepts a
	// connection and never answers blocks its ticker forever, with no error.
	requestTimeout = 30 * time.Second
	hoursPerDay    = 24
	// The hourly and daily runs start offset from the lives poll. Their periods are
	// multiples of its period, so otherwise all three would fire in the same
	// instant, and Pandascore's burst limit answers that with 429s.
	matchTickerOffset = 2 * time.Minute
	setupTickerOffset = 7 * time.Minute
)

// DatabaseConnector is a struct that holds the database connection and the dbtypes.Queries object.
// It is used to interact with the database.
// @param Db - the database connection.
// @param DbConn - the queries object interacting with the database (sqlc).
type DatabaseConnector struct {
	DB     *pgxpool.Pool
	DBConn *dbtypes.Queries
}

// Init initializes the database connection and returns a DatabaseConnector.
// It reads the connection parameters from environment variables.
// @param ctx - the context to use for the database connection.
// @param log - the logger to use for logging.
// @returns a DatabaseConnector and an error if one occurred.
func Init(ctx context.Context, log *zap.SugaredLogger) (DatabaseConnector, error) {
	connStr := fmt.Sprintf("host=postgres port=5432 user=%s password=%s dbname=esports sslmode=disable",
		os.Getenv("postgres_user"),
		os.Getenv("postgres_password"))
	log.Info("Connecting to database")
	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Error(err, "Failed to connect to database")
		return DatabaseConnector{}, err
	}
	// Ping the database to ensure the connection is established.
	err = db.Ping(context.Background())
	if err != nil {
		log.Error(err)
		return DatabaseConnector{}, err
	}
	log.Info("Connected to database, running migrations")
	_, err = db.Exec(ctx, schema)
	if err != nil {
		log.Error(err, "Failed to run migrations")
		return DatabaseConnector{}, err
	}
	log.Info("Migrations completed successfully")
	// Create a dbtypes.Queries object to interact with the database.
	dbConn := dbtypes.New(db)
	return DatabaseConnector{
		DB:     db,
		DBConn: dbConn,
	}, nil
}

// refresh re-fetches the slow-moving entities. Each ticker goroutine keeps its
// own error: sharing one would let a tolerated /lives failure trip a Fatal here.
// every runs fn each period, starting one period after offset. Runs never
// overlap each other: a tick that lands during a run is dropped.
func every(period, offset time.Duration, fn func()) {
	go func() {
		time.Sleep(offset)
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for range ticker.C {
			fn()
		}
	}()
}

func refresh(panda *client.PandaClient) error {
	steps := []func() error{
		panda.UpdateGames,
		func() error { return panda.GetLeagues(false) },
		func() error { return panda.GetSeries(false) },
		func() error { return panda.GetTeams(false) },
		func() error { return panda.GetTournaments(false) },
	}
	var errs []error
	for _, step := range steps {
		if err := step(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func main() {
	ctx := context.Background()
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}
	logger, _ := config.Build()
	sugar := logger.Sugar()
	database, err := Init(ctx, sugar)
	if err != nil {
		sugar.Fatal(err)
	}
	defer database.DB.Close()

	// Initialize the PandaClient with the database connector and logger.
	// The PandaClient will be used to make requests to the Pandascore API.
	client := client.PandaClient{
		BaseURL:     "https://api.pandascore.co/",
		Pandasecret: os.Getenv("pandascore_secret"),
		Logger:      sugar,
		HTTPClient:  &http.Client{Timeout: requestTimeout},
		DBConnector: database.DBConn,
		Run:         atomic.Int64{},
		Ctx:         ctx,
	}
	// Failures past this point are logged, not fatal: a restart would repeat every
	// request Startup made and fix nothing, while the tickers keep data moving.
	err = client.Startup()
	if err != nil {
		sugar.Errorf("Startup incomplete, the next daily run will repeat it: %v", err)
	}
	needsStartup := err != nil
	every(LivesPollInterval, 0, func() {
		sugar.Info("Lives ticker fired")
		if liveErr := client.GetLives(); liveErr != nil {
			sugar.Error(liveErr)
		}
		sugar.Infof("Done with lives update, made %d requests so far", client.Run.Load())
	})
	every(time.Hour, matchTickerOffset, func() {
		sugar.Info("Matchticker fired")
		if matchErr := client.GetMatches(false); matchErr != nil {
			sugar.Error(matchErr)
		}
		if reconcileErr := client.ReconcileMatches(); reconcileErr != nil {
			sugar.Error(reconcileErr)
		}
		sugar.Infof("Done with run, made %d requests so far", client.Run.Load())
	})
	every(hoursPerDay*time.Hour, setupTickerOffset, func() {
		sugar.Info("Setupticker fired")
		var setupErr error
		if needsStartup {
			setupErr = client.Startup()
			needsStartup = setupErr != nil
		} else {
			setupErr = refresh(&client)
		}
		if setupErr != nil {
			sugar.Error(setupErr)
		}
		sugar.Infof("Done with setup, made %d requests so far", client.Run.Load())
	})
	for {
		time.Sleep(time.Hour)
	}
}
