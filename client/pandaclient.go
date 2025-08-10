package client

import (
	"context"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"github.com/feimaomiao/stalka/dbtypes"
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
	BaseURL     string
	Pandasecret string
	Logger      *zap.SugaredLogger
	HTTPClient  *http.Client
	DBConnector *dbtypes.Queries
	Run         int
	Ctx         context.Context
}

// Startup performs the initial setup for the PandaClient, which includes
// updating games, leagues, series, tournaments, and matches.
// @returns an error if any of the requests fail.
func (client *PandaClient) Startup() error {
	err := client.UpdateGames()
	if err != nil {
		return err
	}
	err = client.GetLeagues(true)
	if err != nil {
		return err
	}
	err = client.GetSeries(true)
	if err != nil {
		return err
	}
	err = client.GetTournaments(true)
	if err != nil {
		return err
	}
	err = client.GetTeams(true)
	if err != nil {
		return err
	}
	err = client.GetMatches(true)
	if err != nil {
		return err
	}
	client.Logger.Infof("Done with initial setup, made %d requests", client.Run)
	return nil
}

// MakeRequest creates a new HTTP request to the Pandascore API.
// @param paths - the paths to append to the base URL
// @param params - the query parameters to add to the request
// @returns the HTTP response and an error if one occurred.
func (client *PandaClient) MakeRequest(paths []string, params map[string]string) (*http.Response, error) {
	searchurl, err := url.Parse(client.BaseURL)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		searchurl.Path += path + "/"
	}

	req, err := http.NewRequestWithContext(client.Ctx, http.MethodGet, searchurl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+client.Pandasecret)
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	q.Set("per_page", "100")
	req.URL.RawQuery = q.Encode()
	client.Logger.Info("Making request to " + req.URL.String())
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	client.Run++
	return resp, nil
}
