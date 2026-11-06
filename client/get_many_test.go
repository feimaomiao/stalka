package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/feimaomiao/stalka/dbtypes"
	"github.com/feimaomiao/stalka/pandatypes"
	"github.com/h2non/gock"
	"github.com/nbio/st"
	"github.com/pashagolub/pgxmock/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestUpdateGames(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - UpdateGames", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gameData, err := os.ReadFile("../static/fetch_data/videogames.json")
		st.Assert(t, err, nil)

		var gameResponse pandatypes.GameLike
		err = json.Unmarshal(gameData, &gameResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			Reply(200).
			BodyString("[" + string(gameData) + "]")

		mockDB.ExpectExec("INSERT INTO games").
			WithArgs(
				int32(gameResponse.ID),
				gameResponse.Name,
				pgxmock.AnyArg(),
			).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.UpdateGames()
		st.Expect(t, err, nil)
		st.Expect(t, gock.IsDone(), true)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.UpdateGames()
		st.Reject(t, err, nil)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/videogames").
			Reply(200).
			BodyString("invalid json")

		err := client.UpdateGames()
		st.Reject(t, err, nil)
	})
}

func TestGetLeagues(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - GetLeagues with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		leagueData, err := os.ReadFile("../static/fetch_data/leagues.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/leagues").
			MatchParam("sort", "-modified_at").
			MatchParam("page", "1").
			Reply(200).
			BodyString("[" + string(leagueData) + "]")

		mockDB.ExpectExec("INSERT INTO leagues").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetLeagues(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/leagues").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetLeagues(false)
		st.Reject(t, err, nil)
	})
}

func TestGetSeries(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - GetSeries with existing league", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		seriesData, err := os.ReadFile("../static/fetch_data/series.json")
		st.Assert(t, err, nil)

		var seriesResponse pandatypes.SeriesLike
		err = json.Unmarshal(seriesData, &seriesResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/series").
			Reply(200).
			BodyString("[" + string(seriesData) + "]")

		// Mock LeagueExist to return that league exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(seriesResponse.LeagueID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO series").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetSeries(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/series").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetSeries(false)
		st.Reject(t, err, nil)
	})
}

func TestGetTournaments(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - GetTournaments with existing series", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		tournamentData, err := os.ReadFile("../static/fetch_data/tournaments.json")
		st.Assert(t, err, nil)

		var tournamentResponse pandatypes.TournamentLike
		err = json.Unmarshal(tournamentData, &tournamentResponse)
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/tournaments").
			MatchParam("sort", "-modified_at").
			MatchParam("page", "1").
			Reply(200).
			BodyString("[" + string(tournamentData) + "]")

		// Mock SeriesExist to return that series exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(tournamentResponse.SerieID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO tournaments").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetTournaments(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/tournaments").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetTournaments(false)
		st.Reject(t, err, nil)
	})
}

func TestGetMatchPage(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - getMatchPage for upcoming matches (even page)", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			MatchParam("page", "1").
			Reply(200).
			BodyString("[" + string(matchData) + "]")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Expect(t, result.Err, nil)
		st.Reject(t, result.Matches, nil)
		st.Expect(t, len(result.Matches) > 0, true)
	})

	t.Run("Success - getMatchPage for past matches (odd page)", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/matches/past").
			MatchParam("page", "1").
			Reply(200).
			BodyString("[" + string(matchData) + "]")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(1, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Expect(t, result.Err, nil)
		st.Reject(t, result.Matches, nil)
		st.Expect(t, len(result.Matches) > 0, true)
	})

	t.Run("Error - Non-200 status code", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(404)

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		// Note: The implementation has a bug where err is nil on non-200 status
		// but Matches is still an empty slice
		st.Expect(t, len(result.Matches), 0)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(200).
			BodyString("invalid json")

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		st.Reject(t, result.Err, nil)
		st.Expect(t, len(result.Matches), 0)
	})
}

func TestGetMatches(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - GetMatches with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var matchResponse pandatypes.MatchLike
		err = json.Unmarshal(matchData, &matchResponse)
		st.Assert(t, err, nil)

		// Mock both upcoming and past matches endpoints for all pages
		for i := 0; i < Pages; i++ {
			if i%2 == 0 {
				gock.New("https://api.pandascore.io").
					Get("/matches/upcoming").
					MatchParam("page", strings.Join([]string{}, "")).
					Reply(200).
					BodyString("[" + string(matchData) + "]")
			} else {
				gock.New("https://api.pandascore.io").
					Get("/matches/past").
					Reply(200).
					BodyString("[" + string(matchData) + "]")
			}
		}

		// Mock database expectations for tournament existence check and match write
		for i := 0; i < Pages; i++ {
			mockDB.ExpectQuery("SELECT COUNT").
				WithArgs(int32(matchResponse.TournamentID)).
				WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

			mockDB.ExpectExec("INSERT INTO matches").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		}

		err = client.GetMatches(false)
		st.Expect(t, err, nil)
	})
}

func TestGetTeams(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - GetTeams with setup=false", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		teamData, err := os.ReadFile("../static/fetch_data/teams.json")
		st.Assert(t, err, nil)

		gock.New("https://api.pandascore.io").
			Get("/teams").
			Reply(200).
			BodyString("[" + string(teamData) + "]")

		mockDB.ExpectExec("INSERT INTO teams").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = client.GetTeams(false)
		st.Expect(t, err, nil)
	})

	t.Run("Error - MakeRequest fails", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/teams").
			ReplyError(io.ErrUnexpectedEOF)

		err := client.GetTeams(false)
		st.Reject(t, err, nil)
	})

	t.Run("Error - Invalid JSON response", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/teams").
			Reply(200).
			BodyString("invalid json")

		err := client.GetTeams(false)
		st.Reject(t, err, nil)
	})
}

func TestWriteMatches(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Success - WriteMatches with existing tournament", func(t *testing.T) {
		matchData, err := os.ReadFile("../static/fetch_data/matches.json")
		st.Assert(t, err, nil)

		var match pandatypes.MatchLike
		err = json.Unmarshal(matchData, &match)
		st.Assert(t, err, nil)

		matches := pandatypes.MatchLikes{match}

		// Mock TournamentExist to return that tournament exists
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.TournamentID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectExec("INSERT INTO matches").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		// Expect team checks for both teams
		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.Opponents[0].Opponent.ID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		mockDB.ExpectQuery("SELECT COUNT").
			WithArgs(int32(match.Opponents[1].Opponent.ID)).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

		client.WriteMatches(matches)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})
}

func TestConstants(t *testing.T) {
	// Test that the constants have the expected values
	st.Expect(t, sortedBy, "-modified_at")
	st.Expect(t, Pages, 20)
	st.Expect(t, SetupPages, 50)
}

func TestGetMatchPageReadError(t *testing.T) {
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
		Ctx:         context.Background(),
	}

	t.Run("Error - Non-200 status code", func(t *testing.T) {
		gock.InterceptClient(client.HTTPClient)
		defer gock.Off()

		gock.New("https://api.pandascore.io").
			Get("/matches/upcoming").
			Reply(404)

		var wg sync.WaitGroup
		ch := make(chan pandatypes.ResultMatchLikes, 1)

		wg.Add(1)
		client.getMatchPage(0, &wg, ch)
		wg.Wait()
		close(ch)

		result := <-ch
		// Note: The implementation has a bug where err is nil on non-200 status
		// but Matches is still an empty slice
		st.Expect(t, len(result.Matches), 0)
	})
}

const livesTestBaseURL = "https://api.pandascore.test/"

func newLivesClient(t *testing.T) (*PandaClient, pgxmock.PgxPoolIface) {
	t.Helper()
	mockDB, err := pgxmock.NewPool()
	st.Assert(t, err, nil)
	t.Cleanup(mockDB.Close)
	client := &PandaClient{
		BaseURL:     livesTestBaseURL,
		Pandasecret: "test_secret",
		Logger:      zap.NewNop().Sugar(),
		HTTPClient:  &http.Client{},
		DBConnector: dbtypes.New(mockDB),
		Ctx:         context.Background(),
	}
	gock.InterceptClient(client.HTTPClient)
	t.Cleanup(gock.Off)
	return client, mockDB
}

func TestGetLives(t *testing.T) {
	const running = `[
		{"id": 1, "name": "Live 1", "tournament_id": 1000, "number_of_games": 3,
		 "scheduled_at": "2026-05-01T14:00:00Z"},
		{"id": 2, "name": "Live 2", "tournament_id": 1000, "number_of_games": 3,
		 "scheduled_at": "2026-05-01T15:00:00Z"}
	]`

	t.Run("Success - syncs is_live to the running ids", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).
			Get("/matches/running/").
			MatchHeader("Authorization", "Bearer test_secret").
			Reply(http.StatusOK).
			BodyString(running)
		for range 2 {
			mockDB.ExpectQuery("SELECT COUNT").
				WithArgs(int32(1000)).
				WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))
			mockDB.ExpectExec("INSERT INTO matches").
				WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
					pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
					pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
					pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
		}
		mockDB.ExpectQuery("UPDATE MATCHES SET is_live").
			WithArgs([]int32{1, 2}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "is_live"}).AddRow(int32(1), true).AddRow(int32(2), true))

		st.Expect(t, client.GetLives(), nil)
		st.Expect(t, gock.IsDone(), true)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	// A nil slice reaches Postgres as NULL, which is why this pins the empty
	// non-nil slice rather than just "the query ran".
	t.Run("Success - nothing running sends an empty list", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[]")
		mockDB.ExpectQuery("UPDATE MATCHES SET is_live").
			WithArgs([]int32{}).
			WillReturnRows(pgxmock.NewRows([]string{"id", "is_live"}))

		st.Expect(t, client.GetLives(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("Error - non-200 leaves is_live untouched", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusForbidden)

		st.Reject(t, client.GetLives(), nil)
		st.Expect(t, mockDB.ExpectationsWereMet(), nil)
	})

	t.Run("Error - sync failure is returned", func(t *testing.T) {
		client, mockDB := newLivesClient(t)
		gock.New(livesTestBaseURL).Get("/matches/running/").Reply(http.StatusOK).BodyString("[]")
		mockDB.ExpectQuery("UPDATE MATCHES SET is_live").
			WithArgs([]int32{}).
			WillReturnError(io.ErrUnexpectedEOF)

		st.Reject(t, client.GetLives(), nil)
	})
}

func TestGetMatchPageTransportError(t *testing.T) {
	client, _ := newLivesClient(t)
	gock.New(livesTestBaseURL).Get("/matches/upcoming/").ReplyError(io.ErrUnexpectedEOF)

	var wg sync.WaitGroup
	ch := make(chan pandatypes.ResultMatchLikes, 1)
	wg.Add(1)
	client.getMatchPage(0, &wg, ch)
	wg.Wait()

	result := <-ch
	st.Reject(t, result.Err, nil)
	st.Expect(t, len(result.Matches), 0)
}

func TestGetMatchPageNon200IsAnError(t *testing.T) {
	client, _ := newLivesClient(t)
	gock.New(livesTestBaseURL).Get("/matches/upcoming/").Reply(http.StatusForbidden)

	var wg sync.WaitGroup
	ch := make(chan pandatypes.ResultMatchLikes, 1)
	wg.Add(1)
	client.getMatchPage(0, &wg, ch)
	wg.Wait()

	st.Reject(t, (<-ch).Err, nil)
}

// Pages are 1-indexed: asking for page=0 returns nothing, so a refresh that
// starts there silently does no work.
func TestMatchPagesStartAtOne(t *testing.T) {
	cases := []struct {
		name     string
		page     int
		endpoint string
		want     string
	}{
		{name: "first upcoming", page: 0, endpoint: "/matches/upcoming/", want: "1"},
		{name: "first past", page: 1, endpoint: "/matches/past/", want: "1"},
		{name: "second upcoming", page: 2, endpoint: "/matches/upcoming/", want: "2"},
		{name: "second past", page: 3, endpoint: "/matches/past/", want: "2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newLivesClient(t)
			gock.New(livesTestBaseURL).
				Get(tc.endpoint).
				MatchParam("page", tc.want).
				Reply(http.StatusOK).
				BodyString("[]")

			var wg sync.WaitGroup
			ch := make(chan pandatypes.ResultMatchLikes, 1)
			wg.Add(1)
			client.getMatchPage(tc.page, &wg, ch)
			wg.Wait()

			st.Expect(t, (<-ch).Err, nil)
			st.Expect(t, gock.IsDone(), true)
		})
	}
}

// Every bulk fetcher must send sort and page, and setup must walk distinct
// pages instead of re-fetching the first one Pages times.
func TestBulkFetchersPaginate(t *testing.T) {
	fetchers := []struct {
		name     string
		endpoint string
		fetch    func(*PandaClient, bool) error
	}{
		{name: "leagues", endpoint: "/leagues/", fetch: (*PandaClient).GetLeagues},
		{name: "series", endpoint: "/series/", fetch: (*PandaClient).GetSeries},
		{name: "tournaments", endpoint: "/tournaments/", fetch: (*PandaClient).GetTournaments},
		{name: "teams", endpoint: "/teams/", fetch: (*PandaClient).GetTeams},
	}
	for _, fetcher := range fetchers {
		t.Run(fetcher.name+" refresh reads the first page", func(t *testing.T) {
			client, _ := newLivesClient(t)
			gock.New(livesTestBaseURL).
				Get(fetcher.endpoint).
				MatchParam("page", "1").
				MatchParam("sort", sortedBy).
				Reply(http.StatusOK).
				BodyString("[]")

			st.Expect(t, fetcher.fetch(client, false), nil)
			st.Expect(t, gock.IsDone(), true)
		})
		t.Run(fetcher.name+" setup walks every page once", func(t *testing.T) {
			client, _ := newLivesClient(t)
			for page := firstPage; page < firstPage+Pages; page++ {
				gock.New(livesTestBaseURL).
					Get(fetcher.endpoint).
					MatchParam("page", strconv.Itoa(page)+"$").
					MatchParam("sort", sortedBy).
					Reply(http.StatusOK).
					BodyString("[]")
			}

			st.Expect(t, fetcher.fetch(client, true), nil)
			st.Expect(t, gock.IsDone(), true)
		})
	}
}

func TestMakeRequestCountsConcurrentRequests(t *testing.T) {
	client, _ := newLivesClient(t)
	const requests = 20
	gock.New(livesTestBaseURL).Get("/videogames/").Times(requests).Reply(http.StatusOK).BodyString("[]")

	var wg sync.WaitGroup
	for range requests {
		wg.Go(func() {
			resp, err := client.MakeRequest([]string{"videogames"}, nil)
			if err == nil {
				_ = resp.Body.Close()
			}
		})
	}
	wg.Wait()

	st.Expect(t, client.Run.Load(), int64(requests))
}
