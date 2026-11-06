package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
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
	// maxRequestAttempts caps how many times a request is sent while Pandascore
	// answers 429 or 5xx.
	maxRequestAttempts = 6
	defaultRetryAfter  = 2 * time.Second
	maxRetryAfter      = time.Minute
	perPage            = 100
)

// PandaClient is a client for the Pandascore API.
type PandaClient struct {
	BaseURL     string
	Pandasecret string
	Logger      *zap.SugaredLogger
	HTTPClient  *http.Client
	DBConnector *dbtypes.Queries
	Run         atomic.Int64
	Ctx         context.Context
}

// Startup performs the initial setup for the PandaClient, which includes
// updating games, leagues, series, tournaments, and matches.
// @returns an error if any of the requests fail.
func (client *PandaClient) Startup() error {
	steps := []func() error{
		client.UpdateGames,
		func() error { return client.GetLeagues(true) },
		func() error { return client.GetSeries(true) },
		func() error { return client.GetTournaments(true) },
		func() error { return client.GetTeams(true) },
		func() error { return client.GetMatches(true) },
	}
	// Run every step even if one fails: each is useful alone, and missing
	// parents get fetched on demand.
	var errs []error
	for _, step := range steps {
		if err := step(); err != nil {
			client.Logger.Errorf("Startup step failed: %v", err)
			errs = append(errs, err)
		}
	}
	// Prime the is_live flag immediately so the UI doesn't have to wait for
	// the first livesTicker tick. A failure here shouldn't block startup —
	// the ticker will retry every LivesPollInterval.
	if liveErr := client.GetLives(); liveErr != nil {
		client.Logger.Errorf("Initial /lives fetch failed (will retry on ticker): %v", liveErr)
	}
	client.Logger.Infof("Done with initial setup, made %d requests", client.Run.Load())
	return errors.Join(errs...)
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
	q.Set("per_page", strconv.Itoa(perPage))
	req.URL.RawQuery = q.Encode()
	client.Logger.Debug("Making request to " + req.URL.String())
	for attempt := 1; ; attempt++ {
		resp, doErr := client.HTTPClient.Do(req)
		if doErr != nil {
			return nil, doErr
		}
		client.Run.Add(1)
		if !isRetryable(resp.StatusCode) || attempt >= maxRequestAttempts {
			return resp, nil
		}
		// Pandascore enforces a short burst limit on top of the hourly quota and
		// answers with Retry-After. Concurrent page fetches all trip it at once,
		// so wait it out (plus jitter to spread the retries) instead of failing.
		// 5xx responses are retried the same way, with the default wait.
		_ = resp.Body.Close()
		wait := retryAfter(resp.Header.Get("Retry-After"))
		client.Logger.Warnf("Got %d on %s, retrying in %s (attempt %d)", resp.StatusCode, req.URL.Path, wait, attempt)
		select {
		case <-client.Ctx.Done():
			return nil, client.Ctx.Err()
		case <-time.After(wait):
		}
	}
}

// isRetryable reports whether a response is worth sending the request again for.
func isRetryable(code int) bool {
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

// StatusError is a non-200 response from Pandascore, after any retries.
type StatusError struct {
	Path string
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("pandascore %s: status %d", e.Path, e.Code)
}

// getBody makes a request and returns the body of a 200 response. Anything else
// is an error, so a failed fetch can't pass for an empty page.
func (client *PandaClient) getBody(paths []string, params map[string]string) ([]byte, error) {
	resp, err := client.MakeRequest(paths, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &StatusError{Path: strings.Join(paths, "/"), Code: resp.StatusCode}
	}
	return io.ReadAll(resp.Body)
}

// retryAfter parses a Retry-After header in seconds, capped at maxRetryAfter so a
// large value can't stall polling, and adds up to the same amount again as
// jitter, falling back to defaultRetryAfter.
func retryAfter(header string) time.Duration {
	wait := defaultRetryAfter
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		wait = min(time.Duration(secs)*time.Second, maxRetryAfter)
	}
	return wait + rand.N(wait) //nolint:gosec // jitter only, not security sensitive
}
