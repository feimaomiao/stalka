package client

import (
	"database/sql"
	"errors"
	"fmt"
)

// TeamExists checks if a team exists in the database.
// @returns true if the team exists, false otherwise, and an error if one occurred.
func TeamExists(db *sql.DB, teamID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM teams WHERE id = $1", teamID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

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
	err := client.dbConnector.QueryRow(dbString, id).Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if id == 0 {
		client.logger.Infof("%d %d currently does not exist", flag, id)
		err = client.GetOne(id, flag)
		if err != nil {
			return err
		}
	}
	return nil
}
