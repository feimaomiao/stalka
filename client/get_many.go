package client

import (
	"io"
	"net/http"
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
	SetupPages = 30
)

// UpdateGames updates all games in the database.
// @returns an error if one occurred.
func (client *PandaClient) UpdateGames() error {
	client.Logger.Info("Updating games")
	resp, err := client.MakeRequest([]string{"videogames"}, nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		client.Logger.Errorf("Error making request to Pandascore API: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
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
			return err
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
	for i := range 15 {
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"leagues"}, keys)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.Logger.Errorf("Error making request to Pandascore API: %v", err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			client.Logger.Errorf("Error reading response: %v", err)
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
				return err
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
	for i := range 20 {
		client.Logger.Debugf("Getting series page %d", i)
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"series"}, nil)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.Logger.Errorf("Error making request to Pandascore API: %v", err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			client.Logger.Errorf("Error reading response: %v", err)
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
				return err
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
	for i := range 10 {
		keys["page"] = strconv.Itoa(i)
		client.Logger.Debugf("Getting tournaments page %d", i)
		resp, err := client.MakeRequest([]string{"tournaments"}, keys)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.Logger.Errorf("Error making request to Pandascore API: %v", err)
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
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
				return err
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
	pageMap["page"] = strconv.Itoa(page / polarity)
	client.Logger.Debugf("Getting %s matches page %d", reqStr, pageMap["page"])
	resp, err := client.MakeRequest([]string{"matches", reqStr}, pageMap)
	if err != nil {
		client.Logger.Errorf("Error making request to Pandascore API %v, %d on request %d", err, resp.StatusCode, page)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	if resp.StatusCode != http.StatusOK {
		client.Logger.Errorf("Error making request to Pandascore API %d on request %d", resp.StatusCode, page)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.Logger.Errorf("Error reading response: %v", err)
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
	for i := 1; i <= pageCount; i++ {
		wg.Add(1)
		go client.getMatchPage(i, &wg, varChan)
	}
	wg.Wait()
	close(varChan)
	for res := range varChan {
		if res.Err != nil {
			continue
		}
		result = append(result, res.Matches...)
	}
	client.WriteMatches(result)
	return nil
}

func (client *PandaClient) GetTeams(setup bool) error {
	client.Logger.Info("Getting teams")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range 20 {
		client.Logger.Debugf("Getting teams page %d", i)
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"teams"}, nil)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.Logger.Error("Error making request to Pandascore API")
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			client.Logger.Errorf("Error reading response: %v", err)
			return err
		}

		var result pandatypes.TeamLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.Logger.Errorf("Error unmarshalling response: %v", err)
			return err
		}
		for _, teams := range result {
			client.Logger.Debugf("Writing team %s in game %d", teams.Name, teams.CurrentVideogame.ID)
			err = teams.ToRow().WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Errorf("Error writing team to database: %v", err)
				return err
			}
		}
		if !setup {
			break
		}
	}
	return nil
}
