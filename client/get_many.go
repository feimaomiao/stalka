package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"encoding/json"

	"github.com/feimaomiao/stalka/pandatypes"

	// loads .env file automatically.
	_ "github.com/joho/godotenv/autoload"
)

const (
	sortedBy   = "-modified_at"
	Pages      = 20
	SetupPages = 50
	// firstPage is where Pandascore pagination starts; page=0 is an empty list.
	firstPage = 1
	// ReconcileBatch caps the matches ReconcileMatches refetches per run. Each
	// costs one request against the 1k/hour quota.
	ReconcileBatch       = 100
	maxReconcileFailures = 3
)

// UpdateGames updates all games in the database.
// @returns an error if one occurred.
func (client *PandaClient) UpdateGames() error {
	client.Logger.Info("Updating games")
	body, err := client.getBody([]string{"videogames"}, nil)
	if err != nil {
		client.Logger.Errorf("Error fetching from Pandascore: %v", err)
		return err
	}

	var result pandatypes.GameLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.Logger.Errorf("Error unmarshalling response: %v", err)
		return err
	}
	for _, game := range result {
		client.Logger.Debugf("Writing game %s", game.Name)
		err = game.ToRow().WriteToDB(client.Ctx, client.DBConnector)
		if err != nil {
			client.Logger.Errorf("Error writing game %s to database: %v", game.Name, err)
			continue
		}
	}
	return nil
}

// GetLeagues gets the first leagues from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetLeagues(setup bool) error {
	client.Logger.Info("Getting leagues")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range Pages {
		keys["page"] = strconv.Itoa(i + firstPage)
		body, err := client.getBody([]string{"leagues"}, keys)
		if err != nil {
			client.Logger.Errorf("Error fetching from Pandascore: %v", err)
			return err
		}

		var result pandatypes.LeagueLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.Logger.Errorf("Error unmarshalling response: %v", err)
			return err
		}
		for _, league := range result {
			client.Logger.Debugf("Writing league %s", league.Name)
			err = league.ToRow().WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Errorf("Error writing league %s to database: %v", league.Name, err)
				continue
			}
		}
		if !setup {
			break
		}
	}
	return nil
}

// GetSeries gets the first 200 series from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetSeries(setup bool) error {
	client.Logger.Info("Getting series")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range Pages {
		client.Logger.Debugf("Getting series page %d", i)
		keys["page"] = strconv.Itoa(i + firstPage)
		body, err := client.getBody([]string{seriesEndpoint}, keys)
		if err != nil {
			client.Logger.Errorf("Error fetching from Pandascore: %v", err)
			return err
		}

		var result pandatypes.SeriesLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.Logger.Errorf("Error unmarshalling response: %v", err)
			return err
		}
		var exists bool
		for _, series := range result {
			exists, err = client.ExistCheck(series.League.ID, FlagLeague)
			if err != nil {
				continue
			}
			if !exists {
				err = client.GetOne(series.League.ID, FlagLeague)
				if err != nil {
					continue
				}
			}
			client.Logger.Debugf("Writing series %s, with league_id %d", series.Name, series.LeagueID)
			err = series.ToRow().WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Errorf("Error writing series %s to database: %v", series.Name, err)
				continue
			}
		}
		if !setup {
			break
		}
	}
	return nil
}

// GetTournaments gets the first 1000 tournaments from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetTournaments(setup bool) error {
	client.Logger.Info("Getting tournaments")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range Pages {
		keys["page"] = strconv.Itoa(i + firstPage)
		client.Logger.Debugf("Getting tournaments page %d", i)
		body, err := client.getBody([]string{"tournaments"}, keys)
		if err != nil {
			client.Logger.Errorf("Error fetching from Pandascore: %v", err)
			return err
		}

		var result pandatypes.TournamentLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		var exists bool
		for _, tournament := range result {
			exists, err = client.ExistCheck(tournament.SerieID, FlagSeries)
			if err != nil {
				continue
			}
			if !exists {
				client.Logger.Debugf("Serie %d does not exist, getting", tournament.SerieID)
				err = client.GetOne(tournament.SerieID, FlagSeries)
				if err != nil {
					continue
				}
			}
			client.Logger.Debugf("Writing tournament %s", tournament.Name)
			err = tournament.ToRow().WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Errorf("Error writing tournament %s to database: %v", tournament.Name, err)
				continue
			}
		}
		if !setup {
			break
		}
	}
	return nil
}

// Goroutine to get one page of matches. Sends the data to the channel.
func (client *PandaClient) getMatchPage(page int, wg *sync.WaitGroup, ch chan<- pandatypes.ResultMatchLikes) {
	polarity := 2
	defer wg.Done()
	reqStr := "upcoming"
	if page%2 == 1 {
		reqStr = "past"
	}
	pageMap := make(map[string]string)
	// odd pages are past matches, even pages are upcoming matches
	pageMap["page"] = strconv.Itoa(page/polarity + firstPage)
	client.Logger.Debugf("Getting %s matches page %d", reqStr, pageMap["page"])
	body, err := client.getBody([]string{matchesEndpoint, reqStr}, pageMap)
	if err != nil {
		client.Logger.Errorf("Error fetching match page %d: %v", page, err)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}

	var result pandatypes.MatchLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.Logger.Errorf("Error unmarshalling response: %v", err)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	client.Logger.Infof("Got %d %s matches on page %d", len(result), reqStr, page)
	// channel is bounded to the amount of pages get
	ch <- pandatypes.ResultMatchLikes{Matches: result, Err: nil}
}

// GetMatches gets all upcoming and past matches and writes to the database.
// @returns an error if one occurred.
func (client *PandaClient) GetMatches(setup bool) error {
	client.Logger.Info("Getting matches")
	var result pandatypes.MatchLikes
	var wg sync.WaitGroup
	var pageCount int
	if setup {
		pageCount = SetupPages
	} else {
		pageCount = Pages
	}
	varChan := make(chan pandatypes.ResultMatchLikes, pageCount)
	for i := range pageCount {
		wg.Add(1)
		go client.getMatchPage(i, &wg, varChan)
	}
	wg.Wait()
	close(varChan)
	failed := 0
	var lastErr error
	for res := range varChan {
		if res.Err != nil {
			failed++
			lastErr = res.Err
			continue
		}
		result = append(result, res.Matches...)
	}
	client.WriteMatches(result)
	if failed == pageCount {
		return fmt.Errorf("all %d match pages failed: %w", pageCount, lastErr)
	}
	if failed > 0 {
		client.Logger.Warnf("%d of %d match pages failed: %v", failed, pageCount, lastErr)
	}
	return nil
}

func (client *PandaClient) GetTeams(setup bool) error {
	client.Logger.Info("Getting teams")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range Pages {
		client.Logger.Debugf("Getting teams page %d", i)
		keys["page"] = strconv.Itoa(i + firstPage)
		body, err := client.getBody([]string{"teams"}, keys)
		if err != nil {
			client.Logger.Errorf("Error fetching from Pandascore: %v", err)
			return err
		}

		var result pandatypes.TeamLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.Logger.Errorf("Error unmarshalling response: %v", err)
			return err
		}
		for _, teams := range result {
			// current_videogame is nullable and teams.game_id is a foreign key.
			if teams.CurrentVideogame.ID == 0 {
				client.Logger.Debugf("Skipping team %d, it has no game", teams.ID)
				continue
			}
			client.Logger.Debugf("Writing team %s in game %d", teams.Name, teams.CurrentVideogame.ID)
			err = teams.ToRow().WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Errorf("Error writing team to database: %v", err)
				continue
			}
		}
		if !setup {
			break
		}
	}
	return nil
}

// GetLives polls the /matches/running endpoint and updates the is_live flag for all matches.
// Matches in the response have is_live=true; matches not in the response have is_live=false.
// We use /matches/running (not /lives) because /lives only contains matches PandaScore's
// broadcast detector has actively verified — it lags real-time by minutes and misses
// matches whose streams haven't been picked up. /matches/running flips on when a match
// actually starts, which can be before or after its scheduled time.
func (client *PandaClient) GetLives() error {
	client.Logger.Info("Getting live matches")

	result, err := client.getRunningMatches()
	if err != nil {
		return err
	}

	client.Logger.Infof("Got %d live matches", len(result))

	// Write live matches to DB (refreshes stream_url etc.)
	client.WriteMatches(result)

	// Extract IDs of live matches. Never nil: "nothing is live" must clear flags.
	liveIDs := make([]int32, 0, len(result))
	for _, match := range result {
		// A match can end between the feed updating its status and dropping it.
		if match.IsFinished() {
			continue
		}
		id32, convErr := pandatypes.SafeIntToInt32(match.ID)
		if convErr != nil {
			client.Logger.Errorf("Error converting match ID %d to int32: %v", match.ID, convErr)
			continue
		}
		liveIDs = append(liveIDs, id32)
	}

	flipped, err := client.DBConnector.SyncMatchesIsLive(client.Ctx, liveIDs)
	if err != nil {
		client.Logger.Errorf("Error syncing is_live: %v", err)
		return err
	}
	client.Logger.Debugf("Synced is_live for %d live matches", len(liveIDs))

	// A match that just left the running feed has its result on no list we poll
	// until the hourly run, so fetch it now. If this fails, that run still has it.
	for _, row := range flipped {
		if row.IsLive {
			continue
		}
		if refreshErr := client.refreshMatch(row.ID); refreshErr != nil {
			client.Logger.Errorf("Error refreshing ended match %d: %v", row.ID, refreshErr)
		}
	}

	return nil
}

// getRunningMatches reads every page of /matches/running. All of them must load:
// a match missing from this list has is_live cleared, so a partial list would
// mark running matches as over.
func (client *PandaClient) getRunningMatches() (pandatypes.MatchLikes, error) {
	var result pandatypes.MatchLikes
	keys := make(map[string]string)
	for i := range Pages {
		keys["page"] = strconv.Itoa(i + firstPage)
		body, err := client.getBody([]string{matchesEndpoint, "running"}, keys)
		if err != nil {
			client.Logger.Errorf("Error fetching running matches: %v", err)
			return nil, err
		}
		var page pandatypes.MatchLikes
		if err = json.Unmarshal(body, &page); err != nil {
			client.Logger.Errorf("Error unmarshalling response: %v", err)
			return nil, err
		}
		result = append(result, page...)
		if len(page) < perPage {
			break
		}
	}
	return result, nil
}

// refreshMatch rewrites one match from its own endpoint, or records that
// Pandascore no longer has it. Only a 404 for the match itself counts: GetOne
// also resolves parents, and a missing parent says nothing about the match.
func (client *PandaClient) refreshMatch(id int32) error {
	err := client.GetOne(int(id), FlagMatch)
	var statusErr *StatusError
	if errors.As(err, &statusErr) && statusErr.Code == http.StatusNotFound &&
		statusErr.Path == fmt.Sprintf("%s/%d", matchesEndpoint, id) {
		return client.DBConnector.MarkMatchRemoved(client.Ctx, id)
	}
	return err
}

// ReconcileMatches refetches matches that left every feed without reaching a
// terminal status, which the feeds alone would never resolve. Canceled matches
// are the main case. Stops early if Pandascore itself keeps failing, since the
// rest of the batch would only burn the hourly quota, and the next run picks
// them up. A failure confined to one match is skipped: stopping for it would
// halt every run at the same row, because the batch is newest first.
func (client *PandaClient) ReconcileMatches() error {
	ids, err := client.DBConnector.GetMatchesToReconcile(client.Ctx, ReconcileBatch)
	if err != nil {
		return err
	}
	client.Logger.Infof("Reconciling %d matches", len(ids))
	failures := 0
	for _, id := range ids {
		if err = client.refreshMatch(id); err != nil {
			client.Logger.Errorf("Error reconciling match %d: %v", id, err)
			if !isUpstreamFailure(err) {
				continue
			}
			failures++
			if failures >= maxReconcileFailures {
				return fmt.Errorf("stopped reconciling after %d failures in a row: %w", failures, err)
			}
			continue
		}
		failures = 0
	}
	return nil
}

// isUpstreamFailure reports whether an error is Pandascore being unavailable,
// rather than something wrong with one match.
func isUpstreamFailure(err error) bool {
	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		return isRetryable(statusErr.Code)
	}
	var urlErr *url.Error
	return errors.As(err, &urlErr)
}
