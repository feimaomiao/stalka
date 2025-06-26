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
	sortedBy = "-modified_at"
	Pages    = 20
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
		client.logger.Infof("Writing game %s", game.Name)
		err = game.ToRow().WriteToDB(client.dbConnector)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLeagues gets the first 100 leagues from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetLeagues() error {
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
			client.logger.Infof("Writing league %s", league.Name)
			err = league.ToRow().WriteToDB(client.dbConnector)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetSeries gets the first 100 series from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetSeries() error {
	client.logger.Info("Getting series")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
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
		client.logger.Infof("Writing series %s, with league_id %d", series.Name, series.LeagueID)
		err = series.ToRow().WriteToDB(client.dbConnector)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetTournaments gets the first 100 tournaments from the Pandascore API.
// @returns an error if one occurred.
func (client *PandaClient) GetTournaments() error {
	client.logger.Info("Getting tournaments")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
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
		client.logger.Infof("Writing tournament %s", tournament.Name)
		err = tournament.ToRow().WriteToDB(client.dbConnector)
		if err != nil {
			return err
		}
	}
	return nil
}

// Goroutine to get one page of matches. Sends the data to the channel.
func (client *PandaClient) getMatchPage(page int, wg *sync.WaitGroup, ch chan<- pandatypes.ResultMatchLikes) {
	polarity := 2
	defer wg.Done()
	client.logger.Infof("Getting upcoming matches page %d", page)
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
func (client *PandaClient) GetMatches() error {
	client.logger.Info("Getting matches")
	var result pandatypes.MatchLikes
	var wg sync.WaitGroup

	varChan := make(chan pandatypes.ResultMatchLikes, Pages)
	for i := 1; i <= Pages; i++ {
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
