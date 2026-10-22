package client

import (
	"context"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

const (
	seriesEndpoint  = "series"
	matchesEndpoint = "matches"
)

const (
	// maxRateLimitAttempts caps how many times a request is sent while
	// Pandascore keeps answering 429.
	maxRateLimitAttempts = 6
	defaultRetryAfter    = 2 * time.Second
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
	// Prime the is_live flag immediately so the UI doesn't have to wait for
	// the first livesTicker tick. A failure here shouldn't block startup —
	// the ticker will retry every LivesPollInterval.
	if liveErr := client.GetLives(); liveErr != nil {
		client.Logger.Errorf("Initial /lives fetch failed (will retry on ticker): %v", liveErr)
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
	client.Logger.Debug("Making request to " + req.URL.String())
	for attempt := 1; ; attempt++ {
		resp, doErr := client.HTTPClient.Do(req)
		if doErr != nil {
			return nil, doErr
		}
		client.Run++
		if resp.StatusCode != http.StatusTooManyRequests || attempt >= maxRateLimitAttempts {
			return resp, nil
		}
		// Pandascore enforces a short burst limit on top of the hourly quota and
		// answers with Retry-After. Concurrent page fetches all trip it at once,
		// so wait it out (plus jitter to spread the retries) instead of failing.
		_ = resp.Body.Close()
		wait := retryAfter(resp.Header.Get("Retry-After"))
		client.Logger.Warnf("Rate limited on %s, retrying in %s (attempt %d)", req.URL.Path, wait, attempt)
		select {
		case <-client.Ctx.Done():
			return nil, client.Ctx.Err()
		case <-time.After(wait):
		}
	}
}

// retryAfter parses a Retry-After header in seconds and adds up to the same
// amount again as jitter, falling back to defaultRetryAfter.
func retryAfter(header string) time.Duration {
	wait := defaultRetryAfter
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		wait = time.Duration(secs) * time.Second
	}
	return wait + rand.N(wait) //nolint:gosec // jitter only, not security sensitive
}
