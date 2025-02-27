package client

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/feimaomiao/stalka/JsonTypes"
)

// / GetGame gets a game from the Pandascore API
// /@param gameId the id of the game
func (client *PandaClient) GetGame(gameId int) {
	client.logger.Infof("Getting game %d", gameId)
	resp, err := client.MakeRequest([]string{"videogames", fmt.Sprint(gameId)}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result JsonTypes.GameLike
	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}
	client.logger.Infof("Writing game %s", result.Name)
	err = result.ToRow().WriteToDB(client.dbConnector)
	if err != nil {
		return
	}
}

// / GetLeague gets a league from the Pandascore API
// /@param leagueId the id of the league
func (client *PandaClient) GetLeague(leagueId int) {
	client.logger.Infof("Getting league %d", leagueId)
	resp, err := client.MakeRequest([]string{"leagues", fmt.Sprint(leagueId)}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Error("Error reading response: %v", err)
		return
	}

	var result JsonTypes.LeagueLike
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Error("Error unmarshalling response: %v", err)
		return
	}
	// dependency for League: game
	client.gameExistsCheck(result.Videogame.ID)
	client.logger.Infof("Writing league %s", result.Name)
	err = result.ToRow().WriteToDB(client.dbConnector)
	if err != nil {
		client.logger.Error("Error writing league to database: %v", err)
		return
	}

}

// / GetSerie gets a series from the Pandascore API
// /@param seriesId the id of the series
func (client *PandaClient) GetSerie(seriesId int) {
	client.logger.Infof("Getting series %d", seriesId)
	resp, err := client.MakeRequest([]string{"series", fmt.Sprint(seriesId)}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result JsonTypes.SeriesLike
	err = json.Unmarshal(body, &result)
	if err != nil {
		return
	}
	// dependency for Series: league
	client.leagueExistsCheck(result.League.ID)
	client.logger.Infof("Writing Series %s", result.Name)
	err = result.ToRow().WriteToDB(client.dbConnector)
	if err != nil {
		return
	}

}

func (client *PandaClient) GetTournament(tournamentId int) {
	client.logger.Infof("Getting tournament %d", tournamentId)
	resp, err := client.MakeRequest([]string{"tournaments", fmt.Sprint(tournamentId)}, nil)
	if err != nil || resp.StatusCode != 200 {
		client.logger.Error("Error making request to Pandascore API")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		client.logger.Error("Error reading response: %v", err)
		return
	}

	var result JsonTypes.TournamentLike
	err = json.Unmarshal(body, &result)
	if err != nil {
		client.logger.Error("Error unmarshalling response: %v", err)
		return
	}
	// dependency for Tournament: series
	client.seriesExistsCheck(result.Serie.ID)
	client.logger.Infof("Writing tournament %s", result.Name)
	err = result.ToRow().WriteToDB(client.dbConnector)
	if err != nil {
		client.logger.Error("Error writing tournament to database: %v", err)
		return
	}

}
