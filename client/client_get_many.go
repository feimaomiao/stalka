package client

import (
	"io"
	"net/http"
	"strconv"
	"sync"

	"encoding/json"

	"github.com/feimaomiao/stalka/jsontypes"

	// loads .env file automatically.
	_ "github.com/joho/godotenv/autoload"
)

const (
	sortedBy = "-modified_at"
)

// UpdateGames updates all games in the database.
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

	var result jsontypes.GameLikes
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
func (client *PandaClient) GetLeagues() error {
	client.logger.Info("Getting leagues")
	keys := make(map[string]string)
	keys["sort"] = sortedBy
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

	var result jsontypes.LeagueLikes
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
	return nil
}

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

	var result jsontypes.SeriesLikes
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

	var result jsontypes.TournamentLikes
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

// / Goroutine to get one page of matches.
func (client *PandaClient) getMatchPage(page int, wg *sync.WaitGroup, ch chan<- jsontypes.ResultMatchLikes) {
	polarity := 2
	defer wg.Done()
	client.logger.Infof("Getting upcoming matches page %d", page)
	reqStr := "upcoming"
	if page%2 == 1 {
		reqStr = "past"
	}
	pageMap := make(map[string]string)
	pageMap["page"] = strconv.Itoa(page / polarity)
	resp, err := client.MakeRequest([]string{"matches", reqStr}, pageMap)
	if err != nil || resp.StatusCode != http.StatusOK {
		client.logger.Errorf("Error making request to Pandascore API %v, %d on request %d", err, resp.StatusCode, page)
		ch <- jsontypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Errorf("Error reading response: %v", err)
		ch <- jsontypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}

	var result jsontypes.MatchLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Errorf("Error unmarshalling response: %v", err)
		ch <- jsontypes.ResultMatchLikes{Matches: nil, Err: err}
		return
	}
	ch <- jsontypes.ResultMatchLikes{Matches: result, Err: nil}
}

const Pages = 20

// GetMatches gets all upcoming matches and writes to the database.
func (client *PandaClient) GetMatches() error {
	client.logger.Info("Getting matches")
	var result jsontypes.MatchLikes
	var wg sync.WaitGroup

	varChan := make(chan jsontypes.ResultMatchLikes, Pages)
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
