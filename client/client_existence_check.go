package client

import (
	"database/sql"
)

// / Check if game exists in the database
// / @param gameID the id of the game
func GameExists(db *sql.DB, gameID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM games WHERE id = $1", gameID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// / Check if league exists in the database
// / @param leagueID the id of the league
func LeagueExists(db *sql.DB, leagueID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM leagues WHERE id = $1", leagueID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// / Check if series exists in the database
// / @param seriesID the id of the series
func SeriesExists(db *sql.DB, seriesID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM series WHERE id = $1", seriesID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// / Check if tournament exists in the database
// / @param tournamentID the id of the tournament
func TournamentExists(db *sql.DB, tournamentID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM tournaments WHERE id = $1", tournamentID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// / Check if team exists in the database
// / @param teamID the id of the team
func TeamExists(db *sql.DB, teamID int) (bool, error) {
	var id int
	err := db.QueryRow("SELECT id FROM teams WHERE id = $1", teamID).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return id != 0, nil
}

// / Check if game exists in the database, if not, fetch it from the API
// / @param gameID the id of the game
func (client *PandaClient) gameExistsCheck(gameID int) error {
	client.logger.Infof("Checking if game %d exists", gameID)
	exists, err := GameExists(client.dbConnector, gameID)
	if err != nil {
		client.logger.Errorf("Error checking if game exists: %v", err)
		return err
	}
	if !exists {
		client.logger.Infof("Game %d does not exist", gameID)
		client.GetGame(gameID)
	}
	return nil
}

// / Check if league exists in the database, if not, fetch it from the API
// /@param leagueID the id of the league
func (client *PandaClient) leagueExistsCheck(leagueID int) error {
	client.logger.Infof("Checking if league %d exists", leagueID)
	exists, err := LeagueExists(client.dbConnector, leagueID)
	if err != nil {
		client.logger.Errorf("Error checking if league exists: %v", err)
	}
	if !exists {
		client.logger.Infof("League %d does not exist", leagueID)
		client.GetLeague(leagueID)
	} else {
		client.logger.Infof("League %d exists", leagueID)
	}
	return nil
}

// / Check if series exists in the database, if not, fetch it from the API
// /@param seriesID the id of the series
func (client *PandaClient) seriesExistsCheck(seriesID int) error {
	client.logger.Infof("Checking if series %d exists", seriesID)
	exists, err := SeriesExists(client.dbConnector, seriesID)
	if err != nil {
		client.logger.Errorf("Error checking if series exists: %v", err)
		return err
	}
	if !exists {
		client.logger.Infof("Series %d does not exist", seriesID)
		client.GetSerie(seriesID)
	} else {
		client.logger.Infof("Series %d exists", seriesID)
	}
	return nil
}

// / Check if tournament exists in the database, if not, fetch it from the API
// /@param tournamentID the id of the tournament
func (client *PandaClient) tournamentExistsCheck(tournamentID int) error {
	client.logger.Infof("Checking if tournament %d exists", tournamentID)
	exists, err := TournamentExists(client.dbConnector, tournamentID)
	if err != nil {
		client.logger.Errorf("Error checking if tournament exists: %v", err)
	}
	if !exists {
		client.logger.Infof("Tournament %d does not exist", tournamentID)
		client.GetTournament(tournamentID)
	} else {
		client.logger.Infof("Tournament %d exists", tournamentID)
	}
	return nil
}
