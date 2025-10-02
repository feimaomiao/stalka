package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/h2non/gock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nbio/st"
	"github.com/pashagolub/pgxmock/v4"
	"go.uber.org/zap/zaptest"
)

func TestFlagToString(t *testing.T) {
	testCases := []struct {
		name           string
		flag           GetChoice
		expectedString string
		expectError    bool
	}{
		{
			name:           "Game flag",
			flag:           FlagGame,
			expectedString: "videogames",
			expectError:    false,
		},
		{
			name:           "League flag",
			flag:           FlagLeague,
			expectedString: "leagues",
			expectError:    false,
		},
		{
			name:           "Series flag",
			flag:           FlagSeries,
			expectedString: "series",
			expectError:    false,
		},
		{
			name:           "Tournament flag",
			flag:           FlagTournament,
			expectedString: "tournaments",
			expectError:    false,
		},
		{
			name:           "Match flag",
			flag:           FlagMatch,
			expectedString: "matches",
			expectError:    false,
		},
		{
			name:           "Team flag",
			flag:           FlagTeam,
			expectedString: "teams",
			expectError:    false,
		},
		{
			name:           "Invalid flag",
			flag:           GetChoice(999),
			expectedString: "",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flagToString(tc.flag)

			if tc.expectError {
				st.Reject(t, err, nil)
				st.Expect(t, strings.Contains(err.Error(), "invalid flag"), true)
			} else {
				st.Expect(t, err, nil)
				st.Expect(t, result, tc.expectedString)
			}
		})
	}
}

func TestGetDependency(t *testing.T) {
	client := &PandaClient{}
	nilDependency := (*Dependency)(nil)

	t.Run("League dependency - Game", func(t *testing.T) {
		league := pandatypes.LeagueLike{
			Videogame: struct {
				ID             int    `json:"id"`
				Name           string `json:"name"`
				CurrentVersion string `json:"current_version"`
				Slug           string `json:"slug"`
			}{
				ID:   123,
				Name: "Counter-Strike 2",
				Slug: "cs2",
			},
		}

		dep := client.getDependency(league, FlagLeague)
		st.Reject(t, dep, nilDependency)
		st.Expect(t, dep.id, 123)
		st.Expect(t, dep.flag, FlagGame)
		st.Expect(t, dep.name, "game")
	})

	t.Run("Series dependency - League", func(t *testing.T) {
		series := pandatypes.SeriesLike{
			LeagueID: 456,
		}

		dep := client.getDependency(series, FlagSeries)
		st.Reject(t, dep, nilDependency)
		st.Expect(t, dep.id, 456)
		st.Expect(t, dep.flag, FlagLeague)
		st.Expect(t, dep.name, "league")
	})

	t.Run("Tournament dependency - Series", func(t *testing.T) {
		tournament := pandatypes.TournamentLike{
			SerieID: 789,
		}

		dep := client.getDependency(tournament, FlagTournament)

		st.Reject(t, dep, nilDependency)
		st.Expect(t, dep.id, 789)
		st.Expect(t, dep.flag, FlagSeries)
		st.Expect(t, dep.name, "series")
	})

	t.Run("Match dependency - Tournament", func(t *testing.T) {
		match := pandatypes.MatchLike{
			TournamentID: 101112,
		}

		dep := client.getDependency(match, FlagMatch)

		st.Reject(t, dep, nilDependency)
		st.Expect(t, dep.id, 101112)
		st.Expect(t, dep.flag, FlagTournament)
		st.Expect(t, dep.name, "tournament")
	})

	t.Run("Game - no dependency", func(t *testing.T) {
		game := pandatypes.GameLike{
			ID:   1,
			Name: "Counter-Strike 2",
			Slug: "cs2",
		}

		dep := client.getDependency(game, FlagGame)

		st.Expect(t, dep, nilDependency)
	})

	t.Run("Team - no dependency", func(t *testing.T) {
		team := pandatypes.TeamLike{
			ID:      500,
			Name:    "Team Liquid",
			Acronym: "TL",
		}

		dep := client.getDependency(team, FlagTeam)

		st.Expect(t, dep, nilDependency)
	})

	t.Run("Invalid type assertion", func(t *testing.T) {
		// Pass wrong type for the flag
		game := pandatypes.GameLike{
			ID:   1,
			Name: "Counter-Strike 2",
		}

		// This should return nil because GameLike cannot be asserted as LeagueLike
		dep := client.getDependency(game, FlagLeague)

		st.Expect(t, dep, nilDependency)
	})
}

func TestUnmarshalByFlag(t *testing.T) {
	client := &PandaClient{
		Logger: zaptest.NewLogger(t).Sugar(),
	}

	t.Run("Unmarshal Game", func(t *testing.T) {
		gameData, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(gameData, FlagGame)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		game, ok := result.(pandatypes.GameLike)
		st.Expect(t, ok, true)
		st.Expect(t, game.ID, 34)
		st.Expect(t, game.Name, "Mobile Legends: Bang Bang")
		st.Expect(t, game.Slug, "mlbb")
	})

	t.Run("Unmarshal League", func(t *testing.T) {
		leagueData, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(leagueData, FlagLeague)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		league, ok := result.(pandatypes.LeagueLike)
		st.Expect(t, ok, true)
		st.Expect(t, league.ID, 289)
		st.Expect(t, league.Name, "NA LCS")
		st.Expect(t, league.Videogame.ID, 1)
		st.Expect(t, league.Videogame.Name, "LoL")
	})

	t.Run("Unmarshal Series", func(t *testing.T) {
		seriesData, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(seriesData, FlagSeries)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		series, ok := result.(pandatypes.SeriesLike)
		st.Expect(t, ok, true)
		st.Expect(t, series.ID, 346)
		st.Expect(t, series.Name, "")
		st.Expect(t, series.LeagueID, 299)
		st.Expect(t, series.Videogame.ID, 1)
	})

	t.Run("Unmarshal Tournament", func(t *testing.T) {
		tournamentData, err := os.ReadFile("../static/fetch_data/tournaments.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(tournamentData, FlagTournament)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		tournament, ok := result.(pandatypes.TournamentLike)
		st.Expect(t, ok, true)
		st.Expect(t, tournament.ID, 17283)
		st.Expect(t, tournament.Name, "Group Stage")
		st.Expect(t, tournament.SerieID, 9555)
	})

	t.Run("Unmarshal Match", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(matchData, FlagMatch)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		match, ok := result.(pandatypes.MatchLike)
		st.Expect(t, ok, true)
		st.Expect(t, match.ID, 21655)
		st.Expect(t, match.Name, "Qf 2")
		st.Expect(t, match.TournamentID, 696)
		st.Expect(t, match.NumberOfGames, 3)
	})

	t.Run("Unmarshal Team", func(t *testing.T) {
		teamData, err := os.ReadFile("../static/fetch_data/teams.json")
		st.Assert(t, err, nil)

		result, err := client.unmarshalByFlag(teamData, FlagTeam)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		team, ok := result.(pandatypes.TeamLike)
		st.Expect(t, ok, true)
		st.Expect(t, team.ID, 127652)
		st.Expect(t, team.Name, "Ares Gaming")
		st.Expect(t, team.Acronym, "")
		st.Expect(t, team.CurrentVideogame.ID, 4)
	})

	t.Run("Invalid flag", func(t *testing.T) {
		jsonData := `{"id": 1, "name": "Test"}`

		result, err := client.unmarshalByFlag([]byte(jsonData), GetChoice(999))

		st.Reject(t, err, nil)
		st.Expect(t, strings.Contains(err.Error(), "invalid flag"), true)
		st.Expect(t, result, nil)
	})

	t.Run("Bad json", func(t *testing.T) {
		badJson := `{"id": "hello", "name": "Test", "extra_field": "unexpected"}`

		result, err := client.unmarshalByFlag([]byte(badJson), FlagGame)

		st.Reject(t, err, nil)
		st.Expect(t, result, nil)
	})
}

func TestDependencyStruct(t *testing.T) {
	dep := Dependency{
		id:   123,
		flag: FlagGame,
		name: "test_entity",
	}

	st.Expect(t, dep.id, 123)
	st.Expect(t, dep.flag, FlagGame)
	st.Expect(t, dep.name, "test_entity")
}

func TestGetChoiceConstants(t *testing.T) {
	// Test that the constants have the expected values
	st.Expect(t, FlagGame, GetChoice(0))
	st.Expect(t, FlagLeague, GetChoice(1))
	st.Expect(t, FlagSeries, GetChoice(2))
	st.Expect(t, FlagTournament, GetChoice(3))
	st.Expect(t, FlagMatch, GetChoice(4))
	st.Expect(t, FlagTeam, GetChoice(5))
}

func TestExistCheck(t *testing.T) {
	//create logger
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "",
		Pandasecret: "",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         t.Context(),
	}
	// this should fail
	val, err := client.ExistCheck(0, GetChoice(1000))
	st.Reject(t, err, nil)
	st.Expect(t, val, false)
	// this also should fail due to int out of range
	val, err = client.ExistCheck(math.MaxInt32+10, GetChoice(1))
	st.Reject(t, err, nil)
	st.Expect(t, val, false)

	expectedArgs := int32(1)

	//this should fail
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnError(fmt.Errorf("some error"))
	val, err = client.ExistCheck(1, GetChoice(0))
	st.Reject(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery(`SELECT COUNT`).
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	// this should pass
	val, err = client.ExistCheck(1, GetChoice(0))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)

	// this should also pass
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(0))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	// check league existence
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(1))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	val, err = client.ExistCheck(1, GetChoice(1))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)

	// check series existence
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(2))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	val, err = client.ExistCheck(1, GetChoice(2))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)

	// check tournament existence
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(3))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	val, err = client.ExistCheck(1, GetChoice(3))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)

	// check match existence
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(4))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	val, err = client.ExistCheck(1, GetChoice(4))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)

	// check team existence
	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))

	val, err = client.ExistCheck(1, GetChoice(5))
	st.Assert(t, err, nil)
	st.Expect(t, val, false)

	mockDB.ExpectQuery("SELECT COUNT").
		WithArgs(expectedArgs).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))

	val, err = client.ExistCheck(1, GetChoice(5))
	st.Assert(t, err, nil)
	st.Expect(t, val, true)
}

func TestParseResponse(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "",
		Pandasecret: "",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - ParseResponse for League with existing game", func(t *testing.T) {
		leagueData, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		// Mock GameExist to return that game exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(1)). // videogame.id from leagues.json
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		result, err := client.ParseResponse(leagueData, FlagLeague)
		st.Expect(t, err, nil)
		st.Reject(t, result, nil)

		leagueResult, ok := result.(pandatypes.LeagueLike)
		st.Expect(t, ok, true)
		st.Expect(t, leagueResult.ID, 289)
	})

	t.Run("Success - ParseResponse for Series with missing league", func(t *testing.T) {
		seriesData, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		seriesResult, err := client.unmarshalByFlag(seriesData, FlagSeries)
		st.Assert(t, err, nil)
		series, ok := seriesResult.(pandatypes.SeriesLike)
		st.Assert(t, ok, true)

		// Mock LeagueExist to return false (league doesn't exist)
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(series.LeagueID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

		// Since league doesn't exist, we expect GetOne to fail with MakeRequest error
		result, err := client.ParseResponse(seriesData, FlagSeries)
		st.Reject(t, err, nil)
		st.Expect(t, result, nil)
	})

	t.Run("Error - ParseResponse with invalid JSON", func(t *testing.T) {
		badJson := []byte("invalid json")

		result, err := client.ParseResponse(badJson, FlagGame)
		st.Reject(t, err, nil)
		st.Expect(t, result, nil)
	})
}

func TestEnsureDependencies(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "",
		Pandasecret: "",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Success - No dependencies for Game", func(t *testing.T) {
		game := pandatypes.GameLike{
			ID:   1,
			Name: "Counter-Strike 2",
			Slug: "cs2",
		}

		err := client.ensureDependencies(game, FlagGame)
		st.Expect(t, err, nil)
	})

	t.Run("Success - Dependency exists", func(t *testing.T) {
		league := pandatypes.LeagueLike{
			ID:   289,
			Name: "NA LCS",
			Videogame: struct {
				ID             int    `json:"id"`
				Name           string `json:"name"`
				CurrentVersion string `json:"current_version"`
				Slug           string `json:"slug"`
			}{
				ID:   1,
				Name: "LoL",
				Slug: "lol",
			},
		}

		// Mock that the game exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(1)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		err := client.ensureDependencies(league, FlagLeague)
		st.Expect(t, err, nil)
	})

	t.Run("Error - ExistCheck fails", func(t *testing.T) {
		league := pandatypes.LeagueLike{
			ID:   289,
			Name: "NA LCS",
			Videogame: struct {
				ID             int    `json:"id"`
				Name           string `json:"name"`
				CurrentVersion string `json:"current_version"`
				Slug           string `json:"slug"`
			}{
				ID:   1,
				Name: "LoL",
				Slug: "lol",
			},
		}

		// Mock ExistCheck error
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(1)).
			WillReturnError(fmt.Errorf("database error"))

		err := client.ensureDependencies(league, FlagLeague)
		st.Reject(t, err, nil)
	})

	t.Run("Error - Dependency doesn't exist and GetOne fails", func(t *testing.T) {
		league := pandatypes.LeagueLike{
			ID:   289,
			Name: "NA LCS",
			Videogame: struct {
				ID             int    `json:"id"`
				Name           string `json:"name"`
				CurrentVersion string `json:"current_version"`
				Slug           string `json:"slug"`
			}{
				ID:   1,
				Name: "LoL",
				Slug: "lol",
			},
		}

		// Mock that game doesn't exist
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(1)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

		// GetOne will fail because MakeRequest will fail
		err := client.ensureDependencies(league, FlagLeague)
		st.Reject(t, err, nil)
	})
}

func TestGetOne(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "https://api.pandascore.io",
		Pandasecret: "fakesecret",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Error - Invalid flag", func(t *testing.T) {
		err := client.GetOne(1, GetChoice(999))
		st.Reject(t, err, nil)
	})

	t.Run("Error - Non-200 status code", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/videogames/1").
			Reply(404)

		err := client.GetOne(1, FlagGame)
		st.Reject(t, err, nil)
	})

	t.Run("Error - WriteToDB fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gameData, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/videogames/34").
			Reply(200).
			BodyString(string(gameData))

		// Mock database error
		mockDB.ExpectExec("INSERT INTO games").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("database error"))

		err = client.GetOne(34, FlagGame)
		st.Reject(t, err, nil)
	})
}

func TestWriteMatchesErrorPaths(t *testing.T) {
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "",
		Pandasecret: "",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         context.Background(),
	}

	t.Run("Error - Tournament ExistCheck fails", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match pandatypes.MatchLike
		err = json.Unmarshal(matchData, &match)
		st.Assert(t, err, nil)

		matches := pandatypes.MatchLikes{match}

		// Mock TournamentExist error
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.TournamentID)).
			WillReturnError(fmt.Errorf("database error"))

		client.WriteMatches(matches)
		// Should continue on error, no assertion needed
	})

	t.Run("Error - Tournament doesn't exist and GetOne fails", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match pandatypes.MatchLike
		err = json.Unmarshal(matchData, &match)
		st.Assert(t, err, nil)

		matches := pandatypes.MatchLikes{match}

		// Mock that tournament doesn't exist
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.TournamentID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

		// GetOne will fail because there's no HTTP mock
		client.WriteMatches(matches)
		// Should continue on error, no assertion needed
	})

	t.Run("Error - WriteToDB fails", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match pandatypes.MatchLike
		err = json.Unmarshal(matchData, &match)
		st.Assert(t, err, nil)

		matches := pandatypes.MatchLikes{match}

		// Mock that tournament exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.TournamentID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		// Mock WriteToDB error
		mockDB.ExpectExec("INSERT INTO matches").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(fmt.Errorf("database error"))

		client.WriteMatches(matches)
		// Should continue on error, no assertion needed
	})
}

func TestCheckTeam(t *testing.T) {
	//create logger
	logger := zaptest.NewLogger(t).Sugar()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	defer mockDB.Close()
	mockQueries := dbtypes.New(mockDB)

	client := &PandaClient{
		BaseURL:     "",
		Pandasecret: "",
		Logger:      logger,
		HTTPClient:  &http.Client{},
		DBConnector: mockQueries,
		Run:         0,
		Ctx:         t.Context(),
	}

	//setup MatchLike material by opening static/fetch_data/matches.json
	matchData, err := os.ReadFile("../static/fetch_data/matches.json")
	st.Assert(t, err, nil)
	pdDataLike, err := client.unmarshalByFlag(matchData, FlagMatch)
	st.Assert(t, err, nil)
	st.Reject(t, pdDataLike, nil)
	match, ok := pdDataLike.(pandatypes.MatchLike)
	st.Assert(t, ok, true)
	st.Reject(t, match, nil)
	//we first run the test where the winner type is not a team
	match.WinnerType = "Player"
	// this has no outputs
	client.checkTeam(match)
	r := recover()
	st.Expect(t, r, nil)
	match, ok = pdDataLike.(pandatypes.MatchLike)
	st.Assert(t, ok, true)
	st.Reject(t, match, nil)

	// we first set two cases where the request errors
	cancelContext, cancel := context.WithCancel(context.Background())
	client.Ctx = cancelContext
	// the first case fails because the client.ExistCheck fails (out of range)
	match.Opponents[0].Opponent.ID = math.MaxInt32 + 1
	client.checkTeam(match)
	r = recover()
	st.Expect(t, r, nil)
	match.Opponents[0].Opponent.ID = 1

	mockDB.ExpectQuery("SELECT COUNT").WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))
	mockDB.ExpectQuery("SELECT COUNT").WithArgs(int32(match.Opponents[1].Opponent.ID)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))
	client.checkTeam(match)
	cancel()
	r = recover()
	st.Expect(t, r, nil)
	st.Expect(t, mockDB.ExpectationsWereMet(), nil)

	client.Ctx = context.Background()
	match = pdDataLike.(pandatypes.MatchLike)
	mockDB.ExpectQuery("SELECT COUNT").WithArgs(int32(1)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(1)))
	mockDB.ExpectQuery("SELECT COUNT").WithArgs(int32(match.Opponents[1].Opponent.ID)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int32(0)))
	mockDB.ExpectExec("INSERT INTO teams").
		WithArgs(
			int32(match.Opponents[1].Opponent.ID),
			match.Opponents[1].Opponent.Name,
			pgtype.Text{String: match.Opponents[1].Opponent.Slug, Valid: match.Opponents[1].Opponent.Slug != ""},
			pgtype.Text{String: match.Opponents[1].Opponent.Acronym,
				Valid: match.Opponents[1].Opponent.Acronym != ""},
			pgtype.Text{String: match.Opponents[1].Opponent.ImageURL, Valid: match.Opponents[1].Opponent.ImageURL != ""},
			int32(match.Videogame.ID)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	t.Log(match.Opponents)
	client.checkTeam(match)
	r = recover()
	st.Expect(t, r, nil)
	st.Expect(t, mockDB.ExpectationsWereMet(), nil)
}
