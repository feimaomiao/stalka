package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/jackc/pgx/v5"
)

// converts the flag to a pandaapi recognized string.
// @param flag - the flag to convert.
// @returns the string representation of the flag and an error if one occurred.
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

// ParseResponse parses the response body to a datatype based on the flag.
// It also ensures that all dependencies are checked and created.
// @param body - the response body to parse.
// @param flag - the type of entity to parse.
// @returns the parsed entity and an error if one occurred.
func (client *PandaClient) ParseResponse(body []byte, flag GetChoice) (pandatypes.PandaDataLike, error) {
	result, err := client.unmarshalByFlag(body, flag)
	if err != nil {
		return nil, err
	}

	if err = client.ensureDependencies(result, flag); err != nil {
		return nil, err
	}

	return result, nil
}

// unmarshalByFlag unmarshals the body into the appropriate type based on flag.
func (client *PandaClient) unmarshalByFlag(body []byte, flag GetChoice) (pandatypes.PandaDataLike, error) {
	var result pandatypes.PandaDataLike
	var err error

	switch flag {
	case FlagGame:
		var r pandatypes.GameLike
		err = json.Unmarshal(body, &r)
		result = r
	case FlagLeague:
		var r pandatypes.LeagueLike
		err = json.Unmarshal(body, &r)
		result = r
	case FlagSeries:
		var r pandatypes.SeriesLike
		err = json.Unmarshal(body, &r)
		result = r
	case FlagTournament:
		var r pandatypes.TournamentLike
		err = json.Unmarshal(body, &r)
		result = r
	case FlagMatch:
		var r pandatypes.MatchLike
		err = json.Unmarshal(body, &r)
		result = r
	case FlagTeam:
		var r pandatypes.TeamLike
		err = json.Unmarshal(body, &r)
		result = r
	default:
		return nil, fmt.Errorf("invalid flag: %d", flag)
	}

	if err != nil {
		client.Logger.Errorf("Error unmarshalling response: %v", err)
		return nil, err
	}

	return result, nil
}

// ensureDependencies checks and creates dependencies for the parsed entity.
func (client *PandaClient) ensureDependencies(result pandatypes.PandaDataLike, flag GetChoice) error {
	dep := client.getDependency(result, flag)
	if dep == nil {
		return nil // No dependencies to check
	}

	exists, err := client.ExistCheck(dep.id, dep.flag)
	if err != nil {
		client.Logger.Errorf("Error checking if %s %d exists: %v", dep.name, dep.id, err)
		return err
	}

	if !exists {
		if err = client.GetOne(dep.id, dep.flag); err != nil {
			client.Logger.Errorf("Error getting %s %d: %v", dep.name, dep.id, err)
			return err
		}
	}

	return nil
}

// Dependency represents a Dependency that needs to be checked.
type Dependency struct {
	id   int
	flag GetChoice
	name string
}

// getDependency returns the dependency Debug for a given entity type.
func (client *PandaClient) getDependency(result pandatypes.PandaDataLike, flag GetChoice) *Dependency {
	switch flag {
	case FlagLeague:
		if r, ok := result.(pandatypes.LeagueLike); ok {
			return &Dependency{
				id:   r.Videogame.ID,
				flag: FlagGame,
				name: "game",
			}
		}
	case FlagSeries:
		if r, ok := result.(pandatypes.SeriesLike); ok {
			return &Dependency{
				id:   r.LeagueID,
				flag: FlagLeague,
				name: "league",
			}
		}
	case FlagTournament:
		if r, ok := result.(pandatypes.TournamentLike); ok {
			return &Dependency{
				id:   r.SerieID,
				flag: FlagSeries,
				name: "series",
			}
		}
	case FlagMatch:
		if r, ok := result.(pandatypes.MatchLike); ok {
			return &Dependency{
				id:   r.TournamentID,
				flag: FlagTournament,
				name: "tournament",
			}
		}
	// No dependencies for FlagGame and FlagTeam
	case FlagTeam:
		return nil
	case FlagGame:
		return nil
	}
	return nil
}

// GetOne gets a single entity from the Pandascore API.
// @param id - the ID of the entity to get.
// @param flag - the type of entity to get.
// @returns an error if one occurred.
func (client *PandaClient) GetOne(id int, flag GetChoice) error {
	searchString, err := flagToString(flag)
	if err != nil {
		client.Logger.Error("Error converting flag to string: %v", err)
		return err
	}
	client.Logger.Debugf("Getting %s %d", searchString, id)
	resp, err := client.MakeRequest([]string{searchString, strconv.Itoa(id)}, nil)
	if err != nil {
		client.Logger.Error("Error making request to Pandascore API: %v", err)
		return err
	}
	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		client.Logger.Error("Error: received status code %d", resp.StatusCode)
		return fmt.Errorf("received status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result, err := client.ParseResponse(body, flag)
	if err != nil {
		client.Logger.Error("Error parsing response: %v", err)
	}

	err = result.ToRow().WriteToDB(client.Ctx, client.DBConnector)
	if err != nil {
		return err
	}
	return nil
}

// WriteMatches writes the matches to the database.
// @param matches - the matches to write.
func (client *PandaClient) WriteMatches(matches pandatypes.MatchLikes) {
	for _, match := range matches {
		exists, err := client.ExistCheck(match.TournamentID, FlagTournament)
		if err != nil {
			client.Logger.Error(err)
			continue
		}
		if !exists {
			err = client.GetOne(match.TournamentID, FlagTournament)
			if err != nil {
				client.Logger.Error(err)
				continue
			}
		}
		client.Logger.Debugf("Writing match %s", match.Name)
		row, success := match.ToRow().(pandatypes.MatchRow)
		if !success {
			client.Logger.Errorf("Error converting match row to match row (??), %v", row)
			continue
		}
		err = row.WriteToDB(client.Ctx, client.DBConnector)
		if err != nil {
			client.Logger.Error(err)
			continue
		}
		client.checkTeam(match)
	}
}

// checkTeam checks if the teams in the match exist in the database.
// @param match - the match to check.
func (client *PandaClient) checkTeam(match pandatypes.MatchLike) {
	if match.WinnerType != "Team" {
		client.Logger.Debugf("Match %d is not a team match, but is a %s match", match.ID, match.WinnerType)
		return
	}
	for _, opponent := range match.Opponents {
		exists, err := client.ExistCheck(opponent.Opponent.ID, FlagTeam)
		if err != nil {
			client.Logger.Error(err)
			continue
		}
		if !exists {
			client.Logger.Debugf("Team %s does not exist", opponent.Opponent.Name)
			err = pandatypes.TeamRow{
				ID:        opponent.Opponent.ID,
				GameID:    match.Videogame.ID,
				Name:      opponent.Opponent.Name,
				Acronym:   opponent.Opponent.Acronym,
				Slug:      opponent.Opponent.Slug,
				ImageLink: opponent.Opponent.ImageURL,
			}.WriteToDB(client.Ctx, client.DBConnector)
			if err != nil {
				client.Logger.Error(err)
				continue
			}
		} else {
			client.Logger.Debugf("Team %s exists", opponent.Opponent.Name)
		}
	}
}

// ExistCheck checks if an entity exists in the database.
// @param id - the ID of the entity to check.
// @param flag - the type of entity to check.
// @returns an error if one occurred.
func (client *PandaClient) ExistCheck(id int, flag GetChoice) (bool, error) {
	var dbResult int64
	var err error
	stringFlag, err := flagToString(flag)
	if err != nil {
		client.Logger.Error("Error converting flag to string: %v", err)
		return false, err
	}
	client.Logger.Debugf("Checking if %s with ID %d exists", stringFlag, id)

	// Safely convert int to int32
	id32, err := pandatypes.SafeIntToInt32(id)
	if err != nil {
		client.Logger.Errorf("Error converting ID %d to int32: %v", id, err)
		return false, err
	}

	switch flag {
	case FlagGame:
		dbResult, err = client.DBConnector.GameExist(client.Ctx, id32)
	case FlagLeague:
		dbResult, err = client.DBConnector.LeagueExist(client.Ctx, id32)
	case FlagSeries:
		dbResult, err = client.DBConnector.SeriesExist(client.Ctx, id32)
	case FlagTournament:
		dbResult, err = client.DBConnector.TournamentExist(client.Ctx, id32)
	case FlagMatch:
		dbResult, err = client.DBConnector.MatchExist(client.Ctx, id32)
	case FlagTeam:
		dbResult, err = client.DBConnector.TeamExist(client.Ctx, id32)
	// this would never happen as we vet the flags before calling
	default:
		client.Logger.Error("Invalid flag")
		return false, fmt.Errorf("invalid flag: %d", flag)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		client.Logger.Errorf("Error checking if entity exists: %v", err)
		return false, err
	}
	if dbResult == 0 {
		client.Logger.Debugf("%s with ID %d does not exist", stringFlag, id)
		return false, nil
	}
	client.Logger.Debugf("%s with ID %d exists", stringFlag, id)
	return true, nil
}
