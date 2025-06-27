package client

import (
	"io"
	"net/http"
	"strconv"
	"sync"

	"encoding/json"

	pandatypes "github.com/feimaomiao/stalka/pandatypes"
	// loads .env file automatically.
	_ "github.com/joho/godotenv/autoload"
)

const (
	sortedBy   = "-modified_at"
	Pages      = 20
	SetupPages = 60
)

// UpdateGames updates all games in the database.
// @returns an error if one occurred.
func (client *PandaClient) UpdateGames() error {
	client.logger.Info("Updating games")
	resp, err := client.MakeRequest([]string{"videogames"}, nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		client.logger.Error("Error making request to Pandascore API")
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
		return err
	}
	for _, game := range result {
		client.logger.Debugf("Writing game %s", game.Name)
		err = game.ToRow().WriteToDB(client.dbConnector)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLeagues gets the first leagues from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetLeagues(setup bool) error {
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range 15 {
		client.logger.Info("Getting leagues page " + strconv.Itoa(i))
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"leagues"}, keys)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.logger.Error("Error making request to Pandascore API")
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var result pandatypes.LeagueLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			return err
		}
		for _, league := range result {
			client.logger.Debugf("Writing league %s", league.Name)
			err = league.ToRow().WriteToDB(client.dbConnector)
			if err != nil {
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
	client.logger.Info("Getting series")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range 20 {
		client.logger.Debugf("Getting series page %d", i)
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"series"}, nil)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.logger.Error("Error making request to Pandascore API")
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			client.logger.Error("Error reading response: %v", err)
			return err
		}

		var result pandatypes.SeriesLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.logger.Error("Error unmarshalling response: %v", err)
			return err
		}
		for _, series := range result {
			err = client.ExistCheck(series.League.ID, FlagLeague)
			if err != nil {
				client.logger.Error("Error checking if league exists: %v", err)
				continue
			}
			client.logger.Debugf("Writing series %s, with league_id %d", series.Name, series.LeagueID)
			err = series.ToRow().WriteToDB(client.dbConnector)
			if err != nil {
				return err
			}
		}
		if !setup {
			break
		}
	}
	return nil
}

// GetTournaments gets the first 200 tournaments from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetTournaments(setup bool) error {
	client.logger.Info("Getting tournaments")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range 20 {
		keys["page"] = strconv.Itoa(i)
		client.logger.Debugf("Getting tournaments page %d", i)
		resp, err := client.MakeRequest([]string{"tournaments"}, keys)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.logger.Error("Error making request to Pandascore API")
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
		for _, tournament := range result {
			client.logger.Debugf("Checking if series exists %d", tournament.SerieID)
			err = client.ExistCheck(tournament.SerieID, FlagSeries)
			if err != nil {
				client.logger.Error(err)
				continue
			}
			client.logger.Debugf("Writing tournament %s", tournament.Name)
			err = tournament.ToRow().WriteToDB(client.dbConnector)
			if err != nil {
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
	client.logger.Debugf("Getting upcoming matches page %d", page)
	reqStr := "upcoming"
	if page%2 == 1 {
		reqStr = "past"
	}
	pageMap := make(map[string]string)
	// odd pages are past matches, even pages are upcoming matches
	// there are 20 pages in total
	pageMap["page"] = strconv.Itoa(page / polarity)
	resp, err := client.MakeRequest([]string{"matches", reqStr}, pageMap)
	if err != nil || resp.StatusCode != http.StatusOK {
		client.logger.Errorf("Error making request to Pandascore API %v, %d on request %d", err, resp.StatusCode, page)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Errorf("Error reading response: %v", err)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}

	var result pandatypes.MatchLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Errorf("Error unmarshalling response: %v", err)
		ch <- pandatypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	// channel is bounded to the amount of pages get
	ch <- pandatypes.ResultMatchLikes{Matches: result, Err: nil}
}

// GetMatches gets all upcoming and past matches and writes to the database.
// @returns an error if one occurred.
func (client *PandaClient) GetMatches(setup bool) error {
	client.logger.Info("Getting matches")
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
			client.logger.Error(res.Err)
			continue
		}
		result = append(result, res.Matches...)
	}
	client.WriteMatches(result)
	return nil
}

func (client *PandaClient) GetTeams(setup bool) error {
	client.logger.Info("Getting teams")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
	for i := range 20 {
		client.logger.Debugf("Getting teams page %d", i)
		keys["page"] = strconv.Itoa(i)
		resp, err := client.MakeRequest([]string{"teams"}, nil)
		if err != nil || resp.StatusCode != http.StatusOK {
			client.logger.Error("Error making request to Pandascore API")
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			client.logger.Error("Error reading response: %v", err)
			return err
		}

		var result pandatypes.TeamLikes
		err = json.Unmarshal(body, &result)
		if err != nil {
			client.logger.Error("Error unmarshalling response: %v", err)
			return err
		}
		for _, teams := range result {
			client.logger.Debugf("Writing team %s in game %d", teams.Name, teams.CurrentVideogame.ID)
			err = teams.ToRow().WriteToDB(client.dbConnector)
			if err != nil {
				client.logger.Error("Error writing team to database: %v", err)
				return err
			}
		}
		if !setup {
			break
		}
	}
	return nil
}
