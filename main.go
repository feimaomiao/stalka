package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/feimaomiao/stalka/client"
	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	_ "embed"
)

//go:embed static/schema.sql
var schema string

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
	log.Info("Connecting to database with connection string: ", connStr)
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

func main() {
	ctx := context.Background()
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
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
		HTTPClient:  &http.Client{},
		DBConnector: database.DBConn,
		Run:         0,
		Ctx:         ctx,
	}
	if err != nil {
		sugar.Fatal(err)
	}
	err = client.Startup()
	if err != nil {
		sugar.Fatal(err)
	}
	day := 24
	matchTicker := time.NewTicker(time.Hour)
	setupTicker := time.NewTicker(time.Duration(day) * time.Hour)
	defer matchTicker.Stop()
	defer setupTicker.Stop()
	go func() {
		for range matchTicker.C {
			sugar.Info("Matchticker fired")
			err = client.GetMatches(false)
			if err != nil {
				sugar.Fatal(err)
			}
			sugar.Infof("Done with run, made %d requests so far", client.Run)
		}
	}()
	go func() {
		for range setupTicker.C {
			sugar.Info("Setupticker fired")
			err = client.UpdateGames()
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetLeagues(false)
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetSeries(false)
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetTeams(false)
			if err != nil {
				sugar.Fatal(err)
			}
			err = client.GetTournaments(false)
			if err != nil {
				sugar.Fatal(err)
			}
			sugar.Infof("Done with setup, made %d requests so far", client.Run)
		}
	}()
	for {
		time.Sleep(time.Hour)
	}
}
