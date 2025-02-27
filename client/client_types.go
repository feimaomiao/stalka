package client

import (
	"database/sql"
	"net/http"
	"net/url"
	"os"

	"github.com/feimaomiao/stalka/database"
	"go.uber.org/zap"
)

// / PandaClient is a client for the Pandascore API
type PandaClient struct {
	pandasecret string
	logger      *zap.SugaredLogger
	httpClient  *http.Client
	dbConnector *sql.DB
	run         int
}

func (client *PandaClient) GetRun() int {
	return client.run
}

// / NewPandaClient creates a new PandaClient
// / @return the PandaClient
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
	}, nil
}

// / MakeRequest makes a request to the Pandascore API
// / @param paths the paths to append to the base url
// / @param params the parameters to add to the url
// / @return the response
func (client *PandaClient) MakeRequest(paths []string, params map[string]string) (*http.Response, error) {
	searchurl, err := url.Parse("https://api.pandascore.co/")
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		searchurl.Path += path + "/"
	}

	req, err := http.NewRequest("GET", searchurl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", "Bearer "+client.pandasecret)
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	q.Set("per_page", "100")
	req.URL.RawQuery = q.Encode()
	client.logger.Info("Making request to " + req.URL.String())
	resp, err := client.httpClient.Do(req)
	client.run += 1
	if err != nil {
		return nil, err
	}
	return resp, nil
}
