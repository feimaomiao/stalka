package client

import (
	"io"
	"sync"

	"encoding/json"

	"fmt"

	"github.com/feimaomiao/stalka/JsonTypes"
	_ "github.com/joho/godotenv/autoload"
)

// / UpdateGames updates all games in the database
func (client *PandaClient) UpdateGames() error {
	client.logger.Info("Updating games")
	resp, err := client.MakeRequest([]string{"videogames"}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result JsonTypes.GameLikes
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

// / GetLeagues gets the first 100 leagues from the Pandascore API
func (client *PandaClient) GetLeagues() error {
	client.logger.Info("Getting leagues")
	keys := make(map[string]string)
	keys["sort"] = "-modified_at"
	resp, err := client.MakeRequest([]string{"leagues"}, keys)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result JsonTypes.LeagueLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	for _, league := range result {
		client.logger.Infof("Writing league %s", league.Name)
		err := league.ToRow().WriteToDB(client.dbConnector)
		if err != nil {
			return err
		}
	}
	return nil
}

// / GetSeries gets the first 100 series from the Pandascore API
// / This should be called only at startup
func (client *PandaClient) GetSeries() error {
	client.logger.Info("Getting series")
	keys := make(map[string]string)
	keys["sort"] = "-modified_at"
	resp, err := client.MakeRequest([]string{"series"}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Error("Error reading response: %v", err)
		return err
	}

	var result JsonTypes.SeriesLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Error("Error unmarshalling response: %v", err)
		return err
	}
	for _, series := range result {
		client.logger.Debugf("Checking if leagueid = .League.ID exists, %t, %d ,%d", series.LeagueID == series.League.ID, series.LeagueID, series.League.ID)
		err = client.leagueExistsCheck(series.League.ID)
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
	keys["sort"] = "-modified_at"
	resp, err := client.MakeRequest([]string{"tournaments"}, keys)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result JsonTypes.TournamentLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	for _, tournament := range result {
		client.logger.Debugf("Checking if series exists %d", tournament.SerieID)
		err := client.seriesExistsCheck(tournament.SerieID)
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

// / Goroutine to get one page of matches
func (client *PandaClient) getMatchPage(page int, wg *sync.WaitGroup, ch chan<- JsonTypes.ResultMatchLikes) {
	defer wg.Done()
	client.logger.Infof("Getting upcoming matches page %d", page)
	reqStr := "upcoming"
	if page%2 == 1 {
		reqStr = "past"
	}
	pageMap := make(map[string]string)
	pageMap["page"] = fmt.Sprint(page / 2)
	resp, err := client.MakeRequest([]string{"matches", reqStr}, pageMap)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Errorf("Error making request to Pandascore API %v, %d on request %d", err, resp.StatusCode, page)
		ch <- JsonTypes.ResultMatchLikes{Err: err}
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Errorf("Error reading response: %v", err)
		ch <- JsonTypes.ResultMatchLikes{Err: err}
		return
	}

	var result JsonTypes.MatchLikes
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Errorf("Error unmarshalling response: %v", err)
		ch <- JsonTypes.ResultMatchLikes{Err: err}
		return
	}
	ch <- JsonTypes.ResultMatchLikes{Matches: result}
}

// / GetUpcoming gets all upcoming matches and writes to the database
func (client *PandaClient) GetMatches() error {
	client.logger.Info("Getting matches")
	var result JsonTypes.MatchLikes
	var wg sync.WaitGroup

	varChan := make(chan JsonTypes.ResultMatchLikes, 20)
	for i := 1; i <= 20; i++ {
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
	for _, match := range result {
		client.logger.Debugf("Checking if tournament exists %d", match.TournamentID)
		err := client.tournamentExistsCheck(match.TournamentID)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		client.logger.Infof("Writing match %s", match.Name)
		row := match.ToRow()
		err = row.WriteToDB(client.dbConnector)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		if row.Finished {
			if match.WinnerType != "Team" {
				client.logger.Errorf("Match %d winner type is not team", match.ID)
				continue
			}
			for _, opponents := range match.Opponents {
				exists, err := TeamExists(client.dbConnector, opponents.Opponent.ID)
				if err != nil {
					client.logger.Error(err)
					continue
				}
				if !exists {
					client.logger.Infof("Team %d does not exist", opponents.Opponent.ID)
					tr := JsonTypes.TeamRow{
						ID:         opponents.Opponent.ID,
						Name:       opponents.Opponent.Name,
						Acronym:    opponents.Opponent.Acronym,
						Slug:       opponents.Opponent.Slug,
						Image_link: opponents.Opponent.ImageURL}
					err = tr.WriteToDB(client.dbConnector)
					if err != nil {
						client.logger.Error(err)
						continue
					}
				} else {
					client.logger.Infof("Team %d exists", opponents.Opponent.ID)
				}
			}
		}
	}
	return nil
}
