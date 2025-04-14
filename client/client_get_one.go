package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/feimaomiao/stalka/jsontypes"
)

func flagToString(flag GetChoice) (string, error) {
	switch flag {
	case FlagGame:
		return "videogames", nil
	case FlagLeague:
		return "leagues", nil
	case FlagSeries:
		return "series", nil
	case FlagTournament:
		return "tournaments", nil
	case FlagMatch:
		return "matches", nil
	case FlagTeam:
		return "teams", nil
	default:
		return "", fmt.Errorf("invalid flag: %d", flag)
	}
}

func (client *PandaClient) ParseResponse(body []byte, flag GetChoice) (jsontypes.PandaDataLike, error) {
	switch flag {
	case FlagGame:
		var result jsontypes.GameLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagLeague:
		var result jsontypes.LeagueLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		err = client.ExistCheck(result.Videogame.ID, FlagGame)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagSeries:
		var result jsontypes.SeriesLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		err = client.ExistCheck(result.LeagueID, FlagLeague)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagTournament:
		var result jsontypes.TournamentLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		err = client.ExistCheck(result.SerieID, FlagSeries)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagMatch:
		var result jsontypes.MatchLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		err = client.ExistCheck(result.TournamentID, FlagTournament)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagTeam:
		return nil, errors.New("should not call getone on team")
	default:
		return nil, fmt.Errorf("invalid flag: %d", flag)
	}
}
func (client *PandaClient) GetOne(id int, flag GetChoice) error {
	searchString, err := flagToString(flag)
	if err != nil {
		client.logger.Error("Error converting flag to string: %v", err)
		return err
	}
	client.logger.Infof("Getting %s %d", searchString, id)
	resp, err := client.MakeRequest([]string{searchString, strconv.Itoa(id)}, nil)
	if err != nil {
		client.logger.Error("Error making request to Pandascore API: %v", err)
		return err
	}
	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		client.logger.Error("Error: received status code %d", resp.StatusCode)
		return fmt.Errorf("received status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result, err := client.ParseResponse(body, flag)
	if err != nil {
		client.logger.Error("Error parsing response: %v", err)
	}

	err = result.ToRow().WriteToDB(client.dbConnector)
	if err != nil {
		return err
	}
	return nil
}

func (client *PandaClient) WriteMatches(matches jsontypes.MatchLikes) {
	for _, match := range matches {
		client.logger.Debugf("Checking if tournament %d exists", match.TournamentID)
		err := client.ExistCheck(match.TournamentID, FlagTournament)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		client.logger.Infof("Writing match %s", match.Name)
		row, success := match.ToRow().(jsontypes.MatchRow)
		if !success {
			client.logger.Errorf("Error converting match row to match row (??), %v", row)
			continue
		}
		err = row.WriteToDB(client.dbConnector)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		if row.Finished {
			client.checkTeam(match)
		}
	}
}

func (client *PandaClient) checkTeam(match jsontypes.MatchLike) {
	if match.WinnerType != "Team" {
		client.logger.Infof("Match %d is not a team match", match.ID)
		return
	}
	for _, opponent := range match.Opponents {
		exists, err := TeamExists(client.dbConnector, opponent.Opponent.ID)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		if !exists {
			client.logger.Infof("Team %s does not exist", opponent.Opponent.Name)
			err = jsontypes.TeamRow{
				ID:        opponent.Opponent.ID,
				Name:      opponent.Opponent.Name,
				Acronym:   opponent.Opponent.Acronym,
				Slug:      opponent.Opponent.Slug,
				ImageLink: opponent.Opponent.ImageURL,
			}.WriteToDB(client.dbConnector)
			if err != nil {
				client.logger.Error(err)
				continue
			}
		} else {
			client.logger.Infof("Team %s exists", opponent.Opponent.Name)
		}
	}
}
