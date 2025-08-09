package client

import (
	"os"
	"strings"
	"testing"

	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/nbio/st"
	"go.uber.org/zap"
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
		Logger: zap.NewNop().Sugar(),
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
