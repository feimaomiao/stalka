package client

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	pandatypes "github.com/feimaomiao/stalka/pandatypes"
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

func (client *PandaClient) ParseResponse(body []byte, flag GetChoice) (pandatypes.PandaDataLike, error) {
	switch flag {
	case FlagGame:
		var result pandatypes.GameLike
		err := json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	case FlagLeague:
		var result pandatypes.LeagueLike
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
		var result pandatypes.SeriesLike
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
		var result pandatypes.TournamentLike
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
		var result pandatypes.MatchLike
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

func (client *PandaClient) WriteMatches(matches pandatypes.MatchLikes) {
	for _, match := range matches {
		client.logger.Debugf("Checking if tournament %d exists", match.TournamentID)
		err := client.ExistCheck(match.TournamentID, FlagTournament)
		if err != nil {
			client.logger.Error(err)
			continue
		}
		client.logger.Infof("Writing match %s", match.Name)
		row, success := match.ToRow().(pandatypes.MatchRow)
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

func (client *PandaClient) checkTeam(match pandatypes.MatchLike) {
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
			err = pandatypes.TeamRow{
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

// TeamExists checks if a team exists in the database.
// @param db - the database connection
// @param teamID - the ID of the team to check
// @returns true if the team exists, false otherwise, and an error if one occurred.
func TeamExists(db *sql.DB, teamID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM teams WHERE id = $1", teamID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// ExistCheck checks if an entity exists in the database.
// If entity does not exist, get it from the api
// @param id - the ID of the entity to check
// @param flag - the type of entity to check
// It takes an ID and a flag indicating the type of entity to check.
func (client *PandaClient) ExistCheck(id int, flag GetChoice) error {
	var dbString string
	switch flag {
	case FlagGame:
		dbString = "SELECT id FROM games WHERE id = $1"
	case FlagLeague:
		dbString = "SELECT id FROM leagues WHERE id = $1"
	case FlagSeries:
		dbString = "SELECT id FROM series WHERE id = $1"
	case FlagTournament:
		dbString = "SELECT id FROM tournaments WHERE id = $1"
	case FlagMatch:
		dbString = "SELECT id FROM matches WHERE id = $1"
	case FlagTeam:
		dbString = "SELECT id FROM teams WHERE id = $1"
	default:
		client.logger.Error("Invalid flag")
		return fmt.Errorf("invalid flag: %d", flag)
	}
	// checks whether it exists
	err := client.dbConnector.QueryRow(dbString, id).Scan(&id)
	// error exists
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	// does not exist
	if id == 0 || errors.Is(err, sql.ErrNoRows) {
		client.logger.Infof("%d %d currently does not exist", flag, id)
		err = client.GetOne(id, flag)
		if err != nil {
			return err
		}
	}
	return nil
}
