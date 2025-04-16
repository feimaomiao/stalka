package client

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"os"

	"github.com/feimaomiao/stalka/database"
	"go.uber.org/zap"
)

type GetChoice int

const (
	FlagGame GetChoice = iota
	FlagLeague
	FlagSeries
	FlagTournament
	FlagMatch
	FlagTeam
)

// PandaClient is a client for the Pandascore API.
type PandaClient struct {
	pandasecret string
	logger      *zap.SugaredLogger
	httpClient  *http.Client
	dbConnector *sql.DB
	run         int
	ctx         context.Context
}

// GetRun is a wrapper for the run variable, which counts the number of requests made.
func (client *PandaClient) GetRun() int {
	return client.run
}

// NewPandaClient creates a new PandaClient.
// @param logger - the logger to use
// @returns a new PandaClient and an error if one occurred.
func NewPandaClient(logger *zap.SugaredLogger) (PandaClient, error) {
	dbConnector, err := database.Connect("writer", os.Getenv("writer_password"))
	if err != nil {
		return PandaClient{}, err
	}
	return PandaClient{
		pandasecret: os.Getenv("pandascore_secret"),
		logger:      logger,
		httpClient:  &http.Client{},
		dbConnector: dbConnector,
		run:         0,
		ctx:         context.Background(),
	}, nil
}

// Startup performs the initial setup for the PandaClient, which includes
// updating games, leagues, series, tournaments, and matches.
// @returns an error if any of the requests fail.
func (client *PandaClient) Startup() error {
	err := client.UpdateGames()
	if err != nil {
		return err
	}
	err = client.GetLeagues()
	if err != nil {
		return err
	}
	err = client.GetSeries()
	if err != nil {
		return err
	}
	err = client.GetTournaments()
	if err != nil {
		return err
	}
	err = client.GetMatches()
	if err != nil {
		return err
	}
	client.logger.Infof("Done with initial setup, made %d requests", client.GetRun())
	return nil
}

// MakeRequest creates a new HTTP request to the Pandascore API.
// @param paths - the paths to append to the base URL
// @param params - the query parameters to add to the request
// @returns the HTTP response and an error if one occurred.
func (client *PandaClient) MakeRequest(paths []string, params map[string]string) (*http.Response, error) {
	searchurl, err := url.Parse("https://api.pandascore.co/")
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		searchurl.Path += path + "/"
	}

	req, err := http.NewRequestWithContext(client.ctx, http.MethodGet, searchurl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.pandasecret)
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	q.Set("per_page", "100")
	req.URL.RawQuery = q.Encode()
	client.logger.Info("Making request to " + req.URL.String())
	resp, err := client.httpClient.Do(req)
	client.run++
	if err != nil {
		return nil, err
	}
	return resp, nil
}
